package pkg

import (
	"context"
	"os"

	"github.com/go-logr/logr"
)

type releaseWorkflow struct {
	releaseStruct
	execdir        string
	log            logr.Logger
	keys           map[string]execStruct
	files          map[string][]*file
	flow           string
	processes      map[string]*os.Process
	switchUpstream func(string)
	state          stateClient
	slack          *slackNotifier
}

func deepCopy(e execStruct) execStruct {
	output := execStruct{
		log:     e.log,
		name:    e.name,
		Command: e.Command,
		Args:    []string{},
		WorkDir: e.WorkDir,
		Envs:    envVars{},
		Output:  e.Output,
		files:   e.files,
	}
	output.Args = append(output.Args, e.Args...)
	for _, v := range e.Envs {
		env := envVar{
			Name:  v.Name,
			Value: v.Value,
		}
		output.Envs = append(output.Envs, env)
	}
	return output
}

func (w *releaseWorkflow) start(ctx context.Context, action <-chan event) error {
	log := w.log.WithName("release")
	port, err := createPortSequence(w.PortRange.Min, w.PortRange.Max)
	if err != nil {
		log.Error(err, "could not start due to unvailable ports")
		return err
	}
	for {
		select {
		case action := <-action:
			log.Info("release started...")
			switch action.id {
			case deployedMessage:
				vars := newEnvVars(action.envs...)
				p, err := port.getPort()
				if err != nil {
					log.Error(err, "could not reserve port")
					continue
				}
				cmd := deepCopy(w.keys[w.flow])
				cmd.log = log
				if cmd.Command == "" {
					continue
				}
				err = w.switchProcesses(p, cmd, vars)
				if err != nil {
					log.Error(err, "execution error")
					continue
				}
				// TODO: use sendMessage()
				// - if success : send a message on slack to say `command.Command` has started on `port`
				// - if failure, send a message on slack to say `command.Command` could not start
				log.Info("release execution succeeded...")
			}
		case <-ctx.Done():
			w.killAll()
			return nil
		}
	}
}

func (r *releaseWorkflow) killAll() error {
	for k, v := range r.processes {
		err := v.Kill()
		if err != nil {
			return err
		}
		delete(r.processes, k)
		delete(r.files, k)
	}
	return nil
}

func (r *releaseWorkflow) switchProcesses(port string, command execStruct, envs envVars) error {
	envs.addOne("port", port)
	workflow := &workflow{
		log:     r.log,
		version: envs.get("version"),
		name:    "release",
		basedir: r.execdir,
		envs:    envs,
		state:   r.state,
	}
	process, err := workflow.start(&command)
	if err != nil {
		return err
	}
	r.processes[port] = process
	r.files[port] = command.files
	r.switchUpstream("localhost:" + port)
	for k, v := range r.processes {
		if k != port {
			err := v.Kill()
			if err != nil {
				return err
			}
			delete(r.processes, k)
			delete(r.files, k)
		}
	}
	return nil
}
