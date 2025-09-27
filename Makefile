# Variables
BINARY_NAME=finance-mcp
MAIN_PACKAGE=./cmd/server/main.go
BUILD_DIR=bin
VERSION ?=v1.0.0
LDFLAGS=-ldflags "-X main.version=${VERSION}"

# Go related variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint

# Colors for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
BLUE=\033[0;34m
NC=\033[0m # No Color

.PHONY: all build clean test coverage deps fmt lint vet run dev help install docker

# Default target
all: clean deps fmt lint test build

# Help target - shows available commands
help: ## Show this help message
	@echo "$(BLUE)Simple MCP Market Data Server$(NC)"
	@echo ""
	@echo "$(YELLOW)Available commands:$(NC)"
	@awk 'BEGIN {FS = ":.*##"}; /^[a-zA-Z_-]+:.*?##/ { printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

# Build the application
build: ## Build the application binary
	@echo "$(YELLOW)Building $(BINARY_NAME)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "$(GREEN)Build completed: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

# Build for multiple platforms
build-all: ## Build for multiple platforms (Linux, macOS, Windows)
	@echo "$(YELLOW)Building for multiple platforms...$(NC)"
	@mkdir -p $(BUILD_DIR)

	# Linux AMD64
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PACKAGE)

	# macOS AMD64
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PACKAGE)

	# macOS ARM64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PACKAGE)

	# Windows AMD64
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PACKAGE)

	@echo "$(GREEN)Multi-platform build completed$(NC)"

# Run the application
run: ## Run the application directly
	@echo "$(YELLOW)Running $(BINARY_NAME)...$(NC)"
	$(GOCMD) run $(MAIN_PACKAGE)

# Development run with hot reload (requires air: go install github.com/cosmtrek/air@latest)
dev: ## Run with hot reload (requires 'air' tool)
	@echo "$(YELLOW)Starting development server with hot reload...$(NC)"
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "$(RED)Error: 'air' not found. Install it with: go install github.com/cosmtrek/air@latest$(NC)"; \
	fi

# Test the application
test: ## Run all tests
	@echo "$(YELLOW)Running tests...$(NC)"
	$(GOTEST) -v ./...

# Run tests with coverage
coverage: ## Run tests with coverage report
	@echo "$(YELLOW)Running tests with coverage...$(NC)"
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

# Run tests in watch mode (requires gotestsum: go install gotest.tools/gotestsum@latest)
test-watch: ## Run tests in watch mode
	@if command -v gotestsum > /dev/null; then \
		gotestsum --watch ./...; \
	else \
		echo "$(RED)Error: 'gotestsum' not found. Install it with: go install gotest.tools/gotestsum@latest$(NC)"; \
	fi

# Benchmark tests
bench: ## Run benchmark tests
	@echo "$(YELLOW)Running benchmarks...$(NC)"
	$(GOTEST) -bench=. -benchmem ./...

# Install/update dependencies
deps: ## Download and install dependencies
	@echo "$(YELLOW)Installing dependencies...$(NC)"
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "$(GREEN)Dependencies installed$(NC)"

# Update dependencies
deps-update: ## Update all dependencies
	@echo "$(YELLOW)Updating dependencies...$(NC)"
	$(GOGET) -u ./...
	$(GOMOD) tidy
	@echo "$(GREEN)Dependencies updated$(NC)"

# Format code
fmt: ## Format Go code
	@echo "$(YELLOW)Formatting code...$(NC)"
	$(GOFMT) -s -w .
	@echo "$(GREEN)Code formatted$(NC)"

# Lint code
lint: ## Run linter (requires golangci-lint)
	@echo "$(YELLOW)Running linter...$(NC)"
	@if command -v $(GOLINT) > /dev/null; then \
		$(GOLINT) run ./...; \
		echo "$(GREEN)Linting completed$(NC)"; \
	else \
		echo "$(RED)Error: golangci-lint not found. Install it from: https://golangci-lint.run/usage/install/$(NC)"; \
	fi

