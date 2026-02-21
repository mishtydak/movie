# Design Document

This document outlines the architecture and design decisions behind the Movie API project.

## Overview

The application is a lightweight RESTful service written in Go and backed by SQLite. It
interfaces with the external OMDB API for movie information and maintains a local cache
of movie details to reduce external calls and support offline access. Users can register
and manage personalized watchlists, marking movies as watched or rating them.

A simple static HTML/JavaScript frontend consumes the API and provides an interactive
experience. The front end stores the current user in `localStorage` and adapts the
request URLs accordingly.

## Components

### Server (Go)

- **Entry point (`main.go`)**
  - Loads environment variables (OMDB key, port).
  - Connects to SQLite and ensures tables exist.
  - Registers HTTP routes using [Gin](https://github.com/gin-gonic/gin).

- **Database layer (`db.go`)**
  - Holds global `*sql.DB` connection.
  - Defines functions for all CRUD operations: user creation/lookup, movie cache, watchlist
    operations.
  - Table creation SQL executed at startup.

- **Handlers (`handlers.go`)**
  - Map HTTP endpoints to business logic.
  - Perform input validation and return JSON responses.
  - Wrap database calls and external OMDB requests.
  - Includes error handling and special cases (e.g. caching when adding to watchlist).

- **Models/structs (`models.go`)**
  - Define structs representing OMDB responses and request payloads.

- **OMDB integration (`omdb.go`)**
  - Contains functions to call the OMDB API using the key from environment variables.
  - Parses JSON into Go structs.

### Frontend (`frontend/index.html`)

- Single-file static page with embedded CSS styles and JavaScript.
- Dark cinematic theme with film grain overlay and stylized buttons.
- Contains user registration/login UI, movie search form, and watchlist display.
- Communicates with backend via `fetch` API; handles login state, caching, retries.

## Data Flow

1. **User registration/login:** user enters name/email; frontend stores user info,
   subsequent requests include the user ID.
2. **Searching:** query is sent to `/movies/search`; server forwards to OMDB and returns results.
3. **Caching:** when `/movies/{imdbID}` or watchlist POST is called, server checks cache;
   if missing, it fetches from OMDB and stores deep metadata in `movies` table.
4. **Watchlist management:** watchlist rows link `user_id` to `movie_id`. Updates and
   deletes are simple SQL statements executed by handlers.

## Caching Strategy

Caching is eager when needed by watchlist functionality.  The design avoids repeated
OMDB calls by storing a full copy of each movie detail locally.  On subsequent requests
for the same IMDb ID, the database record is returned directly.  This also means that
a watchlist entry always contains the movie details at the time of caching – changes on
OMDB will not reflect until the movie is re-fetched manually (currently there is no
re-fetch mechanism).

The caching logic is encapsulated in `GetMovieFromDB` / `SaveMovieToDB` and used by both
`GetMovieDetailsHandler` and `AddToWatchlistHandler`.

## Security and Validation

- Input is validated at handler level; IDs are converted with `strconv.Atoi`.
- SQL statements use parameterized queries (`?`) to avoid injection.
- Email uniqueness enforced by database constraint.
- No authentication beyond email lookup; passwords can be added in a future iteration.

## Extensibility

- Migrating to a different database requires swapping out the `sql` package and
  adjusting SQL syntax; the rest of the code is database-agnostic.
- Additional fields from OMDB can be added to the `movies` table as needed.
- The front end could be refactored into a framework (React, etc.) if required.

## Logging and Errors

Gin's default logger is enabled. Errors from database or HTTP calls are returned as JSON
with an `error` field; the frontend displays alerts when operations fail.

## Deployment

The app can be built into a single binary with `go build` and deployed to any host that
can run Go. The SQLite file persists data and can be backed up or moved as needed.
