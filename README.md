# Go-Linko

A simple and fast URL shortener service built with Go.

## Features

- Fast URL shortening
- PostgreSQL database storage
- RESTful API
- Docker support
- Health check endpoint

## Quick Start

### Using Docker (Recommended)

```bash
# Clone the repository
git clone https://github.com/hbrawnak/go-linko.git
cd go-linko

# Start the application with Docker Compose
docker-compose up --build
```

### Local Development

```bash
# Install dependencies
go mod tidy

# Run the application
go run ./cmd/api
```

The service will be available at `http://localhost:8080`

## API Endpoints

### Welcome
```
GET /
```

### Shorten URL
```
POST /shorten
Content-Type: application/json

{
  "url": "https://example.com"
}
```

### Health Check
```
GET /ping
```

## Environment Variables

- `DSN`: PostgreSQL connection string
- Default: `"host=postgres port=5432 user=postgres password=password dbname=users sslmode=disable timezone=UTC connect_timeout=5"`

## Project Structure

```
├── cmd/api/          # Application entry point
├── data/             # Database models and operations
├── internal/service/ # Business logic layer
├── docker-compose.yaml
├── Dockerfile
└── Makefile
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
