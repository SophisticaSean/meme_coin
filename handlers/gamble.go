package handlers

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/SophisticaSean/meme_coin/interaction"
	"github.com/dustin/go-humanize"
	"github.com/jmoiron/sqlx"
)

func betToPayout(bet int, payoutMultiplier float64) int {
	payout := int(math.Floor(float64(bet) * payoutMultiplier))
	return payout
}

func gambleProcess(content string, author *User, db *sqlx.DB) string {
	message := ""
	args := strings.Split(content, " ")
	if len(args) == 4 || len(args) == 5 {
		bet, err := strconv.Atoi(args[1])
		if err != nil {
			message = "amount is too large or not a number, try again."
			return message
		}
		if bet <= 0 {
			message = "amount has to be more than 0"
			return message
		}

		game := args[2]
		gameInput := args[3]

		loopAmount := 1

		if len(args) == 5 {
			convertedLoopAmount, err := strconv.Atoi(args[4])
			if err != nil || convertedLoopAmount < 1 || convertedLoopAmount > 500 {
				message = "amount of times to run the gamble is too high or not a number, try again."
				return message
			}
			loopAmount = convertedLoopAmount
		}

		totalBet := bet
		if len(args) == 5 {
			totalBet = totalBet * loopAmount
			if totalBet < 1 {
				message = "your bet * loopamount is too big, or not a number, try lowering the loop amount or bet amount"
				return message
			}
		}

		loopAmount = 1

		if bet > author.CurMoney {
			message = "not enough funds to complete transaction, total: " + humanize.Comma(int64(author.CurMoney)) + " needed:" + humanize.Comma(int64(bet))
			return message
		}

		if totalBet > author.CurMoney {
			message = "not enough funds to complete transaction, total: " + humanize.Comma(int64(author.CurMoney)) + " needed:" + humanize.Comma(int64(totalBet))
			return message
		}

		isTest, _ := os.LookupEnv("TEST")
		if isTest == "" {
			rand.Seed(time.Now().UnixNano())
		}

		if loopAmount == 1 {
			// Pick a number game
			if game == "number" {
				message, _ = numberGame(gameInput, bet, author, db)
				fmt.Println(message)
				return message
			}

			// Coin flip game
			if game == "coin" || game == "flip" {
				message, _ = coinGame(gameInput, bet, author, db)
				fmt.Println(message)
				return message
			}
		} else {
			wins := 0
			winAmount := 0
			losses := 0
			lossAmount := 0
			loop := loopAmount
			for loop > 0 {
				curMessage := ""
				curAmount := 0
				if game == "number" {
					curMessage, curAmount = numberGame(gameInput, bet, author, db)
				}

				// Coin flip game
				if game == "coin" || game == "flip" {
					curMessage, curAmount = coinGame(gameInput, bet, author, db)
				}

				if strings.Contains(curMessage, "won") {
					wins++
					winAmount = winAmount + curAmount
				} else if strings.Contains(curMessage, "lost") {
					losses++
					lossAmount = lossAmount + curAmount
				} else {
					loop = 0
					return curMessage
				}
				loop--
			}
			message = author.Username + " gambled " + strconv.Itoa(loopAmount) + " times. You won " + strconv.Itoa(wins) + " times, and lost " + strconv.Itoa(losses) + ".\r"
			fmt.Println(message)
			message2 := "Your net gamble gain was " + humanize.Comma(int64((winAmount)-(lossAmount)))
			fmt.Println(message2)
			return message + message2
		}
	} else if args[0] == "!gamble" {
		message = `
			Gamble command is used as follows: ` + "`" + `!gamble <amount> <game> <gameInput> <loopAmount>` + "`" + `
			 ` + "`" + `!gamble <amount> coin|flip heads|tails` + "`" + ` payout is 1x
			 ` + "`" + `!gamble <amount> number <numberToGuess>:<highestNumberInRange>` + "`" + ` payout is whatever the <highestNumberInRange - 1> is.
				ex: ` + "`" + `!gamble 10 coin heads 5` + "`" + `
				ex: ` + "`" + `!gamble 10 number 3:4` + "`" + `
				`
		return message
	}
	return message
}

func winLoseProcessor(answer string, pickedItem string, payout float64, bet int, author *User, db *sqlx.DB) (string, int) {
	message := "The result was " + answer
	if answer == pickedItem {
		payout := betToPayout(bet, payout)
		MoneyAdd(author, payout, "gamble", db)
		message = message + ". Congrats, " + author.Username + " won " + humanize.Comma(int64(payout)) + " memes."
		return message, payout
	}
	MoneyDeduct(author, bet, "gamble", db)
	message = message + ". Bummer, " + author.Username + " lost " + humanize.Comma(int64(bet)) + " memes. :("
	return message, bet
}

// Gamble is the function that handles the interaction of a user and gambling their memes
func Gamble(s interaction.Session, m *interaction.MessageCreate, db *sqlx.DB) {
	author := UserGet(m.Author, db)
	message := gambleProcess(m.Content, &author, db)
	_, _ = s.ChannelMessageSend(m.ChannelID, message)
	return
}

func numberGame(gameInput string, bet int, author *User, db *sqlx.DB) (string, int) {
	numberErrMessage := "!gamble <amount> number <numberToGuess>:<highestNumberInRange>. So !gamble 100 number 10:100 will run a pick a number game between 1 and 100 and the payout will be x100, because you have a 1  in 100 chance to win."
	message := ""
	gameInputs := strings.Split(gameInput, ":")

	if len(gameInputs) != 2 {
		return numberErrMessage, 0
	}
	pickedNumber, err := strconv.Atoi(gameInputs[0])
	if err != nil || pickedNumber < 1 {
		return numberErrMessage, 0
	}
	rangeNumber, err := strconv.Atoi(gameInputs[1])
	if err != nil || rangeNumber < pickedNumber {
		return numberErrMessage, 0
	}
	if rangeNumber <= 1 {
		message = "your highestNumberInRange needs to be greater than 1"
		return message, 0
	}

	answer := humanize.Comma(int64(rand.Intn(rangeNumber) + 1))
	strPickedNumber := humanize.Comma(int64(pickedNumber))
	amount := 0
	message, amount = winLoseProcessor(answer, strPickedNumber, float64(rangeNumber-1), bet, author, db)
	return message, amount
}

func coinGame(gameInput string, bet int, author *User, db *sqlx.DB) (string, int) {
	message := ""
	if gameInput == "heads" || gameInput == "tails" {
		num := rand.Intn(99)
		answer := ""
		if num > 50 {
			answer = gameInput
		} else {
			if gameInput == "heads" {
				answer = "tails"
			} else {
				answer = "heads"
			}
		}
		amount := 0
		message, amount = winLoseProcessor(answer, gameInput, 1.0, bet, author, db)
		return message, amount
	}
	message = "pick heads or tails bud. `!gamble <amount> coin heads|tails`"
	return message, 0
}
