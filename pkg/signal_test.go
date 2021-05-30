package pkg

import (
	"context"
	"os"
	"testing"

	log "github.com/go-crzy/crzy/logr"
	"golang.org/x/sync/errgroup"
)

func Test_newSignalWithInterrupt(t *testing.T) {
	run := &runContainer{
		Log: &log.MockLogger{},
	}
	signal := run.newSignalHandler()
	g, ctx := errgroup.WithContext(context.TODO())
	ctx, cancel := context.WithCancel(ctx)
	g.Go(func() error { return signal.run(ctx, cancel) })
	signal.signalc <- os.Interrupt
	if err := g.Wait(); err != nil {
		t.Error("should not receive an error")
	}
}

func Test_newSignalWithCancel(t *testing.T) {
	run := &runContainer{
		Log: &log.MockLogger{},
	}
	signal := run.newSignalHandler()
	g, ctx := errgroup.WithContext(context.TODO())
	ctx, cancel := context.WithCancel(ctx)
	g.Go(func() error { return signal.run(ctx, cancel) })
	cancel()
	if err := g.Wait(); err != nil && err.Error() != "context canceled" {
		t.Error("should receive a context canceled message", err)
	}
}
