package handlers

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"
)

// MineResponse is a struct for possible events to the !mine action
type MineResponse struct {
	amount   int
	response string
	chance   int
}

func GenerateResponseList() []MineResponse {
	mineResponses := []MineResponse{
		MineResponse{
			amount:   100,
			response: " mined for a while and managed to scrounge up 100 dusty memes",
			chance:   50,
		},
		MineResponse{
			amount:   300,
			response: " mined for a bit and found an uncommon pepe worth 300 memes!",
			chance:   30,
		},
		MineResponse{
			amount:   1000,
			response: " fell down a meme-shaft and found a very dank rare pepe worth 1000 memes!",
			chance:   15,
		},
		MineResponse{
			amount:   5000,
			response: " wandered in the meme mine for what seems like forever, eventually stumbling upon a vintage 1980's pepe worth 5000 memes!",
			chance:   5,
		},
		MineResponse{
			amount:   25000,
			response: "'s meme mining has made the Maymay gods happy, they rewarded them with a MLG-shiny-holofoil-dankasheck Pepe Diamond worth 25000 memes!",
			chance:   1,
		}}

	responseList := []MineResponse{}

	for _, response := range mineResponses {
		counter := response.chance
		for counter > 0 {
			responseList = append(responseList, response)
			counter--
		}
	}
	return responseList
}

func Mine(s *discordgo.Session, m *discordgo.MessageCreate, responseList []MineResponse, db *sqlx.DB) {
	author := UserGet(m.Author, db)
	difference := time.Now().Sub(author.MineTime)
	timeLimit := 5
	channel, _ := s.Channel(m.ChannelID)

	if channel.IsPrivate {
		_, _ = s.ChannelMessageSend(m.ChannelID, "you think you're slick, eh? gotta mine in a public room bro.")
		return
	}

	// check to make sure user is not trying to mine before timeLimit has passed
	if difference.Minutes() < float64(timeLimit) {
		waitTime := strconv.Itoa(int(math.Ceil((float64(timeLimit) - difference.Minutes()))))
		_, _ = s.ChannelMessageSend(m.ChannelID, m.Author.Username+" is too tired to mine, they must rest their meme muscles for "+waitTime+" more minute(s)")
		return
	}

	// pick a response out of the responses in responseList
	pickedIndex := rand.Intn(len(responseList))
	mineResponse := responseList[pickedIndex]
	MoneyAdd(&author, mineResponse.amount, "mined", db)
	_, _ = s.ChannelMessageSend(m.ChannelID, m.Author.Username+mineResponse.response)
	fmt.Println(m.Author.Username + " mined " + strconv.Itoa(mineResponse.amount))
	return
}
