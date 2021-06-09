package pkg

import (
	"context"
	"os"
	"runtime"
	"testing"
	"time"

	log "github.com/go-crzy/crzy/logr"
	"golang.org/x/sync/errgroup"
)

func Test_releaseWorkflow(t *testing.T) {
	release := &releaseWorkflow{
		log: &log.MockLogger{},
		releaseStruct: releaseStruct{
			PortRange: portRangeStruct{
				Min: 8090,
				Max: 8100,
			},
			Run: execStruct{
				log:     &log.MockLogger{},
				Command: "tail",
				Args:    []string{"-f", "config.go"},
				WorkDir: ".",
				Envs:    []envVar{}},
		},
		state:          &stateMockClient{},
		flow:           "run",
		switchUpstream: func(string) {},
		processes:      map[string]*os.Process{},
		files:          make(map[string][]*file),
	}
	if runtime.GOOS == "windows" {
		release.Run.Command = "powershell"
		release.Run.Args = []string{"-Command", "Get-Content config.go -Wait"}
	}
	release.keys = map[string]execStruct{"run": release.releaseStruct.Run}
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

func Test_deepCopy(t *testing.T) {
	input := execStruct{
		log:     &log.MockLogger{},
		name:    "release",
		Command: "go",
		Args:    []string{"run", "."},
		WorkDir: ".",
		Envs:    envVars{envVar{Name: "ADDR", Value: "localhost:${port}"}},
		files:   []*file{{filename: "log-1.out"}, {filename: "err-1.out"}},
	}
	o := deepCopy(input)
	if len(o.Envs) == 0 || o.Envs[0].Value != "localhost:${port}" {
		t.Error("should have one unsubtituted value")
	}
}
