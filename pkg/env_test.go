package pkg

import (
	"testing"
)

func Test_ReplaceEnv_and_succeed(t *testing.T) {
	input := []string{`abc-${version}-${x}`, `abc-${version}`, `${version}`}
	envs := map[string]string{"version": "123", "x": "abc"}
	expected := []string{`abc-123-abc`, `abc-123`, "123"}

	for k, v := range input {
		output, err := replaceEnvs(v, envs)
		if err != nil {
			t.Error(err, "error replacing key")
		}
		if output != expected[k] {
			t.Errorf("expected value %s, return %s", expected[k], output)
		}
	}
}

func Test_ReplaceEnv_and_failure(t *testing.T) {
	input := []string{`abc-${version}-${x}`, `abc-${version}`}
	envs := map[string]string{"x": "abc"}

	for _, v := range input {
		_, err := replaceEnvs(v, envs)
		if err != ErrMissingEnv {
			t.Error("we should fail and we are not")
		}
	}
}

func Test_GroupEnvs_and_succeed(t *testing.T) {
	input := []envVar{{Name: "VERSION", Value: "1.0"}, {Name: "PORT", Value: "8080"}}
	mapOfEnvs, err := groupEnvs(input...)
	if err != nil {
		t.Error("groupEnvs should succeed")
	}
	version, ok := mapOfEnvs["VERSION"]
	if !ok {
		t.Error("We should have a version")
	}
	if version != "1.0" {
		t.Errorf("version should be 1.0, it is %s", version)
	}
	port, ok := mapOfEnvs["PORT"]
	if !ok {
		t.Error("We should have a version")
	}
	if port != "8080" {
		t.Errorf("port should be 8080, it is %s", port)
	}
}

func Test_GroupEnvs_and_fail(t *testing.T) {
	input := []envVar{{Name: "VERSION", Value: "1.0"}, {Name: "VERSION", Value: "2.0"}}
	mapOfEnvs, err := groupEnvs(input...)
	if err != ErrDuplicateKeys {
		t.Error("groupEnvs should return ErrDuplicateKeys")
	}
	if len(mapOfEnvs) != 0 {
		t.Error("mapOfEnvs should be empty")
	}
}
