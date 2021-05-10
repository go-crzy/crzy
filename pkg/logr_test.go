package pkg

import (
	"bytes"
	"errors"
	"testing"
)

func Test_NewLogger(t *testing.T) {
	v := NewLogger("main")
	if v == nil {
		t.Error("should not be nil")
		t.FailNow()
	}
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
