package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

func initDirs() {
	os.MkdirAll("./plugins/PluginManager", os.ModePerm)
	os.MkdirAll("./plugins/PluginManager/pkg", os.ModePerm)
	os.MkdirAll("./plugins/PluginManager/cache", os.ModePerm)
}
func findEmptyFolder(path string) (err error) {
	dir, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}
	if len(dir) == 0 {
		os.Remove(path)
		return nil
	}
	for _, file := range dir {
		if file.IsDir() {
			err := findEmptyFolder(filepath.Join(path, file.Name()))
			if err != nil {
				return err
			}
		}
	}
	dir, err = ioutil.ReadDir(path)
	if err != nil {
		return err
	}
	if len(dir) == 0 {
		os.Remove(path)
		return nil
	}
	return nil
}

func removeEmptyFolders(path string) error {
	err := findEmptyFolder(path)
	if err != nil {
		return err
	}
	return nil
}
