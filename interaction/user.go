package interaction

import (
	"github.com/bwmarrin/discordgo"
)

type User interface {
	GetID() string
}

type ConsoleUser struct {
	ID string
}

func (cu *ConsoleUser) GetID() string {
	return cu.ID
}

func NewConsoleUser(userID string) *ConsoleUser {
	return &ConsoleUser{
		ID: userID,
	}
}

type DiscordUser struct {
	*discordgo.User
}

func (du *DiscordUser) GetID() string {
	return du.User.ID
}

func NewDiscordUser(dsUsr *discordgo.User) *DiscordUser {
	return &DiscordUser{
		User: dsUsr,
	}
}
