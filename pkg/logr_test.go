package pkg

import (
	"bytes"
	"errors"
	"os"
	"testing"
)

func Test_newCrzyLogger_from_interface(t *testing.T) {
	v := newCrzyLogger("main", false)
	if v == nil {
		t.Error("should not be nil")
		t.FailNow()
	}
	if !v.Enabled() {
		t.Error("should be enabled")
	}
	v.Info("hi", "data", "data")
	v.Error(errors.New("error"), "error", "data", "data")
	_ = v.V(2)
	w := v.WithValues("data", "data")
	w.Info("hello")
	w = w.WithValues("msg", "msg")
	w.Info("hello")
	w = w.WithValues("msg", "msgmsgmsgmsgmsgmsgmsgmsgmsgmsgmsgmsgmsgmsgmsgmsgmsgmsgmsgmsgmsgmsgmsgmsgmsgmsgmsgmsgmsgmsg")
	w.Info("hello")
	_ = v.WithName("main")
}

func Test_newCrzyLogger_from_crzyLogger(t *testing.T) {
	c := crzyLogger{
		color: true,
		name:  "git",
		out:   os.Stdout,
	}
	c.Info("hi", "data", "data")
	d := crzyLogger{
		color: true,
		name:  "doesnotexist",
		out:   os.Stdout,
	}
	d.Info("hi", "data", "data")
}

func Test_mockLogger(t *testing.T) {
	c := mockLogger{}
	_ = c.WithName("xxx")
	_ = c.V(10)
	_ = c.WithValues("data", "data")
}

func Test_colorPrint(t *testing.T) {
	out := bytes.Buffer{}
	c := &crzyLogger{
		name: "main",
		out:  &out,
	}
	c.colorPrint("main", "123")
	if out.String() != "123\n" {
		t.Errorf("should return 123, current: %s", out.String())
	}
}

func Test_Log(t *testing.T) {
	out := bytes.Buffer{}
	c := &crzyLogger{
		name: "main",
		out:  &out,
	}
	c.Log("main", "123")
	if out.String()[13:35] != "[main ] main       123" {
		t.Errorf("should return \"[main ] main       123\", current: \"%s\"", out.String()[13:35])
	}
}

func Test_Error(t *testing.T) {
	out := bytes.Buffer{}
	c := &crzyLogger{
		name: "main",
		out:  &out,
	}
	c.Error(errors.New("error"), "123")
	if out.String()[13:50] != "[error] main       err:error, msg:123" {
		t.Errorf("should return \"[error] main       err:error, msg:123\", current: \"%s\"", out.String()[13:50])
	}
}
