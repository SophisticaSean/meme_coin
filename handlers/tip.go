package handlers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/SophisticaSean/meme_coin/interaction"
	"github.com/dustin/go-humanize"
	"github.com/jmoiron/sqlx"
)

// Tip is the function that handles the act of giving another player memes
func Tip(s interaction.Session, m *interaction.MessageCreate, db *sqlx.DB) {
	args := strings.Split(m.Content, " ")
	if len(args) >= 3 && strings.ToLower(args[0]) == "!tip" {
		amountInContent := []string{}
		amount := "-1"
		currencyName := "super dank memes"

		amountRegex := regexp.MustCompile(` \d+`)
		tipRegex := regexp.MustCompile(`^!\w* `)
		nameRegex := regexp.MustCompile(`@!*\w+`)
		carrotRegex := regexp.MustCompile(`<|>`)
		spaceReplaceRegex := regexp.MustCompile(` `)
		symbolRegex := regexp.MustCompile(`\W+`)
		twelveRegex := regexp.MustCompile(`12ww12ww12`)
		spaceRegex := regexp.MustCompile(`^ *| *$`)

		// find amount via some regex
		fmt.Println(m.Content)
		amountInContent = amountRegex.FindAllString(m.Content, -1)
		if len(amountInContent) > 0 {
			amount = spaceRegex.ReplaceAllString(amountInContent[0], "")
		}

		// bunch of regex replacement to support all types of currencies
		processedContent := amountRegex.ReplaceAllString(m.Content, "")
		processedContent = tipRegex.ReplaceAllString(processedContent, "")
		processedContent = nameRegex.ReplaceAllString(processedContent, "")
		processedContent = carrotRegex.ReplaceAllString(processedContent, "")
		processedContent = spaceReplaceRegex.ReplaceAllString(processedContent, "12ww12ww12")
		processedContent = symbolRegex.ReplaceAllString(processedContent, "")
		processedContent = twelveRegex.ReplaceAllString(processedContent, " ")
		processedContent = spaceRegex.ReplaceAllString(processedContent, "")

		if len(processedContent) > 0 {
			currencyName = processedContent
		}

		intAmount, err := strconv.Atoi(amount)
		if err != nil {
			fmt.Println(err)
			_, _ = s.ChannelMessageSend(m.ChannelID, "amount is too large or not a number, try again.")
			return
		}

		if intAmount <= 0 {
			_, _ = s.ChannelMessageSend(m.ChannelID, "amount has to be more than 0")
			return
		}

		totalDeduct := intAmount * len(m.Mentions)
		from := UserGet(m.Author, db)
		if totalDeduct > from.CurMoney {
			_, _ = s.ChannelMessageSend(m.ChannelID, "not enough funds to complete transaction, total: "+humanize.Comma(int64(from.CurMoney))+" needed:"+humanize.Comma(int64(totalDeduct)))
			return
		}
		for _, to := range m.Mentions {
			toUser := UserGet(to, db)
			// check prestige level to prevent prestige tip cheese
			if from.PrestigeLevel >= toUser.PrestigeLevel {
				if (toUser.CurMoney + intAmount) < 1 {
					message := "You're trying to tip too many memes, try tipping less memes."
					s.ChannelMessageSend(m.ChannelID, message)
					return
				}
				MoneyDeduct(&from, intAmount, "tip", db)
				// refresh the touser, handles the tipping self problem
				toUser = UserGet(to, db)
				MoneyAdd(&toUser, intAmount, "tip", db)
				message := from.Username + " gave " + humanize.Comma(int64(intAmount)) + " " + currencyName + " to " + to.Username
				_, _ = s.ChannelMessageSend(m.ChannelID, message)
				fmt.Println(message)
			} else {
				message := from.Username + " tried to give " + humanize.Comma(int64(intAmount)) + " " + currencyName + " to " + to.Username + "; but " + from.Username + "'s prestige level is " + strconv.Itoa(from.PrestigeLevel) + " and " + toUser.Username + "'s prestige level is " + strconv.Itoa(toUser.PrestigeLevel) + ". Memes can not be tipped up prestige levels; only to equal and lower prestige levels."
				_, _ = s.ChannelMessageSend(m.ChannelID, message)
				fmt.Println(message)
			}
		}
		return
	}
	return
}
