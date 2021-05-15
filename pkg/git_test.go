package pkg

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"testing"
)

func Test_execCmd_and_succeed(t *testing.T) {
	output, err := execCmd(".", map[string]string{}, "git", "version")
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

func Test_newDefaultGitCommand(t *testing.T) {
	store := store{
		rootDir: "/root",
		execDir: "/root/execs",
		repoDir: "/root/repository",
		workdir: "/root/workspace",
		log:     &mockLogger{},
	}
	r := &runContainer{
		Log:    &mockLogger{},
		Config: config{},
	}
	g, err := r.newDefaultGitCommand(store)
	if err != nil {
		t.Error("should succeed")
	}
	cmd := g.getBin()
	if cmd == "" {
		t.Error("should not be empty")
	}
	repo := g.getRepository()
	if repo != "/root/repository" {
		t.Error("should return /root/repository")
	}
	repo = g.getExecdir()
	if repo != "/root/execs" {
		t.Error("should return /root/execs")
	}
}

func Test_newGitServer(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "crzytest")
	if err != nil {
		t.Error("could not create tmpdir")
		t.FailNow()
	}
	defer os.RemoveAll(tmpdir)
	for _, v := range []string{
		path.Join(tmpdir, "execs"),
		path.Join(tmpdir, "repository"),
		path.Join(tmpdir, "workspace"),
	} {
		err := os.Mkdir(v, os.ModeDir|os.ModePerm)
		if err != nil {
			t.Error("could not dirrectory", v)
			t.FailNow()
		}
	}
	store := store{
		rootDir: tmpdir,
		execDir: path.Join(tmpdir, "execs"),
		repoDir: path.Join(tmpdir, "repository"),
		workdir: path.Join(tmpdir, "workspace"),
		log:     &mockLogger{},
	}
	r := &runContainer{
		Log:    &mockLogger{},
		Config: config{},
	}
	action := make(chan event)
	_, err = r.newGitServer(store, action)
	if err != nil {
		t.Error("should succeed", err)
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

func Test_captureAndTrigger_and_event(t *testing.T) {
	action := make(chan event, 1)
	g := &gitServer{
		action:   action,
		repoName: "color.git",
	}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	h := g.captureAndTrigger(next)
	server := httptest.NewServer(h)
	client := server.Client()

	request, _ := http.NewRequest("POST", server.URL+"/color.git/git-receive-pack", nil)
	response, err := client.Do(request)
	if err != nil {
		t.Errorf("Should not return %v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf(
			"Status Code should be 200, current: %d",
			response.StatusCode,
		)
	}
	b, err := io.ReadAll(response.Body)
	if err != nil {
		t.Error("body conversion should succeed")
	}
	body := strings.Split(string(b), "\n")[0]
	if body != `ok` {
		t.Errorf("message should be ok, >%s<", body)
	}
	val := <-action
	if val.id != triggeredMessage {
		t.Error("should trigger an action")
	}
}

func Test_captureAndTrigger_and_skip(t *testing.T) {
	action := make(chan event, 1)
	g := &gitServer{
		action:   action,
		repoName: "color.git",
	}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	h := g.captureAndTrigger(next)
	server := httptest.NewServer(h)
	client := server.Client()

	request, _ := http.NewRequest("POST", server.URL+"/git-receive-pack", nil)
	response, err := client.Do(request)
	if err != nil {
		t.Errorf("Should not return %v", err)
	}
	if response.StatusCode != http.StatusNotFound {
		t.Errorf(
			"Status Code should be 404, current: %d",
			response.StatusCode,
		)
	}
}