# Vet code
vet: ## Run go vet
	@echo "$(YELLOW)Running go vet...$(NC)"
	$(GOCMD) vet ./...
	@echo "$(GREEN)Go vet completed$(NC)"

# Security scan (requires gosec: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest)
security: ## Run security scan
	@echo "$(YELLOW)Running security scan...$(NC)"
	@if command -v gosec > /dev/null; then \
		gosec ./...; \
		echo "$(GREEN)Security scan completed$(NC)"; \
	else \
		echo "$(RED)Error: 'gosec' not found. Install it with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest$(NC)"; \
	fi

# Clean build artifacts
clean: ## Clean build artifacts and temporary files
	@echo "$(YELLOW)Cleaning...$(NC)"
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	rm -f *.log
	@echo "$(GREEN)Clean completed$(NC)"

# Install the binary to $GOPATH/bin
install: build ## Install the binary to $GOPATH/bin
	@echo "$(YELLOW)Installing $(BINARY_NAME)...$(NC)"
	cp $(BUILD_DIR)/$(BINARY_NAME) $(shell go env GOPATH)/bin/
	@echo "$(GREEN)$(BINARY_NAME) installed to $(shell go env GOPATH)/bin/$(NC)"

# Create a new release build
release: clean ## Create a release build
	@echo "$(YELLOW)Creating release build...$(NC)"
	$(MAKE) build-all
	@echo "$(GREEN)Release build completed$(NC)"

# Setup development environment
setup: ## Setup development environment
	@echo "$(YELLOW)Setting up development environment...$(NC)"

	# Install development tools
	go install github.com/cosmtrek/air@latest
	go install gotest.tools/gotestsum@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

	# Create .env from example if it doesn't exist
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "$(YELLOW)Created .env file from .env.example. Please edit it with your actual API key.$(NC)"; \
	fi

	@echo "$(GREEN)Development environment setup completed$(NC)"

# Validate the project structure and code quality
validate: fmt lint vet security test ## Run all validation checks

# Docker targets
docker-build: ## Build Docker image
	@echo "$(YELLOW)Building Docker image...$(NC)"
	docker build -t $(BINARY_NAME):$(VERSION) .
	docker build -t $(BINARY_NAME):latest .
	@echo "$(GREEN)Docker image built$(NC)"

docker-run: ## Run Docker container
	@echo "$(YELLOW)Running Docker container...$(NC)"
	docker run --rm -it --env-file .env $(BINARY_NAME):latest

# Documentation targets
docs: ## Generate documentation
	@echo "$(YELLOW)Generating documentation...$(NC)"
	$(GOCMD) doc -all ./... > docs/API.md
	@echo "$(GREEN)Documentation generated$(NC)"

# Show project statistics
stats: ## Show project statistics
	@echo "$(BLUE)Project Statistics:$(NC)"
	@echo "Go files: $$(find . -name '*.go' -not -path './vendor/*' | wc -l)"
	@echo "Lines of code: $$(find . -name '*.go' -not -path './vendor/*' -exec cat {} \; | wc -l)"
	@echo "Test files: $$(find . -name '*_test.go' -not -path './vendor/*' | wc -l)"
	@echo "Dependencies: $$(go list -m all | wc -l)"

# Pre-commit hook setup
pre-commit: ## Setup pre-commit hooks
	@echo "$(YELLOW)Setting up pre-commit hooks...$(NC)"
	@echo '#!/bin/bash' > .git/hooks/pre-commit
	@echo 'make validate' >> .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "$(GREEN)Pre-commit hooks installed$(NC)"

# Development shortcuts
quick-test: fmt vet ## Quick test (format, vet, and test)
	$(MAKE) test

check: fmt lint vet test ## Full check (format, lint, vet, and test)

# Version information
version: ## Show version information
	@echo "$(BLUE)Simple MCP Market Data Server$(NC)"
	@echo "Version: $(VERSION)"
	@echo "Go version: $$(go version)"
	@echo "Build date: $$(date)"
