package pkg

import (
	"context"
	"os"

	"golang.org/x/sync/errgroup"
)

const (
	triggeredMessage string = "triggered"
	deployedMessage  string = "deployed"
)

type event struct {
	id   string
	envs []envVar
}

func (r *runContainer) createAndStartWorkflows(ctx context.Context, git gitCommand, startTrigger chan event, switchUpstream func(string)) error {
	err := git.cloneRepository()
	if err != nil {
		r.Log.Error(err, "error cloning repository")
		return err
	}
	g, ctx := errgroup.WithContext(ctx)
	ctx, cancel := context.WithCancel(ctx)
	deploy := &deployWorkflow{
		deployStruct: r.Config.Deploy,
		workspace:    git.getWorkspace(),
		execdir:      git.getExecdir(),
		log:          r.Log,
		keys: map[string]execStruct{
			"install":   r.Config.Deploy.Install,
			"test":      r.Config.Deploy.Test,
			"pre_build": r.Config.Deploy.PreBuild,
			"build":     r.Config.Deploy.Build,
		},
		flow: []string{"install", "test", "pre_build", "build"},
	}
	trigger := &triggerWorkflow{
		triggerStruct: r.Config.Trigger,
		head:          r.Config.Main.Head,
		log:           r.Log,
		git:           git,
		command:       &defaultTriggerCommand{},
	}
	release := &releaseWorkflow{
		releaseStruct: r.Config.Release,
		log:           r.Log,
		execdir:       git.getExecdir(),
		keys: map[string]execStruct{
			"run": r.Config.Release.Run,
		},
		flow:           "run",
		processes:      map[string]*os.Process{},
		switchUpstream: switchUpstream,
	}
	startDeploy := make(chan event)
	defer close(startDeploy)
	startRelease := make(chan event)
	defer close(startRelease)
	g.Go(func() error { return trigger.start(ctx, startTrigger, startDeploy) })
	g.Go(func() error { return deploy.start(ctx, startDeploy, startRelease, startTrigger) })
	g.Go(func() error { return release.start(ctx, startRelease) })
	<-ctx.Done()
	cancel()
	return g.Wait()
}
