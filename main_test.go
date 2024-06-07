package main

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const RootDir = "C:\\Users\\yangj\\Downloads\\datax-v1.0.4-hashdata\\datax"
const dstDir = "C:\\Users\\yangj\\Downloads\\datax-v1.0.4-hashdata\\data1"

func TestFilePath(t *testing.T) {
	err := filepath.WalkDir(RootDir, func(path string, d fs.DirEntry, err error) error {
		relativePath := strings.Replace(path, RootDir, "", 1)
		if len(relativePath) == 0 {
			return nil
		}

		dstPath := filepath.Join(dstDir, relativePath)
		if d.IsDir() {
			if exists, err := fileExists(dstPath); err != nil {
				log.Fatalln(err)
				return err
			} else if !exists {
				if err := os.Mkdir(dstPath, os.ModeDir); err != nil {
					log.Fatalln(err)
					return err
				}
			}
		} else {
			return copyFile(path, dstPath)
		}

		return nil
	})
	log.Fatalln(err)
}
