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

func (r *runContainer) createAndStartWorkflows(ctx context.Context, conf config, startTrigger <-chan string) error {
	g, ctx := errgroup.WithContext(ctx)
	ctx, cancel := context.WithCancel(ctx)
	deploy := &deployWorkflow{
		deployStruct: conf.Deploy,
		Log:          r.Log,
	}
	trigger := &triggerWorkflow{
		triggerStruct: conf.Trigger,
		Log:           r.Log,
	}
	release := &releaseWorkflow{
		releaseStruct: conf.Release,
		Log:           r.Log,
	}
	version := &versionSync{
		versionStruct: conf.Trigger.Version,
		Log:           r.Log,
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
