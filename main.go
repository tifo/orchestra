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

func main() {
	defer log.Flush()
	app = cli.NewApp()
	app.Name = "Orchestra"
	app.Usage = "Orchestrate Go Services (Tifo)"
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
	app.EnableBashCompletion = true
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "config, c",
			Usage:   "Specify a different config file to use (default: \"orchestra.yml\")",
			EnvVars: []string{"ORCHESTRA_CONFIG"},
		},
	}
	// init checks for an existing orchestra.yml in the current working directory
	// and creates a new .orchestra directory (if doesn't exist)
	app.Before = func(c *cli.Context) error {
		confVal := c.String("config")
		confVal = config.FindProjectConfig(confVal)
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
	app.Version = "0.5.2"
	app.Run(os.Args)
	if commands.HasErrors() {
		os.Exit(1)
	}
}
