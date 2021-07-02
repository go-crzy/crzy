package pkg

import (
	"context"
	"errors"
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

const (
	listenerProxyAddr = "proxy"
	listenerAPIAddr   = "api"
)

var errUnknownListener = errors.New("listener unknown")

func (r *defaultContainer) newHTTPListener(key string) (*HTTPListener, error) {
	addr := ""
	switch key {
	case listenerProxyAddr:
		addr = ":8081"
	case listenerAPIAddr:
		addr = ":8080"
	default:
		return nil, errUnknownListener
	}
	lsnr, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &HTTPListener{
		errc: make(chan error),
		log:  r.log.WithName("http"),
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

// Step 3: Creer un middleware qui va:
// - vérifier le contenu du Header HTTP "authorization"
// - s'il ne contient pas "basic" et une chaine de caractère (tester en case-insensitive) renvoyer http-401
// - décoder la chaine après basic en base64, celle-ci doit être de la forme "mot1:mot2", sinon renvoyer http-401
// - vérifier que mot1==username et mot2==password sinon, renvoyer http-401
