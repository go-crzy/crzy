package pkg

import (
	"errors"
	"testing"

	"github.com/slack-go/slack"
	"gopkg.in/yaml.v3"
)

type mockMessenger struct{}

func (m *mockMessenger) GetConversations(params *slack.GetConversationsParameters) ([]slack.Channel, string, error) {
	channels := []slack.Channel{}
	channels = append(channels, slack.Channel{GroupConversation: slack.GroupConversation{Conversation: slack.Conversation{ID: "123"}, Name: "ops"}})
	return channels, "", nil
}

func (m *mockMessenger) PostMessage(channelID string, options ...slack.MsgOption) (string, string, error) {
	if channelID == "wrong" {
		return "", "", errors.New("wrongChannel")
	}
	return "", "", nil
}

func Test_getChannel(t *testing.T) {
	g := &slackNotifier{messenger: &mockMessenger{}}
	channel := g.getChannel("ops")
	if channel != "123" {
		t.Error("Should succeed", channel)
	}
}

func Test_sendMessage(t *testing.T) {
	s := &slackNotifier{messenger: &mockMessenger{}}
	err := s.sendMessage("", "", "titi")
	if err != nil {
		t.Error("Should succeed", err)
	}
	err = s.sendMessage("", "wrong", "titi")
	if err.Error() != "wrongChannel" {
		t.Error("Should fail with wrongChannel, instead:", err)
	}
}

type notifierTest struct {
	Notifier notifierStruct
}

func Test_notifier(t *testing.T) {
	// token := os.Getenv("SLACK_TOKEN")
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
