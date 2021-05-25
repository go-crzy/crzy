package pkg

import (
	"context"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
)

type mockState struct{}

func (m *mockState) listVersions() []byte {
	return []byte(`{"versions": ["123"]}`)
}

func (m *mockState) listVersionDetails(version string) ([]byte, error) {
	return []byte(`{"runners": {"deploy": {} }}`), nil
}

func (m *mockState) addStep(stepEvent) {

}

func Test_newStateManager(t *testing.T) {
	r := &runContainer{
		Log: &mockLogger{},
	}
	v := r.newStateManager()
	stateClient := &stateDefaultClient{
		notifier: v.notifier,
	}
	g, ctx := errgroup.WithContext(context.TODO())
	ctx, cancel := context.WithCancel(ctx)
	g.Go(func() error { return v.start(ctx) })
	go func() {
		stateClient.notifyStep("123", "trigger", runnerStatusDone, step{
			execStruct: execStruct{
				Command: "version",
			},
		})
	}()
	go func() {
		stateClient.notifyStep("123", "deploy", runnerStatusDone, step{
			execStruct: execStruct{
				Command: "ant",
				Args: []string{
					"build",
				},
				WorkDir: ".",
			},
		})
	}()
	time.Sleep(200 * time.Millisecond)
	data := v.state.listVersions()
	if string(data) != `{"versions":["123"]}` {
		t.Error("We are fucked", string(data))
	}
	cancel()
	if err := g.Wait(); err != nil && err.Error() != "context cancel" {
		t.Error("should receive a context cancel message")
	}
}
