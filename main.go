package main

import (
	"fmt"
	"github.com/op/go-logging"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

const (
	EnvDef  = "ENV_DEF"
	EnvPort = "PORT"
)

var (
	port         = 8081
	workspaceDir = "/workspace"
	rootDir      = "/public"
	log          = logging.MustGetLogger("webapp-go")
	logFormat    = logging.MustStringFormatter(
		`%{color}%{time:15:04:05} [%{level}] %{color:reset} %{message}`,
	)
	replaceEnvMap    = make(map[string]string)
	replaceEnvExtMap = map[string]int{
		".html": 1, ".js": 1, ".css": 1, ".json": 1,
	}
)

type FileHandle func(fileAbs string) error

func init() {
	// fix http.FileServer mime types
	// see: https://stackoverflow.com/questions/70716366/how-can-i-set-correct-http-fileserver-mime-types
	mime.AddExtensionType(".js", "application/javascript")

	// init log
	initLog()

	// 初始化环境变量配置
	var ok bool
	var envResult string
	var envName string
	var envValue string
	envResult, ok = os.LookupEnv(EnvPort)
	if ok {
		port, _ = strconv.Atoi(envResult)
		log.Infof("Read Env: %s => %s", EnvPort, port)
	}
	envResult, ok = os.LookupEnv(EnvDef)
	if ok {
		log.Infof("Read Env: %s => %s", EnvDef, envResult)
		for _, envName = range strings.Split(envResult, " ") {
			envValue, ok = os.LookupEnv(envName)
			if ok {
				log.Infof("Read Env: %s => %s", envName, envValue)
				replaceEnvMap[envName] = envValue
			}
		}
	}
	// 更新静态资源
	initStaticResources()
}

func initLog() {
	infoBackend := logging.NewLogBackend(os.Stdout, "", 0)
	infoFormatter := logging.NewBackendFormatter(infoBackend, logFormat)
	infoLeveled := logging.AddModuleLevel(infoFormatter)
	infoLeveled.SetLevel(logging.INFO, "")

	errorBackend := logging.NewLogBackend(os.Stderr, "", 0)
	errorFormatter := logging.NewBackendFormatter(errorBackend, logFormat)
	errorLeveled := logging.AddModuleLevel(errorFormatter)
	errorLeveled.SetLevel(logging.ERROR, "")

	logging.SetBackend(infoLeveled, errorLeveled)
}

func initStaticResources() {
	log.Info(">>>>>>>>>>>>>>>>>>>>>>>>>>> Deploy Start <<<<<<<<<<<<<<<<<<<<<<<<<<<<")
	// Copy workspace dir to root dir, and replace env
	var err error
	LoopFileHandle(workspaceDir, func(fileAbs string) error {
		toFileAbs := strings.ReplaceAll(fileAbs, workspaceDir, rootDir)
		log.Infof("Copy %s to %s", fileAbs, toFileAbs)
		if err = CopyFile(fileAbs, toFileAbs); err != nil {
			return err
		} else {
			ext := filepath.Ext(toFileAbs)
			if _, ok := replaceEnvExtMap[ext]; !ok {
				return nil
			}
			for name, value := range replaceEnvMap {
				if err = ReplaceEnvFile(toFileAbs, name, value); err != nil {
					return err
				}
			}
			return nil
		}
	})
	log.Info(">>>>>>>>>>>>>>>>>>>>>>>>>>> Deploy End <<<<<<<<<<<<<<<<<<<<<<<<<<<<")
}

func CopyFile(source, target string) error {
	var err error
	var sourceFd *os.File
	var targetFd *os.File
	var sourceInfo os.FileInfo

	if sourceFd, err = os.Open(source); err != nil {
		return err
	}
	defer sourceFd.Close()

	if targetFd, err = os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644); err != nil {
		return err
	}
	defer targetFd.Close()

	if _, err = io.Copy(targetFd, sourceFd); err != nil {
		return err
	}

	if sourceInfo, err = os.Stat(source); err != nil {
		return err
	}
	return os.Chmod(target, sourceInfo.Mode())
}

func ReplaceEnvFile(fileAbs, envName, envValue string) error {
	readBytes, err := ioutil.ReadFile(fileAbs)
	if err != nil {
		return err
	}
	log.Infof("Replace %s with %s in the %s", envName, envValue, fileAbs)
	var replaceContent = string(readBytes)
	if envValue == "/" {
		replaceContent = strings.ReplaceAll(replaceContent, "/"+envName+"/", envName)
		replaceContent = strings.ReplaceAll(replaceContent, "/"+envName, envName)
		replaceContent = strings.ReplaceAll(replaceContent, envName+"/", envName)
	} else {
		replaceContent = strings.ReplaceAll(replaceContent, "/"+envName, envName)
		replaceContent = strings.ReplaceAll(replaceContent, envName+"/", envName)
	}
	replaceContent = strings.ReplaceAll(replaceContent, envName, envValue)
	err = ioutil.WriteFile(fileAbs, []byte(replaceContent), 0)
	if err != nil {
		return err
	}
	return nil
}

func LoopFileHandle(fileAbs string, fileHandle FileHandle) {
	var err error
	var fileInfo os.FileInfo
	var fileInfos []os.FileInfo
	if fileInfo, err = os.Stat(fileAbs); err != nil {
		log.Error(err.Error())
	} else {
		if fileInfo.IsDir() {
			if fileInfos, err = ioutil.ReadDir(fileAbs); err != nil {
				log.Error(err.Error())
			} else {
				for _, fileInfo = range fileInfos {
					LoopFileHandle(path.Join(fileAbs, fileInfo.Name()), fileHandle)
				}
			}
		} else {
			if err = fileHandle(fileAbs); err != nil {
				log.Error(err.Error())
			}
		}
	}
}

func WebAppFileServer(root http.FileSystem) http.Handler {
	fileServer := http.FileServer(root)
	http.Handle("/", fileServer)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fileServer.ServeHTTP(w, r)
	})
}

func main() {
	runtime.GOMAXPROCS(1)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), WebAppFileServer(http.Dir(rootDir)))
	if err != nil {
		log.Error(err.Error())
	}
}
