package pkg

import (
	"testing"
)

func Test_header(t *testing.T) {
	r := runContainer{
		Log: &mockLogger{},
	}
	r.heading()
}
