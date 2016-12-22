package handlers

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/SophisticaSean/meme_coin/interaction"
	humanize "github.com/dustin/go-humanize"
	"github.com/jmoiron/sqlx"
)

var helpMessage = "The goal of hacking is to get your hacker's performance to 100 percent and then your botnet performance to 100 percent. You're trying to hack someone's password to steal all their uncollected memes. If your hacker's performance is under 100 percent, it means you need to increase the amount of botnets you're using, if your botnet's performance is overperforming, you'll need to decrease the amount of botnets you're using. There is a magic number of botnets and hackers that will crack the target's password successfully. You only have a fixed amount of tries at someone's password before it resets! The more cyphers someone has, the more difficult their password is going to be to crack!\r"

const (
	globalHackerLimit = 7
	botnetLimit       = 5000
	cypherPadding     = 5
)

var (
	hackAttempts int
	lossChances  int
)

// Ftoa is the Float64 equivalent of strconv.Itoa
func Ftoa(float float64) string {
	floatString := strconv.FormatFloat(float, 'f', -1, 64)
	return floatString
}

func processHackingLosses(units *User, usedHackers int, usedBotnets int, seed int64, db *sqlx.DB) string {
	rand.Seed(seed)
	message := ""
	hackerLosses := 0
	botnetLosses := 0
	for i := 0; i <= usedHackers; i++ {
		if rand.Intn(400) < lossChances {
			hackerLosses++
		}
	}
	if hackerLosses != 0 {
		units.Hacker = units.Hacker - hackerLosses
		message = message + "`Your hacking got " + humanize.Comma(int64(hackerLosses)) + " hackers arrested by the FBI!`\r`You now have " + humanize.Comma(int64(units.Hacker)) + " hackers left.`\r"
	}
	for i := 0; i <= usedBotnets; i++ {
		if rand.Intn(100) < lossChances {
			botnetLosses++
		}
	}
	if botnetLosses != 0 {
		units.Botnet = units.Botnet - botnetLosses
		message = message + "`Your hacking was detected and got some botnets discovered, some whitehat released a zero day whitepaper and patch defeating " + humanize.Comma(int64(botnetLosses)) + " of your botnets.`\r`You now have " + humanize.Comma(int64(units.Botnet)) + " botnets left.`\r"
	}
	UpdateUnits(units, db)
	rand.Seed(time.Now().UnixNano())
	return message
}

// Hack handles all the logic for the !hack command
func Hack(s interaction.Session, m *interaction.MessageCreate, db *sqlx.DB) {
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
	// set the vars we care about
	totalMemes, _, target := totalMemesEarned(mentions[0], db)
	strTotalMemes := humanize.Comma(int64(totalMemes))
	lossChances = int(math.Floor(math.Abs(float64(float64((len(strTotalMemes) - 3)) * 6))))
	hackAttempts = int((math.Floor(float64(target.Cypher/75.0) + 4)))
	maxStringLength := target.Cypher + cypherPadding
	maxStringCapped := maxStringLength
	//if maxStringLength > 75 {
	//maxStringLength = 75
	//}
	if maxStringCapped > 80 {
		maxStringCapped = 80
	}
	hackerLimit := globalHackerLimit + int(math.Floor(float64(maxStringCapped)*float64(0.5)))
	author := UserGet(m.Author, db)

	if target.HackSeed == 0 {
		discordID, err := strconv.Atoi(target.DID)
		// shouldn't happen
		if err != nil {
			fmt.Println("SUPER BAD ERROR: ", err)
			return
		}
		target.HackAttempts = 0
		target.HackSeed = (time.Now().UnixNano() + int64(discordID))
		UpdateUnits(&target, db)
	}
	seed := target.HackSeed

	hackerCount, err := strconv.Atoi(args[1])
	if err != nil || hackerCount < 1 {
		message = "your amount of hackers argument (`!hack <this_argument> <amount_of_botnets>`) is too large, too small, or not a number.\r"
	}
	if hackerCount > author.Hacker {
		message = message + "You don't have enough hackers for the requested hack need: " + args[1] + " have: " + humanize.Comma(int64(author.Hacker)) + "\r"
	}
	if hackerCount > hackerLimit {
		hackerCount = hackerLimit
	}

	botnetCount, err := strconv.Atoi(args[2])
	if err != nil || botnetCount < 1 {
		message = message + "your amount of botnets argument (`!hack <amount_of_hackers> <this_argument>`) is too large, too small, or not a number.\r"
	}
	if botnetCount > author.Botnet {
		message = message + "You don't have enough botnets for the requested hack need: " + args[2] + " have: " + humanize.Comma(int64(author.Botnet))
	}
	if botnetCount > botnetLimit {
		botnetCount = botnetLimit
	}
	if message != "" {
		_, _ = s.ChannelMessageSend(m.ChannelID, message)
		return
	}

	// check prestige level to prevent prestige tip cheese
	if author.PrestigeLevel != target.PrestigeLevel {
		message := author.Username + " tried to hack " + target.Username + "; but " + author.Username + "'s prestige level is " + strconv.Itoa(author.PrestigeLevel) + " and " + target.Username + "'s prestige level is " + strconv.Itoa(target.PrestigeLevel) + ". Memes can not be hacked from different prestige levels."
		s.ChannelMessageSend(m.ChannelID, message)
		fmt.Println(message)
		return
	}

	fitnessPercentage, generationPercentage := hackSimulate(seed, hackerCount, botnetCount, maxStringLength)
	// randomize which direction we round in
	roundedGenerationPercentage := 0.0
	if rand.Intn(1) == 1 {
		roundedGenerationPercentage = math.Floor(generationPercentage * 10)
	} else {
		roundedGenerationPercentage = math.Ceil(generationPercentage * 10)
	}
	if fitnessPercentage == 1 && generationPercentage == 1 {
		// reset target collectTime, HackSeed, and HackAttempts
		target.CollectTime = time.Now()
		target.HackSeed = 0
		target.HackAttempts = 0
		totalMemes := PrestigeBonus(totalMemes, &author)
		message = "The hack was successful, " + author.Username + " stole " + humanize.Comma(int64(totalMemes)) + " dank memes from " + target.Username
		MoneyAdd(&author, totalMemes, "hacked", db)
		MoneyDeduct(&target, totalMemes, "hacked", db)
		fmt.Println(message, lossChances)
	} else {
		// update the target's hacked count and possibly CollectTime
		target.HackAttempts = target.HackAttempts + 1
		lossesMessage := ""
		if target.DID != author.DID {
			lossesMessage = processHackingLosses(&author, hackerCount, botnetCount, seed, db)
		}
		message = "`" + author.Username + " is trying to hack " + target.Username + "!\rhacking report:`"
		message = message + "\r`hackers performed at: " + Ftoa(fitnessPercentage*100) + "%`\r"
		if fitnessPercentage == 1 {
			message = message + "`botnets overperformed at: ~" + Ftoa(roundedGenerationPercentage*10) + "%`\r"
		}
		message = message + lossesMessage
		// handle hackAttempts limit reached
		if target.HackAttempts >= hackAttempts {
			discordID, err := strconv.Atoi(target.DID)
			// shouldn't happen
			if err != nil {
				fmt.Println("SUPER BAD ERROR: ", err)
				return
			}
			target.HackAttempts = 0
			target.HackSeed = (time.Now().UnixNano() + int64(discordID))
			UpdateUnits(&target, db)
			message = message + "`Your hacking attempts were detected! The target's password has been reset!`"
		}
	}
	UpdateUnits(&target, db)
	_, _ = s.ChannelMessageSend(m.ChannelID, message)
	return
}
