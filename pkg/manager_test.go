package pkg

import (
	"fmt"
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

# Exemple de liste
maliste:
- value1
- value2

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
	for _, v := range c.Maliste {
		fmt.Println(v)
	}
	//  go test -v -run Test_ConfigUnmarshal
}
