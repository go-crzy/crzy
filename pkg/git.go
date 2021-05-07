package pkg

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-logr/logr"
	"github.com/gregoryguillou/go-git-http-xfer/githttpxfer"
)

const (
	workarea  = "workarea"
	artifacts = "execs"
)

var (
	ErrRepositoryNotSync = errors.New("notsync")
	ErrCommitNotFound    = errors.New("notfound")
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
	log         logr.Logger
}

func NewGitServer(
	repository, head string,
	upstream Upstream,
	action chan<- func()) (*GitServer, error) {
	log := NewLogger("git")

	gitBinPath, err := exec.LookPath("git")
	if err != nil {
		log.Info("git not found...")
		return nil, err
	}
	gitRootPath, err := os.MkdirTemp("", "crzy")
	if err != nil {
		log.Info("unable to create temporary directory")
		return nil, err
	}
	err = os.Chdir(gitRootPath)
	if err != nil {
		log.Error(err, "unable to chdir", "data", gitRootPath)
		return nil, err
	}

	ghx, err := githttpxfer.New(gitRootPath, gitBinPath)
	if err != nil {
		log.Error(err, "GitHTTPXfer instance could not be created")
		return nil, err
	}

	ghx.Event.On(githttpxfer.BeforeUploadPack, func(ctx githttpxfer.Context) {
		log.Info("prepare run service rpc upload.")
	})
	ghx.Event.On(githttpxfer.BeforeReceivePack, func(ctx githttpxfer.Context) {
		log.Info("prepare run service rpc receive.")
	})
	ghx.Event.On(githttpxfer.AfterMatchRouting, func(ctx githttpxfer.Context) {
		log.Info("after match routing.")
	})
	absRepoPath := ghx.Git.GetAbsolutePath(repository)

	os.Mkdir(absRepoPath, os.ModeDir|os.ModePerm)
	if _, err := execCmd(absRepoPath, "git", "init", "--bare", "--shared"); err != nil {
		log.Error(err, "execute command error")
		return nil, err
	}

	os.Mkdir(absRepoPath, os.ModeDir|os.ModePerm)
	workspace, err := filepath.Abs(path.Join(gitRootPath, workarea))
	if err != nil {
		log.Error(err, "Could not get directory", "data", workarea)
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
		log:         log,
	}

	g.ghx = g.Updater(Logging(NewLogger("updater"), ghx))
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
	log := g.log
	f := func() {
		if _, err := os.Stat(g.workspace); err != nil && errors.Is(err, os.ErrNotExist) {
			if _, err := execCmd(g.gitRootPath, "git", "clone", g.absRepoPath, g.workspace); err != nil {
				log.Error(err, "could not clone", "data", g.absRepoPath)
				return
			}
			return
		}
		output, err := os.ReadFile(path.Join(g.workspace, ".git/HEAD"))
		if err != nil {
			log.Error(err, "cannot read .git/HEAD")
			return
		}
		current := strings.Join(strings.Split(strings.TrimSuffix(string(output), "\n"), "/")[2:], "/")
		if current != g.head {
			if output, err := execCmd(g.workspace, "git", "fetch", "-p"); err != nil {
				log.Error(err, "could not run git fetch,", "data", string(output))
				return
			}
			if output, err := execCmd(g.workspace, "git", "checkout", g.head); err != nil {
				log.Error(err, "could not run git checkout,", "data", string(output))
				return
			}
		}
		if output, err := execCmd(g.workspace, "git", "pull"); err != nil {
			log.Error(err, "could not run git pull,", "data", string(output))
			return
		}
		workspace, err := filepath.Abs(path.Join(g.workspace, conf.Deployment.Test.Directory))
		if err != nil {
			log.Error(err, "Could not build path")
			return
		}
		output, err = execCmd(workspace, conf.Deployment.Test.Command, conf.Deployment.Test.Args...)
		for _, v := range strings.Split(string(output), "\n") {
			log.Info(v)
		}
		if err != nil {
			log.Error(err, "tests fail")
			return
		}
		workspace, err = filepath.Abs(path.Join(g.workspace, conf.Version.Directory))
		if err != nil {
			log.Error(err, "Could not build path")
			return
		}
		output, err = execCmd(workspace, conf.Version.Command, conf.Version.Args...)
		if err != nil {
			log.Error(err, "could not get version")
			return
		}
		re := regexp.MustCompile(`([0-9a-f]*)`)
		match := re.FindStringSubmatch(string(output))
		if len(match) < 2 {
			log.Error(errors.New("wrongversion"), string(output))
			return
		}
		version := match[1]
		artipath := path.Join(g.gitRootPath, artifacts)
		if err := os.Mkdir(artipath, os.ModeDir|os.ModePerm); err != nil && !os.IsExist(err) {
			log.Error(err, "artipath directory creation failed", "data", artipath)
			return
		}
		replaceVersion := regexp.MustCompile(`(\$\{version\})`)
		exe := replaceVersion.ReplaceAllString(conf.Deployment.Artifact.Pattern, version)
		artifact := path.Join(artipath, exe)
		args := []string{}
		for _, arg := range conf.Deployment.Build.Args {
			replaceArtifact := regexp.MustCompile(`(\$\{artifact\})`)
			args = append(args, replaceArtifact.ReplaceAllString(arg, artifact))
		}
		workspace, err = filepath.Abs(path.Join(g.workspace, conf.Deployment.Build.Directory))
		if err != nil {
			log.Error(err, "Could not build path")
			return
		}
		output, err = execCmd(workspace, conf.Deployment.Build.Command, args...)
		for _, v := range strings.Split(string(output), "\n") {
			log.Info(v)
		}
		if err != nil {
			log.Error(err, "build fail")
			return
		}
		old, _ := g.upstream.GetDefault()
		_, _, err = g.upstream.Lookup(exe + "/v1")
		if err == nil {
			log.Info("executable is already running", "data", exe)
			return
		}
		addr, err := g.upstream.NextAddr()
		if err != nil {
			log.Error(err, "no address available")
			return
		}
		cmd := exec.Command(artifact)
		cmd.Env = []string{fmt.Sprintf("ADDR=%s", addr)}
		log.Info("starting instance", "data", fmt.Sprintf("%s,%s", exe, addr))
		g.upstream.Register(exe, "v1", HTTPProcess{Addr: addr, Cmd: cmd}, true)
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
		log.Info("stopping instance", "data", fmt.Sprintf("%s,%s", strings.Join(key[0:len(key)-1], "/"), key[len(key)-1]))
		g.upstream.Unregister(strings.Join(key[0:len(key)-1], "/"), key[len(key)-1])
	}
	g.action <- f
}
