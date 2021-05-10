package pkg

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"golang.org/x/sync/errgroup"
)

var (
	conf = &config{}
)

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
		log.Error(err, "msg", "could nor initialize GIT server")
		return
	}
	log.Info("temporary directory", "payload", git.absRepoPath)
	proxy := NewReverseProxy(upstream)
	listener1, err := NewHTTPListener(":8080")
	if err != nil {
		log.Error(err, "msg", "could not start the GIT listener")
		return
	}
	listener2, err := NewHTTPListener(":8081")
	if err != nil {
		log.Error(err, "msg", "could not start the proxy listener")
		return
	}
	startContext, startCancel := context.WithCancel(context.Background())
	endContext, endCancel := context.WithCancel(context.Background())
	startGroup.Go(func() error { return NewSignalHandler().Run(startContext, startCancel) })
	startGroup.Go(func() error { return listener1.Run(startContext, git.ghx) })
	startGroup.Go(func() error { return listener2.Run(startContext, proxy) })
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
	flag.StringVar(&configFile, "config", defaultConfigFile, "configuration file")
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
	var err error
	conf, err = getConfig(defaultLanguage, configFile)
	if err != nil {
		fmt.Printf("could not read file %s\n", configFile)
		os.Exit(1)
	}
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
