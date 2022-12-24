package main

import (
	"errors"
	"fmt"
	"github.com/op/go-logging"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	AppName      = "webapp-go"
	AppEnvPrefix = "WEBAPP_ENV."
	EnvPort      = "PORT"
	EnvDef       = "ENV_DEF"
)

var (
	port         = 80
	workspaceDir = "/workspace"
	rootDir      = "/public"
	log          = logging.MustGetLogger(AppName)
	logFormat    = logging.MustStringFormatter(
		`%{color}%{time:15:04:05} [%{level}] %{color:reset} %{message}`,
	)
	replaceEnvMap    = make(map[string]string)
	replaceEnvExtMap = map[string]bool{
		".html": true, ".js": true, ".css": true, ".json": true,
	}
)

type FileHandle func(fileAbs string) error

func init() {
	// fix http.FileServer mime types
	// see: https://stackoverflow.com/questions/70716366/how-can-i-set-correct-http-fileserver-mime-types
	mime.AddExtensionType(".js", "application/javascript")

	// init log
	initLog()

	// init env
	initEnv()

	// init static resources
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

func initEnv() {
	log.Info(">>>>>>>>>>>>>>>>>>>>>>>>>>> LoadEnv Start <<<<<<<<<<<<<<<<<<<<<<<<<<<<")
	var ok bool
	var envResult string
	var envName string
	var envValue string
	envResult, ok = os.LookupEnv(EnvPort)
	if ok {
		port, _ = strconv.Atoi(envResult)
		log.Infof("Read Env: %s => %d", EnvPort, port)
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
	systemEnvs := os.Environ()
	for _, systemEnv := range systemEnvs {
		parts := strings.SplitN(systemEnv, "=", 2)
		if len(parts) == 2 {
			systemEnvName := parts[0]
			if strings.HasPrefix(systemEnvName, AppEnvPrefix) {
				envValue, ok = os.LookupEnv(systemEnvName)
				if ok {
					replaceEnvName := strings.Replace(systemEnvName, AppEnvPrefix, "", 1)
					log.Infof("Read Env: %s => %s", replaceEnvName, envValue)
					replaceEnvMap[replaceEnvName] = envValue
				}
			}
		}
	}
	log.Info(">>>>>>>>>>>>>>>>>>>>>>>>>>> LoadEnv End <<<<<<<<<<<<<<<<<<<<<<<<<<<<")
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
				log.Infof("Ext: %s, Ignore replace.", ext)
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

func ReplaceEnvFile(fileAbs, envName, envValue string) error {
	readBytes, err := ioutil.ReadFile(fileAbs)
	if err != nil {
		return errors.New(fmt.Sprintf("[ReplaceEnvFile] Read file [%s] error, %s", fileAbs, err.Error()))
	}
	log.Infof("Replace %s with %s in the %s", envName, envValue, fileAbs)
	var replaceContent = string(readBytes)
	if envValue == "/" {
		replaceContent = strings.ReplaceAll(replaceContent, "/"+envName+"/", envName)
		replaceContent = strings.ReplaceAll(replaceContent, "/"+envName, envName)
		replaceContent = strings.ReplaceAll(replaceContent, envName+"/", envName)
	} else {
		if strings.HasPrefix(envValue, "/") {
			replaceContent = strings.ReplaceAll(replaceContent, "/"+envName, envName)
		}
		if strings.HasSuffix(envValue, "/") {
			replaceContent = strings.ReplaceAll(replaceContent, envName+"/", envName)
		}
	}
	replaceContent = strings.ReplaceAll(replaceContent, envName, envValue)
	err = ioutil.WriteFile(fileAbs, []byte(replaceContent), 0)
	if err != nil {
		return errors.New(fmt.Sprintf("[ReplaceEnvFile] Write file [%s] error, %s", fileAbs, err.Error()))
	}
	return nil
}

func LoopFileHandle(fileAbs string, fileHandle FileHandle) {
	var err error
	var fileInfo os.FileInfo
	var fileInfos []os.FileInfo
	if fileInfo, err = os.Stat(fileAbs); err != nil {
		log.Errorf("Stat file [%s] error, %s", fileAbs, err.Error())
	} else {
		if fileInfo.IsDir() {
			if fileInfos, err = ioutil.ReadDir(fileAbs); err != nil {
				log.Errorf("Read dir [%s] error, %s", fileAbs, err.Error())
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
		w.Header().Add("Server", AppName)
		fileServer.ServeHTTP(w, r)
	})
}

func main() {
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), WebAppFileServer(http.Dir(rootDir)))
	if err != nil {
		log.Error(err.Error())
	} else {
		log.Infof("webapp-go started on port %s", port)
	}
}
