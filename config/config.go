package config

import (
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"strings"

	log "github.com/cihub/seelog"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
)

var orchestra *Config
var ConfigPath string
var globalEnvs []string

type ContextConfig struct {
	Env    map[string]string `yaml:"env,omitempty"`
	Before []string          `yaml:"before,omitempty"`
	After  []string          `yaml:"after,omitempty"`
}

type Config struct {
	// Global Configuration
	Env    map[string]string `yaml:"env,omitempty"`
	Before []string          `yaml:"before,omitempty"`
	After  []string          `yaml:"after,omitempty"`
	GoRun  bool              `yaml:"gorun,omitempty"`

	// Stacks configuration (includes subfolders)
	Stacks []string `yaml:"stacks,omitempty"`

	// Configuration for Commands
	Build   ContextConfig `yaml:"build,omitempty"`
	Export  ContextConfig `yaml:"export,omitempty"`
	Install ContextConfig `yaml:"install,omitempty"`
	Logs    ContextConfig `yaml:"logs,omitempty"`
	Ps      ContextConfig `yaml:"ps,omitempty"`
	Restart ContextConfig `yaml:"restart,omitempty"`
	Start   ContextConfig `yaml:"start,omitempty"`
	Stop    ContextConfig `yaml:"stop,omitempty"`
	Test    ContextConfig `yaml:"test,omitempty"`
}

func GetBaseEnvVars() map[string]string {
	return orchestra.Env
}

func UseGoRun() bool {
	return orchestra.GoRun
}

func GetStacks() []string {
	if len(orchestra.Stacks) == 0 {
		return []string{""}
	}
	return orchestra.Stacks
}

func ParseGlobalConfig() {
	orchestra = &Config{}
	b, err := os.ReadFile(ConfigPath)
	if err != nil {
		_ = log.Criticalf(err.Error())
		os.Exit(1)
	}
	_ = yaml.Unmarshal(b, &orchestra)

	globalEnvs = os.Environ()
	for k, v := range orchestra.Env {
		globalEnvs = append([]string{fmt.Sprintf("%s=%s", k, v)}, globalEnvs...)
	}
}

func GetEnvForCommand(c *cli.Context) []string {
	envs := globalEnvs
	for k, v := range getConfigFieldByName(c.Command.Name).Env {
		envs = append(envs, fmt.Sprintf("%s=%s", k, v))
	}
	return envs
}

func runCommands(c *cli.Context, cmds []string) error {
	for _, command := range cmds {
		cmdLine := strings.Split(command, " ")
		cmd := exec.Command(cmdLine[0], cmdLine[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = GetEnvForCommand(c)
		err := cmd.Start()
		if err != nil {
			return err
		}
		_ = cmd.Wait()
		if !cmd.ProcessState.Success() {
			return fmt.Errorf("Command %s exited with error", cmdLine[0])
		}
	}
	return nil
}

func GetBeforeFunc() func(c *cli.Context) error {
	return func(c *cli.Context) error {
		err := runCommands(c, orchestra.Before)
		if err != nil {
			return err
		}
		err = runCommands(c, getConfigFieldByName(c.Command.Name).Before)
		if err != nil {
			return err
		}
		return nil
	}
}

func GetAfterFunc() func(c *cli.Context) error {
	return func(c *cli.Context) error {
		err := runCommands(c, orchestra.After)
		if err != nil {
			return err
		}
		err = runCommands(c, getConfigFieldByName(c.Command.Name).After)
		if err != nil {
			return err
		}
		return nil
	}
}

func getConfigFieldByName(name string) ContextConfig {
	initial := strings.Split(name, "")[0]
	value := reflect.ValueOf(orchestra)
	f := reflect.Indirect(value).FieldByName(strings.Replace(name, initial, strings.ToUpper(initial), 1))
	return f.Interface().(ContextConfig)
}
