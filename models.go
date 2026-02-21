package main

import "time"

type OMDBMovie struct {
	Title  string `json:"Title"`
	Year   string `json:"Year"`
	ImdbID string `json:"imdbID"`
	Type   string `json:"Type"`
	Poster string `json:"Poster"`
}

type OMDBSearchResponse struct {
	Search   []OMDBMovie `json:"Search"`
	Response string      `json:"Response"`
	Error    string      `json:"Error"`
}
type OMDBMovieDetail struct {
	Title      string `json:"Title"`
	Year       string `json:"Year"`
	Rated      string `json:"Rated"`
	Released   string `json:"Released"`
	Runtime    string `json:"Runtime"`
	Genre      string `json:"Genre"`
	Director   string `json:"Director"`
	Writer     string `json:"Writer"`
	Actors     string `json:"Actors"`
	Plot       string `json:"Plot"`
	Language   string `json:"Language"`
	Country    string `json:"Country"`
	Awards     string `json:"Awards"`
	Poster     string `json:"Poster"`
	IMDBRating string `json:"imdbRating"`
	IMDBVotes  string `json:"imdbVotes"`
	IMDBID     string `json:"imdbID"`
	Type       string `json:"Type"`
	Response   string `json:"Response"`
	Error      string `json:"Error"`
}
type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type AddToWatchlistRequest struct {
	ID         uint `gorm:"primaryKey"`
	UserID     uint `gorm:"uniqueIndex:idx_user_movie"`
	MovieID    uint `gorm:"uniqueIndex:idx_user_movie"`
	Status     string
	UserRating int
	CreatedAt  time.Time // WATCHLIST or WATCHED
}
