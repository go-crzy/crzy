package crzy

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"

	"golang.org/x/sync/errgroup"
)

var (
	service = "color"
)

func SyncRepository() {
	dir := fmt.Sprintf("/tmp/workspace/%s", service)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		cmd := exec.Cmd{
			Path: "git",
			Dir:  "/tmp/workspace",
			Args: []string{"clone", dir, "workdir"},
		}
		cmd.Run()
	}
	cmd := exec.Cmd{
		Path: "git",
		Dir:  dir,
		Args: []string{"pull"},
	}
	cmd.Run()
}

func RefreshRepository(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		method := r.Method
		next.ServeHTTP(w, r)
		if path != "" && method != "" {
			log.Printf("%s on %s continue", path, method)
		}
	})
}

func Startup() {
	g := new(errgroup.Group)
	heading()
	ctx, cancel := context.WithCancel(context.Background())
	Upstream := &DefaultUpstreams{
		Versions: map[string]HTTPProcess{},
	}
	admin := http.NewServeMux()
	gitHandler := NewGITServer("/tmp/workspace")
	admin.Handle(fmt.Sprintf("/%s/", service), gitHandler)
	api := NewUpstreamAPI(Upstream)
	admin.Handle("/", api)
	proxy := NewReverseProxy(Upstream)
	g.Go(func() error { return NewSignalHandler().Run(ctx, cancel) })
	g.Go(func() error { return NewHTTPListener().Run(ctx, ":8080", admin) })
	g.Go(func() error { return NewHTTPListener().Run(ctx, ":8081", proxy) })
	g.Go(func() error { return NewCronService().Run(ctx) })
	g.Go(func() error { return NewStoreService().Run(ctx) })
	g.Go(func() error { return NewStateMachine().Run(ctx) })
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
