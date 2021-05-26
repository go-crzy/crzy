package pkg

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/go-logr/logr"
)

type execStruct struct {
	log     logr.Logger
	name    string
	Command string   `json:"command"`
	Args    []string `json:"args,omitempty"`
	WorkDir string   `json:"workdir"`
	Envs    envVars  `json:"envs,omitempty"`
	Output  string   `json:"output,omitempty"`
}

func getCmd(dir string, envs envVars, name string, args ...string) *exec.Cmd {
	c := exec.Command(name, args...)
	c.Dir = dir
	if len(envs) > 0 {
		vars := os.Environ()
		for _, e := range envs {
			vars = append(vars, fmt.Sprintf("%s=%s", e.Name, e.Value))
		}
		c.Env = vars
	}
	return c
}

func (e *execStruct) prepare(workspace string, envs envVars) (*exec.Cmd, error) {
	dir := path.Join(workspace, e.WorkDir)
	e.WorkDir = dir
	command, err := envs.replace(e.Command)
	if err != nil {
		return nil, err
	}
	e.Command = command
	args := []string{}
	full := command
	for _, arg := range e.Args {
		arg, err = envs.replace(arg)
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
		full = fmt.Sprintf("%s %s", full, arg)
	}
	e.Args = args
	for index, value := range e.Envs {
		if v := envs.get(value.Name); v == "" {
			if e.Envs[index].Value, err = envs.replace(value.Value); err != nil {
				return nil, err
			}
		}
	}
	e.log.Info(full)
	return getCmd(dir, e.Envs, command, args...), nil
}
