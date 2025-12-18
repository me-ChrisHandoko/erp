.PHONY: help build run test lint migrate migrate-down migrate-create seed clean docker-build docker-up docker-down

# Default target
.DEFAULT_GOAL := help

# Colors for output
BLUE := \033[36m
GREEN := \033[32m
YELLOW := \033[33m
RED := \033[31m
RESET := \033[0m

help: ## Show this help message
	@echo "$(BLUE)Available targets:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(RESET) %s\n", $$1, $$2}'

# Development
build: ## Build the application
	@echo "$(BLUE)Building API server...$(RESET)"
	@go build -o bin/api ./cmd/api
	@echo "$(BLUE)Building worker...$(RESET)"
	@go build -o bin/worker ./cmd/worker
	@echo "$(GREEN)✓ Build complete$(RESET)"

run: ## Run the API server
	@echo "$(BLUE)Starting API server...$(RESET)"
	@go run ./cmd/api

run-worker: ## Run the background worker
	@echo "$(BLUE)Starting background worker...$(RESET)"
	@go run ./cmd/worker

dev: ## Run with live reload (requires air)
	@echo "$(BLUE)Starting development server with live reload...$(RESET)"
	@air

# Testing
test: ## Run all tests
	@echo "$(BLUE)Running tests...$(RESET)"
	@go test -v -race -coverprofile=coverage.out ./...
	@echo "$(GREEN)✓ Tests complete$(RESET)"

test-coverage: test ## Run tests with coverage report
	@echo "$(BLUE)Generating coverage report...$(RESET)"
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report generated: coverage.html$(RESET)"

test-unit: ## Run unit tests only
	@echo "$(BLUE)Running unit tests...$(RESET)"
	@go test -v -short ./...

test-integration: ## Run integration tests
	@echo "$(BLUE)Running integration tests...$(RESET)"
	@go test -v -run Integration ./test/integration/...

test-e2e: ## Run end-to-end tests
	@echo "$(BLUE)Running E2E tests...$(RESET)"
	@go test -v ./test/e2e/...

# Code Quality
lint: ## Run linter
	@echo "$(BLUE)Running linter...$(RESET)"
	@golangci-lint run
	@echo "$(GREEN)✓ Lint complete$(RESET)"

fmt: ## Format code
	@echo "$(BLUE)Formatting code...$(RESET)"
	@go fmt ./...
	@echo "$(GREEN)✓ Format complete$(RESET)"

vet: ## Run go vet
	@echo "$(BLUE)Running go vet...$(RESET)"
	@go vet ./...
	@echo "$(GREEN)✓ Vet complete$(RESET)"

# Dependencies
deps: ## Install dependencies
	@echo "$(BLUE)Installing dependencies...$(RESET)"
	@go mod download
	@go mod tidy
	@echo "$(GREEN)✓ Dependencies installed$(RESET)"

deps-update: ## Update dependencies
	@echo "$(BLUE)Updating dependencies...$(RESET)"
	@go get -u ./...
	@go mod tidy
	@echo "$(GREEN)✓ Dependencies updated$(RESET)"

# Database
migrate: ## Run database migrations (up)
	@echo "$(BLUE)Running migrations...$(RESET)"
	@./scripts/migrate.sh up
	@echo "$(GREEN)✓ Migrations complete$(RESET)"

migrate-down: ## Rollback last migration
	@echo "$(YELLOW)Rolling back migration...$(RESET)"
	@./scripts/migrate.sh down
	@echo "$(GREEN)✓ Rollback complete$(RESET)"

migrate-reset: ## Reset database (rollback all migrations)
	@echo "$(RED)WARNING: This will reset the database!$(RESET)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		./scripts/migrate.sh reset; \
		echo "$(GREEN)✓ Database reset complete$(RESET)"; \
	fi

migrate-create: ## Create a new migration file (usage: make migrate-create name=create_users)
	@if [ -z "$(name)" ]; then \
		echo "$(RED)Error: Please provide migration name$(RESET)"; \
		echo "Usage: make migrate-create name=create_users"; \
		exit 1; \
	fi
	@echo "$(BLUE)Creating migration: $(name)$(RESET)"
	@./scripts/migrate.sh create $(name)
	@echo "$(GREEN)✓ Migration created$(RESET)"

seed: ## Seed database with test users for login
	@echo "$(BLUE)Seeding database with test users...$(RESET)"
	@go run ./cmd/seed
	@echo "$(GREEN)✓ Database seeded with test users$(RESET)"

seed-sql: ## Seed database with SQL files (dev)
	@echo "$(BLUE)Seeding database from SQL files...$(RESET)"
	@./scripts/seed.sh dev
	@echo "$(GREEN)✓ Database seeded$(RESET)"

seed-test: ## Seed database with test data
	@echo "$(BLUE)Seeding test database...$(RESET)"
	@./scripts/seed.sh test
	@echo "$(GREEN)✓ Test database seeded$(RESET)"

# Docker
docker-build: ## Build Docker image
	@echo "$(BLUE)Building Docker image...$(RESET)"
	@docker build -f scripts/docker/Dockerfile -t erp-backend:latest .
	@echo "$(GREEN)✓ Docker image built$(RESET)"

docker-up: ## Start Docker containers
	@echo "$(BLUE)Starting Docker containers...$(RESET)"
	@docker-compose -f scripts/docker/docker-compose.yml up -d
	@echo "$(GREEN)✓ Docker containers started$(RESET)"

docker-down: ## Stop Docker containers
	@echo "$(BLUE)Stopping Docker containers...$(RESET)"
	@docker-compose -f scripts/docker/docker-compose.yml down
	@echo "$(GREEN)✓ Docker containers stopped$(RESET)"

docker-logs: ## Show Docker logs
	@docker-compose -f scripts/docker/docker-compose.yml logs -f

# Cleanup
clean: ## Clean build artifacts
	@echo "$(BLUE)Cleaning build artifacts...$(RESET)"
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@rm -rf tmp/
	@echo "$(GREEN)✓ Cleanup complete$(RESET)"

clean-all: clean ## Clean all generated files including dependencies
	@echo "$(BLUE)Cleaning all generated files...$(RESET)"
	@go clean -cache -testcache -modcache
	@echo "$(GREEN)✓ Full cleanup complete$(RESET)"

# Setup
install-tools: ## Install development tools
	@echo "$(BLUE)Installing development tools...$(RESET)"
	@go install github.com/cosmtrek/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "$(GREEN)✓ Tools installed$(RESET)"

setup: deps install-tools ## Setup development environment
	@echo "$(BLUE)Setting up development environment...$(RESET)"
	@cp .env.example .env
	@chmod +x scripts/*.sh
	@echo "$(GREEN)✓ Setup complete$(RESET)"
	@echo "$(YELLOW)Please edit .env file with your configuration$(RESET)"

# Documentation
docs: ## Generate API documentation
	@echo "$(BLUE)Generating API documentation...$(RESET)"
	@swag init -g cmd/api/main.go -o api/openapi
	@echo "$(GREEN)✓ Documentation generated$(RESET)"

# All-in-one commands
all: clean deps build test ## Clean, install deps, build, and test
	@echo "$(GREEN)✓ All tasks complete$(RESET)"

check: fmt vet lint test ## Run all checks (format, vet, lint, test)
	@echo "$(GREEN)✓ All checks passed$(RESET)"
