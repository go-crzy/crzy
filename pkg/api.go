package pkg

import (
	"net/http"
)

type api struct {
	state *stateManager
}

func (h api) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		switch r.URL.Path {
		case "/v0/version":
			w.Write([]byte("version"))
		case "/v0/versions":
			w.Write([]byte(h.state.state.listVersions()))
		}
	}
}
