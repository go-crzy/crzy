package pkg

import (
	"fmt"

	"github.com/slack-go/slack"
)

type notifierStruct struct {
	Slack slackStruct
}

type slackStruct struct {
	Token   string `default:"${SLACK_TOKEN}"`
	Channel string
}

type slackNotifier struct {
	messenger messenger
}

type messenger interface {
	GetConversations(params *slack.GetConversationsParameters) (channels []slack.Channel, nextCursor string, err error)
	PostMessage(channelID string, options ...slack.MsgOption) (string, string, error)
}

func newSlackNotifier(token string) slackNotifier {
	api := slack.New(token)
	return slackNotifier{
		messenger: api,
	}
}

func (n *slackNotifier) getChannel(channel string) string {
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

func (n *slackNotifier) sendMessage(token, channelID, msg string) error {
	channelID, timestamp, err := n.messenger.PostMessage(
		channelID,
		slack.MsgOptionText(msg, false),
	)
	if err != nil {
		fmt.Printf("%s\n", err)
		return err
	}
	fmt.Printf("Message successfully sent to channel %s at %s", channelID, timestamp)
	return nil
}
