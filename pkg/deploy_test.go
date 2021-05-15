package pkg

import (
	"context"
	"testing"

	"golang.org/x/sync/errgroup"
)

func Test_deployWorkflow_and_succeed(t *testing.T) {
	deploy := &deployWorkflow{
		log:          &mockLogger{},
		deployStruct: deployStruct{},
		workspace:    ".",
		execdir:      ".",
		keys: map[string]execStruct{
			"test": {
				Command: "echo",
				Args:    []string{"version"},
				WorkDir: ".",
				Envs:    []envVar{{Name: "version", Value: "123"}},
				Output:  "version",
			},
		},
		flow: []string{"test"},
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
	startDeploy <- event{id: triggeredMessage}
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

func Test_deployWorkflow_and_fail(t *testing.T) {
	deploy := &deployWorkflow{
		log:          &mockLogger{},
		deployStruct: deployStruct{},
		keys: map[string]execStruct{
			"test": {
				Command: "doesnotexist",
				WorkDir: ".",
			},
		},
		flow: []string{"test"},
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
	startDeploy <- event{id: triggeredMessage}
	trigger := <-startTrigger
	if trigger.id != deployedMessage {
		t.Error("deploy should send deployed to version")
	}
	cancel()
	if err := g.Wait(); err != nil && err.Error() != "context cancel" {
		t.Error("should receive a context cancel message")
	}
}
