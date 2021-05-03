package pkg

import (
	"context"
	"errors"
	"os"
	"testing"

	"golang.org/x/sync/errgroup"
)

// Test_StoreClose makes sure the directory associated with the store is cleaned
// up when the program stops.
func Test_StoreClose(t *testing.T) {
	dir := "./store-test"
	err := os.Mkdir(dir, os.ModeDir|os.ModePerm)
	if err != nil {
		t.Errorf("store start failed, err: %v", err)
	}

	g := new(errgroup.Group)
	storeservice := NewStoreService(dir)
	ctx := context.TODO()
	storeCtx, cancel := context.WithCancel(ctx)
	g.Go(func() error { return storeservice.Run(storeCtx) })
	cancel()
	if err := g.Wait(); err != context.Canceled {
		t.Errorf("program has stopped (%v)", err)
	}
	_, err = os.Stat(dir)
	if err == nil {
		t.Error("directory should have been cleaned up")
		os.RemoveAll("./store-test")
		t.FailNow()
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("directory should have been cleaned up: %v", err)
	}
}
