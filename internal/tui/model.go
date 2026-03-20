// Package tui provides the Bubble Tea-based terminal user interface.
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/marstid/synmodels/internal/config"
	"github.com/marstid/synmodels/internal/opencode"
	"github.com/marstid/synmodels/internal/types"
)

// KeyMap defines the key bindings for the TUI.
type KeyMap struct {
	Up          key.Binding
	Down        key.Binding
	Toggle      key.Binding
	SelectAll   key.Binding
	DeselectAll key.Binding
	Generate    key.Binding
	Apply       key.Binding
	Quit        key.Binding
	Help        key.Binding
}

// DefaultKeyMap returns the default key bindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Toggle: key.NewBinding(
			key.WithKeys(" ", "x"),
			key.WithHelp("space/x", "toggle"),
		),
		SelectAll: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "select all"),
		),
		DeselectAll: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "deselect all"),
		),
		Generate: key.NewBinding(
			key.WithKeys("g", "enter"),
			key.WithHelp("g/enter", "generate config"),
		),
		Apply: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "apply to opencode"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q/ctrl+c", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
	}
}

// ShortHelp returns a concise help view.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns a complete help view.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Toggle},
		{k.SelectAll, k.DeselectAll, k.Generate},
		{k.Apply, k.Help, k.Quit},
	}
}

// Styles holds the lipgloss styles for the TUI.
type Styles struct {
	Title        lipgloss.Style
	ModelItem    lipgloss.Style
	SelectedItem lipgloss.Style
	CheckboxOn   lipgloss.Style
	CheckboxOff  lipgloss.Style
	Cursor       lipgloss.Style
	Help         lipgloss.Style
	Error        lipgloss.Style
	Success      lipgloss.Style
}

// DefaultStyles returns the default styles.
func DefaultStyles() Styles {
	return Styles{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED")).
			MarginBottom(1),
		ModelItem: lipgloss.NewStyle().
			PaddingLeft(2),
		SelectedItem: lipgloss.NewStyle().
			PaddingLeft(2).
			Foreground(lipgloss.Color("#10B981")),
		CheckboxOn: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981")),
		CheckboxOff: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")),
		Cursor: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C3AED")).
			Bold(true),
		Help: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")),
		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")),
		Success: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981")),
	}
}

// applyResultMsg is sent when the apply operation completes.
type applyResultMsg struct {
	err error
}

// Model represents the Bubble Tea model for the TUI.
type Model struct {
	models          []types.SelectedModel
	cursor          int
	keys            KeyMap
	help            help.Model
	styles          Styles
	spinner         spinner.Model
	loading         bool
	err             error
	generated       string
	showHelp        bool
	width           int
	height          int
	quitting        bool
	generator       *config.Generator
	waitingForApply bool
	applyError      error
	applySuccess    bool
	configManager   *opencode.Manager
}

// NewModel creates a new TUI model with the given models.
func NewModel(models []types.Model) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED"))

	// Wrap models with selection state
	selectedModels := make([]types.SelectedModel, len(models))
	for i, model := range models {
		selectedModels[i] = types.SelectedModel{
			Model:    model,
			Selected: false,
		}
	}

	return Model{
		models:        selectedModels,
		cursor:        0,
		keys:          DefaultKeyMap(),
		help:          help.New(),
		styles:        DefaultStyles(),
		spinner:       s,
		loading:       false,
		generator:     config.NewGenerator(config.FormatJSON),
		configManager: opencode.NewManager(),
	}
}

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, m.keys.Help):
			m.showHelp = !m.showHelp
			return m, nil

		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil

		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(m.models)-1 {
				m.cursor++
			}
			return m, nil

		case key.Matches(msg, m.keys.Toggle):
			if len(m.models) > 0 {
				m.models[m.cursor].Selected = !m.models[m.cursor].Selected
			}
			return m, nil

		case key.Matches(msg, m.keys.SelectAll):
			for i := range m.models {
				m.models[i].Selected = true
			}
			return m, nil

		case key.Matches(msg, m.keys.DeselectAll):
			for i := range m.models {
				m.models[i].Selected = false
			}
			return m, nil

		case key.Matches(msg, m.keys.Generate):
			output, err := m.generator.GenerateFromSelected(m.models)
			if err != nil {
				m.err = err
			} else {
				m.generated = output
				m.waitingForApply = true
				m.applySuccess = false
				m.applyError = nil
			}
			return m, nil

		case key.Matches(msg, m.keys.Apply):
			if m.waitingForApply && m.generated != "" {
				return m, m.applyConfig()
			}
			return m, nil
		}

	case applyResultMsg:
		m.waitingForApply = false
		if msg.err != nil {
			m.applyError = msg.err
			m.applySuccess = false
		} else {
			m.applySuccess = true
			m.applyError = nil
		}
		return m, nil

	case spinner.TickMsg:
		if m.loading {
			newSpinner, cmd := m.spinner.Update(msg)
			m.spinner = newSpinner
			return m, cmd
		}
	}

	return m, nil
}

