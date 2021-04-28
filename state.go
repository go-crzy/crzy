package crzy

import (
	"context"
	"log"
)

type StateMachine struct {
	action chan func()
	state  string
}

func NewStateMachine() *StateMachine {
	return &StateMachine{
		state:  "initial",
		action: make(chan func()),
	}
}

func (m *StateMachine) Run(ctx context.Context) error {
	log.Println("starting state machine....")
	for {
		select {
		case f := <-m.action:
			log.Println("action captured, ready to run....")
			f()
		case <-ctx.Done():
			log.Println("stopping state machine....")
			close(m.action)
			return ctx.Err()
		}
	}
}
