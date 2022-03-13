package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type ModuleVersionInfo struct {
	Version string    // version string
	Time    time.Time // commit time
}

func parseUrl(url string) (ret string) {
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
	return fmt.Sprintf("%s/%s/@v/%s.zip", goproxyUrl, parseUrl(modulePath), versionStr)
}

func getModuleVersionInfo(modulePath, goproxyUrl, versionStr string) (ret ModuleVersionInfo, err error) {
	url := fmt.Sprintf("%s/%s/@v/%s.info", goproxyUrl, parseUrl(modulePath), versionStr)
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
	realUrl := fmt.Sprintf("%s/%s/@latest", goproxyUrl, parseUrl(modulePath))
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
	realUrl := fmt.Sprintf("%s/%s/@v/list", goproxyUrl, parseUrl(modulePath))
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
