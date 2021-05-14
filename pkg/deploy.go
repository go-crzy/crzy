package pkg

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
)

type deployWorkflow struct {
	deployStruct
	Log logr.Logger
}

func (w *deployWorkflow) start(ctx context.Context, action <-chan string, release, version chan<- string) error {
	for {
		select {
		case action := <-action:
			msg := fmt.Sprintf("action %s started...", action)
			w.Log.Info(msg)
			release <- deployedMessage
			version <- deployedMessage
		case <-ctx.Done():
			return nil
		}
	}
}
