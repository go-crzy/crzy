package pkg

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
)

type deployWorkflow struct {
	deployStruct
	log logr.Logger
}

func (w *deployWorkflow) start(ctx context.Context, action <-chan string, release, trigger chan<- string) error {
	log := w.log.WithName("deploy")
	for {
		select {
		case action := <-action:
			msg := fmt.Sprintf("action %s started...", action)
			log.Info(msg)
			release <- deployedMessage
			trigger <- deployedMessage
		case <-ctx.Done():
			return nil
		}
	}
}
