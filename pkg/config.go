package pkg

import (
	"embed"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"sync"

	"github.com/go-logr/logr"
	"gopkg.in/yaml.v3"
)

//go:embed templates/*
var langTemplate embed.FS

const (
	defaultConfigFile = "crzy.yaml"
	defaultLanguage   = "golang"
	golangLanguage    = "golang"
)

var (
	errUnsupportedLang   = errors.New("unsupportedlang")
	errLoadingConfigFile = errors.New("loadingfile")
)

type config struct {
	*sync.Mutex
	Main     mainStruct
	Trigger  triggerStruct
	Deploy   deployStruct
	Release  releaseStruct
	Notifier notifierStruct
}

type mainStruct struct {
	Repository string
	Head       string
	Color      bool
	ApiPort    int `yaml:"api_port"`
	ProxyPort  int `yaml:"proxy_port"`
}

type triggerStruct struct {
	Version versionStruct
}

type deployStruct struct {
	Artifact artifactStruct
	Install  execStruct
	Test     execStruct
	PreBuild execStruct `yaml:"pre_build"`
	Build    execStruct
}

type artifactStruct struct {
	Filename  string
	Directory string
	Extension string
}

type versionStruct struct {
	Command string
	Args    []string
	WorkDir string
	Envs    []envVar
}

type portRangeStruct struct {
	Min int `yaml:"min"`
	Max int `yaml:"max"`
}

type releaseStruct struct {
	PortRange portRangeStruct `yaml:"port_range"`
	Run       execStruct
}

type execStruct struct {
	log     logr.Logger
	Command string
	Args    []string
	WorkDir string
	Envs    []envVar
	Output  string
}

func execCmdBackground(dir string, envs map[string]string, name string, arg ...string) (*os.Process, error) {
	c := exec.Command(name, arg...)
	c.Dir = dir
	if envs != nil && len(envs) > 0 {
		vars := []string{}
		for k, v := range envs {
			vars = append(vars, fmt.Sprintf("%s=%s", k, v))
		}
		c.Env = vars
	}
	err := c.Start()
	return c.Process, err
}

func (e *execStruct) run(workspace string, envs map[string]string) (*envVar, error) {
	dir := path.Join(workspace, e.WorkDir)
	command, err := replaceEnvs(e.Command, envs)
	if err != nil {
		return nil, err
	}
	args := []string{}
	full := command
	for _, arg := range e.Args {
		arg, err = replaceEnvs(arg, envs)
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
		full = fmt.Sprintf("%s %s", full, arg)
	}
	for key, value := range envs {
		value, err = replaceEnvs(value, envs)
		if err != nil {
			return nil, err
		}
		envs[key] = value
	}
	e.log.Info(full)
	output, err := execCmd(dir, envs, command, args...)
	if err != nil {
		return nil, err
	}
	results := strings.Split(string(output), "\n")
	for _, v := range results {
		e.log.Info(v)
	}
	if e.Output != "" {
		return &envVar{Name: e.Output, Value: results[0]}, nil
	}
	return nil, nil
}

func (e *execStruct) runBackground(workspace string, envs map[string]string) (*os.Process, error) {
	dir := path.Join(workspace, e.WorkDir)
	command, err := replaceEnvs(e.Command, envs)
	if err != nil {
		return nil, err
	}
	args := []string{}
	full := command
	for _, arg := range e.Args {
		arg, err = replaceEnvs(arg, envs)
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
		full = fmt.Sprintf("%s %s", full, arg)
	}
	for key, value := range envs {
		value, err = replaceEnvs(value, envs)
		if err != nil {
			return nil, err
		}
		envs[key] = value
	}
	e.log.Info(full)
	return execCmdBackground(dir, envs, command, args...)
}

func getConfig(lang string, configFile string) (*config, error) {
	conf, err := defaultConf(lang)
	if err != nil {
		return nil, errUnsupportedLang
	}
	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil && configFile != defaultConfigFile {
		return nil, errLoadingConfigFile
	}
	if err == nil {
		yaml.Unmarshal(yamlFile, &conf)
	}
	return conf, nil
}

func defaultConf(lang string) (*config, error) {
	switch lang {
	case golangLanguage:
		yamlFile, _ := langTemplate.ReadFile("templates/golang.yaml")
		conf := &config{}
		yaml.Unmarshal(yamlFile, conf)
		if runtime.GOOS == "windows" {
			conf.Deploy.Artifact.Extension = ".exe"
		}
		return conf, nil
	default:
		return nil, errUnsupportedLang
	}
}
