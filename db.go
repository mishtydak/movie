package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func ConnectDB() {
	var err error

	DB, err = sql.Open("sqlite", "file:movieapi.db?cache=shared&mode=rwc")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatal("Database not reachable:", err)
	}

	fmt.Println("✅ Connected to SQLite database")

	createTables()
}

func createTables() {

	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		email TEXT UNIQUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	createMoviesTable := `
	CREATE TABLE IF NOT EXISTS movies (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		imdb_id TEXT UNIQUE,
		title TEXT,
		year TEXT,
		rated TEXT,
		released TEXT,
		runtime TEXT,
		genre TEXT,
		director TEXT,
		writer TEXT,
		actors TEXT,
		plot TEXT,
		language TEXT,
		country TEXT,
		awards TEXT,
		poster TEXT,
		imdb_rating TEXT,
		imdb_votes TEXT,
		type TEXT,
		cached_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	createWatchlistTable := `
	CREATE TABLE IF NOT EXISTS watchlist (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER ,
		movie_id INTEGER,
		status TEXT CHECK(status IN ('WATCHLIST','WATCHED')),
		user_rating INTEGER CHECK(user_rating BETWEEN 1 AND 5),
		added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(user_id) REFERENCES users(id),
		FOREIGN KEY(movie_id) REFERENCES movies(id)
		UNIQUE(user_id, movie_id)
	);`

	_, err := DB.Exec(createUsersTable)
	if err != nil {
		log.Fatal("Failed to create users table:", err)
	}

	_, err = DB.Exec(createMoviesTable)
	if err != nil {
		log.Fatal("Failed to create movies table:", err)
	}

	_, err = DB.Exec(createWatchlistTable)
	if err != nil {
		log.Fatal("Failed to create watchlist table:", err)
	}
}

//
// 🔥 CACHING FUNCTIONS
//

func GetMovieFromDB(imdbID string) (*OMDBMovieDetail, error) {

	row := DB.QueryRow(`
		SELECT imdb_id, title, year, rated, released, runtime,
		       genre, director, writer, actors, plot, language,
		       country, awards, poster, imdb_rating, imdb_votes, type
		FROM movies WHERE imdb_id = ?
	`, imdbID)

	var movie OMDBMovieDetail

	err := row.Scan(
		&movie.IMDBID,
		&movie.Title,
		&movie.Year,
		&movie.Rated,
		&movie.Released,
		&movie.Runtime,
		&movie.Genre,
		&movie.Director,
		&movie.Writer,
		&movie.Actors,
		&movie.Plot,
		&movie.Language,
		&movie.Country,
		&movie.Awards,
		&movie.Poster,
		&movie.IMDBRating,
		&movie.IMDBVotes,
		&movie.Type,
	)

	if err != nil {
		return nil, err
	}

	return &movie, nil
}

func SaveMovieToDB(movie *OMDBMovieDetail) error {

	_, err := DB.Exec(`
		INSERT OR REPLACE INTO movies (
			imdb_id, title, year, rated, released, runtime,
			genre, director, writer, actors, plot, language,
			country, awards, poster, imdb_rating, imdb_votes, type
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		movie.IMDBID,
		movie.Title,
		movie.Year,
		movie.Rated,
		movie.Released,
		movie.Runtime,
		movie.Genre,
		movie.Director,
		movie.Writer,
		movie.Actors,
		movie.Plot,
		movie.Language,
		movie.Country,
		movie.Awards,
		movie.Poster,
		movie.IMDBRating,
		movie.IMDBVotes,
		movie.Type,
	)

	return err
}
func CreateUser(name, email string) (int64, error) {
	result, err := DB.Exec(
		"INSERT INTO users (name, email) VALUES (?, ?)",
		name, email,
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}
func AddMovieToWatchlist(userID int, imdbID string, status string) error {

	// Get movie internal ID
	var movieID int
	err := DB.QueryRow(
		"SELECT id FROM movies WHERE imdb_id = ?",
		imdbID,
	).Scan(&movieID)

	if err != nil {
		return fmt.Errorf("movie not cached yet, fetch details first")
	}

	_, err = DB.Exec(`
		INSERT INTO watchlist (user_id, movie_id, status)
		VALUES (?, ?, ?)
	`, userID, movieID, status)

	return err
}
func GetUserWatchlist(userID int) ([]map[string]interface{}, error) {

	rows, err := DB.Query(`
		SELECT w.id, m.title, m.year, m.poster, w.status, w.user_rating
		FROM watchlist w
		JOIN movies m ON w.movie_id = m.id
		WHERE w.user_id = ?
	`, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}

	for rows.Next() {
		var id int
		var title, year, poster, status string
		var rating sql.NullInt64

		rows.Scan(&id, &title, &year, &poster, &status, &rating)

		results = append(results, map[string]interface{}{
			"watchlist_id": id,
			"title":        title,
			"year":         year,
			"poster":       poster,
			"status":       status,
			"user_rating":  rating.Int64,
		})
	}

	return results, nil
}
func UpdateWatchlist(watchlistID int, status string, rating int) error {

	_, err := DB.Exec(`
		UPDATE watchlist
		SET status = ?, user_rating = ?
		WHERE id = ?
	`, status, rating, watchlistID)

	return err
}
