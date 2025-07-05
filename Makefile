.PHONY: all build test test-unit test-e2e clean install lint help

# Variables
BINARY_NAME := devslot
GO_FILES := $(shell find . -name '*.go' -not -path './test/*')
VERSION := $(shell grep 'const version' main.go | cut -d'"' -f2)

# Default target
all: test build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME) .

# Install the binary to $GOPATH/bin
install: build
	@echo "Installing $(BINARY_NAME)..."
	@go install .

# Run all tests
test: test-unit test-e2e

# Run unit tests
test-unit:
	@echo "Running unit tests..."
	@go test -v ./...

# Run unit tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run E2E tests
test-e2e: build
	@echo "Running E2E tests..."
	@mise exec -- zx test/e2e/init.test.mjs

# Run linter
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: brew install golangci-lint"; \
		go vet ./...; \
	fi

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -f coverage.out coverage.html
	@go clean -testcache

# Run the application
run: build
	@./$(BINARY_NAME)

# Display help
help:
	@echo "Available targets:"
	@echo "  make build       - Build the binary"
	@echo "  make install     - Install the binary to \$$GOPATH/bin"
	@echo "  make test        - Run all tests (unit + E2E)"
	@echo "  make test-unit   - Run unit tests only"
	@echo "  make test-e2e    - Run E2E tests only"
	@echo "  make test-coverage - Run tests with coverage report"
	@echo "  make lint        - Run linter"
	@echo "  make fmt         - Format code"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make run         - Build and run the application"
	@echo "  make help        - Show this help message"