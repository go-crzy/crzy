package pkg

import (
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/go-crzy/go-git-http-xfer/githttpxfer"
	"github.com/go-logr/logr"
)

type gitCommand interface {
	initRepository() error
	cloneRepository() error
	getBin() string
	getRepository() string
	getWorkspace() string
	getExecdir() string
	syncWorkspace(string) error
}

type defaultGitCommand struct {
	bin   string
	store store
	log   logr.Logger
}

func (r *defaultContainer) newDefaultGitCommand(store store) (gitCommand, error) {
	bin, err := exec.LookPath("git")
	if err != nil {
		r.log.Info("git not found...")
		return nil, err
	}
	return &defaultGitCommand{bin: bin, store: store, log: r.log}, nil
}

func (git *defaultGitCommand) initRepository() error {
	if _, err := getCmd(git.store.repoDir, envVars{}, git.bin, "init", "--bare", "--shared").CombinedOutput(); err != nil {
		git.log.Error(err, "could not initialize repository")
		return err
	}
	return nil
}

func (git *defaultGitCommand) cloneRepository() error {
	if _, err := getCmd(git.store.workdir, envVars{}, git.bin, "clone", git.store.repoDir, ".").CombinedOutput(); err != nil {
		git.log.Error(err, "could not clone repository")
		return err
	}
	return nil
}

func (git *defaultGitCommand) syncWorkspace(head string) error {
	log := git.log
	output, err := os.ReadFile(path.Join(git.store.workdir, ".git/HEAD"))
	if err != nil {
		git.log.Error(err, "cannot read .git/HEAD")
		return err
	}
	current := strings.Join(strings.Split(strings.TrimSuffix(string(output), "\n"), "/")[2:], "/")
	if current != head {
		if output, err := getCmd(git.store.workdir, envVars{}, "git", "fetch", "-p").CombinedOutput(); err != nil {
			log.Error(err, "could not run git fetch,", "data", string(output))
			return err
		}
		if output, err := getCmd(git.store.workdir, envVars{}, "git", "checkout", head).CombinedOutput(); err != nil {
			log.Error(err, "could not run git checkout,", "data", string(output))
			return err
		}
		return nil
	}
	if output, err := getCmd(git.store.workdir, envVars{}, "git", "pull").CombinedOutput(); err != nil {
		log.Error(err, "could not run git pull,", "data", string(output))
		return err
	}
	return nil
}

func (git *defaultGitCommand) getBin() string {
	return git.bin
}

func (git *defaultGitCommand) getRepository() string {
	return git.store.repoDir
}

func (git *defaultGitCommand) getWorkspace() string {
	return git.store.workdir
}

func (git *defaultGitCommand) getExecdir() string {
	return git.store.execDir
}

type gitServer struct {
	repoName   string
	head       string
	gitCommand gitCommand
	ghx        *http.Handler
	action     chan<- event
	release    chan<- event
	log        logr.Logger
	state      *stateManager
}

func (r *defaultContainer) newGitServer(store store, state *stateManager, action chan<- event, release chan<- event) (*gitServer, error) {
	log := r.log.WithName("git")
	command, err := r.newDefaultGitCommand(store)
	if err != nil {
		r.log.Error(err, "unable to find git")
		return nil, err
	}
	err = os.Chdir(command.getRepository())
	if err != nil {
		r.log.Error(err, "unable to change directory", "data", command.getRepository())
		return nil, err
	}
	ghx, err := githttpxfer.New(command.getRepository(), command.getBin())
	if err != nil {
		r.log.Error(err, "unable to create git server instance")
		return nil, err
	}
	// prepare run service rpc upload.
	ghx.Event.On(githttpxfer.BeforeUploadPack, func(ctx githttpxfer.Context) {})
	// prepare run service rpc receive."
	ghx.Event.On(githttpxfer.BeforeReceivePack, func(ctx githttpxfer.Context) {})
	// after match routing.
	ghx.Event.On(githttpxfer.AfterMatchRouting, func(ctx githttpxfer.Context) {})
	if err := command.initRepository(); err != nil {
		return nil, err
	}
	server := &gitServer{
		repoName:   r.config.Main.Repository,
		head:       r.config.Main.Head,
		gitCommand: command,
		action:     action,
		release:    release,
		log:        log,
		state:      state,
	}
	handler := loggingMiddleware(r.log.WithName("git"), server.captureAndTrigger(ghx))
	handler = r.config.authMiddleware(handler)
	server.ghx = &handler
	return server, nil
}

func (g *gitServer) captureAndTrigger(next http.Handler) http.Handler {
	mux := newAPI(g.state)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		method := r.Method
		if len(path) >= 3 && path[:3] == "/v0" {
			mux.ServeHTTP(w, r)
			return
		}
		if len(path) >= len(g.repoName)+1 && path[:len(g.repoName)+1] == "/"+g.repoName {
			r.URL.Path = path[len(g.repoName)+1:]
		}
		if path == r.URL.Path {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		path = r.URL.Path
		next.ServeHTTP(w, r)
		if path == "/git-receive-pack" && method == http.MethodPost {
			g.action <- event{id: triggeredMessage}
		}
	})
}
