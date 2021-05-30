package pkg

import (
	"testing"

	log "github.com/go-crzy/crzy/logr"
)

// Test_StoreCreateAndDelete
func Test_storeCreateAndDelete(t *testing.T) {
	run := defaultContainer{
		log: &log.MockLogger{},
	}
	store, err := run.createStore()
	if err != nil {
		t.Error("could not create store", err)
		t.FailNow()
	}
	defer store.delete()
}
