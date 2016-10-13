package interaction

import "github.com/bwmarrin/discordgo"

type Message struct {
	*discordgo.Message
}

type MessageCreate struct {
	*discordgo.MessageCreate
}

func NewMessage() Message {
	message := Message{
		Message: &discordgo.Message{},
	}
	return message
}

func NewMessageEvent() MessageCreate {
	message := discordgo.MessageCreate{Message: NewMessage().Message}
	return MessageCreate{MessageCreate: &message}
}
