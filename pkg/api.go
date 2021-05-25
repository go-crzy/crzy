package pkg

import (
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
	if len(r.URL.Path) < 13 && r.URL.Path[:13] != "/v0/versions/" {
		w.Write([]byte(`{"message": "Not Found"}`))
		return
	}
	output, err := v.state.state.listVersionDetails(r.URL.Path[13:])
	if err != nil {
		w.Write([]byte(`{"message": "Not Found"}`))
		return
	}
	w.Write([]byte(output))
}