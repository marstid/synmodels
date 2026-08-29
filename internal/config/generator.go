// Package config provides functionality for generating configuration output.
package config

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/marstid/synmodels/internal/types"
	"github.com/marstid/synmodels/internal/variants"
)

const defaultBaseURL = "https://api.synthetic.new/openai/v1"

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
	format  OutputFormat
	baseURL string
	apiKey  string
}

// NewGenerator creates a new config generator with the specified format.
func NewGenerator(format OutputFormat) *Generator {
	return &Generator{
		format:  format,
		baseURL: defaultBaseURL,
	}
}

// WithBaseURL sets the provider base URL to include in the generated config preview.
func (g *Generator) WithBaseURL(baseURL string) *Generator {
	if baseURL != "" {
		g.baseURL = baseURL
	}
	return g
}

// WithAPIKey sets the API key to use in the generated config preview.
// When non-empty, the key value is used directly. When empty, the placeholder
// text "SYNTHETIC_API_KEY" is shown instead.
func (g *Generator) WithAPIKey(apiKey string) *Generator {
	g.apiKey = apiKey
	return g
}

// apiKeyLabel returns the preview label for the API key.
func (g *Generator) apiKeyLabel() string {
	if g.apiKey != "" {
		return g.apiKey
	}
	return "SYNTHETIC_API_KEY"
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
			fmt.Fprintf(&sb, "  - %s\n", id)
		}
		return sb.String(), nil
	default:
		return "", fmt.Errorf("unsupported format: %s", g.format)
	}
}

// ModelConfig represents the configuration for a single model.
type ModelConfig struct {
	Name        string                    `json:"name,omitempty"`
	Reasoning   bool                      `json:"reasoning,omitempty"`
	ToolCall    bool                      `json:"tool_call"`
	Interleaved any                       `json:"interleaved,omitempty"`
	Limit       ModelLimits               `json:"limit"`
	Modalities  ModelModalities           `json:"modalities"`
	Cost        *CostInfo                 `json:"cost,omitempty"`
	Options     map[string]any            `json:"options,omitempty"`
	Variants    map[string]map[string]any `json:"variants,omitempty"`
}

