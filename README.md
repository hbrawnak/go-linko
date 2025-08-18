# Go-Linko

A high-performance URL shortener service built with Go, featuring Redis caching and asynchronous background processing.

## Features

- **Fast URL Shortening** - Base62 encoded short codes
- **Redis Caching** - Hash-based caching for lightning-fast redirects
- **Async Processing** - Background workers for database persistence
- **Analytics** - Hit count tracking for shortened URLs
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

## Future Modifications

### Stats API Implementation
**Planned Feature**: Comprehensive analytics and statistics tracking for shortened URLs.

#### Implementation Plan:
1. **New Database Table**: Create a dedicated stats table for time-based logging
   - Track individual URL access events with timestamps
   - Store metadata like referrer, user agent, IP address (anonymized)
   - Index by short code and timestamp for efficient queries

2. **Stats API Endpoints**:
   ```http
   GET /stats/{code}
   ```
   - Return comprehensive statistics for a specific short code
   - Include metrics like total clicks, daily/weekly/monthly trends
   - Provide time-series data for visualization


#### Technical Considerations:
- Use time-series optimized database schema
- Implement efficient aggregation queries
- Add caching layer for frequently requested stats
- Consider data privacy and anonymization requirements

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
