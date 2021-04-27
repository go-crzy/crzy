package crzy

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os/exec"
	"strings"
	"time"
)

var (
	ErrInvalidService = errors.New("invalidservice")
	ErrServiceRunning = errors.New("servicerunning")
)

func NewUpstreamAPI(u Upstreamer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		command := "unknown"
		program := "red"
		if len(strings.Split(r.URL.Path, "/")) >= 2 {
			command = strings.Split(r.URL.Path, "/")[1]
		}
		if len(strings.Split(r.URL.Path, "/")) >= 3 {
			program = strings.Split(r.URL.Path, "/")[2]
		}
		switch command {
		case "start":
			_, _, err := u.Lookup(program + "/v1")
			if err == nil {
				log.Printf("%s/v1 already started, set default", program)
				u.SetDefault(program, "v1")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"message": "OK"}`))
				return
			}
			port, err := u.NextPort()
			if err == ErrNoAvailablePort {
				w.WriteHeader(http.StatusGone)
				w.Write([]byte(`{"message": "Gone"}`))
				return
			}
			cmd := &exec.Cmd{
				Dir:  fmt.Sprintf("./sample/%s", program),
				Path: program,
				Env:  []string{fmt.Sprintf("PORT=%s", port)},
			}
			log.Printf("starting %s/v1 with port %s", program, port)
			u.Register(program, "v1", HTTPProcess{Addr: port, Cmd: cmd}, true)
			cmd.Start()
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message": "OK"}`))
			return
		case "stop":
			_, cmd, err := u.Lookup(program + "/v1")
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"message": "NotFound"}`))
				return
			}
			log.Printf("stopping %s/v1", program)
			cmd.Process.Kill()
			u.Unregister(program, "v1")
			w.WriteHeader(http.StatusNoContent)
			w.Write([]byte(`{"message": "Stopped"}`))
			return
		case "rollout":
			err := Rollout(u, "red", "black")
			if err == nil {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"message": "OK"}`))
				return
			}
			if err == ErrNoAvailablePort {
				w.WriteHeader(http.StatusGone)
				w.Write([]byte(`{"message": "NoAvailablePort"}`))
				return
			}
			err = Rollout(u, "black", "red")
			if err == nil {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"message": "OK"}`))
				return
			}
			if err == ErrNoAvailablePort {
				w.WriteHeader(http.StatusGone)
				w.Write([]byte(`{"message": "NoAvailablePort"}`))
				return
			}
			v, err := u.GetDefault()
			newVersion := "red"
			if err == nil && v == fmt.Sprintf("%s:v1", newVersion) {
				newVersion = "black"
			}
			err = u.SetDefault(newVersion, "v1")
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"message": "InternalServerError"}`))
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message": "OK"}`))
			return
		default:
			log.Printf("unknown command")
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message": "BadRequest"}`))
	})
}

// Rollout start newprogram and stop oldprogram
func Rollout(u Upstreamer, newprogram, oldprogram string) error {
	_, _, err := u.Lookup(newprogram + "/v1")
	if err == nil {
		return ErrServiceRunning
	}
	port, err := u.NextPort()
	if err != nil {
		return err
	}
	cmd := &exec.Cmd{
		Dir:  fmt.Sprintf("./sample/%s", newprogram),
		Path: newprogram,
		Env:  []string{fmt.Sprintf("PORT=%s", port)},
	}
	log.Printf("starting %s/v1 with port %s", newprogram, port)
	u.Register(newprogram, "v1", HTTPProcess{Addr: port, Cmd: cmd}, true)
	cmd.Start()
	_, cmd, err = u.Lookup(oldprogram + "/v1")
	if err == nil {
		log.Printf("stopping %s/v1", oldprogram)
		u.Unregister(oldprogram, "v1")
		cmd.Process.Kill()
	}
	return nil
}

// NewReverseProxy creates a reverse proxy for the existing service
func NewReverseProxy(u Upstreamer) http.HandlerFunc {
	transport := &http.Transport{
		TLSHandshakeTimeout: 10 * time.Second,
	}
	return func(w http.ResponseWriter, r *http.Request) {
		service := r.Header.Get("service")
		v, _, err := u.Lookup(service)
		if err == ErrServiceNotFound {
			http.Error(w, `{"message": "NotFound"}`, http.StatusNotFound)
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
