package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type PluginManifest struct {
	Author      string
	Description string
	License     string

	Install   string
	Uninstall string
}

type PluginInfo struct {
	Name    string
	Path    string
	Version string

	Manifest PluginManifest
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
				_, mainName := filepath.Split(path)
				index := strings.LastIndex(mainName, "@")
				if index != -1 {
					rel, err := filepath.Rel("./plugins/PluginManager/pkg", path)
					if err != nil {
						return
					}
					rel, _ = filepath.Split(rel)
					Plugin.Name = filepath.Join(rel, mainName[:index])
					Plugin.Version = mainName[index+1:]
					Plugin.Path = path

					data, err := os.Open(filepath.Join(path, "manifest.json"))
					if err != nil {
						return
					}
					var manifest PluginManifest
					err = json.NewDecoder(data).Decode(&manifest)
					if err != nil {
						return
					}

					Plugin.Manifest = manifest

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

func installPlugin(path string) error {
	return nil
}

func uninstallPlugin(path string) error {
	return nil
}
