package pkg

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
)

var (
	conf = &config{}
)
type config struct {
	sync.Mutex
	Main    MainStruct
	Version ExecStruct
	Deployment DeploymentStruct
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
	Build ExecStruct
}

type ArtifactStruct struct {
	Type string
	Pattern string
}

type ExecStruct struct {
	Command string
	Args []string
	Directory string
}

func Startup(version, commit, date, builtBy string) {
	log := NewLogger("main")
	usage(version, commit, date, builtBy)
	heading()
	startGroup := new(errgroup.Group)
	endGroup := new(errgroup.Group)
	upstream := NewUpstream()
	machine := NewStateMachine(upstream)
	git, err := NewGitServer(conf.Main.Repository, conf.Main.Head, upstream, machine.action)
	if err != nil {
		log.Error(err, "msg", "could nor initialize GIT server: %v")
		return
	}
	log.Info("temporary directory", "payload", git.absRepoPath)
	proxy := NewReverseProxy(upstream)
	startContext, startCancel := context.WithCancel(context.Background())
	endContext, endCancel := context.WithCancel(context.Background())
	startGroup.Go(func() error { return NewSignalHandler().Run(startContext, startCancel) })
	startGroup.Go(func() error { return NewHTTPListener().Run(startContext, ":8080", git.ghx) })
	startGroup.Go(func() error { return NewHTTPListener().Run(startContext, ":8081", proxy) })
	startGroup.Go(func() error { return NewCronService().Run(startContext) })
	endGroup.Go(func() error { return NewStoreService(git.gitRootPath).Run(endContext) })
	startGroup.Go(func() error { return machine.Run(startContext) })
	if err := startGroup.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		log.Error(err, "compute have stopped with error")
	}
	log.Info("stopping store")
	endCancel()
	if err := startGroup.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		log.Error(err, "store has stopped with error")
	}
}

func heading() {
	log := NewLogger("")
	log.Info("")
	log.Info("█▀▀ █▀▀█ ▀▀█ █░░█")
	log.Info("█░░ █▄▄▀ ▄▀░ █▄▄█")
	log.Info("▀▀▀ ▀░▀▀ ▀▀▀ ▄▄▄█")
	log.Info("")
}

func usage(version, commit, date, builtBy string) {
	configFile := ""
	flag.StringVar(&configFile, "config", "crzy.yaml", "configuration file")
	repository := ""
	head := ""
	server := false
	colorize := false
	flag.StringVar(&repository, "repository", "myrepo", "GIT repository target name")
	flag.StringVar(&head, "head", "main", "GIT repository target name")
	flag.BoolVar(&colorize, "color", false, "colorize logs")
	flag.BoolVar(&server, "server", false, "run as a server")
	v := false
	flag.BoolVar(&v, "version", false, "display the version")
	flag.Parse()
	if v {
		fmt.Printf("crzy version %s\n", version)
		os.Exit(0)
	}
	getConfig(configFile)
	if repository != "myrepo" || conf.Main.Repository == "" {
		conf.Main.Repository = repository
	}
	if head != "main" || conf.Main.Head == "" {
		conf.Main.Head = head
	}
	if colorize {
		conf.Main.Color = colorize
	}
	if server {
		conf.Main.Server = server
	}
	if !conf.Main.Server {
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func getConfig(configFile string) {
	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil && configFile != "crzy.yaml" {
		fmt.Printf("Can not access file %s \n", configFile)
		os.Exit(1)
	}
	yaml.Unmarshal(yamlFile, &conf)
}
