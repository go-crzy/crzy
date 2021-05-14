package pkg

import (
	"context"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
)

func Test_WorkflowsAndCancel(t *testing.T) {
	conf := config{
		Deploy:  deployStruct{},
		Trigger: triggerStruct{},
		Release: releaseStruct{},
	}
	r := runContainer{
		Log: &mockLogger{},
	}
	g, ctx := errgroup.WithContext(context.TODO())
	ctx, cancel := context.WithCancel(ctx)
	startTrigger := make(chan string)
	defer close(startTrigger)
	g.Go(func() error { return r.createAndStartWorkflows(ctx, conf, startTrigger) })
	time.Sleep(200 * time.Microsecond)
	cancel()
	if err := g.Wait(); err != nil && err.Error() != "context cancel" {
		t.Error("should receive a context cancel message")
	}
}
