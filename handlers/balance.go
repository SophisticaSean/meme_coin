package handlers

import (
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"
	"github.com/sophisticasean/meme_coin/dbHandler"
)

func Balance(s *discordgo.Session, m *discordgo.MessageCreate, db *sqlx.DB) {
	args := strings.Split(m.Content, " ")
	if len(args) == 1 {
		author := dbHandler.UserGet(m.Author, db)
		_, _ = s.ChannelMessageSend(m.ChannelID, "total balance is: "+strconv.Itoa(author.CurMoney))
	}
}
