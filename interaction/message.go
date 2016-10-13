package interaction

import "github.com/bwmarrin/discordgo"

type MessageCreate struct {
	*discordgo.MessageCreate
}

func NewMessage() MessageCreate {
	message := discordgo.MessageCreate{}
	return MessageCreate{MessageCreate: &message}
}
