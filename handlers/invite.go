package handlers

import "github.com/SophisticaSean/meme_coin/interaction"

// Invite is a wrapper for returning a message showing the invite link for the bot
func Invite(s interaction.Session, m *interaction.MessageCreate) {
	message := "Here's the invite link: https://discordapp.com/oauth2/authorize?&client_id=226569059385212929&scope=bot&permissions=0"
	s.ChannelMessageSend(m.ChannelID, message)
	return
}
