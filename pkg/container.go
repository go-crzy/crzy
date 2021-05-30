package pkg

import (
	"context"
	"fmt"
	"io"

	"github.com/go-logr/logr"
)

type container interface {
	load() error
	createStore() (*store, error)
	newStateManager() *stateManager
	newDefaultGitCommand(store store) (gitCommand, error)
	newGitServer(store store, state *stateManager, action chan<- event, release chan<- event) (*gitServer, error)
	newHTTPListener(addr string) (*HTTPListener, error)
	newSignalHandler() *signalHandler
	createAndStartWorkflows(ctx context.Context, state *stateManager, git gitCommand, startTrigger chan event, startRelease chan event, switchUpstream func(string)) error
}

type defaultContainer struct {
	log    logr.Logger
	out    io.Writer
	config *config
	parser parser
}

var (
	version = "dev"
	commit  = "unknown"
)

func (c *defaultContainer) load() error {
	args := c.parser.parse()
	if args.version {
		fmt.Fprintf(c.out, "crzy version %s(%s)\n", version, commit)
		return ErrVersionRequested
	}
	conf, err := getConf(args)
	c.config = conf
	return err
}
