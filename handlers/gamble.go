package handlers

import (
	"math"
	"math/rand"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"
	"github.com/sophisticasean/meme_coin/dbHandler"
)

func BetToPayout(bet int, payoutMultiplier float64) int {
	payout := int(math.Floor(float64(bet) * payoutMultiplier))
	return payout
}

func Gamble(s *discordgo.Session, m *discordgo.MessageCreate, db *sqlx.DB) {
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
				payout := BetToPayout(bet, float64(rangeNumber+1))
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
					payout := BetToPayout(bet, 1.0)
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
