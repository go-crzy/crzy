package pkg

import (
	"context"
	"testing"

	log "github.com/go-crzy/crzy/logr"
)

func Test_NewCrzy(t *testing.T) {
	args := Args{}
	c, _ := NewCrzy(args)
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
	_, err := NewCrzy(Args{Version: true})
	if err == nil {
		t.Error("return an error, current:", err)
	}
}

func Test_Run_and_succeed(t *testing.T) {
	log := &log.MockLogger{}
	c := &DefaultRunner{
		log: log,
		container: &defaultContainer{
			log: log,
		},
	}
	err := c.Run(context.TODO())
	if err == nil {
		t.Error("should simply fail")
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
