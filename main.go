package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	loadEnv()
	ConnectDB()

	r := gin.Default()
	r.Use(cors.Default())
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Server running"})
	})

	r.GET("/movies/search", SearchMoviesHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.GET("/movies/:imdbID", GetMovieDetailsHandler)
	// user endpoints
	r.POST("/users", CreateUserHandler)
	r.GET("/users", FindUserByEmailHandler) // query param email
	r.GET("/users/:id", GetUserHandler)

	// watchlist management
	r.POST("/watchlist", AddToWatchlistHandler)
	r.GET("/users/:id/watchlist", GetUserWatchlistHandler)
	r.PUT("/watchlist/:id", UpdateWatchlistHandler)
	r.DELETE("/watchlist/:id", DeleteWatchlistItemHandler)
	r.DELETE("/users/:id/watchlist", ClearUserWatchlistHandler)

	r.Run(":" + port)

}
