package pkg

import (
	"embed"
	"errors"
	"io/ioutil"
	"runtime"
	"sync"

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
		conf.Deploy.Artifact.Extension = map[string]string{"windows": ".exe"}[runtime.GOOS]
		return conf, nil
	default:
		return nil, errUnsupportedLang
	}
}
