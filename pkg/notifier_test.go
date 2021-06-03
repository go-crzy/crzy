package pkg

import (
	"testing"

	"gopkg.in/yaml.v3"
)

// TODO:
// 1- cree un mock qui implémente l'interface messenger et renvoie
//   des valeurs par défaut pour les 2 fonctions de l'interfaces
// par exemple, le mock est une struct vide mockMessenger
//      - la fonction  (m *mockMessenger) GetConversations(params *slack.GetConversationsParameters) (channels []slack.Channel, nextCursor string, err error)
//        renvoie []slack.Channel avec 1 channel, "" et nil
//      - la fonction  (m *mockMessenger) PostMessage(channelID string, options ...slack.MsgOption) (string, string, error)
//        renvoie "", "", nil
// 2- Créer un test, qui :
//     - instancie le mock
//     - crée un objet slackNotifier dans lequel on met le mock
//     - teste la fonction getChannel pour vérifier que l'ensemble du code "en dehors" de slack fonctionne
//     - teste la fonction sendMessage pour vérifier que l'ensemble du code "en dehors" de slack fonctionne

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

	// if os.Getenv("INTEGRATION") != "true" {
	// 	return
	// }

	// channelID := getChannel(token, "demo")
	// if channelID != "CKG85VC9G" {
	// 	t.Error("error token is wrong; it is not demo ID")
	// }
	// sendMessage(token, channelID, "Hi")
}
