package main

import (
	"github.com/carnage-sh/crzy/pkg"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {
	pkg.Startup(version, commit, date, builtBy)
}
