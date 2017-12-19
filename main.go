package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"time"

	_ "database/sql"
	"strings"

	"github.com/SophisticaSean/meme_coin/api"
	"github.com/SophisticaSean/meme_coin/events"
	"github.com/SophisticaSean/meme_coin/interaction"
	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"

	_ "github.com/bmizerany/pq"
)

// Import vars and init vars
var (
	Console  string
	Token    string
	Email    string
	Password string
)

func init() {
	Console, _ = os.LookupEnv("console")
	Token, _ = os.LookupEnv("bot_token")
	Email, _ = os.LookupEnv("email")
	Password, _ = os.LookupEnv("password")
}

func main() {
	// more sample text
	var botSess interaction.Session
	if Console != "" && Console != "false" {
		botSess = interaction.NewConsoleSession()
	} else if Token != "" {
		gin.SetMode(gin.ReleaseMode)
		var err error
		botSess, err = interaction.NewDiscordSessionToken("Bot " + Token)
		if err != nil {
			fmt.Println("Unable to create Discord session with given token,", err)
			return
		}
	} else if Email != "" && Password != "" {
		gin.SetMode(gin.ReleaseMode)
		var err error
		botSess, err = interaction.NewDiscordSession(Email, Password)
		if err != nil {
			fmt.Println("Unable to create Discord session with given email and password,", err)
			return
		}
	} else {
		fmt.Println("Please provide either a bot token or an email and password")
		return
	}

	// generate a new rand seed, else all results will be deterministic
	rand.Seed(time.Now().UnixNano())

	u, err := botSess.User("@me")
	if err != nil {
		fmt.Println("Error obtaining account details,", err)
	}

	// set var BotID so events know the ID of the bot
	if Console != "" && Console != "false" {
		botSess.AddHandler(events.MessageHandler)
		os.Setenv("BotID", "1")
	} else {
		botSess.AddHandler(events.DiscordMessageHandler)
		os.Setenv("BotID", u.GetID())
	}

	err = botSess.Open()
	if err != nil {
		fmt.Println("Error opening connection:", err)
		return
	}

	db, router := api.RouterConfigure()
	defer db.Close()
	go router.Run(":8080")

	fmt.Println("Bot is now running!")
	if Console != "" && Console != "false" {
		reader := bufio.NewReader(os.Stdin)
		// TODO: parse text and pass it into botSess as a MessageCreate event so our handlers can handle it and respond in kind
		//var message *interaction.MessageCreate
		message := interaction.NewMessageEvent()
		author := discordgo.User{
			ID:       "2",
			Username: "admin",
		}
		message.Message.Author = &author
		for {
			line, _, _ := reader.ReadLine()
			text := string(line)
			if text != "" {
				message.Message.Content = strings.TrimSpace(text)
				events.MessageHandler(botSess, &message)
			}
			time.Sleep(100 * time.Millisecond)
		}
	} else {
		//do some busy work indefinitely
		<-make(chan struct{})
	}
	return
}
