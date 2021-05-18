package pkg

import (
	"context"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
)

func Test_releaseWorkflow(t *testing.T) {
	release := &releaseWorkflow{
		log: &mockLogger{},
		releaseStruct: releaseStruct{
			PortRange: portRangeStruct{
				Min: 8090,
				Max: 8100,
			},
		},
		state: &stateMockClient{},
	}
	g, ctx := errgroup.WithContext(context.TODO())
	ctx, cancel := context.WithCancel(ctx)
	startRelease := make(chan event)
	defer close(startRelease)
	g.Go(func() error { return release.start(ctx, startRelease) })
	startRelease <- event{
		id: deployedMessage,
		envs: []envVar{
			{Name: "version", Value: "version"},
		}}
	time.Sleep(200 * time.Millisecond)
	cancel()
	if err := g.Wait(); err != nil && err.Error() != "context cancel" {
		t.Error("should receive a context cancel message")
	}
}
