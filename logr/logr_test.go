package logr

import (
	"errors"
	"io"
	"testing"
)

func Test_NewLogger_with_option(t *testing.T) {
	v := NewLogger("main", OptionColor, OptionNoOutput, OptionNoPrefix)
	if v == nil {
		t.Error("should not be nil")
	}
}

func Test_Heading(t *testing.T) {
	log := &MockLogger{}
	heading(log)
	if len(log.logs) != 5 {
		t.Error("should return 5 rows")
	}
	if log.logs[0] != "" {
		t.Error("I'm done")
	}
}

func Test_defaultLogger_nocolor(t *testing.T) {
	v := &defaultLogger{
		name:          "main",
		keysAndValues: map[string]string{},
		out:           io.Discard,
		color:         false,
	}

	if !v.Enabled() {
		t.Error("should be enabled")
	}
	v.Info("hi", "data", "data")
	v.Error(errors.New("error"), "error", "data", "data")
	v.Error(nil, "error")
	v.Error(nil, "error", "data", "data")
	v.Error(nil, "error", "data", "data")
	_ = v.V(2)
	_ = v.V(2)
	w := v.WithValues("data", "data")
	w.Info("hello")
	w = w.WithValues("msg", "msg")
	w.Info("hello")
	w = w.WithValues("msg", "msgmsgmsgmsgmsgmsgmsgmsgmsgmsgmsgmsgmsgmsgmsgmsgmsgmsgmsgmsgmsgmsgmsgmsgmsgmsgmsgmsgmsgmsg")
	w.Info("hello")
	_ = v.WithName("main")
}

func Test_defaultLogger_color(t *testing.T) {
	c := defaultLogger{
		color: true,
		name:  "git",
		out:   io.Discard,
	}
	c.Info("hi", "data", "data")
	d := defaultLogger{
		color: true,
		name:  "doesnotexist",
		out:   io.Discard,
	}
	d.Info("hi", "data", "data")
}

func Test_mockLogger(t *testing.T) {
	c := MockLogger{}
	_ = c.WithName("xxx")
	_ = c.V(10)
	_ = c.WithValues("data", "data")
	c.Info("info")
	c.Error(errors.New("error"), "data")
	b := c.Enabled()
	if !b {
		t.Error("log should be enabled")
	}
}
