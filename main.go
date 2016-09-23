package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

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

// User is a struct that maps 1 to 1 with 'money' db table
type User struct {
	ID        int       `db:"id"`
	DID       string    `db:"discord_id"`
	Username  string    `db:"name"`
	CurMoney  int       `db:"current_money"`
	TotMoney  int       `db:"total_money"`
	WonMoney  int       `db:"won_money"`
	LostMoney int       `db:"lost_money"`
	GiveMoney int       `db:"given_money"`
	RecMoney  int       `db:"received_money"`
	EarMoney  int       `db:"earned_money"`
	MineTime  time.Time `db:"mine_time"`
}

type MineResponse struct {
	amount   int
	response string
	chance   int
}

func dbGet() *sqlx.DB {
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
	db = dbGet()
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
	var newUser User
	newUser.DID = user.ID
	newUser.Username = user.Username
	_, err := db.NamedExec(`INSERT INTO money (discord_id, name) VALUES (:discord_id, :name)`, newUser)
	if err != nil {
		log.Fatal(err)
	}
}

func userGet(discordUser *discordgo.User) User {
	var users []User
	//fmt.Println(discordUser.ID)
	err := db.Select(&users, `SELECT id, discord_id, name, current_money, total_money, won_money, lost_money, given_money, received_money, earned_money, mine_time FROM money WHERE discord_id = $1`, discordUser.ID)
	if err != nil {
		log.Fatal(err)
	}
	var user User
	if len(users) == 0 {
		fmt.Println("creating user: " + discordUser.ID)
		createUser(discordUser)
		user = userGet(discordUser)
	} else {
		user = users[0]
	}
	return user
}

func moneyDeduct(user *User, amount int, deduction string) {
	newCurrentMoney := user.CurMoney - amount
	newDeductionAmount := -1
	dbString := ``
	deductionRecord := -1

	if deduction == "tip" {
		dbString = `UPDATE money SET (current_money, given_money) = ($1, $2) WHERE discord_id = `
		deductionRecord = user.GiveMoney
		newDeductionAmount = user.GiveMoney + amount
		user.CurMoney = newCurrentMoney
		user.GiveMoney = newDeductionAmount
	}
	if deduction == "gamble" {
		dbString = `UPDATE money SET (current_money, lost_money) = ($1, $2) WHERE discord_id = `
		deductionRecord = user.LostMoney
		newDeductionAmount = user.LostMoney + amount
		user.CurMoney = newCurrentMoney
		user.LostMoney = newDeductionAmount
	}

	if dbString != `` && deductionRecord != -1 && newDeductionAmount != -1 {
		dbString = dbString + `'` + user.DID + `'`
		db.MustExec(dbString, newCurrentMoney, newDeductionAmount)
	}
}

func moneyAdd(user *User, amount int, addition string) {
	newCurrentMoney := user.CurMoney + amount
	newAdditionAmount := -1
	dbString := ``
	additionRecord := -1

	if addition == "tip" {
		dbString = `UPDATE money SET (current_money, received_money) = ($1, $2) WHERE discord_id = `
		additionRecord = user.RecMoney
		newAdditionAmount = user.RecMoney + amount
		user.CurMoney = newCurrentMoney
		user.RecMoney = newAdditionAmount
	}
	if addition == "gamble" {
		dbString = `UPDATE money SET (current_money, won_money) = ($1, $2) WHERE discord_id = `
		additionRecord = user.WonMoney
		newAdditionAmount = user.WonMoney + amount
		user.CurMoney = newCurrentMoney
		user.WonMoney = newAdditionAmount
	}
	if addition == "mined" {
		dbString = `UPDATE money SET (current_money, earned_money, mine_time) = ($1, $2, $3) WHERE discord_id = `
		additionRecord = user.EarMoney
		newAdditionAmount = user.EarMoney + amount
		user.CurMoney = newCurrentMoney
		user.EarMoney = newAdditionAmount
		// bindvars can only be used as values so we have to concat the user.DID onto the db string
		dbString = dbString + `'` + user.DID + `'`
		db.MustExec(dbString, newCurrentMoney, newAdditionAmount, time.Now())
	} else {
		if dbString != `` && additionRecord != -1 && newAdditionAmount != -1 {
			// bindvars can only be used as values so we have to concat the user.DID onto the db string
			dbString = dbString + `'` + user.DID + `'`
			db.MustExec(dbString, newCurrentMoney, newAdditionAmount)
		}
	}
}

