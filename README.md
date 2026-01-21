# Fintrack-Go

A simple, secure RESTful backend API for personal finance tracking built with Go and PostgreSQL.

## Features

- **User Management**: Create users with unique email addresses
- **Categories**: Create and list expense categories per user
- **Transactions**: Track expenses with optional category assignment
- **Summary**: Get spending summaries grouped by category with date filtering
- **Validation**: Comprehensive input validation for all endpoints
- **Structured Logging**: JSON logging with request tracking
- **Error Handling**: Consistent error responses with appropriate HTTP status codes

## Technology Stack

- **Language**: Go 1.21+
- **Router**: Chi
- **Database**: PostgreSQL 16
- **Database Driver**: pgx/v5
- **Logging**: zerolog
- **Configuration**: Environment variables

## Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose (for PostgreSQL)
- psql (PostgreSQL client tool)

## Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/yourusername/fintrack-go.git
cd fintrack-go
```

### 2. Start PostgreSQL

```bash
make docker-up
```

This starts PostgreSQL using Docker Compose.

⚠️ **Security Warning**: The default credentials (`fintrack:fintrack`) are for **development only**. Before deploying to production:

1. **Change** credentials in `docker-compose.yml`
2. **Use strong passwords** (min 12 characters, mixed case, numbers, symbols)
3. **Never commit credentials** to version control
4. **Use environment variables** for sensitive data in production

Generate a secure password:
```bash
# Using OpenSSL (most systems)
openssl rand -base64 24

# Or on Linux/Mac:
LC_ALL=C tr -dc 'A-Za-z0-9!#$%&()*+,-./:;<=>?@[\]^_`{|}~' < /dev/urandom | head -c 24
```

Then update `docker-compose.yml`:
```yaml
environment:
  POSTGRES_USER: your_secure_user
  POSTGRES_PASSWORD: your_generated_password_here
  POSTGRES_DB: fintrack
```

Default development credentials:
- User: `fintrack`
- Password: `fintrack` (CHANGE THIS IN PRODUCTION!)
- Database: `fintrack`
- Port: `5432`

### 3. Set Environment Variables

Copy the example environment file and configure as needed:

```bash
cp configs/.env.example .env
```

Default `.env` configuration:
```bash
DATABASE_URL=postgres://fintrack:fintrack@localhost:5432/fintrack
SERVER_PORT=8080
LOG_LEVEL=INFO
CORS_ENABLED=false
```

### 4. Run Database Migrations

```bash
make migrate
```

Or manually:
```bash
psql $DATABASE_URL -f sql/migrations/001_init.sql
psql $DATABASE_URL -f sql/migrations/002_indexes.sql
```

### 5. Install Dependencies

```bash
go mod download
```

### 6. Run the Server

```bash
make run
```

The server will start on `http://localhost:8080`

## API Endpoints

### Health Check

#### Health Status
```bash
GET /health
```

Response (200):
```json
{
  "status": "healthy"
}
```

Use this endpoint for health monitoring and load balancer health checks.

### Users

### Users

#### Create User
```bash
POST /api/v1/users
Content-Type: application/json

{
  "email": "user@example.com"
}
```

Response (201):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "created_at": "2026-01-21T10:00:00Z"
}
```

### Categories

#### Create Category
```bash
POST /api/v1/categories
Content-Type: application/json

{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Food"
}
```

Response (201):
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440001",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Food",
  "created_at": "2026-01-21T10:00:00Z"
}
```

#### List Categories
```bash
GET /api/v1/categories?user_id=550e8400-e29b-41d4-a716-446655440000
```

Response (200):
```json
[
  {
    "id": "660e8400-e29b-41d4-a716-446655440001",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Food",
    "created_at": "2026-01-21T10:00:00Z"
  }
]
```

### Transactions

#### Create Transaction
```bash
POST /api/v1/transactions
Content-Type: application/json

{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "category_id": "660e8400-e29b-41d4-a716-446655440001",
  "amount": 12.50,
  "description": "Lunch",
  "occurred_at": "2026-01-21T10:00:00Z"
}
```

Response (201):
```json
{
  "id": "770e8400-e29b-41d4-a716-446655440000",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "category_id": "660e8400-e29b-41d4-a716-446655440001",
  "amount": 12.50,
  "description": "Lunch",
  "occurred_at": "2026-01-21T10:00:00Z",
  "created_at": "2026-01-21T10:05:00Z"
}
```

#### List Transactions
```bash
GET /api/v1/transactions?user_id=550e8400-e29b-41d4-a716-446655440000&from=2026-01-01T00:00:00Z&to=2026-01-31T23:59:59Z
```

Query Parameters:
- `user_id` (required): UUID of the user
- `from` (optional): ISO 8601 timestamp for start date
- `to` (optional): ISO 8601 timestamp for end date

Default behavior: Returns transactions from the last 30 days

