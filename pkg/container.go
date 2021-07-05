package pkg

import (
	"context"
	"io"
	"net/http"

	"github.com/go-logr/logr"
)

type container interface {
	getConf(args Args) error
	createStore() (*store, error)
	newStateManager() *stateManager
	newDefaultGitCommand(store store) (gitCommand, error)
	newGitServer(store store, state *stateManager, action chan<- event, release chan<- event) (*gitServer, error)
	newReverseProxy(u upstream) http.Handler
	newHTTPListener(addr string) (*HTTPListener, error)
	newSignalHandler() *signalHandler
	createAndStartWorkflows(ctx context.Context, state *stateManager, git gitCommand, startTrigger chan event, startRelease chan event, switchUpstream func(string)) error
}

type defaultContainer struct {
	log    logr.Logger
	out    io.Writer
	config *config
}

var (
	version = "dev"
	commit  = "unknown"
)
