package handlers

import (
	_ "database/sql" // necessary for sqlx
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/SophisticaSean/meme_coin/interaction"
	_ "github.com/bmizerany/pq" // necessary for sqlx
	"github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"
)

// User is a struct that maps 1 to 1 with 'money' db table
type User struct {
	ID              int       `db:"id"`
	DID             string    `db:"money_discord_id"`
	Username        string    `db:"name"`
	CurMoney        int       `db:"current_money"`
	TotMoney        int       `db:"total_money"`
	WonMoney        int       `db:"won_money"`
	LostMoney       int       `db:"lost_money"`
	GiveMoney       int       `db:"given_money"`
	RecMoney        int       `db:"received_money"`
	EarMoney        int       `db:"earned_money"`
	SpentMoney      int       `db:"spent_money"`
	CollectedMoney  int       `db:"collected_money"`
	HackedMoney     int       `db:"hacked_money"`
	StolenFromMoney int       `db:"stolen_money"`
	MineTime        time.Time `db:"mine_time"`
	Miner           int       `db:"miner"`
	Robot           int       `db:"robot"`
	Swarm           int       `db:"swarm"`
	Fracker         int       `db:"fracker"`
	Cypher          int       `db:"cyphers"`
	Hacker          int       `db:"hackers"`
	Botnet          int       `db:"botnets"`
	HackSeed        int64     `db:"hack_seed"`
	HackAttempts    int       `db:"hack_attempts"`
	PrestigeLevel   int       `db:"prestige_level"`
	CollectTime     time.Time `db:"collect_time"`
	UnitsDID        string    `db:"units_discord_id"`
}

// UserUnits is a struct that maps 1 to 1 with units db table, keeps track of what units users have purchased
type UserUnits struct {
	DID           string    `db:"units_discord_id"`
	Miner         int       `db:"miner"`
	Robot         int       `db:"robot"`
	Swarm         int       `db:"swarm"`
	Fracker       int       `db:"fracker"`
	Cypher        int       `db:"cyphers"`
	Hacker        int       `db:"hackers"`
	Botnet        int       `db:"botnets"`
	HackSeed      int64     `db:"hack_seed"`
	HackAttempts  int       `db:"hack_attempts"`
	PrestigeLevel int       `db:"prestige_level"`
	CollectTime   time.Time `db:"collect_time"`
}

var schema = `
CREATE TABLE IF NOT EXISTS money(id SERIAL PRIMARY KEY, money_discord_id VARCHAR(100), name VARCHAR(100), current_money numeric DEFAULT(1000), total_money numeric DEFAULT(0), won_money numeric DEFAULT(0), lost_money numeric DEFAULT(0), given_money numeric DEFAULT(0), received_money numeric DEFAULT(0), earned_money numeric DEFAULT(1000), spent_money numeric DEFAULT(0), collected_money numeric DEFAULT(0), hacked_money numeric DEFAULT(0), stolen_money numeric DEFAULT(0), mine_time timestamptz NOT NULL DEFAULT(now()));

CREATE TABLE IF NOT EXISTS units(units_discord_id VARCHAR(100) PRIMARY KEY, miner numeric DEFAULT(0), robot numeric DEFAULT(0), swarm numeric DEFAULT(0), fracker numeric DEFAULT(0), hackers numeric DEFAULT(0), botnets numeric DEFAULT(0), cyphers numeric DEFAULT(0), hack_seed numeric DEFAULT(0), hack_attempts numeric DEFAULT(0), prestige_level numeric DEFAULT(0), collect_time timestamptz NOT NULL DEFAULT(now()));

CREATE TABLE IF NOT EXISTS transactions(id SERIAL PRIMARY KEY, transactions_discord_id VARCHAR(100), amount numeric DEFAULT(0), type VARCHAR(100), time timestamptz NOT NULL DEFAULT(now()));
`

var dropSchema = `
DROP TABLE IF EXISTS money;
DROP TABLE IF EXISTS units;
DROP TABLE IF EXISTS transactions;
`

// DbGet returns a pointer to a db connection
func DbGet() *sqlx.DB {
	isTest, _ := os.LookupEnv("TEST")
	var db *sqlx.DB
	var err error
	if isTest != "" {
		db, err = sqlx.Connect("postgres", "host=localhost user=memebot dbname=test password=password sslmode=disable parseTime=true")
	} else {
		db, err = sqlx.Connect("postgres", "host=localhost user=memebot dbname=money password=password sslmode=disable parseTime=true")
	}
	if err != nil {
		log.Fatal(err)
	}
	db.MustExec(schema)
	return db
}

