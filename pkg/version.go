package pkg

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
)

type versionSync struct {
	versionStruct
	log     logr.Logger
	command runner
}

func (w *versionSync) start(ctx context.Context, action <-chan string, deploy chan<- string) error {
	log := w.log.WithName("version")
	deploying := false
	triggered := false
	for {
		select {
		case action := <-action:
			switch action {
			case triggeredMessage:
				triggered = true
				if !deploying {
					_, err := w.command.run()
					if err != nil {
						log.Error(err, "error during sync/version of the repository")
						continue
					}
					// TODO: check the version does not exist yet, if it does not kick off the deploy
					deploying = true
					deploy <- versionedMessage
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
					deploy <- versionedMessage
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
