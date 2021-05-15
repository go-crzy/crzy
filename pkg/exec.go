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
	Envs    []envVar
	Output  string
}

func getCmd(dir string, envs map[string]string, name string, arg ...string) *exec.Cmd {
	c := exec.Command(name, arg...)
	c.Dir = dir
	if len(envs) > 0 {
		vars := []string{}
		for k, v := range envs {
			vars = append(vars, fmt.Sprintf("%s=%s", k, v))
		}
		c.Env = vars
	}
	return c
}

func (e *execStruct) prepare(workspace string, envs map[string]string) (*exec.Cmd, error) {
	dir := path.Join(workspace, e.WorkDir)
	command, err := replaceEnvs(e.Command, envs)
	if err != nil {
		return nil, err
	}
	args := []string{}
	full := command
	for _, arg := range e.Args {
		arg, err = replaceEnvs(arg, envs)
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
		full = fmt.Sprintf("%s %s", full, arg)
	}
	for _, value := range e.Envs {
		if _, ok := envs[value.Name]; !ok {
			v, err := replaceEnvs(value.Value, envs)
			if err != nil {
				return nil, err
			}
			envs[value.Name] = v
		}
	}
	e.log.Info(full)
	return getCmd(dir, envs, command, args...), nil
}

func (e *execStruct) run(workspace string, envs map[string]string) (*envVar, error) {
	cmd, err := e.prepare(workspace, envs)
	if err != nil {
		return nil, err
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	results := strings.Split(string(output), "\n")
	for _, v := range results {
		e.log.Info(v)
	}
	if e.Output != "" {
		return &envVar{Name: e.Output, Value: results[0]}, nil
	}
	return nil, nil
}

func (e *execStruct) runBackground(workspace string, envs map[string]string) (*os.Process, error) {
	cmd, err := e.prepare(workspace, envs)
	if err != nil {
		return nil, err
	}
	err = cmd.Start()
	return cmd.Process, err
}
