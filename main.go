package main

import (
	"fmt"
	"os"
	"strings"

	_ "database/sql"
	_ "strings"

	"github.com/sophisticasean/meme_coin/dbHandler"
	"github.com/sophisticasean/meme_coin/handlers"

	_ "github.com/bmizerany/pq"
	"github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"
)

// Import vars and init vars
var (
	Token        string
	Email        string
	PW           string
	BotID        string
	db           *sqlx.DB
	responseList []handlers.MineResponse
)

func init() {
	Token, _ = os.LookupEnv("bot_token")
	Email, _ = os.LookupEnv("email")
	PW, _ = os.LookupEnv("pw")
	db = dbHandler.DbGet()
	responseList = handlers.GenerateResponseList()
}

func main() {
	// Create a new Discord session using the provided login information.
	dg, err := discordgo.New(Email, PW)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	u, err := dg.User("@me")
	if err != nil {
		fmt.Println("error obtaining account details,", err)
	}

	BotID = u.ID

	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection:", err)
		return
	}

	fmt.Println("Bot is now running!")
	// do some busy work indefinitely
	<-make(chan struct{})
	return
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == BotID {
		return
	}

	if strings.Contains(m.Content, "!tip") == true {
		handlers.Tip(s, m, db)
	}

	if m.Content == "!balance" || m.Content == "!memes" {
		handlers.Balance(s, m, db)
	}

	if strings.Contains(m.Content, "!gamble") {
		handlers.Gamble(s, m, db)
	}

	if m.Content == "!mine" {
		handlers.Mine(s, m, responseList, db)
	}

	if m.Content == "meme" {
		_, _ = s.ChannelMessageSend(m.ChannelID, "you're a memestar harry")
	}
}
