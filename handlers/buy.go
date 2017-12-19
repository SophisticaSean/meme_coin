package handlers

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/SophisticaSean/meme_coin/interaction"
	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/jmoiron/sqlx"
)

// Unit is a struct defining a unit you can buy
type Unit struct {
	Name       string
	Cost       int
	Production int
	Amount     int
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
	!buy max miners
	(passively generated memes are added to your account with !collect command)
	(buying units resets the time on your generated memes, so remember to collect before
	you buy!)
	(you can also use 'max' as a number and it will buy the max amount of those units you can afford)
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
		{
			Name:       "miner",
			Cost:       1000,
			Production: 1,
		},
		{
			Name:       "robot",
			Cost:       50000,
			Production: 60,
		},
		{
			Name:       "swarm",
			Cost:       2500000,
			Production: 3600,
		},
		{
			Name:       "fracker",
			Cost:       125000000,
			Production: 216000,
		},
		{
			Name:       "cypher",
			Cost:       10000,
			Production: 2,
		},
		{
			Name:       "hacker",
			Cost:       500,
			Production: 5,
		},
		{
			Name:       "botnet",
			Cost:       100,
			Production: 1,
		},
	}
	return unitList
}

// Balance is a function that returns the current balance for a user
func Balance(s interaction.Session, m *interaction.MessageCreate, db *sqlx.DB) {
	args := strings.Split(m.Content, " ")
	if len(args) == 1 {
		author := UserGet(m.Author, db)
		if author.CurMoney < 0 {
			author.CurMoney = author.CurMoney * -1
			MoneySet(&author, author.CurMoney, db)
		}
		message := author.Username + "\r"
		message = message + "Prestige Level " + strconv.Itoa(author.PrestigeLevel) + "\r"
		message = message + "total balance is: " + humanize.Comma(int64(author.CurMoney))
		_, production, _ := ProductionSum(m.Author, db)
		production = PrestigeBonus(production, &author)
		message = message + "\ntotal memes per minute: " + Ftoa(float64(production) / 10)
		message = message + "\nnet gambling balance: " + humanize.Comma(int64(author.WonMoney-author.LostMoney))
		_, _ = s.ChannelMessageSend(m.ChannelID, message)
	} else {
		_, _ = s.ChannelMessageSend(m.ChannelID, "usage: !balance [user]")
	}
}

func totalMemesEarned(user *discordgo.User, db *sqlx.DB) (int, string, User) {
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
			memes = int(((int(60*productionPerMinute) + memes) * productionMultiplier) / 100)
			roundedDifference = roundedDifference - 60
		}
	}
	if roundedDifference > 0 {
		memes = memes + int(productionPerMinute*roundedDifference)
	}
	if memes == 0.0 {
		message = "you don't have enough memes to collect right now."
		return memes, message, userUnits
	}

	//if memes < 0 {
	////message = "looks like you're trying to collect too many memes! You can fix this by `!buy`ing some units to reset your collect time. It's probably time for you to prestige and reset your meme production for a percentage bonus."
	////return memes, message, userUnits
	//memes = 9223372036854775807
	//}

	return memes, message, userUnits
}

func collectHelper(author *discordgo.User, db *sqlx.DB) (message string) {
	user := UserGet(author, db)
	totalMemesEarned, _, _ := totalMemesEarned(author, db)
	totalMemesEarned = PrestigeBonus(totalMemesEarned, &user)
	if totalMemesEarned < 0 {
		totalMemesEarned = 9223372036854775807
	}
	MoneyAdd(&user, totalMemesEarned, "collected", db)
	user.HackSeed = 0
	user.HackAttempts = 0
	user.CollectTime = time.Now()
	UpdateUnits(&user, db)
	message = user.Username + " collected " + humanize.Comma(int64(totalMemesEarned)) + " memes!"
	fmt.Println(message)
	message = message + "\rDon't forget, you get a " + strconv.Itoa(multiplier) + "% compound interest for every hour you let your memes stay uncollected (up to 24 hours)."
	return message
}

// Collect is a function that moves uncollected memes into the memebank/user's balance CurMoney
func Collect(s interaction.Session, m *interaction.MessageCreate, db *sqlx.DB) {
	_, message, _ := totalMemesEarned(m.Author, db)
	if message != "" {
		s.ChannelMessageSend(m.ChannelID, message)
		return
	}

	message = collectHelper(m.Author, db)

	s.ChannelMessageSend(m.ChannelID, message)
	return
}

// FakeCollect is a function that reports how many uncollected memes could be collected at the time it was called
func FakeCollect(s interaction.Session, m *interaction.MessageCreate, db *sqlx.DB) {
	user := UserGet(m.Author, db)
	totalMemesEarned, message, _ := totalMemesEarned(m.Author, db)
	if message != "" {
		_, _ = s.ChannelMessageSend(m.ChannelID, message)
		return
	}
	totalMemesEarned = PrestigeBonus(totalMemesEarned, &user)
	message = user.Username + " can collect " + humanize.Comma(int64(totalMemesEarned)) + " memes right now."
	s.ChannelMessageSend(m.ChannelID, message)
	return
}

