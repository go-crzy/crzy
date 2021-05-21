package pkg

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/go-logr/logr"
)

type execStruct struct {
	log     logr.Logger
	Command string
	Args    []string
	WorkDir string
	Envs    envVars
	Output  string
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
	command, err := envs.replace(e.Command)
	if err != nil {
		return nil, err
	}
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
	for _, value := range e.Envs {
		if v := envs.get(value.Name); v == "" {
			if _, err := envs.replace(value.Value); err != nil {
				return nil, err
			}
		}
	}
	e.log.Info(full)
	return getCmd(dir, envs, command, args...), nil
}

func (e *execStruct) run(workspace string, envs envVars) (*envVar, error) {
	cmd, err := e.prepare(workspace, envs)
	if err != nil {
		return nil, err
	}
	output, err := cmd.CombinedOutput()
	results := strings.Split(string(output), "\n")
	for _, v := range results {
		e.log.Info(v)
	}
	if err != nil {
		return nil, err
	}
	if e.Output != "" {
		return &envVar{Name: e.Output, Value: results[0]}, nil
	}
	return nil, nil
}

func (e *execStruct) runBackground(workspace string, envs envVars) (*os.Process, error) {
	cmd, err := e.prepare(workspace, envs)
	if err != nil {
		return nil, err
	}
	err = cmd.Start()
	return cmd.Process, err
}
