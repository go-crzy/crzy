package pkg

import (
	"errors"
	"fmt"

	"github.com/slack-go/slack"
)

type notifierStruct struct {
	Slack slackStruct
}

type slackStruct struct {
	Token   string
	Channel string
}

type slackNotifier struct {
	messenger messenger
	channelID string
}

type messenger interface {
	GetConversations(params *slack.GetConversationsParameters) (channels []slack.Channel, nextCursor string, err error)
	PostMessage(channelID string, options ...slack.MsgOption) (string, string, error)
}

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

func newSlackNotifier(s slackStruct) *slackNotifier {
	output := &slackNotifier{
		messenger: &mockMessenger{},
	}
	evs := &envVars{}
	token, err := evs.replace(s.Token)
	if err != nil || token == "" {
		return output
	}
	channel, err := evs.replace(s.Channel)
	if err != nil || channel == "" {
		return output
	}
	api := slack.New(token)
	notifier := &slackNotifier{
		messenger: api,
	}
	channelID := getChannel(notifier, s.Channel)
	if channelID == "" {
		return output
	}
	notifier.channelID = channelID
	return notifier
}

func getChannel(n *slackNotifier, channel string) string {
	channels, _, err := n.messenger.GetConversations(
		&slack.GetConversationsParameters{
			Types: []string{"public_channel"},
		})
	if err != nil {
		fmt.Printf("%s\n", err)
		return ""
	}
	for _, c := range channels {
		if channel == c.Name {
			return c.ID
		}
	}
	return ""
}

func (n *slackNotifier) sendMessage(msg string) error {
	channelID, timestamp, err := n.messenger.PostMessage(
		n.channelID,
		slack.MsgOptionText(msg, false),
	)
	if err != nil {
		fmt.Printf("%s\n", err)
		return err
	}
	fmt.Printf("Message successfully sent to channel %s at %s", channelID, timestamp)
	return nil
}
