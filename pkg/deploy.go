package pkg

import (
	"context"
	"os"
	"path"

	"github.com/go-logr/logr"
)

type deployWorkflow struct {
	deployStruct
	workspace string
	execdir   string
	log       logr.Logger
	keys      map[string]execStruct
	flow      []string
	state     stateClient
}

func (w *deployWorkflow) start(ctx context.Context, action <-chan event, release, trigger chan<- event) error {
	log := w.log.WithName("deploy")
	for {
		select {
		case action := <-action:
			log.Info("deploy started...")
			switch action.id {
			case triggeredMessage:
				vars := newEnvVars(action.envs...)
				if _, err := vars.toMap(); err != nil {
					log.Error(err, "could not map envs")
					trigger <- event{id: deployedMessage}
					continue
				}
				artifactDirectory, err := vars.replace(w.Artifact.Directory)
				if err != nil {
					log.Error(err, "could not transform directory")
					trigger <- event{id: deployedMessage}
					continue
				}
				artifactDirectory = path.Join(w.execdir, artifactDirectory)
				if err := os.MkdirAll(artifactDirectory, os.ModeDir|os.ModePerm); err != nil {
					log.Error(err, "could not create directory", "data", artifactDirectory)
					trigger <- event{id: deployedMessage}
					continue
				}
				vars.addOne("artifactDirectory", artifactDirectory)
				artifactFilename, err := vars.replace(w.Artifact.Filename + w.Artifact.Extension)
				if err != nil {
					log.Error(err, "could not transform filename")
					trigger <- event{id: deployedMessage}
					continue
				}
				vars.addOne("artifactFilename", artifactFilename)
				artifact := path.Join(artifactDirectory, artifactFilename)
				vars.addOne("artifact", artifact)

				if err := w.startFlows(action, &vars); err != nil {
					trigger <- event{id: deployedMessage}
					continue
				}
				log.Info("deploy execution succeeded...")
				release <- event{id: deployedMessage, envs: vars}
				trigger <- event{id: deployedMessage}
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (w *deployWorkflow) startFlows(action event, vars *envVars) error {
	log := w.log.WithName("deploy")
	for _, v := range w.flow {
		cmd := w.keys[v]
		cmd.log = log
		if cmd.Command == "" {
			continue
		}
		log.Info("running...", "data", v)
		e, err := cmd.run(w.workspace, *vars)
		if err != nil {
			return err
		}
		w.state.notifyStep(action.envs.get("version"), "deploy", runnerStatusDone, step{execStruct: cmd, Name: v})
		if e != nil {
			vars.add(*e)
		}
	}
	return nil
}