func handleTip(s *discordgo.Session, m *discordgo.MessageCreate) {
	args := strings.Split(m.Content, " ")
	if len(args) >= 3 && args[0] == "!tip" {
		intAmount, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Println(err)
			_, _ = s.ChannelMessageSend(m.ChannelID, "amount is too large or not a number, try again.")
			return
		}
		if intAmount <= 0 {
			_, _ = s.ChannelMessageSend(m.ChannelID, "amount has to be more than 0")
			return
		}
		amount := args[1]
		currencyName := "super dank memes"
		if len(args) > 3 {
			currencyName = args[2]
		}
		totalDeduct := intAmount * len(m.Mentions)
		from := userGet(m.Author)
		if totalDeduct > from.CurMoney {
			_, _ = s.ChannelMessageSend(m.ChannelID, "not enough funds to complete transaction, total: "+strconv.Itoa(from.CurMoney)+" needed:"+strconv.Itoa(totalDeduct))
			return
		}
		moneyDeduct(&from, totalDeduct, "tip")
		for _, to := range m.Mentions {
			toUser := userGet(to)
			moneyAdd(&toUser, intAmount, "tip")
			message := "you gave " + amount + " " + currencyName + " to " + to.Username
			_, _ = s.ChannelMessageSend(m.ChannelID, message)
			fmt.Println(message + " from " + from.Username)
		}
		return
	} else {
		return
	}
}

func handleBalance(s *discordgo.Session, m *discordgo.MessageCreate) {
	args := strings.Split(m.Content, " ")
	if len(args) == 1 {
		author := userGet(m.Author)
		_, _ = s.ChannelMessageSend(m.ChannelID, "total balance is: "+strconv.Itoa(author.CurMoney))
	}
}

func betToPayout(bet int, payoutMultiplier float64) int {
	payout := int(math.Floor(float64(bet) * payoutMultiplier))
	return payout
}

func handleGamble(s *discordgo.Session, m *discordgo.MessageCreate) {
	args := strings.Split(m.Content, " ")
	if len(args) == 4 {
		author := userGet(m.Author)
		bet, err := strconv.Atoi(args[1])
		if err != nil {
			_, _ = s.ChannelMessageSend(m.ChannelID, "amount is too large or not a number, try again.")
			return
		}
		if bet <= 0 {
			_, _ = s.ChannelMessageSend(m.ChannelID, "amount has to be more than 0")
			return
		}
		game := args[2]
		gameInput := args[3]

		if bet > author.CurMoney {
			_, _ = s.ChannelMessageSend(m.ChannelID, "not enough funds to complete transaction, total: "+strconv.Itoa(author.CurMoney)+" needed:"+strconv.Itoa(bet))
			return
		}

		// Pick a number game
		if game == "number" {
			numberErrMessage := "!gamble <amount> number <numberToGuess>:<highestNumberInRange>. So !gamble 100 number 10:100 will run a pick a number game between 1 and 100 and the payout will be x100, because you have a 1  in 100 chance to win."
			gameInputs := strings.Split(gameInput, ":")

			if len(gameInputs) != 2 {
				_, _ = s.ChannelMessageSend(m.ChannelID, numberErrMessage)
				return
			}
			pickedNumber, err := strconv.Atoi(gameInputs[0])
			if err != nil {
				_, _ = s.ChannelMessageSend(m.ChannelID, numberErrMessage)
				return
			}
			rangeNumber, err := strconv.Atoi(gameInputs[1])
			if err != nil || rangeNumber < pickedNumber {
				_, _ = s.ChannelMessageSend(m.ChannelID, numberErrMessage)
				return
			}
			if rangeNumber <= 1 {
				_, _ = s.ChannelMessageSend(m.ChannelID, "your highestNumberInRange needs to be greater than 1")
				return
			}

			answer := rand.Intn(rangeNumber)
			if answer == pickedNumber {
				payout := betToPayout(bet, float64(rangeNumber+1))
				moneyAdd(&author, payout, "gamble")
				_, _ = s.ChannelMessageSend(m.ChannelID, "The result was "+strconv.Itoa(answer)+". Congrats, you won "+strconv.Itoa(payout)+" memes.")
			} else {
				moneyDeduct(&author, bet, "gamble")
				_, _ = s.ChannelMessageSend(m.ChannelID, "The result was "+strconv.Itoa(answer)+". Bummer, you lost "+strconv.Itoa(bet)+" memes. :(")
			}
		}

		// Coin flip game
		if game == "coin" || game == "flip" {
			if gameInput == "heads" || gameInput == "tails" {
				answers := []string{"heads", "tails"}
				answer := answers[rand.Intn(len(answers))]

				if answer == gameInput {
					// 1x payout
					payout := betToPayout(bet, 1.0)
					moneyAdd(&author, payout, "gamble")
					_, _ = s.ChannelMessageSend(m.ChannelID, "The result was "+answer+". Congrats, you won "+strconv.Itoa(payout)+" memes.")
				} else {
					moneyDeduct(&author, bet, "gamble")
					_, _ = s.ChannelMessageSend(m.ChannelID, "The result was "+answer+". Bummer, you lost "+strconv.Itoa(bet)+" memes. :(")
				}
			} else {
				_, _ = s.ChannelMessageSend(m.ChannelID, "pick heads or tails bud. `!gamble <amount> coin heads|tails`")
			}
		}
	} else if args[0] == "!gamble" {
		_, _ = s.ChannelMessageSend(m.ChannelID,
			`Gamble command is used as follows: '!gamble <amount> <game> <gameInput>
			 '!gamble <amount> coin|flip heads|tails' payout is 1x
			 '!gamble <amount> number <numberToGuess>:<highestNumberInRange>' payout is whatever the <highestNumberInRange> is.`,
		)
	}
}

