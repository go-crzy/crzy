package crzy

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"golang.org/x/sync/errgroup"
)

var (
	service = "color"
)

func RefreshRepository(updater Updater, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		method := r.Method
		next.ServeHTTP(w, r)
		if path == "/color/.git/git-receive-pack" && method == "POST" {
			log.Printf("%s on %s continue", path, method)
			_, err := updater.Update("color")
			if err != nil {
				log.Printf("error updating color, %v", err)
			}
		}
	})
}

func Startup() {
	g := new(errgroup.Group)
	Upstream := &DefaultUpstreams{
		Versions: map[string]HTTPProcess{},
	}
	rand.Seed(time.Now().UTC().UnixNano())
	workspace := os.TempDir() + fmt.Sprintf("crzy-%d", rand.Intn(99999999))
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
	admin.Handle(fmt.Sprintf("/%s/", service), gitHandler)
	api := NewUpstreamAPI(Upstream)
	admin.Handle("/", api)
	proxy := NewReverseProxy(Upstream)
	g.Go(func() error { return NewSignalHandler().Run(ctx, cancel) })
	g.Go(func() error { return NewHTTPListener().Run(ctx, ":8080", admin) })
	g.Go(func() error { return NewHTTPListener().Run(ctx, ":8081", proxy) })
	g.Go(func() error { return NewCronService().Run(ctx) })
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
