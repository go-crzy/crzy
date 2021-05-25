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

// TODO: create a handler that checks the r.URL.Path is /v0/versions/xxxx si c'est le cas
// utiliser listVersionDetails pour vérifier que xxxx existe: (e.g. caa0ea14d84ca40c)
// - s'il n'exsiste pas, renvoyer un message `{"message": "Not Found"}`
// - s'il existe, renvoyer le retour de listVersionDetails()
