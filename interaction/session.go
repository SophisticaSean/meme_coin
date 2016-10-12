package interaction

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type Session interface {
	ChannelMessageSend(string, string) (string, error)
	AddHandler(event interface{}) func()
	User(userID string) (User, error)
	Open() error
}

type DiscordSession struct {
	*discordgo.Session
}

type ConsoleSession struct{}

func (ds *DiscordSession) ChannelMessageSend(id string, message string) (string, error) {
	msg, err := ds.Session.ChannelMessageSend(id, message)
	return msg.Content, err
}

func (ds *DiscordSession) AddHandler(event interface{}) func() {
	return ds.Session.AddHandler(event)
}

func (ds *DiscordSession) User(userID string) (User, error) {
	dsUsr, err := ds.Session.User(userID)
	if err != nil {
		return nil, err
	}
	return NewDiscordUser(dsUsr), nil
}

func (ds *DiscordSession) Open() error {
	return ds.Session.Open()
}

func NewDiscordSession(email string, pass string) (*DiscordSession, error) {
	s, e := discordgo.New(email, pass)
	if e != nil {
		return nil, e
	}
	return &DiscordSession{
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

func NewConsoleSession() *ConsoleSession {
	return &ConsoleSession{}
}
