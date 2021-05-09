package pkg

import (
	"context"
	"testing"

	cron "github.com/robfig/cron/v3"
)

func Test_NewCronService(t *testing.T) {
	v := NewCronService()
	if v == nil {
		t.Error("should not be nil")
		t.FailNow()
	}
	if v.log == nil || !v.log.Enabled() {
		t.Error("log should be enabled")
	}
	if v.cron == nil {
		t.Error("cron should not be empty")
	}
}

func Test_RunCronService_and_succeed(t *testing.T) {
	s := &CronService{
		cron: cron.New(),
		log:  NewLogger("test"),
	}
	c, cancel := context.WithCancel(context.Background())
	charErr := make(chan error)
	go func() { charErr <- s.Run(c) }()
	cancel()
	err := <-charErr
	if err == nil || err.Error() != "context canceled" {
		t.Error("there should be a context canceled; error:", err)
	}
}
