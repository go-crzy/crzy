package pkg

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"golang.org/x/sync/errgroup"
)

var (
	repository = "myrepo"
	head       = "master"
	server     = false
)

func Startup() {
	usage()
	heading()
	g := new(errgroup.Group)
	upstream := NewUpstream()
	machine := NewStateMachine()
	git, err := NewGitServer(repository, head, upstream, machine.action)
	if err != nil {
		log.Printf("could nor initialize GIT server: %v", err)
		return
	}
	log.Printf("temporary directory: %s", git.absRepoPath)
	proxy := NewReverseProxy(upstream)
	ctx, cancel := context.WithCancel(context.Background())
	g.Go(func() error { /* yellow */ return NewSignalHandler().Run(ctx, cancel) })
	g.Go(func() error { /* red */ return NewHTTPListener().Run(ctx, ":8080", git.ghx) })
	g.Go(func() error { /* blue */ return NewHTTPListener().Run(ctx, ":8081", proxy) })
	g.Go(func() error { /* green */ return NewCronService().Run(ctx) })
	g.Go(func() error { return NewStoreService(git.gitRootPath).Run(ctx) })
	g.Go(func() error { return machine.Run(ctx) })
	if err := g.Wait(); err != nil {
		log.Printf("program has stopped (%v)", err)
	}
}

func heading() {
	log.Println()
	log.Println("  █▀▀ █▀▀█ ▀▀█ █░░█")
	log.Println("  █░░ █▄▄▀ ▄▀░ █▄▄█")
	log.Println("  ▀▀▀ ▀░▀▀ ▀▀▀ ▄▄▄█")
	log.Println()
	log.Println()
}

func usage() {
	flag.StringVar(&repository, "repository", "myrepo", "GIT repository target name")
	flag.StringVar(&head, "head", "main", "GIT repository target name")
	flag.BoolVar(&server, "server", false, "run as a server")
	flag.Parse()
	if !server {
		fmt.Println("crzy start a GIT server to start/stop services")
		flag.PrintDefaults()
		os.Exit(1)
	}
}
