# Forum (Go + SQLite)

A minimal web forum built with Go, SQLite, and server-side HTML templates.

## Features
- Register/login with hashed passwords (bcrypt).
- Sessions stored in SQLite with expiration cookies.
- Create posts with categories.
- Comment on posts.
- Like/dislike posts and comments (toggle on/off).
- Filter posts by category, your posts, or your liked posts.

## Local run

1. Install Go 1.22+.
2. Start the server:

```bash
go run ./cmd/forum
```

Open `http://localhost:8080`.

### Environment variables
- `PORT` (default `8080`)
- `DB_PATH` (default `./forum.db`)
- `COOKIE_SECURE` (set to `1` to mark cookies as secure)

## Docker

Build and run:

```bash
docker build -t forum .
docker run -p 8080:8080 -e DB_PATH=/app/data/forum.db -v ${PWD}/data:/app/data forum
```

Or with Compose:

```bash
docker compose up --build
```

## Notes
- The database is initialized on startup using `schema.sql`.
- Default categories are seeded automatically when the database is empty.
