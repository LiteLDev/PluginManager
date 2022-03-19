package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/mod/modfile"
	"io"
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

	Manifest   *PluginManifest
	ModuleInfo *modfile.File
}

type PluginInfos []PluginInfo

func getPluginInfo(path string) (Plugin PluginInfo, err error) {
	_, mainName := filepath.Split(path)
	index := strings.LastIndex(mainName, "@")
	if index == -1 {
		err = fmt.Errorf("invalid plugin name: %s", mainName)
		return
	}
	rel, err := filepath.Rel(filepath.Join(PluginManagerRoot, "pkg"), path)
	if err != nil {
		return
	}
	rel, _ = filepath.Split(rel)
	Plugin.Name = filepath.Join(rel, mainName[:index])
	Plugin.Version = mainName[index+1:]
	Plugin.Path = path

	manifestFileReader, err := os.Open(filepath.Join(path, "manifest.json"))
	if err != nil {
		return
	}
	var manifest PluginManifest
	err = json.NewDecoder(manifestFileReader).Decode(&manifest)
	if err != nil {
		return
	}

	Plugin.Manifest = &manifest
	modFileReader, err := os.Open(filepath.Join(path, "go.mod"))
	if err != nil {
		return
	}
	modFileData, err := io.ReadAll(modFileReader)
	if err != nil {
		return
	}
	modFile, err := modfile.Parse("go.mod", modFileData, nil)
	Plugin.ModuleInfo = modFile
	return
}

func (w *PluginInfos) isPluginFolder(path string) {
	isPluginPath := false

	_, err1 := os.Stat(filepath.Join(path, "manifest.json"))
	_, err2 := os.Stat(filepath.Join(path, "go.mod"))
	maybePluginPath := err1 == nil && err2 == nil

	if maybePluginPath {
		Plugin, err := getPluginInfo(path)
		if err == nil {
			*w = append(*w, Plugin)
			isPluginPath = true
		}
	}

	if !isPluginPath {
		dir, err := ioutil.ReadDir(path)
		if err != nil {
			return
		}
		for _, file := range dir {
			if file.IsDir() {
				w.isPluginFolder(filepath.Join(path, file.Name()))
			}
		}
	}
	return
}

func getLocalPackages() (PluginInfos, error) {
	dir, err := ioutil.ReadDir(filepath.Join(PluginManagerRoot, "pkg"))
	if err != nil {
		return nil, err
	}
	plugins := PluginInfos{}
	for _, file := range dir {
		if file.IsDir() {
			plugins.isPluginFolder(filepath.Join(PluginManagerRoot, "pkg", file.Name()))
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
