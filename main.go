// synmodel-selector is a TUI tool for selecting Synthetic models and generating configuration.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/marstid/synmodels/internal/api"
	"github.com/marstid/synmodels/internal/config"
	"github.com/marstid/synmodels/internal/opencode"
	"github.com/marstid/synmodels/internal/tui"
	"github.com/marstid/synmodels/internal/types"
)

func main() {
	printMode := flag.Bool("print", false, "skip the TUI, select all models, and print the generated config to stdout")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "A TUI tool for selecting Synthetic models and generating opencode configuration.\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s              Launch the interactive TUI\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --print      Print config with all models to stdout\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --print > opencode.json   Write the config to a file\n", os.Args[0])
	}
	flag.Parse()

	if err := run(*printMode); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(printMode bool) error {
	// Create API client
	client := api.NewClient()

	// Fetch models from the API
	fmt.Fprintln(os.Stderr, "Fetching models from API...")
	models, err := client.FetchModels(context.Background())
	if err != nil {
		return fmt.Errorf("failed to fetch models: %w", err)
	}

	if len(models) == 0 {
		return fmt.Errorf("no models found")
	}

	fmt.Fprintf(os.Stderr, "Found %d models\n", len(models))

	if printMode {
		return printConfig(models)
	}

	// Create and run the TUI
	m := tui.NewModel(models)
	p := tea.NewProgram(m, tea.WithAltScreen())

	_, err = p.Run()
	if err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}

	return nil
}

// printConfig generates a config with all models selected and prints it to stdout.
func printConfig(models []types.Model) error {
	// Resolve base URL (same logic as tui.NewModel)
	baseURL := opencode.DefaultBaseURL
	if envURL, ok := os.LookupEnv(opencode.EnvAPIBaseURL); ok && envURL != "" {
		baseURL = envURL
	}
	apiKey := os.Getenv("SYNTHETIC_API_KEY")

	gen := config.NewGenerator(config.FormatJSON).WithBaseURL(baseURL).WithAPIKey(apiKey)

	// Mark every model as selected
	selected := make([]types.SelectedModel, len(models))
	for i, m := range models {
		selected[i] = types.SelectedModel{Model: m, Selected: true}
	}

	output, err := gen.GenerateFromSelected(selected)
	if err != nil {
		return fmt.Errorf("failed to generate config: %w", err)
	}

	fmt.Println(output)
	return nil
}
