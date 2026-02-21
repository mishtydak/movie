package main

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

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

	// 3️⃣ Save to DB (must succeed before returning)
	saveErr := SaveMovieToDB(movieFromAPI)
	if saveErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to cache movie: " + saveErr.Error(),
		})
		return
	}

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

	c.JSON(http.StatusOK, gin.H{"user_id": id})
}

func GetUserHandler(c *gin.Context) {
	userID, _ := strconv.Atoi(c.Param("id"))
	user, err := GetUserByID(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

func FindUserByEmailHandler(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email query required"})
		return
	}
	user, err := GetUserByEmail(email)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

func AddToWatchlistHandler(c *gin.Context) {
	var req struct {
		UserID int    `json:"user_id"`
		ImdbID string `json:"imdb_id"`
		Status string `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	alreadyExists, err := IsMovieInWatchlist(req.UserID, req.ImdbID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking watchlist"})
		return
	}
	if alreadyExists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Movie is already in your watchlist"})
		return
	}

	err = AddMovieToWatchlist(req.UserID, req.ImdbID, req.Status)
	if err != nil {
		if strings.Contains(err.Error(), "movie not cached") {
			movie, fetchErr := GetMovieDetails(req.ImdbID)
			if fetchErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch movie details"})
				return
			}
			saveErr := SaveMovieToDB(movie)
			if saveErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cache movie"})
				return
			}
			err = AddMovieToWatchlist(req.UserID, req.ImdbID, req.Status)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Added to watchlist successfully"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Added to watchlist successfully"})
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
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid watchlist item ID"})
		return
	}

	var body struct {
		Status     string `json:"status"`
		Rating     int    `json:"rating"`
		UserRating int    `json:"user_rating"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rating := body.Rating
	if body.UserRating != 0 {
		rating = body.UserRating
	}

	err = UpdateWatchlist(id, body.Status, rating)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Updated successfully"})
}

func DeleteWatchlistItemHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid watchlist item ID"})
		return
	}

	var count int
	err = DB.QueryRow(`SELECT COUNT(*) FROM watchlist WHERE id = ?`, id).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking watchlist item"})
		return
	}

	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Watchlist item not found"})
		return
	}

	result, err := DB.Exec(`DELETE FROM watchlist WHERE id = ?`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete watchlist item"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Watchlist item not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Watchlist item deleted successfully"})
}

func ClearUserWatchlistHandler(c *gin.Context) {
	userID, _ := strconv.Atoi(c.Param("id"))
	result, err := DB.Exec(`DELETE FROM watchlist WHERE user_id = ?`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear watchlist"})
		return
	}

	rc, _ := result.RowsAffected()
	c.JSON(http.StatusOK, gin.H{"message": "Watchlist cleared successfully", "deleted_count": rc})
}
