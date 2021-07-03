package pkg

import (
	"encoding/json"
	"os"
	"reflect"
	"runtime"
	"testing"
)

func Test_defaultConf_and_succeed(t *testing.T) {
	d := &config{
		Main: mainStruct{
			Head:       "main",
			Color:      true,
			Repository: "myrepo",
			API: apiStruct{
				Port: 8080,
			},
			Proxy: proxyStruct{
				Port: 8081,
			},
		},
		Deploy: deployStruct{
			Artifact: artifactStruct{
				Filename: "go-${version}",
			},
			Build: execStruct{
				Command: "go",
				Args:    []string{"build", "-o", `${artifact}`, "."},
				WorkDir: ".",
			},
			Test: execStruct{
				Command: "go",
				Args:    []string{"test", "-v", "./..."},
				WorkDir: ".",
			},
		},
		Release: releaseStruct{
			Run: execStruct{
				Command: "./go-${version}",
				WorkDir: ".",
				Envs: []envVar{
					{Name: "ADDR", Value: "localhost:${port}"},
					{Name: "PORT", Value: ":${port}"},
				},
			},
			PortRange: portRangeStruct{
				Min: 8090,
				Max: 8100,
			},
		},
		Notifier: notifierStruct{
			Slack: slackStruct{
				Channel: "general",
				Token:   "${SLACK_TOKEN}",
			},
		},
	}
	if runtime.GOOS == "windows" {
		d.Deploy.Artifact.Extension = ".exe"
	}
	c, err := defaultConf("go")
	if err != nil {
		t.Error("expect defaultConf with go to succeed")
	}
	if !reflect.DeepEqual(c, d) {
		text, _ := json.Marshal(&c)
		t.Error("values do not match", string(text))
	}
}

func Test_defaultConf_and_fail(t *testing.T) {
	_, err := defaultConf("java")
	if err != errUnsupportedLang {
		t.Error("java should not be supported for now")
	}
}

func Test_getConfig_and_fail_java(t *testing.T) {
	_, err := getConfig("java", "")
	if err != errUnsupportedLang {
		t.Error("java should not be supported for now")
	}
}

func Test_getConfig_and_fail_golang(t *testing.T) {
	_, err := getConfig("go", "fail.yaml")
	if err != errLoadingConfigFile {
		t.Error("should fail reading file, instead:", err)
	}
}

func Test_getConf_and_fail_golang(t *testing.T) {
	c := &defaultContainer{}
	err := c.getConf(Args{
		ConfigFile: "fail.yaml",
	})
	if err != errLoadingConfigFile {
		t.Error("should fail reading file, instead:", err)
	}
}

func Test_getConfig_without_file_and_succeed(t *testing.T) {
	_, err := getConfig("go", DefaultConfigFile)
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
	_, err = getConfig("go", f.Name())
	if err != nil {
		t.Error("should be able to read file")
	}
}

func Test_getConf_with_file_and_succeed(t *testing.T) {
	args := Args{
		ConfigFile: DefaultConfigFile,
		NoColor:    false,
	}
	c := &defaultContainer{}
	err := c.getConf(args)
	if err != nil {
		t.Error("should be able to read file")
	}
}
