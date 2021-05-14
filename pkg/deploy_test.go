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
	startDeploy := make(chan string)
	defer close(startDeploy)
	startRelease := make(chan string)
	defer close(startRelease)
	startVersion := make(chan string)
	defer close(startVersion)
	g.Go(func() error { return deploy.start(ctx, startDeploy, startRelease, startVersion) })
	startDeploy <- deployedMessage
	release := <-startRelease
	if release != deployedMessage {
		t.Error("deploy should start release, current:", release)
	}
	version := <-startVersion
	if version != deployedMessage {
		t.Error("deploy should send deployed to version")
	}
	cancel()
	if err := g.Wait(); err != nil && err.Error() != "context cancel" {
		t.Error("should receive a context cancel message")
	}
}
