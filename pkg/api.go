package pkg

import (
	"net/http"
)

type api struct {
	state *stateManager
}

func (h api) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	method := r.Method
	if path == "/v0/version" && method == http.MethodGet {
		// api := api{state: h.state}
		// api.ServeHTTP(w, r)
		// return
		w.Write([]byte("version"))
	}
	// _ = h.state.state.listVersions()
}