func handleMine(s *discordgo.Session, m *discordgo.MessageCreate) {
	author := userGet(m.Author)
	lastMineTime := author.MineTime
	now := time.Now()
	difference := now.Sub(lastMineTime)
	timeLimit := 5

	mineResponses := []MineResponse{
		MineResponse{
			amount:   100,
			response: "you mined for a while and managed to scrounge up 100 dusty memes",
			chance:   50,
		},
		MineResponse{
			amount:   300,
			response: "you mined for a bit and found an uncommon pepe worth 300 memes!",
			chance:   30,
		},
		MineResponse{
			amount:   1000,
			response: "you fell down a meme-shaft and found a very dank rare pepe worth 1000 memes!",
			chance:   15,
		},
		MineResponse{
			amount:   5000,
			response: "you wandered in the meme mine for what seems like forever, eventually stumbling upon a vintage 1980's pepe worth 5000 memes!",
			chance:   5,
		},
		MineResponse{
			amount:   25000,
			response: "your meme mining has made the Maymay gods happy, they rewarded you with an MLG-shiny-holofoil-dankasfuck Pepe Diamond worth 25000 memes!",
			chance:   1,
		}}

	if difference.Minutes() < float64(timeLimit) {
		waitTime := strconv.Itoa(int(math.Ceil((float64(timeLimit) - difference.Minutes()))))
		_, _ = s.ChannelMessageSend(m.ChannelID, "you're too tired to mine, you must rest your meme muscles for "+waitTime+" more minute(s)")
		return
	} else {
		responseList := []MineResponse{}
		for _, response := range mineResponses {
			counter := response.chance
			for counter > 0 {
				responseList = append(responseList, response)
				counter--
			}
		}
		// pick a response out of the responses in responseList
		mineResponse := responseList[(rand.Intn(len(responseList)))]
		moneyAdd(&author, mineResponse.amount, "mined")
		_, _ = s.ChannelMessageSend(m.ChannelID, mineResponse.response)
		fmt.Println(author.Username + " mined " + strconv.Itoa(mineResponse.amount))
		return
	}
	return
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == BotID {
		return
	}

	if strings.Contains(m.Content, "!tip") == true {
		handleTip(s, m)
	}

	if m.Content == "!balance" || m.Content == "!memes" {
		handleBalance(s, m)
	}

	if strings.Contains(m.Content, "!gamble") {
		handleGamble(s, m)
	}

	if m.Content == "!mine" {
		handleMine(s, m)
	}

	if m.Content == "meme" {
		_, _ = s.ChannelMessageSend(m.ChannelID, "you're a memestar harry")
	}
}
