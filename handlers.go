package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func SearchMoviesHandler(c *gin.Context) {
	query := c.Query("q")

	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Query parameter 'q' is required",
		})
		return
	}

	movies, err := SearchOMDB(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": movies,
	})
}
func GetMovieDetailsHandler(c *gin.Context) {
	imdbID := c.Param("imdbID")

	// 1️⃣ Try DB first
	movie, err := GetMovieFromDB(imdbID)
	if err == nil && movie != nil {
		c.JSON(http.StatusOK, gin.H{
			"source": "database",
			"movie":  movie,
		})
		return
	}

	// 2️⃣ Fetch from OMDB
	movieFromAPI, err := GetMovieDetails(imdbID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 3️⃣ Save to DB
	_ = SaveMovieToDB(movieFromAPI)

	c.JSON(http.StatusOK, gin.H{
		"source": "omdb",
		"movie":  movieFromAPI,
	})

}
func CreateUserHandler(c *gin.Context) {

	var req CreateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := CreateUser(req.Name, req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": id,
	})
}
func AddToWatchlistHandler(c *gin.Context) {

	var req AddToWatchlistRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := AddMovieToWatchlist(req.UserID, req.ImdbID, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Added to watchlist",
	})

}
func GetUserWatchlistHandler(c *gin.Context) {

	userID, _ := strconv.Atoi(c.Param("id"))

	list, err := GetUserWatchlist(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, list)
}
func UpdateWatchlistHandler(c *gin.Context) {

	id, _ := strconv.Atoi(c.Param("id"))

	var body struct {
		Status string `json:"status"`
		Rating int    `json:"rating"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := UpdateWatchlist(id, body.Status, body.Rating)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Updated successfully",
	})
}
