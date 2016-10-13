package interaction

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type Session interface {
	ChannelMessageSend(string, string) (string, error)
	AddHandler(interface{}) func()
	User(string) (User, error)
	Open() error
	Channel(string) (discordgo.Channel, error)
}

type DiscordSession struct {
	*discordgo.Session
}

type ConsoleSession struct {
	Session string
}

func (ds DiscordSession) ChannelMessageSend(id string, message string) (string, error) {
	msg, err := ds.Session.ChannelMessageSend(id, message)
	return msg.Content, err
}

func (ds DiscordSession) AddHandler(event interface{}) func() {
	return ds.Session.AddHandler(event)
}

func (ds DiscordSession) User(userID string) (User, error) {
	dsUsr, err := ds.Session.User(userID)
	if err != nil {
		return nil, err
	}
	return NewDiscordUser(dsUsr), nil
}

func (ds DiscordSession) Open() error {
	return ds.Session.Open()
}

func (ds DiscordSession) Channel(channelID string) (discordgo.Channel, error) {
	channel, err := ds.Channel(channelID)
	return channel, err
}

func NewDiscordSession(email string, pass string) (DiscordSession, error) {
	s, e := discordgo.New(email, pass)
	if e != nil {
		return DiscordSession{}, e
	}
	return DiscordSession{
		Session: s,
	}, nil
}

func (cs *ConsoleSession) ChannelMessageSend(id string, message string) (string, error) {
	msg := fmt.Sprintf("Channel: %s\nSent: %s", id, message)
	fmt.Println(msg)
	return msg, nil
}

func (cs *ConsoleSession) AddHandler(event interface{}) func() {
	return nil
	//ds.Session.AddHandler(event)
}

func (cs *ConsoleSession) User(userID string) (User, error) {
	return NewConsoleUser(userID), nil
}

func (cs *ConsoleSession) Open() error {
	return nil
}

func (ds *ConsoleSession) Channel(channelID string) (discordgo.Channel, error) {
	return discordgo.Channel{ID: channelID}, nil
}

func NewConsoleSession() *ConsoleSession {
	return &ConsoleSession{}
}
