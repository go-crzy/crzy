package pkg

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/go-logr/logr"
)

func NewLogger(name string) logr.Logger {
	return &crzyLogger{
		name:          name,
		keysAndValues: map[string]string{},
	}
}

type crzyLogger struct {
	name          string
	keysAndValues map[string]string
	level         int
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
		name:          c.name,
		keysAndValues: c.keysAndValues,
		level:         level,
	}
}

func (c *crzyLogger) WithValues(keysAndValues ...interface{}) logr.Logger {
	output := &crzyLogger{
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
	payload := c.keysAndValues
	for i < len(keysAndValues)-1 {
		key := fmt.Sprintf("%v", keysAndValues[i])
		val := fmt.Sprintf("%v", keysAndValues[i+1])
		i += 2
		if key == "msg" || key == "data" {
			payload[key] = val
		}
	}
	if msg, ok := payload["msg"]; ok {
		log += fmt.Sprintf(", msg:%-50s", msg)
		if len(msg) > 50 {
			log += "... "
		} else {
			log += "    "
		}
	}
	if data, ok := payload["data"]; ok {
		n := 50
		if len(data) < n {
			n = len(data)
		}
		log += fmt.Sprintf(" [%s]", data[0:n])
	}
	colorPrint(c.name, log)
}

func colorPrint(name, log string) {
	if conf.Main.Color {
		fgColor := color.FgRed
		switch name {
		case "":
			fgColor = color.FgYellow
		case "machine":
			fgColor = color.FgBlue
		case "store":
			fgColor = color.FgCyan
		case "http":
			fgColor = color.FgRed
		case "main":
			fgColor = color.FgGreen
		default:
			fgColor = color.FgMagenta
		}
		color.Set(fgColor)
		defer color.Unset()
	}
	fmt.Println(log)
}
