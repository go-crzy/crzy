package pkg

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func Test_ConfigUnmarshal(t *testing.T) {
	fileContent := `
main:
  head: main
  server: true
  color: true
  repository: color.git
  api_port: 8080
  proxy_port: 8081

# utilisé à la ligne 177 du fichier git.go
version:
  command: git
  args:
  - log
  - "-1"
  - "--format=%H"
  - "."
  directory: "."
  
deployment:
  artifact:
    type: executable
    pattern: go-${version}
  # utilisé à la ligne 196 du fichier git.go
  build:
    command: go
    args:
    - build
    - "-o"
    - ${artifact}
    - "."
    directory: "."
`
	c := config{}
	err := yaml.Unmarshal([]byte(fileContent), &c)
	if err != nil {
		t.Error(err, "error unmarshalling file")
	}
	if c.Main.Repository != "color.git" {
		t.Error("error repository should be color.git; it is", c.Main.Repository)
	}

	if c.Version.Args[0] != "log" {
		t.Error("error repository should be log; it is", c.Version.Args)
	}

	if c.Deployment.Artifact.Pattern != "go-${version}" {
		t.Error("error repository should be go-${version}; it is", c.Version.Args)
	}

}
