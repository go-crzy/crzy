package pkg

import (
	"context"
	"os"
	"testing"
)

func Test_Startup(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	os.Args = []string{"crzy", "-repository", "color.git", "-color", "-head", "example"}
	go Startup(ctx, "", "", "", "")
	cancel()
}
