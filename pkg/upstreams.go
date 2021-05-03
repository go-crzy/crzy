package pkg

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os/exec"
	"sync"
	"time"
)

// NewReverseProxy creates a reverse proxy for the existing service
func NewReverseProxy(u Upstream) http.HandlerFunc {
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

type HTTPProcess struct {
	Addr string
	Cmd  *exec.Cmd
}

type DefaultUpstream struct {
	sync.RWMutex
	Versions map[string]HTTPProcess
	Default  *string
}

func NewUpstream() Upstream {
	return &DefaultUpstream{
		Versions: map[string]HTTPProcess{},
	}
}

// ErrServiceNotFound default service
var (
	ErrServiceNotFound = errors.New("notfound")
	ErrNoAvailablePort = errors.New("noport")
)

// Upstreamer the backend registration interface
type Upstream interface {
	Register(string, string, HTTPProcess, bool)
	SetDefault(string, string) error
	GetDefault() (string, error)
	Unregister(string, string)
	Lookup(string) (string, *exec.Cmd, error)
	NextPort() (string, error)
	KillAll() error
}

func (u *DefaultUpstream) KillAll() error {
	u.Lock()
	defer u.Unlock()
	for k, v := range u.Versions {
		if v.Cmd.Process != nil {
			v.Cmd.Process.Kill()
			delete(u.Versions, k)
		}
	}
	return nil
}

// Register an upstream server for a service version
func (u *DefaultUpstream) Register(name, version string, process HTTPProcess, def bool) {
	u.Lock()
	defer u.Unlock()
	key := fmt.Sprintf("%s/%s", name, version)
	u.Versions[key] = process
	if def {
		u.Default = &key
	}
}

// SetDefault an upstream server for a service version
func (u *DefaultUpstream) SetDefault(name, version string) error {
	u.Lock()
	defer u.Unlock()
	key := fmt.Sprintf("%s/%s", name, version)
	_, ok := u.Versions[key]
	if !ok {
		return ErrServiceNotFound
	}
	u.Default = &key
	return nil
}

// GetDefault an upstream server for a service version
func (u *DefaultUpstream) GetDefault() (string, error) {
	u.Lock()
	defer u.Unlock()
	if u.Default == nil {
		return "", ErrServiceNotFound
	}
	return *u.Default, nil
}

// Next provides a port
func (u *DefaultUpstream) NextPort() (string, error) {
	u.Lock()
	defer u.Unlock()
	for i := 8090; i < 8100; i++ {
		addr := fmt.Sprintf(":%d", i)
		found := false
		for k := range u.Versions {
			if addr == u.Versions[k].Addr {
				found = true
				break
			}
		}
		if !found {
			return addr, nil
		}
	}
	return "", ErrNoAvailablePort
}

// Unregister an upstream server for a service version
func (u *DefaultUpstream) Unregister(name, version string) {
	u.Lock()
	defer u.Unlock()
	key := fmt.Sprintf("%s/%s", name, version)
	delete(u.Versions, key)
	if u.Default != nil && *u.Default == key {
		u.Default = nil
	}
}

// Lookup returns the port for the version to find
func (u *DefaultUpstream) Lookup(service string) (string, *exec.Cmd, error) {
	u.RLock()
	defer u.RUnlock()
	if service == "" || service == "main" {
		if u.Default == nil {
			return "", nil, ErrServiceNotFound
		}
		service = *u.Default
	}
	process, ok := u.Versions[service]
	if !ok {
		return "", nil, ErrServiceNotFound
	}
	return process.Addr, process.Cmd, nil
}
