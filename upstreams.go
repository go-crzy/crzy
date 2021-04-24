package crzy

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
	Register(string, HTTPProcess, bool)
	Unregister(string)
	Lookup(string) (string, error)
	Next() (string, error)
}

// Register an upstream server for a service version
func (u *DefaultUpstreams) Register(version string, process HTTPProcess, def bool) {
	u.Lock()
	defer u.Unlock()
	u.Versions[version] = process
	if def {
		u.Default = &version
	}
}

// Next provides a port
func (u *DefaultUpstreams) Next() (string, error) {
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
func (u *DefaultUpstreams) Unregister(version string) {
	u.Lock()
	defer u.Unlock()
	delete(u.Versions, version)
	if u.Default != nil && *u.Default == version {
		u.Default = nil
	}
}

// Lookup returns the port for the version to find
func (u *DefaultUpstreams) Lookup(version string) (string, error) {
	u.RLock()
	defer u.RUnlock()
	if version == "" || version == "main" {
		if u.Default == nil {
			return "", ErrServiceNotFound
		}
		version = *u.Default
	}
	process, ok := u.Versions[version]
	if !ok {
		return "", ErrServiceNotFound
	}
	return process.Addr, nil
}
