.PHONY: all build test clean install help

# Default target
all: check build

##@ Build

BINARY_NAME := devslot
VERSION := $(shell grep 'var version' main.go | cut -d'"' -f2)

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME) .

install: ## Install the binary to $GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	@go install .

##@ Test

test: test-unit test-e2e ## Run all tests (unit + E2E)

test-unit: ## Run unit tests
	@echo "Running unit tests..."
	@go test -v ./...

test-e2e: build ## Run E2E tests using zx
	@echo "Running E2E tests..."
	@echo "Note: Requires Node.js and zx. Install with mise or 'npm install -g zx'"
	@mise exec -- zx test/e2e/init.test.mjs

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

##@ Quality

check: fmt lint ## Run all quality checks

fmt: ## Format code using go fmt
	@echo "Formatting code..."
	@go fmt ./...

lint: ## Run golangci-lint or go vet
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Running go vet instead..."; \
		go vet ./...; \
	fi

##@ Utility

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -f coverage.out coverage.html
	@go clean -testcache

run: build ## Build and run the application
	./$(BINARY_NAME)

help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)