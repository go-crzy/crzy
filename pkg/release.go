package pkg

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
)

type releaseWorkflow struct {
	releaseStruct
	Log logr.Logger
}

func (w *releaseWorkflow) start(ctx context.Context, action <-chan string) error {
	for {
		select {
		case action := <-action:
			msg := fmt.Sprintf("release %s started...", action)
			w.Log.Info(msg)
		case <-ctx.Done():
			return nil
		}
	}
}
