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

// Etape 1: ajouter 2 champs
// - une section `API` de type apiStruct
// - une section `Proxy` de type proxyStruct

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

// Etape 2: Créer 2 type struct
// apiStruct qui contient, un `Username` et un `Password` tous les 2 `string`
// proxyStruct qui contient un slice de string pour un champs `Origin`

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
	colorize   bool
	version    bool
}

func (p *argsParser) parse() args {
	a := args{}
	flag.StringVar(&a.configFile, "config", defaultConfigFile, "configuration file")
	flag.StringVar(&a.repository, "repository", "myrepo", "GIT repository target name")
	flag.StringVar(&a.head, "head", "main", "GIT repository target name")
	flag.BoolVar(&a.colorize, "color", false, "colorize logs")
	flag.BoolVar(&a.version, "version", false, "display the version")
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
	if a.colorize {
		conf.Main.Color = a.colorize
	}
	return conf, nil
}
