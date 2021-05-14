package pkg

import (
	"context"
	"testing"

	"golang.org/x/sync/errgroup"
)

func Test_TriggerWorkflow(t *testing.T) {
	trigger := &triggerWorkflow{
		Log: &mockLogger{},
	}
	g, ctx := errgroup.WithContext(context.TODO())
	ctx, cancel := context.WithCancel(ctx)
	startTrigger := make(chan string)
	defer close(startTrigger)
	startVersion := make(chan string)
	defer close(startVersion)
	g.Go(func() error { return trigger.start(ctx, startTrigger, startVersion) })
	startTrigger <- "start"
	version := <-startVersion
	if version != triggeredMessage {
		t.Error("deploy should start release")
	}
	cancel()
	if err := g.Wait(); err != nil && err.Error() != "context cancel" {
		t.Error("should receive a context cancel message")
	}
}
