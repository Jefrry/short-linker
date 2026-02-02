# Short Linker

A production-ready URL shortener service built with Go.

## Features

- **URL Shortening** - Convert long URLs into short, shareable links
- **User Authentication** - JWT-based authentication with signup/signin/signout
- **User Link Management** - Users can view and delete their own links
- **Batch Operations** - Create or delete multiple links in a single request
- **Automatic Migrations** - Database schema managed via golang-migrate
- **Gzip Compression** - Transparent request/response compression
- **Swagger Documentation** - Interactive API documentation
- **Health Check** - Built-in endpoint for monitoring
- **Structured Logging** - Production-grade logging with Zap
- **Clean Architecture** - Layered design with handlers, services, and repositories

## Tech Stack

- **Go 1.25+**
- **Chi** - Lightweight HTTP router
- **PostgreSQL** - Primary database
- **JWT** - Token-based authentication
- **golang-migrate** - Database migrations
- **Zap** - Structured logging
- **Swaggo** - Swagger documentation generator
- **Docker Compose** - Container orchestration

## Quick Start

### Prerequisites

- Go 1.25 or higher
- PostgreSQL 14+ (or Docker)
- Make (optional)

### 1. Clone the repository

```bash
git clone https://github.com/yourusername/short-linker.git
cd short-linker
```

### 2. Set up the database

Using Docker Compose:

```bash
# Create .env file from example
cp .example.env .env

# Edit .env with your database credentials
# DB_USER=your_user
# DB_PASSWORD=your_password
# DB_NAME=short_linker

# Start PostgreSQL
docker-compose up -d
```

Or use an existing PostgreSQL instance.

### 3. Configure the application

The application can be configured via environment variables or command-line flags. Environment variables take precedence over flags.

| Environment Variable | Flag | Default | Description |
|---------------------|------|---------|-------------|
| `SERVER_ADDRESS` | `-a` | `localhost:8080` | Server address and port |
| `BASE_URL` | `-b` | `http://localhost:8080` | Base URL for shortened links |
| `DATABASE_DSN` | `-d` | (required) | PostgreSQL connection string |
| `JWT_SECRET` | `-j` | `super-secret-key` | Secret key for JWT signing |

Example DSN format:
```
postgres://user:password@localhost:5434/short_linker?sslmode=disable
```

### 4. Run the application

```bash
# Install dependencies
go mod download

# Run the server
go run cmd/server/main.go \
  -d "postgres://user:password@localhost:5434/short_linker?sslmode=disable" \
  -j "your-secure-jwt-secret"
```

Or using environment variables:

```bash
export DATABASE_DSN="postgres://user:password@localhost:5434/short_linker?sslmode=disable"
export JWT_SECRET="your-secure-jwt-secret"
go run cmd/server/main.go
```

The server will start and automatically run database migrations.

## Development

### Generate Swagger Documentation

```bash
go generate ./cmd/server/main.go
```

Or manually:

```bash
swag init -g cmd/server/main.go --parseInternal -d . -o docs
```

### Run Tests

```bash
go test ./...
```

### Build for Production

```bash
go build -ldflags="-s -w" -o short-linker ./cmd/server
```

## Database Schema

The application uses PostgreSQL with the following main tables:

**links**
- `id` - Short link identifier (primary key)
- `original_url` - Original URL
- `user_id` - Owner user ID (optional)
- `deleted` - Soft delete flag
- `created_at` - Creation timestamp

**users**
- `id` - User ID (auto-increment)
- `name` - User name
- `email` - User email (unique)
- `password` - Hashed password
- `created_at` - Registration timestamp

## Security Considerations

- Passwords are hashed using bcrypt
- JWT tokens are used for stateless authentication
- Session tokens are stored in HTTP-only cookies
- Input validation on all endpoints
- SQL injection protection via parameterized queries

## Production Deployment

For production deployment, ensure:

1. Set a strong `JWT_SECRET` (at least 32 characters)
2. Use HTTPS with a reverse proxy (nginx, Caddy)
3. Configure proper database connection pooling
4. Set appropriate resource limits
5. Enable monitoring and alerting

Example production environment variables:

```bash
SERVER_ADDRESS=0.0.0.0:8080
BASE_URL=https://your-domain.com
DATABASE_DSN=postgres://user:pass@db-host:5432/short_linker?sslmode=require
JWT_SECRET=your-very-long-and-secure-secret-key-here
```

## License

This project is open source and available under the [MIT License](LICENSE).
