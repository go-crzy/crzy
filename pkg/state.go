package pkg

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/go-logr/logr"
)

const (
	runnerStatusStarted = "started"
	runnerStatusFailed  = "failure"
	runnerStatusDone    = "success"
)

type syntheticWorkflow struct {
	Runners map[string]runner `json:"runners"`
	Version string            `json:"version"`
}

type runner struct {
	Steps  []step `json:"steps"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type step struct {
	execStruct
	Name      string     `json:"name"`
	StartTime *time.Time `json:"start_time,omitempty"`
	Duration  *string    `json:"duration,omitempty"`
	Variables []envVar   `json:"flow.envs,omitempty"`
}

type stepEvent struct {
	version        string
	workflow       string
	step           step
	workflowStatus string
}

type state interface {
	listVersions() []byte
	listVersionDetails(string) ([]byte, error)
	addStep(stepEvent)
}

type defaultState struct {
	sync.Mutex
	state map[string]syntheticWorkflow
}

type stateManager struct {
	notifier chan stepEvent
	state    state
	log      logr.Logger
}

type stateDefaultClient struct {
	notifier chan stepEvent
}

type stateClient interface {
	notifyStep(version, workflow, status string, step step)
}

type stateMockClient struct {
}

type dataVersion struct {
	Versions []string `json:"versions"`
}

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
		state: &defaultState{
			state: map[string]syntheticWorkflow{},
		},
		log: r.Log.WithName("state"),
	}
}

func (s *defaultState) addStep(stepEvent stepEvent) {
	s.Lock()
	defer s.Unlock()
	version, ok := s.state[stepEvent.version]
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
	s.state[stepEvent.version] = version
}

func (s *defaultState) listVersions() []byte {
	s.Lock()
	defer s.Unlock()
	data := dataVersion{
		Versions: []string{},
	}
	for k := range s.state {
		data.Versions = append(data.Versions, k)
	}
	output, _ := json.Marshal(&data)
	return output
}

var errNoVersion = errors.New("noVersion")

func (s *defaultState) listVersionDetails(version string) ([]byte, error) {
	s.Lock()
	defer s.Unlock()
	x, ok := s.state[version]
	if !ok {
		return []byte{}, errNoVersion
	}
	return json.Marshal(x)
}

func (w *stateManager) start(ctx context.Context) error {
	defer close(w.notifier)
	log := w.log
	log.Info("starting state manager...")
	for {
		select {
		case stepEvent := <-w.notifier:
			w.state.addStep(stepEvent)
		case <-ctx.Done():
			log.Info("stopping state manager...")
			return nil
		}
	}
}
