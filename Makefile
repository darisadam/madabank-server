.PHONY: help build run test test-coverage docker-up docker-down docker-logs migrate-up migrate-down migrate-create lint fmt clean deps install-tools dev

# Variables
BINARY_NAME=api
DOCKER_COMPOSE_FILE=docker/docker-compose.yml
MIGRATIONS_DIR=migrations
CMD_API=cmd/api/main.go
CMD_MIGRATE=cmd/migrate/main.go

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build the application binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	go build -o bin/$(BINARY_NAME) $(CMD_API)
	@echo "Build complete: bin/$(BINARY_NAME)"

run: ## Run the application locally
	@echo "Running application..."
	go run $(CMD_API)

dev: ## Run with auto-reload (requires air: go install github.com/air-verse/air@latest)
	@command -v air > /dev/null 2>&1 || { echo "air not found. Install with: go install github.com/air-verse/air@latest"; exit 1; }
	air

test: ## Run all tests
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Generate and open coverage report
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out

test-short: ## Run tests without race detector (faster)
	go test -v -short ./...

docker-up: ## Start all services with docker-compose
	@test -f $(DOCKER_COMPOSE_FILE) || { echo "Error: $(DOCKER_COMPOSE_FILE) not found"; exit 1; }
	docker-compose -f $(DOCKER_COMPOSE_FILE) up -d
	@echo "Services started. Use 'make docker-logs' to view logs"

docker-down: ## Stop all services
	@test -f $(DOCKER_COMPOSE_FILE) || { echo "Error: $(DOCKER_COMPOSE_FILE) not found"; exit 1; }
	docker-compose -f $(DOCKER_COMPOSE_FILE) down

docker-logs: ## View docker logs (follow mode)
	@test -f $(DOCKER_COMPOSE_FILE) || { echo "Error: $(DOCKER_COMPOSE_FILE) not found"; exit 1; }
	docker-compose -f $(DOCKER_COMPOSE_FILE) logs -f

docker-restart: docker-down docker-up ## Restart all docker services

migrate-up: ## Run database migrations up
	@echo "Running migrations up..."
	@test -f $(CMD_MIGRATE) || { echo "Error: Migration tool not found at $(CMD_MIGRATE)"; exit 1; }
	go run $(CMD_MIGRATE) up

migrate-down: ## Rollback last migration
	@echo "Rolling back last migration..."
	@test -f $(CMD_MIGRATE) || { echo "Error: Migration tool not found at $(CMD_MIGRATE)"; exit 1; }
	go run $(CMD_MIGRATE) down

migrate-create: ## Create new migration (usage: make migrate-create name=create_users_table)
	@if [ -z "$(name)" ]; then \
		echo "Error: name is required"; \
		echo "Usage: make migrate-create name=your_migration_name"; \
		exit 1; \
	fi
	@mkdir -p $(MIGRATIONS_DIR)
	@timestamp=$$(date -u +%Y%m%d%H%M%S 2>/dev/null || date +%Y%m%d%H%M%S); \
	up_file="$(MIGRATIONS_DIR)/$${timestamp}_$(name).up.sql"; \
	down_file="$(MIGRATIONS_DIR)/$${timestamp}_$(name).down.sql"; \
	touch "$$up_file" "$$down_file"; \
	echo "-- Write your UP migration here" > "$$up_file"; \
	echo "-- Write your DOWN migration here" > "$$down_file"; \
	echo "✓ Created: $$up_file"; \
	echo "✓ Created: $$down_file"

lint: ## Run golangci-lint
	@command -v golangci-lint > /dev/null 2>&1 || { echo "golangci-lint not found. Install: https://golangci-lint.run/usage/install/"; exit 1; }
	golangci-lint run ./...

lint-fix: ## Run linter and auto-fix issues
	@command -v golangci-lint > /dev/null 2>&1 || { echo "golangci-lint not found. Install: https://golangci-lint.run/usage/install/"; exit 1; }
	golangci-lint run --fix ./...

fmt: ## Format all Go code
	@echo "Formatting code..."
	go fmt ./...
	@echo "Code formatted"

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

clean: ## Clean build artifacts and cache
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out
	go clean -cache -testcache
	@echo "Clean complete"

deps: ## Download and tidy dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy
	@echo "Dependencies updated"

deps-upgrade: ## Upgrade all dependencies
	@echo "Upgrading dependencies..."
	go get -u ./...
	go mod tidy

install-tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/air-verse/air@latest
	@echo "Tools installed"

check: fmt vet lint test ## Run all checks (format, vet, lint, test)

.DEFAULT_GOAL := help