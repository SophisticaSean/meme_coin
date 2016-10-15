package events

import (
	"io/ioutil"
	"math/rand"
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
	db := handlers.DbGet()
	user := handlers.UserGet(&author, db)
	if user.CurMoney != 1000 {
		t.Error("Did not give new user 1000 starting money.")
	}
}

func TestGambleCoinWin(t *testing.T) {
	botSess := interaction.NewConsoleSession()
	message := interaction.NewMessageEvent()
	id := strconv.Itoa(int(time.Now().UnixNano()))
	author := discordgo.User{
		ID:       id,
		Username: "admin",
	}
	rand.Seed(37)
	message.Message.Author = &author
	gambleAmount := 1000
	text := "!gamble " + strconv.Itoa(gambleAmount) + " coin heads"
	message.Message.Content = text
	db := handlers.DbGet()
	user := handlers.UserGet(&author, db)

	output := capStdout(botSess, message)
	if !strings.Contains(output, "result was heads.") {
		t.Log("Coin game did not report result.")
		t.Error(output)
	}
	if !strings.Contains(output, author.Username+" won "+strconv.Itoa(gambleAmount)+" memes.") {
		t.Log("Coin toss did not report win properly.")
		t.Error(output)
	}
	user = handlers.UserGet(&author, db)
	if user.CurMoney != 2000 {
		t.Log("Coin toss did not award proper amount of memes!")
		t.Log("We expected 2000 memes, but got " + strconv.Itoa(user.CurMoney))
		t.Error(output)
	}
	if user.WonMoney != 1000 {
		t.Log("Coin game didn't compute WonMoney Properly!")
		t.Log("User has " + strconv.Itoa(user.WonMoney) + " but we expected 1000.")
		t.Error(output)
	}
}

func TestGambleCoinLoss(t *testing.T) {
	botSess := interaction.NewConsoleSession()
	message := interaction.NewMessageEvent()
	id := strconv.Itoa(int(time.Now().UnixNano()))
	author := discordgo.User{
		ID:       id,
		Username: "admin",
	}
	rand.Seed(37)
	message.Message.Author = &author
	gambleAmount := 1000
	text := "!gamble " + strconv.Itoa(gambleAmount) + " coin tails"
	message.Message.Content = text
	db := handlers.DbGet()
	user := handlers.UserGet(&author, db)

	output := capStdout(botSess, message)
	if !strings.Contains(output, "result was heads.") {
		t.Log("Coin game did not report result.")
		t.Error(output)
	}
	if !strings.Contains(output, author.Username+" lost "+strconv.Itoa(gambleAmount)+" memes.") {
		t.Log("Coin toss did not report loss properly.")
		t.Error(output)
	}
	user = handlers.UserGet(&author, db)
	if user.CurMoney != 0 {
		t.Log("Coin toss did not take away memes!")
		t.Log("User still has " + strconv.Itoa(user.CurMoney) + " memes.")
		t.Error(output)
	}
	if user.LostMoney != 1000 {
		t.Log("Coin game didn't compute LostMoney Properly!")
		t.Log("User has " + strconv.Itoa(user.LostMoney) + " but we expected 1000.")
		t.Error(output)
	}
}

func TestGambleNumberWin(t *testing.T) {
	botSess := interaction.NewConsoleSession()
	message := interaction.NewMessageEvent()
	id := strconv.Itoa(int(time.Now().UnixNano()))
	author := discordgo.User{
		ID:       id,
		Username: "admin",
	}
	rand.Seed(37)
	message.Message.Author = &author
	gambleAmount := 1000
	text := "!gamble " + strconv.Itoa(gambleAmount) + " number 35:100"
	message.Message.Content = text
	db := handlers.DbGet()
	user := handlers.UserGet(&author, db)

	output := capStdout(botSess, message)
	if !strings.Contains(output, "result was 35.") {
		t.Log("Number game did not report result.")
		t.Error(output)
	}
	if !strings.Contains(output, author.Username+" won "+strconv.Itoa(gambleAmount*99)+" memes.") {
		t.Log("Number game did not report win properly.")
		t.Error(output)
	}
	user = handlers.UserGet(&author, db)
	if user.CurMoney != 100000 {
		t.Log("Number game did not award proper amount of memes!")
		t.Log("We expected 10000 memes, but got " + strconv.Itoa(user.CurMoney))
		t.Error(output)
	}
	if user.WonMoney != 99000 {
		t.Log("Number game didn't compute WonMoney Properly!")
		t.Log("User has " + strconv.Itoa(user.WonMoney) + " but we expected 99000.")
		t.Error(output)
	}
}

func TestGambleNumberLoss(t *testing.T) {
	botSess := interaction.NewConsoleSession()
	message := interaction.NewMessageEvent()
	id := strconv.Itoa(int(time.Now().UnixNano()))
	author := discordgo.User{
		ID:       id,
		Username: "admin",
	}
	rand.Seed(37)
	message.Message.Author = &author
	gambleAmount := 1000
	text := "!gamble " + strconv.Itoa(gambleAmount) + " number 36:100"
	message.Message.Content = text
	db := handlers.DbGet()
	user := handlers.UserGet(&author, db)

	output := capStdout(botSess, message)
	if !strings.Contains(output, "result was 35.") {
		t.Log("Number game did not report result.")
		t.Error(output)
	}
	if !strings.Contains(output, author.Username+" lost "+strconv.Itoa(gambleAmount)+" memes.") {
		t.Log("Number game did not report loss properly.")
		t.Error(output)
	}
	user = handlers.UserGet(&author, db)
	if user.CurMoney != 0 {
		t.Log("Number game did not take away memes!")
		t.Log("User still has " + strconv.Itoa(user.CurMoney) + " memes.")
		t.Error(output)
	}
	if user.LostMoney != 1000 {
		t.Log("Number game didn't compute LostMoney Properly!")
		t.Log("User has " + strconv.Itoa(user.LostMoney) + " but we expected 1000.")
		t.Error(output)
	}
}
