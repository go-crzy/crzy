package pkg

import (
	"context"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
)

func Test_ReleaseWorkflow(t *testing.T) {
	release := &releaseWorkflow{
		log: &mockLogger{},
	}
	g, ctx := errgroup.WithContext(context.TODO())
	ctx, cancel := context.WithCancel(ctx)
	startRelease := make(chan string)
	defer close(startRelease)
	g.Go(func() error { return release.start(ctx, startRelease) })
	startRelease <- "start"
	time.Sleep(200 * time.Millisecond)
	cancel()
	if err := g.Wait(); err != nil && err.Error() != "context cancel" {
		t.Error("should receive a context cancel message")
	}
}
