package pkg

import (
	"context"

	"github.com/go-logr/logr"
)

type releaseWorkflow struct {
	releaseStruct
	execdir string
	log     logr.Logger
	keys    map[string]execStruct
	flow    []string
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
				for _, v := range w.flow {
					cmd := w.keys[v]
					cmd.log = log
					if cmd.Command == "" {
						continue
					}
					log.Info("running...", "data", v)
					_, err := (&cmd).run(w.execdir, m)
					if err != nil {
						log.Error(err, "execution error")
						continue OUTER
					}
				}
				log.Info("release execution succeeded...")
			}
		case <-ctx.Done():
			return nil
		}
	}
}
