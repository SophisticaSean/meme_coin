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

func log(t *testing.T, expected string, actual string) {
	t.Log("We expected " + expected + ", but got: " + actual)
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
		log(t, strconv.Itoa(2000), strconv.Itoa(user.CurMoney))
		t.Error(output)
	}
	if user.WonMoney != 1000 {
		t.Log("Coin game didn't compute WonMoney Properly!")
		log(t, strconv.Itoa(user.WonMoney), strconv.Itoa(1000))
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
		log(t, strconv.Itoa(user.LostMoney), strconv.Itoa(1000))
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
		log(t, strconv.Itoa(1000), strconv.Itoa(user.CurMoney))
		t.Error(output)
	}
	if user.WonMoney != 99000 {
		t.Log("Number game didn't compute WonMoney Properly!")
		log(t, strconv.Itoa(user.WonMoney), strconv.Itoa(99000))
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
		log(t, strconv.Itoa(user.CurMoney), strconv.Itoa(0))
		t.Error(output)
	}
	if user.LostMoney != 1000 {
		t.Log("Number game didn't compute LostMoney Properly!")
		log(t, strconv.Itoa(user.LostMoney), strconv.Itoa(1000))
		t.Error(output)
	}
}

func TestHackWin(t *testing.T) {
	botSess := interaction.NewConsoleSession()
	message := interaction.NewMessageEvent()
	db := handlers.DbGet()
	id := strconv.Itoa(int(time.Now().UnixNano()))
	targetID := strconv.Itoa(int(time.Now().UnixNano()))
	author := discordgo.User{
		ID:       id,
		Username: "thief",
	}
	target := discordgo.User{
		ID:       targetID,
		Username: "target",
	}
	rand.Seed(37)
	message.Message.Author = &author

	user := handlers.UserGet(&author, db)
	userUnits := handlers.UnitsGet(&author, db)
	userUnits.Hacker = 100
	userUnits.Botnet = 100
	handlers.UpdateUnits(&userUnits, db)
	userUnits = handlers.UnitsGet(&target, db)

	targetUser := handlers.UserGet(&target, db)
	targetUnits := handlers.UnitsGet(&target, db)
	targetUnits.Miner = 14
	targetUnits.HackSeed = 37
	targetUnits.CollectTime = targetUnits.CollectTime.Add(-10 * time.Minute)
	handlers.UpdateUnits(&targetUnits, db)

	text := "!hack 100 67 @target"
	message.Message.Content = text
	message.Message.Mentions = append(message.Message.Mentions, &target)

	output := capStdout(botSess, message)
	newUser := handlers.UserGet(&author, db)
	newTargetUser := handlers.UserGet(&target, db)
	newTargetUnits := handlers.UnitsGet(&target, db)

	userMoneyDiff := newUser.CurMoney - user.CurMoney
	if userMoneyDiff != targetUnits.Miner {
		t.Log("The thief's money wasn't updated properly.")
		log(t, strconv.Itoa(targetUnits.Miner), strconv.Itoa(userMoneyDiff))
		t.Error(output)
	}

	userStoleDiff := newUser.HackedMoney - user.HackedMoney
	if userStoleDiff != targetUnits.Miner {
		t.Log("The thief's HackedMoney wasn't updated properly.")
		log(t, strconv.Itoa(targetUnits.Miner), strconv.Itoa(userStoleDiff))
		t.Log(newUser.HackedMoney)
		t.Error(output)
	}

	targetStoleDiff := newTargetUser.StolenFromMoney - targetUser.StolenFromMoney
	if targetStoleDiff != targetUnits.Miner {
		t.Log("The target's StolenFromMoney wasn't updated properly.")
		log(t, strconv.Itoa(targetUnits.Miner), strconv.Itoa(targetStoleDiff))
		t.Log(newUser.HackedMoney)
		t.Error(output)
	}

	if newTargetUnits.CollectTime == targetUnits.CollectTime {
		t.Log("The target's CollectTime wasn't updated properly.")
		t.Error(output)
	}

	if newTargetUnits.HackAttempts != 0 {
		t.Log("The target's HackAttempts was not reset back to 0")
		log(t, strconv.Itoa(0), strconv.Itoa(newTargetUnits.HackAttempts))
		t.Error(output)
	}

	if newTargetUnits.HackSeed != 0 {
		t.Log("The target's HackSeed was not reset")
		t.Log("The seed was still: " + strconv.Itoa(int(newTargetUnits.HackSeed)))
		t.Error(output)
	}

	// analyze the output
}
