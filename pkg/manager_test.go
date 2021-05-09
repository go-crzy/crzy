package pkg

import (
	"embed"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

//go:embed templates/*
var langTemplate embed.FS

// Test_Template_golang checks the template matches the default Golang Model
// as it is expected to work
func Test_Template_golang(t *testing.T) {
	c := config{}
	golang, err := langTemplate.ReadFile("templates/golang.yaml")
	if err != nil {
		t.Error(err, "error reading file")
	}
	err = yaml.Unmarshal(golang, &c)
	if err != nil {
		t.Error(err, "error Unmarshalling file")
	}
	d := config{
		Main: MainStruct{
			Head:       "main",
			Server:     true,
			Color:      true,
			Repository: "color.git",
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
			Test: TestStruct{
				Command:   "go",
				Args:      []string{"test", "./..."},
				Directory: ".",
			},
		},
	}
	if !reflect.DeepEqual(c, d) {
		t.Error("Parsed and original values do not match")
		t.Error("get", string(golang))
	}
}
