// synmodel-selector is a TUI tool for selecting Synthetic models and generating configuration.
package main

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/marstid/synmodel/internal/api"
	"github.com/marstid/synmodel/internal/tui"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Create API client
	client := api.NewClient()

	// Fetch models from the API
	fmt.Println("Fetching models from API...")
	models, err := client.FetchModels(context.Background())
	if err != nil {
		return fmt.Errorf("failed to fetch models: %w", err)
	}

	if len(models) == 0 {
		return fmt.Errorf("no models found")
	}

	fmt.Printf("Found %d models\n", len(models))

	// Create and run the TUI
	m := tui.NewModel(models)
	p := tea.NewProgram(m, tea.WithAltScreen())

	_, err = p.Run()
	if err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}

	return nil
}
