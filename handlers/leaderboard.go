package handlers

import "github.com/SophisticaSean/meme_coin/interaction"

// Leaderboard returns the link to the publicly available leaderboard for this bot.
func Leaderboard(s interaction.Session, m *interaction.MessageCreate) {
	message := "Here's the leaderboard link: https://sophisticasean.github.io/meme_dashboard/"
	s.ChannelMessageSend(m.ChannelID, message)
	return
}
