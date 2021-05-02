package pkg

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-logr/logr"
)

type SignalHandler struct {
	signalc chan os.Signal
	log     logr.Logger
}

func NewSignalHandler() *SignalHandler {
	return &SignalHandler{
		signalc: make(chan os.Signal, 1),
		log:     NewLogger("signal"),
	}
}

func (c *SignalHandler) Run(ctx context.Context, cancel context.CancelFunc) error {
	log := c.log
	log.Info("starting signal handler....")
	signal.Notify(c.signalc, os.Interrupt, syscall.SIGTERM)
	for {
		select {
		case <-c.signalc:
			fmt.Println()
			log.Info("sigterm captured, stopping processes...")
			cancel()
			return nil
		case <-ctx.Done():
			log.Info("signal handler stop requested...")
			return ctx.Err()
		}
	}
}
