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
	flow           string
	processes      map[string]*os.Process
	switchUpstream func(string)
}

func (w *releaseWorkflow) start(ctx context.Context, action <-chan event) error {
	log := w.log.WithName("release")
	port, err := createPortSequence(w.PortRange.Min, w.PortRange.Max)
	if err != nil {
		log.Error(err, "could not start due to unvailable ports")
		return err
	}
OUTER:
	for {
		select {
		case action := <-action:
			log.Info("release started...")
			switch action.id {
			case deployedMessage:
				vars := []envVar{}
				vars = append(vars, action.envs...)
				p, err := port.getPort()
				if err != nil {
					log.Error(err, "could not reserve port")
					continue OUTER
				}
				vars = append(vars, envVar{Name: "port", Value: p})
				m, err := groupEnvs(vars...)
				if err != nil {
					log.Error(err, "could not map envs")
					continue OUTER
				}
				cmd := w.keys[w.flow]
				cmd.log = log
				if cmd.Command == "" {
					continue
				}
				err = w.switchProcesses(p, cmd, m)
				if err != nil {
					log.Error(err, "execution error")
					continue OUTER
				}
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
	}
	return nil
}

func (r *releaseWorkflow) switchProcesses(port string, command execStruct, envs map[string]string) error {
	process, err := command.runBackground(r.execdir, envs)
	if err != nil {
		return err
	}
	r.processes[port] = process
	r.switchUpstream(port)
	for k, v := range r.processes {
		if k != port {
			err := v.Kill()
			if err != nil {
				return err
			}
			delete(r.processes, k)
		}
	}
	return nil
}
