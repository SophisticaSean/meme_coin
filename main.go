package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	_ "database/sql"
	_ "strings"

	"github.com/sophisticasean/meme_coin/events"

	_ "github.com/bmizerany/pq"
	"github.com/bwmarrin/discordgo"
)

// Import vars and init vars
var (
	Token string
	Email string
	PW    string
)

func init() {
	Token, _ = os.LookupEnv("bot_token")
	Email, _ = os.LookupEnv("email")
	PW, _ = os.LookupEnv("pw")
}

func main() {
	// Create a new Discord session using the provided login information.
	dg, err := discordgo.New(Email, PW)
	// generate a new rand seed, else all results will be deterministic
	rand.Seed(time.Now().UnixNano())
	fmt.Println(rand.Int())
	fmt.Println(rand.Int())
	fmt.Println(rand.Int())
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	u, err := dg.User("@me")
	if err != nil {
		fmt.Println("error obtaining account details,", err)
	}

	// set var BotID so events know the ID of the bot
	os.Setenv("BotID", u.ID)

	dg.AddHandler(events.MessageCreate)

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
