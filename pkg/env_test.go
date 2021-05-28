package pkg

import (
	"testing"
)

func Test_get_succeed(t *testing.T) {
	e := envVars{{Name: "version", Value: "123"}}
	out := e.get("version")
	if out != "123" {
		t.Error("getEnv should return 123")
	}
}

func Test_get_failed(t *testing.T) {
	e := envVars{{Name: "version", Value: "123"}}
	out := e.get("release")
	if out != "" {
		t.Error("getEnv should return an empty string")
	}
}

func Test_replaceEnv_and_succeed(t *testing.T) {
	input := []string{`abc-${version}-${x}`, `abc-${version}`, `${version}`}
	envs := newEnvVars(envVar{Name: "version", Value: "123"}, envVar{Name: "x", Value: "abc"})
	expected := []string{`abc-123-abc`, `abc-123`, "123"}

	for k, v := range input {
		output, err := envs.replace(v)
		if err != nil {
			t.Error(err, "error replacing key")
		}
		if output != expected[k] {
			t.Errorf("expected value %s, return %s", expected[k], output)
		}
	}
}

func Test_replaceEnv_and_failure(t *testing.T) {
	input := []string{`abc-${version}-${x}`, `abc-${version}`}
	envs := newEnvVars(envVar{Name: "x", Value: "abc"})

	for _, v := range input {
		if _, err := envs.replace(v); err != errMissingEnv {
			t.Error("we should fail and we are not")
		}
	}
}

func Test_replaceEnv_duplicate_key(t *testing.T) {
	envs := newEnvVars(envVar{Name: "VERSION", Value: "1.0"}, envVar{Name: "VERSION", Value: "2.0"})

	if _, err := envs.replace("abc"); err != errDuplicateKeys {
		t.Error("should fail with errDuplicateKeys, error:", err)
	}
}

func Test_groupEnvs_and_succeed(t *testing.T) {
	input := envVars{{Name: "VERSION", Value: "1.0"}, {Name: "PORT", Value: "8080"}}
	mapOfEnvs, err := input.toMap()
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

func Test_groupEnvs_and_fail(t *testing.T) {
	input := envVars{{Name: "VERSION", Value: "1.0"}, {Name: "VERSION", Value: "2.0"}}
	mapOfEnvs, err := input.toMap()
	if err != errDuplicateKeys {
		t.Error("groupEnvs should return ErrDuplicateKeys")
	}
	if len(mapOfEnvs) != 0 {
		t.Error("mapOfEnvs should be empty")
	}
}
