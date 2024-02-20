package services

import (
	"fmt"
	"go/build"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"

	log "github.com/cihub/seelog"
	"gopkg.in/yaml.v3"

	"github.com/tifo/orchestra/config"
)

var (
	// Internal Service Registry
	Registry      map[string]*Service
	StackRegistry map[string][]*Service

	// Path variables
	OrchestraServicePath string
	ProjectPath          string

	// Other internal variables
	MaxServiceNameLength int
	colors               = []string{"g", "b", "c", "m", "y", "w"}
)

func init() {
	Registry = make(map[string]*Service)
	StackRegistry = make(map[string][]*Service)
}

// Init initializes the OrchestraServicePath to the workingdir/.orchestra path
// and starts the service discovery
func Init() {
	DiscoverServices()
}

func Sort(r map[string]*Service) SortableRegistry {
	sr := make(SortableRegistry, 0)
	for _, v := range r {
		sr = append(sr, v)
	}
	sort.Sort(sr)
	return sr
}

type SortableRegistry []*Service

func (s SortableRegistry) Len() int {
	return len(s)
}

func (s SortableRegistry) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s SortableRegistry) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}

// Service encapsulates all the information needed for a service
type Service struct {
	Name        string
	Stack       string
	Description string
	Path        string
	Color       string

	// Path
	OrchestraPath string
	LogFilePath   string
	PidFilePath   string
	BinPath       string

	// Process, Service and Package information
	FileInfo    fs.DirEntry
	PackageInfo *build.Package
	Process     *os.Process
	Env         []string
	Ports       string
}

func (s *Service) IsRunning() bool {
	if _, err := os.Stat(s.PidFilePath); err == nil {
		bytes, _ := os.ReadFile(s.PidFilePath)
		pid, _ := strconv.Atoi(string(bytes))
		proc, procErr := os.FindProcess(pid)
		if procErr == nil {
			sigError := proc.Signal(syscall.Signal(0))
			if sigError == nil {
				s.Process = proc
				return true
			} else {
				os.Remove(s.PidFilePath)
			}
		}
	} else {
		os.Remove(s.PidFilePath)
	}
	return false
}

func discoverStack(stack string) {
	fd, err := os.ReadDir(path.Join(ProjectPath, stack))
	if err != nil {
		_ = log.Errorf("Error registering stack %s")
		_ = log.Error(err.Error())
		return
	}
	if stack != "" {
		StackRegistry[stack] = make([]*Service, 0)
	}
	for _, item := range fd {
		serviceName := item.Name()
		if stack != "" {
			serviceName = path.Join(stack, serviceName)
		}
		if item.IsDir() && !strings.HasPrefix(serviceName, ".") {
			serviceConfigPath := path.Join(ProjectPath, serviceName, "service.yml")
			if _, err := os.Stat(serviceConfigPath); err == nil {
				// Check for service.yml and try to import the package
				pkg, err := build.ImportDir(path.Join(ProjectPath, serviceName), build.FindOnly)
				if err != nil {
					_ = log.Errorf("Error registering %s", item.Name())
					_ = log.Error(err.Error())
					continue
				}

				service := &Service{
					Name:          serviceName,
					Stack:         stack,
					Description:   "",
					FileInfo:      item,
					PackageInfo:   pkg,
					OrchestraPath: OrchestraServicePath,
					LogFilePath:   path.Join(OrchestraServicePath, strings.Replace(serviceName, "/", "_", -1)+".log"),
					PidFilePath:   path.Join(OrchestraServicePath, strings.Replace(serviceName, "/", "_", -1)+".pid"),
					Color:         colors[len(Registry)%len(colors)],
					Path:          path.Join(ProjectPath, serviceName),
				}

				// Parse env variable in configuration
				var serviceConfig struct {
					Env map[string]string `yaml:"env,omitempty"`
				}
				b, err := os.ReadFile(serviceConfigPath)
				if err != nil {
					_ = log.Criticalf(err.Error())
					os.Exit(1)
				}
				_ = yaml.Unmarshal(b, &serviceConfig)
				for k, v := range serviceConfig.Env {
					service.Env = append(service.Env, fmt.Sprintf("%s=%s", k, v))
				}

				// Because I like nice logging
				if len(serviceName) > MaxServiceNameLength {
					MaxServiceNameLength = len(serviceName)
				}

				if binPath := os.Getenv("GOBIN"); binPath != "" {
					service.BinPath = path.Join(binPath, path.Base(serviceName))
				} else {
					service.BinPath = path.Join(os.Getenv("GOPATH"), "bin", path.Base(serviceName))
				}

				// Add the service to the registry
				Registry[serviceName] = service
				if stack != "" {
					StackRegistry[stack] = append(StackRegistry[stack], service)
				}
				// When registering, we take care, on every run, to check
				// if the process is still alive.
				service.IsRunning()
			}
		}
	}
}

// DiscoverServices walks into the project path and looks in every subdirectory
// for the service.yml file. For every service it registers it after trying
// to import the package using Go's build.Import package
func DiscoverServices() {
	for _, stack := range config.GetStacks() {
		if stack == "" || stack == "." {
			discoverStack("")
		} else {
			if !filepath.IsLocal(stack) {
				_ = log.Errorf("Can't register stack %s, path is not local", stack)
				continue
			}
			discoverStack(stack)
		}
	}
}
