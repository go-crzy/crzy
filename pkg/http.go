package pkg

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
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
type expl struct {
	next               http.Handler
	username, password string
}

func (ex *expl) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if ex.password == "" || ex.username == "" {
		ex.next.ServeHTTP(w, r)
		return
	}
	authorization := r.Header.Get("authorization")
	keys := strings.Split(authorization, " ")
	if len(keys) != 2 || keys[0] != "Basic" {
		w.Header().Add("WWW-Authenticate", "Basic realm=\"require auth\"")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	body, err := base64.StdEncoding.DecodeString(keys[1])
	if err != nil {
		w.Header().Add("WWW-Authenticate", "Basic realm=\"require auth\"")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	keys = strings.Split(string(body), ":")
	if len(keys) != 2 || keys[0] != ex.username || keys[1] != ex.password {
		w.Header().Add("WWW-Authenticate", "Basic realm=\"require auth\"")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	ex.next.ServeHTTP(w, r)
}

func (c *config) authMiddleware(h http.Handler) http.Handler {
	return &expl{
		next:     h,
		username: c.API.Username,
		password: c.API.Password,
	}
}
