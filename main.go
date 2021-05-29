package main

import (
	"context"

	"github.com/go-crzy/crzy/pkg"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {
	pkg.NewCrzy().Run(context.Background(), version, commit, date, builtBy)
}
