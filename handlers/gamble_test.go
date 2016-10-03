package handlers

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

var db = TestDbGet()
var author = User{
	DID:      "1234",
	Username: "test",
}

func UserGet(*discordgo.User, db) {
	return author
}

func TestGambleProcess(t *testing.T) {
	user := UserGet()
	message := gambleProcess("!gamble lol", &user, db)
	t.Log(message)
}
