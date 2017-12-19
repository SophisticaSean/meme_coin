package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/SophisticaSean/meme_coin/interaction"
	"github.com/jmoiron/sqlx"
)

// Prestige handles resetting a user and giving them a bonus
func Prestige(s interaction.Session, m *interaction.MessageCreate, db *sqlx.DB) {
	infoMessage := `
	usage: !prestige <are_you_sure>
	!prestige
	!prestige YESIMSURE
	!prestige help
	if you have enough of the requisite units, prestige will reset all your accumulated wealth
		and start you off at the beginning with 1000 memes and a flat +100% bonus multiplier for all future
		meme income.
	every prestige level doubles the amount of units you need to proceed
	`
	args := strings.Split(m.Content, " ")
	if strings.ToLower(args[0]) != "!prestige" {
		return
	}

	user := UserGet(m.Author, db)

	unitMultiplier := (1 + user.PrestigeLevel) * (1 + user.PrestigeLevel)
	necessaryUnitAmount := unitMultiplier * 100

	if len(args) == 1 {
		message := canPrestige(&user, necessaryUnitAmount)

		if message == "" {
			message = "You have enough units to prestige. If you are sure you want to prestige, type `!prestige YESIMSURE`"
		}

		message = user.Username + " is prestige level " + strconv.Itoa(user.PrestigeLevel) + " which is a bonus of +" + strconv.Itoa(user.PrestigeLevel*100) + " percent to all meme income.\r" + message
		s.ChannelMessageSend(m.ChannelID, message)
		return
	}

	if len(args) == 2 {
		if args[1] == "help" {
			s.ChannelMessageSend(m.ChannelID, infoMessage)
			return
		}
		if args[1] == "YESIMSURE" {
			message := canPrestige(&user, necessaryUnitAmount)

			if message != "" {
				s.ChannelMessageSend(m.ChannelID, message)
				return
			}

			ResetUser(user, db)
			// get fresh reset user before updating units
			newUser := UserGet(m.Author, db)
			newUser.PrestigeLevel = user.PrestigeLevel + 1
			fmt.Println(newUser.PrestigeLevel)
			UpdateUnits(&newUser, db)

			message = "You have been reset! Congratulations, you are now prestige level " + strconv.Itoa(newUser.PrestigeLevel) + ", which means you get a " + strconv.Itoa(newUser.PrestigeLevel*100) + " percentage bonus on all new meme income!"

			s.ChannelMessageSend(m.ChannelID, message)
			fmt.Println(user.Username + " prestiged to level " + strconv.Itoa(user.PrestigeLevel))
			return
		}
	}
	s.ChannelMessageSend(m.ChannelID, infoMessage)
	return
}

func canPrestige(user *User, necessaryUnitAmount int) (message string) {
	message = ""
	if user.Miner < (necessaryUnitAmount) {
		message = message + "You do not have enough miners to Prestige, you need " + strconv.Itoa(necessaryUnitAmount-user.Miner) + " more.\n"
	}
	if user.Robot < (necessaryUnitAmount) {
		message = message + "You do not have enough robots to Prestige, you need " + strconv.Itoa(necessaryUnitAmount-user.Robot) + " more.\n"
	}
	if user.Swarm < (necessaryUnitAmount) {
		message = message + "You do not have enough swarms to Prestige, you need " + strconv.Itoa(necessaryUnitAmount-user.Swarm) + " more.\n"
	}
	if user.Fracker < (necessaryUnitAmount) {
		message = message + "You do not have enough frackers to Prestige, you need " + strconv.Itoa(necessaryUnitAmount-user.Fracker) + " more.\n"
	}
	return message
}
