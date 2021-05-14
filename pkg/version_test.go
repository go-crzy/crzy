package pkg

import (
	"context"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
)

func Test_VersionWorkflowWithSuccess(t *testing.T) {
	version := &versionSync{
		Log:     &mockLogger{},
		command: &MockVersionAndSyncSucceed{},
	}
	g, ctx := errgroup.WithContext(context.TODO())
	ctx, cancel := context.WithCancel(ctx)
	startVersion := make(chan string)
	defer close(startVersion)
	startDeploy := make(chan string)
	defer close(startDeploy)
	g.Go(func() error { return version.start(ctx, startVersion, startDeploy) })
	startVersion <- triggeredMessage
	deploy := <-startDeploy
	if deploy != versionedMessage {
		t.Error("deploy should start version")
	}
	startVersion <- triggeredMessage
	startVersion <- deployedMessage
	deploy = <-startDeploy
	if deploy != versionedMessage {
		t.Error("deploy should start version")
	}
	cancel()
	if err := g.Wait(); err != nil && err.Error() != "context cancel" {
		t.Error("should receive a context cancel message")
	}
}

func Test_VersionWorkflowWithFailure(t *testing.T) {
	version := &versionSync{
		Log:     &mockLogger{},
		command: &MockVersionAndSyncFailed{},
	}
	g, ctx := errgroup.WithContext(context.TODO())
	ctx, cancel := context.WithCancel(ctx)
	startVersion := make(chan string)
	defer close(startVersion)
	startDeploy := make(chan string)
	defer close(startDeploy)
	g.Go(func() error { return version.start(ctx, startVersion, startDeploy) })
	startVersion <- triggeredMessage
	version.command = &MockVersionAndSyncSucceed{}
	startVersion <- triggeredMessage
	deploy := <-startDeploy
	if deploy != versionedMessage {
		t.Error("deploy should start version")
	}
	startVersion <- triggeredMessage
	version.command = &MockVersionAndSyncFailed{}
	startVersion <- deployedMessage
	time.Sleep(200 * time.Millisecond)
	cancel()
	if err := g.Wait(); err != nil && err.Error() != "context cancel" {
		t.Error("should receive a context cancel message")
	}
}

func Test_VersionCommand(t *testing.T) {
	version := &DefaultVersionAndSync{}
	x, err := version.run()
	if err != nil || x != "1" {
		t.Error("version should succeed and return 1")
	}
}
