package pkg

import (
	"bytes"
	"context"
	"os"
	"testing"

	log "github.com/go-crzy/crzy/logr"
)

func Test_argsParser(t *testing.T) {
	p := &argsParser{}
	os.Args = []string{"crzy", "-repository", "color.git", "-color"}
	a := p.parse()
	if a.colorize != true {
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
		parser: &mockParser{
			version: true,
		},
		out: buf,
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
		parser: &mockParser{
			colorize: true,
		},
	}
	err := c.Run(context.TODO())
	if err != errLoadingConfigFile {
		t.Error("should get errLoadingConfigFile, current:", err)
	}
}

func Test_Run_and_succeed(t *testing.T) {
	log := &log.MockLogger{}
	c := &DefaultRunner{
		log: log,
		parser: &mockParser{
			configFile: defaultConfigFile,
			colorize:   true,
		},
	}
	ctx, cancel := context.WithCancel(context.TODO())
	go c.Run(ctx)
	cancel()
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
