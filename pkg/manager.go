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
	workspace := os.TempDir() + fmt.Sprintf("/crzy-%d", rand.Intn(99999999))
	machine := NewStateMachine()
	updater, err := NewUpdater(workspace, Upstream, machine.action)
	if err != nil {
		log.Printf("could nor initialize workspace: %v", err)
		return
	}
	heading()
	log.Printf("temporary directory: %s", workspace)
	ctx, cancel := context.WithCancel(context.Background())
	admin := http.NewServeMux()
	gitHandler := RefreshRepository(updater, NewGITServer(workspace))
	admin.Handle(fmt.Sprintf("/%s/", repository), gitHandler)
	api := NewUpstreamAPI(Upstream)
	admin.Handle("/", api)
	proxy := NewReverseProxy(Upstream)
	g.Go(func() error { /* yellow */ return NewSignalHandler().Run(ctx, cancel) })
	g.Go(func() error { /* red */ return NewHTTPListener().Run(ctx, ":8080", admin) })
	g.Go(func() error { /* blue */ return NewHTTPListener().Run(ctx, ":8081", proxy) })
	g.Go(func() error { /* green */ return NewCronService().Run(ctx) })
	g.Go(func() error { return NewStoreService(workspace).Run(ctx) })
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
