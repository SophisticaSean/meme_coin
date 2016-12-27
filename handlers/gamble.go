package handlers

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"

	"github.com/SophisticaSean/meme_coin/interaction"
	humanize "github.com/dustin/go-humanize"
	"github.com/jmoiron/sqlx"
)

func betToPayout(bet int, payoutMultiplier float64) int {
	payout := int(math.Floor(float64(bet) * payoutMultiplier))
	return payout
}

func gambleProcess(content string, author *User, db *sqlx.DB) string {
	message := ""
	args := strings.Split(content, " ")
	if len(args) == 4 {
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

		if bet > author.CurMoney {
			message = "not enough funds to complete transaction, total: " + humanize.Comma(int64(author.CurMoney)) + " needed:" + humanize.Comma(int64(bet))
			return message
		}

		// Pick a number game
		if game == "number" {
			numberErrMessage := "!gamble <amount> number <numberToGuess>:<highestNumberInRange>. So !gamble 100 number 10:100 will run a pick a number game between 1 and 100 and the payout will be x100, because you have a 1  in 100 chance to win."
			gameInputs := strings.Split(gameInput, ":")

			if len(gameInputs) != 2 {
				return numberErrMessage
			}
			pickedNumber, err := strconv.Atoi(gameInputs[0])
			if err != nil || pickedNumber < 1 {
				return numberErrMessage
			}
			rangeNumber, err := strconv.Atoi(gameInputs[1])
			if err != nil || rangeNumber < pickedNumber {
				return numberErrMessage
			}
			if rangeNumber <= 1 {
				message = "your highestNumberInRange needs to be greater than 1"
				return message
			}

			answer := humanize.Comma(int64(rand.Intn(rangeNumber) + 1))
			strPickedNumber := humanize.Comma(int64(pickedNumber))
			message = winLoseProcessor(answer, strPickedNumber, float64(rangeNumber-1), bet, author, db)
			return message
		}

		// Coin flip game
		if game == "coin" || game == "flip" {
			if gameInput == "heads" || gameInput == "tails" {
				num := rand.Intn(99)
				answer := ""
				if num > 47 {
					answer = gameInput
				} else {
					if gameInput == "heads" {
						answer = "tails"
					} else {
						answer = "heads"
					}
				}
				message = winLoseProcessor(answer, gameInput, 1.0, bet, author, db)
				return message
			}
			message = "pick heads or tails bud. `!gamble <amount> coin heads|tails`"
			return message
		}
	} else if args[0] == "!gamble" {
		message = `
			Gamble command is used as follows: ` + "`" + `'!gamble <amount> <game> <gameInput>` + "`" + `
			 ` + "`" + `'!gamble <amount> coin|flip heads|tails` + "`" + ` payout is 1x
			 ` + "`" + `!gamble <amount> number <numberToGuess>:<highestNumberInRange>` + "`" + ` payout is whatever the <highestNumberInRange - 1> is.`
		return message
	}
	return message
}

func winLoseProcessor(answer string, pickedItem string, payout float64, bet int, author *User, db *sqlx.DB) string {
	message := "The result was " + answer
	if answer == pickedItem {
		payout := betToPayout(bet, payout)
		MoneyAdd(author, payout, "gamble", db)
		message = message + ". Congrats, " + author.Username + " won " + humanize.Comma(int64(payout)) + " memes."
		fmt.Println(message)
		return message
	}
	MoneyDeduct(author, bet, "gamble", db)
	message = message + ". Bummer, " + author.Username + " lost " + humanize.Comma(int64(bet)) + " memes. :("
	fmt.Println(message)
	return message
}

// Gamble is the function that handles the interaction of a user and gambling their memes
func Gamble(s interaction.Session, m *interaction.MessageCreate, db *sqlx.DB) {
	author := UserGet(m.Author, db)
	message := gambleProcess(m.Content, &author, db)
	_, _ = s.ChannelMessageSend(m.ChannelID, message)
	return
}
