package pkg

import (
	"os"
	"path"

	"github.com/go-logr/logr"
)

type store struct {
	rootDir string
	repoDir string
	execDir string
	workdir string
	log     logr.Logger
}

func (r *runContainer) createStore() (*store, error) {
	log := r.Log.WithName("store")
	rootDir, err := os.MkdirTemp("", "crzy")
	if err != nil {
		log.Info("unable to create temporary directory")
		return nil, err
	}
	repoDir := path.Join(rootDir, "repository")
	workDir := path.Join(rootDir, "workspace")
	execDir := path.Join(rootDir, "execs")
	for _, dir := range []string{repoDir, execDir, workDir} {
		err = os.Mkdir(dir, os.ModeDir|os.ModePerm)
		if err != nil {
			log.Info("unable to create directory")
			return nil, err
		}
	}
	log.Info("directory created", "data", rootDir)
	return &store{
		execDir: execDir,
		log:     log,
		repoDir: repoDir,
		rootDir: rootDir,
		workdir: workDir,
	}, nil
}

func (s *store) delete() error {
	err := os.RemoveAll(s.rootDir)
	if err != nil {
		s.log.Error(err, "error deleting store...")
		return err
	}
	s.log.Info("store deleted with success....")
	return nil
}
