package handlers

import (
	_ "database/sql"
	"fmt"
	"log"
	_ "strings"
	"time"

	_ "github.com/bmizerany/pq"
	"github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"
)

// User is a struct that maps 1 to 1 with 'money' db table
type User struct {
	ID             int       `db:"id"`
	DID            string    `db:"discord_id"`
	Username       string    `db:"name"`
	CurMoney       int       `db:"current_money"`
	TotMoney       int       `db:"total_money"`
	WonMoney       int       `db:"won_money"`
	LostMoney      int       `db:"lost_money"`
	GiveMoney      int       `db:"given_money"`
	RecMoney       int       `db:"received_money"`
	EarMoney       int       `db:"earned_money"`
	SpentMoney     int       `db:"spent_money"`
	CollectedMoney int       `db:"collected_money"`
	MineTime       time.Time `db:"mine_time"`
}

// UserUnits is a struct that maps 1 to 1 with units db table, keeps track of what units users have purchased
type UserUnits struct {
	DID         string    `db:"discord_id"`
	Miner       int       `db:"miner"`
	Robot       int       `db:"robot"`
	Swarm       int       `db:"swarm"`
	Fracker     int       `db:"fracker"`
	CollectTime time.Time `db:"collect_time"`
}

var schema = `
CREATE table money(id SERIAL PRIMARY KEY, discord_id VARCHAR(100), name VARCHAR(100), current_money numeric DEFAULT(1000), total_money numeric DEFAULT(0), won_money numeric DEFAULT(0), lost_money numeric DEFAULT(0), given_money numeric DEFAULT(0), received_money numeric DEFAULT(0), earned_money numeric DEFAULT(1000), spent_money numeric DEFAULT(0), collected_money numeric DEFAULT(0), mine_time timestamptz NOT NULL DEFAULT(now()));

create table units(discord_id VARCHAR(100) PRIMARY KEY, miner numeric DEFAULT(0), robot numeric DEFAULT(0), swarm numeric DEFAULT(0), fracker numeric DEFAULT(0), collect_time timestamptz NOT NULL DEFAULT(now()));

create table transactions(discord_id VARCHAR(100), amount numeric DEFAULT(0), type VARCHAR(100), time timestamptz NOT NULL DEFAULT(now()));
`

