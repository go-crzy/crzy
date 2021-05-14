package pkg

import (
	"context"
	"testing"

	"golang.org/x/sync/errgroup"
)

func Test_DeployWorkflow(t *testing.T) {
	deploy := &deployWorkflow{
		log: &mockLogger{},
	}
	g, ctx := errgroup.WithContext(context.TODO())
	ctx, cancel := context.WithCancel(ctx)
	startDeploy := make(chan event)
	defer close(startDeploy)
	startRelease := make(chan event)
	defer close(startRelease)
	startTrigger := make(chan event)
	defer close(startTrigger)
	g.Go(func() error { return deploy.start(ctx, startDeploy, startRelease, startTrigger) })
	startDeploy <- event{id: deployedMessage}
	release := <-startRelease
	if release.id != deployedMessage {
		t.Error("deploy should start release, current:", release)
	}
	version := <-startTrigger
	if version.id != deployedMessage {
		t.Error("deploy should send deployed to version")
	}
	cancel()
	if err := g.Wait(); err != nil && err.Error() != "context cancel" {
		t.Error("should receive a context cancel message")
	}
}
