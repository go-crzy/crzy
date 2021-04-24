package crzy

import (
	"context"
	"log"

	"golang.org/x/sync/errgroup"
)

func Startup() {
	g := new(errgroup.Group)
	gitHandler := NewGITServer("/tmp/workspace")
	heading()
	ctx, cancel := context.WithCancel(context.Background())
	Upstream := &DefaultUpstreams{
		Versions: map[string]HTTPProcess{},
	}
	proxy := NewReverseProxy(Upstream)
	api := NewUpstreamAPI(Upstream)
	g.Go(func() error { return NewSignalHandler().Run(ctx, cancel) })
	g.Go(func() error { return NewHTTPListener().Run(ctx, ":8080", gitHandler) })
	g.Go(func() error { return NewHTTPListener().Run(ctx, ":8081", proxy) })
	g.Go(func() error { return NewHTTPListener().Run(ctx, ":8082", api) })
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
