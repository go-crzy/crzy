package pkg

import (
	"context"
	"log"
	"net"
	"net/http"
)

type HTTPListener struct {
	errc chan error
}

func NewHTTPListener() *HTTPListener {
	return &HTTPListener{
		errc: make(chan error),
	}
}

func (l *HTTPListener) Run(ctx context.Context, addr string, handler http.Handler) error {
	log.Printf("starting HTTP listener on %s....", addr)
	lsnr, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("failed to start HTTP listener on %s....", addr)
		return err
	}
	defer lsnr.Close()
	go func() {
		l.errc <- http.Serve(lsnr, handler)
	}()
	for {
		select {
		case err := <-l.errc:
			if err != nil {
				log.Printf("HTTP listener on %s failed: %v", addr, err)
				return err
			}
			log.Printf("HTTP listener on %s stopped...", addr)
			return nil
		case <-ctx.Done():
			log.Printf("stopping HTTP listener on %s....", addr)
			return ctx.Err()
		}
	}
}
