package pkg

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

	"github.com/gregoryguillou/go-git-http-xfer/githttpxfer"
)

type GitServer struct {
	gitRootPath string
	gitBinPath  string
	repoName    string
	absRepoPath string
	head        string
	ghx         *githttpxfer.GitHTTPXfer
}

func NewGitServer(repository, head string) (*GitServer, error) {

	gitBinPath, err := exec.LookPath("git")
	if err != nil {
		log.Println("git not found...")
		return nil, err
	}
	gitRootPath, err := os.MkdirTemp("", "crzy")
	if err != nil {
		log.Println("unable to create temporary directory")
		return nil, err
	}
	err = os.Chdir(gitRootPath)
	if err != nil {
		log.Printf("unable to chdir to %s, %v", gitRootPath, err)
		return nil, err
	}

	ghx, err := githttpxfer.New(gitRootPath, gitBinPath)
	if err != nil {
		log.Printf("GitHTTPXfer instance could not be created. %v", err)
		return nil, err
	}

	ghx.Event.On(githttpxfer.BeforeUploadPack, func(ctx githttpxfer.Context) {
		log.Printf("prepare run service rpc upload.")
	})
	ghx.Event.On(githttpxfer.BeforeReceivePack, func(ctx githttpxfer.Context) {
		log.Printf("prepare run service rpc receive.")
	})
	ghx.Event.On(githttpxfer.AfterMatchRouting, func(ctx githttpxfer.Context) {
		log.Printf("after match routing.")
	})
	absRepoPath := ghx.Git.GetAbsolutePath(repository)

	os.Mkdir(absRepoPath, os.ModeDir|os.ModePerm)
	if _, err := execCmd(absRepoPath, "git", "init", "--bare", "--shared"); err != nil {
		log.Printf("execute command error: %s", err.Error())
		return nil, err
	}

	return &GitServer{
		gitRootPath: gitRootPath,
		gitBinPath:  gitBinPath,
		repoName:    repository,
		absRepoPath: absRepoPath,
		head:        head,
		ghx:         ghx,
	}, nil
}

func (g *GitServer) cleanupRepository() {
	os.RemoveAll(g.gitRootPath)
}

func (g *GitServer) Updater(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
}

func execCmd(dir string, name string, arg ...string) ([]byte, error) {
	c := exec.Command(name, arg...)
	c.Dir = dir
	return c.CombinedOutput()
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
