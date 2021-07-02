package pkg

import (
	"embed"
	"errors"
	"flag"
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
	if err != nil && configFile != defaultConfigFile {
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

type parser interface {
	parse() args
}

type argsParser struct{}

type args struct {
	configFile string
	repository string
	head       string
	nocolor    bool
	version    bool
	lang       string
}

func (p *argsParser) parse() args {
	a := args{}
	flag.StringVar(&a.configFile, "config", defaultConfigFile, "configuration file")
	flag.StringVar(&a.repository, "repository", "myrepo", "GIT repository URI")
	flag.StringVar(&a.head, "head", "main", "GIT branch to build from")
	flag.BoolVar(&a.nocolor, "nocolor", false, "disable log color")
	flag.BoolVar(&a.version, "version", false, "crzy version")
	flag.StringVar(&a.lang, "template", "go", "template for language")
	flag.Parse()
	return a
}

func getConf(a args) (*config, error) {
	conf, err := getConfig(defaultLanguage, a.configFile)
	if err != nil {
		return nil, err
	}
	if a.repository != "myrepo" || conf.Main.Repository == "" {
		conf.Main.Repository = a.repository
	}
	if a.head != "main" || conf.Main.Head == "" {
		conf.Main.Head = a.head
	}
	if a.nocolor {
		conf.Main.Color = false
	}
	return conf, nil
}
