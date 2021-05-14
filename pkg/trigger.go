package pkg

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
)

type triggerWorkflow struct {
	triggerStruct
	log     logr.Logger
	command runner
	git     gitCommand
}

func (w *triggerWorkflow) start(ctx context.Context, action <-chan string, deploy chan<- string) error {
	log := w.log.WithName("trigger")
	deploying := false
	triggered := false
	for {
		select {
		case action := <-action:
			switch action {
			case triggeredMessage:
				triggered = true
				if !deploying {
					triggered = false
					version, err := w.command.run()
					if err != nil {
						log.Error(err, "error during sync/version of the repository")
						continue
					}
					// TODO: check the version does not exist yet, if it does not kick off the deploy
					log.Info("version computed, deploying now...", "data", version)
					deploying = true
					deploy <- triggeredMessage
				}
			case deployedMessage:
				deploying = false
				if triggered {
					triggered = false
					_, err := w.command.run()
					if err != nil {
						log.Error(err, "error during sync/version of the repository")
						continue
					}
					// TODO: check the version does not exist yet, if it does not kick off the deploy
					deploying = true
					deploy <- triggeredMessage
				}
			}
		case <-ctx.Done():
			return nil
		}
	}
}

type runner interface {
	run() (string, error)
}

type DefaultVersionAndSync struct{}

func (d *DefaultVersionAndSync) run() (string, error) {
	return "1", nil
}

type MockVersionAndSyncSucceed struct{}

func (w *MockVersionAndSyncSucceed) run() (string, error) {
	return "1", nil
}

type MockVersionAndSyncFailed struct{}

func (w *MockVersionAndSyncFailed) run() (string, error) {
	return "", errors.New("error")
}
