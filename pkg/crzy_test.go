package pkg

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"testing"

	log "github.com/go-crzy/crzy/logr"
	"golang.org/x/sync/errgroup"
)

func Test_argsParser(t *testing.T) {
	p := &argsParser{}
	os.Args = []string{"crzy", "-repository", "color.git", "-nocolor"}
	a := p.parse()
	if a.nocolor != true {
		t.Error("args not parsed as expected")
	}
}

func Test_NewCrzy(t *testing.T) {
	c := NewCrzy()
	if c == nil {
		t.Error("should return a DefaultRunner")
	}
}

func Test_uninitilizedRunner(t *testing.T) {
	c := &DefaultRunner{}
	err := c.Run(context.TODO())
	if err != ErrWronglyInitialized {
		t.Error("should return ErrWronglyInitialized, current:", err)
	}
}

func Test_displayVersion(t *testing.T) {
	log := &log.MockLogger{}
	buf := new(bytes.Buffer)
	c := &DefaultRunner{
		log: log,
		container: &defaultContainer{
			out: buf,
			parser: &mockParser{
				version: true,
			},
		},
	}
	c.Run(context.TODO())
	if buf.String() != "crzy version dev(unknown)\n" {
		t.Error("should return version, current:", buf.String())
	}
}

func Test_Run_and_fails(t *testing.T) {
	log := &log.MockLogger{}
	c := &DefaultRunner{
		log: log,
		container: &defaultContainer{
			log: log,
			out: io.Discard,
			parser: &mockParser{
				version: true,
			},
		},
	}
	err := c.Run(context.TODO())
	if err != ErrVersionRequested {
		t.Error("should get ErrWronglyInitialized, current:", err)
	}
}

func Test_Run_and_succeed(t *testing.T) {
	log := &log.MockLogger{}
	c := &DefaultRunner{
		log: log,
		container: &defaultContainer{
			log: log,
			parser: &mockParser{
				configFile: defaultConfigFile,
				version:    false,
			},
		},
	}
	g, ctx := errgroup.WithContext(context.TODO())
	ctx, cancel := context.WithCancel(ctx)
	g.Go(func() error { return c.Run(ctx) })
	cancel()
	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		t.Error("should simply succeed", err)
	}
}

func Test_Heading(t *testing.T) {
	log := &log.MockLogger{}
	heading(log)
	if len(log.Logs) != 5 {
		t.Error("should return 5 rows")
	}
	if log.Logs[0] != "" {
		t.Error("first line should be empty")
	}
}

var containerData = []string{"load", "store", "git", "gitserver", "proxy", "api"}

func Test_new_with_mock_runner_and_fail(t *testing.T) {
	log := &log.MockLogger{}
	for _, v := range containerData {
		r := &DefaultRunner{
			log: log,
			container: &mockContainer{
				step: v,
			},
		}
		t.Log(v)
		err := r.Run(context.TODO())
		if err == nil || err.Error() != v {
			t.Error("be empty:", err)
		}
	}
}
