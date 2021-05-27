package pkg

import (
	"encoding/json"
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
	// TODO: si une personne tape /v0/versions/abc sans rien derriere,
	// renvoyer listVersionDetails
	output, err := v.state.state.listVersionDetails(version)
	if err != nil {
		w.Write([]byte(fmt.Sprintf(`{"message": "version %s not found"}`, version)))
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Write([]byte(output))
	// Sinon si une personne renvoie /v0/versions/abc/log ou /v0/versions/abc/err,
	// renvoyer state.logVersion(version, "log") ou state.logVersion(version, "err")
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
