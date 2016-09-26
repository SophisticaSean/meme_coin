package handlers

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"
	"github.com/sophisticasean/meme_coin/dbHandler"
)

type Unit struct {
	name       string
	cost       int
	production int
	amount     int
}

var (
	infoMessage string
	unitList    []Unit
)

func init() {
	infoMessage = `
	usage: !buy <amount> <unitType>
	!buy 10 miners
	(passively generated memes are added to your account with !collect command)
	(buying units resets the time on your generated memes, so remember to collect before
	you buy!)
	Unit list:
	Unit          Cost           Memes per 10 minutes
	miner         1k             1 m/m
	robot         50k            60 m/m
	swarm         2,500k         360 m/m
	fracker       125,000k       2160 m/m
	`
	unitList = UnitList()
}

func UnitList() []Unit {
	unitList := []Unit{
		Unit{
			name:       "miner",
			cost:       1000,
			production: 1,
		},
		Unit{
			name:       "robot",
			cost:       50000,
			production: 60,
		},
		Unit{
			name:       "swarm",
			cost:       250000,
			production: 360,
		},
		Unit{
			name:       "fracker",
			cost:       1250000,
			production: 2160,
		},
	}
	return unitList
}

func Collect(s *discordgo.Session, m *discordgo.MessageCreate, db *sqlx.DB) {
	_, production, userUnits := ProductionSum(m.Author, db)
	difference := time.Now().Sub(userUnits.CollectTime).Minutes()
	fmt.Println(difference)
	if difference < 1.0 {
		_, _ = s.ChannelMessageSend(m.ChannelID, "have to wait at least 1 minute between collections. \r its better to wait longer between collections, as we round down when computing how much memes you earned.")
		return
	}
	maxDifference := float64(24 * 60) //max difference is 1 days worth of minutes
	if difference > maxDifference {
		difference = maxDifference
	}
	roundedDifference := math.Floor(difference)
	productionPerMinute := math.Floor(float64(production) / 10.0)
	totalMemesEarned := int(roundedDifference * productionPerMinute)
	if totalMemesEarned < 1.0 {
		_, _ = s.ChannelMessageSend(m.ChannelID, "you don't have enough memes to collect right now.")
		return
	}
	user := dbHandler.UserGet(m.Author, db)
	dbHandler.MoneyAdd(&user, totalMemesEarned, "collected", db)
	dbHandler.UpdateUnitsTimestamp(&userUnits, db)
	message := m.Author.Username + " collected " + strconv.Itoa(totalMemesEarned) + " memes!"
	fmt.Println(message)
	_, _ = s.ChannelMessageSend(m.ChannelID, message)
	return
}

func ProductionSum(user *discordgo.User, db *sqlx.DB) (string, int, dbHandler.UserUnits) {
	userUnits := dbHandler.UnitsGet(user, db)
	tempUnitList := UnitList()
	message := ""
	production := 0
	for _, unit := range tempUnitList {
		if unit.name == "miner" {
			unit.amount = userUnits.Miner
			production = production + (unit.amount * unit.production)
			message = message + "`" + strconv.Itoa(unit.amount) + " " + unit.name + "(s)` \r"
		}
		if unit.name == "robot" {
			unit.amount = userUnits.Robot
			production = production + (unit.amount * unit.production)
			message = message + "`" + strconv.Itoa(unit.amount) + " " + unit.name + "(s)` \r"
		}
		if unit.name == "swarm" {
			unit.amount = userUnits.Swarm
			production = production + (unit.amount * unit.production)
			message = message + "`" + strconv.Itoa(unit.amount) + " " + unit.name + "(s)` \r"
		}
		if unit.name == "fracker" {
			unit.amount = userUnits.Fracker
			production = production + (unit.amount * unit.production)
			message = message + "`" + strconv.Itoa(unit.amount) + " " + unit.name + "(s)` \r"
		}
	}
	message = message + "total memes per 10 minutes: " + strconv.Itoa(production)
	return message, production, userUnits
}

func UnitInfo(s *discordgo.Session, m *discordgo.MessageCreate, db *sqlx.DB) {
	message, _, _ := ProductionSum(m.Author, db)
	_, _ = s.ChannelMessageSend(m.ChannelID, message)
	return
}

func Buy(s *discordgo.Session, m *discordgo.MessageCreate, db *sqlx.DB) {
	args := strings.Split(m.Content, " ")
	if args[0] != "!buy" {
		return
	}
	if len(args) == 1 {
		_, _ = s.ChannelMessageSend(m.ChannelID, infoMessage)
		return
	}

	amount, err := strconv.Atoi(args[1])
	if err != nil {
		_, _ = s.ChannelMessageSend(m.ChannelID, "1st argument needs to be a number. `!buy 10 miners`")
		return
	}

	unit := Unit{}
	validUnit := false
	for _, i := range unitList {
		if args[2] == i.name {
			unit = i
			validUnit = true
		}
	}

	if validUnit == false {
		_, _ = s.ChannelMessageSend(m.ChannelID, "2nd argument needs to be a correct unit type, check `!buy` for valid unit types")
		return
	}

	user := dbHandler.UserGet(m.Author, db)
	totalCost := (unit.cost * amount)

	if totalCost > user.CurMoney {
		strTotalCost := strconv.Itoa(totalCost)
		_, _ = s.ChannelMessageSend(m.ChannelID, "not enough money for transaction, need "+strTotalCost)
		return
	}

	dbHandler.MoneyDeduct(&user, totalCost, "buy", db)
	userUnits := dbHandler.UnitsGet(m.Author, db)
	// gross if statements to determine what number to increment
	if unit.name == "miner" {
		userUnits.Miner = userUnits.Miner + amount
	}
	if unit.name == "robot" {
		userUnits.Robot = userUnits.Robot + amount
	}
	if unit.name == "swarm" {
		userUnits.Swarm = userUnits.Swarm + amount
	}
	if unit.name == "fracker" {
		userUnits.Fracker = userUnits.Fracker + amount
	}
	dbHandler.UpdateUnits(&userUnits, db)
	dbHandler.UpdateUnitsTimestamp(&userUnits, db)
	message := m.Author.Username + " successfully bought " + strconv.Itoa(amount) + " " + args[2] + "(s)"
	fmt.Println(message)
	_, _ = s.ChannelMessageSend(m.ChannelID, message)
	return
}
