package handlers

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/SophisticaSean/meme_coin/interaction"
	"github.com/bwmarrin/discordgo"
	humanize "github.com/dustin/go-humanize"
	"github.com/jmoiron/sqlx"
)

// Unit is a struct defining a unit you can buy
type Unit struct {
	name       string
	cost       int
	production int
	amount     int
}

const (
	multiplier = 3
)

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
	cypher        10k            +1 password strength
	hacker        500            +5 hacking strength
	botnet        100            +1 hacking strength
	`
	unitList = UnitList()
}

// UnitList returns a struct of Units with defined values
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
		Unit{
			name:       "cypher",
			cost:       10000,
			production: 2,
		},
		Unit{
			name:       "hacker",
			cost:       500,
			production: 5,
		},
		Unit{
			name:       "botnet",
			cost:       100,
			production: 1,
		},
	}
	return unitList
}

// Balance is a function that returns the current balance for a user
func Balance(s interaction.Session, m *interaction.MessageCreate, db *sqlx.DB) {
	args := strings.Split(m.Content, " ")
	if len(args) == 1 {
		author := UserGet(m.Author, db)
		message := "total balance is: " + humanize.Comma(int64(author.CurMoney))
		_, production, _ := ProductionSum(m.Author, db)
		message = message + "\ntotal memes per minute: " + humanize.Comma(int64((float64(production) / 10)))
		message = message + "\nnet gambling balance: " + humanize.Comma(int64(author.WonMoney-author.LostMoney))
		_, _ = s.ChannelMessageSend(m.ChannelID, message)
	}
}

func totalMemesEarned(user *discordgo.User, db *sqlx.DB) (int, string, UserUnits) {
	memes := 0
	message := ""
	_, production, userUnits := ProductionSum(user, db)
	difference := time.Now().Sub(userUnits.CollectTime)
	diffMinutes := difference.Minutes()
	if diffMinutes < 1.0 {
		message = "have to wait at least 1 minute between collections. \r its better to wait longer between collections, as we round down when computing how much memes you earned."
		return memes, message, userUnits
	}
	maxDifference := float64(24 * 60) //max difference is 1 days worth of minutes
	if diffMinutes > maxDifference {
		diffMinutes = maxDifference
	}
	roundedDifference := math.Floor(diffMinutes)
	roundedHours := math.Floor(diffMinutes / 60)
	productionMultiplier := int((multiplier) + 100)
	productionPerMinute := float64(production) / 10.0
	if int(roundedHours) > 0 {
		for i := 0; i < int(roundedHours); i++ {
			memes = int((((int(60*productionPerMinute) + memes) * productionMultiplier) / 100))
			roundedDifference = roundedDifference - 60
		}
	}
	if roundedDifference > 0 {
		memes = memes + int(productionPerMinute*roundedDifference)
	}
	if memes < 1.0 {
		message = "you don't have enough memes to collect right now."
		return memes, message, userUnits
	}
	return memes, message, userUnits
}

// Collect is a function that moves uncollected memes into the memebank/user's balance CurMoney
func Collect(s interaction.Session, m *interaction.MessageCreate, db *sqlx.DB) {
	user := UserGet(m.Author, db)
	totalMemesEarned, message, userUnits := totalMemesEarned(m.Author, db)
	if message != "" {
		_, _ = s.ChannelMessageSend(m.ChannelID, message)
		return
	}
	MoneyAdd(&user, totalMemesEarned, "collected", db)
	userUnits.HackSeed = 0
	userUnits.HackAttempts = 0
	userUnits.CollectTime = time.Now()
	UpdateUnits(&userUnits, db)
	message = user.Username + " collected " + humanize.Comma(int64(totalMemesEarned)) + " memes!"
	fmt.Println(message)
	message = message + "\rYou now get a " + strconv.Itoa(multiplier) + "% multiplier for every hour you let your memes stay uncollected."

	_, _ = s.ChannelMessageSend(m.ChannelID, message)
	return
}

// ProductionSum computes the amount of memes/minute someone has and returns a message
// with that information, the int of the memes/minute and the user's userUnits struct
func ProductionSum(user *discordgo.User, db *sqlx.DB) (string, int, UserUnits) {
	userUnits := UnitsGet(user, db)
	tempUnitList := UnitList()
	message := ""
	production := 0
	productionUnit := false
	for _, unit := range tempUnitList {
		if unit.name == "miner" {
			unit.amount = userUnits.Miner
			productionUnit = true
		}
		if unit.name == "robot" {
			unit.amount = userUnits.Robot
			productionUnit = true
		}
		if unit.name == "swarm" {
			unit.amount = userUnits.Swarm
			productionUnit = true
		}
		if unit.name == "fracker" {
			unit.amount = userUnits.Fracker
			productionUnit = true
		}
		if productionUnit == true {
			production = production + (unit.amount * unit.production)
			message = message + "`" + humanize.Comma(int64(unit.amount)) + " " + unit.name + "(s)` \r"
		}
		productionUnit = false
	}
	message = message + "total memes per minute: " + strconv.FormatFloat((float64(production)/10), 'f', -1, 64)
	return message, production, userUnits
}

func militarySum(user *discordgo.User, db *sqlx.DB) (string, int, int, int, UserUnits) {
	userUnits := UnitsGet(user, db)
	tempUnitList := UnitList()
	message := ""
	botnet := 0
	defense := 0
	hacking := 0
	militaryUnit := false
	for _, unit := range tempUnitList {
		if unit.name == "cypher" {
			unit.amount = userUnits.Cypher
			defense = defense + (unit.amount * unit.production)
			militaryUnit = true
		}
		if unit.name == "hacker" {
			unit.amount = userUnits.Hacker
			hacking = hacking + (unit.amount * unit.production)
			militaryUnit = true
		}
		if unit.name == "botnet" {
			unit.amount = userUnits.Botnet
			botnet = botnet + (unit.amount * unit.production)
			militaryUnit = true
		}
		if militaryUnit {
			message = message + "`" + humanize.Comma(int64(unit.amount)) + " " + unit.name + "(s)` \r"
		}
		militaryUnit = false
	}
	//message = message + "total botnets: " + strconv.Itoa(botnet)
	//message = message + "\rtotal cypher strength: " + strconv.Itoa(defense)
	//message = message + "\rtotal hackers: " + strconv.Itoa(hacking)
	return message, botnet, defense, hacking, userUnits
}

// UnitInfo returns the ProductionSum for a user
func UnitInfo(s interaction.Session, m *interaction.MessageCreate, db *sqlx.DB) {
	message, _, _ := ProductionSum(m.Author, db)
	_, _ = s.ChannelMessageSend(m.ChannelID, message)
	return
}

// MilitaryUnitInfo returns the militarySum info for a user
func MilitaryUnitInfo(s interaction.Session, m *interaction.MessageCreate, db *sqlx.DB) {
	message, _, _, _, _ := militarySum(m.Author, db)
	_, _ = s.ChannelMessageSend(m.ChannelID, message)
	return
}

// Buy handles unit buying for users
func Buy(s interaction.Session, m *interaction.MessageCreate, db *sqlx.DB) {
	args := strings.Split(m.Content, " ")
	if args[0] != "!buy" {
		return
	}
	if len(args) == 1 {
		_, _ = s.ChannelMessageSend(m.ChannelID, infoMessage)
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
	maxAmountToBuy := int(math.Floor(float64(user.CurMoney / unit.cost)))
	var amount int
	var err error
	var totalCost int

	if strings.ToUpper(args[1]) == strings.ToUpper("max") {
		if maxAmountToBuy > 0 {
			totalCost = (unit.cost * maxAmountToBuy)
			amount = maxAmountToBuy
			if totalCost < 0 {
				// handle the totalCost overflow case
				s.ChannelMessageSend(m.ChannelID, "You're trying to buy too many units at once, please lower the number and try again.")
				return
			}
			if totalCost == 0 || totalCost > user.CurMoney {
				s.ChannelMessageSend(m.ChannelID, "You ain't got enough cash to buy any of those units bro.")
				return
			}
		} else {
			s.ChannelMessageSend(m.ChannelID, "You ain't got enough cash to buy any of those units bro.")
			return
		}
	} else {
		amount, err = strconv.Atoi(args[1])
		if amount < 1 {
			_, _ = s.ChannelMessageSend(m.ChannelID, "1st argument needs to be a number or the word 'max', and it needs to be greater than 0. `!buy 10 miners`")
			return
		}
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "trying to buy too many units at once. try buying fewer units.")
			return
		}

		totalCost = (unit.cost * amount)

		if totalCost < 0 {
			// handle the totalCost overflow case
			s.ChannelMessageSend(m.ChannelID, "You're trying to buy too many units at once, please lower the number and try again.")
			return
		}

		if totalCost > user.CurMoney {
			strTotalCost := humanize.Comma(int64(totalCost))
			s.ChannelMessageSend(m.ChannelID, "not enough money for transaction, need "+strTotalCost+"\rYou can currently afford "+humanize.Comma(int64(maxAmountToBuy)))
			return
		}
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
	if unit.name == "cypher" {
		userUnits.Cypher = userUnits.Cypher + amount
	}
	if unit.name == "hacker" {
		userUnits.Hacker = userUnits.Hacker + amount
	}
	if unit.name == "botnet" {
		userUnits.Botnet = userUnits.Botnet + amount
	}
	userUnits.CollectTime = time.Now()
	UpdateUnits(&userUnits, db)
	message := user.Username + " successfully bought " + humanize.Comma(int64(amount)) + " " + unit.name + "(s)"
	fmt.Println(message)
	_, _ = s.ChannelMessageSend(m.ChannelID, message)
	return
}
