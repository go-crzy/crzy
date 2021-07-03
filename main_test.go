package main

import (
	"os"
	"testing"
)

func Test_argsParser(t *testing.T) {
	os.Args = []string{"crzy", "-repository", "color.git", "-nocolor"}
	a := parse()
	if a.NoColor != true {
		t.Error("args not parsed as expected")
	}
}
