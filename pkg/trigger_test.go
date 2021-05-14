package pkg

import (
	"context"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
)

func Test_TriggerWorkflowWithSuccess(t *testing.T) {
	trigger := &triggerWorkflow{
		trigger: triggerStruct{},
		log:     &mockLogger{},
		command: &mockVersionAndSync{output: true},
		head:    "main",
		git:     &mockGitCommand{},
	}
	g, ctx := errgroup.WithContext(context.TODO())
	ctx, cancel := context.WithCancel(ctx)
	startTrigger := make(chan string)
	defer close(startTrigger)
	startDeploy := make(chan string)
	defer close(startDeploy)
	g.Go(func() error { return trigger.start(ctx, startTrigger, startDeploy) })
	startTrigger <- triggeredMessage
	startTrigger <- triggeredMessage
	deploy := <-startDeploy
	if deploy != triggeredMessage {
		t.Error("deploy should start version")
	}
	startTrigger <- triggeredMessage
	startTrigger <- deployedMessage
	deploy = <-startDeploy
	if deploy != triggeredMessage {
		t.Error("deploy should start version")
	}
	cancel()
	if err := g.Wait(); err != nil && err.Error() != "context cancel" {
		t.Error("should receive a context cancel message")
	}
}

func Test_VersionWorkflowWithFailure(t *testing.T) {
	command := mockVersionAndSync{output: true}
	trigger := &triggerWorkflow{
		trigger: triggerStruct{},
		log:     &mockLogger{},
		command: &command,
		head:    "main",
		git:     &mockGitCommand{},
	}
	g, ctx := errgroup.WithContext(context.TODO())
	ctx, cancel := context.WithCancel(ctx)
	startTrigger := make(chan string)
	defer close(startTrigger)
	startDeploy := make(chan string)
	defer close(startDeploy)
	g.Go(func() error { return trigger.start(ctx, startTrigger, startDeploy) })
	startTrigger <- triggeredMessage
	time.Sleep(100 * time.Microsecond)
	command.output = false
	startTrigger <- triggeredMessage
	time.Sleep(100 * time.Microsecond)
	command.output = true
	startTrigger <- triggeredMessage
	deploy := <-startDeploy
	if deploy != triggeredMessage {
		t.Error("deploy should start version")
	}
	startTrigger <- triggeredMessage
	time.Sleep(100 * time.Microsecond)
	command.output = false
	startTrigger <- deployedMessage
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
