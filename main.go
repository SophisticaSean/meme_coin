package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"time"

	_ "database/sql"
	"strings"

	"github.com/SophisticaSean/meme_coin/events"
	"github.com/bwmarrin/discordgo"

	"github.com/SophisticaSean/meme_coin/interaction"
	_ "github.com/bmizerany/pq"
)

// Import vars and init vars
var (
	Console string
	Token   string
	Email   string
	PW      string
)

func init() {
	Console, _ = os.LookupEnv("console")
	Token, _ = os.LookupEnv("bot_token")
	Email, _ = os.LookupEnv("email")
	PW, _ = os.LookupEnv("pw")
}

func main() {
	var botSess interaction.Session
	fmt.Println(Console)
	if Console != "" {
		botSess = interaction.NewConsoleSession()
	} else {
		var err error
		botSess, err = interaction.NewDiscordSession(Email, PW)
		if err != nil {
			fmt.Println("Unable to create Discord session with given Email and Password,", err)
			return
		}
	}
	// generate a new rand seed, else all results will be deterministic
	rand.Seed(time.Now().UnixNano())

	u, err := botSess.User("@me")
	if err != nil {
		fmt.Println("error obtaining account details,", err)
	}

	// set var BotID so events know the ID of the bot
	os.Setenv("BotID", u.GetID())

	botSess.AddHandler(events.DiscordMessageHandler)

	err = botSess.Open()
	if err != nil {
		fmt.Println("error opening connection:", err)
		return
	}

	fmt.Println("Bot is now running!")
	reader := bufio.NewReader(os.Stdin)
	// TODO: parse text and pass it into botSess as an MessageCreate event so our handlers can handle it and respond in kind
	//var message *interaction.MessageCreate
	message := interaction.NewMessage()
	author := discordgo.User{
		ID:       "1",
		Username: "admin",
	}
	message.Author = &author
	for {
		text, _ := reader.ReadString('\n')
		if text != "" {
			message.Content = strings.TrimSpace(text)
			events.DiscordMessageHandler(botSess, &message)
		}
		time.Sleep(100 * time.Millisecond)
	}
	// do some busy work indefinitely
	//<-make(chan struct{})
	return
}
