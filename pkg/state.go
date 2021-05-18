package pkg

import (
	"context"
	"time"

	"github.com/go-logr/logr"
)

const runnerStatusStarted = "started"
const runnerStatusDone = "success"

type syntheticWorkflow struct {
	Runners map[string]runner
	Version string
}

type runner struct {
	Steps  []step
	Name   string
	Status string
}

type step struct {
	execStruct
	Name      string
	StartTime *time.Time
	Duration  *time.Duration
}

type stepEvent struct {
	version        string
	workflow       string
	step           step
	workflowStatus string
}

type stateManager struct {
	notifier stateNotifier
	state    map[string]syntheticWorkflow
	log      logr.Logger
}

type stateDefaultClient struct {
	notifier stateNotifier
}

type stateClient interface {
	notifyStep(version, workflow, status string, step step)
}

type stateMockClient struct {
	notifier stateNotifier
}

type stateNotifier chan stepEvent

func (a *stateDefaultClient) notifyStep(version, workflow, status string, step step) {
	a.notifier <- stepEvent{
		version:        version,
		workflow:       workflow,
		step:           step,
		workflowStatus: status,
	}
}

func (a *stateMockClient) notifyStep(version, workflow, status string, step step) {
}

func (r *runContainer) newStateManager() *stateManager {
	return &stateManager{
		notifier: make(chan stepEvent),
		state:    map[string]syntheticWorkflow{},
		log:      r.Log.WithName("state"),
	}
}

func (w *stateManager) start(ctx context.Context) error {
	defer close(w.notifier)
	log := w.log
	log.Info("starting state manager...")
	for {
		select {
		case stepEvent := <-w.notifier:
			version, ok := w.state[stepEvent.version]
			if !ok {
				version = syntheticWorkflow{
					Runners: map[string]runner{},
					Version: stepEvent.version,
				}
			}
			workflow, ok := version.Runners[stepEvent.workflow]
			if !ok {
				workflow = runner{
					Steps:  []step{},
					Name:   stepEvent.workflow,
					Status: stepEvent.workflowStatus,
				}
			}
			workflow.Status = stepEvent.workflowStatus
			workflow.Steps = append(workflow.Steps, stepEvent.step)
			version.Runners[stepEvent.workflow] = workflow
			w.state[stepEvent.version] = version
		case <-ctx.Done():
			log.Info("stopping state manager...")
			return nil
		}
	}
}
