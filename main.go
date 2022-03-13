package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"path/filepath"
)

func main() {
	app := &cli.App{
		Name:  "BDSLiteLoader Plugin Manager",
		Usage: "BDSLiteLoader Plugin Manager that helps you download third-party plugins",
		Commands: []*cli.Command{
			{
				Name:  "list",
				Usage: "get version list of specified url",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "url",
						Aliases:     []string{"u"},
						Usage:       "specify plugin url",
						Value:       "github.com/WangYneos/GoModuleTest",
						DefaultText: "github.com/WangYneos/GoModuleTest",
					},
				},
				Action: func(c *cli.Context) error {
					versions, err := getModuleVersionList(c.String("url"), "https://goproxy.cn")
					if err != nil {
						return err
					}
					for _, v := range versions {
						log.Printf("%s\t%s\n", v.Version, v.Time)
					}
					return nil
				},
			},
			{
				Name:  "local",
				Usage: "get local plugin list",
				Action: func(c *cli.Context) error {

					plugins, _ := getLocalPackages()
					for _, p := range plugins {
						log.Println("Name\t", p.Name)
						log.Println("Version\t", p.Version)
						log.Println("Path\t", p.Path)
						log.Print("\n")
					}
					return nil
				},
			},
			{
				Name:  "download",
				Usage: "download specified version of specified url",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "url",
						Aliases:     []string{"u"},
						Usage:       "specify plugin url",
						Value:       "github.com/WangYneos/GoModuleTest",
						DefaultText: "github.com/WangYneos/GoModuleTest",
					},
					&cli.StringFlag{
						Name:        "version",
						Aliases:     []string{"v"},
						Usage:       "specify plugin version",
						Value:       "@latest",
						DefaultText: "@latest",
					},
				},
				Action: func(c *cli.Context) error {
					var err error

					os.MkdirAll("plugins/PluginManager", os.ModePerm)
					os.MkdirAll("plugins/PluginManager/pkg", os.ModePerm)

					version := ModuleVersionInfo{}
					if c.String("version") == "@latest" {
						version, err = getModuleVersionLatest(c.String("url"), "https://goproxy.cn")
						if err != nil {
							log.Fatalf("failed to get latest version: %v", err)
							return nil
						}
					} else {
						version, err = getModuleVersionInfo(c.String("url"), "https://goproxy.cn", c.String("version"))
						if err != nil {
							log.Fatalf("failed to get version info for %s: %v", c.String("version"), err)
							return nil
						}
					}
					_, file := filepath.Split(c.String("url"))

					log.Printf("downloading %s@%s [%v]", version.Version, c.String("url"), version.Time)
					downloadUrl := getDownloadUrl(c.String("url"), "https://goproxy.cn", version.Version)
					fileName := fmt.Sprintf("plugins/PluginManager/pkg/%s-%s.zip", file, version.Version)
					DownloadFile(fileName, downloadUrl)
					UnzipFiles(fileName, "plugins/PluginManager/pkg")
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}