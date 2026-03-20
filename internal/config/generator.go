// Package config provides functionality for generating configuration output.
package config

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/marstid/synmodels/internal/types"
)

// OutputFormat represents the format of the generated config.
type OutputFormat string

const (
	// FormatJSON outputs the config as formatted JSON.
	FormatJSON OutputFormat = "json"
	// FormatYAML outputs the config as YAML-like format.
	FormatYAML OutputFormat = "yaml"
)

// Generator creates configuration output from selected models.
type Generator struct {
	format OutputFormat
}

// NewGenerator creates a new config generator with the specified format.
func NewGenerator(format OutputFormat) *Generator {
	return &Generator{
		format: format,
	}
}

// Generate creates a configuration string from the selected models.
func (g *Generator) Generate(models []types.Model) (string, error) {
	if len(models) == 0 {
		return "", fmt.Errorf("no models selected")
	}

	// Extract model IDs
	modelIDs := make([]string, len(models))
	for i, model := range models {
		modelIDs[i] = model.ID
	}

	// Create a config structure similar to the sample
	config := map[string]interface{}{
		"models": modelIDs,
	}

	switch g.format {
	case FormatJSON:
		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal config: %w", err)
		}
		return string(data), nil
	case FormatYAML:
		// Simple YAML-like output
		var sb strings.Builder
		sb.WriteString("models:\n")
		for _, id := range modelIDs {
			sb.WriteString(fmt.Sprintf("  - %s\n", id))
		}
		return sb.String(), nil
	default:
		return "", fmt.Errorf("unsupported format: %s", g.format)
	}
}

// ModelConfig represents the configuration for a single model.
type ModelConfig struct {
	Name       string          `json:"name"`
	ToolCall   bool            `json:"tool_call"`
	Limit      ModelLimits     `json:"limit"`
	Modalities ModelModalities `json:"modalities"`
	Pricing    PricingInfo     `json:"pricing,omitempty"`
}

// PricingInfo represents the pricing information for a model configuration.
type PricingInfo struct {
	Prompt           string `json:"prompt,omitempty"`
	Completion       string `json:"completion,omitempty"`
	Image            string `json:"image,omitempty"`
	Request          string `json:"request,omitempty"`
	InputCacheReads  string `json:"input_cache_reads,omitempty"`
	InputCacheWrites string `json:"input_cache_writes,omitempty"`
}

// ModelLimits represents the context and output token limits.
type ModelLimits struct {
	Context int `json:"context"`
	Output  int `json:"output"`
}

// ModelModalities represents the input and output modalities.
type ModelModalities struct {
	Input  []string `json:"input"`
	Output []string `json:"output"`
}

// GetModelConfig generates a ModelConfig from actual API data.
func GetModelConfig(model types.Model) ModelConfig {
	// Use actual API data for context limit
	contextLimit := model.ContextLength

	// Use actual API data for output limit, with default fallback
	outputLimit := model.MaxOutputLength
	if outputLimit == 0 {
		outputLimit = 4096 // Default output limit if not specified
	}

	// Use actual API data for input modalities
	inputModalities := model.InputModalities
	if len(inputModalities) == 0 {
		inputModalities = []string{"text"} // Default to text only if not specified
	}

	// Use actual API data for output modalities
	outputModalities := model.OutputModalities
	if len(outputModalities) == 0 {
		outputModalities = []string{"text"} // Default to text only if not specified
	}

	// Check if tools feature is supported
	toolCall := false
	for _, feature := range model.SupportedFeatures {
		if feature == "tools" {
			toolCall = true
			break
		}
	}

	// Extract the last part of the model ID for the name
	// e.g., "hf:zai-org/GLM-4.7-Flash" -> "GLM 4.7 Flash"
	// e.g., "nvidia/Kimi-K2.5-NVFP4" -> "Kimi K2.5 NVFP4"
	parts := strings.Split(model.ID, "/")
	name := parts[len(parts)-1]
	name = strings.ReplaceAll(name, "-", " ")

	return ModelConfig{
		Name:     name,
		ToolCall: toolCall,
		Limit: ModelLimits{
			Context: contextLimit,
			Output:  outputLimit,
		},
		Modalities: ModelModalities{
			Input:  inputModalities,
			Output: outputModalities,
		},
		Pricing: PricingInfo{
			Prompt:           model.Pricing.Prompt,
			Completion:       model.Pricing.Completion,
			Image:            model.Pricing.Image,
			Request:          model.Pricing.Request,
			InputCacheReads:  model.Pricing.InputCacheReads,
			InputCacheWrites: model.Pricing.InputCacheWrites,
		},
	}
}

