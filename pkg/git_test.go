package pkg

import (
	"os/exec"
	"runtime"
	"testing"
)

func Test_execCmd(t *testing.T) {
	output, err := execCmd(".", "git", "version")
	if err != nil {
		t.Error("test fails", err)
	}
	if string(output[0:11]) != "git version" {
		t.Errorf("output should be git version, current %q", output)
	}
}

func Test_cmdProcessKill(t *testing.T) {
	name := "tail"
	args := []string{
		"-f", "/dev/null",
	}
	if runtime.GOOS == "windows" {
		name = "powershell"
		args = []string{"-Command", "Get-Content cron.go -Wait"}
	}
	cmd := exec.Command(name, args...)
	err := cmd.Start()
	if err != nil {
		t.Error(err, "start failed")
	}
	err = cmd.Process.Kill()
	if err != nil {
		t.Error(err, "kill failed")
	}
}
