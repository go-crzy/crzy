package pkg

import (
	"net/http"
	"net/http/httputil"
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
