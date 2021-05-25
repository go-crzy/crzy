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

func Test_runBackground_no_envs(t *testing.T) {
	e := &execStruct{
		log:     &mockLogger{},
		Command: "tail",
		Args:    []string{"-f", "config.go"},
		WorkDir: ".",
		Envs:    envVars{},
	}
	if runtime.GOOS == "windows" {
		e.Command = "powershell"
		e.Args = []string{"-Command", "Get-Content config.go -Wait"}
	}
	p, err := e.runBackground(".", envVars{})
	if err != nil {
		t.Error(err, "start failed")
		t.FailNow()
	}
	if p == nil {
		t.Error("process is empty")
	}
	err = p.Kill()
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

func Test_run_and_succeed(t *testing.T) {
	e := &execStruct{
		log:     &mockLogger{},
		Command: "git",
		Args:    []string{"version"},
		WorkDir: ".",
		Envs:    envVars{},
		Output:  "",
	}
	_, err := e.run(".", envVars{})
	if err != nil {
		t.Error("should succeed")
	}
}

func Test_run_and_succeed_with_output(t *testing.T) {
	e := &execStruct{
		log:     &mockLogger{},
		Command: "git",
		Args:    []string{"version"},
		WorkDir: ".",
		Envs:    envVars{},
		Output:  "data",
	}
	env, err := e.run(".", envVars{})
	if err != nil {
		t.Error("should succeed")
	}
	if env == nil || env.Name != "data" ||
		len(env.Value) < 11 || env.Value[0:11] != "git version" {
		t.Error("should return data=\"git version\"", env.Value)
	}
}

func Test_run_and_fail_command(t *testing.T) {
	e := &execStruct{
		log:     &mockLogger{},
		Command: "${xxx}",
		Args:    []string{"-f", "config.go"},
		WorkDir: ".",
		Envs:    envVars{},
		Output:  "",
	}
	_, err := e.run(".", envVars{})
	if err != errMissingEnv {
		t.Error(err, "should fail")
	}
}

func Test_run_and_fail_combinedoutput(t *testing.T) {
	e := &execStruct{
		log:     &mockLogger{},
		Command: "doesnotexist",
		Args:    []string{"-f", "config.go"},
		WorkDir: ".",
		Envs:    envVars{},
	}
	_, err := e.run(".", envVars{})
	if err == nil ||
		(err.Error() != "exec: \"doesnotexist\": executable file not found in $PATH" &&
			err.Error() != "exec: \"doesnotexist\": executable file not found in %PATH%") {

		t.Error(err, "should fail")
	}
}

func Test_runBackground_and_fail_command(t *testing.T) {
	e := &execStruct{
		log:     &mockLogger{},
		Command: "${xxx}",
		Args:    []string{"-f", "config.go"},
		WorkDir: ".",
		Envs:    envVars{},
	}
	_, err := e.runBackground(".", envVars{})
	if err != errMissingEnv {
		t.Error(err, "should fail")
	}
}

func Test_runBackground_with_envs(t *testing.T) {
	e := &execStruct{
		log:     &mockLogger{},
		Command: "tail",
		Args:    []string{"-f", "config.go"},
		WorkDir: ".",
		Envs:    envVars{{Name: "port", Value: "1234"}},
	}
	if runtime.GOOS == "windows" {
		e.Command = "powershell"
		e.Args = []string{"-Command", "Get-Content config.go -Wait"}
	}
	_, err := e.Envs.toMap()
	if err != nil {
		t.Error(err, "should be able to convert envs")
		t.FailNow()
	}
	p, err := e.runBackground(".", e.Envs)
	if err != nil {
		t.Error(err, "start failed")
		t.FailNow()
	}
	if p == nil {
		t.Error("process is empty")
	}
	err = p.Kill()
	if err != nil {
		t.Error(err, "kill failed")
	}
}
