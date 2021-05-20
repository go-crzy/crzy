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
OUTER:
	for {
		select {
		case action := <-action:
			log.Info("deploy started...")
			switch action.id {
			case triggeredMessage:
				vars := []envVar{}
				vars = append(vars, action.envs...)
				m, err := groupEnvs(vars...)
				if err != nil {
					log.Error(err, "could not map envs")
					trigger <- event{id: deployedMessage}
					continue OUTER
				}
				artifactDirectory, err := replaceEnvs(w.Artifact.Directory, m)
				if err != nil {
					log.Error(err, "could not transform directory")
					trigger <- event{id: deployedMessage}
					continue OUTER
				}
				artifactDirectory = path.Join(w.execdir, artifactDirectory)
				err = os.MkdirAll(artifactDirectory, os.ModeDir|os.ModePerm)
				if err != nil {
					log.Error(err, "could not create directory", "data", artifactDirectory)
					trigger <- event{id: deployedMessage}
					continue OUTER
				}
				vars = append(vars, envVar{Name: "artifactDirectory", Value: artifactDirectory})
				m["artifactDirectory"] = artifactDirectory
				artifactFilename, err := replaceEnvs(w.Artifact.Filename+w.Artifact.Extension, m)
				if err != nil {
					log.Error(err, "could not transform filename")
					trigger <- event{id: deployedMessage}
					continue OUTER
				}
				vars = append(vars, envVar{Name: "artifactFilename", Value: artifactFilename})
				m["artifactFilename"] = artifactFilename
				artifact := path.Join(artifactDirectory, artifactFilename)
				vars = append(vars, envVar{Name: "artifact", Value: artifact})
				m["artifact"] = artifact
				for _, v := range w.flow {
					cmd := w.keys[v]
					cmd.log = log
					if cmd.Command == "" {
						continue
					}
					log.Info("running...", "data", v)
					e, err := (&cmd).run(w.workspace, m)
					if err != nil {
						w.state.notifyStep(
							getEnv(action.envs, "version"), "deploy",
							runnerStatusFailed,
							step{execStruct: cmd, Name: v})
						log.Error(err, "execution error")
						trigger <- event{id: deployedMessage}
						continue OUTER
					}
					w.state.notifyStep(
						getEnv(action.envs, "version"), "deploy",
						runnerStatusDone,
						step{execStruct: cmd, Name: v})
					if e != nil {
						vars = append(vars, *e)
					}
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
