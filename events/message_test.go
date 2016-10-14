package events

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/SophisticaSean/meme_coin/interaction"
	"github.com/bwmarrin/discordgo"
)

func TestHelp(t *testing.T) {
	botSess := interaction.NewConsoleSession()
	message := interaction.NewMessageEvent()
	author := discordgo.User{
		ID:       "2",
		Username: "admin",
	}
	message.Message.Author = &author
	text := "!memes"
	message.Message.Content = text
	fname := filepath.Join(os.TempDir(), "stdout")
	old := os.Stdout
	temp, _ := os.Create(fname)
	os.Stdout = temp

	MessageHandler(botSess, &message)

	os.Stdout = old
	out, _ := ioutil.ReadFile(fname)
	//t.Log(string(out))
	t.Error(string(out))
	//message := help()
	//if len(message) != 126 {
	//t.Error("help message was this long: " + strconv.Itoa(len(message)))
	//}
}
