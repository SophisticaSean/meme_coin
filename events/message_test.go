package events

import (
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

func numLog(t *testing.T, expected int, actual int) {
	t.Log("We expected " + strconv.Itoa(expected) + ", but got: " + strconv.Itoa(actual))
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
	if !strings.Contains(output, author.Username+" won "+strconv.Itoa(gambleAmount*99)+" memes.") {
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
	if !strings.Contains(output, author.Username+" lost "+strconv.Itoa(gambleAmount)+" memes.") {
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
	hackers := "100"
	botnets := "88"

	rand.Seed(seed)

	user := handlers.UserGet(&author, db)
	userUnits := handlers.UnitsGet(&author, db)
	userUnits.Hacker = 100
	userUnits.Botnet = 100
	handlers.UpdateUnits(&userUnits, db)
	userUnits = handlers.UnitsGet(&author, db)

	targetUser := handlers.UserGet(&target, db)
	targetUnits := handlers.UnitsGet(&target, db)
	targetUnits.Miner = 14
	targetUnits.HackSeed = seed
	targetUnits.CollectTime = targetUnits.CollectTime.Add(-10 * time.Minute)
	handlers.UpdateUnits(&targetUnits, db)

	text := "!hack " + hackers + " " + botnets + " @target"
	message.Message.Content = text
	message.Message.Mentions = append(message.Message.Mentions, &target)
	message.Message.Author = &author

	output := capStdout(botSess, message)
	newUser := handlers.UserGet(&author, db)
	newUserUnits := handlers.UnitsGet(&author, db)
	newTargetUser := handlers.UserGet(&target, db)
	newTargetUnits := handlers.UnitsGet(&target, db)

	// verify the thief's stats
	userMoneyDiff := newUser.CurMoney - user.CurMoney
	if userMoneyDiff != targetUnits.Miner {
		t.Log("The thief's money wasn't updated properly.")
		numLog(t, targetUnits.Miner, userMoneyDiff)
		t.Error(output)
	}

	userStoleDiff := newUser.HackedMoney - user.HackedMoney
	if userStoleDiff != targetUnits.Miner {
		t.Log("The thief's HackedMoney wasn't updated properly.")
		numLog(t, targetUnits.Miner, userStoleDiff)
		t.Log(newUser.HackedMoney)
		t.Error(output)
	}

	if newUserUnits.Hacker != userUnits.Hacker {
		t.Log("The thief lost hackers on a successful hack!")
		numLog(t, newUserUnits.Hacker, userUnits.Hacker)
		t.Error(output)
	}

	if newUserUnits.Botnet != userUnits.Botnet {
		t.Log("The thief lost botnets on a successful hack!")
		numLog(t, newUserUnits.Botnet, userUnits.Botnet)
		t.Error(output)
	}

	// verify the target's stats
	targetStoleDiff := newTargetUser.StolenFromMoney - targetUser.StolenFromMoney
	if targetStoleDiff != targetUnits.Miner {
		t.Log("The target's StolenFromMoney wasn't updated properly.")
		numLog(t, targetUnits.Miner, targetStoleDiff)
		t.Log(newUser.HackedMoney)
		t.Error(output)
	}

	if newTargetUnits.CollectTime == targetUnits.CollectTime {
		t.Log("The target's CollectTime wasn't updated properly.")
		t.Error(output)
	}

	if newTargetUnits.HackAttempts != 0 {
		t.Log("The target's HackAttempts was not reset back to 0")
		numLog(t, 0, newTargetUnits.HackAttempts)
		t.Error(output)
	}

	if newTargetUnits.HackSeed != 0 {
		t.Log("The target's HackSeed was not reset")
		t.Log("The seed was still: " + strconv.Itoa(int(newTargetUnits.HackSeed)))
		t.Error(output)
	}

	// verify the output
	if !strings.Contains(output, "totalIterations: "+botnets) {
		t.Log("Number of iterations was incorrect.")
		t.Error(output)
	}
	if !strings.Contains(output, "iterationLimit: "+botnets) {
		t.Log("IterationLimit was incorrect.")
		t.Error(output)
	}
	if !strings.Contains(output, "seed: "+strconv.Itoa(int(seed))) {
		t.Log("The seed was incorrect.")
		t.Error(output)
	}
	if !strings.Contains(output, "totalFitness: 32") {
		t.Log("totalFitness was not what we expected.")
		t.Error(output)
	}
	if !strings.Contains(output, "targetLength: 32") {
		t.Log("targetLength of password was incorrect.")
		t.Error(output)
	}
	if !strings.Contains(output, "populationSize: 9") {
		t.Log("populationSize was not hardcapped to what we expected!")
		t.Error(output)
	}
	expectedOutput := ("The hack was successful, " + user.Username + " stole " + strconv.Itoa(targetUnits.Miner) + " dank memes from " + target.Username)
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
	seed := int64(1281) // botnet loss
	hackers := "300"
	botnets := "2000"

	user := handlers.UserGet(&author, db)
	userUnits := handlers.UnitsGet(&author, db)
	userUnits.Hacker = 300
	userUnits.Botnet = 2000
	handlers.UpdateUnits(&userUnits, db)
	userUnits = handlers.UnitsGet(&author, db)

	targetUser := handlers.UserGet(&target, db)
	targetUnits := handlers.UnitsGet(&target, db)
	targetUnits.Miner = 14
	targetUnits.Cypher = 307
	targetUnits.HackSeed = seed
	targetUnits.CollectTime = targetUnits.CollectTime.Add(-10 * time.Minute)
	handlers.UpdateUnits(&targetUnits, db)

	rand.Seed(seed)

	text := "!hack " + hackers + " " + botnets + " @target"
	message.Message.Content = text
	message.Message.Mentions = append(message.Message.Mentions, &target)
	message.Message.Author = &author

	output := capStdout(botSess, message)
	newUser := handlers.UserGet(&author, db)
	newUserUnits := handlers.UnitsGet(&author, db)
	newTargetUser := handlers.UserGet(&target, db)
	newTargetUnits := handlers.UnitsGet(&target, db)

	// verify the thief's stats
	userMoneyDiff := newUser.CurMoney - user.CurMoney
	if userMoneyDiff != 0 {
		t.Log("The thief's money wasn't updated properly.")
		numLog(t, targetUnits.Miner, userMoneyDiff)
		t.Error(output)
	}

	userStoleDiff := newUser.HackedMoney - user.HackedMoney
	if userStoleDiff != 0 {
		t.Log("The thief's HackedMoney wasn't updated properly.")
		numLog(t, targetUnits.Miner, userStoleDiff)
		t.Log(newUser.HackedMoney)
		t.Error(output)
	}

	if newUserUnits.Hacker != userUnits.Hacker-2 {
		t.Log("The thief did not lose hackers on a failed hack!")
		numLog(t, newUserUnits.Hacker, userUnits.Hacker)
		t.Error(output)
	}

	if newUserUnits.Botnet != userUnits.Botnet-23 {
		t.Log("The thief did not lose botnets on a failed hack!")
		numLog(t, newUserUnits.Botnet, userUnits.Botnet)
		t.Error(output)
	}

	// verify the target's stats
	targetStoleDiff := newTargetUser.StolenFromMoney - targetUser.StolenFromMoney
	if targetStoleDiff != 0 {
		t.Log("The target's StolenFromMoney was updated on a failed hack!")
		numLog(t, targetUnits.Miner, targetStoleDiff)
		t.Log(newUser.HackedMoney)
		t.Error(output)
	}

	if newTargetUnits.CollectTime != targetUnits.CollectTime {
		t.Log("The target's CollectTime was reset on a failed hack!")
		t.Error(output)
	}

	if newTargetUnits.HackAttempts != targetUnits.HackAttempts+1 {
		t.Log("The target's HackAttempts was not incremented.")
		numLog(t, 0, newTargetUnits.HackAttempts)
		t.Error(output)
	}

	if newTargetUnits.HackSeed != targetUnits.HackSeed {
		t.Log("The target's HackSeed reset on a failed Hack!")
		t.Error(output)
	}

	// verify the output
	if !strings.Contains(output, "totalIterations: 426") {
		t.Log("Number of iterations was incorrect.")
		t.Error(output)
	}
	if !strings.Contains(output, "iterationLimit: "+botnets) {
		t.Log("IterationLimit was incorrect.")
		t.Error(output)
	}
	if !strings.Contains(output, "seed: "+strconv.Itoa(int(seed))) {
		t.Log("The seed was incorrect.")
		t.Error(output)
	}
	if !strings.Contains(output, "totalFitness: 640") {
		t.Log("totalFitness was not what we expected.")
		t.Error(output)
	}
	if !strings.Contains(output, "targetLength: 640") {
		t.Log("targetLength of password was incorrect.")
		t.Error(output)
	}
	if !strings.Contains(output, "populationSize: 47") {
		t.Log("populationSize was not hardcapped to what we expected!")
		t.Error(output)
	}
	expectedOutput := ("1977 botnets left")
	if !strings.Contains(output, expectedOutput) {
		t.Log("Failed hacking output to channel was not what was expected.")
		t.Error(output)
	}
	expectedOutput = ("298 hackers left")
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
	userUnits := handlers.UnitsGet(&author, db)
	userUnits.Hacker = 100
	userUnits.Botnet = 100
	handlers.UpdateUnits(&userUnits, db)
	userUnits = handlers.UnitsGet(&author, db)

	targetUser := handlers.UserGet(&target, db)
	targetUnits := handlers.UnitsGet(&target, db)
	targetUnits.Miner = 14
	targetUnits.Cypher = 307
	targetUnits.HackSeed = seed
	targetUnits.CollectTime = targetUnits.CollectTime.Add(-10 * time.Minute)
	handlers.UpdateUnits(&targetUnits, db)

	rand.Seed(seed)

	text := "!hack " + hackers + " " + botnets + " @target"
	message.Message.Content = text
	message.Message.Mentions = append(message.Message.Mentions, &target)
	message.Message.Author = &author

	output := capStdout(botSess, message)
	newUser := handlers.UserGet(&author, db)
	newUserUnits := handlers.UnitsGet(&author, db)
	newTargetUser := handlers.UserGet(&target, db)
	newTargetUnits := handlers.UnitsGet(&target, db)

	// compare old and new users, make sure nothing has changed.
	if !reflect.DeepEqual(user, newUser) {
		t.Log("user did not equal newUser")
		t.Error(output)
	}
	if !reflect.DeepEqual(userUnits, newUserUnits) {
		t.Log("userUnits did not equal newUserUnits")
		t.Error(output)
	}
	if !reflect.DeepEqual(targetUser, newTargetUser) {
		t.Log("targetUser did not equal newTargetUser")
		t.Error(output)
	}
	if !reflect.DeepEqual(targetUnits, newTargetUnits) {
		t.Log("targetUnits did not equal newTargetUnits")
		t.Error(output)
	}

	// make sure output is correct
	expectedOutput := ("You don't have enough botnets for the requested hack need: " + botnets + " have: " + strconv.Itoa(userUnits.Botnet))
	if !strings.Contains(output, expectedOutput) {
		t.Log("Hacking output did not report botnet mismatch.")
		t.Error(output)
	}
	expectedOutput = ("You don't have enough hackers for the requested hack need: " + hackers + " have: " + strconv.Itoa(userUnits.Hacker))
	if !strings.Contains(output, expectedOutput) {
		t.Log("Hacking output did not report hacker mismatch.")
		t.Error(output)
	}
}
