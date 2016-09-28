package handlers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"
	"github.com/sophisticasean/meme_coin/dbHandler"
)

func Tip(s *discordgo.Session, m *discordgo.MessageCreate, db *sqlx.DB) {
	args := strings.Split(m.Content, " ")
	if len(args) >= 3 && args[0] == "!tip" {
		amountInContent := []string{}
		amount := "-1"
		currencyName := "super dank memes"

		amountRegex := regexp.MustCompile(` \d+`)
		tipRegex := regexp.MustCompile(`!tip `)
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
		from := dbHandler.UserGet(m.Author, db)
		if totalDeduct > from.CurMoney {
			_, _ = s.ChannelMessageSend(m.ChannelID, "not enough funds to complete transaction, total: "+strconv.Itoa(from.CurMoney)+" needed:"+strconv.Itoa(totalDeduct))
			return
		}
		dbHandler.MoneyDeduct(&from, totalDeduct, "tip", db)
		for _, to := range m.Mentions {
			toUser := dbHandler.UserGet(to, db)
			dbHandler.MoneyAdd(&toUser, intAmount, "tip", db)
			message := from.Username + " gave " + amount + " " + currencyName + " to " + to.Username
			_, _ = s.ChannelMessageSend(m.ChannelID, message)
			fmt.Println(message)
		}
		return
	}
	return
}
