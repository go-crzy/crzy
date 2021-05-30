package pkg

import (
	"context"
	"runtime"
	"testing"

	log "github.com/go-crzy/crzy/logr"
	"golang.org/x/sync/errgroup"
)

func Test_deployWorkflow_and_succeed(t *testing.T) {
	deploy := &deployWorkflow{
		log:          &log.MockLogger{},
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
		flow:  []string{"test"},
		state: &stateMockClient{},
	}
	if runtime.GOOS == "windows" {
		deploy.keys["test"] = execStruct{
			Command: "powershell",
			Args:    []string{"-Command", "write-output version"},
			WorkDir: ".",
			Envs:    []envVar{{Name: "version", Value: "123"}},
			Output:  "version"}
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

func Test_deployWorkflow_with_duplicate_envvar(t *testing.T) {
	deploy := &deployWorkflow{
		log:          &log.MockLogger{},
		deployStruct: deployStruct{},
		keys: map[string]execStruct{
			"test": {
				Command: "doesnotexist",
				WorkDir: ".",
			},
		},
		flow:  []string{"test"},
		state: &stateMockClient{},
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
	startDeploy <- event{
		id:   triggeredMessage,
		envs: envVars{{Name: "VERSION", Value: "123"}, {Name: "VERSION", Value: "abc"}}}
	trigger := <-startTrigger
	if trigger.id != deployedMessage {
		t.Error("deploy should send deployed to version")
	}
	cancel()
	if err := g.Wait(); err != nil && err.Error() != "context cancel" {
		t.Error("should receive a context cancel message")
	}
}

func Test_deployWorkflow_with_wrong_artifact_directory(t *testing.T) {
	deploy := &deployWorkflow{
		log: &log.MockLogger{},
		deployStruct: deployStruct{
			Artifact: artifactStruct{
				Filename:  "go",
				Directory: "go-${oops}",
			},
		},
		keys: map[string]execStruct{
			"test": {
				Command: "doesnotexist",
				WorkDir: ".",
			},
		},
		flow:  []string{"test"},
		state: &stateMockClient{},
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
	startDeploy <- event{
		id: triggeredMessage,
	}
	trigger := <-startTrigger
	if trigger.id != deployedMessage {
		t.Error("deploy should send deployed to version")
	}
	cancel()
	if err := g.Wait(); err != nil && err.Error() != "context cancel" {
		t.Error("should receive a context cancel message")
	}
}

func Test_deployWorkflow_with_wrong_artifact_filename(t *testing.T) {
	deploy := &deployWorkflow{
		log: &log.MockLogger{},
		deployStruct: deployStruct{
			Artifact: artifactStruct{
				Filename:  "go-${oops}",
				Directory: ".",
			},
		},
		keys: map[string]execStruct{
			"test": {
				Command: "doesnotexist",
				WorkDir: ".",
			},
		},
		flow:  []string{"test"},
		state: &stateMockClient{},
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
	startDeploy <- event{
		id: triggeredMessage,
	}
	trigger := <-startTrigger
	if trigger.id != deployedMessage {
		t.Error("deploy should send deployed to version")
	}
	cancel()
	if err := g.Wait(); err != nil && err.Error() != "context cancel" {
		t.Error("should receive a context cancel message")
	}
}

func Test_deployWorkflow_with_wrong_flow(t *testing.T) {
	deploy := &deployWorkflow{
		log: &log.MockLogger{},
		deployStruct: deployStruct{
			Artifact: artifactStruct{
				Filename:  "go",
				Directory: ".",
			},
			Install: execStruct{
				Command: "doesnotexist",
				WorkDir: ".",
			},
		},
		keys: map[string]execStruct{
			"install": {
				Command: "",
				WorkDir: ".",
			},
			"test": {
				Command: "doesnotexist",
				WorkDir: ".",
			},
		},
		flow:  []string{"install", "test"},
		state: &stateMockClient{},
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
	startDeploy <- event{
		id: triggeredMessage,
	}
	trigger := <-startTrigger
	if trigger.id != deployedMessage {
		t.Error("deploy should send deployed to version")
	}
	cancel()
	if err := g.Wait(); err != nil && err.Error() != "context cancel" {
		t.Error("should receive a context cancel message")
	}
}

func Test_deployWorkflow_and_fail(t *testing.T) {
	deploy := &deployWorkflow{
		log:          &log.MockLogger{},
		deployStruct: deployStruct{},
		keys: map[string]execStruct{
			"test": {
				Command: "doesnotexist",
				WorkDir: ".",
			},
		},
		flow:  []string{"test"},
		state: &stateMockClient{},
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
