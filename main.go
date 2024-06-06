package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

const (
	EnvDef    = "ENV_DEF"
	EnvPrefix = "WEBAPP_ENV."
)

type fileHandle func(f string) error

var (
	workspaceDir  = "/workspace"
	nginxRootDir  = "/usr/share/nginx/html"
	sortEnvKeys   []string
	replaceEnvMap = make(map[string]string)
	replaceExtMap = map[string]bool{
		".html": true, ".js": true, ".css": true, ".json": true,
	}
)

func init() {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Lmicroseconds)
	loadWebappEnv()
}

func loadWebappEnv() {
	log.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>> LoadEnv Start <<<<<<<<<<<<<<<<<<<<<<<<<<<<")
	var (
		ok        bool
		lookupVal string
		envKey    string
		envVal    string
	)

	// 兼容老版本写法
	lookupVal, ok = os.LookupEnv(EnvDef)
	if ok {
		for _, envKey = range strings.Split(lookupVal, " ") {
			envVal, ok = os.LookupEnv(envKey)
			if ok {
				log.Printf("Read Env: %s => %s", envKey, envVal)
				replaceEnvMap[envKey] = envVal
			}
		}
	}

	// 读取系统环境变量
	// 获取 WEBAPP_ENV. 为前缀的环境变量
	systemEnvs := os.Environ()
	for _, systemEnv := range systemEnvs {
		parts := strings.SplitN(systemEnv, "=", 2)
		if len(parts) == 2 {
			systemEnvKey := parts[0]
			if strings.HasPrefix(systemEnvKey, EnvPrefix) {
				envVal, ok = os.LookupEnv(systemEnvKey)
				if ok {
					envKey = strings.Replace(systemEnvKey, EnvPrefix, "", 1)
					log.Printf("Read Env: %s => %s", envKey, envVal)
					replaceEnvMap[envKey] = envVal
				}
			}
		}
	}

	sortEnvKeys = make([]string, 0, len(replaceEnvMap))
	for k := range replaceEnvMap {
		sortEnvKeys = append(sortEnvKeys, k)
	}

	// 排序，优先替换名称长的环境变量
	sort.Slice(sortEnvKeys, func(i, j int) bool {
		return len(sortEnvKeys[i]) > len(sortEnvKeys[j])
	})

	for _, key := range sortEnvKeys {
		log.Printf("Sort Env: %s", key)
	}
	log.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>> LoadEnv End <<<<<<<<<<<<<<<<<<<<<<<<<<<<")
}

func copyFile(source, target string) error {
	var err error
	var sourceName string
	var targetDir string
	var sourceFd *os.File
	var targetFd *os.File
	var sourceInfo os.FileInfo

	if sourceInfo, err = os.Stat(source); err != nil {
		return errors.New(fmt.Sprintf("[CopyFile] Stat source file [%s] error, %s", source, err.Error()))
	}
	sourceName = sourceInfo.Name()
	targetDir = strings.Replace(target, sourceName, "", -1)
	if _, err = os.Stat(targetDir); err != nil {
		if err = os.MkdirAll(targetDir, os.ModeDir); err != nil {
			return errors.New(fmt.Sprintf("[CopyFile] Mkdir target dir [%s] error, %s", targetDir, err.Error()))
		}
	}

	if sourceFd, err = os.Open(source); err != nil {
		return errors.New(fmt.Sprintf("[CopyFile] Open source file [%s] error, %s", source, err.Error()))
	}
	defer sourceFd.Close()

	if targetFd, err = os.Create(target); err != nil {
		return errors.New(fmt.Sprintf("[CopyFile] Craete target file [%s] error, %s", target, err.Error()))
	}
	defer targetFd.Close()

	if _, err = io.Copy(targetFd, sourceFd); err != nil {
		return errors.New(fmt.Sprintf("[CopyFile] Copy file [%s] to [%s] error, %s", source, target, err.Error()))
	}

	return os.Chmod(target, sourceInfo.Mode())
}

func replaceEnv(f, envKey, envVal string) error {
	bytes, err := os.ReadFile(f)
	if err != nil {
		return errors.New(fmt.Sprintf("[ReplaceEnvFile] Read file [%s] error, %s", f, err.Error()))
	}
	log.Printf("Replace %s with %s in the %s", envKey, envVal, f)
	var content = string(bytes)
	if envVal == "/" {
		content = strings.ReplaceAll(content, "/"+envKey+"/", envKey)
		content = strings.ReplaceAll(content, "/"+envKey, envKey)
		content = strings.ReplaceAll(content, envKey+"/", envKey)
	} else {
		if strings.HasPrefix(envVal, "/") {
			content = strings.ReplaceAll(content, "/"+envKey, envKey)
		}
		if strings.HasSuffix(envVal, "/") {
			content = strings.ReplaceAll(content, envKey+"/", envKey)
		}
	}
	content = strings.ReplaceAll(content, envKey, envVal)
	err = os.WriteFile(f, []byte(content), 0)
	if err != nil {
		return errors.New(fmt.Sprintf("[ReplaceEnvFile] Write file [%s] error, %s", f, err.Error()))
	}
	return nil
}

func recursionFileHandle(fileAbs string, handle fileHandle) {
	fi, err := os.Stat(fileAbs)
	if err != nil {
		log.Println(err.Error())
		return
	}

	if fi.IsDir() {
		entries, err := os.ReadDir(fileAbs)
		if err != nil {
			log.Println(err.Error())
			return
		}
		for _, entry := range entries {
			recursionFileHandle(path.Join(fileAbs, entry.Name()), handle)
		}
	} else {
		if err = handle(fileAbs); err != nil {
			log.Println(err.Error())
		}
	}
}

func loadNginxDist() {
	log.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>> Deploy Start <<<<<<<<<<<<<<<<<<<<<<<<<<<<")
	var err error
	recursionFileHandle(workspaceDir, func(f string) error {
		toAbs := strings.Replace(f, workspaceDir, nginxRootDir, 1)
		log.Printf("Copy %s to %s", f, toAbs)
		if err = copyFile(f, toAbs); err != nil {
			return err
		} else {
			ext := filepath.Ext(toAbs)
			if _, ok := replaceExtMap[ext]; !ok {
				log.Printf("Ext: %s, Ignore replace.", ext)
				return nil
			}

			for _, k := range sortEnvKeys {
				if err = replaceEnv(toAbs, k, replaceEnvMap[k]); err != nil {
					return err
				}
			}
			return nil
		}
	})
	log.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>> Deploy End <<<<<<<<<<<<<<<<<<<<<<<<<<<<")
}

func main() {
	loadNginxDist()
}
