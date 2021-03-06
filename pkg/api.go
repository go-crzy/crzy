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
	route := r.URL.Path[13:]
	keys := strings.Split(route, "/")
	if len(keys) == 1 {
		output, err := v.state.state.listVersionDetails(keys[0])
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message":"not found"}`))
			return
		}
		w.Write([]byte(output))
		return
	}

	if len(keys) == 2 && (keys[1] == "log" || keys[1] == "err") {
		output, err := v.state.state.logVersion(keys[0], keys[1])
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message":"not found"}`))
			return
		}
		w.Write([]byte(output))
		return
	}
	w.Write([]byte(`{"message":"error"}`))
}

type actionHandler struct{}

type action struct {
	Command string
}

func (a *actionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var p action

	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"bad request"}`))
		return
	}
	if p.Command == "start" {
		w.Write([]byte(`{"message":"started"}`))
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(`{"message":"bad request"}`))
}

type configHandler struct{}

type configuration struct {
	Head string
}

func (c *configHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	var config configuration

	if r.Method == http.MethodPut {
		err := json.NewDecoder(r.Body).Decode(&config)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message":"bad request"}`))
			return
		}
	}
	w.Write([]byte(`{"message":"bad request"}`))
}

func newAPI(state *stateManager) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/v0/version", &versionHandler{})
	mux.Handle("/v0/versions", &versionsHandler{state: state})
	mux.Handle("/v0/versions/", &verHandler{state: state})
	mux.Handle("/v0/actions", &actionHandler{})
	mux.Handle("/v0/configuration", &configHandler{})
	mux.Handle("/v0/scripts", &scriptHandler{})
	return mux
}

type scriptHandler struct{}

func (s *scriptHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"scripts":["1", "2"]}`))

}
