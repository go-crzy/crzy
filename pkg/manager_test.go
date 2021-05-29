package pkg

import (
	"context"
	"os"
	"testing"

	log "github.com/go-crzy/crzy/logr"
)

func Test_Startup(t *testing.T) {
	r := &defaultRunner{
		log: &log.MockLogger{},
	}
	ctx, cancel := context.WithCancel(context.TODO())
	os.Args = []string{"crzy", "-repository", "color.git", "-color", "-head", "example"}
	go r.Run(ctx, "", "", "", "")
	cancel()
}
