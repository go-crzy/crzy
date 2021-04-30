package pkg

import (
	"errors"
	"fmt"
	"os/exec"
	"sync"
)

type HTTPProcess struct {
	Addr string
	Cmd  *exec.Cmd
}

type DefaultUpstreams struct {
	sync.RWMutex
	Versions map[string]HTTPProcess
	Default  *string
}

// ErrServiceNotFound default service
var (
	ErrServiceNotFound = errors.New("notfound")
	ErrNoAvailablePort = errors.New("noport")
)

// Upstreamer the backend registration interface
type Upstreamer interface {
	Register(string, string, HTTPProcess, bool)
	SetDefault(string, string) error
	GetDefault() (string, error)
	Unregister(string, string)
	Lookup(string) (string, *exec.Cmd, error)
	NextPort() (string, error)
}

// Register an upstream server for a service version
func (u *DefaultUpstreams) Register(name, version string, process HTTPProcess, def bool) {
	u.Lock()
	defer u.Unlock()
	key := fmt.Sprintf("%s/%s", name, version)
	u.Versions[key] = process
	if def {
		u.Default = &key
	}
}

// SetDefault an upstream server for a service version
func (u *DefaultUpstreams) SetDefault(name, version string) error {
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
func (u *DefaultUpstreams) GetDefault() (string, error) {
	u.Lock()
	defer u.Unlock()
	if u.Default == nil {
		return "", ErrServiceNotFound
	}
	return *u.Default, nil
}

// Next provides a port
func (u *DefaultUpstreams) NextPort() (string, error) {
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
func (u *DefaultUpstreams) Unregister(name, version string) {
	u.Lock()
	defer u.Unlock()
	key := fmt.Sprintf("%s/%s", name, version)
	delete(u.Versions, key)
	if u.Default != nil && *u.Default == key {
		u.Default = nil
	}
}

// Lookup returns the port for the version to find
func (u *DefaultUpstreams) Lookup(service string) (string, *exec.Cmd, error) {
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
