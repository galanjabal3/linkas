# Linkas — URL Shortener API

## Overview
A simple, fast URL shortener API built with Go + Gin + PostgreSQL.

## Tech Stack
- **Language:** Go 1.22+
- **Framework:** Gin
- **Database:** PostgreSQL
- **ORM:** sqlx (lightweight, close to raw SQL)
- **Migration:** golang-migrate
- **Testing:** testify

## Features
- [x] POST /api/shorten — Create short URL
- [x] GET /:slug — Redirect to original URL
- [x] GET /api/stats/:slug — Get click analytics
- [x] DELETE /api/:slug — Delete short URL
- [x] Rate limiting per IP
- [x] Custom slug support
- [x] URL validation
- [x] Click tracking (user-agent, referrer, IP)

## Database Schema

```sql
CREATE TABLE urls (
    id SERIAL PRIMARY KEY,
    slug VARCHAR(10) UNIQUE NOT NULL,
    original_url TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP,
    click_count INTEGER DEFAULT 0
);

CREATE TABLE clicks (
    id SERIAL PRIMARY KEY,
    url_id INTEGER REFERENCES urls(id) ON DELETE CASCADE,
    ip_address VARCHAR(45),
    user_agent TEXT,
    referrer TEXT,
    clicked_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_urls_slug ON urls(slug);
CREATE INDEX idx_clicks_url_id ON clicks(url_id);
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/shorten` | Create short URL |
| GET | `/:slug` | Redirect to original URL |
| GET | `/api/stats/:slug` | Get click stats |
| DELETE | `/api/:slug` | Delete short URL |
| GET | `/api/health` | Health check |

## Project Structure

```
linkas/
├── main.go              # Entry point, router setup
├── go.mod
├── go.sum
├── cmd/
│   └── migrate/         # Migration runner
│       └── main.go
├── internal/
│   ├── handler/         # HTTP handlers
│   │   └── url.go
│   ├── model/           # Data models
│   │   └── url.go
│   ├── repository/      # Database queries
│   │   └── url_repo.go
│   ├── service/         # Business logic
│   │   └── url_service.go
│   └── middleware/       # Rate limiting, etc.
│       └── rate_limit.go
├── migrations/          # SQL migrations
│   └── 001_init.sql
├── docker-compose.yml
├── Dockerfile
├── .env.example
└── README.md
```

## Slug Generation
- Use base62 encoding (a-z, A-Z, 0-9)
- Default length: 6 characters
- Collision check: regenerate if slug exists

## Rate Limiting
- 30 requests per minute per IP
- Return 429 Too Many Requests when exceeded

## Environment Variables

```env
PORT=8080
DATABASE_URL=postgres://user:pass@localhost:5432/linkas?sslmode=disable
BASE_URL=http://localhost:8080
```

## Quick Commands

```bash
# Copy env file
cp .env.example .env

# Run migrations
go run cmd/migrate/main.go up

# Check migration version
go run cmd/migrate/main.go version

# Rollback migration
go run cmd/migrate/main.go down

# Start API
go run main.go
```

## Phase 1 (MVP)
- [ ] Project setup (go mod, gin, docker)
- [ ] Database connection and migrations
- [ ] URL CRUD (create, read, delete)
- [ ] Redirect with click tracking
- [ ] Basic stats endpoint

## Phase 2 (Enhancement)
- [ ] Rate limiting middleware
- [ ] URL expiration
- [ ] Input validation
- [ ] Error handling
- [ ] Unit tests

## Phase 3 (Polish)
- [ ] Docker optimization
- [ ] README with examples
- [ ] CI/CD setup
- [ ] Performance testing
