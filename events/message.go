package events

import (
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"
	"github.com/sophisticasean/meme_coin/dbHandler"
	"github.com/sophisticasean/meme_coin/handlers"
)

var (
	db           *sqlx.DB
	responseList []handlers.MineResponse
	BotID        string
)

func init() {
	db = dbHandler.DbGet()
	responseList = handlers.GenerateResponseList()
}

func MessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if BotID == "" {
		BotID, _ = os.LookupEnv("BotID")
	}
	if m.Author.ID == BotID {
		return
	}

	lowerMessage := strings.ToLower(m.Content)
	if strings.Contains(lowerMessage, "!tip") {
		handlers.Tip(s, m, db)
	}

	if lowerMessage == "!balance" || lowerMessage == "!memes" {
		handlers.Balance(s, m, db)
	}

	if strings.Contains(lowerMessage, "!gamble") {
		handlers.Gamble(s, m, db)
	}

	if lowerMessage == "!mine" {
		handlers.Mine(s, m, responseList, db)
	}

	if strings.Contains(lowerMessage, "!buy") {
		handlers.Buy(s, m, db)
	}

	if lowerMessage == "!units" {
		handlers.UnitInfo(s, m, db)
	}

	if lowerMessage == "!collect" {
		handlers.Collect(s, m, db)
	}

	if lowerMessage == "!memehelp" {
		handlers.Help(s, m)
	}

	if lowerMessage == "meme" {
		_, _ = s.ChannelMessageSend(m.ChannelID, "you're a dank maymay-er, harry")
	}
}
