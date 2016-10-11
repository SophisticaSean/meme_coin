package events

import (
	"os"
	"strings"

	"github.com/SophisticaSean/meme_coin/handlers"
	"github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"
)

var (
	db           *sqlx.DB
	responseList []handlers.MineResponse
	BotID        string
)

func init() {
	db = handlers.DbGet()
	responseList = handlers.GenerateResponseList()
}

func DiscordMessageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	lowerMessage := strings.ToLower(m.Content)

	if BotID == "" {
		BotID, _ = os.LookupEnv("BotID")
	}

	if m.Author.ID == BotID {
		if strings.Contains(lowerMessage, "!reset") {
			handlers.Reset(s, m, db)
		}
		if strings.Contains(lowerMessage, "!ban") {
			handlers.TempBan(s, m, db)
		}
		return
	}

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

	if lowerMessage == "!military" {
		handlers.MilitaryUnitInfo(s, m, db)
	}

	if lowerMessage == "!collect" {
		handlers.Collect(s, m, db)
	}

	if strings.Contains(lowerMessage, "!hack") {
		handlers.Hack(s, m, db)
	}

	if lowerMessage == "!memehelp" {
		handlers.Help(s, m)
	}

	if lowerMessage == "meme" {
		_, _ = s.ChannelMessageSend(m.ChannelID, "you're a dank maymay-er, harry")
	}
}
