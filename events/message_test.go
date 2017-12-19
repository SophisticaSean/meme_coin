package events

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/SophisticaSean/meme_coin/handlers"
	"github.com/SophisticaSean/meme_coin/interaction"
	"github.com/bwmarrin/discordgo"
	"github.com/davecgh/go-spew/spew"
	humanize "github.com/dustin/go-humanize"
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
	//handlers.DbReset()
	exit := m.Run()
	handlers.DbReset()

	os.Exit(exit)
}

func numLog(t *testing.T, expected int, actual int) {
	t.Log("We expected " + strconv.Itoa(expected) + ", but got: " + strconv.Itoa(actual))
}

func TestHelp(t *testing.T) {
	targetString := "yo, whaddup. Here are the commands I know:\r`!military` `!hack` `!buy` `!mine` `!units` `!collect` `!gamble` `!tip` `!balance` `!memes` `!memehelp` `!prestige` `!fakecollect` `!check` `!invite`"
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

	if !strings.Contains(output, targetString) {
		t.Log("help output didn't report what we expected.")
		t.Error(output)
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
	if !strings.Contains(output, "total balance is: 1,000") {
		t.Error(output)
	}
	db := handlers.DbGet()
	user := handlers.UserGet(&author, db)
	if user.CurMoney != 1000 {
		t.Error("Did not give new user 1000 starting money.")
	}
}

func TestGambleNegativeAmount(t *testing.T) {
	botSess := interaction.NewConsoleSession()
	message := interaction.NewMessageEvent()
	id := strconv.Itoa(int(time.Now().UnixNano()))
	author := discordgo.User{
		ID:       id,
		Username: "admin",
	}
	rand.Seed(37)
	message.Message.Author = &author
	gambleAmount := -1000
	text := "!gamble " + strconv.Itoa(gambleAmount) + " coin heads"
	message.Message.Content = text
	db := handlers.DbGet()
	user := handlers.UserGet(&author, db)

	output := capStdout(botSess, message)
	if !strings.Contains(output, "amount has to be more than 0") {
		t.Log("Coin game did not error on bad input.")
		t.Error(output)
	}
	user = handlers.UserGet(&author, db)
	if user.CurMoney != 1000 {
		t.Log("Coin toss did not award proper amount of memes!")
		numLog(t, 1000, user.CurMoney)
		t.Error(output)
	}
	if user.WonMoney != 0 {
		t.Log("Coin game didn't compute WonMoney Properly!")
		numLog(t, user.WonMoney, 0)
		t.Error(output)
	}
}

func TestGambleOverflow(t *testing.T) {
	botSess := interaction.NewConsoleSession()
	message := interaction.NewMessageEvent()
	id := strconv.Itoa(int(time.Now().UnixNano()))
	author := discordgo.User{
		ID:       id,
		Username: "admin",
	}
	rand.Seed(37)
	message.Message.Author = &author
	text := "!gamble 120301020300120301028012310929301923091093 coin heads"
	message.Message.Content = text
	db := handlers.DbGet()
	user := handlers.UserGet(&author, db)

	output := capStdout(botSess, message)
	if !strings.Contains(output, "amount is too large or not a number, try again.") {
		t.Log("Coin game did not error on bad input.")
		t.Error(output)
	}
	user = handlers.UserGet(&author, db)
	if user.CurMoney != 1000 {
		t.Log("Coin toss did not award proper amount of memes!")
		numLog(t, 1000, user.CurMoney)
		t.Error(output)
	}
	if user.WonMoney != 0 {
		t.Log("Coin game didn't compute WonMoney Properly!")
		numLog(t, user.WonMoney, 0)
		t.Error(output)
	}
}

func TestGambleNaN(t *testing.T) {
	botSess := interaction.NewConsoleSession()
	message := interaction.NewMessageEvent()
	id := strconv.Itoa(int(time.Now().UnixNano()))
	author := discordgo.User{
		ID:       id,
		Username: "admin",
	}
	rand.Seed(37)
	message.Message.Author = &author
	text := "!gamble hello coin heads"
	message.Message.Content = text
	db := handlers.DbGet()
	user := handlers.UserGet(&author, db)

	output := capStdout(botSess, message)
	if !strings.Contains(output, "amount is too large or not a number, try again.") {
		t.Log("Coin game did not error on bad input.")
		t.Error(output)
	}
	user = handlers.UserGet(&author, db)
	if user.CurMoney != 1000 {
		t.Log("Coin toss did not award proper amount of memes!")
		numLog(t, 1000, user.CurMoney)
		t.Error(output)
	}
	if user.WonMoney != 0 {
		t.Log("Coin game didn't compute WonMoney Properly!")
		numLog(t, user.WonMoney, 0)
		t.Error(output)
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
	result := "tails"
	text := "!gamble " + strconv.Itoa(gambleAmount) + " coin " + result
	message.Message.Content = text
	db := handlers.DbGet()
	user := handlers.UserGet(&author, db)

	output := capStdout(botSess, message)
	if !strings.Contains(output, "result was "+result+".") {
		t.Log("Coin game did not report result.")
		t.Error(output)
	}
	if !strings.Contains(output, author.Username+" won "+humanize.Comma(int64(gambleAmount))+" memes.") {
		t.Log("Coin toss did not report win properly.")
		t.Error(output)
	}
	user = handlers.UserGet(&author, db)
	if user.CurMoney != 2000 {
		t.Log("Coin toss did not award proper amount of memes!")
		numLog(t, 2000, user.CurMoney)
		t.Error(output)
	}
	if user.WonMoney != 1000 {
		t.Log("Coin game didn't compute WonMoney Properly!")
		numLog(t, user.WonMoney, 1000)
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
	rand.Seed(38)
	message.Message.Author = &author
	gambleAmount := 1000
	result := "heads"
	text := "!gamble " + strconv.Itoa(gambleAmount) + " coin " + result
	message.Message.Content = text
	db := handlers.DbGet()
	user := handlers.UserGet(&author, db)

	output := capStdout(botSess, message)
	if !strings.Contains(output, "result was tails.") {
		t.Log("Coin game did not report result.")
		t.Error(output)
	}
	if !strings.Contains(output, author.Username+" lost "+humanize.Comma(int64(gambleAmount))+" memes.") {
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
		numLog(t, user.LostMoney, 1000)
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
	if !strings.Contains(output, author.Username+" won "+humanize.Comma(int64(gambleAmount*99))+" memes.") {
		t.Log("Number game did not report win properly.")
		t.Error(output)
	}
	user = handlers.UserGet(&author, db)
	if user.CurMoney != 100000 {
		t.Log("Number game did not award proper amount of memes!")
		numLog(t, 1000, user.CurMoney)
		t.Error(output)
	}
	if user.WonMoney != 99000 {
		t.Log("Number game didn't compute WonMoney Properly!")
		numLog(t, user.WonMoney, 99000)
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
	if !strings.Contains(output, author.Username+" lost "+humanize.Comma(int64((gambleAmount)))+" memes.") {
		t.Log("Number game did not report loss properly.")
		t.Error(output)
	}
	user = handlers.UserGet(&author, db)
	if user.CurMoney != 0 {
		t.Log("Number game did not take away memes!")
		numLog(t, user.CurMoney, 0)
		t.Error(output)
	}
	if user.LostMoney != 1000 {
		t.Log("Number game didn't compute LostMoney Properly!")
		numLog(t, user.LostMoney, 1000)
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
	seed := int64(37)
	hackers := "38"
	botnets := "29"

	rand.Seed(seed)

	user := handlers.UserGet(&author, db)
	user.Hacker = 100
	user.Botnet = 100
	handlers.UpdateUnits(&user, db)
	//t.Fatal(user.Botnet)
	//user = handlers.UserGet(&author, db)

	targetUser := handlers.UserGet(&target, db)
	targetUser.Miner = 14
	targetUser.HackSeed = seed
	targetUser.CollectTime = targetUser.CollectTime.Add(-10 * time.Minute)
	handlers.UpdateUnits(&targetUser, db)

	text := "!hack " + hackers + " " + botnets + " @target"
	message.Message.Content = text
	message.Message.Mentions = append(message.Message.Mentions, &target)
	message.Message.Author = &author

	output := capStdout(botSess, message)
	newUser := handlers.UserGet(&author, db)
	newTarget := handlers.UserGet(&target, db)

	// verify the thief's stats
	userMoneyDiff := newUser.CurMoney - user.CurMoney
	if userMoneyDiff != targetUser.Miner {
		t.Log("The thief's money wasn't updated properly.")
		numLog(t, targetUser.Miner, userMoneyDiff)
		spew.Dump(output)
		t.Error(output)
	}

	userStoleDiff := newUser.HackedMoney - user.HackedMoney
	if userStoleDiff != targetUser.Miner {
		t.Log("The thief's HackedMoney wasn't updated properly.")
		numLog(t, targetUser.Miner, userStoleDiff)
		t.Log(newUser.HackedMoney)
		t.Error(output)
	}

	if newUser.Hacker != user.Hacker {
		t.Log("The thief lost hackers on a successful hack!")
		numLog(t, newUser.Hacker, user.Hacker)
		t.Error(output)
	}

	if newUser.Botnet != user.Botnet {
		t.Log("The thief lost botnets on a successful hack!")
		numLog(t, newUser.Botnet, user.Botnet)
		t.Error(output)
	}

	// verify the target's stats
	targetStoleDiff := newTarget.StolenFromMoney - targetUser.StolenFromMoney
	if targetStoleDiff != targetUser.Miner {
		t.Log("The target's StolenFromMoney wasn't updated properly.")
		numLog(t, targetUser.Miner, targetStoleDiff)
		t.Log(newUser.HackedMoney)
		t.Error(output)
	}

	if newTarget.CollectTime == targetUser.CollectTime {
		t.Log("The target's CollectTime wasn't updated properly.")
		t.Error(output)
	}

	if newTarget.HackAttempts != 0 {
		t.Log("The target's HackAttempts was not reset back to 0")
		numLog(t, 0, newTarget.HackAttempts)
		t.Error(output)
	}

	if newTarget.HackSeed != 0 {
		t.Log("The target's HackSeed was not reset")
		t.Log("The seed was still: " + strconv.Itoa(int(newTarget.HackSeed)))
		t.Error(output)
	}

	// verify the output
	// OLD TEST
	//if !strings.Contains(output, "totalIterations: "+botnets) {
	//t.Log("Number of iterations was incorrect.")
	//t.Error(output)
	//}
	//if !strings.Contains(output, "iterationLimit: "+botnets) {
	//t.Log("IterationLimit was incorrect.")
	//t.Error(output)
	//}
	//if !strings.Contains(output, "seed: "+strconv.Itoa(int(seed))) {
	//t.Log("The seed was incorrect.")
	//t.Error(output)
	//}
	//if !strings.Contains(output, "totalFitness: 32") {
	//t.Log("totalFitness was not what we expected.")
	//t.Error(output)
	//}
	//if !strings.Contains(output, "targetLength: 32") {
	//t.Log("targetLength of password was incorrect.")
	//t.Error(output)
	//}
	//if !strings.Contains(output, "populationSize: 9") {
	//t.Log("populationSize was not hardcapped to what we expected!")
	//t.Error(output)
	//}

	expectedOutput := ("The hack was successful, " + user.Username + " stole " + humanize.Comma(int64(targetUser.Miner)) + " dank memes from " + target.Username)
	if !strings.Contains(output, expectedOutput) {
		t.Log("Successful hacking output to channel was not what was expected.")
		t.Error(output)
	}
}

func TestHackLoss(t *testing.T) {
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
	seed := int64(293920) // botnet loss
	hackers := "3000"
	botnets := "2000"

	user := handlers.UserGet(&author, db)
	user.Hacker = 3000
	user.Botnet = 2000
	handlers.UpdateUnits(&user, db)
	user = handlers.UserGet(&author, db)

	targetUser := handlers.UserGet(&target, db)
	targetUser.Miner = 1400000000000
	targetUser.Cypher = 307
	targetUser.HackSeed = seed
	targetUser.CollectTime = targetUser.CollectTime.Add(-10 * time.Minute)
	handlers.UpdateUnits(&targetUser, db)

	rand.Seed(seed)

	text := "!hack " + hackers + " " + botnets + " @target"
	message.Message.Content = text
	message.Message.Mentions = append(message.Message.Mentions, &target)
	message.Message.Author = &author

	output := capStdout(botSess, message)
	newUser := handlers.UserGet(&author, db)
	newTarget := handlers.UserGet(&target, db)

	// verify the thief's stats
	userMoneyDiff := newUser.CurMoney - user.CurMoney
	if userMoneyDiff != 0 {
		t.Log("The thief's money wasn't updated properly.")
		numLog(t, targetUser.Miner, userMoneyDiff)
		t.Error(output)
	}

	userStoleDiff := newUser.HackedMoney - user.HackedMoney
	if userStoleDiff != 0 {
		t.Log("The thief's HackedMoney wasn't updated properly.")
		numLog(t, targetUser.Miner, userStoleDiff)
		t.Log(newUser.HackedMoney)
		t.Error(output)
	}

	if newUser.Hacker == user.Hacker {
		t.Log("The thief did not lose hackers on a failed hack!")
		numLog(t, newUser.Hacker, user.Hacker)
		t.Error(output)
	}

	if newUser.Botnet == user.Botnet {
		t.Log("The thief did not lose botnets on a failed hack!")
		numLog(t, newUser.Botnet, user.Botnet)
		t.Error(output)
	}

	// verify the target's stats
	targetStoleDiff := newTarget.StolenFromMoney - targetUser.StolenFromMoney
	if targetStoleDiff != 0 {
		t.Log("The target's StolenFromMoney was updated on a failed hack!")
		numLog(t, targetUser.Miner, targetStoleDiff)
		t.Log(newUser.HackedMoney)
		t.Error(output)
	}

	if newTarget.CollectTime != targetUser.CollectTime {
		t.Log("The target's CollectTime was reset on a failed hack!")
		t.Error(output)
	}

	if newTarget.HackAttempts != targetUser.HackAttempts+1 {
		t.Log("The target's HackAttempts was not incremented.")
		numLog(t, 0, newTarget.HackAttempts)
		t.Error(output)
	}

	if newTarget.HackSeed != targetUser.HackSeed {
		t.Log("The target's HackSeed reset on a failed Hack!")
		t.Error(output)
	}

	expectedOutput := ("1,960 botnets left")
	if !strings.Contains(output, expectedOutput) {
		t.Log("Failed hacking output to channel was not what was expected.")
		t.Error(output)
	}
	expectedOutput = ("2,970 hackers left")
	if !strings.Contains(output, expectedOutput) {
		t.Log("Failed hacking output to channel was not what was expected.")
		t.Error(output)
	}
}

func TestHackInsufficientUnits(t *testing.T) {
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
	seed := int64(1281) // botnet loss
	hackers := "300"
	botnets := "300"

	user := handlers.UserGet(&author, db)
	user.Hacker = 100
	user.Botnet = 100
	handlers.UpdateUnits(&user, db)
	user = handlers.UserGet(&author, db)

	targetUser := handlers.UserGet(&target, db)
	targetUser.Miner = 14
	targetUser.Cypher = 307
	targetUser.HackSeed = seed
	targetUser.CollectTime = targetUser.CollectTime.Add(-10 * time.Minute)
	handlers.UpdateUnits(&targetUser, db)

	rand.Seed(seed)

	text := "!hack " + hackers + " " + botnets + " @target"
	message.Message.Content = text
	message.Message.Mentions = append(message.Message.Mentions, &target)
	message.Message.Author = &author

	output := capStdout(botSess, message)
	newUser := handlers.UserGet(&author, db)
	newTarget := handlers.UserGet(&target, db)

	// compare old and new users, make sure nothing has changed.
	if !reflect.DeepEqual(user, newUser) {
		t.Log("user did not equal newUser")
		t.Error(output)
	}
	if !reflect.DeepEqual(user, newUser) {
		t.Log("user did not equal newUser")
		t.Error(output)
	}
	if !reflect.DeepEqual(targetUser, newTarget) {
		t.Log("targetUser did not equal newTarget")
		t.Error(output)
	}
	if !reflect.DeepEqual(targetUser, newTarget) {
		t.Log("targetUser did not equal newTarget")
		t.Error(output)
	}

	// make sure output is correct
	expectedOutput := ("You don't have enough botnets for the requested hack need: " + botnets + " have: " + humanize.Comma(int64(user.Botnet)))
	if !strings.Contains(output, expectedOutput) {
		t.Log("Hacking output did not report botnet mismatch.")
		t.Error(output)
	}
	expectedOutput = ("You don't have enough hackers for the requested hack need: " + hackers + " have: " + humanize.Comma(int64(user.Hacker)))
	if !strings.Contains(output, expectedOutput) {
		t.Log("Hacking output did not report hacker mismatch.")
		t.Error(output)
	}
}

func TestCollectTenMinutes(t *testing.T) {
	botSess := interaction.NewConsoleSession()
	message := interaction.NewMessageEvent()
	db := handlers.DbGet()
	id := strconv.Itoa(int(time.Now().UnixNano()))
	author := discordgo.User{
		ID:       id,
		Username: "admin",
	}

	user := handlers.UserGet(&author, db)
	user.Miner = 1000
	user.HackAttempts = 2
	user.HackSeed = int64(12301002)
	user.CollectTime = user.CollectTime.Add(-10 * time.Minute)
	handlers.UpdateUnits(&user, db)
	user = handlers.UserGet(&author, db)

	text := "!collect"
	message.Message.Content = text
	message.Message.Author = &author

	output := capStdout(botSess, message)
	newUser := handlers.UserGet(&author, db)

	if newUser.CurMoney != (1000 + 1000) {
		t.Log("User CurMoney was not updated with collected amount!")
		t.Error(output)
	}

	if newUser.CollectTime == user.CollectTime {
		t.Log("User CollectTime was not reset upon collection.")
		t.Error(output)
	}

	expectedOutput := ("admin collected 1,000 memes!")
	if !strings.Contains(output, expectedOutput) {
		t.Log("Collecting output did not report proper meme collection amount!")
		t.Error(output)
	}
	if newUser.HackAttempts != 0 {
		t.Log("HackAttempts were not reset to 0 after collection.")
		t.Error(output)
	}
	if newUser.HackSeed != 0 {
		t.Log("HackSeed was not reset to 0 after collection.")
		t.Error(output)
	}
}

func TestCollectTwoHours(t *testing.T) {
	botSess := interaction.NewConsoleSession()
	message := interaction.NewMessageEvent()
	db := handlers.DbGet()
	id := strconv.Itoa(int(time.Now().UnixNano()))
	author := discordgo.User{
		ID:       id,
		Username: "admin",
	}

	user := handlers.UserGet(&author, db)
	user.Miner = 1000
	user.CollectTime = user.CollectTime.Add(-120 * time.Minute)
	handlers.UpdateUnits(&user, db)
	user = handlers.UserGet(&author, db)

	text := "!collect"
	message.Message.Content = text
	message.Message.Author = &author

	output := capStdout(botSess, message)
	newUser := handlers.UserGet(&author, db)

	if newUser.CurMoney != (1000 + 12545) {
		t.Log("User CurMoney was not updated with collected amount!")
		t.Error(output)
	}

	if newUser.CollectTime == user.CollectTime {
		t.Log("User CollectTime was not reset upon collection.")
		t.Error(output)
	}

	expectedOutput := ("admin collected 12,545 memes!")
	if !strings.Contains(output, expectedOutput) {
		t.Log("Collecting output did not report proper meme collection amount!")
		t.Error(output)
	}
}

func TestMineNoUnits(t *testing.T) {
	botSess := interaction.NewConsoleSession()
	message := interaction.NewMessageEvent()
	db := handlers.DbGet()
	id := strconv.Itoa(int(time.Now().UnixNano()))
	author := discordgo.User{
		ID:       id,
		Username: "admin",
	}
	rand.Seed(37)

	user := handlers.UserGet(&author, db)

	text := "!mine"
	message.Message.Content = text
	message.Message.Author = &author

	output := capStdout(botSess, message)
	newUser := handlers.UserGet(&author, db)

	if newUser.CurMoney != (1000 + 100) {
		t.Log("User CurMoney was not updated with mined amount!")
		t.Error(output)
	}

	if newUser.MineTime == user.MineTime {
		t.Log("User MineTime was not reset upon mining.")
		t.Error(output)
	}

	expectedOutput := ("admin mined 100")
	if !strings.Contains(output, expectedOutput) {
		t.Log("Mineing output did not report proper meme mine amount!")
		t.Error(output)
	}
}

func TestMineWithUnits(t *testing.T) {
	botSess := interaction.NewConsoleSession()
	message := interaction.NewMessageEvent()
	db := handlers.DbGet()
	id := strconv.Itoa(int(time.Now().UnixNano()))
	author := discordgo.User{
		ID:       id,
		Username: "admin",
	}
	rand.Seed(37)

	user := handlers.UserGet(&author, db)
	user.Miner = 1000
	//user.CollectTime = user.CollectTime.Add(-120 * time.Minute)
	handlers.UpdateUnits(&user, db)
	user = handlers.UserGet(&author, db)

	text := "!mine"
	message.Message.Content = text
	message.Message.Author = &author

	output := capStdout(botSess, message)
	newUser := handlers.UserGet(&author, db)

	if newUser.CurMoney != (1000 + 600) {
		t.Log("User CurMoney was not updated with mined amount!")
		t.Error(output)
	}

	if newUser.MineTime == user.MineTime {
		t.Log("User MineTime was not reset upon mineing.")
		t.Error(output)
	}

	expectedOutput := ("admin mined 600")
	if !strings.Contains(output, expectedOutput) {
		t.Log("Mineing output did not report proper meme mine amount!")
		t.Error(output)
	}
}

func TestMineFrequency(t *testing.T) {
	botSess := interaction.NewConsoleSession()
	message := interaction.NewMessageEvent()
	db := handlers.DbGet()
	id := strconv.Itoa(int(time.Now().UnixNano()))
	author := discordgo.User{
		ID:       id,
		Username: "admin",
	}
	rand.Seed(37)

	user := handlers.UserGet(&author, db)

	text := "!mine"
	message.Message.Content = text
	message.Message.Author = &author

	capStdout(botSess, message)
	newUser := handlers.UserGet(&author, db)
	output := capStdout(botSess, message)

	if newUser.CurMoney != (1000 + 100) {
		t.Log(newUser.CurMoney)
		t.Log("User CurMoney was not updated with mined amount!")
		t.Error(output)
	}

	if newUser.MineTime == user.MineTime {
		t.Log("User MineTime was not reset upon mineing.")
		t.Error(output)
	}

	expectedOutput := ("admin mined")
	if strings.Contains(output, expectedOutput) {
		t.Log("Mining reported success before time limit!")
		t.Error(output)
	}
}

func TestBuyAutoCollect(t *testing.T) {
	botSess := interaction.NewConsoleSession()
	message := interaction.NewMessageEvent()
	db := handlers.DbGet()
	id := strconv.Itoa(int(time.Now().UnixNano()))
	author := discordgo.User{
		ID:       id,
		Username: "admin",
	}

	user := handlers.UserGet(&author, db)
	user.Robot = 10
	user.CollectTime = user.CollectTime.Add(-10 * time.Minute)
	handlers.UpdateUnits(&user, db)
	user = handlers.UserGet(&author, db)

	text := "!buy 1 miner"
	message.Message.Content = text
	message.Message.Author = &author

	output := capStdout(botSess, message)
	newUser := handlers.UserGet(&author, db)

	if reflect.DeepEqual(user, newUser) {
		t.Log(user)
		t.Log(newUser)
		t.Log("User was not updated.")
		t.Error(output)
	}

	expectedOutput := ("admin successfully bought 1 miner")
	if !strings.Contains(output, expectedOutput) {
		t.Log("Buying did not report success.")
		t.Error(output)
	}

	if newUser.Miner != 1 {
		t.Log("Users miner count was not updated on successful purchase.")
		t.Error(output)
	}

	if newUser.CurMoney != 600 {
		t.Log("Users current money was not updated on successful purchase.")
		spew.Dump(newUser)
		t.Error(output)
	}
}

func TestBuyOverflow(t *testing.T) {
	botSess := interaction.NewConsoleSession()
	message := interaction.NewMessageEvent()
	db := handlers.DbGet()
	id := strconv.Itoa(int(time.Now().UnixNano()))
	author := discordgo.User{
		ID:       id,
		Username: "admin",
	}

	user := handlers.UserGet(&author, db)

	text := "!buy 99999999999 fracker"
	message.Message.Content = text
	message.Message.Author = &author

	output := capStdout(botSess, message)
	newUser := handlers.UserGet(&author, db)

	if !reflect.DeepEqual(user, newUser) {
		t.Log(user)
		t.Log(newUser)
		t.Log("UserUnits was updated even though the transaction shouldn't have gone through.")
		t.Error(output)
	}

	expectedOutput := ("admin successfully bought")
	if strings.Contains(output, expectedOutput) {
		t.Log("Buying reported success when it should have failed.")
		t.Error(output)
	}

	expectedOutput = ("You're trying to buy too many units at once")
	if !strings.Contains(output, expectedOutput) {
		t.Log("Buying should have returned a message regarding unit purchase amount.")
		t.Error(output)
	}
}

func TestBuy(t *testing.T) {
	botSess := interaction.NewConsoleSession()
	message := interaction.NewMessageEvent()
	db := handlers.DbGet()
	id := strconv.Itoa(int(time.Now().UnixNano()))
	author := discordgo.User{
		ID:       id,
		Username: "admin",
	}

	user := handlers.UserGet(&author, db)
	handlers.MoneyAdd(&user, 10000000000, "tip", db)
	handlers.UpdateUnits(&user, db)
	user = handlers.UserGet(&author, db)
	units := handlers.UnitList()

	text := "!buy 1 fracker"
	message.Message.Content = text
	message.Message.Author = &author

	output := capStdout(botSess, message)
	newUser := handlers.UserGet(&author, db)

	if reflect.DeepEqual(user, newUser) {
		t.Log(user)
		t.Log(newUser)
		t.Log("User was not updated.")
		t.Error(output)
	}

	expectedOutput := ("admin successfully bought 1 fracker")
	if !strings.Contains(output, expectedOutput) {
		t.Log("Buying did not report success.")
		t.Error(output)
	}

	if newUser.Fracker != 1 {
		t.Log("Users fracker count was not updated on successful purchase.")
		t.Error(output)
	}

	if newUser.CurMoney != (user.CurMoney - units[3].Cost) {
		t.Log("Users current money was not updated on successful purchase.")
		spew.Dump(units)
		t.Error(output)
	}

	text = "!buy 1 robot"
	message.Message.Content = text
	message.Message.Author = &author
	user = handlers.UserGet(&author, db)

	output = capStdout(botSess, message)
	newUser = handlers.UserGet(&author, db)

	if reflect.DeepEqual(user, newUser) {
		t.Log(user)
		t.Log(newUser)
		t.Log("User was not updated.")
		t.Error(output)
	}

	expectedOutput = ("admin successfully bought 1 robot")
	if !strings.Contains(output, expectedOutput) {
		t.Log("Buying did not report success.")
		t.Error(output)
	}

	if newUser.Robot != 1 {
		t.Log("Users robot count was not updated on successful purchase.")
		t.Error(output)
	}

	if newUser.CurMoney != (user.CurMoney - units[1].Cost) {
		t.Log("Users current money was not updated on successful purchase.")
		spew.Dump(units)
		t.Error(output)
	}

	text = "!buy 1 swarm"
	message.Message.Content = text
	message.Message.Author = &author
	user = handlers.UserGet(&author, db)

	output = capStdout(botSess, message)
	newUser = handlers.UserGet(&author, db)

	if reflect.DeepEqual(user, newUser) {
		t.Log(user)
		t.Log(newUser)
		t.Log("User was not updated.")
		t.Error(output)
	}

	expectedOutput = ("admin successfully bought 1 swarm")
	if !strings.Contains(output, expectedOutput) {
		t.Log("Buying did not report success.")
		t.Error(output)
	}

	if newUser.Swarm != 1 {
		t.Log("Users robot count was not updated on successful purchase.")
		t.Error(output)
	}
	if newUser.CurMoney != (user.CurMoney - units[2].Cost) {
		t.Log("Users current money was not updated on successful purchase.")
		spew.Dump(units)
		t.Error(output)
	}

	text = "!buy 1 miner"
	message.Message.Content = text
	message.Message.Author = &author
	user = handlers.UserGet(&author, db)

	output = capStdout(botSess, message)
	newUser = handlers.UserGet(&author, db)

	if reflect.DeepEqual(user, newUser) {
		t.Log(user)
		t.Log(newUser)
		t.Log("User was not updated.")
		t.Error(output)
	}

	expectedOutput = ("admin successfully bought 1 miner")
	if !strings.Contains(output, expectedOutput) {
		t.Log("Buying did not report success.")
		t.Error(output)
	}

	if newUser.Miner != 1 {
		t.Log("Users miner count was not updated on successful purchase.")
		t.Error(output)
	}

	if newUser.CurMoney != (user.CurMoney - units[0].Cost) {
		t.Log("Users current money was not updated on successful purchase.")
		spew.Dump(units)
		t.Error(output)
	}
}

/*

Fails for some reason.
FIXME debug this later.

func TestCollectOverflow(t *testing.T) {
	botSess := interaction.NewConsoleSession()
	message := interaction.NewMessageEvent()
	db := handlers.DbGet()
	id := strconv.Itoa(int(time.Now().UnixNano()))
	author := discordgo.User{
		ID:       id,
		Username: "admin",
	}

	user := handlers.UserGet(&author, db)
	user.Fracker = 90000000000000
	user.CollectTime = user.CollectTime.Add(-24 * time.Hour)
	handlers.UpdateUnits(&user, db)
	user = handlers.UserGet(&author, db)

	text := "!collect"
	message.Message.Content = text
	message.Message.Author = &author

	output := capStdout(botSess, message)

	newUser := handlers.UserGet(&author, db)

	if newUser.CurMoney != 1000 {
		t.Log("User CurMoney was updated with an invalid amount!")
		t.Error(output)
	}

	if newUser.CollectTime != user.CollectTime {
		t.Log("User CollectTime was reset on overflowed collection.")
		t.Error(output)
	}

	expectedOutput := ("looks like you're trying to collect too many memes!")
	if !strings.Contains(output, expectedOutput) {
		t.Log("Collecting output did not report proper meme collection error!")
		t.Error(output)
	}
}
*/

func TestPrestigeSuccess(t *testing.T) {
	botSess := interaction.NewConsoleSession()
	message := interaction.NewMessageEvent()
	db := handlers.DbGet()
	id := strconv.Itoa(int(time.Now().UnixNano()))
	author := discordgo.User{
		ID:       id,
		Username: "admin",
	}

	user := handlers.UserGet(&author, db)
	handlers.MoneyAdd(&user, 10000, "tip", db)
	user.Miner = 100
	user.Robot = 100
	user.Swarm = 100
	user.Fracker = 100
	handlers.UpdateUnits(&user, db)

	user = handlers.UserGet(&author, db)

	text := "!prestige YESIMSURE"
	message.Message.Content = text
	message.Message.Author = &author

	output := capStdout(botSess, message)

	user = handlers.UserGet(&author, db)
	if user.CurMoney != 1000 {
		t.Log("User CurMoney was not changed, it should have reset to 1000.")
		t.Error(output)
	}

	if user.Miner != 0 {
		t.Log("User Miner count was not reset to 0.")
		t.Error(user.Miner)
	}

	if user.PrestigeLevel != 1 {
		t.Log("User prestige level was not set to 1.")
		t.Error(user.Miner)
	}

	expectedOutput := "You have been reset! Congratulations, you are now prestige level 1, which means you get a 100 percentage bonus on all new meme income!"
	if !strings.Contains(output, expectedOutput) {
		t.Log("Prestige output did not report proper prestige error!")
		t.Error(output)
	}
}

func TestPrestigeFail(t *testing.T) {
	botSess := interaction.NewConsoleSession()
	message := interaction.NewMessageEvent()
	db := handlers.DbGet()
	id := strconv.Itoa(int(time.Now().UnixNano()))
	author := discordgo.User{
		ID:       id,
		Username: "admin",
	}

	user := handlers.UserGet(&author, db)
	handlers.MoneyAdd(&user, 10000, "tip", db)
	user = handlers.UserGet(&author, db)

	text := "!prestige"
	message.Message.Content = text
	message.Message.Author = &author

	output := capStdout(botSess, message)

	user = handlers.UserGet(&author, db)

	if user.CurMoney != 11000 {
		t.Log("User CurMoney was changed, it should have remained at 11000.")
		t.Error(output)
	}

	expectedOutputs := []string{"You do not have enough miners to Prestige, you need 100 more.", "You do not have enough robots to Prestige, you need 100 more.", "You do not have enough swarms to Prestige, you need 100 more.", "You do not have enough frackers to Prestige, you need 100 more."}
	for _, expectedOutput := range expectedOutputs {
		if !strings.Contains(output, expectedOutput) {
			t.Log("Prestige output did not report proper prestige error!")
			t.Error(output)
		}
	}
}

func TestPrestigeTipFail(t *testing.T) {
	botSess := interaction.NewConsoleSession()
	message := interaction.NewMessageEvent()
	db := handlers.DbGet()
	id := strconv.Itoa(int(time.Now().UnixNano()))
	author := discordgo.User{
		ID:       id,
		Username: "admin",
	}
	user := handlers.UserGet(&author, db)
	handlers.MoneyAdd(&user, 10000, "tip", db)
	user = handlers.UserGet(&author, db)

	id = strconv.Itoa(int(time.Now().UnixNano()))
	tipeeDiscordUser := discordgo.User{
		ID:       id,
		Username: "tipee",
	}
	tipee := handlers.UserGet(&tipeeDiscordUser, db)
	tipee.PrestigeLevel = 1
	handlers.UpdateUnits(&tipee, db)

	text := "!tip 10000 memes @tipee"
	message.Message.Content = text
	message.Message.Author = &author
	message.Message.Mentions = []*discordgo.User{&tipeeDiscordUser}

	output := capStdout(botSess, message)

	tipee = handlers.UserGet(&tipeeDiscordUser, db)
	user = handlers.UserGet(&author, db)

	if tipee.CurMoney != 1000 {
		t.Log("tipee CurMoney was changed, it should have remained at 1000.")
		t.Error(output)
	}

	if user.CurMoney != 11000 {
		t.Log("User CurMoney was changed, it should have remained at 11000.")
		t.Error(output)
	}

	expectedOutput := ("admin tried to give 10,000 memes to tipee; but admin's prestige level is 0 and tipee's prestige level is 1. Memes can not be tipped up prestige levels; only to equal and lower prestige levels.")
	if !strings.Contains(output, expectedOutput) {
		t.Log("Tipping output did not report proper prestige tipping error!")
		t.Error(output)
	}
}

func TestPrestigeTipSuccess(t *testing.T) {
	botSess := interaction.NewConsoleSession()
	message := interaction.NewMessageEvent()
	db := handlers.DbGet()
	id := strconv.Itoa(int(time.Now().UnixNano()))
	author := discordgo.User{
		ID:       id,
		Username: "admin",
	}
	user := handlers.UserGet(&author, db)
	handlers.MoneyAdd(&user, 10000, "tip", db)
	user.PrestigeLevel = 1
	handlers.UpdateUnits(&user, db)
	user = handlers.UserGet(&author, db)

	id = strconv.Itoa(int(time.Now().UnixNano()))
	tipeeDiscordUser := discordgo.User{
		ID:       id,
		Username: "tipee",
	}
	tipee := handlers.UserGet(&tipeeDiscordUser, db)
	tipee.PrestigeLevel = 1
	handlers.UpdateUnits(&tipee, db)

	text := "!tip 10000 memes @tipee"
	message.Message.Content = text
	message.Message.Author = &author
	message.Message.Mentions = []*discordgo.User{&tipeeDiscordUser}

	output := capStdout(botSess, message)

	tipee = handlers.UserGet(&tipeeDiscordUser, db)
	user = handlers.UserGet(&author, db)

	if tipee.CurMoney != 11000 {
		t.Log("tipee CurMoney was not changed, it should be 11000.")
		t.Error(output)
	}

	if user.CurMoney != 1000 {
		t.Log("User CurMoney was not changed, it should be 0.")
		t.Error(output)
	}

	expectedOutput := ("admin gave 10,000 memes to tipeeadmin gave 10,000 memes to tipee")
	if !strings.Contains(output, expectedOutput) {
		t.Log("Tipping output did not report proper prestige tipping error!")
		t.Error(output)
	}
}

func TestPrestigeMineSuccess(t *testing.T) {
	botSess := interaction.NewConsoleSession()
	message := interaction.NewMessageEvent()
	db := handlers.DbGet()
	id := strconv.Itoa(int(time.Now().UnixNano()))
	author := discordgo.User{
		ID:       id,
		Username: "admin",
	}
	user := handlers.UserGet(&author, db)
	user.PrestigeLevel = 1
	handlers.UpdateUnits(&user, db)
	user = handlers.UserGet(&author, db)

	text := "!mine"
	message.Message.Content = text
	message.Message.Author = &author

	output := capStdout(botSess, message)

	user = handlers.UserGet(&author, db)

	if user.CurMoney != 1000+((1+user.PrestigeLevel)*100) {
		t.Log("User CurMoney was not changed, it should be 1200.")
		t.Error(output)
	}

	expectedOutput := ("admin mined for a while and managed to scrounge up 200 dusty memes")
	if !strings.Contains(output, expectedOutput) {
		t.Log("Mining output did not report proper prestige mining output!")
		t.Error(output)
	}
}

func TestPrestigeMineSuccessHighPrestige(t *testing.T) {
	botSess := interaction.NewConsoleSession()
	message := interaction.NewMessageEvent()
	db := handlers.DbGet()
	id := strconv.Itoa(int(time.Now().UnixNano()))
	author := discordgo.User{
		ID:       id,
		Username: "admin",
	}
	user := handlers.UserGet(&author, db)
	user.PrestigeLevel = 1000
	handlers.UpdateUnits(&user, db)
	user = handlers.UserGet(&author, db)

	text := "!mine"
	message.Message.Content = text
	message.Message.Author = &author

	output := capStdout(botSess, message)

	user = handlers.UserGet(&author, db)

	if user.CurMoney != 1000+((1+user.PrestigeLevel)*300) {
		t.Log("User CurMoney was not changed, it should be 301300.")
		t.Error(output)
	}

	expectedOutput := ("admin mined for a bit and found an uncommon pepe worth 300,300 memes!")
	if !strings.Contains(output, expectedOutput) {
		t.Log("Mining output did not report proper prestige mining output!")
		t.Error(output)
	}
}

func TestPrestigeHackWin(t *testing.T) {
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
	seed := int64(37)
	hackers := "38"
	botnets := "29"

	rand.Seed(seed)

	user := handlers.UserGet(&author, db)
	user.Hacker = 100
	user.Botnet = 100
	user.PrestigeLevel = 1
	handlers.UpdateUnits(&user, db)
	//t.Fatal(user.Botnet)
	//user = handlers.UserGet(&author, db)

	targetUser := handlers.UserGet(&target, db)
	targetUser.Miner = 14
	targetUser.HackSeed = seed
	targetUser.CollectTime = targetUser.CollectTime.Add(-10 * time.Minute)
	targetUser.PrestigeLevel = 1
	handlers.UpdateUnits(&targetUser, db)

	text := "!hack " + hackers + " " + botnets + " @target"
	message.Message.Content = text
	message.Message.Mentions = append(message.Message.Mentions, &target)
	message.Message.Author = &author

	output := capStdout(botSess, message)
	newUser := handlers.UserGet(&author, db)
	newTarget := handlers.UserGet(&target, db)

	// verify the thief's stats
	userMoneyDiff := newUser.CurMoney - user.CurMoney
	if userMoneyDiff != targetUser.Miner*(1+user.PrestigeLevel) {
		t.Log("The thief's money wasn't updated properly.")
		numLog(t, targetUser.Miner*user.PrestigeLevel, userMoneyDiff)
		t.Error(output)
	}

	userStoleDiff := newUser.HackedMoney - user.HackedMoney
	if userStoleDiff != targetUser.Miner*(1+user.PrestigeLevel) {
		t.Log("The thief's HackedMoney wasn't updated properly.")
		numLog(t, targetUser.Miner*(1+user.PrestigeLevel), userStoleDiff)
		t.Log(newUser.HackedMoney)
		t.Error(output)
	}

	if newUser.Hacker != user.Hacker {
		t.Log("The thief lost hackers on a successful hack!")
		numLog(t, newUser.Hacker, user.Hacker)
		t.Error(output)
	}

	if newUser.Botnet != user.Botnet {
		t.Log("The thief lost botnets on a successful hack!")
		numLog(t, newUser.Botnet, user.Botnet)
		t.Error(output)
	}

	// verify the target's stats
	targetStoleDiff := newTarget.StolenFromMoney - targetUser.StolenFromMoney
	if targetStoleDiff != targetUser.Miner*(1+user.PrestigeLevel) {
		t.Log("The target's StolenFromMoney wasn't updated properly.")
		numLog(t, targetUser.Miner*(1+user.PrestigeLevel), targetStoleDiff)
		t.Log(newUser.HackedMoney)
		t.Error(output)
	}

	if newTarget.CollectTime == targetUser.CollectTime {
		t.Log("The target's CollectTime wasn't updated properly.")
		t.Error(output)
	}

	if newTarget.HackAttempts != 0 {
		t.Log("The target's HackAttempts was not reset back to 0")
		numLog(t, 0, newTarget.HackAttempts)
		t.Error(output)
	}

	if newTarget.HackSeed != 0 {
		t.Log("The target's HackSeed was not reset")
		t.Log("The seed was still: " + strconv.Itoa(int(newTarget.HackSeed)))
		t.Error(output)
	}

	expectedOutput := ("The hack was successful, " + user.Username + " stole " + humanize.Comma(int64(targetUser.Miner*(1+user.PrestigeLevel))) + " dank memes from " + target.Username)
	if !strings.Contains(output, expectedOutput) {
		t.Log("Successful hacking output to channel was not what was expected.")
		t.Error(output)
	}
}

func TestPrestigeHackFail(t *testing.T) {
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
	seed := int64(37)
	hackers := "100"
	botnets := "88"

	rand.Seed(seed)

	user := handlers.UserGet(&author, db)
	user.Hacker = 100
	user.Botnet = 100
	user.PrestigeLevel = 1
	handlers.UpdateUnits(&user, db)
	//t.Fatal(user.Botnet)
	//user = handlers.UserGet(&author, db)

	targetUser := handlers.UserGet(&target, db)
	targetUser.Miner = 14
	targetUser.HackSeed = seed
	targetUser.CollectTime = targetUser.CollectTime.Add(-10 * time.Minute)
	handlers.UpdateUnits(&targetUser, db)

	text := "!hack " + hackers + " " + botnets + " @target"
	message.Message.Content = text
	message.Message.Mentions = append(message.Message.Mentions, &target)
	message.Message.Author = &author

	output := capStdout(botSess, message)
	newUser := handlers.UserGet(&author, db)
	newTarget := handlers.UserGet(&target, db)

	// verify the thief's stats
	userMoneyDiff := newUser.CurMoney - user.CurMoney
	if userMoneyDiff != 0 {
		t.Log("The thief's money was updated on an improper hack.")
		numLog(t, targetUser.Miner*user.PrestigeLevel, userMoneyDiff)
		t.Error(output)
	}

	userStoleDiff := newUser.HackedMoney - user.HackedMoney
	if userStoleDiff != 0 {
		t.Log("The thief's HackedMoney was update on an improper hack.")
		numLog(t, targetUser.Miner*(1+user.PrestigeLevel), userStoleDiff)
		t.Log(newUser.HackedMoney)
		t.Error(output)
	}

	if newUser.Hacker != user.Hacker {
		t.Log("The thief lost hackers on a improper hack!")
		numLog(t, newUser.Hacker, user.Hacker)
		t.Error(output)
	}

	if newUser.Botnet != user.Botnet {
		t.Log("The thief lost botnets on a improper hack!")
		numLog(t, newUser.Botnet, user.Botnet)
		t.Error(output)
	}

	// verify the target's stats
	targetStoleDiff := newTarget.StolenFromMoney - targetUser.StolenFromMoney
	if targetStoleDiff != 0 {
		t.Log("The target's StolenFromMoney was updated on an improper hack.")
		numLog(t, targetUser.Miner*(1+user.PrestigeLevel), targetStoleDiff)
		t.Log(newUser.HackedMoney)
		t.Error(output)
	}

	if newTarget.CollectTime != targetUser.CollectTime {
		t.Log("The target's CollectTime was changed on an improper hack.")
		t.Error(output)
	}

	if newTarget.HackAttempts != 0 {
		t.Log("The target's HackAttempts should remain as 0")
		numLog(t, 0, newTarget.HackAttempts)
		t.Error(output)
	}

	if newTarget.HackSeed == 0 {
		t.Log("The target's HackSeed was reset")
		t.Log("The seed was still: " + strconv.Itoa(int(newTarget.HackSeed)))
		t.Error(output)
	}

	// verify the output
	expectedOutput := ("thief tried to hack target; but thief's prestige level is 1 and target's prestige level is 0.")
	if !strings.Contains(output, expectedOutput) {
		t.Log("Successful hacking output to channel was not what was expected.")
		t.Error(output)
	}
}

func TestPrestigeHackWinHighPrestige(t *testing.T) {
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
	seed := int64(37)
	hackers := "38"
	botnets := "29"

	rand.Seed(seed)

	user := handlers.UserGet(&author, db)
	user.Hacker = 100
	user.Botnet = 100
	user.PrestigeLevel = 1000
	handlers.UpdateUnits(&user, db)
	//t.Fatal(user.Botnet)
	//user = handlers.UserGet(&author, db)

	targetUser := handlers.UserGet(&target, db)
	targetUser.Miner = 14
	targetUser.HackSeed = seed
	targetUser.CollectTime = targetUser.CollectTime.Add(-10 * time.Minute)
	targetUser.PrestigeLevel = 1000
	handlers.UpdateUnits(&targetUser, db)

	text := "!hack " + hackers + " " + botnets + " @target"
	message.Message.Content = text
	message.Message.Mentions = append(message.Message.Mentions, &target)
	message.Message.Author = &author

	output := capStdout(botSess, message)
	newUser := handlers.UserGet(&author, db)
	newTarget := handlers.UserGet(&target, db)

	// verify the thief's stats
	userMoneyDiff := newUser.CurMoney - user.CurMoney
	if userMoneyDiff != targetUser.Miner*(1+user.PrestigeLevel) {
		t.Log("The thief's money wasn't updated properly.")
		numLog(t, targetUser.Miner*user.PrestigeLevel, userMoneyDiff)
		t.Error(output)
	}

	userStoleDiff := newUser.HackedMoney - user.HackedMoney
	if userStoleDiff != targetUser.Miner*(1+user.PrestigeLevel) {
		t.Log("The thief's HackedMoney wasn't updated properly.")
		numLog(t, targetUser.Miner*(1+user.PrestigeLevel), userStoleDiff)
		t.Log(newUser.HackedMoney)
		t.Error(output)
	}

	if newUser.Hacker != user.Hacker {
		t.Log("The thief lost hackers on a successful hack!")
		numLog(t, newUser.Hacker, user.Hacker)
		t.Error(output)
	}

	if newUser.Botnet != user.Botnet {
		t.Log("The thief lost botnets on a successful hack!")
		numLog(t, newUser.Botnet, user.Botnet)
		t.Error(output)
	}

	// verify the target's stats
	targetStoleDiff := newTarget.StolenFromMoney - targetUser.StolenFromMoney
	if targetStoleDiff != targetUser.Miner*(1+user.PrestigeLevel) {
		t.Log("The target's StolenFromMoney wasn't updated properly.")
		numLog(t, targetUser.Miner*(1+user.PrestigeLevel), targetStoleDiff)
		t.Log(newUser.HackedMoney)
		t.Error(output)
	}

	if newTarget.CollectTime == targetUser.CollectTime {
		t.Log("The target's CollectTime wasn't updated properly.")
		t.Error(output)
	}

	if newTarget.HackAttempts != 0 {
		t.Log("The target's HackAttempts was not reset back to 0")
		numLog(t, 0, newTarget.HackAttempts)
		t.Error(output)
	}

	if newTarget.HackSeed != 0 {
		t.Log("The target's HackSeed was not reset")
		t.Log("The seed was still: " + strconv.Itoa(int(newTarget.HackSeed)))
		t.Error(output)
	}

	expectedOutput := ("The hack was successful, " + user.Username + " stole " + humanize.Comma(int64(targetUser.Miner*(1+user.PrestigeLevel))) + " dank memes from " + target.Username)
	if !strings.Contains(output, expectedOutput) {
		t.Log("Successful hacking output to channel was not what was expected.")
		t.Error(output)
	}
}

func TestPrestigeCollect(t *testing.T) {
	botSess := interaction.NewConsoleSession()
	message := interaction.NewMessageEvent()
	db := handlers.DbGet()
	id := strconv.Itoa(int(time.Now().UnixNano()))
	author := discordgo.User{
		ID:       id,
		Username: "admin",
	}

	user := handlers.UserGet(&author, db)
	user.Miner = 1000
	user.HackAttempts = 2
	user.HackSeed = int64(12301002)
	user.CollectTime = user.CollectTime.Add(-10 * time.Minute)
	user.PrestigeLevel = 1000
	handlers.UpdateUnits(&user, db)
	user = handlers.UserGet(&author, db)

	text := "!collect"
	message.Message.Content = text
	message.Message.Author = &author

	output := capStdout(botSess, message)
	newUser := handlers.UserGet(&author, db)

	if newUser.CurMoney != (1000 + 1000*(1+user.PrestigeLevel)) {
		t.Log("User CurMoney was not updated with collected amount!")
		t.Error(output)
	}

	if newUser.CollectTime == user.CollectTime {
		t.Log("User CollectTime was not reset upon collection.")
		t.Error(output)
	}

	expectedOutput := ("admin collected " + humanize.Comma(int64((1000 * (1 + user.PrestigeLevel)))) + " memes!")
	if !strings.Contains(output, expectedOutput) {
		t.Log("Collecting output did not report proper meme collection amount!")
		t.Error(output)
	}
	if newUser.HackAttempts != 0 {
		t.Log("HackAttempts were not reset to 0 after collection.")
		t.Error(output)
	}
	if newUser.HackSeed != 0 {
		t.Log("HackSeed was not reset to 0 after collection.")
		t.Error(output)
	}
}

func TestPrestigeTwoSuccess(t *testing.T) {
	botSess := interaction.NewConsoleSession()
	message := interaction.NewMessageEvent()
	db := handlers.DbGet()
	id := strconv.Itoa(int(time.Now().UnixNano()))
	author := discordgo.User{
		ID:       id,
		Username: "admin",
	}

	user := handlers.UserGet(&author, db)
	handlers.MoneyAdd(&user, 10000, "tip", db)
	user.Miner = 400
	user.Robot = 400
	user.Swarm = 400
	user.Fracker = 400
	user.PrestigeLevel = 1
	handlers.UpdateUnits(&user, db)

	user = handlers.UserGet(&author, db)

	text := "!prestige YESIMSURE"
	message.Message.Content = text
	message.Message.Author = &author

	output := capStdout(botSess, message)

	user = handlers.UserGet(&author, db)
	if user.CurMoney != 1000 {
		t.Log("User CurMoney was not changed, it should have reset to 1000.")
		t.Error(output)
	}

	if user.Miner != 0 {
		t.Log("User Miner count was not reset to 0.")
		t.Error(user.Miner)
	}

	if user.PrestigeLevel != 2 {
		t.Log(output)
		t.Error("User Prestige level is not 2.")
	}

	expectedOutput := "You have been reset! Congratulations, you are now prestige level 2, which means you get a 200 percentage bonus on all new meme income!"
	if !strings.Contains(output, expectedOutput) {
		t.Log("Prestige output did not report proper prestige error!")
		t.Error(output)
	}
}

func TestTipSelf(t *testing.T) {
	botSess := interaction.NewConsoleSession()
	message := interaction.NewMessageEvent()
	db := handlers.DbGet()
	id := strconv.Itoa(int(time.Now().UnixNano()))
	author := discordgo.User{
		ID:       id,
		Username: "admin",
	}
	user := handlers.UserGet(&author, db)
	handlers.MoneyAdd(&user, 10000, "tip", db)
	user.PrestigeLevel = 1
	handlers.UpdateUnits(&user, db)
	user = handlers.UserGet(&author, db)

	text := "!tip 10000 memes @tipee"
	message.Message.Content = text
	message.Message.Author = &author
	message.Message.Mentions = []*discordgo.User{&author}

	output := capStdout(botSess, message)

	user = handlers.UserGet(&author, db)

	if user.CurMoney != 11000 {
		t.Log("author CurMoney was changed, it should be 11000.")
		t.Error(user.CurMoney)
		t.Error(output)
	}

	expectedOutput := ("admin gave 10,000 memes to adminadmin gave 10,000 memes to admin")
	if !strings.Contains(output, expectedOutput) {
		t.Log("Tipping output did not report proper prestige tipping error!")
		t.Error(output)
	}
	//spew.Dump(user)
	//t.Error(output)
}

func TestTipOverflow(t *testing.T) {
	botSess := interaction.NewConsoleSession()
	message := interaction.NewMessageEvent()
	db := handlers.DbGet()
	id := strconv.Itoa(int(time.Now().UnixNano()))
	author := discordgo.User{
		ID:       id,
		Username: "admin",
	}
	user := handlers.UserGet(&author, db)
	handlers.MoneyAdd(&user, 1000000000000000000, "tip", db)
	handlers.UpdateUnits(&user, db)
	user = handlers.UserGet(&author, db)

	id = strconv.Itoa(int(time.Now().UnixNano()))
	tipeeDiscordUser := discordgo.User{
		ID:       id,
		Username: "tipee",
	}
	tipee := handlers.UserGet(&tipeeDiscordUser, db)
	handlers.MoneyAdd(&tipee, 9000000000000000000, "tip", db)
	handlers.UpdateUnits(&tipee, db)

	text := "!tip 1000000000000000000 memes @tipee"
	message.Message.Content = text
	message.Message.Author = &author
	message.Message.Mentions = []*discordgo.User{&tipeeDiscordUser}

	output := capStdout(botSess, message)

	tipee = handlers.UserGet(&tipeeDiscordUser, db)
	user = handlers.UserGet(&author, db)

	fmt.Println(tipee.CurMoney)
	if tipee.CurMoney != 9000000000000001000 {
		t.Log("tipee CurMoney was changed, it should have remained at 1000000000000001000.")
		t.Error(output)
	}

	if user.CurMoney != 1000000000000001000 {
		t.Log("User CurMoney was changed, it should have remained at 1000000000000001000.")
		t.Error(output)
	}

	expectedOutput := ("You're trying to tip too many memes, try tipping less memes.")
	if !strings.Contains(output, expectedOutput) {
		t.Log("Tipping output did not report proper prestige tipping error!")
		t.Error(output)
	}
}

func TestNegativeBalanceReset(t *testing.T) {
	botSess := interaction.NewConsoleSession()
	message := interaction.NewMessageEvent()
	db := handlers.DbGet()
	id := strconv.Itoa(int(time.Now().UnixNano()))
	author := discordgo.User{
		ID:       id,
		Username: "admin",
	}
	user := handlers.UserGet(&author, db)
	handlers.MoneyAdd(&user, -20000, "tip", db)
	handlers.UpdateUnits(&user, db)
	user = handlers.UserGet(&author, db)

	text := "!balance"
	message.Message.Content = text
	message.Message.Author = &author

	output := capStdout(botSess, message)

	user = handlers.UserGet(&author, db)

	if user.CurMoney != 19000 {
		t.Log("User CurMoney was not changed, it should 19000.")
		t.Log(user.CurMoney)
		t.Error(output)
	}

	expectedOutput := ("19,000")
	if !strings.Contains(output, expectedOutput) {
		t.Log("Tipping output did not report meme amount.")
		t.Error(output)
	}
}
