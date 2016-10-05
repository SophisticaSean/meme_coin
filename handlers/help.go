package handlers

import "github.com/bwmarrin/discordgo"

func Help(s *discordgo.Session, m *discordgo.MessageCreate) {
	message := help()
	_, _ = s.ChannelMessageSend(m.ChannelID, message)
	return
}

func help() string {
	message := "yo, whaddup. Here are the commands I know:\r"
	message = message + "`!military` `!hack` `!buy` `!mine` `!units` `!collect` `!gamble` `!tip` `!balance` `!memes` `!memehelp`"
	return message
}
