package pkg

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
)

type releaseWorkflow struct {
	releaseStruct
	log logr.Logger
}

func (w *releaseWorkflow) start(ctx context.Context, action <-chan string) error {
	log := w.log.WithName("release")
	for {
		select {
		case action := <-action:
			msg := fmt.Sprintf("release %s started...", action)
			log.Info(msg)
		case <-ctx.Done():
			return nil
		}
	}
}
