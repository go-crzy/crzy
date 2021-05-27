package pkg

import (
	"encoding/json"
	"net/http"
	"strings"
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

	body := strings.Split(version, "/")
	if len(body) == 1 {
		output, err := v.state.state.listVersionDetails(r.URL.Path[13:])
		if err != nil {
			w.Write([]byte(`{"message":"not found"}`))
			return
		}
		w.Write([]byte(output))
		return
	}

	if len(body) == 2 && (body[1] == "log" || body[1] == "err") {
		output, err := v.state.state.logVersion(body[0], body[1])
		if err != nil {
			w.Write([]byte(`{"message":"not found"}`))
			return
		}
		w.Write([]byte(output))
		return
	}

	w.Write([]byte(`{"message":"error"}`))

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
