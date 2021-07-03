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
	DefaultConfigFile = "crzy.yaml"
	defaultLanguage   = "go"
	golangLanguage    = "go"
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
	Scripts  []string
}

type mainStruct struct {
	Repository string
	Head       string
	Color      bool
	API        apiStruct   `yaml:"api"`
	Proxy      proxyStruct `yaml:"proxy"`
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

type apiStruct struct {
	Username, Password *string
	Port               int `yaml:"port"`
}

type proxyStruct struct {
	Origins []string `yaml:"origins"`
	Port    int      `yaml:"port"`
}

func getConfig(lang string, configFile string) (*config, error) {
	conf, err := defaultConf(lang)
	if err != nil {
		return nil, errUnsupportedLang
	}
	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil && configFile != DefaultConfigFile {
		return nil, errLoadingConfigFile
	}
	if err == nil {
		yaml.Unmarshal(yamlFile, &conf)
	}
	return conf, nil
}

func defaultConf(lang string) (conf *config, err error) {
	switch lang {
	case golangLanguage:
		yamlFile, _ := langTemplate.ReadFile("templates/golang.yaml")
		yaml.Unmarshal(yamlFile, &conf)
		conf.Deploy.Artifact.Extension = map[string]string{"windows": ".exe"}[runtime.GOOS]
		return
	default:
		return nil, errUnsupportedLang
	}
}

type Args struct {
	ConfigFile string
	Repository string
	Head       string
	NoColor    bool
	Version    bool
	Lang       string
}

func (c *defaultContainer) getConf(a Args) error {
	conf, err := getConfig(defaultLanguage, a.ConfigFile)
	c.config = conf
	if err != nil {
		return err
	}
	if a.Repository != "myrepo" || conf.Main.Repository == "" {
		conf.Main.Repository = a.Repository
	}
	if a.Head != "main" || conf.Main.Head == "" {
		conf.Main.Head = a.Head
	}
	if a.NoColor {
		conf.Main.Color = false
	}
	return nil
}