// View renders the TUI.
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	var sb strings.Builder

	// Title
	title := m.styles.Title.Render("🤖 Synthetic Model Selector")
	sb.WriteString(title)
	sb.WriteString("\n\n")

	// Error display
	if m.err != nil {
		sb.WriteString(m.styles.Error.Render(fmt.Sprintf("Error: %v\n\n", m.err)))
		m.err = nil // Clear error after displaying
	}

	// Generated config display
	if m.generated != "" {
		sb.WriteString(m.styles.Success.Render("✓ Generated Config:\n"))
		sb.WriteString("\n")
		sb.WriteString(m.generated)
		sb.WriteString("\n\n")

		// Show apply prompt or result
		if m.applySuccess {
			sb.WriteString(m.styles.Success.Render("✓ Configuration applied to opencode successfully!\n"))
			sb.WriteString(m.styles.Help.Render("\nPress 'q' to quit or 'g' to generate again\n"))
		} else if m.applyError != nil {
			sb.WriteString(m.styles.Error.Render(fmt.Sprintf("✗ Failed to apply configuration: %v\n", m.applyError)))
			sb.WriteString(m.styles.Help.Render("\nPress 'y' to retry, 'q' to quit, or 'g' to generate again\n"))
		} else if m.waitingForApply {
			sb.WriteString("Apply this configuration to your opencode config? [y/N]\n")
			sb.WriteString(m.styles.Help.Render("\nPress 'y' to apply, 'q' to quit, or 'g' to generate again\n"))
		} else {
			sb.WriteString(m.styles.Help.Render("\nPress 'q' to quit or 'g' to generate again\n"))
		}
		return sb.String()
	}

	// Loading state
	if m.loading {
		sb.WriteString(fmt.Sprintf("%s Loading models...\n", m.spinner.View()))
		return sb.String()
	}

	// No models
	if len(m.models) == 0 {
		sb.WriteString(m.styles.Error.Render("No models available\n"))
		return sb.String()
	}

	// Model list
	sb.WriteString("Available Models:\n")
	sb.WriteString(strings.Repeat("─", min(m.width, 60)))
	sb.WriteString("\n\n")

	for i, model := range m.models {
		line := m.renderModelItem(i, model)
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("─", min(m.width, 60)))
	sb.WriteString("\n\n")

	// Help
	if m.showHelp {
		sb.WriteString(m.help.View(m.keys))
	} else {
		sb.WriteString(m.styles.Help.Render("Press '?' for help, 'q' to quit"))
	}

	return sb.String()
}

// renderModelItem renders a single model item with checkbox and cursor.
func (m Model) renderModelItem(index int, model types.SelectedModel) string {
	var sb strings.Builder

	// Cursor
	if index == m.cursor {
		sb.WriteString(m.styles.Cursor.Render("> "))
	} else {
		sb.WriteString("  ")
	}

	// Checkbox
	if model.Selected {
		sb.WriteString(m.styles.CheckboxOn.Render("[✓] "))
	} else {
		sb.WriteString(m.styles.CheckboxOff.Render("[ ] "))
	}

	// Model name
	if index == m.cursor {
		sb.WriteString(m.styles.SelectedItem.Render(model.Model.ID))
	} else {
		sb.WriteString(m.styles.ModelItem.Render(model.Model.ID))
	}

	return sb.String()
}

// GetSelectedModels returns the currently selected models.
func (m Model) GetSelectedModels() []types.Model {
	var selected []types.Model
	for _, model := range m.models {
		if model.Selected {
			selected = append(selected, model.Model)
		}
	}
	return selected
}

// applyConfig applies the generated configuration to the opencode config file.
func (m Model) applyConfig() tea.Cmd {
	return func() tea.Msg {
		// Get selected models
		var selected []types.Model
		for _, sm := range m.models {
			if sm.Selected {
				selected = append(selected, sm.Model)
			}
		}

		if len(selected) == 0 {
			return applyResultMsg{err: fmt.Errorf("no models selected")}
		}

		// Build models map with detailed config for each model
		modelsMap := make(map[string]config.ModelConfig)
		for _, model := range selected {
			modelID := model.ID
			if !strings.HasPrefix(modelID, "hf:") {
				modelID = "hf:" + modelID
			}
			modelsMap[modelID] = config.GetModelConfig(model)
		}

		// Apply to opencode config
		if err := m.configManager.AddModels(modelsMap); err != nil {
			return applyResultMsg{err: err}
		}

		return applyResultMsg{err: nil}
	}
}

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
