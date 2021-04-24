package crzy

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type SignalHandler struct {
	signalc chan os.Signal
}

func NewSignalHandler() *SignalHandler {
	return &SignalHandler{
		signalc: make(chan os.Signal, 1),
	}
}

func (c *SignalHandler) Run(ctx context.Context, cancel context.CancelFunc) error {
	log.Println("starting signal handler....")
	signal.Notify(c.signalc, os.Interrupt, syscall.SIGTERM)
	for {
		select {
		case <-c.signalc:
			fmt.Println()
			log.Println("sigterm captured, stopping processes...")
			cancel()
			return nil
		case <-ctx.Done():
			log.Println("signal handler stop requested...")
			return ctx.Err()
		}
	}
}
