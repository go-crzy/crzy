package main

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"github.com/go-crzy/crzy/pkg"
	"golang.org/x/sync/errgroup"
)

func parse() pkg.Args {
	a := pkg.Args{}
	flag.StringVar(&a.ConfigFile, "config", pkg.DefaultConfigFile, "configuration file")
	flag.StringVar(&a.Repository, "repository", "myrepo", "GIT repository URI")
	flag.StringVar(&a.Head, "head", "main", "GIT branch to build from")
	flag.BoolVar(&a.NoColor, "nocolor", false, "disable log color")
	flag.BoolVar(&a.Version, "version", false, "crzy version")
	flag.StringVar(&a.Lang, "template", "go", "template for language")
	flag.Parse()
	return a
}

func main() {
	args := parse()
	group, ctx := errgroup.WithContext(context.Background())
	runner, err := pkg.NewCrzy(args)
	if err != nil {
		return
	}
	group.Go(func() error { return runner.Run(ctx) })
	if err := group.Wait(); err != nil &&
		!errors.Is(err, context.Canceled) &&
		!errors.Is(err, pkg.ErrVersionRequested) {
		fmt.Println("error detected: ", err)
	}
}
