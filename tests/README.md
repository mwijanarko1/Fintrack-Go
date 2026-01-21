# Testing Documentation

## Overview

The fintrack-go project uses an enterprise-grade testing framework with comprehensive test coverage across multiple layers and types.

## Test Structure

```
tests/
├── testutil/              # Test utilities and helpers
├── testconfig/            # Test configuration management
├── integration/            # Integration tests with real database
├── e2e/                  # End-to-end workflow tests
├── load/                  # Load and stress testing
└── security/              # Security vulnerability testing

internal/
├── db/                    # Database layer unit tests
├── http/                  # HTTP handler unit tests (with mocks)
└── benchmarks/            # Performance benchmarks
```

## Test Types

### Unit Tests

**Purpose**: Test individual functions and methods in isolation using mocks

**Location**:
- `internal/db/*_test.go` - Database query tests with mocked connections
- `internal/http/*_handler_test.go` - HTTP handler tests with mocked DB
- `internal/validator/validator_test.go` - Validation tests (already 100% coverage)
- `internal/http/middleware_test.go` - Middleware tests
- `internal/http/handler_test.go` - Handler utility tests

**Running**:
```bash
# Run all unit tests
make test-unit

# Run specific package
go test -v ./internal/db
go test -v ./internal/http
go test -v ./internal/validator
```

**Coverage Goal**: >90% for all packages

### Integration Tests

**Purpose**: Test interaction between components with real database

**Location**:
- `tests/integration/api_test.go` - API endpoint integration tests
- `tests/integration/edge_cases_test.go` - Edge cases and error scenarios

**Running**:
```bash
# Run all integration tests
make test-integration

# Requires test database to be running
TEST_DATABASE_URL="postgres://..." make test-integration
```

**Setup**:
1. Start PostgreSQL: `make docker-up`
2. Run migrations: `make migrate`
3. Create test database: `createdb fintrack_test`
4. Run tests

### E2E Tests

**Purpose**: Test complete user workflows end-to-end

**Location**:
- `tests/e2e/workflow_test.go` - Complete user journeys

**Scenarios**:
- User lifecycle (create → use → verify)
- Category management (create → list → update → delete)
- Transaction tracking (create with category → list → summary)
- Summary generation (full analytics workflow)
- Multi-user isolation (data separation between users)

**Running**:
```bash
make test-e2e
```

### Load Tests

**Purpose**: Test system performance under concurrent and high-volume load

**Location**:
- `tests/load/load_test.go`

**Scenarios**:
- Concurrent requests (50+ simultaneous)
- Stress transaction creation (100+ transactions)
- Burst traffic handling
- Repeated summary generation

**Running**:
```bash
make test-load
```

**Performance Targets**:
- API response time < 200ms (p95)
- Handle 50+ concurrent requests
- 90%+ success rate under load

### Security Tests

**Purpose**: Identify and prevent security vulnerabilities

**Location**:
- `tests/security/security_test.go`

**Scenarios**:
- SQL injection attempts (all input fields)
- XSS prevention (error messages, responses)
- User isolation (cross-user data access prevention)
- UUID enumeration prevention

**Running**:
```bash
make test-security
```

### Benchmarks

**Purpose**: Measure performance of critical code paths

**Location**:
- `internal/benchmarks/validator_benchmark_test.go`

**Running**:
```bash
make test-bench
```

## Test Configuration

### Environment Variables

| Variable | Purpose | Default |
|----------|---------|---------|
| `TEST_DATABASE_URL` | Test database connection string | `postgres://fintrack:fintrack@localhost:5432/fintrack_test` |
| `TEST_TIMEOUT` | Test context timeout | `10s` |
| `VERBOSE_LOGGING` | Enable verbose test logging | `false` |
| `CLEANUP_TEST_DATA` | Auto-cleanup test data | `true` |

### CI/CD Configuration

**GitHub Actions**: `.github/workflows/test.yml`

Workflows run on:
- Push to main/develop
- Pull requests to main/develop

Jobs:
1. **Unit Tests** - Run with race detector
2. **Integration Tests** - With test database
3. **Security Tests** - Vulnerability scanning
4. **E2E Tests** - Complete workflow validation
5. **Cleanup** - Stop services after tests

## Test Utilities

### Fixtures

Test data factories for consistent test data:

```go
user := testutil.NewUserFixture()
category := testutil.NewCategoryFixture(userID)
transaction := testutil.NewTransactionFixture(userID, &categoryID)
```

### Assertions

Custom assertion helpers:

```go
testutil.AssertJSONResponse(t, w, http.StatusOK, expectedBody)
testutil.AssertErrorResponse(t, w, http.StatusBadRequest, "expected message")
testutil.AssertRowExists(t, pool, query, args...)
testutil.AssertRowCount(t, pool, expectedCount, query, args...)
```

