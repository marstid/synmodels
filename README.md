# synmodel

A Terminal User Interface (TUI) tool for selecting Synthetic models and generating [opencode](https://opencode.ai) configuration.

## Overview

synmodel is a command-line application that:
1. Fetches available Synthetic models from the Synthetic API (`https://api.synthetic.new/openai/v1/models`)
2. Presents them in an interactive terminal interface with checkboxes and keyboard navigation
3. Generates configuration (JSON/YAML) with per-model details: limits, modalities, cost, reasoning support, and variant presets
4. Applies the configuration directly to `~/.config/opencode/opencode.json`, creating a `synthetic` provider with the selected models, API key, and pricing data

## Features

- Interactive TUI built with [Bubbletea](https://github.com/charmbracelet/bubbletea)
- Fetches model listings from the Synthetic API, including pricing and capability data
- Generates opencode-compatible config with:
  - `cost` (numeric input/output/cache pricing)
  - `reasoning` flag for reasoning-capable models (detected from `supported_features`)
  - `variants` presets for reasoning models (e.g. GLM-5.2: `none`/`high`/`max` → `reasoningEffort`)
  - `tool_call`, `limit`, `modalities` from real API data
- Writes directly to `opencode.json` with atomic writes, timestamped backups, and full field preservation (your existing config, MCP servers, other providers, etc. are kept)
- Secure: config and backups written with `0600` permissions; config directory `0700`

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
1. Fetch models from the Synthetic API
2. Launch the TUI for model selection
3. Let you navigate, toggle, and select models
4. Press `g` to generate a config preview, then `y` to apply it to opencode

### Environment Variables

| Variable | Description | Default |
|---|---|---|
| `SYN_API` | Base URL for the Synthetic API (used for both fetching models and writing the config) | `https://api.synthetic.new/openai/v1` |
| `OPENCODE_CONFIG` | Path to the opencode config file | `~/.config/opencode/opencode.json` |

## Variant Support

Some reasoning models (e.g. GLM-5.2) support adjustable reasoning effort levels. synmodel emits `variants` presets in the opencode config so you can switch between them in opencode's model picker.

Variant definitions live in a built-in registry (`internal/variants/registry.go`), since the Synthetic API does not expose variant presets. To add support for a new reasoning model, add a single entry to the registry:

```go
var registry = map[string]Spec{
    "hf:zai-org/GLM-5.2": {
        Reasoning: true,
        Variants: map[string]map[string]any{
            "none": {"reasoningEffort": "none"},
            "high": {"reasoningEffort": "high"},
            "max":  {"reasoningEffort": "xhigh"},
        },
    },
}
```

## Keyboard Controls

| Key | Action |
|---|---|
| `↑`/`k` | Move cursor up |
| `↓`/`j` | Move cursor down |
| `space`/`x` | Toggle model selection |
| `a` | Select all |
| `d` | Deselect all |
| `g`/`enter` | Generate config |
| `y` | Apply to opencode |
| `esc`/`b` | Back to model list (from config view) |
| `q`/`ctrl+c` | Quit |
| `?` | Toggle help |

## Project Structure

```
synmodels/
├── main.go                  # Application entry point
├── internal/
│   ├── api/                 # HTTP client for the Synthetic API
│   ├── config/              # Config generation (JSON/YAML)
│   ├── opencode/            # opencode.json read/write/backup
│   ├── tui/                 # Bubble Tea terminal UI
│   ├── types/               # Shared data structures
│   └── variants/            # Built-in variant registry
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
