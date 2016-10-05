package handlers

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"
)

var helpMessage = "The goal of hacking is to get your hacker's performance to 100 percent and then your botnet performance to 100 percent. You're trying to hack someone's password to steal all their uncollected memes. If your hacker's performance is under 100 percent, it means you need to increase the amount of botnets you're using, if your botnet's performance is overperforming, you'll need to decrease the amount of botnets you're using. There is a magic number of botnets and hackers that will crack the target's password successfully. You only have a fixed amount of tries at someone's password before it resets! The more cyphers someone has, the more difficult their password is going to be to crack!\r"

const (
	hackAttempts  = 4
	hackerLimit   = 10
	botnetLimit   = 5000
	cypherPadding = 5
)

var (
	lossChances int
)

func Ftoa(float float64) string {
	floatString := strconv.FormatFloat(float, 'f', -1, 64)
	return floatString
}

func processHackingLosses(units *UserUnits, usedHackers int, usedBotnets int, db *sqlx.DB) string {
	message := ""
	hackerLosses := 0
	botnetLosses := 0
	for i := 0; i <= usedHackers; i++ {
		if rand.Intn(400) < lossChances {
			hackerLosses += 1
		}
	}
	if hackerLosses != 0 {
		units.Hacker = units.Hacker - hackerLosses
		message = message + "`Your hacking got " + strconv.Itoa(hackerLosses) + " hackers arrested by the FBI!`\r`You now have " + strconv.Itoa(units.Hacker) + " hackers left.`\r"
	}
	for i := 0; i <= usedBotnets; i++ {
		if rand.Intn(100) < lossChances {
			botnetLosses += 1
		}
	}
	if botnetLosses != 0 {
		units.Botnet = units.Botnet - botnetLosses
		message = message + "`Your hacking was detected and got some botnets discovered, some whitehat released a zero day whitepaper and patch defeating " + strconv.Itoa(botnetLosses) + " of your botnets.`\r`You now have " + strconv.Itoa(units.Botnet) + " botnets left.`\r"
	}
	UpdateUnits(units, db)
	return message
}

func Hack(s *discordgo.Session, m *discordgo.MessageCreate, db *sqlx.DB) {
	args := strings.Split(m.Content, " ")
	defaultSyntax := "`!hack <amount_of_hackers> <amount_of_botnets> @person`\r`!hack 3 12 @some_body`"
	message := ""
	if len(args) == 1 {
		_, _ = s.ChannelMessageSend(m.ChannelID, helpMessage+defaultSyntax)
		return
	}
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
	strTotalMemes := strconv.Itoa(totalMemes)
	lossChances = int(math.Abs(float64((len(strTotalMemes) - 3) * 2)))
	target := UserGet(mentions[0], db)
	authorUnits := UnitsGet(m.Author, db)
	author := UserGet(m.Author, db)

	if targetUnits.HackSeed == 0 {
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

	hackerCount, err := strconv.Atoi(args[1])
	if err != nil || hackerCount < 1 {
		message = "your amount of hackers argument (`!hack <this_argument> <amount_of_botnets>`) is too large, too small, or not a number.\r"
	}
	if hackerCount > authorUnits.Hacker {
		message = message + "You don't have enough hackers for the requested hack need: " + args[1] + " have: " + strconv.Itoa(authorUnits.Hacker) + "\r"
	}
	if hackerCount > hackerLimit {
		hackerCount = hackerLimit
	}

	botnetCount, err := strconv.Atoi(args[2])
	if err != nil || botnetCount < 1 {
		message = message + "your amount of botnets argument (`!hack <amount_of_hackers> <this_argument>`) is too large, too small, or not a number.\r"
	}
	if botnetCount > authorUnits.Botnet {
		message = message + "You don't have enough botnets for the requested hack need: " + args[2] + " have: " + strconv.Itoa(authorUnits.Botnet)
	}
	if botnetCount > botnetLimit {
		botnetCount = botnetLimit
	}
	if message != "" {
		_, _ = s.ChannelMessageSend(m.ChannelID, message)
		return
	}

	maxStringLength := targetUnits.Cypher + cypherPadding
	fitnessPercentage, generationPercentage := hackSimulate(seed, hackerCount, botnetCount, maxStringLength)
	// randomize which direction we round in
	roundedGenerationPercentage := 0.0
	if rand.Intn(1) == 1 {
		roundedGenerationPercentage = math.Floor(generationPercentage * 10)
	} else {
		roundedGenerationPercentage = math.Ceil(generationPercentage * 10)
	}
	if fitnessPercentage == 1 && generationPercentage == 1 {
		message = "The hack was successful, " + author.Username + " stole " + strconv.Itoa(totalMemes) + " dank memes from " + target.Username
		// reset targetUnits collectTime, HackSeed, and HackAttempts
		targetUnits.CollectTime = time.Now()
		targetUnits.HackSeed = 0
		targetUnits.HackAttempts = 0
		UpdateUnitsTimestamp(&targetUnits, db)
		MoneyAdd(&author, totalMemes, "hacked", db)
		MoneyDeduct(&target, totalMemes, "hacked", db)
		fmt.Println(message, lossChances)
	} else {
		// update the target's hacked count and possibly CollectTime
		targetUnits.HackAttempts = targetUnits.HackAttempts + 1
		lossesMessage := ""
		if target.DID != author.DID {
			lossesMessage = processHackingLosses(&authorUnits, hackerCount, botnetCount, db)
		}
		message = "`hacking was not successful! hacking report:`"
		message = message + "\r `hackers performed at: " + Ftoa(fitnessPercentage*100) + "%`\r"
		if fitnessPercentage == 1 {
			message = message + "`botnets overperformed at: ~" + Ftoa(roundedGenerationPercentage*10) + "%`\r"
		}
		message = message + lossesMessage
		// handle hackAttempts limit reached
		if targetUnits.HackAttempts >= hackAttempts {
			discordID, err := strconv.Atoi(targetUnits.DID)
			// shouldn't happen
			if err != nil {
				fmt.Println("SUPER BAD ERROR: ", err)
				return
			}
			targetUnits.HackAttempts = 0
			targetUnits.HackSeed = (time.Now().UnixNano() + int64(discordID))
			UpdateUnits(&targetUnits, db)
			message = message + "`Your hacking attempts were detected! The target's password has been reset!`"
		}
	}
	UpdateUnits(&targetUnits, db)
	_, _ = s.ChannelMessageSend(m.ChannelID, message)
	return
}
