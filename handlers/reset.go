package handlers

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"
)

func Reset(s *discordgo.Session, m *discordgo.MessageCreate, db *sqlx.DB) {
	message := "yo, whaddup. Here are the commands I know:\r"
	message = message + "`!buy` `!mine` `!units` `!collect` `!gamble` `!tip` `!balance` `!memes` `!memehelp`"
	_, _ = s.ChannelMessageSend(m.ChannelID, message)
	return
}
