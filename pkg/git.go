package pkg

import (
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/go-logr/logr"
	"github.com/gregoryguillou/go-git-http-xfer/githttpxfer"
)

func execCmd(dir string, name string, arg ...string) ([]byte, error) {
	c := exec.Command(name, arg...)
	c.Dir = dir
	return c.CombinedOutput()
}

type gitCommand interface {
	initRepository() error
	cloneRepository() error
	getBin() string
	getRepository() string
	getWorkspaceSync(string) (bool, error)
}

type defaultGitCommand struct {
	bin   string
	store store
	log   logr.Logger
}

func (r *runContainer) newDefaultGitCommand(store store) (gitCommand, error) {
	bin, err := exec.LookPath("git")
	if err != nil {
		r.Log.Info("git not found...")
		return nil, err
	}
	return &defaultGitCommand{
		bin:   bin,
		store: store,
		log:   r.Log}, nil
}

func (git *defaultGitCommand) initRepository() error {
	if _, err := execCmd(git.store.repoDir, git.bin, "init", "--bare", "--shared"); err != nil {
		git.log.Error(err, "could not initialize repository")
		return err
	}
	return nil
}

func (git *defaultGitCommand) cloneRepository() error {
	if _, err := execCmd(git.store.workdir, git.bin, "clome", git.store.repoDir, "."); err != nil {
		git.log.Error(err, "could not clone repository")
		return err
	}
	return nil
}

func (git *defaultGitCommand) getWorkspaceSync(head string) (bool, error) {
	log := git.log
	output, err := os.ReadFile(path.Join(git.store.workdir, ".git/HEAD"))
	if err != nil {
		git.log.Error(err, "cannot read .git/HEAD")
		return false, err
	}
	current := strings.Join(strings.Split(strings.TrimSuffix(string(output), "\n"), "/")[2:], "/")
	if current != head {
		if output, err := execCmd(git.store.repoDir, "git", "fetch", "-p"); err != nil {
			log.Error(err, "could not run git fetch,", "data", string(output))
			return true, err
		}
		if output, err := execCmd(git.store.repoDir, "git", "checkout", head); err != nil {
			log.Error(err, "could not run git checkout,", "data", string(output))
			return true, err
		}
		return true, nil
	}
	if output, err := execCmd(git.store.repoDir, "git", "pull"); err != nil {
		log.Error(err, "could not run git pull,", "data", string(output))
		return false, err
	}
	return false, nil
}

func (git *defaultGitCommand) getBin() string {
	return git.bin
}

func (git *defaultGitCommand) getRepository() string {
	return git.store.repoDir
}

type mockGitCommand struct {
}

func (git *mockGitCommand) initRepository() error {
	return nil
}

func (git *mockGitCommand) cloneRepository() error {
	return nil
}

func (git *mockGitCommand) getBin() string {
	return "git"
}

func (git *mockGitCommand) getRepository() string {
	return "/repository"
}

func (git *mockGitCommand) getWorkspaceSync(head string) (bool, error) {
	return false, nil
}

type gitServer struct {
	repoName   string
	head       string
	gitCommand gitCommand
	ghx        *http.Handler
	action     chan<- string
	log        logr.Logger
}

func (r *runContainer) newGitServer(store store, action chan<- string) (*gitServer, error) {
	log := r.Log.WithName("git")
	command, err := r.newDefaultGitCommand(store)
	if err != nil {
		r.Log.Error(err, "unable to find git")
		return nil, err
	}
	err = os.Chdir(command.getRepository())
	if err != nil {
		r.Log.Error(err, "unable to change directory", "data", command.getRepository())
		return nil, err
	}
	ghx, err := githttpxfer.New(command.getRepository(), command.getBin())
	if err != nil {
		r.Log.Error(err, "unable to create git server instance")
		return nil, err
	}
	ghx.Event.On(githttpxfer.BeforeUploadPack, func(ctx githttpxfer.Context) {
		// log.Info("prepare run service rpc upload.")
	})
	ghx.Event.On(githttpxfer.BeforeReceivePack, func(ctx githttpxfer.Context) {
		// log.Info("prepare run service rpc receive.")
	})
	ghx.Event.On(githttpxfer.AfterMatchRouting, func(ctx githttpxfer.Context) {
		// log.Info("after match routing.")
	})
	if err := command.initRepository(); err != nil {
		return nil, err
	}
	server := &gitServer{
		repoName:   r.Config.Main.Repository,
		head:       r.Config.Main.Head,
		gitCommand: command,
		action:     action,
		log:        log,
	}
	handler := LoggingMiddleware(r.Log.WithName("git"), server.captureAndTrigger(ghx))
	server.ghx = &handler
	return server, nil
}

func (g *gitServer) captureAndTrigger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		method := r.Method
		if len(path) >= len(g.repoName)+1 && path[0:len(g.repoName)+1] == "/"+g.repoName {
			r.URL.Path = path[len(g.repoName)+1:]
		}
		path = r.URL.Path
		next.ServeHTTP(w, r)
		if path == "/git-receive-pack" && method == http.MethodPost {
			g.action <- triggeredMessage
		}
	})
}
