package pkg

import (
	"fmt"
	"encoding/json"
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

type actionHandler struct{}

type action struct {
	Command string
}

func (a *actionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var p action

	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if p.Command == "start" {
		w.Write([]byte(`{"message":"started"}`))
		return
	}
	w.Write([]byte(`{"message":"failed"}`))
}