// GenerateFromSelected generates config from a list of SelectedModel.
func (g *Generator) GenerateFromSelected(selectedModels []types.SelectedModel) (string, error) {
	var selected []types.Model
	for _, sm := range selectedModels {
		if sm.Selected {
			selected = append(selected, sm.Model)
		}
	}

	if len(selected) == 0 {
		return "", fmt.Errorf("no models selected")
	}

	// Build models map with detailed config for each model
	modelsMap := make(map[string]ModelConfig)
	for _, model := range selected {
		modelsMap[model.ID] = GetModelConfig(model)
	}

	// Create the final config structure
	config := map[string]interface{}{
		"models": modelsMap,
	}

	switch g.format {
	case FormatJSON:
		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal config: %w", err)
		}
		return string(data), nil
	case FormatYAML:
		// Simple YAML-like output
		var sb strings.Builder
		sb.WriteString("models:\n")
		for id, cfg := range modelsMap {
			sb.WriteString(fmt.Sprintf("  %s:\n", id))
			sb.WriteString(fmt.Sprintf("    name: %s\n", cfg.Name))
			sb.WriteString(fmt.Sprintf("    tool_call: %t\n", cfg.ToolCall))
			sb.WriteString("    limit:\n")
			sb.WriteString(fmt.Sprintf("      context: %d\n", cfg.Limit.Context))
			sb.WriteString(fmt.Sprintf("      output: %d\n", cfg.Limit.Output))
			sb.WriteString("    modalities:\n")
			sb.WriteString(fmt.Sprintf("      input: %v\n", cfg.Modalities.Input))
			sb.WriteString(fmt.Sprintf("      output: %v\n", cfg.Modalities.Output))
		}
		return sb.String(), nil
	default:
		return "", fmt.Errorf("unsupported format: %s", g.format)
	}
}

// DefaultConfig returns a default config structure with the given model IDs.
func DefaultConfig(modelIDs []string) map[string]interface{} {
	return map[string]interface{}{
		"models": modelIDs,
		"settings": map[string]interface{}{
			"temperature": 0.7,
			"max_tokens":  2048,
		},
		"endpoints": map[string]interface{}{
			"completions": "/v1/completions",
			"chat":        "/v1/chat/completions",
		},
	}
}

// GenerateFullConfig creates a full configuration with additional settings.
func (g *Generator) GenerateFullConfig(models []types.Model) (string, error) {
	if len(models) == 0 {
		return "", fmt.Errorf("no models selected")
	}

	modelIDs := make([]string, len(models))
	for i, model := range models {
		modelIDs[i] = model.ID
	}

	config := DefaultConfig(modelIDs)

	switch g.format {
	case FormatJSON:
		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal config: %w", err)
		}
		return string(data), nil
	case FormatYAML:
		// Simple YAML-like output
		var sb strings.Builder
		sb.WriteString("models:\n")
		for _, id := range modelIDs {
			sb.WriteString(fmt.Sprintf("  - %s\n", id))
		}
		sb.WriteString("\nsettings:\n")
		sb.WriteString("  temperature: 0.7\n")
		sb.WriteString("  max_tokens: 2048\n")
		sb.WriteString("\nendpoints:\n")
		sb.WriteString("  completions: /v1/completions\n")
		sb.WriteString("  chat: /v1/chat/completions\n")
		return sb.String(), nil
	default:
		return "", fmt.Errorf("unsupported format: %s", g.format)
	}
}
