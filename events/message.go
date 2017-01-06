package events

import (
	"os"
	"strings"

	"github.com/SophisticaSean/meme_coin/handlers"
	"github.com/SophisticaSean/meme_coin/interaction"
	"github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"
)

var (
	db           *sqlx.DB
	responseList []handlers.MineResponse
	botID        string
	adminID      string
)

func validateSingleArg(validator string, validatee string) bool {
	newValidateeSlice := strings.Split(validatee, " ")
	newValidatee := ""
	if len(newValidateeSlice) > 0 {
		newValidatee = newValidateeSlice[0]
		newValidatee = strings.TrimSpace(newValidatee)
	} else {
		return false
	}
	return validator == newValidatee
}

func init() {
	db = handlers.DbGet()
	responseList = handlers.GenerateResponseList()
}

// DiscordMessageHandler is a wrapper for creating a message handler
func DiscordMessageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	sess := interaction.DiscordSession{Session: s}
	messageCreate := &interaction.MessageCreate{MessageCreate: m}
	MessageHandler(sess, messageCreate)
}

// MessageHandler is where we define all the handlers we want
func MessageHandler(s interaction.Session, m *interaction.MessageCreate) {
	lowerMessage := strings.ToLower(m.Content)

	if botID == "" {
		botID, _ = os.LookupEnv("BotID")
	}

	if adminID == "" {
		adminID, _ = os.LookupEnv("AdminID")
	}

	if m.Author.ID == botID || m.Author.ID == adminID {
		if strings.Contains(lowerMessage, "!reset") {
			handlers.Reset(s, m, db)
		}
		if strings.Contains(lowerMessage, "!ban") {
			handlers.TempBan(s, m, db)
		}
		if strings.Contains(lowerMessage, "!unban") {
			handlers.Unban(s, m, db)
		}
		if m.Author.ID == botID {
			return
		}
	}

	if strings.Contains(lowerMessage, "!tip") {
		handlers.Tip(s, m, db)
	}

	if lowerMessage == "!balance" || lowerMessage == "!memes" || lowerMessage == "!maymays" || lowerMessage == "!memez" {
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

	if lowerMessage == "!fakecollect" || lowerMessage == "!check" {
		handlers.FakeCollect(s, m, db)
	}

	if strings.Contains(lowerMessage, "!hack") {
		handlers.Hack(s, m, db)
	}

	if strings.Contains(lowerMessage, "!prestige") {
		handlers.Prestige(s, m, db)
	}

	if lowerMessage == "!help" || lowerMessage == "!memehelp" {
		handlers.Help(s, m)
	}

	if lowerMessage == "!invite" {
		handlers.Invite(s, m)
	}

	if lowerMessage == "meme" {
		_, _ = s.ChannelMessageSend(m.ChannelID, "you're a dank maymay-er, harry")
	}
}
