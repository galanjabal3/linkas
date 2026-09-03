# Linkas

A simple, fast URL shortener API built with Go + Gin + PostgreSQL.

## Features

- **Shorten URLs** — Create short URLs with custom or auto-generated slugs
- **Path-based Slugs** — Support nested paths like `video/01`
- **QR Code** — Generate QR code images for short URLs
- **Redirect** — Fast 301 redirects to original URLs
- **Analytics** — Track clicks with IP, user-agent, and referrer
- **Rate Limiting** — 30 requests per minute per IP
- **URL Expiration** — Set TTL for short links with specific error handling
- **Swagger Docs** — Interactive API documentation

## Tech Stack

- Go 1.26+
- Gin (HTTP framework)
- PostgreSQL (database)
- sqlx (database toolkit)
- go-qrcode (QR code generation)
- Swaggo (Swagger docs)

## Quick Start

### 1. Clone & Setup

```bash
git clone https://github.com/galanjabal3/linkas.git
cd linkas
cp .env.example .env
```

### 2. Start PostgreSQL

```bash
# Option A: Docker Compose (recommended)
docker-compose up -d postgres

# Option B: Manual Docker
docker run -d --name linkas-db -p 5432:5432 -e POSTGRES_DB=linkas postgres:16-alpine
```

### 3. Run Migrations

```bash
go run cmd/migrate/main.go up
```

### 4. Start API

```bash
go run main.go
```

### Full Stack with Docker

```bash
docker-compose up -d
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/shorten` | Create short URL |
| GET | `/:slug` | Redirect to original URL |
| GET | `/api/stats/:slug` | Get click analytics |
| GET | `/api/qrcode/:slug` | Generate QR code image |
| DELETE | `/api/:slug` | Delete short URL |
| GET | `/api/health` | Health check |

## Usage Examples

### Shorten a URL

```bash
curl -X POST http://localhost:8080/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://github.com", "expires_in": "24h"}'
```

Response:
```json
{
  "slug": "aB3xYz",
  "short_url": "http://localhost:8080/aB3xYz",
  "original_url": "https://github.com",
  "expires_at": "2026-08-06T10:00:00Z"
}
```

### Shorten with Custom Slug

```bash
curl -X POST http://localhost:8080/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://github.com", "custom_slug": "my-link", "expires_in": "168h"}'
```

### Get QR Code

```bash
curl -o qr.png http://localhost:8080/api/qrcode/aB3xYz
```

Returns PNG image (256x256).

### Get Stats

```bash
curl http://localhost:8080/api/stats/aB3xYz
```

Response:
```json
{
  "slug": "aB3xYz",
  "original_url": "https://github.com",
  "click_count": 42,
  "created_at": "2026-08-05T10:00:00Z",
  "recent_clicks": [...]
}
```

### Delete URL

```bash
curl -X DELETE http://localhost:8080/api/aB3xYz
```

## Error Responses

| Status | Error | Description |
|--------|-------|-------------|
| 400 | `invalid_request` | Invalid URL or request body |
| 404 | `not_found` | URL not found |
| 409 | `slug_exists` | Custom slug already taken |
| 410 | `url_expired` | URL has expired |
| 429 | `rate_limit` | Too many requests |

## Project Structure

```
linkas/
├── main.go              # Entry point
├── cmd/
│   └── migrate/         # Migration runner
├── internal/
│   ├── handler/         # HTTP handlers
│   ├── model/           # Data models
│   ├── repository/      # Database queries
│   ├── service/         # Business logic
│   └── middleware/       # Rate limiting
├── migrations/          # SQL migrations
├── docker-compose.yml
├── Dockerfile
└── .env.example
```

## Migrations

```bash
# Apply all migrations
go run cmd/migrate/main.go up

# Rollback last migration
go run cmd/migrate/main.go down

# Check current version
go run cmd/migrate/main.go version
```

## License

MIT
