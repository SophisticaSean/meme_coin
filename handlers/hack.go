package handlers

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"
)

const (
	lossChances = 10
)

func Ftoa(float float64) string {
	floatString := strconv.FormatFloat(float, 'f', -1, 64)
	return floatString
}

func processHackingLosses(units *UserUnits, db *sqlx.DB) string {
	message := ""
	hackerLosses := 0
	botnetLosses := 0
	for i := 0; i <= units.Hacker; i++ {
		if rand.Intn(100) < lossChances {
			hackerLosses += 1
		}
	}
	if hackerLosses != 0 {
		units.Hacker = units.Hacker - hackerLosses
		message = message + "`Your hacking got " + strconv.Itoa(hackerLosses) + " hackers arrested by the FBI!`\r`You now have " + strconv.Itoa(units.Hacker) + " hackers left.`\r"
	}
	for i := 0; i <= units.Botnet; i++ {
		if rand.Intn(100) < lossChances {
			botnetLosses += 1
		}
	}
	if hackerLosses != 0 {
		units.Botnet = units.Botnet - botnetLosses
		message = message + "`Your hacking was detected and got some botnets discovered, some whitehat released a Day 0 paper and patch defeating " + strconv.Itoa(botnetLosses) + " of your botnets.`\r`You now have " + strconv.Itoa(units.Botnet) + " botnets left.`\r"
	}
	UpdateUnits(units, db)
	return message
}

func Hack(s *discordgo.Session, m *discordgo.MessageCreate, db *sqlx.DB) {
	args := strings.Split(m.Content, " ")
	defaultSyntax := "`!hack <amount_of_hackers> <amount_of_botnets> @person`"
	message := ""
	if len(args) != 4 {
		message = "Too many or too few arguments, hack like this boi: " + defaultSyntax
		_, _ = s.ChannelMessageSend(m.ChannelID, message)
		return
	}
	mentions := m.Mentions
	if len(mentions) != 1 {
		message := "Have to hack 1 person " + defaultSyntax
		_, _ = s.ChannelMessageSend(m.ChannelID, message)
		return
	}
	totalMemes, _, targetUnits := totalMemesEarned(mentions[0], db)
	target := UserGet(mentions[0], db)
	authorUnits := UnitsGet(m.Author, db)
	author := UserGet(m.Author, db)

	if targetUnits.HackSeed == 0 || targetUnits.HackAttempts >= 10 {
		discordID, err := strconv.Atoi(targetUnits.DID)
		// shouldn't happen
		if err != nil {
			fmt.Println("SUPER BAD ERROR: ", err)
			return
		}
		targetUnits.HackAttempts = 0
		targetUnits.HackSeed = (time.Now().UnixNano() + int64(discordID))
		UpdateUnits(&targetUnits, db)
	}
	seed := targetUnits.HackSeed

	popSize, err := strconv.Atoi(args[1])
	if err != nil || popSize < 1 {
		message = "your amount of hackers argument (`!hack <this_argument> <amount_of_botnets>`) is too large, too small, or not a number."
		_, _ = s.ChannelMessageSend(m.ChannelID, message)
		return
	}
	if popSize > authorUnits.Hacker {
		message = "You don't have enough hackers for the requested hack need: " + args[1] + " have: " + strconv.Itoa(authorUnits.Hacker)
		_, _ = s.ChannelMessageSend(m.ChannelID, message)
		return
	}

	iterationLimit, err := strconv.Atoi(args[2])
	if err != nil || iterationLimit < 1 {
		message = "your amount of bot_nets argument (`!hack <amount_of_hackers> <this_argument>`) is too large, too small, or not a number."
		_, _ = s.ChannelMessageSend(m.ChannelID, message)
		return
	}
	if iterationLimit > authorUnits.Botnet {
		message = "You don't have enough hackers for the requested hack need: " + args[2] + " have: " + strconv.Itoa(authorUnits.Botnet)
		_, _ = s.ChannelMessageSend(m.ChannelID, message)
		return
	}

	maxStringLength := targetUnits.Cypher + 5
	fitnessPercentage, generationPercentage := hackSimulate(seed, popSize, iterationLimit, maxStringLength)
	if fitnessPercentage == 1 && generationPercentage == 1 {
		message = "The hack was successful, " + author.Username + " stole " + strconv.Itoa(totalMemes) + " dank memes from " + target.Username
		// reset targetUnits collectTime, HackSeed, and HackAttempts
		targetUnits.CollectTime = time.Now()
		targetUnits.HackSeed = 0
		targetUnits.HackAttempts = 0
	} else {
		lossesMessage := processHackingLosses(&authorUnits, db)
		message = "`hacking was not successful! hacking report:`"
		message = message + "\r `hackers performed at: " + Ftoa(fitnessPercentage*100) + "%`"
		message = message + "\r `botnets performed at: " + Ftoa(generationPercentage*100) + "%`\r" + lossesMessage
	}
	// update the target's hacked count and possibly CollectTime
	targetUnits.HackAttempts = targetUnits.HackAttempts + 1
	UpdateUnits(&targetUnits, db)
	_, _ = s.ChannelMessageSend(m.ChannelID, message)
	return
}
