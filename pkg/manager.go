package pkg

import (
	"context"
	"flag"
	"fmt"
	"os"

	"golang.org/x/sync/errgroup"
)

var (
	repository = "myrepo"
	head       = "master"
	server     = false
)

func Startup(version, commit, date, builtBy string) {
	log := NewLogger("main")
	usage(version, commit, date, builtBy)
	heading()
	g := new(errgroup.Group)
	upstream := NewUpstream()
	machine := NewStateMachine()
	git, err := NewGitServer(repository, head, upstream, machine.action)
	if err != nil {
		log.Error(err, "msg", "could nor initialize GIT server: %v")
		return
	}
	log.Info("temporary directory", "payload", git.absRepoPath)
	proxy := NewReverseProxy(upstream)
	ctx, cancel := context.WithCancel(context.Background())
	g.Go(func() error { /* yellow */ return NewSignalHandler().Run(ctx, cancel) })
	g.Go(func() error { /* red */ return NewHTTPListener().Run(ctx, ":8080", git.ghx) })
	g.Go(func() error { /* blue */ return NewHTTPListener().Run(ctx, ":8081", proxy) })
	g.Go(func() error { /* green */ return NewCronService().Run(ctx) })
	g.Go(func() error { return NewStoreService(git.gitRootPath).Run(ctx) })
	g.Go(func() error { return machine.Run(ctx) })
	if err := g.Wait(); err != nil {
		log.Error(err, "program has stopped")
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
	v := false
	flag.StringVar(&repository, "repository", "myrepo", "GIT repository target name")
	flag.StringVar(&head, "head", "main", "GIT repository target name")
	flag.BoolVar(&server, "server", false, "run as a server")
	flag.BoolVar(&v, "version", false, "display the version")
	flag.Parse()
	if v {
		fmt.Printf("crzy version %s\n", version)
		os.Exit(0)
	}
	if !server {
		flag.PrintDefaults()
		os.Exit(1)
	}
}
