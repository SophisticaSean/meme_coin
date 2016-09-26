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

	if strings.Contains(m.Content, "!tip") == true {
		handlers.Tip(s, m, db)
	}

	if m.Content == "!balance" || m.Content == "!memes" {
		handlers.Balance(s, m, db)
	}

	if strings.Contains(m.Content, "!gamble") {
		handlers.Gamble(s, m, db)
	}

	if m.Content == "!mine" {
		handlers.Mine(s, m, responseList, db)
	}

	if strings.Contains(m.Content, "!buy") {
		handlers.Buy(s, m, db)
	}

	if m.Content == "!units" {
		handlers.UnitInfo(s, m, db)
	}

	if m.Content == "meme" {
		_, _ = s.ChannelMessageSend(m.ChannelID, "you're a dank maymay-er, harry")
	}
}
