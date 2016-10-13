package handlers

import "github.com/SophisticaSean/meme_coin/interaction"

func Help(s interaction.Session, m *interaction.MessageCreate) {
	message := help()
	_, _ = s.ChannelMessageSend(m.ChannelID, message)
	return
}

func help() string {
	message := "yo, whaddup. Here are the commands I know:\r"
	message = message + "`!military` `!hack` `!buy` `!mine` `!units` `!collect` `!gamble` `!tip` `!balance` `!memes` `!memehelp`"
	return message
}
