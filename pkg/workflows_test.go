package pkg

import (
	"context"
	"os"
	"runtime"
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
	g.Go(func() error {
		return r.createAndStartWorkflows(ctx, &stateManager{
			notifier: make(chan stepEvent),
			log:      &mockLogger{},
			state:    &defaultState{},
		},
			git,
			startTrigger,
			f,
		)
	})
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
	err := r.createAndStartWorkflows(
		context.TODO(),
		&stateManager{
			notifier: make(chan stepEvent),
			log:      &mockLogger{},
			state:    &defaultState{},
		},
		git,
		startTrigger,
		f)
	if err == nil {
		t.Error("should receive an error message")
	}
}

func Test_execute_and_succeed(t *testing.T) {
	workflow := &workflow{
		log:     &mockLogger{},
		version: "version",
		name:    "deploy",
		basedir: ".",
		envs:    envVars{},
		state:   &stateMockClient{},
	}
	e := &execStruct{
		log:     &mockLogger{},
		Command: "git",
		Args:    []string{"version"},
		WorkDir: ".",
		Envs:    envVars{},
		Output:  "",
	}
	_, err := workflow.execute(e)
	if err != nil {
		t.Error("should succeed")
	}
}

func Test_execute_and_succeed_with_output(t *testing.T) {
	workflow := &workflow{
		log:     &mockLogger{},
		version: "version",
		name:    "deploy",
		basedir: ".",
		envs:    envVars{},
		state:   &stateMockClient{},
	}
	e := &execStruct{
		log:     &mockLogger{},
		Command: "git",
		Args:    []string{"version"},
		WorkDir: ".",
		Envs:    envVars{},
		Output:  "data",
	}
	env, err := workflow.execute(e)
	if err != nil {
		t.Error("should succeed")
	}
	if env == nil || env.Name != "data" ||
		len(env.Value) < 11 || env.Value[0:11] != "git version" {
		t.Error("should return data=\"git version\"", env.Value)
	}
}

func Test_execute_and_fail_command(t *testing.T) {
	workflow := &workflow{
		log:     &mockLogger{},
		version: "version",
		name:    "deploy",
		basedir: ".",
		envs:    envVars{},
		state:   &stateMockClient{},
	}
	e := &execStruct{
		log:     &mockLogger{},
		Command: "${xxx}",
		Args:    []string{"-f", "config.go"},
		WorkDir: ".",
		Envs:    envVars{},
		Output:  "",
	}
	_, err := workflow.execute(e)
	if err != errMissingEnv {
		t.Error(err, "should fail")
	}
}

func Test_execute_and_fail_combinedoutput(t *testing.T) {
	workflow := &workflow{
		log:     &mockLogger{},
		version: "version",
		name:    "deploy",
		basedir: ".",
		envs:    envVars{},
		state:   &stateMockClient{},
	}
	e := &execStruct{
		log:     &mockLogger{},
		Command: "doesnotexist",
		Args:    []string{"-f", "config.go"},
		WorkDir: ".",
		Envs:    envVars{},
	}
	_, err := workflow.execute(e)
	if err == nil ||
		(err.Error() != "exec: \"doesnotexist\": executable file not found in $PATH" &&
			err.Error() != "exec: \"doesnotexist\": executable file not found in %PATH%") {

		t.Error(err, "should fail")
	}
}

func Test_execute_without_exec(t *testing.T) {
	workflow := &workflow{
		log:     &mockLogger{},
		version: "version",
		name:    "deploy",
		basedir: ".",
		envs:    envVars{},
		state:   &stateMockClient{},
	}
	_, err := workflow.execute(nil)
	if err != errNoExcution {
		t.Error(err, "should fail")
	}
}

func Test_start_no_envs(t *testing.T) {
	workflow := &workflow{
		log:     &mockLogger{},
		version: "version",
		name:    "deploy",
		basedir: ".",
		envs:    envVars{{Name: "version", Value: "version"}},
		state:   &stateMockClient{},
	}
	e := execStruct{
		log:     &mockLogger{},
		Command: "tail",
		Args:    []string{"-f", "config.go"},
		WorkDir: ".",
		Envs:    envVars{},
	}
	if runtime.GOOS == "windows" {
		e.Command = "powershell"
		e.Args = []string{"-Command", "Get-Content config.go -Wait"}
	}
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Error(err, "start failed")
		t.FailNow()
	}
	workflow.basedir = dir
	p, err := workflow.start(&e)
	if err != nil {
		t.Error(err, "start failed")
		t.FailNow()
	}
	if p == nil {
		t.Error("process is empty")
	}
	err = p.Kill()
	if err != nil {
		t.Error(err, "kill failed")
	}
	err = os.RemoveAll(dir)
	if err != nil {
		t.Error(err, "should remove all files")
	}
}

func Test_start_and_fail_command(t *testing.T) {
	workflow := &workflow{
		log:     &mockLogger{},
		version: "version",
		name:    "deploy",
		basedir: ".",
		envs:    envVars{},
		state:   &stateMockClient{},
	}
	e := &execStruct{
		log:     &mockLogger{},
		Command: "${xxx}",
		Args:    []string{"-f", "config.go"},
		WorkDir: ".",
		Envs:    envVars{},
	}
	_, err := workflow.start(e)
	if err != errMissingEnv {
		t.Error(err, "should fail")
	}
}

func Test_start_without_exec(t *testing.T) {
	workflow := &workflow{
		log:     &mockLogger{},
		version: "version",
		name:    "deploy",
		basedir: ".",
		envs:    envVars{},
		state:   &stateMockClient{},
	}
	_, err := workflow.start(nil)
	if err != errNoExcution {
		t.Error(err, "should fail")
	}
}

func Test_start_with_envs(t *testing.T) {
	workflow := &workflow{
		log:     &mockLogger{},
		version: "version",
		name:    "deploy",
		basedir: ".",
		envs:    envVars{{Name: "version", Value: "version"}},
		state:   &stateMockClient{},
	}
	e := &execStruct{
		log:     &mockLogger{},
		Command: "tail",
		Args:    []string{"-f", "config.go"},
		WorkDir: ".",
		Envs:    envVars{{Name: "port", Value: "1234"}},
	}
	if runtime.GOOS == "windows" {
		e.Command = "powershell"
		e.Args = []string{"-Command", "Get-Content config.go -Wait"}
	}
	_, err := e.Envs.toMap()
	if err != nil {
		t.Error(err, "should be able to convert envs")
		t.FailNow()
	}
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Error(err, "start failed")
		t.FailNow()
	}
	workflow.basedir = dir
	p, err := workflow.start(e)
	if err != nil {
		t.Error(err, "start failed")
		t.FailNow()
	}
	if p == nil {
		t.Error("process is empty")
	}
	err = p.Kill()
	if err != nil {
		t.Error(err, "kill failed")
	}
	err = os.RemoveAll(dir)
	if err != nil {
		t.Error(err, "should remove all files")
	}
}
