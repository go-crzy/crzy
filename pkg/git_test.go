package pkg

import (
	"runtime"
	"testing"
)

func Test_execCmd(t *testing.T) {
	switch runtime.GOOS {
	case "windows":
		output, err := execCmd(".", "git", "version")
		if err != nil {
			t.Error("test fails", err)
		}
		if string(output[0:11]) != "git version" {
			t.Errorf("output should be git version, current %q", output)
		}
		
		//t.Error("this test does not work on windows")
	default:
		output, err := execCmd(".", "echo", "-n", "test")
		if err != nil {
			t.Error("test fails", err)
		}
		if string(output) != "test" {
			t.Errorf("output should be test, current %q", output)
		}
	}
}