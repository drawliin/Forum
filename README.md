# Forum (Go + SQLite)

A minimal web forum built with Go, SQLite, and server-side HTML templates.

## Features
- Register/login with hashed passwords (bcrypt).
- Sessions stored in SQLite with expiration cookies.
- Create posts with categories.
- Comment on posts.
- Like/dislike posts and comments (toggle on/off).
- Filter posts by category, your posts, or your liked posts.

## Structure
- `cmd/forum`: main func
- `data`: database
- `internal`
    - `app`: entry point, server setup
    - `config`: project config variables
    - `handlers`: http handlers
    - `db`: database setup and fetch
    - `models`: db tables to golang structs
    - `templates`: html handling
    - `util`: helper functions
- `static`: static files
- `templates`: html templates
- `tests`: unit tests

## Local run

1. Install Go 1.22+.
2. create data folder
```
mkdir data
```
3. build
```
go build -o forum ./cmd/forum/
```
3. Start the server:

```bash
./forum
```

Open `http://localhost:8080`.

### Environment variables
- `PORT` (default `8080`)
- `DB_PATH` (default `./data/forum.db`)
- `COOKIE_SECURE` (set to `1` to mark cookies as secure)

## Run with Docker

```bash
docker compose up -d --build
```

## Notes
- The database is initialized on startup using `schema.sql`.
- Default categories are seeded automatically when the database is empty.
