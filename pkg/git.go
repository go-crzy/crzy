package pkg

import (
	"net/http"
	"os"
	"os/exec"

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
		log.Info("prepare run service rpc upload.")
	})
	ghx.Event.On(githttpxfer.BeforeReceivePack, func(ctx githttpxfer.Context) {
		log.Info("prepare run service rpc receive.")
	})
	ghx.Event.On(githttpxfer.AfterMatchRouting, func(ctx githttpxfer.Context) {
		log.Info("after match routing.")
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

// Update refresh the workarea from the GIT repository, build the artifact and
// roll the upstream with the latest version
// func (g *GitServer) Update(repo string) {
// 	log := g.log
// 	f := func() {
// 		if _, err := os.Stat(g.workspace); err != nil && errors.Is(err, os.ErrNotExist) {
// 			return
// 		}
// 		output, err := os.ReadFile(path.Join(g.workspace, ".git/HEAD"))
// 		if err != nil {
// 			log.Error(err, "cannot read .git/HEAD")
// 			return
// 		}
// 		current := strings.Join(strings.Split(strings.TrimSuffix(string(output), "\n"), "/")[2:], "/")
// 		if current != g.head {
// 			if output, err := execCmd(g.workspace, "git", "fetch", "-p"); err != nil {
// 				log.Error(err, "could not run git fetch,", "data", string(output))
// 				return
// 			}
// 			if output, err := execCmd(g.workspace, "git", "checkout", g.head); err != nil {
// 				log.Error(err, "could not run git checkout,", "data", string(output))
// 				return
// 			}
// 		}
// 		if output, err := execCmd(g.workspace, "git", "pull"); err != nil {
// 			log.Error(err, "could not run git pull,", "data", string(output))
// 			return
// 		}
// 		workspace, err := filepath.Abs(path.Join(g.workspace, conf.Deploy.Test.WorkDir))
// 		if err != nil {
// 			log.Error(err, "Could not build path")
// 			return
// 		}
// 		output, err = execCmd(workspace, conf.Deploy.Test.Command, conf.Deploy.Test.Args...)
// 		for _, v := range strings.Split(string(output), "\n") {
// 			log.Info(v)
// 		}
// 		if err != nil {
// 			log.Error(err, "tests fail")
// 			return
// 		}
// 		workspace, err = filepath.Abs(path.Join(g.workspace, conf.Trigger.Version.WorkDir))
// 		if err != nil {
// 			log.Error(err, "Could not build path")
// 			return
// 		}
// 		output, err = execCmd(workspace, conf.Trigger.Version.Command, conf.Trigger.Version.Args...)
// 		if err != nil {
// 			log.Error(err, "could not get version")
// 			return
// 		}
// 		re := regexp.MustCompile(`([0-9a-f]*)`)
// 		match := re.FindStringSubmatch(string(output))
// 		if len(match) < 2 {
// 			log.Error(errors.New("wrongversion"), string(output))
// 			return
// 		}
// 		version := match[1]
// 		artipath := path.Join(g.gitRootPath, artifacts)
// 		if err := os.Mkdir(artipath, os.ModeDir|os.ModePerm); err != nil && !os.IsExist(err) {
// 			log.Error(err, "artipath directory creation failed", "data", artipath)
// 			return
// 		}
// 		replaceVersion := regexp.MustCompile(`(\$\{version\})`)
// 		exe := replaceVersion.ReplaceAllString("a", version)
// 		artifact := path.Join(artipath, exe)
// 		args := []string{}
// 		for _, arg := range conf.Deploy.Build.Args {
// 			replaceArtifact := regexp.MustCompile(`(\$\{artifact\})`)
// 			args = append(args, replaceArtifact.ReplaceAllString(arg, artifact))
// 		}
// 		workspace, err = filepath.Abs(path.Join(g.workspace, conf.Deploy.Build.WorkDir))
// 		if err != nil {
// 			log.Error(err, "Could not build path")
// 			return
// 		}
// 		output, err = execCmd(workspace, conf.Deploy.Build.Command, args...)
// 		for _, v := range strings.Split(string(output), "\n") {
// 			log.Info(v)
// 		}
// 		if err != nil {
// 			log.Error(err, "build fail")
// 			return
// 		}
// 		old, _ := g.upstream.GetDefault()
// 		_, _, err = g.upstream.Lookup(exe + "/v1")
// 		if err == nil {
// 			log.Info("executable is already running", "data", exe)
// 			return
// 		}
// 		addr, err := g.upstream.NextAddr()
// 		if err != nil {
// 			log.Error(err, "no address available")
// 			return
// 		}
// 		cmd := exec.Command(artifact)
// 		cmd.Env = []string{fmt.Sprintf("ADDR=%s", addr)}
// 		log.Info("starting instance", "data", fmt.Sprintf("%s,%s", exe, addr))
// 		g.upstream.Register(exe, "v1", HTTPProcess{Addr: addr, Cmd: cmd}, true)
// 		cmd.Start()
// 		if old == "" {
// 			return
// 		}
// 		_, cmd, err = g.upstream.Lookup(old)
// 		if err != nil {
// 			return
// 		}
// 		cmd.Process.Kill()
// 		key := strings.Split(old, "/")
// 		if len(key) < 2 {
// 			return
// 		}
// 		log.Info("stopping instance", "data", fmt.Sprintf("%s,%s", strings.Join(key[0:len(key)-1], "/"), key[len(key)-1]))
// 		g.upstream.Unregister(strings.Join(key[0:len(key)-1], "/"), key[len(key)-1])
// 	}
// 	g.action <- f
// }