// DbReset clears the test db
func DbReset() {
	db, err := sqlx.Connect("postgres", "host=localhost user=memebot dbname=test password=password sslmode=disable parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	db.MustExec(dropSchema)
	db.MustExec(schema)
}

func createUser(user *discordgo.User, db *sqlx.DB) {
	var newUser User
	newUser.DID = user.ID
	newUser.Username = user.Username
	newUser.MineTime = time.Now().Add(-10 * time.Minute)
	dbString := `INSERT INTO money (money_discord_id, name, mine_time) VALUES (:money_discord_id, :name, :mine_time)`
	_, err := db.NamedExec(dbString, newUser)
	if err != nil {
		log.Fatal(err)
	}
	createUserUnits(user, db)
}

// UserGet returns a User by searching/creating the user in the db and using a discordgo.User
func UserGet(discordUser *discordgo.User, db *sqlx.DB) User {
	var users []User
	err := db.Select(&users, `
	SELECT
		m.id as id,
		m.money_discord_id as money_discord_id,
		m.name as name,
		m.current_money as current_money,
		m.total_money as total_money,
		m.won_money as won_money,
		m.lost_money as lost_money,
		m.given_money as given_money,
		m.received_money as received_money,
		m.earned_money as earned_money,
		m.spent_money as spent_money,
		m.collected_money as collected_money,
		m.hacked_money as hacked_money,
		m.stolen_money as stolen_money,
		m.mine_time as mine_time,
		u.miner as miner,
		u.robot as robot,
		u.swarm as swarm,
		u.fracker as fracker,
		u.cyphers as cyphers,
		u.hackers as hackers,
		u.botnets as botnets,
		u.hack_seed as hack_seed,
		u.hack_attempts as hack_attempts,
		u.prestige_level as prestige_level,
		u.collect_time as collect_time
	FROM money as m
	INNER JOIN units as u on m.money_discord_id = u.units_discord_id
	WHERE m.money_discord_id = $1

	`, discordUser.ID)
	var user User
	if err != nil {
		log.Fatal(err)
	}
	if len(users) == 0 {
		createUser(discordUser, db)
		user = UserGet(discordUser, db)
	} else {
		user = users[0]
		if user.Username != discordUser.Username {
			db.MustExec(`UPDATE money SET (name) = ($1) where money_discord_id = '`+user.DID+`'`, discordUser.Username)
			user.Username = discordUser.Username
		}
	}
	return user
}

// GetAllUsers returns a []User slice of all users in the db. Used for the API
func GetAllUsers(db *sqlx.DB) []User {
	var users []User
	err := db.Select(&users, `SELECT * FROM money INNER JOIN units ON (money.money_discord_id = units.units_discord_id);`)
	if err != nil {
		log.Fatal(err)
	}
	// censor the hackseed
	for i := range users {
		users[i].HackSeed = 0
	}
	return users
}

// MoneyDeduct handles all possible deductions
func MoneyDeduct(user *User, amount int, deduction string, db *sqlx.DB) {
	negativeAmount := amount * -1
	newCurrentMoney := user.CurMoney + negativeAmount
	newDeductionAmount := -1
	dbString := ``
	deductionRecord := -1

	if deduction == "tip" {
		dbString = `UPDATE money SET (current_money, given_money) = ($1, $2) WHERE money_discord_id = `
		deductionRecord = user.GiveMoney
		newDeductionAmount = user.GiveMoney + amount
		user.CurMoney = newCurrentMoney
		user.GiveMoney = newDeductionAmount
	}
	if deduction == "gamble" {
		dbString = `UPDATE money SET (current_money, lost_money) = ($1, $2) WHERE money_discord_id = `
		deductionRecord = user.LostMoney
		newDeductionAmount = user.LostMoney + amount
		user.CurMoney = newCurrentMoney
		user.LostMoney = newDeductionAmount
	}

	if deduction == "buy" {
		dbString = `UPDATE money SET (current_money, spent_money) = ($1, $2) WHERE money_discord_id = `
		deductionRecord = user.SpentMoney
		newDeductionAmount = user.SpentMoney + amount
		user.CurMoney = newCurrentMoney
		user.SpentMoney = newDeductionAmount
	}

	if deduction == "hacked" {
		dbString = `UPDATE money SET (current_money, stolen_money) = ($1, $2) WHERE money_discord_id = `
		deductionRecord = user.StolenFromMoney
		newDeductionAmount = user.StolenFromMoney + amount
		// don't actually deduct any money, we're taking it from their uncollected funds
		newCurrentMoney = user.CurMoney
		user.StolenFromMoney = newDeductionAmount
	}

	if dbString != `` && deductionRecord != -1 && newDeductionAmount != -1 {
		dbString = dbString + `'` + user.DID + `'`
		db.MustExec(dbString, newCurrentMoney, newDeductionAmount)
		db.MustExec(`INSERT INTO transactions (transactions_discord_id, amount, type) VALUES ($1, $2, $3)`, user.DID, negativeAmount, deduction)
	}
}

// MoneyAdd handles all possible meme additions
func MoneyAdd(user *User, amount int, addition string, db *sqlx.DB) {
	newCurrentMoney := user.CurMoney + amount

	newAdditionAmount := -1
	dbString := ``
	additionRecord := -1

	if addition == "tip" {
		dbString = `UPDATE money SET (current_money, received_money) = ($1, $2) WHERE money_discord_id = `
		additionRecord = user.RecMoney
		newAdditionAmount = user.RecMoney + amount
		user.CurMoney = newCurrentMoney
		user.RecMoney = newAdditionAmount
	}
	if addition == "gamble" {
		dbString = `UPDATE money SET (current_money, won_money) = ($1, $2) WHERE money_discord_id = `
		additionRecord = user.WonMoney
		newAdditionAmount = user.WonMoney + amount
		user.CurMoney = newCurrentMoney
		user.WonMoney = newAdditionAmount
	}
	if addition == "collected" {
		dbString = `UPDATE money SET (current_money, collected_money) = ($1, $2) WHERE money_discord_id = `
		additionRecord = user.CollectedMoney
		newAdditionAmount = user.CollectedMoney + amount
		user.CurMoney = newCurrentMoney
		user.CollectedMoney = newAdditionAmount
	}
	if addition == "hacked" {
		dbString = `UPDATE money SET (current_money, hacked_money) = ($1, $2) WHERE money_discord_id = `
		additionRecord = user.HackedMoney
		newAdditionAmount = user.HackedMoney + amount
		user.CurMoney = newCurrentMoney
		user.HackedMoney = newAdditionAmount
	}
	if addition == "mined" {
		dbString = `UPDATE money SET (current_money, earned_money, mine_time) = ($1, $2, $3) WHERE money_discord_id = `
		additionRecord = user.EarMoney
		newAdditionAmount = user.EarMoney + amount
		user.CurMoney = newCurrentMoney
		user.EarMoney = newAdditionAmount
		// bindvars can only be used as values so we have to concat the user.DID onto the db string
		dbString = dbString + `'` + user.DID + `'`
		db.MustExec(dbString, newCurrentMoney, newAdditionAmount, time.Now())
		// add the transaction to the database
		db.MustExec(`INSERT INTO transactions (transactions_discord_id, amount, type) VALUES ($1, $2, $3)`, user.DID, amount, addition)
	} else {
		if dbString != `` && additionRecord != -1 && newAdditionAmount != -1 {
			// bindvars can only be used as values so we have to concat the user.DID onto the db string
			dbString = dbString + `'` + user.DID + `'`
			db.MustExec(dbString, newCurrentMoney, newAdditionAmount)
			// add the transaction to the database
			db.MustExec(`INSERT INTO transactions (transactions_discord_id, amount, type) VALUES ($1, $2, $3)`, user.DID, amount, addition)
		}
	}
}

// MoneySet handles setting curmoney directly
func MoneySet(user *User, amount int, db *sqlx.DB) {
	dbString := `UPDATE money SET (current_money) = ($1) WHERE money_discord_id = `
	dbString = dbString + `'` + user.DID + `'`
	db.MustExec(dbString, amount)
}

// UpdateUnits updates all Units table information on a User
func UpdateUnits(userUnits *User, db *sqlx.DB) {
	dbString := `UPDATE units SET (miner, robot, swarm, fracker, collect_time, cyphers, hackers, botnets, hack_seed, hack_attempts, prestige_level) = ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) WHERE units_discord_id = `
	dbString = dbString + `'` + userUnits.DID + `'`
	db.MustExec(dbString, userUnits.Miner, userUnits.Robot, userUnits.Swarm, userUnits.Fracker, userUnits.CollectTime, userUnits.Cypher, userUnits.Hacker, userUnits.Botnet, userUnits.HackSeed, userUnits.HackAttempts, userUnits.PrestigeLevel)
}

func createUserUnits(user *discordgo.User, db *sqlx.DB) {
	var newUser UserUnits
	newUser.DID = user.ID
	_, err := db.NamedExec(`INSERT INTO units (units_discord_id) VALUES (:units_discord_id)`, newUser)
	if err != nil {
		log.Fatal(err)
	}
}

// Reset is a helper that makes setting a user back to scratch super easy
func Reset(s interaction.Session, m *interaction.MessageCreate, db *sqlx.DB) {
	for _, resetUser := range m.Mentions {
		// reset their money
		db.MustExec(`UPDATE money set (current_money, total_money, won_money, lost_money, given_money, received_money, earned_money, spent_money, collected_money) = (1000,0,0,0,0,0,0,0,0) where money_discord_id = '` + resetUser.ID + `'`)
		// reset their units
		db.MustExec(`UPDATE units set (miner, robot, swarm, fracker, cyphers, hackers, botnets, hack_seed, hack_attempts, prestige_level) = (0,0,0,0,0,0,0,0,0,0) where units_discord_id = '` + resetUser.ID + `'`)
		db.MustExec(`DELETE FROM transactions  where transactions_discord_id = '` + resetUser.ID + `'`)
		message := resetUser.Username + " has been reset."
		_, _ = s.ChannelMessageSend(m.ChannelID, message)
	}
	return
}

// ResetUser is for internal Go usage of reseting a user, whereas Reset is for the !reset command
func ResetUser(resetUser User, db *sqlx.DB) {
	// reset their money
	db.MustExec(`UPDATE money set (current_money, total_money, won_money, lost_money, given_money, received_money, earned_money, spent_money, collected_money) = (1000,0,0,0,0,0,0,0,0) where money_discord_id = '` + resetUser.DID + `'`)
	// reset their units
	db.MustExec(`UPDATE units set (miner, robot, swarm, fracker, cyphers, hackers, botnets, hack_seed, hack_attempts, prestige_level) = (0,0,0,0,0,0,0,0,0,0) where units_discord_id = '` + resetUser.DID + `'`)
	db.MustExec(`DELETE FROM transactions where transactions_discord_id = '` + resetUser.DID + `'`)
	return
}

// TempBan blocks a user from mining or collecting for the amount of days passed in
func TempBan(s interaction.Session, m *interaction.MessageCreate, db *sqlx.DB) {
	args := strings.Split(m.Content, " ")
	days := args[1]
	_, err := strconv.Atoi(days)
	if err != nil {
		days = "1"
	}
	for _, resetUser := range m.Mentions {
		// tmp ban their mine timeer
		db.MustExec(`UPDATE money set (mine_time) = (current_timestamp + interval '` + days + ` days') where money_discord_id = '` + resetUser.ID + `'`)
		// tmp ban their collect timer
		db.MustExec(`UPDATE units set (collect_time) = (current_timestamp + interval '` + days + ` days') where units_discord_id = '` + resetUser.ID + `'`)
		message := resetUser.Username + " has been banned for " + days + " day(s)."
		_, _ = s.ChannelMessageSend(m.ChannelID, message)
	}
	return
}

// Unban unblocks a user from mining or collecting
func Unban(s interaction.Session, m *interaction.MessageCreate, db *sqlx.DB) {
	args := strings.Split(m.Content, " ")
	days := args[1]
	_, err := strconv.Atoi(days)
	if err != nil {
		days = "1"
	}
	for _, resetUser := range m.Mentions {
		// tmp ban their mine timeer
		db.MustExec(`UPDATE money set (mine_time) = (current_timestamp) where money_discord_id = '` + resetUser.ID + `'`)
		// tmp ban their collect timer
		db.MustExec(`UPDATE units set (collect_time) = (current_timestamp) where units_discord_id = '` + resetUser.ID + `'`)
		message := resetUser.Username + " has been unbanned."
		_, _ = s.ChannelMessageSend(m.ChannelID, message)
	}
	return
}

// PrestigeBonus calculates the total memes a person receives when their prestigebonus is taken into account.
func PrestigeBonus(amount int, user *User) (total int) {
	// process prestige bonus
	prestigeBonus := user.PrestigeLevel * amount
	total = amount
	// protect from overflow
	if prestigeBonus > 0 {
		total = total + prestigeBonus
	}
	return total
}
