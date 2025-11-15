SHELL := /bin/bash

APP_NAME := user-services
PKG := ./...
BUILD_DIR := bin
COVERAGE_DIR := coverage

.PHONY: run build clean tidy test fmt lint vet migrate compose-up compose-down compose-logs deps check

# Default target
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development targets
run: ## Run the application in development mode
	go run ./cmd/server

run-dev: ## Run with hot reload using air (requires air)
	@command -v air >/dev/null 2>&1 || { echo "air not installed. Install with: go install github.com/cosmtrek/air@latest"; exit 1; }
	air

build: ## Build the application
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server

build-prod: ## Build production binary with optimizations
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server

# Dependency management
deps: ## Download dependencies
	go mod download

tidy: ## Clean up dependencies
	go mod tidy

deps-check: ## Check for outdated dependencies
	go list -u -m all

# Code quality
fmt: ## Format code
	gofmt -s -w .

vet: ## Run go vet
	go vet $(PKG)

lint: ## Run golangci-lint (requires golangci-lint)
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed. Install with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin v1.54.2"; exit 1; }
	golangci-lint run

check: fmt vet ## Run all code quality checks

# Testing
test: ## Run tests
	go test $(PKG) -v

test-coverage: ## Run tests with coverage
	@mkdir -p $(COVERAGE_DIR)
	go test $(PKG) -v -coverprofile=$(COVERAGE_DIR)/coverage.out
	go tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "Coverage report generated at $(COVERAGE_DIR)/coverage.html"

test-race: ## Run tests with race detection
	go test $(PKG) -v -race

benchmark: ## Run benchmarks
	go test $(PKG) -bench=. -benchmem

# Database
migrate: ## Run database migrations
	go run ./cmd/migrate

migrate-create: ## Create a new migration (usage: make migrate-create NAME=migration_name)
	@if [ -z "$(NAME)" ]; then echo "Migration name is required. Usage: make migrate-create NAME=migration_name"; exit 1; fi
	@timestamp=$$(date +%Y%m%d%H%M%S); \
	filename="migrations/$${timestamp}_$(NAME).up.sql"; \
	touch $$filename; \
	echo "Created migration file: $$filename"

# Docker compose targets
compose-up: ## Start infrastructure with Docker Compose
	docker compose up -d

compose-down: ## Stop and remove infrastructure
	docker compose down -v

compose-logs: ## Show Docker Compose logs
	docker compose logs -f --tail=200

compose-rebuild: ## Rebuild and restart services
	docker compose down
	docker compose up --build -d

# Cleanup
clean: ## Clean build artifacts
	rm -rf $(BUILD_DIR)
	rm -rf $(COVERAGE_DIR)

clean-deps: ## Clean dependencies
	go clean -modcache

# Development environment
dev-setup: ## Set up development environment
	@echo "Setting up development environment..."
	@command -v go >/dev/null 2>&1 || { echo "Go is not installed. Please install Go first."; exit 1; }
	@command -v docker >/dev/null 2>&1 || { echo "Docker is not installed. Please install Docker first."; exit 1; }
	@command -v docker compose >/dev/null 2>&1 || { echo "Docker Compose is not installed. Please install Docker Compose first."; exit 1; }
	go install github.com/cosmtrek/air@latest
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.54.2
	@echo "Development environment setup complete!"

# Production
docker-build: ## Build Docker image
	docker build -t $(APP_NAME):latest .

docker-run: ## Run Docker container
	docker run -p 8001:8001 --env-file .env $(APP_NAME):latest

# Security
security-scan: ## Run security scan with gosec (requires gosec)
	@command -v gosec >/dev/null 2>&1 || { echo "gosec not installed. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; exit 1; }
	gosec $(PKG)
