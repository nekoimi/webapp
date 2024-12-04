package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	EnvDef    = "ENV_DEF"
	EnvPrefix = "WEBAPP_ENV."
)

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
	loadEnv()
}

func loadEnv() {
	log.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>> LoadEnv Start <<<<<<<<<<<<<<<<<<<<<<<<<<<<")
	var (
		lookupVal string
		envKey    string
		envVal    string
	)

	// 兼容老版本写法
	if lookupVal = os.Getenv(EnvDef); lookupVal != "" {
		for _, envKey = range strings.Split(lookupVal, " ") {
			if envVal = os.Getenv(envKey); envVal != "" {
				log.Printf("Read Env: %s => %s", envKey, envVal)
				replaceEnvMap[envKey] = envVal
			}
		}
	}

	// 读取系统环境变量
	// 获取 WEBAPP_ENV. 为前缀的环境变量
	sysEnvs := os.Environ()
	for _, sysEnv := range sysEnvs {
		parts := strings.SplitN(sysEnv, "=", 2)
		if len(parts) == 2 {
			sysEnvKey := parts[0]
			if strings.HasPrefix(sysEnvKey, EnvPrefix) {
				if envVal = os.Getenv(sysEnvKey); envVal != "" {
					envKey = strings.Replace(sysEnvKey, EnvPrefix, "", 1)
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

func fileExists(f string) (bool, error) {
	if _, err := os.Stat(f); err == nil {
		// f exists
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		// f does not exists
		return false, nil
	} else {
		// f stat err, return false and err
		return false, err
	}
}

func copyFile(src, dst string) error {
	if exists, err := fileExists(src); err != nil {
		return err
	} else if !exists {
		return errors.New("file " + src + " does not exists")
	}

	if sfi, err := os.Lstat(src); err != nil {
		return err
	} else {
		sf, err := os.Open(src)
		defer sf.Close()
		if err != nil {
			return err
		}

		df, err := os.Create(dst)
		defer df.Close()

		_, err = io.Copy(df, sf)
		if err != nil {
			return err
		}

		return os.Chmod(dst, sfi.Mode())
	}
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		relativePath := strings.Replace(path, src, "", 1)
		if len(relativePath) == 0 {
			return nil
		}

		dstPath := filepath.Join(dst, relativePath)
		if d.IsDir() {
			if exists, err := fileExists(dstPath); err != nil {
				return err
			} else if !exists {
				if err := os.Mkdir(dstPath, os.ModeDir); err != nil {
					return err
				}
			}
		} else {
			return copyFile(path, dstPath)
		}
		return nil
	})
}

func replaceEnv(f, envKey, envVal string) error {
	byteArr, err := os.ReadFile(f)
	if err != nil {
		return errors.New(fmt.Sprintf("[ReplaceEnv] Read file [%s] error, %s", f, err.Error()))
	}
	log.Printf("Replace %s with %s in the %s", envKey, envVal, f)
	var content = string(byteArr)
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
		return errors.New(fmt.Sprintf("[ReplaceEnv] Write file [%s] error, %s", f, err.Error()))
	}
	return nil
}

func deploy() {
	log.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>> Deploy Start <<<<<<<<<<<<<<<<<<<<<<<<<<<<")
	err := copyDir(workspaceDir, nginxRootDir)
	if err != nil {
		log.Fatalln(err.Error())
	}

	_ = filepath.WalkDir(nginxRootDir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if _, ok := replaceExtMap[ext]; !ok {
			log.Printf("Ext: %s, Ignore replace.", ext)
			return nil
		}

		for _, k := range sortEnvKeys {
			if err = replaceEnv(path, k, replaceEnvMap[k]); err != nil {
				return err
			}
		}
		return nil
	})
	log.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>> Deploy End <<<<<<<<<<<<<<<<<<<<<<<<<<<<")
}

func main() {
	deploy()
}
