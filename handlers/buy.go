package handlers

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"
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
	Unit          Cost           Memes minute
	miner         1k             0.1 m/m
	robot         50k            6 m/m
	swarm         2.5mil         360 m/m
	fracker       125mil         21600 m/m
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
			cost:       2500000,
			production: 3600,
		},
		Unit{
			name:       "fracker",
			cost:       125000000,
			production: 216000,
		},
	}
	return unitList
}

func Balance(s *discordgo.Session, m *discordgo.MessageCreate, db *sqlx.DB) {
	args := strings.Split(m.Content, " ")
	if len(args) == 1 {
		author := UserGet(m.Author, db)
		message := "total balance is: " + strconv.Itoa(author.CurMoney)
		_, production, _ := ProductionSum(m.Author, db)
		message = message + "\ntotal memes per minute: " + strconv.FormatFloat((float64(production)/10), 'f', -1, 64)
		message = message + "\nnet gambling balance: " + strconv.Itoa(author.WonMoney-author.LostMoney)
		_, _ = s.ChannelMessageSend(m.ChannelID, message)
	}
}
func Collect(s *discordgo.Session, m *discordgo.MessageCreate, db *sqlx.DB) {
	_, production, userUnits := ProductionSum(m.Author, db)
	difference := time.Now().Sub(userUnits.CollectTime)
	diffMinutes := difference.Minutes()
	if diffMinutes < 1.0 {
		_, _ = s.ChannelMessageSend(m.ChannelID, "have to wait at least 1 minute between collections. \r its better to wait longer between collections, as we round down when computing how much memes you earned.")
		return
	}
	maxDifference := float64(24 * 60) //max difference is 1 days worth of minutes
	if diffMinutes > maxDifference {
		diffMinutes = maxDifference
	}
	roundedDifference := math.Floor(diffMinutes)
	productionPerMinute := float64(production) / 10.0
	totalMemesEarned := int(roundedDifference * productionPerMinute)
	if totalMemesEarned < 1.0 {
		_, _ = s.ChannelMessageSend(m.ChannelID, "you don't have enough memes to collect right now.")
		return
	}
	user := UserGet(m.Author, db)
	MoneyAdd(&user, totalMemesEarned, "collected", db)
	UpdateUnitsTimestamp(&userUnits, db)
	message := m.Author.Username + " collected " + strconv.Itoa(totalMemesEarned) + " memes!"
	fmt.Println(message)
	_, _ = s.ChannelMessageSend(m.ChannelID, message)
	return
}

func ProductionSum(user *discordgo.User, db *sqlx.DB) (string, int, UserUnits) {
	userUnits := UnitsGet(user, db)
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
	message = message + "total memes per minute: " + strconv.FormatFloat((float64(production)/10), 'f', -1, 64)
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
	if err != nil || amount < 1 {
		_, _ = s.ChannelMessageSend(m.ChannelID, "1st argument needs to be a number, and it needs to be greater than 0. `!buy 10 miners`")
		return
	}

	unit := Unit{}
	validUnit := false
	for _, i := range unitList {
		if args[2] == i.name || args[2] == i.name+"s" {
			unit = i
			validUnit = true
		}
	}

	if validUnit == false {
		_, _ = s.ChannelMessageSend(m.ChannelID, "2nd argument needs to be a correct unit type, check `!buy` for valid unit types")
		return
	}

	user := UserGet(m.Author, db)
	totalCost := (unit.cost * amount)

	if totalCost > user.CurMoney {
		strTotalCost := strconv.Itoa(totalCost)
		_, _ = s.ChannelMessageSend(m.ChannelID, "not enough money for transaction, need "+strTotalCost)
		return
	}

	MoneyDeduct(&user, totalCost, "buy", db)
	userUnits := UnitsGet(m.Author, db)
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
	UpdateUnits(&userUnits, db)
	UpdateUnitsTimestamp(&userUnits, db)
	message := m.Author.Username + " successfully bought " + strconv.Itoa(amount) + " " + unit.name + "(s)"
	fmt.Println(message)
	_, _ = s.ChannelMessageSend(m.ChannelID, message)
	return
}
