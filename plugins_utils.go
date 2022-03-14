package main

import (
	"io/ioutil"
	"path/filepath"
	"strings"
)

type PluginInfo struct {
	Name    string
	Path    string
	Version string
}

type PluginInfos []PluginInfo

func (w *PluginInfos) isPluginFolder(path string) {
	subPaths := []string{}
	isPluginPath := false
	dir, _ := ioutil.ReadDir(path)
	for _, file := range dir {
		if file.IsDir() {
			subPaths = append(subPaths, filepath.Join(path, file.Name()))
		} else {
			if file.Name() == "manifest.json" {
				Plugin := PluginInfo{}
				_, file := filepath.Split(path)
				index := strings.LastIndex(file, "@")
				if index != -1 {
					rel, err := filepath.Rel("./plugins/PluginManager/pkg", path)
					if err != nil {
						return
					}
					rel, _ = filepath.Split(rel)
					Plugin.Name = filepath.Join(rel, file[:index])
					Plugin.Version = file[index+1:]
					Plugin.Path = path
					*w = append(*w, Plugin)
					isPluginPath = true
				}
			}
		}
	}
	if !isPluginPath {
		for _, subPath := range subPaths {
			w.isPluginFolder(subPath)
		}
	}
	return
}

func getLocalPackages() (PluginInfos, error) {
	dir, err := ioutil.ReadDir("./plugins/PluginManager/pkg")
	if err != nil {
		return nil, err
	}
	plugins := PluginInfos{}
	for _, file := range dir {
		if file.IsDir() {
			plugins.isPluginFolder("./plugins/PluginManager/pkg/" + file.Name() + "/")
		}
	}
	return plugins, err
}
