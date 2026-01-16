.PHONY: help build run test docker-up docker-down migrate-up migrate-down lint security-scan

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	go build -o bin/api cmd/api/main.go

run: ## Run the application locally
	go run cmd/api/main.go

test: ## Run tests
	go test -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests with coverage report
	go tool cover -html=coverage.out

test-integration: ## Run integration tests
	go test -v ./tests/integration/...

docker-build: ## Build Docker image
	docker build -f docker/Dockerfile -t madabank-api:latest .

docker-up: ## Start all services with docker-compose
	docker-compose -f docker/docker-compose.yml up -d

docker-down: ## Stop all services
	docker-compose -f docker/docker-compose.yml down

docker-logs: ## View docker logs
	docker-compose -f docker/docker-compose.yml logs -f

migrate-up: ## Run database migrations up
	@echo "Running migrations..."
	go run cmd/migrate/main.go up

migrate-down: ## Rollback last migration
	go run cmd/migrate/main.go down

migrate-create: ## Create new migration (usage: make migrate-create name=create_users_table)
	@if [ -z "$(name)" ]; then echo "Error: name is required. Usage: make migrate-create name=your_migration_name"; exit 1; fi
	@echo "Creating migration: $(name)"
	@timestamp=$$(date +%Y%m%d%H%M%S); \
	touch migrations/$${timestamp}_$(name).up.sql; \
	touch migrations/$${timestamp}_$(name).down.sql; \
	echo "Created: migrations/$${timestamp}_$(name).up.sql"; \
	echo "Created: migrations/$${timestamp}_$(name).down.sql"

lint: ## Run linter
	golangci-lint run --timeout=5m

lint-fix: ## Run linter with auto-fix
	golangci-lint run --fix --timeout=5m

fmt: ## Format code
	go fmt ./...
	goimports -w .

vet: ## Run go vet
	go vet ./...

security-scan: ## Run security scanner (gosec)
	@if ! command -v gosec &> /dev/null; then \
		echo "gosec not found in PATH, checking $(go env GOPATH)/bin..."; \
		if [ -x "$(go env GOPATH)/bin/gosec" ]; then \
			export PATH=$$PATH:$$(go env GOPATH)/bin; \
		else \
			echo "Installing gosec..."; \
			go install github.com/securego/gosec/v2/cmd/gosec@latest; \
		fi \
	fi
	@export PATH=$$PATH:$$(go env GOPATH)/bin && gosec -fmt=json -out=gosec-report.json ./...

clean: ## Clean build artifacts
	rm -rf bin/ coverage.out coverage.html gosec-report.json

deps: ## Download dependencies
	go mod download
	go mod tidy

ci: lint vet test ## Run CI checks locally

.DEFAULT_GOAL := help