package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	log "github.com/cihub/seelog"
	"github.com/urfave/cli/v2"

	"github.com/tifo/orchestra/commands"
	"github.com/tifo/orchestra/config"
	"github.com/tifo/orchestra/services"
)

var app *cli.App

const defaultConfigFile = "orchestra.yml"

func findConfigFile() string {
	dir, err := os.Getwd()
	if err != nil {
		return defaultConfigFile
	}
	for {
		file := filepath.Join(dir, defaultConfigFile)
		_, err := os.Stat(file)
		if err == nil {
			return file
		}

		if dir == "/" {
			break
		}

		dir = filepath.Dir(dir)
	}
	return defaultConfigFile
}

func main() {
	defer log.Flush()
	app = cli.NewApp()
	app.Name = "Orchestra"
	app.Usage = "Orchestrate Go Services (Tifo)"
	app.EnableBashCompletion = true
	app.Commands = []*cli.Command{
		commands.BuildCommand,
		commands.ExportCommand,
		commands.InstallCommand,
		commands.LogsCommand,
		commands.PsCommand,
		commands.RestartCommand,
		commands.StartCommand,
		commands.StopCommand,
		commands.TestCommand,
	}
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "config, c",
			Usage:   "Specify a different config file to use (default: \"orchestra.yml\"",
			EnvVars: []string{"ORCHESTRA_CONFIG"},
		},
	}
	// init checks for an existing orchestra.yml in the current working directory
	// and creates a new .orchestra directory (if doesn't exist)
	app.Before = func(c *cli.Context) error {
		confVal := c.String("config")
		if confVal == "" {
			confVal = findConfigFile()
		}

		config.ConfigPath, _ = filepath.Abs(confVal)
		if _, err := os.Stat(config.ConfigPath); os.IsNotExist(err) {
			fmt.Printf("No %s found. Have you specified the right directory?\n", confVal)
			os.Exit(1)
		}
		services.ProjectPath, _ = path.Split(config.ConfigPath)
		services.OrchestraServicePath = services.ProjectPath + ".orchestra"

		if err := os.Mkdir(services.OrchestraServicePath, 0766); err != nil && os.IsNotExist(err) {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		config.ParseGlobalConfig()
		services.Init()
		return nil
	}
	app.Version = "0.5.0"
	app.Run(os.Args)
	if commands.HasErrors() {
		os.Exit(1)
	}
}
