package pkg

import (
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

func Test_notifier(t *testing.T) {
	token := os.Getenv("SLACK_TOKEN")
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

	if os.Getenv("INTEGRATION") != "true" {
		return
	}

	channelID := getChannel(token, "demo")
	if channelID != "CKG85VC9G" {
		t.Error("error token is wrong; it is not demo ID")
	}
	sendMessage(token, channelID, "Hi")
}
