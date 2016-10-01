package handlers

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"
)

func Reset(s *discordgo.Session, m *discordgo.MessageCreate, db *sqlx.DB) {
	for _, resetUser := range m.Mentions {
		// reset their money
		db.MustExec(`UPDATE money set (current_money, total_money, won_money, lost_money, given_money, received_money, earned_money, spent_money, collected_money) = (1000,0,0,0,0,0,0,0,0) where discord_id = '` + resetUser.ID + `'`)
		// reset their units
		db.MustExec(`UPDATE units set (miner, robot, swarm, fracker) = (0,0,0,0) where discord_id = '` + resetUser.ID + `'`)
		message := resetUser.Username + " has been reset."
		_, _ = s.ChannelMessageSend(m.ChannelID, message)
	}
	return
}
