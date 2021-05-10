package pkg

import (
	"os"
	"reflect"
	"runtime"
	"testing"
)

func Test_defaultConf_and_succeed(t *testing.T) {
	d := &config{
		Main: MainStruct{
			Head:       "main",
			Server:     true,
			Color:      true,
			Repository: "myrepo",
			ApiPort:    8080,
			ProxyPort:  8081,
		},
		Version: ExecStruct{
			Command:   "git",
			Args:      []string{"log", "-1", "--format=%H", "."},
			Directory: ".",
		},
		Deployment: DeploymentStruct{
			Artifact: ArtifactStruct{
				Type:    "executable",
				Pattern: `go-${version}`,
			},
			Build: ExecStruct{
				Command:   "go",
				Args:      []string{"build", "-o", `${artifact}`, "."},
				Directory: ".",
			},
			Test: ExecStruct{
				Command:   "go",
				Args:      []string{"test", "./..."},
				Directory: ".",
			},
		},
	}
	if runtime.GOOS == "windows" {
		d.Deployment.Artifact.Pattern += ".exe"
	}
	c, err := defaultConf("golang")
	if err != nil {
		t.Error("expect defaultConf with golang to succeed")
	}
	if !reflect.DeepEqual(c, d) {
		t.Error("Parsed and original values do not match")
	}
}

func Test_defaultConf_and_fail(t *testing.T) {
	_, err := defaultConf("java")
	if err != ErrUnsupportedLang {
		t.Error("java should not be supported for now")
	}
}

func Test_getConfig_and_fail(t *testing.T) {
	_, err := getConfig("golang", "fail.yaml")
	if err != ErrLoadingConfigFile {
		t.Error("loading the file should return an error")
	}
}

func Test_getConfig_without_file_and_succeed(t *testing.T) {
	_, err := getConfig("golang", defaultConfigFile)
	if err != nil {
		t.Error("should not fail if the file is the default file")
	}
}

func Test_getConfig_with_file_and_succeed(t *testing.T) {
	input, err := os.Open("templates/golang.yaml")
	if err != nil {
		t.Error("templates/golang.yaml should exist", err)
	}
	defer input.Close()
	f, err := os.CreateTemp(".", "")
	if err != nil {
		t.Error("should be able to create a file", err)
	}
	defer os.Remove(f.Name())
	defer f.Close()
	f.ReadFrom(input)
	_, err = getConfig("golang", f.Name())
	if err != nil {
		t.Error("should be able to read file")
	}
}
