package pkg

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/go-logr/logr"
)

func newCrzyLogger(name string, color bool) logr.Logger {
	return &crzyLogger{
		name:          name,
		keysAndValues: map[string]string{},
		out:           os.Stdout,
		color:         color,
	}
}

type crzyLogger struct {
	color         bool
	name          string
	keysAndValues map[string]string
	level         int
	out           io.Writer
}

func (c *crzyLogger) Enabled() bool {
	return true
}

func (c *crzyLogger) Info(msg string, keysAndValues ...interface{}) {
	switch len(keysAndValues) {
	case 0:
		c.Log("info", msg)
	default:
		c.Log("info", msg, keysAndValues...)
	}
}

func (c *crzyLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	switch len(keysAndValues) {
	case 0:
		c.Log("error", "err:"+err.Error(), "msg", msg)
	default:
		keysAndValues = append(keysAndValues, "msg")
		keysAndValues = append(keysAndValues, msg)
		c.Log("error", "err:"+err.Error(), keysAndValues...)
	}
}

func (c *crzyLogger) V(level int) logr.Logger {
	return &crzyLogger{
		out:           c.out,
		color:         c.color,
		name:          c.name,
		keysAndValues: c.keysAndValues,
		level:         level,
	}
}

func (c *crzyLogger) WithValues(keysAndValues ...interface{}) logr.Logger {
	output := &crzyLogger{
		out:           c.out,
		color:         c.color,
		name:          c.name,
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

func (c *crzyLogger) WithName(name string) logr.Logger {
	return &crzyLogger{
		out:           c.out,
		color:         c.color,
		name:          name,
		keysAndValues: c.keysAndValues,
		level:         c.level,
	}
}

func (c *crzyLogger) Log(key string, msg string, keysAndValues ...interface{}) {
	log := fmt.Sprintf(
		"%s [%-5s] %-10s %s",
		time.Now().Format("15:04:05.000"),
		key,
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

func (c *crzyLogger) colorPrint(name, log string) {
	if !c.color {
		fmt.Fprintln(c.out, log)
		return
	}
	colorMap := map[string]color.Attribute{
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
	foreground, ok := colorMap[name]
	if !ok {
		foreground = color.FgMagenta
	}
	color.New(foreground).Fprintln(c.out, log)
}

type mockLogger struct {
	sync.Mutex
	logs []string
}

func (l *mockLogger) Enabled() bool {
	return true
}

func (l *mockLogger) Info(msg string, keysAndValues ...interface{}) {
	l.Lock()
	defer l.Unlock()
	l.logs = append(l.logs, msg)
}

func (l *mockLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	l.Lock()
	defer l.Unlock()
	l.logs = append(l.logs, msg)
}

func (l *mockLogger) V(level int) logr.Logger { return &mockLogger{} }

func (c *mockLogger) WithValues(keysAndValues ...interface{}) logr.Logger { return &mockLogger{} }

func (c *mockLogger) WithName(name string) logr.Logger { return &mockLogger{} }
