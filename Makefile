# synmodel Makefile
# A simple TUI tool for selecting Synthetic models and generating configuration

# Variables
BINARY_NAME := synmodel-selector
BINARY_PATH := ./bin/$(BINARY_NAME)
GO := go
GOFLAGS := -v

# Colors for output
BLUE := \033[36m
GREEN := \033[32m
YELLOW := \033[33m
RED := \033[31m
NC := \033[0m # No Color

.PHONY: all build test clean run fmt vet lint install help

# Default target
all: clean fmt vet test build

## build: Build the binary
build:
	@echo "$(BLUE)Building $(BINARY_NAME)...$(NC)"
	@mkdir -p bin
	$(GO) build $(GOFLAGS) -o $(BINARY_PATH) .
	@echo "$(GREEN)✓ Build complete: $(BINARY_PATH)$(NC)"

## test: Run all tests
test:
	@echo "$(BLUE)Running tests...$(NC)"
	$(GO) test $(GOFLAGS) -race -cover ./...
	@echo "$(GREEN)✓ Tests passed$(NC)"

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "$(BLUE)Running tests with coverage...$(NC)"
	$(GO) test -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report generated: coverage.html$(NC)"

## clean: Clean build artifacts
clean:
	@echo "$(BLUE)Cleaning...$(NC)"
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "$(GREEN)✓ Clean complete$(NC)"

## run: Run the application (requires terminal)
run:
	@echo "$(BLUE)Running $(BINARY_NAME)...$(NC)"
	$(GO) run .

## run-dev: Run the application in local dev mode
run-dev:
	@echo "$(BLUE)Running in dev mode...$(NC)"
	$(GO) run $(GOFLAGS) .

## install: Install the binary to GOPATH/bin
install:
	@echo "$(BLUE)Installing $(BINARY_NAME)...$(NC)"
	$(GO) install .
	@echo "$(GREEN)✓ Installed to $(GOPATH)/bin/$(BINARY_NAME)$(NC)"

## fmt: Format Go code
fmt:
	@echo "$(BLUE)Formatting code...$(NC)"
	$(GO) fmt ./...
	@echo "$(GREEN)✓ Format complete$(NC)"

## vet: Run go vet
vet:
	@echo "$(BLUE)Running go vet...$(NC)"
	$(GO) vet ./...
	@echo "$(GREEN)✓ Vet complete$(NC)"

## lint: Run golangci-lint (if installed)
lint:
	@echo "$(BLUE)Running linter...$(NC)"
	@which golangci-lint > /dev/null 2>&1 || (echo "$(YELLOW)golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(NC)" && exit 1)
	golangci-lint run --fix ./...
	@echo "$(GREEN)✓ Lint complete$(NC)"

## tidy: Tidy and verify Go modules
tidy:
	@echo "$(BLUE)Tidying modules...$(NC)"
	$(GO) mod tidy
	$(GO) mod verify
	@echo "$(GREEN)✓ Modules tidied$(NC)"

## deps: Download and verify dependencies
deps:
	@echo "$(BLUE)Downloading dependencies...$(NC)"
	$(GO) mod download
	$(GO) mod verify
	@echo "$(GREEN)✓ Dependencies downloaded$(NC)"

## check: Run all checks (fmt, vet, test)
check: fmt vet test
	@echo "$(GREEN)✓ All checks passed$(NC)"

## ci: Run CI pipeline (build, test, lint)
ci: tidy fmt vet test build
	@echo "$(GREEN)✓ CI pipeline complete$(NC)"

## help: Show this help message
help:
	@echo "$(BLUE)synmodel - TUI tool for selecting Synthetic models$(NC)"
	@echo ""
	@echo "$(YELLOW)Available targets:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "$(YELLOW)Examples:$(NC)"
	@echo "  make build       # Build the binary"
	@echo "  make test        # Run tests"
	@echo "  make run         # Run the application"
	@echo "  make all         # Clean, format, vet, test, and build"
	@echo "  make install     # Install to GOPATH/bin"
