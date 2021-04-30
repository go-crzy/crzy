package pkg

import (
	"os"
	"testing"
)

func TestNewUpdater(t *testing.T) {
	dir := "./test"
	_, err := NewUpdater(dir, &DefaultUpstreams{}, make(chan<- func()))
	if err != nil {
		t.Errorf("updater failed, err: %v", err)
	}
	err = os.Remove(dir)
	if err != nil {
		t.Errorf("cleanup failed, err: %v", err)
	}
}
