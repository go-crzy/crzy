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

func Test_mockGitFailCommand(t *testing.T) {
	g := &mockGitFailCommand{}
	err := g.cloneRepository()
	if err == nil {
		t.Error("should return an error")
	}
	err = g.initRepository()
	if err == nil {
		t.Error("should return an error")
	}
	err = g.initRepository()
	if err == nil {
		t.Error("should return an error")
	}
	err = g.syncWorkspace("head")
	if err == nil {
		t.Error("should return an error")
	}
	bin := g.getBin()
	if bin != "git" {
		t.Error("should return git")
	}
	repo := g.getRepository()
	if repo != "/repository" {
		t.Error("should return /repository")
	}
	repo = g.getWorkspace()
	if repo != "/workspace" {
		t.Error("should return /workspace")
	}
	repo = g.getExecdir()
	if repo != "/executions" {
		t.Error("should return /executions")
	}
}

func Test_mockGitSuccessCommand(t *testing.T) {
	g := &mockGitSuccessCommand{}
	err := g.cloneRepository()
	if err != nil {
		t.Error("should not return an error")
	}
	err = g.initRepository()
	if err != nil {
		t.Error("should not return an error")
	}
	err = g.initRepository()
	if err != nil {
		t.Error("should not return an error")
	}
	err = g.syncWorkspace("head")
	if err != nil {
		t.Error("should not return an error")
	}
	bin := g.getBin()
	if bin != "git" {
		t.Error("should return git")
	}
	repo := g.getRepository()
	if repo != "/repository" {
		t.Error("should return /repository")
	}
	repo = g.getWorkspace()
	if repo != "/workspace" {
		t.Error("should return /workspace")
	}
	repo = g.getExecdir()
	if repo != "/executions" {
		t.Error("should return /executions")
	}
}
