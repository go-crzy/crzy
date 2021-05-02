package pkg

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/go-logr/logr"
)

type HTTPListener struct {
	errc chan error
	log  logr.Logger
}

func NewHTTPListener() *HTTPListener {
	return &HTTPListener{
		errc: make(chan error),
		log:  NewLogger("http"),
	}
}

func (l *HTTPListener) Run(ctx context.Context, addr string, handler http.Handler) error {
	log := l.log
	log.Info("starting HTTP listener", "data", addr)
	lsnr, err := net.Listen("tcp", addr)
	if err != nil {
		log.Error(err, "failed to start HTTP listener", "data", addr)
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
				log.Error(err, "HTTP listener failed", "data", addr)
				return err
			}
			log.Info("HTTP listener stopped", "data", addr)
			return nil
		case <-ctx.Done():
			log.Info("stopping HTTP listener", "data", addr)
			return ctx.Err()
		}
	}
}

func Logging(log logr.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		next.ServeHTTP(w, r)
		t2 := time.Now()
		log.Info(fmt.Sprintf("[%s] %q %v", r.Method, r.URL.String(), t2.Sub(t1)))
	})
}
