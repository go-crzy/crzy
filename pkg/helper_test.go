package pkg

import (
	"testing"
)

func Test_heading(t *testing.T) {
	r := runContainer{
		Log: &mockLogger{},
	}
	r.heading()
}
