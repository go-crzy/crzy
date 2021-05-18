package pkg

import (
	"context"
	"path"
	"runtime"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
)

func Test_triggerWorkflow_and_succeed(t *testing.T) {
	trigger := &triggerWorkflow{
		triggerStruct: triggerStruct{},
		log:           &mockLogger{},
		command:       &mockTriggerCommand{output: true},
		head:          "main",
		git:           &mockGitSuccessCommand{},
		state:         &stateMockClient{},
	}
	g, ctx := errgroup.WithContext(context.TODO())
	ctx, cancel := context.WithCancel(ctx)
	startTrigger := make(chan event)
	defer close(startTrigger)
	startDeploy := make(chan event)
	defer close(startDeploy)
	g.Go(func() error { return trigger.start(ctx, startTrigger, startDeploy) })
	startTrigger <- event{id: triggeredMessage}
	startTrigger <- event{id: triggeredMessage}
	deploy := <-startDeploy
	if deploy.id != triggeredMessage {
		t.Error("deploy should start version")
	}
	startTrigger <- event{id: triggeredMessage}
	startTrigger <- event{id: deployedMessage}
	deploy = <-startDeploy
	if deploy.id != triggeredMessage {
		t.Error("deploy should start version")
	}
	cancel()
	if err := g.Wait(); err != nil && err.Error() != "context cancel" {
		t.Error("should receive a context cancel message")
	}
}

func Test_versionWorkflow_and_fail(t *testing.T) {
	command := &mockTriggerCommand{output: true}
	trigger := &triggerWorkflow{
		triggerStruct: triggerStruct{},
		log:           &mockLogger{},
		command:       command,
		head:          "main",
		git:           &mockGitSuccessCommand{},
		state:         &stateMockClient{},
	}
	g, ctx := errgroup.WithContext(context.TODO())
	ctx, cancel := context.WithCancel(ctx)
	startTrigger := make(chan event)
	defer close(startTrigger)
	startDeploy := make(chan event)
	defer close(startDeploy)
	g.Go(func() error { return trigger.start(ctx, startTrigger, startDeploy) })
	startTrigger <- event{id: triggeredMessage}
	time.Sleep(100 * time.Microsecond)
	command.output = false
	startTrigger <- event{id: triggeredMessage}
	time.Sleep(100 * time.Microsecond)
	command.output = true
	startTrigger <- event{id: triggeredMessage}
	deploy := <-startDeploy
	if deploy.id != triggeredMessage {
		t.Error("deploy should start version")
	}
	startTrigger <- event{id: triggeredMessage}
	time.Sleep(100 * time.Microsecond)
	command.output = false
	startTrigger <- event{id: deployedMessage}
	time.Sleep(200 * time.Millisecond)
	cancel()
	if err := g.Wait(); err != nil && err.Error() != "context cancel" {
		t.Error("should receive a context cancel message")
	}
}

func Test_versionCommand(t *testing.T) {
	w := &triggerWorkflow{
		triggerStruct: triggerStruct{
			Version: versionStruct{
				Command: "echo",
				Args:    []string{"-n", "1"},
			},
		},
		git: &defaultGitCommand{
			store: store{
				workdir: "/tmp",
			},
		},
		state: &stateMockClient{},
	}
	version := &defaultTriggerCommand{}
	version.setTriggerWorkflow(*w)
	x, err := version.version()
	if err != nil || x != "1" {
		t.Error("version should succeed and return 1")
	}
}

func Test_defaultTriggerCommand(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Dir(filename)
	w := &triggerWorkflow{
		triggerStruct: triggerStruct{
			Version: versionStruct{
				Command: "",
				Args:    []string{},
			},
		},
		git: &defaultGitCommand{
			bin: "git",
			store: store{
				workdir: dir,
			},
			log: &mockLogger{},
		},
		log:   &mockLogger{},
		state: &stateMockClient{},
	}
	version := &defaultTriggerCommand{}
	version.setTriggerWorkflow(*w)
	x, err := version.version()
	if err != nil || len(x) != 16 {
		t.Error("version should be 16 length", x)
	}
}

func Test_defaultTriggerCommand_with_error(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Dir(filename)
	w := &triggerWorkflow{
		triggerStruct: triggerStruct{
			Version: versionStruct{
				Command: "",
				Args:    []string{},
			},
		},
		git: &defaultGitCommand{
			bin: "x",
			store: store{
				workdir: dir,
			},
			log: &mockLogger{},
		},
		log:   &mockLogger{},
		state: &stateMockClient{},
	}
	version := &defaultTriggerCommand{}
	version.setTriggerWorkflow(*w)
	_, err := version.version()
	if err == nil || err.Error()[0:4] != "exec" {
		t.Error("should return an error with execution", err)
	}
}
