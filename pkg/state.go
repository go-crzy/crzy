package pkg

import (
	"context"

	"github.com/go-logr/logr"
)

type StateMachine struct {
	action   chan func()
	state    string
	log      logr.Logger
	upstream Upstream
}

func NewStateMachine(upstream Upstream) *StateMachine {
	return &StateMachine{
		state:    "initial",
		action:   make(chan func()),
		log:      NewLogger("machine"),
		upstream: upstream,
	}
}

func (m *StateMachine) Run(ctx context.Context) error {
	log := m.log
	log.Info("starting state machine....")
	for {
		select {
		case f := <-m.action:
			log.Info("action captured, ready to run....")
			f()
		case <-ctx.Done():
			log.Info("stopping state machine....")
			m.upstream.KillAll()
			close(m.action)
			return ctx.Err()
		}
	}
}