### Test Server

Convenient test server wrapper:

```go
server := testutil.SetupTestServer(t)
defer server.Cleanup(t)

// Make requests
resp := server.PostJSON(t, "/api/v1/users", requestBody)
```

## Coverage Goals

| Layer | Target | Current |
|-------|--------|---------|
| Validation | 100% | 100% |
| Database | 90%+ | ~85% |
| HTTP Handlers | 80%+ | ~75% |
| API Endpoints | 100% | 100% |
| Overall | 85%+ | ~80% |

## Running Tests

### Quick Start

```bash
# 1. Start database
make docker-up

# 2. Run migrations
make migrate

# 3. Run tests
make test

# 4. View coverage
make test-coverage
open coverage.html
```

### Running Specific Test Types

```bash
# Unit tests only
make test-unit

# Integration tests only
make test-integration

# Security tests only
make test-security

# E2E tests only
make test-e2e

# Load tests only
make test-load

# Benchmarks only
make test-bench
```

### Running Specific Tests

```bash
# Run specific test file
go test -v ./internal/db/users_test.go

# Run specific test
go test -v ./internal/db/users_test.go -run TestCreateUser

# Run with race detector
go test -race -v ./internal/db

# Run with coverage
go test -v -cover ./internal/validator
```

## Debugging Tests

### Verbose Mode

```bash
VERBOSE_LOGGING=true make test
```

### Run Tests with Debugging

```bash
# Run single test with debug output
go test -v ./internal/db -run TestCreateUser/success -count=1

# Keep test database running
make docker-up
make migrate

# Run tests repeatedly
watchexec -w './internal/...' 'go test ./...'
```

## Continuous Integration

### GitHub Actions

All tests run automatically on:
- Push to `main` or `develop` branches
- Pull requests to `main` or `develop`

### Local CI Simulation

```bash
# Simulate full CI run locally
make ci-test
```

## Test Data Management

### Automatic Cleanup

Tests automatically clean up after completion using:

```go
defer testutil.TeardownTestDB(t, pool)
defer server.Cleanup(t)
```

### Manual Cleanup

```bash
# Reset test database
./scripts/rollback.sh
make migrate
```

## Troubleshooting

### Tests Failing with Database Connection

```bash
# Check if PostgreSQL is running
make docker-logs

# Verify connection string
echo $TEST_DATABASE_URL

# Test connection manually
psql $TEST_DATABASE_URL
```

### Tests Failing with Port Already in Use

```bash
# Check what's using port 5432
lsof -i :5432

# Stop and restart
make docker-down
make docker-up
```

### Tests Timing Out

```bash
# Increase timeout
TEST_TIMEOUT=30s make test

# Run without timeout
go test -v ./tests/integration/...
```

## Best Practices

1. **Use Test Tables** - For multiple test cases with similar logic
2. **Subtest Naming** - Use descriptive names following Go conventions
3. **Parallel Tests** - Use `t.Parallel()` for independent tests
4. **Cleanup** - Always clean up resources in `defer` statements
5. **Test Helpers** - Use fixtures and utilities to reduce duplication
6. **Assertions** - Use descriptive assertion messages
7. **Race Conditions** - Run with `-race` flag in CI
8. **Coverage** - Aim for high coverage without testing implementation details

## Adding New Tests

### Unit Test Template

```go
func TestFunction_Scenario(t *testing.T) {
    t.Run("success", func(t *testing.T) {
        ctx := testutil.CreateTestContext(t)
        // Test logic
    })

    t.Run("error case", func(t *testing.T) {
        ctx := testutil.CreateTestContext(t)
        // Test error handling
    })
}
```

### Integration Test Template

```go
func TestEndpoint_CompleteFlow(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    server := testutil.SetupTestServer(t)
    defer server.Cleanup(t)

    // Setup data
    userResp := server.PostJSON(t, "/api/v1/users", userReq)
    
    // Test behavior
    assert.Equal(t, http.StatusCreated, userResp.StatusCode)
    
    // Cleanup happens automatically via defer
}
```

## Test Statistics

**Total Test Count**: ~300+ test cases

**Breakdown**:
- Unit Tests: ~150
- Integration Tests: ~50
- E2E Tests: ~20
- Security Tests: ~40
- Load Tests: ~15
- Benchmarks: ~10

**Estimated Runtime**: ~10 minutes (full suite)

## Additional Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Testify Assertions](https://pkg.go.dev/github.com/stretchr/testify/assert)
- [Go Race Detector](https://go.dev/doc/race)
- [Test Coverage](https://go.dev/blog/cover)
