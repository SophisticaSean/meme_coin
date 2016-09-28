package handlers

import "github.com/bwmarrin/discordgo"

func Help(s *discordgo.Session, m *discordgo.MessageCreate) {
	message := "yo, whaddup. Here are the commands I know:\r"
	message = message + "`!buy` `!mine` `!units` `!collect` `!gamble` `!tip` `!balance` `!memes` `!memehelp`"
	_, _ = s.ChannelMessageSend(m.ChannelID, message)
	return
}
