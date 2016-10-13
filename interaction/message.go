package interaction

import "github.com/bwmarrin/discordgo"

type MessageCreate struct {
	*discordgo.Message
}

func NewMessage() MessageCreate {
	message := discordgo.Message{}
	return MessageCreate{Message: &message}
}
