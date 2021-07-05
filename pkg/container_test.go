package pkg

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"

	log "github.com/go-crzy/crzy/logr"
)

type mockContainer struct {
	step string
}

func (m *mockContainer) getConf(args Args) error {
	if m.step == "load" {
		return errors.New("load")
	}
	return nil
}

func (m *mockContainer) createStore() (*store, error) {
	if m.step == "store" {
		return nil, errors.New("store")
	}
	return &store{
		log:     &log.MockLogger{},
		rootDir: "unknown",
	}, nil
}

func (m *mockContainer) newStateManager() *stateManager {
	return &stateManager{
		log:   &log.MockLogger{},
		state: &defaultState{},
	}
}
func (m *mockContainer) newDefaultGitCommand(store store) (gitCommand, error) {
	if m.step == "git" {
		return nil, errors.New("git")
	}
	return &defaultGitCommand{}, nil
}

func (m *mockContainer) newReverseProxy(u upstream) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
}
func (m *mockContainer) newGitServer(store store, state *stateManager, action chan<- event, release chan<- event) (*gitServer, error) {
	if m.step == "gitserver" {
		return nil, errors.New("gitserver")
	}
	return nil, nil
}

func (m *mockContainer) newHTTPListener(addr string) (*HTTPListener, error) {
	if m.step == "api" && addr == listenerAPIAddr {
		return nil, errors.New("api")
	}
	if m.step == "proxy" && addr == listenerProxyAddr {
		return nil, errors.New("proxy")
	}
	return &HTTPListener{
		lsnr: &net.TCPListener{},
	}, nil
}

func (m *mockContainer) newSignalHandler() *signalHandler {
	if m.step == "signal" {
		return nil
	}
	return &signalHandler{}
}

func (m *mockContainer) createAndStartWorkflows(ctx context.Context, state *stateManager, git gitCommand, startTrigger chan event, startRelease chan event, switchUpstream func(string)) error {
	if m.step == "workflow" {
		return errors.New("workflow")
	}
	return nil
}

func Test_container_mock(t *testing.T) {
	c := &mockContainer{}
	err := c.getConf(Args{})
	if err != nil {
		t.Error("should succeed, got:", err)
	}
	_, err = c.createStore()
	if err != nil {
		t.Error("should succeed, got:", err)
	}
	_, err = c.newGitServer(store{}, nil, make(chan event), make(chan event))
	if err != nil {
		t.Error("should succeed, got:", err)
	}
	_, err = c.newDefaultGitCommand(store{})
	if err != nil {
		t.Error("should succeed, got:", err)
	}
	state := c.newStateManager()
	if state == nil {
		t.Error("should return a non-empty state")
	}
	_, err = c.newHTTPListener(listenerAPIAddr)
	if err != nil {
		t.Error("should succeed, got:", err)
	}

	signal := c.newSignalHandler()
	if signal == nil {
		t.Error("should return a signal")
	}
	c = &mockContainer{
		step: "signal",
	}
	signal = c.newSignalHandler()
	if signal != nil {
		t.Error("should return a signal")
	}
	err = c.createAndStartWorkflows(context.TODO(), nil, nil, make(chan event), make(chan event), func(string) {})
	if err != nil {
		t.Error("should succeed, got:", err)
	}
	c = &mockContainer{
		step: "workflow",
	}
	err = c.createAndStartWorkflows(context.TODO(), nil, nil, make(chan event), make(chan event), func(string) {})
	if err == nil {
		t.Error("should fail")
	}
}
