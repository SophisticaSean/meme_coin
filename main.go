package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	_ "database/sql"
	_ "strings"

	"github.com/sophisticasean/meme_coin/dbHandler"

	_ "github.com/bmizerany/pq"
	"github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"
)

// Variables used for command line parameters
var (
	Token string
	Email string
	PW    string
	BotID string
	db    *sqlx.DB
)

// MineResponse is a struct for possible events to the !mine action
type MineResponse struct {
	amount   int
	response string
	chance   int
}

var (
	responseList []MineResponse
)

func init() {
	Token, _ = os.LookupEnv("bot_token")
	Email, _ = os.LookupEnv("email")
	PW, _ = os.LookupEnv("pw")
	db = dbHandler.DbGet()
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
		from := dbHandler.UserGet(m.Author, db)
		if totalDeduct > from.CurMoney {
			_, _ = s.ChannelMessageSend(m.ChannelID, "not enough funds to complete transaction, total: "+strconv.Itoa(from.CurMoney)+" needed:"+strconv.Itoa(totalDeduct))
			return
		}
		dbHandler.MoneyDeduct(&from, totalDeduct, "tip", db)
		for _, to := range m.Mentions {
			toUser := dbHandler.UserGet(to, db)
			dbHandler.MoneyAdd(&toUser, intAmount, "tip", db)
			message := from.Username + " gave " + amount + " " + currencyName + " to " + to.Username
			_, _ = s.ChannelMessageSend(m.ChannelID, message)
			fmt.Println(message)
		}
		return
	}
	return
}

func handleBalance(s *discordgo.Session, m *discordgo.MessageCreate) {
	args := strings.Split(m.Content, " ")
	if len(args) == 1 {
		author := dbHandler.UserGet(m.Author, db)
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
		author := dbHandler.UserGet(m.Author, db)
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
				dbHandler.MoneyAdd(&author, payout, "gamble", db)
				_, _ = s.ChannelMessageSend(m.ChannelID, "The result was "+strconv.Itoa(answer)+". Congrats, you won "+strconv.Itoa(payout)+" memes.")
			} else {
				dbHandler.MoneyDeduct(&author, bet, "gamble", db)
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
					dbHandler.MoneyAdd(&author, payout, "gamble", db)
					_, _ = s.ChannelMessageSend(m.ChannelID, "The result was "+answer+". Congrats, you won "+strconv.Itoa(payout)+" memes.")
				} else {
					dbHandler.MoneyDeduct(&author, bet, "gamble", db)
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
	author := dbHandler.UserGet(m.Author, db)
	lastMineTime := author.MineTime
	now := time.Now()
	difference := now.Sub(lastMineTime)
	timeLimit := 5

	mineResponses := []MineResponse{
		MineResponse{
			amount:   100,
			response: " mined for a while and managed to scrounge up 100 dusty memes",
			chance:   50,
		},
		MineResponse{
			amount:   300,
			response: " mined for a bit and found an uncommon pepe worth 300 memes!",
			chance:   30,
		},
		MineResponse{
			amount:   1000,
			response: " fell down a meme-shaft and found a very dank rare pepe worth 1000 memes!",
			chance:   15,
		},
		MineResponse{
			amount:   5000,
			response: " wandered in the meme mine for what seems like forever, eventually stumbling upon a vintage 1980's pepe worth 5000 memes!",
			chance:   5,
		},
		MineResponse{
			amount:   25000,
			response: "'s meme mining has made the Maymay gods happy, they rewarded them with a MLG-shiny-holofoil-dankasfuck Pepe Diamond worth 25000 memes!",
			chance:   1,
		}}

	if difference.Minutes() < float64(timeLimit) {
		waitTime := strconv.Itoa(int(math.Ceil((float64(timeLimit) - difference.Minutes()))))
		_, _ = s.ChannelMessageSend(m.ChannelID, m.Author.Username+" is too tired to mine, they must rest their meme muscles for "+waitTime+" more minute(s)")
		return
	}
	// generate the responseList, and hopefully cache it in the global var
	fmt.Println(len(responseList))
	if len(responseList) == 0 {
		for _, response := range mineResponses {
			counter := response.chance
			for counter > 0 {
				responseList = append(responseList, response)
				counter--
			}
		}
	}
	// pick a response out of the responses in responseList
	mineResponse := responseList[(rand.Intn(len(responseList)))]
	dbHandler.MoneyAdd(&author, mineResponse.amount, "mined", db)
	_, _ = s.ChannelMessageSend(m.ChannelID, author.Username+mineResponse.response)
	fmt.Println(author.Username + " mined " + strconv.Itoa(mineResponse.amount))
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
