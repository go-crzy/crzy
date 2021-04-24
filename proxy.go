package crzy

import (
	"errors"
	"net/http"
	"net/http/httputil"
	"time"
)

var ErrInvalidService = errors.New("invalid service/version")

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
