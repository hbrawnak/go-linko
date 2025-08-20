# Go-Linko

A high-performance URL shortener service built with Go, featuring Redis caching and asynchronous background processing.

## Features

- **Fast URL Shortening** - Base62 encoded short codes
- **Redis Caching** - Hash-based caching for lightning-fast redirects
- **Async Processing** - Background workers for database persistence
- **Statistics API** - Comprehensive URL analytics with hit counts and timestamps
- **PostgreSQL Storage** - Reliable persistent data storage
- **Docker Support** - Full containerization with Docker Compose
- **Input Validation** - URL format and code validation
- **RESTful API** - Clean and simple API endpoints

## Quick Start

### Using Docker (Recommended)

```bash
# Clone the repository
git clone https://github.com/hbrawnak/go-linko.git
cd go-linko

# Start all services (PostgreSQL, Redis, and API)
docker-compose up --build
```

### Local Development

```bash
# Install dependencies
go mod tidy

# Set environment variables
export DSN="your_postgres_connection_string"
export REDIS_DSN="redis://localhost:6379"
export BASE_URL="http://localhost:8080"

# Run the application
go run ./cmd/api
```

The service will be available at `http://localhost:8080`

## Makefile Commands

The project includes a Makefile with convenient commands for development and deployment:

```bash
# Start Docker services in background
make up

# Stop Docker services
make down

# Build the application binary for Linux
make build_app

# Build and start Docker services (rebuilds if needed)
make up_build
```

## API Endpoints

### Shorten URL
```http
POST /shorten
Content-Type: application/json

{
  "url": "https://example.com"
}
```

**Response:**
```json
{
  "error": false,
  "message": "URL Shortened",
  "data": {
    "short_url": "http://localhost:8080/abc123",
    "code": "abc123"
  }
}
```

### Redirect to Original URL
```http
GET /{code}
```
Redirects to the original URL associated with the short code.

### Health Check
```http
GET /ping
```

**Response:**
```json
{
  "error": false,
  "message": "pong"
}
```

### Get URL Statistics
```http
GET /stats/{code}
```
Retrieve comprehensive analytics for a shortened URL.

**Response:**
```json
{
  "error": false,
  "message": "Stats Data",
  "data": {
    "code": "abc123",
    "count": 42,
    "update_at": "2025-08-20 13:45:30",
    "created_at": "2025-08-18 10:15:22",
    "original_url": "https://example.com"
  }
}
```

## Architecture

### Caching Strategy
- **Level 1**: Redis hash-based cache for immediate lookups
- **Level 2**: PostgreSQL database for persistent storage
- **Background Sync**: Async workers ensure data consistency

### Performance Features
- Hash-based Redis operations for faster cache access
- Background task queues to avoid blocking API responses
- Retry mechanisms with exponential backoff
- Connection pooling for database efficiency

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DSN` | PostgreSQL connection string | `host=postgres port=5432 user=postgres password=password dbname=users sslmode=disable timezone=UTC connect_timeout=5` |
| `REDIS_DSN` | Redis connection URL | `redis://redis:6379` |
| `BASE_URL` | Base URL for short links | `http://localhost:8080` |
| `PORT` | Server port | `8080` |

## Project Structure

```
├── cmd/
│   └── api/
│       └── main.go          # Application entry point and server setup
├── internal/
│   ├── data/                # Database models and operations
│   ├── database/            # Database clients (PostgreSQL, Redis)
│   ├── handlers/            # HTTP request handlers
│   │   └── handlers.go
│   ├── routes/              # Route setup and definitions
│   │   └── routes.go
│   ├── service/             # Business logic layer
│   ├── utils/               # Utility functions and helpers
│   └── worker/              # Background task workers
│       └── urlTaskWorker.go
├── db-data/                 # Database persistence (local development)
├── docker-compose.yaml      # Multi-container setup
├── Dockerfile              # Container image definition
├── Makefile                # Build and development commands
├── go.mod                  # Go module dependencies
├── go.sum                  # Go module checksums
└── urlShotenerApp          # Compiled binary
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
