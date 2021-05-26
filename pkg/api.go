package pkg

import (
	"fmt"
	"net/http"
)

type versionHandler struct{}

func (v *versionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("version"))
}

type versionsHandler struct {
	state *stateManager
}

func (v *versionsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(v.state.state.listVersions()))

}

type verHandler struct {
	state *stateManager
}

func (v *verHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	version := r.URL.Path[13:]
	output, err := v.state.state.listVersionDetails(version)
	if err != nil {
		w.Write([]byte(fmt.Sprintf(`{"message": "version %s not found"}`, version)))
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Write([]byte(output))
}
