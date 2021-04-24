package crzy

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os/exec"
	"time"
)

var ErrInvalidService = errors.New("invalidservice")

func NewUpstreamAPI(u Upstreamer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/start":
			port, err := u.Next()
			if err != nil {
				w.WriteHeader(http.StatusGone)
				w.Write([]byte(`{"message": "Gone"}`))
				return
			}
			cmd := &exec.Cmd{
				Dir:  "./sample/color",
				Path: "color",
				Env:  []string{fmt.Sprintf("PORT=%s", port)},
			}
			log.Printf("starting color with port %s", port)
			u.Register("v1", HTTPProcess{Addr: port, Cmd: cmd}, true)
			cmd.Start()
		case "/stop":
			log.Printf("stopping command")
		case "/rollout":
			log.Printf("upgrading command")

		default:
			log.Printf("unknown command")

		}
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "NotFound"}`))
	})
}

func NewReverseProxy(u Upstreamer) http.HandlerFunc {
	transport := &http.Transport{
		TLSHandshakeTimeout: 10 * time.Second,
	}
	return func(w http.ResponseWriter, r *http.Request) {
		version := r.Header.Get("version")
		v, err := u.Lookup(version)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		(&httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req.URL.Scheme = "http"
				req.URL.Host = "localhost" + v
			},
			Transport: transport,
		}).ServeHTTP(w, r)
	}
}
