# synmodel

A simple Terminal User Interface (TUI) tool for selecting Synthetic models and generating configuration.

## Overview

synmodel is a command-line application that:
1. Fetches available Synthetic models from an API (including pricing information)
2. Presents them in an interactive terminal interface
3. Allows users to select models for configuration generation with detailed pricing data

## Features

- Interactive TUI built with [Bubbletea](https://github.com/charmbracelet/bubbletea)
- API integration to fetch model listings including pricing information
- Simple, clean interface for model selection
- Configuration generation capabilities with detailed pricing data

## How to Run

### Prerequisites

- Go 1.26 or higher installed

### Running the Application

The simplest way to run the application is using Go's module support:

```bash
go run github.com/marstid/synmodels@latest
```

For local development, you can also run:

```bash
make run
```

Or build the binary first and then run:

```bash
make build
./bin/synmodel-selector
```

This will:
1. Fetch models from the configured API endpoint
2. Launch the TUI for model selection
3. Allow you to navigate and select models using keyboard controls

## Project Structure

```
synmodels/
├── main.go                  # Application entry point
├── internal/
│   ├── api/                 # API client implementation
│   ├── config/              # Configuration handling
│   ├── opencode/            # OpenCode configuration management
│   ├── tui/                 # Terminal UI components
│   └── types/               # Data type definitions
├── go.mod                   # Go module definition
├── go.sum                   # Dependency checksums
└── Makefile                 # Build automation
```

## Dependencies

- github.com/charmbracelet/bubbletea - TUI framework
- github.com/charmbracelet/bubbles - TUI components
- github.com/charmbracelet/lipgloss - Styling library
- github.com/stretchr/testify - Testing utilities

## Development

### Using Make

The project includes a Makefile for common development tasks:

```bash
# Show all available targets
make help

# Clean, format, vet, test, and build (complete pipeline)
make all

# Build the binary
make build

# Run tests
make test

# Run tests with coverage report
make test-coverage

# Format code
make fmt

# Run go vet
make vet

# Run linter (requires golangci-lint)
make lint

# Clean build artifacts
make clean

# Tidy Go modules
make tidy

# Run CI pipeline (tidy, fmt, vet, test, build)
make ci
```

### Manual Commands

To run tests:
```bash
go test ./...
```

To build the binary:
```bash
go build -o synmodel-selector .
```