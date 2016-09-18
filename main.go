package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	_ "database/sql"
	_ "strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"

	_ "github.com/bmizerany/pq"
)

// Variables used for command line parameters
var (
	Token string
	Email string
	PW    string
	BotID string
	db    *sqlx.DB
)

type User struct {
	ID        string `db:"id"`
	DID       string `db:"discord_id"`
	Username  string `db:"name"`
	CurMoney  int    `db:"current_money"`
	TotMoney  int    `db:"total_money"`
	WonMoney  int    `db:"won_money"`
	LostMoney int    `db:"lost_money"`
	GiveMoney int    `db:"given_money"`
	RecMoney  int    `db:"received_money"`
}

func db_get() *sqlx.DB {
	db, err := sqlx.Connect("postgres", "host=localhost user=memebot dbname=money password=password sslmode=disable parseTime=true")
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func init() {
	Token, _ = os.LookupEnv("bot_token")
	Email, _ = os.LookupEnv("email")
	PW, _ = os.LookupEnv("pw")
	db = db_get()
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
	<-make(chan struct{})
	return
}

func handleTip(s *discordgo.Session, m *discordgo.MessageCreate) {
	args := strings.Split(m.Content, " ")
	if len(args) > 3 {
		amount := args[1]
		from := m.Author
		var users []User
		for _, to := range m.Mentions {
			err := db.Select(&users, `SELECT id, discord_id, name, current_money, total_money, won_money, lost_money, given_money, received_money FROM money WHERE discord_id = $1`, to.ID)
			if err != nil {
				log.Fatal(err)
			}
			err = db.Select(&users, `SELECT id, discord_id, name, current_money, total_money, won_money, lost_money, given_money, received_money FROM money WHERE discord_id = $1`, m.Author.ID)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(len(users))
			_, _ = s.ChannelMessageSend(m.ChannelID, "tip "+amount+" dankmemes to "+to.Username+" from: "+from.Username)

		}
	} else {
		return
	}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == BotID {
		return
	}

	if strings.Contains(m.Content, "!tip") == true {
		handleTip(s, m)
	}

	if m.Content == "meme" {
		_, _ = s.ChannelMessageSend(m.ChannelID, "you're a memestar harry")
	}
}
