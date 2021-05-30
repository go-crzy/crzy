package pkg

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-logr/logr"
)

type signalHandler struct {
	signalc chan os.Signal
	log     logr.Logger
}

func (r *defaultContainer) newSignalHandler() *signalHandler {
	return &signalHandler{
		signalc: make(chan os.Signal, 1),
		log:     r.log.WithName("signal"),
	}
}

func (c *signalHandler) run(ctx context.Context, cancel context.CancelFunc) error {
	defer close(c.signalc)
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
