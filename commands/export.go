package commands

import (
	"fmt"

	"github.com/urfave/cli"
	"github.com/wsxiaoys/terminal"

	"github.com/tifo/orchestra/config"
)

var ExportCommand = &cli.Command{
	Name:         "export",
	Usage:        "Export those *#%&! env vars ",
	Action:       BeforeAfterWrapper(ExportAction),
	BashComplete: ServicesBashComplete,
}

func ExportAction(c *cli.Context) error {
	for key, value := range config.GetBaseEnvVars() {
		terminal.Stdout.Print(fmt.Sprintf("export %s=%s\n", key, value))
	}
	return nil
}
