package pkg

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"golang.org/x/sync/errgroup"
)

var (
	repository = "myrepo"
	head       = "master"
	server     = false
)

func RefreshRepository(updater Updater, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		method := r.Method
		next.ServeHTTP(w, r)
		if path == fmt.Sprintf("/%s/.git/git-receive-pack", repository) && method == "POST" {
			log.Printf("%s on %s continue", path, method)
			_, err := updater.Update(repository)
			if err != nil {
				log.Printf("error updating %s, %v", repository, err)
			}
		}
	})
}

func Usage() {
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

func Startup() {
	Usage()
	g := new(errgroup.Group)
	Upstream := &DefaultUpstreams{
		Versions: map[string]HTTPProcess{},
	}
	rand.Seed(time.Now().UTC().UnixNano())
	machine := NewStateMachine()
	git, err := NewGitServer(repository, head)
	if err != nil {
		log.Printf("could nor initialize GIT server: %v", err)
		return
	}
	updater, err := NewUpdater(git.absRepoPath, Upstream, machine.action)
	if err != nil {
		log.Printf("could nor initialize workspace: %v", err)
		return
	}
	heading()
	gitHandler := RefreshRepository(updater, Logging(git.ghx))
	log.Printf("temporary directory: %s", git.absRepoPath)
	// admin := http.NewServeMux()
	// admin.Handle(fmt.Sprintf("/%s/", repository), gitHandler)
	// api := NewUpstreamAPI(Upstream)
	// admin.Handle("/", api)
	proxy := NewReverseProxy(Upstream)
	ctx, cancel := context.WithCancel(context.Background())
	g.Go(func() error { /* yellow */ return NewSignalHandler().Run(ctx, cancel) })
	g.Go(func() error { /* red */ return NewHTTPListener().Run(ctx, ":8080", gitHandler) })
	g.Go(func() error { /* blue */ return NewHTTPListener().Run(ctx, ":8081", proxy) })
	g.Go(func() error { /* green */ return NewCronService().Run(ctx) })
	g.Go(func() error { return NewStoreService(git.absRepoPath).Run(ctx) })
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
