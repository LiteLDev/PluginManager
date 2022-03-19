package main

import (
	"encoding/json"
	"fmt"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"path/filepath"
)

type Config struct {
	Source string `json:"source"`
}

var GlobalConfig Config

func init() {
	log.SetFlags(log.Ltime | log.Lshortfile)
	initDirs()

	configFilePath := filepath.Join(PluginManagerRoot, "PluginManager.json")

	_, err := os.Stat(configFilePath)
	if err != nil {
		GlobalConfig.Source = DefaultDownloadSource
		configData, err := json.MarshalIndent(GlobalConfig, "", "  ")
		if err != nil {
			log.Fatalln("WriteConfig", err)
		}
		file, err := os.Create(configFilePath)
		if err != nil {
			log.Fatalln("WriteConfig", err)
		}
		defer file.Close()
		_, err = file.Write(configData)
		if err != nil {
			log.Fatalln("WriteConfig", err)
		}

	} else {
		data, err := os.Open(configFilePath)
		if err != nil {
			log.Fatalln("LoadConfig", err)
		}
		err = json.NewDecoder(data).Decode(&GlobalConfig)
		if err != nil {
			log.Fatalln("LoadConfig", err)
		}
	}
}

func main() {
	app := &cli.App{
		Name:  "BDSLiteLoader Plugin Manager",
		Usage: "BDSLiteLoader Plugin Manager that helps you download third-party plugins",
		Commands: []*cli.Command{

			{
				Name:  "test",
				Usage: "test",
				Action: func(c *cli.Context) error {
					vm := newVmInstance()
					_, err := vm.Run(`
filesystem.Mkdir("./test");
filesystem.Write("./test/test.txt", "test");
console.log(filesystem.Exists("./test/test.txt"));
console.log(filesystem.Read("./test/test.txt"));
filesystem.Append("./test/test.txt", "test2");
console.log(filesystem.Read("./test/test.txt"));
filesystem.Delete("./test");
console.log(system.Cmd("cmd", "/C", "pause"));
filesystem.Create("./test.txt");
console.log(system.Cmd("cmd", "/C", "del", "test.txt"));
`)
					return err
				},
			},

			{
				Name:  "list",
				Usage: "list packages or versions",
				Subcommands: []*cli.Command{

					{
						Name:    "remote",
						Aliases: []string{"r"},
						Usage:   "list remote versions",
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
							versions, err := getModuleVersionList(c.String("url"), GlobalConfig.Source)
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
						Name:    "local",
						Aliases: []string{"l"},
						Usage:   "get local plugin list",
						Action: func(c *cli.Context) error {

							plugins, err := getLocalPackages()
							if err != nil {
								return err
							}
							for _, p := range plugins {
								log.Println("Name\t", p.Name)
								log.Println("Version\t", p.Version)
								log.Println("Path\t", p.Path)
								log.Printf("Manifest\t%+v", p.Manifest)
								log.Print("\n")

							}
							return nil
						},
					},
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

					version := ModuleVersionInfo{}
					if c.String("version") == "@latest" {
						version, err = getModuleVersionLatest(c.String("url"), GlobalConfig.Source)
						if err != nil {
							log.Fatalf("failed to get latest version: %v", err)
							return nil
						}
					} else {
						version, err = getModuleVersionInfo(c.String("url"), GlobalConfig.Source, c.String("version"))
						if err != nil {
							log.Fatalf("failed to get version info for %s: %v", c.String("version"), err)
							return nil
						}
					}
					_, file := filepath.Split(c.String("url"))

					log.Printf("downloading %s@%s [%v]", version.Version, c.String("url"), version.Time)
					downloadUrl := getDownloadUrl(c.String("url"), GlobalConfig.Source, version.Version)
					fileName := filepath.Join(PluginManagerRoot, "cache", fmt.Sprintf("%s-%s.zip", file, version.Version))
					DownloadFile(fileName, downloadUrl)
					err = UnzipFiles(fileName, filepath.Join(PluginManagerRoot, "pkg"))
					return err
				},
			},
			{
				Name:  "remove",
				Usage: "remove specified plugin",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "name",
						Aliases:     []string{"n"},
						Usage:       "specify plugin name",
						Value:       "github.com/WangYneos/GoModuleTest",
						DefaultText: "github.com/WangYneos/GoModuleTest",
					},
					&cli.StringFlag{
						Name:        "version",
						Aliases:     []string{"v"},
						Usage:       "specify plugin version",
						Value:       "@all",
						DefaultText: "@all",
					},
				},
				Action: func(c *cli.Context) error {
					packages, err := getLocalPackages()
					if err != nil {
						return err
					}
					for _, v := range packages {
						if v.Name == c.String("name") {
							if c.String("version") == "@all" || v.Version == c.String("version") {
								os.RemoveAll(v.Path)
							}
						}
					}
					err = removeEmptyFolders(filepath.Join(PluginManagerRoot, "pkg"))
					return err
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
