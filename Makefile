.PHONY: all build build.binary build.install
.PHONY: test test.all test.unit test.e2e test.e2e.zx test.coverage test.race
.PHONY: check check.all check.format check.lint check.mod
.PHONY: dev dev.run dev.clean dev.setup
.PHONY: ci ci.setup ci.test ci.check
.PHONY: help

# Default target
all: check.all build.binary

##@ Build

BINARY_NAME := devslot
MAIN_PATH := ./cmd/devslot
BUILD_DIR := ./build
VERSION := $(shell grep 'var version' $(MAIN_PATH)/main.go | cut -d'"' -f2 || echo "dev")

build: build.binary ## Alias for build.binary

build.binary: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

build.install: ## Install the binary to $GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	@go install $(MAIN_PATH)

##@ Test

test: test.all ## Alias for test.all

test.all: test.unit test.e2e ## Run all tests (unit + E2E)

test.unit: ## Run unit tests
	@echo "Running unit tests..."
	@go test -v ./...

test.e2e: test.e2e.zx ## Run E2E tests using zx

test.e2e.zx: build.binary ## Run E2E tests using zx
	@echo "Running E2E tests (zx)..."
	@echo "Note: Requires zx. Install with 'npm install -g zx' or use mise"
	@if [ -f test/e2e/init.test.mjs ]; then \
		zx test/e2e/init.test.mjs; \
	else \
		echo "No zx E2E tests found"; \
	fi

test.coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@mkdir -p $(BUILD_DIR)
	@go test -v -coverprofile=$(BUILD_DIR)/coverage.out ./...
	@go tool cover -html=$(BUILD_DIR)/coverage.out -o $(BUILD_DIR)/coverage.html
	@echo "Coverage report generated: $(BUILD_DIR)/coverage.html"

test.race: ## Run tests with race detector
	@echo "Running tests with race detector..."
	@go test -v -race ./...

##@ Quality Check

check: check.all ## Alias for check.all

check.all: check.format check.lint check.mod ## Run all quality checks

check.format: ## Check code formatting
	@echo "Checking code format..."
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "The following files need formatting:"; \
		gofmt -l .; \
		exit 1; \
	else \
		echo "All files are properly formatted"; \
	fi

check.lint: ## Run linter
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Running go vet instead..."; \
		go vet ./...; \
	fi

check.mod: ## Check go.mod tidiness
	@echo "Checking go.mod..."
	@go mod tidy
	@if [ -n "$$(git status --porcelain go.mod go.sum)" ]; then \
		echo "go.mod or go.sum is not tidy. Run 'go mod tidy' to fix."; \
		exit 1; \
	else \
		echo "go.mod is tidy"; \
	fi

##@ Development

dev: dev.run ## Alias for dev.run

dev.run: build.binary ## Build and run the application
	$(BUILD_DIR)/$(BINARY_NAME)

dev.clean: ## Clean build artifacts and test cache
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f $(BINARY_NAME)
	@rm -f coverage.out coverage.html
	@go clean -testcache

dev.setup: ## Setup development environment
	@echo "Setting up development environment..."
	@if command -v mise >/dev/null 2>&1; then \
		mise install; \
		echo "Development tools installed via mise"; \
	else \
		echo "Please install mise first: https://mise.jdx.dev/getting-started.html"; \
	fi

##@ CI/CD

ci: ci.check ci.test ## Run CI pipeline locally

ci.setup: ## Setup CI environment
	@echo "Setting up CI environment..."
	@go mod download
	@if [ -f test/e2e/init.test.mjs ]; then \
		npm install -g zx; \
	fi

ci.test: test.unit test.e2e ## Run tests for CI

ci.check: check.format check.lint check.mod ## Run checks for CI

##@ Other

help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z][a-zA-Z0-9_.-]*:.*##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)