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
	"github.com/rs/cors"
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

type expl struct {
	next               http.Handler
	username, password *string
}

var errWrongCredentials = errors.New("wrongcredentials")

func checkCredentials(authorization, username, password string) error {
	keys := strings.Split(authorization, " ")
	if len(keys) != 2 || strings.ToLower(keys[0]) != "basic" {
		return errWrongCredentials
	}
	body, err := base64.StdEncoding.DecodeString(keys[1])
	if err != nil {
		return errWrongCredentials
	}
	keys = strings.Split(string(body), ":")
	if len(keys) < 2 || keys[0] != username || strings.Join(keys[1:], ":") != password {
		return errWrongCredentials
	}
	return nil
}

func (ex *expl) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if ex.password != nil && ex.username != nil &&
		checkCredentials(r.Header.Get("authorization"), *ex.username, *ex.password) != nil {
		w.Header().Add("WWW-Authenticate", "Basic realm=\"auth required\"")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	ex.next.ServeHTTP(w, r)
}

func (c *config) authMiddleware(h http.Handler) http.Handler {
	return &expl{
		next:     h,
		username: c.Main.API.Username,
		password: c.Main.API.Password,
	}
}

func (c *config) corsMiddleware(h http.Handler) http.Handler {
	cm := cors.New(cors.Options{
				AllowedOrigins:   c.Main.Proxy.Origins,
		 		AllowCredentials: true,
		 		Debug:            false,
 	})
	return cm.Handler(h)
}
