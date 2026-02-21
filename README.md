# Movie API

A RESTful API for searching movies and managing watchlists using the OMDB API.

## Features

- Search for movies using OMDB API
- Get detailed movie information
- Create users and manage watchlists
- Cache movie data in SQLite database

## Requirements

- Go 1.16+
- OMDB API Key (sign up at http://www.omdbapi.com/)

## Setup

1. Clone the repository
2. Copy `.env.example` to `.env` and add your OMDB API key
3. Install dependencies: `go mod tidy`
4. Run the application: `go run *.go` or `go run main.go db.go handlers.go models.go omdb.go`

## API Endpoints

### Health & diagnostics

- `GET /ping` – simple health check returning `{ "message": "Server running" }`.

### Movie data

- `GET /movies/search?q={query}` – search OMDB for titles matching the query.  Returns JSON with `results` array.
- `GET /movies/{imdbID}` – fetch detailed info for a single IMDb ID.  This endpoint also **caches** the movie in the local database.

### Users

- `POST /users` – register a new user.  Request body: `{ "name": "Alice", "email": "alice@example.com" }`.
- `GET /users?email={address}` – lookup user by email (used for login).
- `GET /users/{id}` – fetch user profile by ID.

### Watchlist

- `POST /watchlist` – add a movie to a user's watchlist.  JSON body: `{ "user_id": 5, "imdb_id": "tt0111161", "status": "WATCHLIST" }`.
- `GET /users/{id}/watchlist` – list all movies in a user's watchlist, including cached metadata such as genre, plot and IMDB rating.
- `PUT /watchlist/{id}` – update status/rating of a watchlist item.  Body fields: `status` and/or `rating` (or `user_rating`).
- `DELETE /watchlist/{id}` – remove a single watchlist entry.
- `DELETE /users/{id}/watchlist` – clear an entire watchlist.

---

## Database schema

SQLite is used with the following tables:

```sql
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT,
    email TEXT UNIQUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

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
);

CREATE TABLE IF NOT EXISTS watchlist (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    movie_id INTEGER,
    status TEXT CHECK(status IN ('WATCHLIST','WATCHED')),
    user_rating INTEGER CHECK(user_rating BETWEEN 1 AND 5),
    added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(user_id) REFERENCES users(id),
    FOREIGN KEY(movie_id) REFERENCES movies(id)
);
```

The Go model `AddToWatchlistRequest` is used internally but the API talks only in terms of IDs.

## Caching strategy

When a client requests `/movies/{imdbID}` or adds a movie to a watchlist, the server:

1. **Checks the `movies` table** for an existing record with that IMDb ID.
2. If not found, calls the external OMDB API to fetch details.
3. Inserts (or replaces) the result into `movies`, timestamping it with `cached_at`.
4. Returns the movie details to the caller.

Adding to a watchlist automatically triggers step 1; if the movie is missing the handler now
fetches and caches it before creating the watchlist row, so clients no longer need to
pre‑cache manually.  Rarely, when the OMDB API is unavailable, the error is propagated to the
user.

Cache refresh is not automatic – the data stays until the database is deleted or updated
explicitly.  Future versions could add an eviction policy or background refresh job.

## Frontend

A simple HTML/JS UI is provided in `frontend/index.html`.  Major features:

- Dark cinematic theme with film‑grain overlay and stylized buttons.
- Registration/login form storing user info in `localStorage`.
- Movie search box triggering the `/movies/search` endpoint.
- Dynamically-generated watchlist section with rating, genres, plot.
- Buttons to mark as watched, remove individual items, clear the watchlist.

The front end no longer hardcodes a user ID; it interacts with the API using the
currently logged-in user.

## External API setup

Obtain an OMDB API key:

1. Visit http://www.omdbapi.com/ and register for a free key (or paid plan).
2. Create a `.env` file in the project root with:

```
OMDB_API_KEY=yourkeyhere
PORT=8080
```

The application loads this key on start.  Without a valid key, movie search/details will
fail with a 500 error.

## Running the application

```bash
# install dependencies
go mod tidy

# start the server (uses SQLite file movieapi.db in the same directory)
go run .
```

Open `frontend/index.html` in a browser (no server required) and interact with the
API.  You can also use `curl`/`Invoke-RestMethod` to exercise the endpoints directly.

## Design & implementation

See `DESIGN.md` for a high-level architecture discussion and rationale.

## Prompts used

The development of this project involved iterating with an AI assistant.  Here are the
primary user prompts given during the session (the AI's responses are omitted):

1. “index.html:397 POST http://localhost:8080/watchlist 500 (Internal Server Error)... solve this error”
2. “now whwn i click mark as watched it is showing failed”
3. “and the rating and add more details from the movie”
4. “remove the word bollywood from frontend make it cinema and adjust the styling of it”
5. “can you make the frontend dark cinematic theme”
6. “can u add a button where i can remove movie from my watchlist?”
7. “add user database”
8. “Database schema for user data and cached movies README with API documentation and external API setup Document explaining your caching strategy update readme according to this neatly and A README.md file explaining the design and implementation A markdown design document (if separate from README) All prompts used (if any AI tools were used to generate the code) all these should be included”

These prompts directed the iterative development of handlers, database functions, and
the front-end UI.

---

**Author:** interactive assistant with live edits in the VS Code workspace
**Date:** February 22, 2026
