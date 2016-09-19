package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
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
	ID        int    `db:"id"`
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

func createUser(user *discordgo.User) {
	var new_user User
	new_user.DID = user.ID
	new_user.Username = user.Username
	_, err := db.NamedExec(`INSERT INTO money (discord_id, name) VALUES (:discord_id, :name)`, new_user)
	if err != nil {
		log.Fatal(err)
	}
}

func userGet(discord_user *discordgo.User) User {
	var users []User
	fmt.Println(discord_user.ID)
	err := db.Select(&users, `SELECT id, discord_id, name, current_money, total_money, won_money, lost_money, given_money, received_money FROM money WHERE discord_id = $1`, discord_user.ID)
	if err != nil {
		log.Fatal(err)
	}
	var user User
	if len(users) == 0 {
		fmt.Println("creating user: " + discord_user.ID)
		createUser(discord_user)
		user = userGet(discord_user)
	} else {
		user = users[0]
	}
	return user
}

func moneyDeduct(user *User, amount int, deduction string) bool {
	if amount <= user.CurMoney {
		new_current_money := user.CurMoney - amount
		new_deduction_amount := -1
		db_string := ``
		deduction_record := -1

		if deduction == "tip" {
			db_string = `UPDATE money SET (current_money, given_money) = (?, ?) WHERE discord_id = '?'`
			deduction_record = user.GiveMoney
			new_deduction_amount = user.GiveMoney + amount
			user.CurMoney = new_current_money
			user.GiveMoney = new_deduction_amount
		}
		if deduction == "gamble" {
			db_string = `UPDATE money SET (current_money, lost_money) = (?, ?) WHERE discord_id = '?'`
			deduction_record = user.LostMoney
			new_deduction_amount = user.LostMoney + amount
			user.CurMoney = new_current_money
			user.LostMoney = new_deduction_amount
		}

		if db_string != `` && deduction_record != -1 && new_deduction_amount != -1 {
			db.MustExec(db_string, new_current_money, new_deduction_amount, user.DID)
			return true
		}
		return false
	} else {
		return false
	}
}

func moneyAdd(user *User, amount int, addition string) {
	new_current_money := user.CurMoney - amount
	new_addition_amount := -1
	db_string := ``
	addition_record := -1

	if addition == "tip" {
		db_string = `UPDATE money SET (current_money, received_money) = (?, ?) WHERE discord_id = '?'`
		addition_record = user.RecMoney
		new_addition_amount = user.RecMoney + amount
		user.CurMoney = new_current_money
		user.RecMoney = new_addition_amount
	}
	if addition == "gamble" {
		db_string = `UPDATE money SET (current_money, won_money) = (?, ?) WHERE discord_id = '?'`
		addition_record = user.WonMoney
		new_addition_amount = user.WonMoney + amount
		user.CurMoney = new_current_money
		user.WonMoney = new_addition_amount
	}

	if db_string != `` && addition_record != -1 && new_addition_amount != -1 {
		db.MustExec(db_string, new_current_money, new_addition_amount, user.DID)
	}
}

func handleTip(s *discordgo.Session, m *discordgo.MessageCreate) {
	args := strings.Split(m.Content, " ")
	if len(args) > 3 && args[0] == "!tip" {
		int_amount, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Println(err)
		}
		amount := args[1]
		total_deduct := int_amount * len(m.Mentions)
		from := userGet(m.Author)
		proceed := moneyDeduct(&from, total_deduct, "tip")
		if proceed != true {
			_, _ = s.ChannelMessageSend(m.ChannelID, "not enough funds to complete transaction, total: "+strconv.Itoa(from.CurMoney)+" needed:"+strconv.Itoa(total_deduct))
			return
		} else {
			for _, to := range m.Mentions {
				toUser := userGet(to)
				_, _ = s.ChannelMessageSend(m.ChannelID, "tip "+amount+" dankmemes to "+toUser.Username+" from: "+from.Username)

			}
		}
	} else {
		return
	}
}

func handleBalance(s *discordgo.Session, m *discordgo.MessageCreate) {
	args := strings.Split(m.Content, " ")
	if len(args) == 1 {
		author := userGet(m.Author)
		_, _ = s.ChannelMessageSend(m.ChannelID, "total balance is :"+strconv.Itoa(author.CurMoney))
	}
}

func handleGamble(s *discordgo.Session, m *discordgo.MessageCreate) {
	args := strings.Split(m.Content, " ")
	if len(args) == 4 {
		//author := userGet(m.Author)
		//bet, err := strconv.Atoi(args[1])
		//game := args[2]
		//game_input := args[3]
	} else if args[0] == "!gamble" {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Gamble command is used as follows: '!gamble <amount> <game> <game_input>")
	}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == BotID {
		return
	}

	if strings.Contains(m.Content, "!tip") == true {
		handleTip(s, m)
	}

	if strings.Contains(m.Content, "!balance") || strings.Contains(m.Content, "!memes") {
		handleBalance(s, m)
	}

	if strings.Contains(m.Content, "!gamble") {
		handleGamble(s, m)
	}

	if m.Content == "meme" {
		_, _ = s.ChannelMessageSend(m.ChannelID, "you're a memestar harry")
	}
}
