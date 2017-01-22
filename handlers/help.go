package handlers

import "github.com/SophisticaSean/meme_coin/interaction"

// Help is a wrapper for returning a help message showing all commands we know
func Help(s interaction.Session, m *interaction.MessageCreate) {
	message := help()
	s.ChannelMessageSend(m.ChannelID, message)
	return
}

func help() string {
	message := "yo, whaddup. Here are the commands I know:\r"
	message = message + "`!military` `!hack` `!buy` `!mine` `!units` `!collect` `!gamble` `!tip` `!balance` `!memes` `!memehelp` `!prestige` `!fakecollect` `!check` `!invite`\r"
	message = message + "report problems, or ask questions to MEMETHAN DANKBERRY with the kermit picture"
	return message
}
