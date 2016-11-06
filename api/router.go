package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/SophisticaSean/meme_coin/handlers"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// RouterConfigure sets up our api routing and content serving
func RouterConfigure() (*sqlx.DB, *gin.Engine) {
	router := gin.Default()
	db := handlers.DbGet()

	router.GET("/help", func(c *gin.Context) {
		c.String(http.StatusOK, "all good")
	})

	router.GET("/users", func(c *gin.Context) {
		users := handlers.GetAllUsers(db)
		usersJSON, err := json.Marshal(users)
		if err != nil {
			log.Fatal(err)
		}
		c.Header("Access-Control-Allow-Origin", "*")
		c.String(http.StatusOK, string(usersJSON))
	})
	return db, router
}
