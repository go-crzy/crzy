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
		Log:    &mockLogger{},
		Config: conf,
	}
	g, ctx := errgroup.WithContext(context.TODO())
	ctx, cancel := context.WithCancel(ctx)
	startTrigger := make(chan string)
	defer close(startTrigger)
	git := &mockGitCommand{}
	g.Go(func() error { return r.createAndStartWorkflows(ctx, git, startTrigger) })
	time.Sleep(200 * time.Microsecond)
	cancel()
	if err := g.Wait(); err != nil && err.Error() != "context cancel" {
		t.Error("should receive a context cancel message")
	}
}
