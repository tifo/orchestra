package commands

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"
	"sync"

	"github.com/urfave/cli"
	"github.com/wsxiaoys/terminal"

	"github.com/tifo/orchestra/services"
)

var PsCommand = &cli.Command{
	Name:   "ps",
	Usage:  "Outputs the status of all services",
	Action: BeforeAfterWrapper(PsAction),
}

// PsAction checks the status for every service and output
func PsAction(c *cli.Context) error {
	svcs := services.Sort(FilterServices(c))

	var wg sync.WaitGroup
	for _, svc := range svcs {
		wg.Add(1)
		go func(s *services.Service) {
			s.Ports = getPorts(s)
			wg.Done()
		}(svc)
	}
	wg.Wait()

	for _, service := range svcs {
		spacing := strings.Repeat(" ", services.MaxServiceNameLength+2-len(service.Name))
		if service.Process != nil {
			terminal.Stdout.Colorf("@{g}%s", service.Name).Reset().Colorf("%s|", spacing).Print(" running ").Colorf("  %d  %s\n", service.Process.Pid, service.Ports)
		} else {
			terminal.Stdout.Colorf("@{r}%s", service.Name).Reset().Colorf("%s|", spacing).Reset().Print(" aborted\n")
		}
	}
	return nil
}

func getPorts(service *services.Service) string {
	if service.Process == nil {
		return ""
	}

	re := regexp.MustCompile("LISTEN")
	cmd := exec.Command("lsof", "-P", "-p", fmt.Sprintf("%d", service.Process.Pid))
	output := bytes.NewBuffer([]byte{})
	cmd.Stdout = output
	cmd.Stderr = output
	err := cmd.Run()
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	lsofOutput := ""
	for {
		s, err := output.ReadString('\n')
		if err == io.EOF {
			break
		}
		matched := re.MatchString(s)
		if matched {
			fields := strings.Fields(s)
			lsofOutput += fmt.Sprintf("%s/%s ", fields[8], strings.ToLower(fields[7]))
		}
	}
	return lsofOutput
}