// CostInfo represents the cost information for a model configuration.
// Fields use numeric values per the opencode config schema (not string prices).
type CostInfo struct {
	Input      float64 `json:"input,omitempty"`
	Output     float64 `json:"output,omitempty"`
	CacheRead  float64 `json:"cache_read,omitempty"`
	CacheWrite float64 `json:"cache_write,omitempty"`
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

// OpencodeModelKey derives the opencode model key from a model's API data.
// If the ID already starts with "hf:", it is used as-is.
// Otherwise, if hugging_face_id is non-empty, "hf:" + hugging_face_id is used.
// Otherwise, the raw ID is returned (no blind "hf:" prefix).
func OpencodeModelKey(model types.Model) string {
	if strings.HasPrefix(model.ID, "hf:") {
		return model.ID
	}
	if model.HuggingFaceID != "" {
		return "hf:" + model.HuggingFaceID
	}
	return model.ID
}

// parsePrice converts a pricing string like "$0.0000014" or "0" to a float64.
func parsePrice(s string) float64 {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "$")
	if s == "" || s == "0" {
		return 0
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

// hasFeature checks if a feature is in the supported features list (case-insensitive).
func hasFeature(features []string, feature string) bool {
	for _, f := range features {
		if strings.EqualFold(f, feature) {
			return true
		}
	}
	return false
}

// extractModelName derives a human-friendly name from the model's API data.
// Prefers the API-provided name field, falling back to deriving from the ID.
func extractModelName(model types.Model) string {
	if model.Name != "" {
		parts := strings.Split(model.Name, "/")
		name := parts[len(parts)-1]
		return strings.ReplaceAll(name, "-", " ")
	}
	id := model.ID
	id = strings.TrimPrefix(id, "hf:")
	id = strings.TrimPrefix(id, "syn:")
	if model.HuggingFaceID != "" {
		hfParts := strings.Split(model.HuggingFaceID, "/")
		if len(hfParts) > 0 {
			return strings.ReplaceAll(hfParts[len(hfParts)-1], "-", " ")
		}
	}
	parts := strings.Split(id, "/")
	name := parts[len(parts)-1]
	return strings.ReplaceAll(name, "-", " ")
}

// mapEffort translates an API reasoning effort level to the opencode
// reasoningEffort value. "max" maps to "xhigh"; all other efforts pass
// through unchanged.
func mapEffort(effort string) string {
	if effort == "max" {
		return "xhigh"
	}
	return effort
}

// variantsFromEfforts builds opencode variant presets from the model's
// reasoning_parameters.efforts. Each effort becomes a variant keyed by its
// own name with a reasoningEffort request option. Returns nil when efforts is
// empty or the model is not a reasoning model.
func variantsFromEfforts(model types.Model, reasoning bool) map[string]map[string]any {
	if !reasoning || len(model.ReasoningParameters.Efforts) == 0 {
		return nil
	}
	variants := make(map[string]map[string]any, len(model.ReasoningParameters.Efforts))
	for _, effort := range model.ReasoningParameters.Efforts {
		variants[effort] = map[string]any{"reasoningEffort": mapEffort(effort)}
	}
	return variants
}

// GetModelConfig generates a ModelConfig from actual API data.
func GetModelConfig(model types.Model) ModelConfig {
	contextLimit := model.ContextLength
	if contextLimit == 0 {
		contextLimit = 8192
	}

	outputLimit := model.MaxOutputLength
	if outputLimit == 0 {
		outputLimit = 4096
	}

	inputModalities := model.InputModalities
	if len(inputModalities) == 0 {
		inputModalities = []string{"text"}
	}

	outputModalities := model.OutputModalities
	if len(outputModalities) == 0 {
		outputModalities = []string{"text"}
	}

	toolCall := hasFeature(model.SupportedFeatures, "tools")
	reasoning := hasFeature(model.SupportedFeatures, "reasoning")

	cfg := ModelConfig{
		Name:      extractModelName(model),
		Reasoning: reasoning,
		ToolCall:  toolCall,
		Limit: ModelLimits{
			Context: contextLimit,
			Output:  outputLimit,
		},
		Modalities: ModelModalities{
			Input:  inputModalities,
			Output: outputModalities,
		},
	}

	cost := &CostInfo{
		Input:      parsePrice(model.Pricing.Prompt),
		Output:     parsePrice(model.Pricing.Completion),
		CacheRead:  parsePrice(model.Pricing.InputCacheReads),
		CacheWrite: parsePrice(model.Pricing.InputCacheWrites),
	}
	if cost.Input != 0 || cost.Output != 0 || cost.CacheRead != 0 || cost.CacheWrite != 0 {
		cfg.Cost = cost
	}

	key := OpencodeModelKey(model)

	// Generate variant presets dynamically from reasoning_parameters.efforts.
	cfg.Variants = variantsFromEfforts(model, reasoning)

	// Apply per-model overrides from the registry. The registry provides
	// values the API does not expose (e.g. interleaved for Kimi-K3) and can
	// optionally replace the generated variants as an escape hatch.
	if spec, ok := variants.Lookup(key); ok {
		if spec.Reasoning {
			cfg.Reasoning = true
		}
		if spec.Interleaved != nil {
			cfg.Interleaved = spec.Interleaved
		}
		if len(spec.Variants) > 0 {
			cfg.Variants = spec.Variants
		}
		if len(spec.Options) > 0 {
			cfg.Options = spec.Options
		}
	}

	return cfg
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

	modelsMap := make(map[string]ModelConfig)
	for _, model := range selected {
		modelsMap[OpencodeModelKey(model)] = GetModelConfig(model)
	}

	providerConfig := map[string]interface{}{
		"provider": map[string]interface{}{
			"synthetic": map[string]interface{}{
				"npm":  "@ai-sdk/openai-compatible",
				"name": "Synthetic",
				"options": map[string]interface{}{
					"baseURL": g.baseURL,
					"apiKey":  g.apiKeyLabel(),
				},
				"models": modelsMap,
			},
		},
	}

	switch g.format {
	case FormatJSON:
		data, err := json.MarshalIndent(providerConfig, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal config: %w", err)
		}
		return string(data), nil
	case FormatYAML:
		return g.marshalYAML(modelsMap), nil
	default:
		return "", fmt.Errorf("unsupported format: %s", g.format)
	}
}

// yamlKey quotes a YAML key if it contains characters that require quoting.
func yamlKey(k string) string {
	if strings.ContainsAny(k, ":#{}[]&*!|>'\"%@`") {
		return fmt.Sprintf("%q", k)
	}
	return k
}

// marshalYAML produces a simple YAML-like representation of the full provider config.
func (g *Generator) marshalYAML(modelsMap map[string]ModelConfig) string {
	var sb strings.Builder
	sb.WriteString("provider:\n")
	sb.WriteString("  synthetic:\n")
	sb.WriteString("    npm: @ai-sdk/openai-compatible\n")
	sb.WriteString("    name: Synthetic\n")
	sb.WriteString("    options:\n")
	fmt.Fprintf(&sb, "      baseURL: %s\n", g.baseURL)
	fmt.Fprintf(&sb, "      apiKey: %s\n", g.apiKeyLabel())
	sb.WriteString("    models:\n")

	keys := make([]string, 0, len(modelsMap))
	for k := range modelsMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, id := range keys {
		cfg := modelsMap[id]
		fmt.Fprintf(&sb, "      %s:\n", yamlKey(id))
		fmt.Fprintf(&sb, "        name: %s\n", cfg.Name)
		if cfg.Reasoning {
			sb.WriteString("        reasoning: true\n")
		}
		fmt.Fprintf(&sb, "        tool_call: %t\n", cfg.ToolCall)
		sb.WriteString("        limit:\n")
		fmt.Fprintf(&sb, "          context: %d\n", cfg.Limit.Context)
		fmt.Fprintf(&sb, "          output: %d\n", cfg.Limit.Output)
		sb.WriteString("        modalities:\n")
		fmt.Fprintf(&sb, "          input: [%s]\n", strings.Join(cfg.Modalities.Input, ", "))
		fmt.Fprintf(&sb, "          output: [%s]\n", strings.Join(cfg.Modalities.Output, ", "))
		if cfg.Cost != nil {
			sb.WriteString("        cost:\n")
			if cfg.Cost.Input != 0 {
				fmt.Fprintf(&sb, "          input: %g\n", cfg.Cost.Input)
			}
			if cfg.Cost.Output != 0 {
				fmt.Fprintf(&sb, "          output: %g\n", cfg.Cost.Output)
			}
			if cfg.Cost.CacheRead != 0 {
				fmt.Fprintf(&sb, "          cache_read: %g\n", cfg.Cost.CacheRead)
			}
			if cfg.Cost.CacheWrite != 0 {
				fmt.Fprintf(&sb, "          cache_write: %g\n", cfg.Cost.CacheWrite)
			}
		}
		if len(cfg.Variants) > 0 {
			sb.WriteString("        variants:\n")
			variantNames := make([]string, 0, len(cfg.Variants))
			for v := range cfg.Variants {
				variantNames = append(variantNames, v)
			}
			sort.Strings(variantNames)
			for _, vname := range variantNames {
				fmt.Fprintf(&sb, "          %s:\n", vname)
				optKeys := make([]string, 0, len(cfg.Variants[vname]))
				for k := range cfg.Variants[vname] {
					optKeys = append(optKeys, k)
				}
				sort.Strings(optKeys)
				for _, ok := range optKeys {
					fmt.Fprintf(&sb, "            %s: %v\n", ok, cfg.Variants[vname][ok])
				}
			}
		}
	}

	return sb.String()
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
			fmt.Fprintf(&sb, "  - %s\n", id)
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
