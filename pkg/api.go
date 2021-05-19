package pkg

import "net/http"

type api struct{}

func (h api)ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("version"))
}

