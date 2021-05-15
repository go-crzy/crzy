package pkg

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/sync/errgroup"
)

func Test_newHTTPListener_and_success(t *testing.T) {
	r := &runContainer{
		Log: &mockLogger{},
	}
	v, err := r.newHTTPListener(":8999")
	if err != nil {
		t.Error("should succeed", err)
	}
	if v == nil {
		t.Error("should not be nil")
		t.FailNow()
	}
	if v.log == nil || !v.log.Enabled() {
		t.Error("log should be enabled")
	}
	if v.errc == nil {
		t.Error("cron should not be empty")
	}
}

func Test_newHTTPListener_and_fail(t *testing.T) {
	r := &runContainer{
		Log: &mockLogger{},
	}
	_, err := r.newHTTPListener("abc")
	if err == nil {
		t.Error("should fail")
	}
}

func Test_loggingMiddleware(t *testing.T) {
	handler := http.NewServeMux()
	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello"))
	})
	server := httptest.NewServer(loggingMiddleware(newCrzyLogger("demo", false), handler))
	client := server.Client()

	request, _ := http.NewRequest("Get", server.URL, nil)
	response, err := client.Do(request)
	if err != nil {
		t.Errorf("Should not return %v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf(
			"Status Code should be 200, current: %d",
			response.StatusCode,
		)
	}
}

func Test_run_and_cancel(t *testing.T) {
	errc := make(chan error)
	l, err := net.Listen("tcp", ":8099")
	if err != nil {
		t.Error(err, "listener should start")
	}
	lsnr := &HTTPListener{
		errc: errc,
		log:  &mockLogger{},
		lsnr: l,
	}
	g, ctx := errgroup.WithContext(context.TODO())
	ctx, cancel := context.WithCancel(ctx)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	g.Go(func() error { return lsnr.run(ctx, next) })
	cancel()
	if err := g.Wait(); err == nil || err.Error() != "context canceled" {
		t.Error(err, "should return context cancel")
	}
}

func Test_run_and_no_error(t *testing.T) {
	errc := make(chan error)
	l, err := net.Listen("tcp", ":8099")
	if err != nil {
		t.Error(err, "listener should start")
	}
	lsnr := &HTTPListener{
		errc: errc,
		log:  &mockLogger{},
		lsnr: l,
	}
	g, ctx := errgroup.WithContext(context.TODO())
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	g.Go(func() error { return lsnr.run(ctx, next) })
	errc <- error(nil)
	if err := g.Wait(); err != nil {
		t.Error(err, "should not return any error")
	}
}

func Test_run_and_error(t *testing.T) {
	errc := make(chan error)
	l, err := net.Listen("tcp", ":8099")
	if err != nil {
		t.Error(err, "listener should start")
	}
	lsnr := &HTTPListener{
		errc: errc,
		log:  &mockLogger{},
		lsnr: l,
	}
	g, ctx := errgroup.WithContext(context.TODO())
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	g.Go(func() error { return lsnr.run(ctx, next) })
	errc <- errors.New("error")
	if err := g.Wait(); err == nil || err.Error() != "error" {
		t.Error(err, "should return error")
	}
}
