package pkg

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
)

type triggerWorkflow struct {
	trigger triggerStruct
	head    string
	log     logr.Logger
	command runner
	git     gitCommand
}

func (w *triggerWorkflow) start(ctx context.Context, action <-chan string, deploy chan<- string) error {
	log := w.log.WithName("trigger")
	firstsync := true
	deploying := false
	triggered := false
	for {
		select {
		case action := <-action:
			switch action {
			case triggeredMessage:
				log.Info("starting trigger...")
				triggered = true
				if !deploying {
					triggered = false
					if firstsync {
						firstsync = false
						continue
					}
					err := w.git.syncWorkspace(w.head)
					if err != nil {
						log.Error(err, "error during sync of the repository")
						continue
					}
					version, err := w.command.run()
					if err != nil {
						log.Error(err, "error during version of the repository")
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

type mockVersionAndSync struct {
	output bool
}

func (w *mockVersionAndSync) run() (string, error) {
	if w.output {
		return "1", nil
	}
	return "1", errors.New("error")
}
