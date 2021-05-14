package pkg

import (
	"errors"
	"regexp"
)

type envVar struct {
	Name  string
	Value string
}

var (
	ErrMissingEnv    = errors.New("missing")
	ErrDuplicateKeys = errors.New("dupkeys")
	envPattern       = regexp.MustCompile(`(\$\{[a-zA-Z0-9_]*\})`)
)

// replaceEnvs replaces variables identified with ${} in param with their
// values picked from the envs map. If one value is missing, it returns the
// ErrMissingEnv error.
func replaceEnvs(param string, envs map[string]string) (string, error) {
	submatches := envPattern.FindAllStringSubmatch(param, -1)
	for _, match := range submatches {
		key := match[0][2 : len(match[0])-1]
		val, ok := envs[key]
		if !ok {
			return "", ErrMissingEnv
		}
		envPatternWithKey := regexp.MustCompile(`(\$\{` + key + `\})`)
		param = envPatternWithKey.ReplaceAllString(param, val)
	}
	return param, nil
}

// groupEnvs returns a map of values built with EnvVar and detect when there
// is a duplicate key in the list
func groupEnvs(envs ...envVar) (map[string]string, error) {
	keys := map[string]string{}
	for _, v := range envs {
		if _, ok := keys[v.Name]; ok {
			return map[string]string{}, ErrDuplicateKeys
		}
		keys[v.Name] = v.Value
	}
	return keys, nil
}
