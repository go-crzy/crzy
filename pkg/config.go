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
	ErrUnsupportedLang   = errors.New("unsupportedlang")
	ErrLoadingConfigFile = errors.New("loadingfile")
)

type config struct {
	sync.Mutex
	Main       MainStruct
	Version    ExecStruct
	Deployment DeploymentStruct
	Release    ReleaseStruct
	Notifier   NotifierStruct
}

type MainStruct struct {
	Head       string
	Server     bool
	Color      bool
	Repository string
	ApiPort    int `yaml:"api_port"`
	ProxyPort  int `yaml:"proxy_port"`
}

type DeploymentStruct struct {
	Artifact ArtifactStruct
	Build    ExecStruct
	Test     ExecStruct
	Run      ExecStruct
}

type ArtifactStruct struct {
	Type    string
	Pattern string
}

type ExecStruct struct {
	Command   string
	Args      []string
	Directory string
}

type ReleaseStruct struct {
	Keep     int
	Model    string
	Accessor AccessorStruct
}

type AccessorStruct struct {
	Type string
	Name string
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
			conf.Deployment.Artifact.Pattern += ".exe"
		}
		return conf, nil
	default:
		return nil, ErrUnsupportedLang
	}
}
