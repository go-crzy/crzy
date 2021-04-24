package crzy

import (
	"github.com/sosedoff/gitkit"
	"log"
	"net/http"
)

func NewGITServer(dir string) http.Handler {
	// Configure git hooks
	hooks := &gitkit.HookScripts{
		PreReceive: `echo "Hello World!"`,
	}

	// Configure git service
	service := gitkit.New(gitkit.Config{
		Dir:        dir,
		AutoCreate: true,
		AutoHooks:  true,
		Hooks:      hooks,
	})

	// Configure git server. Will create git repos path if it does not exist.
	// If hooks are set, it will also update all repos with new version of hook scripts.
	if err := service.Setup(); err != nil {
		log.Fatal(err)
	}
	return service
}