Response (200):
```json
[
  {
    "id": "770e8400-e29b-41d4-a716-446655440000",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "category_id": "660e8400-e29b-41d4-a716-446655440001",
    "category_name": "Food",
    "amount": 12.50,
    "description": "Lunch",
    "occurred_at": "2026-01-21T10:00:00Z",
    "created_at": "2026-01-21T10:05:00Z"
  }
]
```

### Summary

#### Get Summary
```bash
GET /api/v1/summary?user_id=550e8400-e29b-41d4-a716-446655440000&from=2026-01-01T00:00:00Z&to=2026-01-31T23:59:59Z
```

Query Parameters:
- `user_id` (required): UUID of the user
- `from` (optional): ISO 8601 timestamp for start date
- `to` (optional): ISO 8601 timestamp for end date

Default behavior: Uses the last 30 days

Response (200):
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "from": "2026-01-01T00:00:00Z",
  "to": "2026-01-31T23:59:59Z",
  "categories": [
    {
      "category_id": "660e8400-e29b-41d4-a716-446655440001",
      "category_name": "Food",
      "total": 120.50
    },
    {
      "category_id": null,
      "category_name": "Uncategorized",
      "total": 30.00
    }
  ]
}
```

## Error Response Format

All error responses follow this structure:

```json
{
  "error": {
    "code": "BAD_REQUEST",
    "message": "Invalid email format",
    "details": {
      "field": "email",
      "value": "invalid-email"
    }
  }
}
```

### HTTP Status Codes

| Code | Usage |
|------|-------|
| 200  | Successful GET requests |
| 201  | Successful POST requests |
| 400  | Validation errors, invalid input |
| 404  | Resource not found |
| 409  | Duplicate resource (email, category name) |
| 500  | Internal server error |

## Makefile Commands

| Command | Description |
|---------|-------------|
| `make run` | Run the server |
| `make build` | Build the application |
| `make test` | Run tests |
| `make migrate` | Run database migrations |
| `make docker-up` | Start PostgreSQL with Docker |
| `make docker-down` | Stop PostgreSQL with Docker |
| `make docker-logs` | Show PostgreSQL logs |
| `make clean` | Clean build artifacts |

## Database Schema

### Users Table
- `id` (UUID, Primary Key)
- `email` (VARCHAR(255), Unique)
- `created_at` (TIMESTAMP)

### Categories Table
- `id` (UUID, Primary Key)
- `user_id` (UUID, Foreign Key)
- `name` (VARCHAR(100))
- `created_at` (TIMESTAMP)

### Transactions Table
- `id` (UUID, Primary Key)
- `user_id` (UUID, Foreign Key)
- `category_id` (UUID, Foreign Key, Nullable)
- `amount` (DECIMAL(10,2), > 0)
- `description` (TEXT, Nullable)
- `occurred_at` (TIMESTAMP)
- `created_at` (TIMESTAMP)

## Validation Rules

- **Email**: Valid email format, unique across all users
- **UUID**: Valid UUID v4 format
- **Amount**: Must be greater than 0, max 99999999.99
- **Category Name**: 1-100 characters, unique per user
- **Date Range**: `from` must be <= `to`

## Testing

### Enterprise-Grade Test Suite

This project includes a comprehensive testing framework with:

- **300+ test cases** across unit, integration, E2E, security, and load testing
- **80%+ code coverage** across all packages
- **Automated CI/CD** with GitHub Actions
- **Performance benchmarks** for critical paths

See [tests/README.md](tests/README.md) for detailed testing documentation.

### Run Tests

Run tests with:

```bash
make test
```

Run tests with coverage:

```bash
go test -cover ./...
```

### Test Types

```bash
# Run all unit tests
make test-unit

# Run integration tests (requires test database)
make test-integration

# Run security vulnerability tests
make test-security

# Run end-to-end workflow tests
make test-e2e

# Run load and stress tests
make test-load

