# synmodel

A simple Terminal User Interface (TUI) tool for selecting Synthetic models and generating configuration.

## Overview

synmodel is a command-line application that:
1. Fetches available Synthetic models from an API
2. Presents them in an interactive terminal interface
3. Allows users to select models for configuration generation

## Features

- Interactive TUI built with [Bubbletea](https://github.com/charmbracelet/bubbletea)
- API integration to fetch model listings
- Simple, clean interface for model selection
- Configuration generation capabilities

## How to Run

### Prerequisites

- Go 1.26 or higher installed

### Running the Application

The simplest way to run the application is using Go's module support:

```bash
go run github.com/marstid/synmodel@latest
```

For local development, you can also run:

```bash
go run ./cmd/synmodel-selector
```

This will:
1. Fetch models from the configured API endpoint
2. Launch the TUI for model selection
3. Allow you to navigate and select models using keyboard controls

## Project Structure

```
synmodels/
├── cmd/
│   └── synmodel-selector/
│       └── main.go          # Application entry point
├── internal/
│   ├── api/                 # API client implementation
│   ├── config/              # Configuration handling
│   ├── opencode/            # OpenCode configuration management
│   ├── tui/                 # Terminal UI components
│   └── types/               # Data type definitions
├── go.mod                   # Go module definition
└── go.sum                   # Dependency checksums
```

## Dependencies

- github.com/charmbracelet/bubbletea - TUI framework
- github.com/charmbracelet/bubbles - TUI components
- github.com/charmbracelet/lipgloss - Styling library
- github.com/stretchr/testify - Testing utilities

## Development

To run tests:
```bash
go test ./...
```

To build the binary:
```bash
go build -o synmodel-selector ./cmd/synmodel-selector
```