package pkg

import (
	"testing"
)

// Test_StoreCreateAndDelete
func Test_storeCreateAndDelete(t *testing.T) {
	run := runContainer{
		Log: &mockLogger{},
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
		log:     &mockLogger{},
	}
	err := s.delete()
	if err != nil {
		t.Error("could not create store", err)
	}
}
