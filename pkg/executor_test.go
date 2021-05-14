package pkg

import (
	"context"
	"errors"
	"testing"

	"golang.org/x/sync/errgroup"
)

func Test_executorWithSuccess(t *testing.T) {
	r := runContainer{
		Log: &mockLogger{},
	}
	actionChannel := make(chan actionStruct)
	returnChannel := make(chan returnStruct)
	defer close(returnChannel)
	g, ctx := errgroup.WithContext(context.TODO())

	ctxWithCancel, cancel := context.WithCancel(ctx)
	g.Go(func() error { return r.createAndStartExecutor(ctxWithCancel, actionChannel, returnChannel) })
	f := func() ([]envVar, error) {
		return []envVar{{Name: "Key", Value: "Value"}}, nil
	}
	actionChannel <- actionStruct{F: f, JobID: "1"}
	returnMessage := <-returnChannel
	if returnMessage.Err != nil {
		t.Error("error should be nil, instead", returnMessage.Err)
	}
	if len(returnMessage.Envs) != 1 || returnMessage.Envs[0].Value != "Value" {
		t.Error("should return 1 value of Value")
	}
	cancel()
	if err := g.Wait(); err != nil && err.Error() != "context canceled" {
		t.Error(g.Wait(), "should succeed be canceled")
	}
}

func Test_executorWithError(t *testing.T) {
	r := runContainer{
		Log: &mockLogger{},
	}
	actionChannel := make(chan actionStruct)
	returnChannel := make(chan returnStruct)
	defer close(returnChannel)
	g, ctx := errgroup.WithContext(context.TODO())

	ctxWithCancel, cancel := context.WithCancel(ctx)
	g.Go(func() error { return r.createAndStartExecutor(ctxWithCancel, actionChannel, returnChannel) })
	f := func() ([]envVar, error) {
		return nil, errors.New("error")
	}
	actionChannel <- actionStruct{F: f, JobID: "2"}
	returnMessage := <-returnChannel
	if returnMessage.Err == nil || returnMessage.Err.Error() != "error" {
		t.Error("should return an error")
	}
	cancel()
	if err := g.Wait(); err != nil && err.Error() != "context canceled" {
		t.Error(g.Wait(), "should succeed be canceled")
	}
}
