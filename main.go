package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-crzy/crzy/pkg"
	"golang.org/x/sync/errgroup"
)

func main() {
	group, ctx := errgroup.WithContext(context.Background())
	group.Go(func() error { return pkg.NewCrzy().Run(ctx) })
	if err := group.Wait(); err != nil &&
		!errors.Is(err, context.Canceled) &&
		!errors.Is(err, pkg.ErrVersionRequested) {
		fmt.Println("error detected: ", err)
	}
}
