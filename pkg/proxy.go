package pkg

import (
	"errors"
	"net/http"
	"net/http/httputil"
	"sync"
	"time"
)

// newReverseProxy creates a reverse proxy for the existing service
func (r *defaultContainer) newReverseProxy(u upstream) http.Handler {
	transport := &http.Transport{
		TLSHandshakeTimeout: 10 * time.Second,
	}
	return r.config.corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v, err := u.getDefault()
		if err == errServiceNotFound {
			http.Error(w, `{"message": "NotFound"}`, http.StatusNotFound)
			return
		}
		(&httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req.URL.Scheme = "http"
				req.URL.Host = v
			},
			Transport: transport,
		}).ServeHTTP(w, r)
	}))
}

type defaultUpstream struct {
	sync.RWMutex
	defaultUpstream *string
	state           state
}

func newUpstream(state state) upstream {
	return &defaultUpstream{
		state: state,
	}
}

// ErrServiceNotFound default service
var (
	errServiceNotFound = errors.New("notfound")
)

// Upstreamer the backend registration interface
type upstream interface {
	setDefault(string)
	getDefault() (string, error)
	listVersions() []byte
}

// SetDefault an upstream server for a service version
func (u *defaultUpstream) setDefault(name string) {
	u.Lock()
	defer u.Unlock()
	u.defaultUpstream = &name
}

// GetDefault an upstream server for a service version
func (u *defaultUpstream) getDefault() (string, error) {
	u.Lock()
	defer u.Unlock()
	if u.defaultUpstream == nil {
		return "", errServiceNotFound
	}
	return *u.defaultUpstream, nil
}

func (u *defaultUpstream) listVersions() []byte {
	return u.state.listVersions()
}
