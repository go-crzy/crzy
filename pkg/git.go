package pkg

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/gregoryguillou/go-git-http-xfer/githttpxfer"
)

const (
	workarea  = "workarea"
	artifacts = "execs"
)

var (
	ErrRepositoryNotSync = errors.New("notsync")
	ErrCommitNotFound    = errors.New("notfound")
	extension            = ""
)

type GitServer struct {
	gitRootPath string
	gitBinPath  string
	repoName    string
	absRepoPath string
	workspace   string
	head        string
	ghx         http.Handler
	upstream    Upstream
	action      chan<- func()
}

func NewGitServer(
	repository, head string,
	upstream Upstream,
	action chan<- func()) (*GitServer, error) {

	if runtime.GOOS == "windows" {
		extension = ".exe"
	}
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

	os.Mkdir(absRepoPath, os.ModeDir|os.ModePerm)
	workspace, err := filepath.Abs(path.Join(gitRootPath, workarea))
	if err != nil {
		log.Printf("Could not get directory for %s. %v", workarea, err)
		return nil, err
	}

	g := &GitServer{
		gitRootPath: gitRootPath,
		gitBinPath:  gitBinPath,
		repoName:    repository,
		absRepoPath: absRepoPath,
		workspace:   workspace,
		head:        head,
		ghx:         nil,
		upstream:    upstream,
		action:      action,
	}

	g.ghx = g.Updater(Logging(ghx))
	return g, nil

}

func execCmd(dir string, name string, arg ...string) ([]byte, error) {
	c := exec.Command(name, arg...)
	c.Dir = dir
	return c.CombinedOutput()
}

type Updater interface {
	Update(repo string) (string, error)
}

func (g *GitServer) Updater(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		method := r.Method
		next.ServeHTTP(w, r)
		if path == fmt.Sprintf("/%s/git-receive-pack", g.repoName) && method == http.MethodPost {
			g.Update(g.repoName)
		}
	})
}

// Update refresh the workarea from the GIT repository, build the artifact and
// roll the upstream with the latest version
func (g *GitServer) Update(repo string) {
	f := func() {
		if _, err := os.Stat(g.workspace); err != nil && errors.Is(err, os.ErrNotExist) {
			if _, err := execCmd(g.gitRootPath, "git", "clone", g.absRepoPath, g.workspace); err != nil {
				log.Printf("could not clone [%s], error: %v", g.absRepoPath, err)
				return
			}
			return
		}
		if output, err := execCmd(g.workspace, "git", "pull"); err != nil {
			log.Printf("could not run git pull, error: %v\n%q\n", err, output)
			return
		}
		output, err := execCmd(g.workspace, "go", "test", "-v", "./...")
		for _, v := range strings.Split(string(output), "\n") {
			log.Println(v)
		}
		if err != nil {
			log.Printf("tests fail, error: %v", err)
			return
		}
		output, err = execCmd(g.workspace, "git", "log", "-1", "--format=%H", ".")
		if err != nil {
			log.Printf("could not get sha, error: %v", err)
			return
		}
		re := regexp.MustCompile(`([0-9a-f]*)`)
		match := re.FindStringSubmatch(string(output))
		if len(match) < 2 || len(match[1]) != 40 {
			log.Printf("SHA unexpected: %q", output)
			return
		}
		sha := match[1][0:16]
		artipath := path.Join(g.gitRootPath, artifacts)
		if err := os.Mkdir(artipath, os.ModeDir|os.ModePerm); err != nil && !os.IsExist(err) {
			log.Printf("artipath directory [%s] failed with error: %v", artipath, err)
			return
		}
		artifact := fmt.Sprintf("%s/%s-%s%s", artipath, repo, sha, extension)
		exe := fmt.Sprintf("%s-%s%s", repo, sha, extension)
		output, err = execCmd(g.workspace, "go", "build", "-o", artifact, ".")
		for _, v := range strings.Split(string(output), "\n") {
			log.Println(v)
		}
		if err != nil {
			log.Printf("build fail, error: %v", err)
			return
		}
		old, _ := g.upstream.GetDefault()
		_, _, err = g.upstream.Lookup(exe + "/v1")
		if err == nil {
			log.Printf("executable %s already running", exe)
			return
		}
		port, err := g.upstream.NextPort()
		if err != nil {
			log.Printf("no port available: %v", err)
			return
		}
		cmd := exec.Command(artifact)
		cmd.Env = []string{fmt.Sprintf("PORT=%s", port)}
		log.Printf("starting %s/v1 with port %s", exe, port)
		g.upstream.Register(exe, "v1", HTTPProcess{Addr: port, Cmd: cmd}, true)
		cmd.Start()
		if old == "" {
			return
		}
		_, cmd, err = g.upstream.Lookup(old)
		if err != nil {
			return
		}
		cmd.Process.Kill()
		key := strings.Split(old, "/")
		if len(key) < 2 {
			return
		}
		log.Printf("stopping %s/%s", strings.Join(key[0:len(key)-1], "/"), key[len(key)-1])
		g.upstream.Unregister(strings.Join(key[0:len(key)-1], "/"), key[len(key)-1])
	}
	g.action <- f
}
