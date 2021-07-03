package pkg

import (
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

func Test_getChannel(t *testing.T) {
	g := &slackNotifier{messenger: &mockMessenger{}}
	channel := getChannel(g, "ops")
	if channel != "123" {
		t.Error("Should succeed", channel)
	}
}

func Test_sendMessage(t *testing.T) {
	s := &slackNotifier{messenger: &mockMessenger{}}
	err := s.sendMessage("titi")
	if err != nil {
		t.Error("Should succeed", err)
	}
	s.channelID = "wrong"
	err = s.sendMessage("titi")
	if err.Error() != "wrongChannel" {
		t.Error("Should fail with wrongChannel, instead:", err)
	}
}

type notifierTest struct {
	Notifier notifierStruct
}

func Test_notifier(t *testing.T) {
	fileContent := `
notifier:
  slack:
    token: xoxb-xxxx
    channel: demo
`
	c := notifierTest{}
	err := yaml.Unmarshal([]byte(fileContent), &c)
	if err != nil {
		t.Error("error unmarshalling file")
	}
	if c.Notifier.Slack.Channel != "demo" {
		t.Error("error channel should be demo")
	}
	if c.Notifier.Slack.Token != "xoxb-xxxx" {
		t.Error("error channel should be xoxb-xxxx")
	}
}

func Test_newSlackNotifier(t *testing.T) {
	fileContent := `
token: ${SLACK_TOKEN}
channel: ""
`
	c := slackStruct{}
	err := yaml.Unmarshal([]byte(fileContent), &c)
	if err != nil {
		t.Error("error unmarshalling file")
	}
	os.Setenv("SLACK_TOKEN", "xoxb-...")
	newSlackNotifier(c)
}

func Test_notifier_config(t *testing.T) {
	c, err := defaultConf("go")
	if err != nil {
		t.Error("expect defaultConf with golang to succeed")
	}
	if c.Notifier.Slack.Token != "${SLACK_TOKEN}" {
		t.Error("expecting ${SLACK_TOKEN}, got: ", c.Notifier.Slack.Channel)
	}
}
