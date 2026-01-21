.PHONY: help run build test test-coverage test-unit test-integration test-security test-e2e test-load test-bench migrate migrate-rollback docker-up docker-down docker-logs clean lint

help:
	@echo "Available commands:"
	@echo "  make run            - Run server"
	@echo "  make build          - Build application"
	@echo "  make test           - Run all tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make test-unit      - Run only unit tests"
	@echo "  make test-integration  - Run integration tests"
	@echo "  make test-security   - Run security tests"
	@echo "  make test-e2e       - Run E2E tests"
	@echo "  make test-load       - Run load tests"
	@echo "  make test-bench      - Run benchmarks"
	@echo "  make migrate        - Run database migrations"
	@echo "  make migrate-rollback  - Rollback database migrations"
	@echo "  make docker-up      - Start PostgreSQL with Docker"
	@echo "  make docker-down    - Stop PostgreSQL with Docker"
	@echo "  make docker-logs    - Show PostgreSQL logs"
	@echo "  make lint           - Run linter"
	@echo "  make clean          - Clean build artifacts"

run:
	go run cmd/server/main.go

build:
	go build -o bin/fintrack cmd/server/main.go

test:
	go test -v -race ./...

test-coverage:
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@go tool cover -func=coverage.out | grep total

test-unit:
	go test -v ./internal/... -short

test-integration:
	TEST_DATABASE_URL=$(TEST_DATABASE_URL) go test -v ./tests/integration/... -timeout 10m

test-security:
	TEST_DATABASE_URL=$(TEST_DATABASE_URL) go test -v ./tests/security/... -timeout 10m

test-e2e:
	TEST_DATABASE_URL=$(TEST_DATABASE_URL) go test -v ./tests/e2e/... -timeout 10m

test-load:
	TEST_DATABASE_URL=$(TEST_DATABASE_URL) go test -v ./tests/load/... -timeout 10m

test-bench:
	go test -bench=. -benchmem -run=^$$ ./...

migrate:
	@echo "Running migrations..."
	psql $$DATABASE_URL -f sql/migrations/001_init.sql
	psql $$DATABASE_URL -f sql/migrations/002_indexes.sql
	@echo "Migrations completed"

migrate-rollback:
	./scripts/rollback.sh

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f postgres

clean:
	rm -rf bin/ coverage.out coverage.html

lint:
	golangci-lint run ./...

.PHONY: ci-test
ci-test: test-integration test-security test-e2e
