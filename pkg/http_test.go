package pkg

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	log "github.com/go-crzy/crzy/logr"
	"golang.org/x/sync/errgroup"
)

func Test_newHTTPListener_and_success(t *testing.T) {
	r := &defaultContainer{
		log: &log.MockLogger{},
	}
	v, err := r.newHTTPListener(listenerAPIAddr)
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
	r := &defaultContainer{
		log: &log.MockLogger{},
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
	server := httptest.NewServer(loggingMiddleware(&log.MockLogger{}, handler))
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
		log:  &log.MockLogger{},
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
		log:  &log.MockLogger{},
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
		log:  &log.MockLogger{},
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

func Test_AuthdHandler_forbidden(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("Get", "/", nil)
	request.Header.Add("authorization", "Basic aaa")

	handler := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {rw.Write([]byte("hello"))})
	conf := &config{API: apiStruct{Username: "emilie", Password: "emilie"}}
	next := conf.authMiddleware(handler)

	next.ServeHTTP(recorder, request)

	result := recorder.Result()
	if result.StatusCode != http.StatusUnauthorized {
		t.Errorf(
			"Status Code should be 401, current: %d",
			result.StatusCode,
		)
	}
}

func Test_AuthdHandler_authorized(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("Get", "/", nil)
	request.Header.Add("authorization", "Basic ZW1pbGllOmVtaWxpZQ==")

	handler := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {rw.Write([]byte("hello"))})
	conf := &config{API: apiStruct{Username: "emilie", Password: "emilie"}}
	next := conf.authMiddleware(handler)

	next.ServeHTTP(recorder, request)

	result := recorder.Result()
	if result.StatusCode != http.StatusOK {
		t.Errorf(
			"Status Code should be 200, current: %d",
			result.StatusCode,
		)
	}
}