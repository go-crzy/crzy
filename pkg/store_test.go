package pkg

import (
	"testing"

	log "github.com/go-crzy/crzy/logr"
)

// Test_StoreCreateAndDelete
func Test_storeCreateAndDelete(t *testing.T) {
	run := runContainer{
		Log: &log.MockLogger{},
	}
	store, err := run.createStore()
	if err != nil {
		t.Error("could not create store", err)
		t.FailNow()
	}
	err = store.delete()
	if err != nil {
		t.Error("could not delete store", err)
	}
}

// Test_StoreCreateAndDelete
func Test_storeDelete_and_fail(t *testing.T) {
	s := &store{
		rootDir: "/doesnotexist",
		log:     &log.MockLogger{},
	}
	err := s.delete()
	if err != nil {
		t.Error("could not create store", err)
	}
}
