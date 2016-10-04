package handlers

import (
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"
)

func Hack(s *discordgo.Session, m *discordgo.MessageCreate, db *sqlx.DB) {
	args := strings.Split(m.Content, " ")
	seed, _ := strconv.Atoi(args[1])
	seed64 := int64(seed)
	popSize, _ := strconv.Atoi(args[2])
	iterationLimit, _ := strconv.Atoi(args[3])
	maxStringLength, _ := strconv.Atoi(args[4])
	fitness, generations := hackSimulate(seed64, popSize, iterationLimit, maxStringLength)
	message := "fitness: " + strconv.FormatFloat(fitness, 'f', -1, 64) + " generations: " + strconv.Itoa(generations)
	_, _ = s.ChannelMessageSend(m.ChannelID, message)
}
