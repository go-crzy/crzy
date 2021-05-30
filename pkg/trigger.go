package pkg

import (
	"context"
	"errors"
	"path"
	"strings"

	"github.com/go-logr/logr"
)

var errWrongVersionOutput error = errors.New("wrongversion")

type triggerWorkflow struct {
	triggerStruct
	head    string
	log     logr.Logger
	git     gitCommand
	command triggerCommand
	state   stateClient
}

func (w *triggerWorkflow) start(ctx context.Context, action <-chan event, deploy chan<- event) error {
	log := w.log.WithName("trigger")
	firstsync := true
	deploying := false
	triggered := false
	command := w.command
	command.setTriggerWorkflow(w)
	for {
		select {
		case action := <-action:
			switch action.id {
			case triggeredMessage:
				log.Info("starting trigger...")
				triggered = true
				if !deploying {
					triggered = false
					err := w.git.syncWorkspace(w.head)
					if err != nil {
						switch firstsync {
						case true:
							firstsync = false
							log.Info("cannot sync on first capture. do not deploy...")
						default:
							log.Error(err, "error during sync of the repository")
						}
						continue
					}
					firstsync = false
					version, err := command.version()
					if err != nil {
						log.Error(err, "error during version of the repository")
						continue
					}
					w.state.notifyStep(
						version, "trigger",
						runnerStatusDone,
						step{execStruct: execStruct{Command: "version"}, Name: "version"})
					// TODO: check the version does not exist yet, if it does not kick off the deploy
					log.Info("version computed, deploying now...", "data", version)
					deploying = true
					deploy <- event{id: triggeredMessage, envs: []envVar{{Name: "version", Value: version}}}
				}
			case deployedMessage:
				deploying = false
				if triggered {
					triggered = false
					err := w.git.syncWorkspace(w.head)
					if err != nil {
						log.Error(err, "error during sync of the repository")
						continue
					}
					version, err := command.version()
					if err != nil {
						log.Error(err, "error during version of the repository")
						continue
					}
					// TODO: check the version does not exist yet, if it does not kick off the deploy
					log.Info("version computed, deploying now...", "data", version)
					deploying = true
					deploy <- event{id: triggeredMessage, envs: []envVar{{Name: "version", Value: version}}}
				}
			}
		case <-ctx.Done():
			return nil
		}
	}
}

type triggerCommand interface {
	version() (string, error)
	setTriggerWorkflow(*triggerWorkflow)
}

type defaultTriggerCommand struct {
	trigger *triggerWorkflow
}

func (d *defaultTriggerCommand) setTriggerWorkflow(w *triggerWorkflow) {
	d.trigger = w
}

func (d *defaultTriggerCommand) version() (string, error) {
	log := d.trigger.log
	if d.trigger.Version.Command == "" {
		output, err := getCmd(d.trigger.git.getWorkspace(), envVars{}, d.trigger.git.getBin(), "log", "--format=%H", "-1", ".").CombinedOutput()
		if err != nil {
			log.Error(err, "could not get macro version")
			return "", err
		}
		if len(output) < 16 {
			return "", errWrongVersionOutput
		}
		return string(output[0:16]), nil
	}
	workdir := path.Join(d.trigger.git.getWorkspace(), d.trigger.Version.WorkDir)
	output, err := getCmd(workdir, envVars{}, d.trigger.Version.Command, d.trigger.Version.Args...).CombinedOutput()
	if err != nil {
		log.Error(err, "could not get execution version")
		return "", err
	}
	result := strings.Split(string(output), "\n")[0]
	return result, nil
}
