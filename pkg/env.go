package pkg

import (
	"errors"
	"regexp"
)

type envVar struct {
	Name  string
	Value string
}

type envVars []envVar

var (
	errMissingEnv    = errors.New("missing")
	errDuplicateKeys = errors.New("dupkeys")
	envPattern       = regexp.MustCompile(`(\$\{[a-zA-Z0-9_]*\})`)
)

func newEnvVars(evs ...envVar) envVars { return append(envVars{}, evs...) }

func (evs *envVars) add(e ...envVar) { *evs = append(*evs, e...) }

func (evs *envVars) addOne(n, v string) {
	*evs = append(*evs, envVar{Name: n, Value: v})
}

func (evs envVars) get(name string) string {
	for _, v := range evs {
		if v.Name == name {
			return v.Value
		}
	}
	return ""
}

// replaceEnvs replaces variables identified with ${} in param with their
// values picked from the envs map. If one value is missing, it returns the
// ErrMissingEnv error.
func (evs envVars) replace(param string) (string, error) {
	envs, err := evs.toMap()
	if err != nil {
		return "", err
	}
	submatches := envPattern.FindAllStringSubmatch(param, -1)
	for _, match := range submatches {
		key := match[0][2 : len(match[0])-1]
		val, ok := envs[key]
		if !ok {
			return "", errMissingEnv
		}
		envPatternWithKey := regexp.MustCompile(`(\$\{` + key + `\})`)
		param = envPatternWithKey.ReplaceAllString(param, val)
	}
	return param, nil
}

// toMap returns a map of values built with EnvVar and detect when there
// is a duplicate key in the list
func (evs envVars) toMap() (map[string]string, error) {
	keys := map[string]string{}
	for _, v := range evs {
		if _, ok := keys[v.Name]; ok {
			return map[string]string{}, errDuplicateKeys
		}
		keys[v.Name] = v.Value
	}
	return keys, nil
}
