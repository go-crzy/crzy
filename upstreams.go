package crzy

import (
	"errors"
	"sync"
)

type DefaultUpstreams struct {
	sync.RWMutex
	Versions map[string]string
	Default *string
}

var ErrServiceNotFound = errors.New("notfound")

type Upstreamer interface {
	Register(string, string, bool)
	Unregister(string)
	Lookup(string) (string, error)
}

// Register an upstream server for a service version
func (u *DefaultUpstreams) Register(version, port string, def bool) {
	u.Lock()
	defer u.Unlock()
	u.Versions[version] = port
	if def {
		u.Default = &version
	}
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
	port, ok := u.Versions[*u.Default]
	if !ok {
		return "", ErrServiceNotFound
	}
	return port, nil
}