// ProductionSum computes the amount of memes/minute someone has and returns a message
// with that information, the int of the memes/minute and the user's userUnits struct
func ProductionSum(user *discordgo.User, db *sqlx.DB) (string, int, User) {
	userUnits := UserGet(user, db)
	tempUnitList := UnitList()
	message := ""
	production := 0
	productionUnit := false
	for _, unit := range tempUnitList {
		if unit.Name == "miner" {
			unit.Amount = userUnits.Miner
			productionUnit = true
		}
		if unit.Name == "robot" {
			unit.Amount = userUnits.Robot
			productionUnit = true
		}
		if unit.Name == "swarm" {
			unit.Amount = userUnits.Swarm
			productionUnit = true
		}
		if unit.Name == "fracker" {
			unit.Amount = userUnits.Fracker
			productionUnit = true
		}
		if productionUnit == true {
			production = production + (unit.Amount * unit.Production)
			message = message + "`" + humanize.Comma(int64(unit.Amount)) + " " + unit.Name + "(s)` \r"
		}
		productionUnit = false
	}
	prestigeProduction := PrestigeBonus(production, &userUnits)
	message = message + "total memes per minute: " + humanize.Comma(int64(float64(prestigeProduction)/10))
	return message, production, userUnits
}

func militarySum(user *discordgo.User, db *sqlx.DB) (string, int, int, int, User) {
	userUnits := UserGet(user, db)
	tempUnitList := UnitList()
	message := ""
	botnet := 0
	defense := 0
	hacking := 0
	militaryUnit := false
	for _, unit := range tempUnitList {
		if unit.Name == "cypher" {
			unit.Amount = userUnits.Cypher
			defense = defense + (unit.Amount * unit.Production)
			militaryUnit = true
		}
		if unit.Name == "hacker" {
			unit.Amount = userUnits.Hacker
			hacking = hacking + (unit.Amount * unit.Production)
			militaryUnit = true
		}
		if unit.Name == "botnet" {
			unit.Amount = userUnits.Botnet
			botnet = botnet + (unit.Amount * unit.Production)
			militaryUnit = true
		}
		if militaryUnit {
			message = message + "`" + humanize.Comma(int64(unit.Amount)) + " " + unit.Name + "(s)` \r"
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
	if strings.ToLower(args[0]) != "!buy" {
		return
	}
	if len(args) == 1 || len(args) != 3 {
		_, _ = s.ChannelMessageSend(m.ChannelID, infoMessage)
		return
	}

	unit := Unit{}
	validUnit := false
	for _, i := range unitList {
		if strings.ToLower(args[2]) == i.Name || strings.ToLower(args[2]) == i.Name+"s" {
			unit = i
			validUnit = true
		}
	}

	if validUnit == false {
		_, _ = s.ChannelMessageSend(m.ChannelID, "2nd argument needs to be a correct unit type, check `!buy` for valid unit types")
		return
	}

	user := UserGet(m.Author, db)
	maxAmountToBuy := int(math.Floor(float64(user.CurMoney / unit.Cost)))
	var amount int
	var err error
	var totalCost int

	if strings.ToUpper(args[1]) == strings.ToUpper("max") {
		if maxAmountToBuy > 0 {
			totalCost = unit.Cost * maxAmountToBuy
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

		totalCost = unit.Cost * amount

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

	message := ""
	totalMemesEarned, _, _ := totalMemesEarned(m.Author, db)
	if totalMemesEarned > 0 {
		message = collectHelper(m.Author, db)
		message = message + "\n"
		user = UserGet(m.Author, db)
	}

	MoneyDeduct(&user, totalCost, "buy", db)
	userUnits := UserGet(m.Author, db)
	// gross if statements to determine what number to increment
	if unit.Name == "miner" {
		userUnits.Miner = userUnits.Miner + amount
	}
	if unit.Name == "robot" {
		userUnits.Robot = userUnits.Robot + amount
	}
	if unit.Name == "swarm" {
		userUnits.Swarm = userUnits.Swarm + amount
	}
	if unit.Name == "fracker" {
		userUnits.Fracker = userUnits.Fracker + amount
	}
	if unit.Name == "cypher" {
		userUnits.Cypher = userUnits.Cypher + amount
	}
	if unit.Name == "hacker" {
		userUnits.Hacker = userUnits.Hacker + amount
	}
	if unit.Name == "botnet" {
		userUnits.Botnet = userUnits.Botnet + amount
	}
	userUnits.CollectTime = time.Now()
	UpdateUnits(&userUnits, db)
	message = message + user.Username + " successfully bought " + humanize.Comma(int64(amount)) + " " + unit.Name + "(s)"
	fmt.Println(message)
	_, _ = s.ChannelMessageSend(m.ChannelID, message)
	return
}
