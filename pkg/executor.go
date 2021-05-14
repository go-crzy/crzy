package pkg

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
)

type returnStruct struct {
	JobID string
	Err   error
	Envs  []envVar
}

type actionStruct struct {
	F     func() ([]envVar, error)
	JobID string
}

type executor struct {
	uuid          string
	actionChannel chan actionStruct
	returnChannel chan returnStruct
	log           logr.Logger
}

func (r *runContainer) createAndStartExecutor(ctx context.Context, actionChannel chan actionStruct, returnChannel chan returnStruct) error {
	e := executor{
		uuid:          fmt.Sprintf("%d", time.Now().UnixNano()),
		actionChannel: actionChannel,
		returnChannel: returnChannel,
		log:           r.Log.WithName("executor"),
	}
	return e.start(ctx)
}

func (e *executor) start(ctx context.Context) error {
	log := e.log
	log.Info("starting executor uuid:" + e.uuid + "...")
	for {
		select {
		case action := <-e.actionChannel:
			log.Info("action captured, executor uuid:" + e.uuid + "...")
			envs, err := action.F()
			if err != nil {
				e.returnChannel <- returnStruct{JobID: action.JobID, Err: err}
				continue
			}
			e.returnChannel <- returnStruct{JobID: action.JobID, Envs: envs}
		case <-ctx.Done():
			log.Info("stopping executor uuid:" + e.uuid + "...")
			close(e.actionChannel)
			return ctx.Err()
		}
	}
}
