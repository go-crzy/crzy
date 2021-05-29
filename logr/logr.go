package logr

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/go-logr/logr"
)

var colorMap = map[string]color.Attribute{
	"":        color.FgYellow,
	"release": color.FgBlue,
	"store":   color.FgCyan,
	"http":    color.FgRed,
	"main":    color.FgGreen,
	"git":     color.FgHiYellow,
	"deploy":  color.FgHiBlue,
	"signal":  color.FgHiRed,
	"trigger": color.FgHiGreen,
}

var errUnknown = errors.New("unknown")

type defaultLogger struct {
	color         bool
	prefix        bool
	name          string
	keysAndValues map[string]string
	level         int
	out           io.Writer
}

func OptionColor(l *defaultLogger) *defaultLogger {
	l.color = true
	return l
}

func OptionNoOutput(l *defaultLogger) *defaultLogger {
	l.out = io.Discard
	return l
}

func OptionNoPrefix(l *defaultLogger) *defaultLogger {
	l.prefix = false
	return l
}

func NewLogger(name string, f ...(func(l *defaultLogger) *defaultLogger)) logr.Logger {
	log := &defaultLogger{
		name:          name,
		prefix:        true,
		keysAndValues: map[string]string{},
		out:           os.Stdout,
	}
	for _, v := range f {
		log = v(log)
	}
	heading(log)
	return log
}

func (c *defaultLogger) Enabled() bool {
	return true
}

func (c *defaultLogger) Info(msg string, keysAndValues ...interface{}) {
	switch len(keysAndValues) {
	case 0:
		c.Log("info", msg)
	default:
		c.Log("info", msg, keysAndValues...)
	}
}

func (c *defaultLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	switch len(keysAndValues) {
	case 0:
		if err == nil {
			err = errUnknown
		}
		c.Log("error", "err:"+err.Error(), "msg", msg)
	default:
		keysAndValues = append(keysAndValues, "msg")
		keysAndValues = append(keysAndValues, msg)
		if err == nil {
			err = errUnknown
		}
		c.Log("error", "err:"+err.Error(), keysAndValues...)
	}
}

func (c *defaultLogger) V(level int) logr.Logger {
	return &defaultLogger{
		out:           c.out,
		color:         c.color,
		name:          c.name,
		prefix:        c.prefix,
		keysAndValues: c.keysAndValues,
		level:         level,
	}
}

func (c *defaultLogger) WithValues(keysAndValues ...interface{}) logr.Logger {
	output := &defaultLogger{
		out:           c.out,
		color:         c.color,
		name:          c.name,
		prefix:        c.prefix,
		keysAndValues: c.keysAndValues,
		level:         c.level,
	}
	i := 0
	for i < len(keysAndValues) {
		key := fmt.Sprintf("%v", keysAndValues[i])
		val := fmt.Sprintf("%v", keysAndValues[i+1])
		i += 2
		if key == "msg" || key == "data" {
			output.keysAndValues[key] = val
		}
	}
	return output
}

func (c *defaultLogger) WithName(name string) logr.Logger {
	return &defaultLogger{
		out:           c.out,
		color:         c.color,
		name:          name,
		prefix:        c.prefix,
		keysAndValues: c.keysAndValues,
		level:         c.level,
	}
}

func (c *defaultLogger) Log(key string, msg string, keysAndValues ...interface{}) {
	prefix := fmt.Sprintf(
		"%s [%-5s] ",
		time.Now().Format("15:04:05.000"),
		key,
	)
	if !c.prefix {
		prefix = ""
	}
	log := fmt.Sprintf(
		"%s%-10s %s",
		prefix,
		c.name,
		msg,
	)
	i := 0
	keys := map[string]string{}
	if msg, ok := c.keysAndValues["msg"]; ok {
		keys["msg"] = msg
	}
	if data, ok := c.keysAndValues["data"]; ok {
		keys["data"] = data
	}
	for i < len(keysAndValues)-1 {
		key := fmt.Sprintf("%v", keysAndValues[i])
		val := fmt.Sprintf("%v", keysAndValues[i+1])
		i += 2
		if key == "msg" || key == "data" {
			keys[key] = val
		}
	}
	if msg, ok := keys["msg"]; ok {
		log += fmt.Sprintf(", msg:%-50s", msg)
		if len(msg) > 50 {
			log += "... "
		} else {
			log += "    "
		}
	}
	if data, ok := keys["data"]; ok {
		n := 65
		if len(data) < n {
			n = len(data)
		}
		log += fmt.Sprintf(" [%s]", data[0:n])
	}
	c.colorPrint(c.name, log)
}

func (c *defaultLogger) colorPrint(name, log string) {
	if !c.color {
		fmt.Fprintln(c.out, log)
		return
	}

	foreground, ok := colorMap[name]
	if !ok {
		foreground = color.FgMagenta
	}
	color.New(foreground).Fprintln(c.out, log)
}

type MockLogger struct {
	sync.Mutex
	logs []string
}

func (l *MockLogger) Enabled() bool {
	return true
}

func (l *MockLogger) Info(msg string, keysAndValues ...interface{}) {
	l.Lock()
	defer l.Unlock()
	l.logs = append(l.logs, msg)
}

func (l *MockLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	l.Lock()
	defer l.Unlock()
	l.logs = append(l.logs, msg)
}

func (l *MockLogger) V(level int) logr.Logger { return &MockLogger{} }

func (c *MockLogger) WithValues(keysAndValues ...interface{}) logr.Logger { return &MockLogger{} }

func (c *MockLogger) WithName(name string) logr.Logger { return &MockLogger{} }

func heading(log logr.Logger) {
	log.Info("")
	log.Info(" █▀▀ █▀▀█ ▀▀█ █░░█")
	log.Info(" █░░ █▄▄▀ ▄▀░ █▄▄█")
	log.Info(" ▀▀▀ ▀░▀▀ ▀▀▀ ▄▄▄█")
	log.Info("")
}
