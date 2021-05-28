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

var ErrVersionRequested = errors.New("version")

type runContainer struct {
	Log    logr.Logger
	Config config
}

func Startup(ctx context.Context, version, commit, date, builtBy string) {
	log := newCrzyLogger("main", false)
	run := &runContainer{Log: log}
	var err error
	run.Config, err = run.parse(version, commit)
	if err == ErrVersionRequested {
		os.Exit(0)
	}
	log = newCrzyLogger("main", run.Config.Main.Color)
	run.Log = log
	group, ctx := errgroup.WithContext(ctx)
	store, err := run.createStore()
	if err != nil {
		os.Exit(1)
	}
	state := run.newStateManager()
	gitCommand, err := run.newDefaultGitCommand(*store)
	if err != nil {
		log.Error(err, "msg", "could not get GIT")
		return
	}
	defer store.delete()
	trigger := make(chan event)
	defer close(trigger)
	release := make(chan event)
	defer close(release)
	gitServer, err := run.newGitServer(*store, state, trigger, release)
	if err != nil {
		log.Error(err, "msg", "could not initialize GIT server")
		return
	}
	listener1, err := run.newHTTPListener(":8080")
	if err != nil {
		log.Error(err, "msg", "could not start the GIT listener")
		return
	}
	upstream := newUpstream(state.state)
	f := func(port string) { upstream.setDefault(port) }
	proxy := newReverseProxy(upstream)
	listener2, err := run.newHTTPListener(":8081")
	if err != nil {
		log.Error(err, "msg", "could not start the GIT listener")
		return
	}
	ctx, cancel := context.WithCancel(ctx)
	group.Go(func() error { return run.newSignalHandler().run(ctx, cancel) })
	group.Go(func() error { return listener1.run(ctx, *gitServer.ghx) })
	group.Go(func() error { return listener2.run(ctx, proxy) })
	group.Go(func() error { return run.createAndStartWorkflows(ctx, state, gitCommand, trigger, release, f) })
	if err := group.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		log.Error(err, "compute have stopped with error")
	}
}

func (r *runContainer) parse(version, commit string) (config, error) {
	if version == "" {
		version = "dev"
	}
	if commit == "" {
		commit = "unknown"
	}
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
		fmt.Printf("crzy version %s(%s)\n", version, commit)
		return config{}, ErrVersionRequested
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
	return *conf, nil
}
