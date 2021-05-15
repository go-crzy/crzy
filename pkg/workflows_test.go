package pkg

import (
	"context"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
)

func Test_workflowsAndCancel(t *testing.T) {
	conf := config{
		Deploy:  deployStruct{},
		Trigger: triggerStruct{},
		Release: releaseStruct{
			PortRange: portRangeStruct{
				Min: 8090,
				Max: 8100,
			},
		},
	}
	r := runContainer{
		Log:    &mockLogger{},
		Config: conf,
	}
	g := new(errgroup.Group)
	ctx, cancel := context.WithCancel(context.TODO())
	startTrigger := make(chan event)
	defer close(startTrigger)
	git := &mockGitSuccessCommand{}
	f := func(port string) {}
	g.Go(func() error { return r.createAndStartWorkflows(ctx, git, startTrigger, f) })
	time.Sleep(500 * time.Microsecond)
	cancel()
	if err := g.Wait(); err != nil && err.Error() != "context canceled" {
		t.Error("should receive a context canceled message")
	}
}

func Test_workflowsAndFail(t *testing.T) {
	conf := config{
		Deploy:  deployStruct{},
		Trigger: triggerStruct{},
		Release: releaseStruct{
			PortRange: portRangeStruct{
				Min: 8090,
				Max: 8100,
			},
		},
	}
	r := runContainer{
		Log:    &mockLogger{},
		Config: conf,
	}
	startTrigger := make(chan event)
	defer close(startTrigger)
	git := &mockGitFailCommand{}
	f := func(port string) {}
	err := r.createAndStartWorkflows(context.TODO(), git, startTrigger, f)
	if err == nil {
		t.Error("should receive an error message")
	}
}
