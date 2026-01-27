.PHONY: help test test-coverage migrate migrate-up migrate-down migrate-clean build run clean

# Default target
help:
	@echo "Available targets:"
	@echo "  make test              - Run unit tests"
	@echo "  make test-coverage     - Run tests with coverage report"
	@echo "  make migrate           - Run all pending migrations"
	@echo "  make migrate-up       - Apply all up migrations"
	@echo "  make migrate-down     - Rollback last migration"
	@echo "  make migrate-clean     - Drop and recreate database"
	@echo "  make build             - Build the application"
	@echo "  make run               - Run the application"
	@echo "  make clean             - Clean build artifacts"

# Run unit tests
test:
	@echo "Running unit tests..."
	go test -v ./internal/... -timeout 30s

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v ./internal/... -coverprofile=coverage.out -covermode=atomic
	@echo "Coverage report generated: coverage.out"
	@echo "Open coverage report with: go tool cover -html=coverage.out"

# Run migrations
migrate: migrate-up

# Apply all up migrations
migrate-up:
	@echo "Running database migrations..."
	@if command -v migrate >/dev/null 2>&1; then \
		migrate -path ./infrastructure/database/migrations -database "postgresql://$${DB_USER}:$${DB_PASSWORD}@$${DB_HOST}:$${DB_PORT}/$${DB_NAME}?sslmode=disable" up; \
	else \
		echo "migrate tool not found. Installing..."; \
		go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
		migrate -path ./infrastructure/database/migrations -database "postgresql://$${DB_USER}:$${DB_PASSWORD}@$${DB_HOST}:$${DB_PORT}/$${DB_NAME}?sslmode=disable" up; \
	fi
	@echo "Migrations completed successfully!"

# Rollback last migration
migrate-down:
	@echo "Rolling back last migration..."
	migrate -path ./infrastructure/database/migrations -database "postgresql://$${DB_USER}:$${DB_PASSWORD}@$${DB_HOST}:$${DB_PORT}/$${DB_NAME}?sslmode=disable" down 1

# Drop and recreate database (for development only!)
migrate-clean:
	@echo "⚠️  WARNING: This will drop the database!"
	@read -p "Are you sure? (yes/no): " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		dropdb -h $${DB_HOST} -p $${DB_PORT} -U $${DB_USER} $${DB_NAME} || true; \
		createdb -h $${DB_HOST} -p $${DB_PORT} -U $${DB_USER} $${DB_NAME}; \
		make migrate-up; \
	else \
		echo "Cancelled."; \
	fi

# Build the application
build:
	@echo "Building application..."
	go build -o bin/api ./cmd/api
	@echo "Build completed: bin/api"

# Run the application
run:
	@echo "Starting application..."
	go run ./cmd/api

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f bin/api
	rm -f coverage.out
	@echo "Clean completed!"