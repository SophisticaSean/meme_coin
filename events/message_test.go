package events

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/SophisticaSean/meme_coin/handlers"
	"github.com/SophisticaSean/meme_coin/interaction"
	"github.com/bwmarrin/discordgo"
)

func capStdout(botSess interaction.Session, messageEvent interaction.MessageCreate) string {
	fname := filepath.Join(os.TempDir(), "stdout")
	old := os.Stdout
	temp, _ := os.Create(fname)
	os.Stdout = temp

	MessageHandler(botSess, &messageEvent)

	os.Stdout = old
	out, _ := ioutil.ReadFile(fname)
	return string(out)
}

func TestMain(m *testing.M) {

	exit := m.Run()
	handlers.DbReset()

	os.Exit(exit)
}

func TestHelp(t *testing.T) {
	targetString := "yo, whaddup. Here are the commands I know:\r`!military` `!hack` `!buy` `!mine` `!units` `!collect` `!gamble` `!tip` `!balance` `!memes` `!memehelp`"
	splitTargets := strings.Split(targetString, " ")
	botSess := interaction.NewConsoleSession()
	message := interaction.NewMessageEvent()
	author := discordgo.User{
		ID:       "2",
		Username: "admin",
	}
	message.Message.Author = &author
	text := "!help"
	message.Message.Content = text

	output := capStdout(botSess, message)

	outputSplit := strings.Split(output, " ")
	for i := range outputSplit {
		if outputSplit[i] != splitTargets[i] {
			t.Log("Help string was not what we expected!")
			t.Error(outputSplit[i], splitTargets[i])
		}
	}
}

func TestNewUser(t *testing.T) {
	botSess := interaction.NewConsoleSession()
	message := interaction.NewMessageEvent()
	id := strconv.Itoa(int(time.Now().UnixNano()))
	author := discordgo.User{
		ID:       id,
		Username: "admin",
	}
	message.Message.Author = &author
	text := "!memes"
	message.Message.Content = text

	output := capStdout(botSess, message)
	if !strings.Contains(output, "total balance is: 1000") {
		t.Error(output)
	}
	if !strings.Contains(output, "creating user: "+id) {
		t.Log("user table creation didn't fire")
		t.Error(output)
	}
	if !strings.Contains(output, "creating user in units table: "+id) {
		t.Log("user units creation didn't fire")
		t.Error(output)
	}
}
