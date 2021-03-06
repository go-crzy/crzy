package pkg

import (
	"context"
	"errors"
	"testing"
	"time"

	log "github.com/go-crzy/crzy/logr"
	"golang.org/x/sync/errgroup"
)

type mockState struct{}

func (m *mockState) listVersions() []byte {
	return []byte(`{"versions": ["123"]}`)
}

func (m *mockState) listVersionDetails(version string) ([]byte, error) {
	if version == "fail" {
		return nil, errors.New("error")
	}
	return []byte(`{"runners": {"deploy": {} }}`), nil
}

func (m *mockState) addStep(stepEvent) {

}

func (s *mockState) logVersion(version, file string) ([]byte, error) {
	if version == "fail" {
		return nil, errors.New("error")
	}
	return []byte("line1\nline2"), nil
}

func (s *mockState) getConfiguration() []byte {
	return []byte(`{"head": "main"}`)
}

func Test_newStateManager(t *testing.T) {
	r := &defaultContainer{
		config: &config{
			Main: mainStruct{
				Head: "main",
			},
		},
		log: &log.MockLogger{},
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
		t.Error("should return expected message, current:", string(data))
	}
	cancel()
	if err := g.Wait(); err != nil && err.Error() != "context cancel" {
		t.Error("should receive a context cancel message")
	}
}

func Test_listVersionsDetails_succeed(t *testing.T) {
	r := defaultState{
		state: map[string]syntheticWorkflow{
			"abc": {
				Version: "abc",
				Runners: map[string]runner{
					"deploy": {
						Steps: []step{
							{
								execStruct: execStruct{
									Command: "go",
									Args:    []string{"test", "./..."},
									WorkDir: ".",
								},
								Name: "test",
							},
						},
						Name:   "deploy",
						Status: "succeeded",
					},
				},
			},
		},
	}
	data, err := r.listVersionDetails("abc")
	if err != nil {
		t.Error("should succeed; error:", err)
	}
	if string(data) != `{"version":"abc","workflows":[{"steps":[{"command":"go","args":["test","./..."],"workdir":".","name":"test"}],"name":"deploy","status":"succeeded"}]}` {
		t.Error("error, current message is: ", string(data))
	}
}

func Test_listVersionsDetails_fail(t *testing.T) {
	r := defaultState{
		state: map[string]syntheticWorkflow{
			"abc": {
				Version: "abc",
			},
		},
	}
	_, err := r.listVersionDetails("def")
	if err != errNoVersion {
		t.Error("should fail with errNoVersion; error:", err)
	}
}

func Test_logVersion_success(t *testing.T) {
	r := defaultState{
		state: map[string]syntheticWorkflow{
			"abc": {
				Version: "abc",
				Runners: map[string]runner{
					"release": {
						Steps: []step{
							{
								execStruct: execStruct{
									log:   &log.MockLogger{},
									files: []*file{},
								},
								Name: "install",
							},
						},
						Name:   "release",
						Status: "success",
					},
				},
			},
		},
	}
	_, err := r.logVersion("abc", "log")
	if err != errNoLogfile {
		t.Error("should fail with errNoVersion; error:", err)
	}
	_, err = r.logVersion("abc", "err")
	if err != errNoLogfile {
		t.Error("should fail with errNoVersion; error:", err)
	}
}
