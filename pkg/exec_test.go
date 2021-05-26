package pkg

import (
	"os/exec"
	"runtime"
	"testing"
)

func Test_getCmd_and_succeed(t *testing.T) {
	output, err := getCmd(".", envVars{}, "git", "version").CombinedOutput()
	if err != nil {
		t.Error("test fails", err)
	}
	if string(output[0:11]) != "git version" {
		t.Errorf("output should be git version, current %q", output)
	}
}

func Test_getCmd_with_parameters(t *testing.T) {
	output, err := getCmd(".", envVars{{"version", "version"}}, "git", "version").CombinedOutput()
	if err != nil {
		t.Error("test fails", err)
	}
	if string(output[0:11]) != "git version" {
		t.Errorf("output should be git version, current %q", output)
	}
}

func Test_killProcess(t *testing.T) {
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

func Test_prepare_and_fail_command(t *testing.T) {
	e := &execStruct{
		log:     &mockLogger{},
		Command: "${xxx}",
		Args:    []string{"-f", "config.go"},
		WorkDir: ".",
		Envs:    envVars{},
	}
	_, err := e.prepare(".", envVars{})
	if err != errMissingEnv {
		t.Error(err, "should fail")
	}
}

func Test_prepare_and_fail_args(t *testing.T) {
	e := &execStruct{
		log:     &mockLogger{},
		Command: "tail",
		Args:    []string{"${xxx}", "config.go"},
		WorkDir: ".",
		Envs:    envVars{},
	}
	_, err := e.prepare(".", envVars{})
	if err != errMissingEnv {
		t.Error(err, "should fail")
	}
}

func Test_prepare_and_fail_envs(t *testing.T) {
	e := &execStruct{
		log:     &mockLogger{},
		Command: "tail",
		Args:    []string{"-f", "config.go"},
		WorkDir: ".",
		Envs:    envVars{{Name: "xxx", Value: "${xxx}"}},
	}
	_, err := e.prepare(".", envVars{})
	if err != errMissingEnv {
		t.Error(err, "should fail")
	}
}
