package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"
	"github.com/sophisticasean/meme_coin/dbHandler"
)

func Tip(s *discordgo.Session, m *discordgo.MessageCreate, db *sqlx.DB) {
	args := strings.Split(m.Content, " ")
	if len(args) >= 3 && args[0] == "!tip" {
		intAmount, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Println(err)
			_, _ = s.ChannelMessageSend(m.ChannelID, "amount is too large or not a number, try again.")
			return
		}
		if intAmount <= 0 {
			_, _ = s.ChannelMessageSend(m.ChannelID, "amount has to be more than 0")
			return
		}
		amount := args[1]
		currencyName := "super dank memes"
		if len(args) > 3 {
			currencyName = args[2]
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
