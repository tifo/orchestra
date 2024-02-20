package commands

import (
	"fmt"
	"math"
	"os"
	"path"
	"path/filepath"
	"strings"

	log "github.com/cihub/seelog"
	"github.com/urfave/cli/v2"

	"github.com/tifo/orchestra/config"
	"github.com/tifo/orchestra/services"
)

// niceness used for subprocesses
// https://en.wikipedia.org/wiki/Nice_(Unix)
const niceness = "1"

// This is temporary, very very alpha and may change soon
func FilterServices(c *cli.Context) map[string]*services.Service {
	excludeMode := 0
	args := c.Args().Slice()
	included := make(map[string]bool)

	for _, s := range args {
		name := s
		if strings.HasPrefix(s, "~") {
			name = strings.Replace(s, "~", "", 1)
		}
		// Remove trailing slash to help with file autocomplete
		name = strings.TrimRight(name, "/")
		// Alias `.` to the current service if it exists
		if name == "." {
			cwd, _ := os.Getwd()
			name, _ = filepath.Rel(services.ProjectPath, cwd)
		}
		// Check if arg match a service or a stack
		if _, ok := services.Registry[name]; ok {
			if strings.HasPrefix(s, "~") {
				excludeMode += 1
				delete(services.Registry, name)
			} else {
				excludeMode -= 1
				included[name] = true
			}
		} else if stack, ok := services.StackRegistry[name]; ok {
			if strings.HasPrefix(s, "~") {
				excludeMode += 1
				for _, svc := range stack {
					delete(services.Registry, svc.Name)
				}
				delete(services.StackRegistry, name)
			} else {
				excludeMode -= 1
				for _, svc := range stack {
					included[svc.Name] = true
				}
			}
		} else {
			_ = log.Errorf("Service or stack %s not found", name)
			return nil
		}
	}
	if math.Abs(float64(excludeMode)) != float64(len(args)) {
		_ = log.Critical("You can't exclude and include services at the same time")
		os.Exit(1)
	}
	if excludeMode < 0 {
		for name := range services.Registry {
			if !included[name] {
				delete(services.Registry, name)
			}
		}
	}
	return services.Registry
}

func BeforeAfterWrapper(f func(c *cli.Context) error) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		err := config.GetBeforeFunc()(c)
		if err != nil {
			appendError(err)
		}
		_ = f(c)
		err = config.GetAfterFunc()(c)
		if err != nil {
			appendError(err)
		}
		return nil
	}
}

func ServicesBashComplete(c *cli.Context) {
	confVal := config.FindProjectConfig(c.String("config"))
	config.ConfigPath, _ = filepath.Abs(confVal)
	services.ProjectPath, _ = path.Split(config.ConfigPath)
	config.ParseGlobalConfig()
	services.Init()
	for stack := range services.StackRegistry {
		fmt.Println(stack)
	}
	for name := range services.Registry {
		fmt.Println(name)
		fmt.Println("~" + name)
	}
}

// GetEnvForService returns all the environment variables for a given service
// including the ones specified in the global config
func GetEnvForService(c *cli.Context, service *services.Service) []string {
	return append(service.Env, config.GetEnvForCommand(c)...)
}

type workerPool chan struct{}

func (p workerPool) Drain() {
	for i := 0; i < cap(p); i++ {
		p <- struct{}{}
	}
}

func (p workerPool) Do(impl func()) {
	p <- struct{}{}
	go func() {
		defer func() { <-p }()
		impl()
	}()
}
