package pkg

import (
	"context"

	"golang.org/x/sync/errgroup"
)

const (
	triggeredMessage string = "triggered"
	deployedMessage  string = "deployed"
	versionedMessage string = "versioned"
)

func (r *runContainer) createAndStartWorkflows(ctx context.Context, git gitCommand, startTrigger <-chan string) error {
	g, ctx := errgroup.WithContext(ctx)
	ctx, cancel := context.WithCancel(ctx)
	deploy := &deployWorkflow{
		deployStruct: r.Config.Deploy,
		log:          r.Log,
	}
	trigger := &triggerWorkflow{
		triggerStruct: r.Config.Trigger,
		log:           r.Log,
		git:           git,
	}
	release := &releaseWorkflow{
		releaseStruct: r.Config.Release,
		log:           r.Log,
	}
	version := &versionSync{
		versionStruct: r.Config.Trigger.Version,
		log:           r.Log,
		command:       &MockVersionAndSyncSucceed{},
	}
	startVersion := make(chan string)
	defer close(startVersion)
	startDeploy := make(chan string)
	defer close(startDeploy)
	startRelease := make(chan string)
	defer close(startRelease)
	g.Go(func() error { return deploy.start(ctx, startDeploy, startRelease, startVersion) })
	g.Go(func() error { return trigger.start(ctx, startTrigger, startVersion) })
	g.Go(func() error { return release.start(ctx, startRelease) })
	g.Go(func() error { return version.start(ctx, startVersion, startDeploy) })
	<-ctx.Done()
	cancel()
	return g.Wait()
}
