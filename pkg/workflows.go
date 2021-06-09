package pkg

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"golang.org/x/sync/errgroup"
)

const (
	triggeredMessage string = "triggered"
	deployedMessage  string = "deployed"
)

var errNoExcution = errors.New("noexec")

type event struct {
	id   string
	envs envVars
}

func (r *defaultContainer) createAndStartWorkflows(
	ctx context.Context,
	state *stateManager,
	git gitCommand,
	startTrigger chan event,
	startRelease chan event,
	switchUpstream func(string)) error {
	slack := newSlackNotifier(r.config.Notifier.Slack)
	err := git.cloneRepository()
	if err != nil {
		r.log.Error(err, "error cloning repository")
		return err
	}
	g, ctx := errgroup.WithContext(ctx)
	ctx, cancel := context.WithCancel(ctx)
	install := r.config.Deploy.Install
	install.name = "install"
	test := r.config.Deploy.Test
	test.name = "test"
	preBuild := r.config.Deploy.PreBuild
	preBuild.name = "prebuild"
	build := r.config.Deploy.Build
	build.name = "build"
	deploy := &deployWorkflow{
		deployStruct: r.config.Deploy,
		workspace:    git.getWorkspace(),
		execdir:      git.getExecdir(),
		log:          r.log,
		keys: map[string]execStruct{
			"install":   install,
			"test":      test,
			"pre_build": preBuild,
			"build":     build,
		},
		flow:  []string{"install", "test", "pre_build", "build"},
		state: &stateDefaultClient{notifier: state.notifier},
		slack: slack,
	}
	trigger := &triggerWorkflow{
		triggerStruct: r.config.Trigger,
		head:          r.config.Main.Head,
		log:           r.log,
		git:           git,
		command:       &defaultTriggerCommand{},
		state:         &stateDefaultClient{notifier: state.notifier},
	}
	run := r.config.Release.Run
	run.name = "run"
	release := &releaseWorkflow{
		releaseStruct: r.config.Release,
		log:           r.log,
		execdir:       git.getExecdir(),
		keys: map[string]execStruct{
			"run": run,
		},
		flow:           "run",
		processes:      map[string]*os.Process{},
		files:          make(map[string][]*file),
		switchUpstream: switchUpstream,
		state:          &stateDefaultClient{notifier: state.notifier},
		slack:          slack,
	}
	startDeploy := make(chan event)
	defer close(startDeploy)
	g.Go(func() error { return state.start(ctx) })
	g.Go(func() error { return trigger.start(ctx, startTrigger, startDeploy) })
	g.Go(func() error { return deploy.start(ctx, startDeploy, startRelease, startTrigger) })
	g.Go(func() error { return release.start(ctx, startRelease) })
	<-ctx.Done()
	cancel()
	return g.Wait()
}

type workflow struct {
	log     logr.Logger
	version string
	name    string
	basedir string
	envs    envVars
	state   stateClient
}

func (w *workflow) execute(e *execStruct) (*envVar, error) {
	if e == nil {
		return nil, errNoExcution
	}
	cmd, err := e.prepare(w.basedir, w.envs)
	if err != nil {
		return nil, err
	}
	start := time.Now()
	output, err := cmd.CombinedOutput()
	status := runnerStatusDone
	duration := fmt.Sprintf("%dms", time.Since(start).Milliseconds())
	if err != nil {
		status = runnerStatusFailed
	}
	w.state.notifyStep(
		w.version,
		w.name,
		status,
		step{
			execStruct: *e,
			Name:       e.name,
			StartTime:  &start,
			Duration:   &duration,
			Variables:  w.envs,
		})
	results := strings.Split(string(output), "\n")
	for _, v := range results {
		w.log.Info(v)
	}
	if err != nil {
		return nil, err
	}
	if e.Output != "" {
		return &envVar{Name: e.Output, Value: results[0]}, nil
	}
	return nil, nil
}

func (w *workflow) start(e *execStruct) (*os.Process, error) {
	if e == nil {
		return nil, errNoExcution
	}
	cmd, err := e.prepare(w.basedir, w.envs)
	if err != nil {
		return nil, err
	}
	stdout := path.Join(w.basedir, fmt.Sprintf("log-%s.out", w.envs.get("version")))
	stderr := path.Join(w.basedir, fmt.Sprintf("err-%s.out", w.envs.get("version")))
	logWriter := &file{filename: stdout}
	errWriter := &file{filename: stderr}
	e.files = append(e.files, logWriter, errWriter)
	cmd.Stdout = logWriter
	cmd.Stderr = errWriter
	start := time.Now()
	err = cmd.Start()
	status := runnerStatusStarted
	w.state.notifyStep(
		w.version,
		w.name,
		status,
		step{
			execStruct: *e,
			Name:       e.name,
			StartTime:  &start,
			Variables:  w.envs,
		})
	return cmd.Process, err
}
