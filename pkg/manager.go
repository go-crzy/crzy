package pkg

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/go-logr/logr"
	"golang.org/x/sync/errgroup"
)

type runContainer struct {
	Log    logr.Logger
	Config config
}

func Startup(version, commit, date, builtBy string) {
	log := newCrzyLogger("main", false)
	run := &runContainer{Log: log}
	run.Config = run.parse(version, commit)
	log = newCrzyLogger("main", run.Config.Main.Color)
	run.Log = log
	group, ctx := errgroup.WithContext(context.Background())
	store, err := run.createStore()
	if err != nil {
		os.Exit(1)
	}
	defer store.delete()
	trigger := make(chan string)
	defer close(trigger)
	git, err := run.newGitServer(*store, trigger)
	if err != nil {
		log.Error(err, "msg", "could not initialize GIT server")
		return
	}
	listener1, err := NewHTTPListener(":8080")
	if err != nil {
		log.Error(err, "msg", "could not start the GIT listener")
		return
	}
	ctx, cancel := context.WithCancel(ctx)
	group.Go(func() error { return NewSignalHandler().Run(ctx, cancel) })
	group.Go(func() error { return listener1.Run(ctx, *git.ghx) })
	if err := group.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		log.Error(err, "compute have stopped with error")
	}
}

func (r *runContainer) parse(version, commit string) config {
	configFile := ""
	repository := ""
	head := ""
	colorize := false
	v := false
	flag.StringVar(&configFile, "config", defaultConfigFile, "configuration file")
	flag.StringVar(&repository, "repository", "myrepo", "GIT repository target name")
	flag.StringVar(&head, "head", "main", "GIT repository target name")
	flag.BoolVar(&colorize, "color", false, "colorize logs")
	flag.BoolVar(&v, "version", false, "display the version")
	flag.Parse()
	if v {
		if version == "" {
			version = "dev"
		}
		if commit == "" {
			commit = "unknown"
		}
		fmt.Printf("crzy version %s(%s)\n", version, commit)
		os.Exit(0)
	}
	conf, _ := getConfig(defaultLanguage, configFile)
	if repository != "myrepo" || conf.Main.Repository == "" {
		conf.Main.Repository = repository
	}
	if head != "main" || conf.Main.Head == "" {
		conf.Main.Head = head
	}
	if colorize {
		conf.Main.Color = colorize
	}
	return *conf
}
