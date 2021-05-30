package pkg

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	log "github.com/go-crzy/crzy/logr"
	"github.com/go-logr/logr"
	"golang.org/x/sync/errgroup"
)

var (
	ErrWronglyInitialized = errors.New("wronginit")
)

var (
	version = "dev"
	commit  = "unknown"
)

type DefaultRunner struct {
	log    logr.Logger
	parser parser
	out    io.Writer
}

// NewCrzy create the default runner with the various configuration options.
func NewCrzy() *DefaultRunner {
	return &DefaultRunner{
		log:    log.NewLogger("", log.OptionColor),
		parser: &argsParser{},
		out:    os.Stdout,
	}
}

// Run starts the DefaultRunner and runs Crzy
func (c *DefaultRunner) Run(ctx context.Context) error {
	if c.log == nil {
		return ErrWronglyInitialized
	}
	args := c.parser.parse()
	if args.version {
		fmt.Fprintf(c.out, "crzy version %s(%s)\n", version, commit)
		return nil
	}
	conf, err := getConf(args)
	if err != nil {
		return err
	}
	log := c.log.WithName("main")
	heading(log)
	run := &runContainer{
		Log:    log,
		Config: *conf,
	}
	group, ctx := errgroup.WithContext(ctx)
	store, err := run.createStore()
	if err != nil {
		log.Error(err, "msg", "could not create store")
		return err
	}
	defer store.delete()
	state := run.newStateManager()
	gitCommand, err := run.newDefaultGitCommand(*store)
	if err != nil {
		log.Error(err, "msg", "could not get git")
		return err
	}
	trigger := make(chan event)
	defer close(trigger)
	release := make(chan event)
	defer close(release)
	gitServer, err := run.newGitServer(*store, state, trigger, release)
	if err != nil {
		log.Error(err, "msg", "could not initialize git")
		return err
	}
	listener1, err := run.newHTTPListener(fmt.Sprintf(":%d", run.Config.Main.ApiPort))
	if err != nil {
		log.Error(err, "msg", "could not start git listener")
		return err
	}
	upstream := newUpstream(state.state)
	f := func(port string) { upstream.setDefault(port) }
	proxy := newReverseProxy(upstream)
	listener2, err := run.newHTTPListener(fmt.Sprintf(":%d", run.Config.Main.ProxyPort))
	if err != nil {
		log.Error(err, "msg", "could not start proxy listener")
		return err
	}
	ctx, cancel := context.WithCancel(ctx)
	group.Go(func() error { return run.newSignalHandler().run(ctx, cancel) })
	group.Go(func() error { return listener1.run(ctx, *gitServer.ghx) })
	group.Go(func() error { return listener2.run(ctx, proxy) })
	group.Go(func() error { return run.createAndStartWorkflows(ctx, state, gitCommand, trigger, release, f) })
	if err := group.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		log.Error(err, "compute have stopped with error")
		return err
	}
	return nil
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

type mockParser struct {
	configFile string
	repository string
	head       string
	colorize   bool
	version    bool
}

func (p *mockParser) parse() args {
	return args{
		configFile: p.configFile,
		repository: p.repository,
		head:       p.head,
		colorize:   p.colorize,
		version:    p.version,
	}
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

type runContainer struct {
	Log    logr.Logger
	Config config
}

func heading(log logr.Logger) {
	log.Info("")
	log.Info(" █▀▀ █▀▀█ ▀▀█ █░░█")
	log.Info(" █░░ █▄▄▀ ▄▀░ █▄▄█")
	log.Info(" ▀▀▀ ▀░▀▀ ▀▀▀ ▄▄▄█")
	log.Info("")
}
