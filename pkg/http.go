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
	lsnr net.Listener
}

func (r *runContainer) newHTTPListener(addr string) (*HTTPListener, error) {
	lsnr, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &HTTPListener{
		errc: make(chan error),
		log:  r.Log.WithName("http"),
		lsnr: lsnr,
	}, nil
}

func (l *HTTPListener) run(ctx context.Context, handler http.Handler) error {
	log := l.log
	log.Info("starting HTTP listener", "data", l.lsnr.Addr().String())
	go func(lsnr net.Listener) {
		l.errc <- http.Serve(lsnr, handler)
	}(l.lsnr)
	defer l.lsnr.Close()
	for {
		select {
		case err := <-l.errc:
			if err != nil {
				log.Error(err, "HTTP listener failed", "data", l.lsnr.Addr().String())
				return err
			}
			log.Info("HTTP listener stopped", "data", l.lsnr.Addr().String())
			return nil
		case <-ctx.Done():
			log.Info("stopping HTTP listener", "data", l.lsnr.Addr().String())
			return ctx.Err()
		}
	}
}

func loggingMiddleware(log logr.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		next.ServeHTTP(w, r)
		t2 := time.Now()
		log.Info(fmt.Sprintf("[%s] %q %v", r.Method, r.URL.String(), t2.Sub(t1)))
	})
}
