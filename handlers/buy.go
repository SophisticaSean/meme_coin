package handlers

import (
	"fmt"
	"strconv"
	"strings"

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
	(passively generated memes are only added to your account every 10 minutes)
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

func UnitInfo(s *discordgo.Session, m *discordgo.MessageCreate, db *sqlx.DB) {
	userUnits := dbHandler.UnitsGet(m.Author, db)
	tempUnitList := UnitList()
	message := ""
	for _, unit := range tempUnitList {
		if unit.name == "miner" {
			fmt.Println(unit.name)
			unit.amount = userUnits.Miner
			message = message + "`" + strconv.Itoa(unit.amount) + " " + unit.name + "(s)` \r"
		}
		if unit.name == "robot" {
			fmt.Println(unit.name)
			unit.amount = userUnits.Robot
			message = message + "`" + strconv.Itoa(unit.amount) + " " + unit.name + "(s)` \r"
		}
		if unit.name == "swarm" {
			fmt.Println(unit.name)
			unit.amount = userUnits.Swarm
			message = message + "`" + strconv.Itoa(unit.amount) + " " + unit.name + "(s)` \r"
		}
		if unit.name == "fraker" {
			fmt.Println(unit.name)
			unit.amount = userUnits.Fracker
			message = message + "`" + strconv.Itoa(unit.amount) + " " + unit.name + "(s)` \r"
		}
	}
	fmt.Println(message)
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
	if unit.name == "fraker" {
		userUnits.Fracker = userUnits.Fracker + amount
	}
	dbHandler.UpdateUnits(&userUnits, db)
	_, _ = s.ChannelMessageSend(m.ChannelID, "You successfully bought "+strconv.Itoa(amount)+" "+args[2]+"(s)")
	return
}
