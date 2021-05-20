package pkg

import "net/http"

type api struct {
	state stateClient
}

func (h api) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("version"))
}
