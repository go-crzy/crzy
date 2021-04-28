package crzy

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/sosedoff/gitkit"
)

func NewGITServer(dir string) http.Handler {
	// Configure git hooks
	// hooks := &gitkit.HookScripts{
	// 	PreReceive: `echo "Hello World!"`,
	// }

	// Configure git service
	service := gitkit.New(gitkit.Config{
		Dir:        dir,
		AutoCreate: true,
		AutoHooks:  true,
		// Hooks:      hooks,
	})

	// Configure git server. Will create git repos path if it does not exist.
	// If hooks are set, it will also update all repos with new version of hook scripts.
	if err := service.Setup(); err != nil {
		log.Fatal(err)
	}
	return service
}

var (
	ErrRepositoryNotSync = errors.New("notsync")
	ErrCommitNotFound    = errors.New("notfound")
)

type Updater interface {
	Update(repo string) (string, error)
}

type DefaultUpdater struct {
	WorkspaceDir string
	action       chan<- func()
	upstream     Upstreamer
}

func NewUpdater(workspace string, upstream Upstreamer, action chan<- func()) (Updater, error) {
	if workspace == "" {
		return nil, errors.New("unknownworkspace")
	}
	if err := os.Mkdir(workspace, os.ModeDir|os.ModePerm); err != nil && !os.IsExist(err) {
		return nil, err
	}
	return &DefaultUpdater{
		WorkspaceDir: workspace,
		action:       action,
		upstream:     upstream,
	}, nil
}

// Update refresh the workarea and returns the 16-characters of the last commit
func (git *DefaultUpdater) Update(repo string) (string, error) {
	workarea := "workarea"
	workpath := fmt.Sprintf("%s/%s", git.WorkspaceDir, workarea)
	workrepo := fmt.Sprintf("%s/%s/%s", git.WorkspaceDir, workarea, repo)
	if err := os.Mkdir(workpath, os.ModeDir|os.ModePerm); err != nil && !os.IsExist(err) {
		return "", err
	}
	artifacts := "execs"
	extension := ""
	if runtime.GOOS == "windows" {
		extension = ".exe"
	}
	artipath := fmt.Sprintf("%s/%s", git.WorkspaceDir, artifacts)
	if err := os.Mkdir(artipath, os.ModeDir|os.ModePerm); err != nil && !os.IsExist(err) {
		return "", err
	}
	f := func() {
		if _, err := os.Stat(workrepo); os.IsNotExist(err) {
			repopath := fmt.Sprintf("%s/%s/.git", git.WorkspaceDir, repo)
			if _, err := os.Stat(repopath); os.IsNotExist(err) {
				log.Printf("sync failed, repository [%s] does not exists", repopath)
				return
			}
			cmd := exec.Command("git", "clone", repopath, repo)
			cmd.Dir = workpath
			err := cmd.Run()
			if err != nil {
				log.Printf("git clone failed; error: %v", err)
				return
			}
		}
		cmd := exec.Command("git", "pull")
		cmd.Dir = workrepo
		err := cmd.Run()
		if err != nil {
			log.Printf("git pull [%s] failed; error: %v", repo, err)
			return
		}
		cmd = exec.Command("git", "log", "-1", "--format=%H", ".")
		cmd.Dir = workrepo
		out, err := cmd.Output()
		if err != nil {
			log.Printf("git log failed; error: %v", err)
			return
		}
		re := regexp.MustCompile(`([0-9a-f]*)`)
		match := re.FindStringSubmatch(string(out))
		if len(match) < 2 || len(match[1]) != 40 {
			log.Printf("SHA unexpected: %q", out)
			return
		}
		sha := match[1][0:16]
		artifact := fmt.Sprintf("%s/%s/%s-%s%s", git.WorkspaceDir, artifacts, repo, sha, extension)
		exe := fmt.Sprintf("%s-%s%s", repo, sha, extension)
		cmd = exec.Command("go", "test", "-v", "./...")
		cmd.Dir = workrepo
		out, err = cmd.Output()
		if err != nil {
			log.Printf("Error testing the project: %v", err)
			return
		}
		for _, v := range strings.Split(string(out), "\n") {
			log.Println(v)
		}
		cmd = exec.Command("go", "build", "-o", artifact, ".")
		cmd.Dir = workrepo
		err = cmd.Run()
		if err != nil {
			log.Printf("Error building project: %v", err)
			return
		}
		old, _ := git.upstream.GetDefault()
		_, _, err = git.upstream.Lookup(exe + "/v1")
		if err == nil {
			log.Printf("executable %s already running", exe)
			return
		}
		port, err := git.upstream.NextPort()
		if err != nil {
			log.Printf("no port available: %v", err)
			return
		}
		cmd = exec.Command(artifact)
		cmd.Env = []string{fmt.Sprintf("PORT=%s", port)}
		log.Printf("starting %s/v1 with port %s", exe, port)
		git.upstream.Register(exe, "v1", HTTPProcess{Addr: port, Cmd: cmd}, true)
		cmd.Start()
		if old == "" {
			return
		}
		_, cmd, err = git.upstream.Lookup(old)
		if err != nil {
			return
		}
		cmd.Process.Kill()
		key := strings.Split(old, "/")
		if len(key) < 2 {
			return
		}
		log.Printf("stopping %s/%s", strings.Join(key[0:len(key)-1], "/"), key[len(key)-1])
		git.upstream.Unregister(strings.Join(key[0:len(key)-1], "/"), key[len(key)-1])
	}
	git.action <- f
	return "", nil
}