func DbGet() *sqlx.DB {
	db, err := sqlx.Connect("postgres", "host=localhost user=memebot dbname=money password=password sslmode=disable parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	db.MustExec(schema)
	return db
}

func TestDbGet() *sqlx.DB {
	db, err := sqlx.Connect("postgres", "host=localhost user=memebot dbname=test password=password sslmode=disable parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	db.MustExec(schema)
	return db
}

func createUser(user *discordgo.User, db *sqlx.DB) {
	var newUser User
	newUser.DID = user.ID
	newUser.Username = user.Username
	_, err := db.NamedExec(`INSERT INTO money (discord_id, name) VALUES (:discord_id, :name)`, newUser)
	if err != nil {
		log.Fatal(err)
	}
}

func UserGet(discordUser *discordgo.User, db *sqlx.DB) User {
	var users []User
	//fmt.Println(discordUser.ID)
	err := db.Select(&users, `SELECT id, discord_id, name, current_money, total_money, won_money, lost_money, given_money, received_money, earned_money, spent_money, mine_time FROM money WHERE discord_id = $1`, discordUser.ID)
	if err != nil {
		log.Fatal(err)
	}
	var user User
	if len(users) == 0 {
		fmt.Println("creating user: " + discordUser.ID)
		createUser(discordUser, db)
		user = UserGet(discordUser, db)
	} else {
		user = users[0]
		if user.Username != discordUser.Username {
			db.MustExec(`UPDATE money SET (name) = ($1) where discord_id = '`+user.DID+`'`, discordUser.Username)
			user.Username = discordUser.Username
		}
	}
	return user
}

func MoneyDeduct(user *User, amount int, deduction string, db *sqlx.DB) {
	negativeAmount := amount * -1
	newCurrentMoney := user.CurMoney + negativeAmount
	newDeductionAmount := -1
	dbString := ``
	deductionRecord := -1

	if deduction == "tip" {
		dbString = `UPDATE money SET (current_money, given_money) = ($1, $2) WHERE discord_id = `
		deductionRecord = user.GiveMoney
		newDeductionAmount = user.GiveMoney + amount
		user.CurMoney = newCurrentMoney
		user.GiveMoney = newDeductionAmount
	}
	if deduction == "gamble" {
		dbString = `UPDATE money SET (current_money, lost_money) = ($1, $2) WHERE discord_id = `
		deductionRecord = user.LostMoney
		newDeductionAmount = user.LostMoney + amount
		user.CurMoney = newCurrentMoney
		user.LostMoney = newDeductionAmount
	}

	if deduction == "buy" {
		dbString = `UPDATE money SET (current_money, spent_money) = ($1, $2) WHERE discord_id = `
		deductionRecord = user.SpentMoney
		newDeductionAmount = user.SpentMoney + amount
		user.CurMoney = newCurrentMoney
		user.SpentMoney = newDeductionAmount
	}

	if dbString != `` && deductionRecord != -1 && newDeductionAmount != -1 {
		dbString = dbString + `'` + user.DID + `'`
		db.MustExec(dbString, newCurrentMoney, newDeductionAmount)
		db.MustExec(`INSERT INTO transactions (discord_id, amount, type) VALUES ($1, $2, $3)`, user.DID, negativeAmount, deduction)
	}
}

func MoneyAdd(user *User, amount int, addition string, db *sqlx.DB) {
	newCurrentMoney := user.CurMoney + amount
	newAdditionAmount := -1
	dbString := ``
	additionRecord := -1

	if addition == "tip" {
		dbString = `UPDATE money SET (current_money, received_money) = ($1, $2) WHERE discord_id = `
		additionRecord = user.RecMoney
		newAdditionAmount = user.RecMoney + amount
		user.CurMoney = newCurrentMoney
		user.RecMoney = newAdditionAmount
	}
	if addition == "gamble" {
		dbString = `UPDATE money SET (current_money, won_money) = ($1, $2) WHERE discord_id = `
		additionRecord = user.WonMoney
		newAdditionAmount = user.WonMoney + amount
		user.CurMoney = newCurrentMoney
		user.WonMoney = newAdditionAmount
	}
	if addition == "collected" {
		dbString = `UPDATE money SET (current_money, collected_money) = ($1, $2) WHERE discord_id = `
		additionRecord = user.CollectedMoney
		newAdditionAmount = user.CollectedMoney + amount
		user.CurMoney = newCurrentMoney
		user.CollectedMoney = newAdditionAmount
	}
	if addition == "mined" {
		dbString = `UPDATE money SET (current_money, earned_money, mine_time) = ($1, $2, $3) WHERE discord_id = `
		additionRecord = user.EarMoney
		newAdditionAmount = user.EarMoney + amount
		user.CurMoney = newCurrentMoney
		user.EarMoney = newAdditionAmount
		// bindvars can only be used as values so we have to concat the user.DID onto the db string
		dbString = dbString + `'` + user.DID + `'`
		db.MustExec(dbString, newCurrentMoney, newAdditionAmount, time.Now())
		// add the transaction to the database
		db.MustExec(`INSERT INTO transactions (discord_id, amount, type) VALUES ($1, $2, $3)`, user.DID, amount, addition)
	} else {
		if dbString != `` && additionRecord != -1 && newAdditionAmount != -1 {
			// bindvars can only be used as values so we have to concat the user.DID onto the db string
			dbString = dbString + `'` + user.DID + `'`
			db.MustExec(dbString, newCurrentMoney, newAdditionAmount)
			// add the transaction to the database
			db.MustExec(`INSERT INTO transactions (discord_id, amount, type) VALUES ($1, $2, $3)`, user.DID, amount, addition)
		}
	}
}

// units functionality

func UnitsGet(discordUser *discordgo.User, db *sqlx.DB) UserUnits {
	var units []UserUnits
	err := db.Select(&units, `SELECT discord_id, miner, robot, swarm, fracker, collect_time FROM units WHERE discord_id = $1`, discordUser.ID)
	if err != nil {
		log.Fatal(err)
	}
	var unitObj UserUnits
	if len(units) == 0 {
		fmt.Println("creating user in units table: " + discordUser.ID)
		createUserUnits(discordUser, db)
		unitObj = UnitsGet(discordUser, db)
	} else {
		unitObj = units[0]
	}
	return unitObj
}

func UpdateUnits(userUnits *UserUnits, db *sqlx.DB) {
	dbString := `UPDATE units SET (miner, robot, swarm, fracker) = ($1, $2, $3, $4) WHERE discord_id = `
	dbString = dbString + `'` + userUnits.DID + `'`
	db.MustExec(dbString, userUnits.Miner, userUnits.Robot, userUnits.Swarm, userUnits.Fracker)
}

func UpdateUnitsTimestamp(userUnits *UserUnits, db *sqlx.DB) {
	dbString := `UPDATE units SET (collect_time) = ($1) WHERE discord_id = `
	dbString = dbString + `'` + userUnits.DID + `'`
	db.MustExec(dbString, time.Now())
}

func createUserUnits(user *discordgo.User, db *sqlx.DB) {
	var newUser UserUnits
	newUser.DID = user.ID
	_, err := db.NamedExec(`INSERT INTO units (discord_id) VALUES (:discord_id)`, newUser)
	if err != nil {
		log.Fatal(err)
	}
}