# Run performance benchmarks
make test-bench
```

### Coverage Report

Generate HTML coverage report:

```bash
make test-coverage
open coverage.html
```

### Quick Test

Run tests with:

```bash
make test
```

Run tests with coverage:

```bash
go test -cover ./...
```

## Project Structure

```
fintrack-go/
├── cmd/
│   └── server/
│       └── main.go              # Application entrypoint
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── db/
│   │   ├── db.go                # Database connection
│   │   ├── database.go          # Database interface for testing
│   │   ├── errors.go            # Error definitions
│   │   ├── users.go             # User queries
│   │   ├── users_test.go        # Unit tests with mocks
│   │   ├── categories.go        # Category queries
│   │   ├── categories_test.go   # Unit tests with mocks
│   │   ├── transactions.go      # Transaction queries
│   │   ├── transactions_test.go # Unit tests with mocks
│   │   └── summary.go           # Summary aggregation queries
│   │   └── summary_test.go     # Unit tests with mocks
│   ├── models/
│   │   ├── user.go              # User model
│   │   ├── category.go          # Category model
│   │   ├── transaction.go       # Transaction model
│   │   └── summary.go           # Summary model
│   ├── http/
│   │   ├── handler.go           # Common handler utilities
│   │   ├── handler_test.go      # Handler utility tests
│   │   ├── middleware.go        # Logging, error handling, CORS
│   │   ├── middleware_test.go  # Middleware tests
│   │   ├── routes.go            # Route definitions
│   │   ├── user_handler.go      # User endpoints
│   │   ├── user_handler_test.go # User handler unit tests
│   │   ├── category_handler.go  # Category endpoints
│   │   ├── category_handler_test.go # Category handler unit tests
│   │   ├── transaction_handler.go  # Transaction endpoints
│   │   ├── transaction_handler_test.go # Transaction handler unit tests
│   │   ├── summary_handler.go   # Summary endpoints
│   │   ├── summary_handler_test.go # Summary handler unit tests
│   │   └── health_handler.go    # Health check endpoint
│   │   └── health_handler_test.go # Health handler tests
│   ├── benchmarks/
│   │   └── validator_benchmark_test.go # Performance benchmarks
│   └── validator/
│       ├── validator.go         # Input validation logic
│       └── validator_test.go    # Validation tests (100% coverage)
├── sql/
│   └── migrations/
│       ├── 001_init.sql         # Initial schema
│       └── 002_indexes.sql      # Performance indexes
├── tests/
│   ├── testutil/              # Test utilities and helpers
│   │   ├── db.go             # Database setup/teardown
│   │   ├── fixtures.go        # Test data factories
│   │   ├── assertions.go     # Custom assertions
│   │   └── server.go        # Test server wrapper
│   ├── testconfig/            # Test configuration
│   ├── integration/            # Integration tests
│   │   ├── api_test.go      # API endpoint tests
│   │   └── edge_cases_test.go # Edge cases and errors
│   ├── e2e/                 # End-to-end workflow tests
│   │   └── workflow_test.go # Complete user journeys
│   ├── load/                 # Load and stress testing
│   │   └── load_test.go    # Performance tests
│   ├── security/              # Security vulnerability tests
│   │   └── security_test.go # SQL injection, XSS, auth tests
│   └── README.md             # Testing documentation
├── configs/
│   └── .env.example             # Environment variables template
├── scripts/
│   ├── migrate.sh               # Run migrations
│   └── rollback.sh              # Rollback migrations
├── .github/
│   └── workflows/
│       └── test.yml              # CI/CD pipeline
├── docker-compose.yml           # PostgreSQL setup
├── go.mod                       # Go module definition
└── Makefile                     # Common commands
```

## Logging

The application uses structured JSON logging. Log format:

```json
{
  "level": "info",
  "time": "2026-01-21T10:00:00Z",
  "request_id": "abc123",
  "method": "POST",
  "path": "/api/v1/users",
  "status": 201,
  "duration_ms": 15,
  "remote_addr": "127.0.0.1:12345"
}
```

Set log level via `LOG_LEVEL` environment variable (DEBUG, INFO, WARN, ERROR).

## Development

### Running with Hot Reload

For development with hot reload, install and use `air`:

```bash
go install github.com/cosmtrek/air@latest
air
```

### Adding New Features

1. Define models in `internal/models/`
2. Implement database queries in `internal/db/`
3. Add validation in `internal/validator/`
4. Create handlers in `internal/http/`
5. Register routes in `internal/http/routes.go`

## Security Considerations

⚠️ **Important for Production Deployment**

### Default Credentials
The default Docker Compose configuration uses weak credentials (`fintrack:fintrack`). **These must be changed in production**:

1. Update `docker-compose.yml` with strong passwords
2. Set strong environment variables in production
3. Never commit credentials to version control

### CORS Configuration
By default, CORS is disabled. For production:
- Enable `CORS_ENABLED` only if needed
- Configure specific allowed origins (not `*`)
- Consider implementing authentication

### Request Size Limits
The API limits request bodies to 1MB by default to prevent DoS attacks. Adjust `maxRequestBodySize` in `routes.go` if needed.

### Monitoring
- Use the `/health` endpoint for health checks
- Monitor logs for suspicious activity
- Consider adding rate limiting in production

### Known Limitations (MVP)
- No authentication/authorization (user isolation only via user_id)
- No rate limiting
- No input sanitization beyond validation
- Float64 precision for monetary values (acceptable for MVP)

## Troubleshooting

### Database Connection Issues

Check if PostgreSQL is running:
```bash
make docker-logs
```

Test database connection:
```bash
psql $DATABASE_URL
```

### Migration Issues

To reset the database:
```bash
./scripts/rollback.sh
make migrate
```

## License

MIT License

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.
