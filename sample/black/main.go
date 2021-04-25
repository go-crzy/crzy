package main

import (
	"net/http"
	"os"
)

var port = ":8090"

func main() {

	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}
	http.ListenAndServe(
		port,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"color": "black"}`))
		}),
	)
}
