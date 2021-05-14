package pkg

import (
	"embed"
	"errors"
	"fmt"
	"io/ioutil"
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
	ErrUnsupportedLang   = errors.New("unsupportedlang")
	ErrLoadingConfigFile = errors.New("loadingfile")
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
	e.log.Info(full)
	output, err := execCmd(dir, command, args...)
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

func getConfig(lang string, configFile string) (*config, error) {
	conf, err := defaultConf(lang)
	if err != nil {
		return nil, ErrUnsupportedLang
	}
	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil && configFile != defaultConfigFile {
		return nil, ErrLoadingConfigFile
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
		return nil, ErrUnsupportedLang
	}
}
