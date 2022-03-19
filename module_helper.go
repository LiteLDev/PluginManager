package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ModuleVersionInfo struct {
	Version string    // version string
	Time    time.Time // commit time
}

func escapeModuleUrl(url string) (ret string) {
	for _, v := range url {
		if v > 'A' && v < 'Z' {
			ret += "!"
			ret += string(v + 32)
		} else {
			ret += string(v)
		}
	}
	return
}

func getDownloadUrl(modulePath, goproxyUrl, versionStr string) string {
	return fmt.Sprintf("%s/%s/@v/%s.zip", goproxyUrl, escapeModuleUrl(modulePath), versionStr)
}

func getModuleVersionInfo(modulePath, goproxyUrl, versionStr string) (ret ModuleVersionInfo, err error) {
	url := fmt.Sprintf("%s/%s/@v/%s.info", goproxyUrl, escapeModuleUrl(modulePath), versionStr)
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = fmt.Errorf("get module version info failed, status code: %d", resp.StatusCode)
		return
	}
	err = json.NewDecoder(resp.Body).Decode(&ret)
	return
}

func getModuleVersionLatest(modulePath string, goproxyUrl string) (ver ModuleVersionInfo, err error) {
	realUrl := fmt.Sprintf("%s/%s/@latest", goproxyUrl, escapeModuleUrl(modulePath))
	resp, err := http.Get(realUrl)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = fmt.Errorf("get module version list failed, status code: %d", resp.StatusCode)
		return
	}
	err = json.NewDecoder(resp.Body).Decode(&ver)
	return
}

func getModuleVersionList(modulePath string, goproxyUrl string) (list []ModuleVersionInfo, err error) {
	realUrl := fmt.Sprintf("%s/%s/@v/list", goproxyUrl, escapeModuleUrl(modulePath))
	resp, err := http.Get(realUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("get module version list failed, status code: %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	versions := strings.Split(string(data), "\n")
	for _, v := range versions {
		if v != "" {
			ver, err := getModuleVersionInfo(modulePath, goproxyUrl, v)
			if err != nil {
				return nil, err
			}
			list = append(list, ver)
		}
	}
	return
}
func UnzipModule(src, dest string) (string, error) {
	var path string
	r, err := zip.OpenReader(src)
	if err != nil {
		return "", err
	}
	defer r.Close()
	hasManifest := false

	for _, f := range r.File {
		var filename string
		path, filename = filepath.Split(f.Name)
		if !f.FileInfo().IsDir() && filename == "manifest.json" {
			hasManifest = true
			break
		}
	}
	if !hasManifest {
		return "", fmt.Errorf("no manifest.json found in zip file")
	}
	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return "", err
		}
		defer rc.Close()

		path := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			var dir string
			if lastIndex := strings.LastIndex(path, string(os.PathSeparator)); lastIndex > -1 {
				dir = path[:lastIndex]
			}

			err = os.MkdirAll(dir, f.Mode())
			if err != nil {
				log.Fatal(err)
				return "", err
			}
			f, err := os.OpenFile(
				path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return "", err
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				return "", err
			}
		}
	}
	return path, nil
}
