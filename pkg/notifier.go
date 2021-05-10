package pkg

import (
	"fmt"

	"github.com/slack-go/slack"
)

type NotifierStruct struct {
	Slack SlackStruct
}

type SlackStruct struct{
	Token string
	Channel string
}

func getChannel(token, channel string) string {
	api := slack.New(token)
	channels, _, err := api.GetConversations(&slack.GetConversationsParameters{Types: []string{"public_channel"}})
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

func sendMessage(token, channelID, msg string) {
	api := slack.New(token)
	channelID, timestamp, err := api.PostMessage(
		channelID,
		slack.MsgOptionText(msg, false),
	)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	fmt.Printf("Message successfully sent to channel %s at %s", channelID, timestamp)
}
