package pkg

import (
	"context"

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

func (r *runContainer) createAndStartWorkflows(ctx context.Context, git gitCommand, startTrigger chan event) error {
	err := git.cloneRepository()
	if err != nil {
		r.Log.Error(err, "error cloning repository")
		return err
	}
	g, ctx := errgroup.WithContext(ctx)
	ctx, cancel := context.WithCancel(ctx)
	deploy := &deployWorkflow{
		deployStruct: r.Config.Deploy,
		log:          r.Log,
	}
	trigger := &triggerWorkflow{
		trigger: r.Config.Trigger,
		head:    r.Config.Main.Head,
		log:     r.Log,
		git:     git,
		command: &defaultTriggerCommand{},
	}
	release := &releaseWorkflow{
		releaseStruct: r.Config.Release,
		log:           r.Log,
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
