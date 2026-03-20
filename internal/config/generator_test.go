package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/marstid/synmodels/internal/types"
)

func TestGetModelConfig(t *testing.T) {
	tests := []struct {
		name               string
		model              types.Model
		expectedName       string
		expectedContext    int
		expectedOutput     int
		expectedInput      []string
		expectedOutputMods []string
		expectedToolCall   bool
	}{
		{
			name: "model with text only input and no tools",
			model: types.Model{
				ID:                "gpt-3.5-turbo",
				InputModalities:   []string{"text"},
				OutputModalities:  []string{"text"},
				ContextLength:     8192,
				MaxOutputLength:   4096,
				SupportedFeatures: []string{},
			},
			expectedName:       "gpt 3.5 turbo",
			expectedContext:    8192,
			expectedOutput:     4096,
			expectedInput:      []string{"text"},
			expectedOutputMods: []string{"text"},
			expectedToolCall:   false,
		},
		{
			name: "model with image input and tools support",
			model: types.Model{
				ID:                "gpt-4-vision",
				InputModalities:   []string{"text", "image"},
				OutputModalities:  []string{"text"},
				ContextLength:     128000,
				MaxOutputLength:   4096,
				SupportedFeatures: []string{"tools"},
			},
			expectedName:       "gpt 4 vision",
			expectedContext:    128000,
			expectedOutput:     4096,
			expectedInput:      []string{"text", "image"},
			expectedOutputMods: []string{"text"},
			expectedToolCall:   true,
		},
		{
			name: "large model with all features",
			model: types.Model{
				ID:                "claude-3-opus",
				InputModalities:   []string{"text", "image"},
				OutputModalities:  []string{"text", "image"},
				ContextLength:     200000,
				MaxOutputLength:   4096,
				SupportedFeatures: []string{"tools", "vision"},
			},
			expectedName:       "claude 3 opus",
			expectedContext:    200000,
			expectedOutput:     4096,
			expectedInput:      []string{"text", "image"},
			expectedOutputMods: []string{"text", "image"},
			expectedToolCall:   true,
		},
		{
			name: "model with zero output length (uses default)",
			model: types.Model{
				ID:                "custom-model",
				InputModalities:   []string{"text"},
				OutputModalities:  []string{"text"},
				ContextLength:     4096,
				MaxOutputLength:   0,
				SupportedFeatures: []string{},
			},
			expectedName:       "custom model",
			expectedContext:    4096,
			expectedOutput:     4096, // Default value
			expectedInput:      []string{"text"},
			expectedOutputMods: []string{"text"},
			expectedToolCall:   false,
		},
		{
			name: "model with empty modalities (uses defaults)",
			model: types.Model{
				ID:                "minimal-model",
				InputModalities:   []string{},
				OutputModalities:  []string{},
				ContextLength:     8192,
				MaxOutputLength:   2048,
				SupportedFeatures: []string{},
			},
			expectedName:       "minimal model",
			expectedContext:    8192,
			expectedOutput:     2048,
			expectedInput:      []string{"text"}, // Default
			expectedOutputMods: []string{"text"}, // Default
			expectedToolCall:   false,
		},
		{
			name: "model with namespace prefix extracts last part and replaces hyphens",
			model: types.Model{
				ID:                "nvidia/Kimi-K2.5-NVFP4",
				InputModalities:   []string{"text"},
				OutputModalities:  []string{"text"},
				ContextLength:     8192,
				MaxOutputLength:   4096,
				SupportedFeatures: []string{},
			},
			expectedName:       "Kimi K2.5 NVFP4",
			expectedContext:    8192,
			expectedOutput:     4096,
			expectedInput:      []string{"text"},
			expectedOutputMods: []string{"text"},
			expectedToolCall:   false,
		},
		{
			name: "model with hf prefix and org extracts last part and replaces hyphens",
			model: types.Model{
				ID:                "hf:zai-org/GLM-4.7-Flash",
				InputModalities:   []string{"text"},
				OutputModalities:  []string{"text"},
				ContextLength:     8192,
				MaxOutputLength:   4096,
				SupportedFeatures: []string{},
			},
			expectedName:       "GLM 4.7 Flash",
			expectedContext:    8192,
			expectedOutput:     4096,
			expectedInput:      []string{"text"},
			expectedOutputMods: []string{"text"},
			expectedToolCall:   false,
		},
		{
			name: "model with multiple slashes extracts last part and replaces hyphens",
			model: types.Model{
				ID:                "hf:org/team/model-name",
				InputModalities:   []string{"text"},
				OutputModalities:  []string{"text"},
				ContextLength:     8192,
				MaxOutputLength:   4096,
				SupportedFeatures: []string{},
			},
			expectedName:       "model name",
			expectedContext:    8192,
			expectedOutput:     4096,
			expectedInput:      []string{"text"},
			expectedOutputMods: []string{"text"},
			expectedToolCall:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := GetModelConfig(tt.model)

			assert.Equal(t, tt.expectedName, config.Name, "model name should be last part after '/'")
			assert.Equal(t, tt.expectedToolCall, config.ToolCall, "tool_call should match")
			assert.Equal(t, tt.expectedContext, config.Limit.Context, "context limit should match")
			assert.Equal(t, tt.expectedOutput, config.Limit.Output, "output limit should match")
			assert.Equal(t, tt.expectedInput, config.Modalities.Input, "input modalities should match")
			assert.Equal(t, tt.expectedOutputMods, config.Modalities.Output, "output modalities should match")
		})
	}
}

func TestGetModelConfig_NameExtraction(t *testing.T) {
	// Test that model names extract the last part after the final "/" and replace hyphens with spaces
	t.Run("model without slash has hyphens replaced", func(t *testing.T) {
		model := types.Model{
			ID:              "gpt-4",
			InputModalities: []string{"text"},
			ContextLength:   8192,
		}
		config := GetModelConfig(model)
		assert.Equal(t, "gpt 4", config.Name)
	})

	t.Run("model with nvidia prefix extracts last part and replaces hyphens", func(t *testing.T) {
		model := types.Model{
			ID:              "nvidia/Kimi-K2.5-NVFP4",
			InputModalities: []string{"text"},
			ContextLength:   8192,
		}
		config := GetModelConfig(model)
		assert.Equal(t, "Kimi K2.5 NVFP4", config.Name)
	})

	t.Run("model with hf prefix and org extracts last part and replaces hyphens", func(t *testing.T) {
		model := types.Model{
			ID:              "hf:zai-org/GLM-4.7-Flash",
			InputModalities: []string{"text"},
			ContextLength:   8192,
		}
		config := GetModelConfig(model)
		assert.Equal(t, "GLM 4.7 Flash", config.Name)
	})

	t.Run("model with multiple slashes extracts last part and replaces hyphens", func(t *testing.T) {
		model := types.Model{
			ID:              "hf:org/team/model-name",
			InputModalities: []string{"text"},
			ContextLength:   8192,
		}
		config := GetModelConfig(model)
		assert.Equal(t, "model name", config.Name)
	})

	t.Run("model with MiniMax format", func(t *testing.T) {
		model := types.Model{
			ID:              "MiniMax-M2.5",
			InputModalities: []string{"text"},
			ContextLength:   8192,
		}
		config := GetModelConfig(model)
		assert.Equal(t, "MiniMax M2.5", config.Name)
	})
}

func TestGetModelConfig_CaseInsensitivity(t *testing.T) {
	t.Run("model with image input", func(t *testing.T) {
		model := types.Model{
			ID:              "MODEL-VISION-LARGE",
			InputModalities: []string{"text", "image"},
			ContextLength:   191488,
		}
		config := GetModelConfig(model)
		assert.Equal(t, 191488, config.Limit.Context)
		assert.Contains(t, config.Modalities.Input, "image")
		assert.Contains(t, config.Modalities.Input, "text")
	})

	t.Run("large model with multimodal", func(t *testing.T) {
		model := types.Model{
			ID:                "Llama-M2.5-Multimodal",
			InputModalities:   []string{"text", "image"},
			OutputModalities:  []string{"text"},
			ContextLength:     128000,
			MaxOutputLength:   4096,
			SupportedFeatures: []string{"tools"},
		}
		config := GetModelConfig(model)
		assert.Equal(t, 128000, config.Limit.Context)
		assert.Equal(t, 4096, config.Limit.Output)
		assert.Contains(t, config.Modalities.Input, "image")
		assert.True(t, config.ToolCall)
	})
}

func TestGenerator_GenerateFromSelected(t *testing.T) {
	tests := []struct {
		name        string
		models      []types.SelectedModel
		format      OutputFormat
		wantErr     bool
		errContains string
	}{
		{
			name:   "single selected model JSON",
			format: FormatJSON,
			models: []types.SelectedModel{
				{
					Model:    types.Model{ID: "gpt-4"},
					Selected: true,
				},
			},
			wantErr: false,
		},
		{
			name:   "multiple selected models JSON",
			format: FormatJSON,
			models: []types.SelectedModel{
				{
					Model:    types.Model{ID: "gpt-4"},
					Selected: true,
				},
				{
					Model:    types.Model{ID: "claude-3"},
					Selected: true,
				},
				{
					Model:    types.Model{ID: "unselected-model"},
					Selected: false,
				},
			},
			wantErr: false,
		},
		{
			name:   "single selected model YAML",
			format: FormatYAML,
			models: []types.SelectedModel{
				{
					Model:    types.Model{ID: "gpt-4"},
					Selected: true,
				},
			},
			wantErr: false,
		},
		{
			name:        "no models selected",
			format:      FormatJSON,
			models:      []types.SelectedModel{},
			wantErr:     true,
			errContains: "no models selected",
		},
		{
			name: "all models unselected",
			models: []types.SelectedModel{
				{
					Model:    types.Model{ID: "gpt-4"},
					Selected: false,
				},
				{
					Model:    types.Model{ID: "claude-3"},
					Selected: false,
				},
			},
			format:      FormatJSON,
			wantErr:     true,
			errContains: "no models selected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGenerator(tt.format)
			output, err := g.GenerateFromSelected(tt.models)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, output)

			switch tt.format {
			case FormatJSON:
				// Verify valid JSON
				var result map[string]interface{}
				err = json.Unmarshal([]byte(output), &result)
				require.NoError(t, err, "output should be valid JSON")

				// Verify structure
				modelsMap, ok := result["models"].(map[string]interface{})
				require.True(t, ok, "models should be a map")

				// Count selected models
				selectedCount := 0
				for _, m := range tt.models {
					if m.Selected {
						selectedCount++
					}
				}
				assert.Len(t, modelsMap, selectedCount, "should have correct number of models")

				// Verify each selected model is in the output
				for _, m := range tt.models {
					if m.Selected {
						modelKey := m.Model.ID
						modelConfig, exists := modelsMap[modelKey]
						require.True(t, exists, "model %s should exist in output", modelKey)
						assert.NotNil(t, modelConfig)

						// Verify model config structure
						configMap, ok := modelConfig.(map[string]interface{})
						require.True(t, ok, "model config should be a map")
						assert.Contains(t, configMap, "name")
						assert.Contains(t, configMap, "tool_call")
						assert.Contains(t, configMap, "limit")
						assert.Contains(t, configMap, "modalities")
					}
				}

			case FormatYAML:
				// Verify YAML-like output contains expected structure
				assert.Contains(t, output, "models:")
				for _, m := range tt.models {
					if m.Selected {
						modelKey := m.Model.ID
						assert.Contains(t, output, modelKey+":")
					}
				}
				assert.Contains(t, output, "name:")
				assert.Contains(t, output, "tool_call:")
				assert.Contains(t, output, "limit:")
				assert.Contains(t, output, "modalities:")
				assert.Contains(t, output, "context:")
				assert.Contains(t, output, "output:")
			}
		})
	}
}

func TestGenerator_GenerateFromSelected_OutputStructure(t *testing.T) {
	g := NewGenerator(FormatJSON)
	models := []types.SelectedModel{
		{
			Model: types.Model{
				ID:                "vision-model",
				InputModalities:   []string{"text", "image"},
				OutputModalities:  []string{"text"},
				ContextLength:     128000,
				MaxOutputLength:   4096,
				SupportedFeatures: []string{"tools"},
			},
			Selected: true,
		},
	}

	output, err := g.GenerateFromSelected(models)
	require.NoError(t, err)

	// Parse the JSON output
	var result map[string]interface{}
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)

	// Verify top-level structure
	modelsMap, ok := result["models"].(map[string]interface{})
	require.True(t, ok)

	// Verify the model config structure
	visionConfig, ok := modelsMap["vision-model"].(map[string]interface{})
	require.True(t, ok)

	// Check all required fields
	assert.Equal(t, "vision model", visionConfig["name"])
	assert.Equal(t, true, visionConfig["tool_call"])

	limit, ok := visionConfig["limit"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, limit, "context")
	assert.Contains(t, limit, "output")

	modalities, ok := visionConfig["modalities"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, modalities, "input")
	assert.Contains(t, modalities, "output")

	// Should have image input based on actual API data
	inputModalities, ok := modalities["input"].([]interface{})
	require.True(t, ok)
	assert.Contains(t, inputModalities, "text")
	assert.Contains(t, inputModalities, "image")

	outputModalities, ok := modalities["output"].([]interface{})
	require.True(t, ok)
	assert.Contains(t, outputModalities, "text")
}

func TestGenerator_GenerateFromSelected_LargeModelLimits(t *testing.T) {
	g := NewGenerator(FormatJSON)
	models := []types.SelectedModel{
		{
			Model: types.Model{
				ID:              "nvidia/llama-3.2-m2.5-32k",
				ContextLength:   191488,
				MaxOutputLength: 655360,
				InputModalities: []string{"text"},
			},
			Selected: true,
		},
	}

	output, err := g.GenerateFromSelected(models)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)

	modelsMap := result["models"].(map[string]interface{})
	modelConfig := modelsMap["nvidia/llama-3.2-m2.5-32k"].(map[string]interface{})
	limits := modelConfig["limit"].(map[string]interface{})

	assert.Equal(t, float64(191488), limits["context"])
	assert.Equal(t, float64(655360), limits["output"])
	assert.Equal(t, "llama 3.2 m2.5 32k", modelConfig["name"])
}

func TestGenerator_GenerateFromSelected_SmallModelLimits(t *testing.T) {
	g := NewGenerator(FormatJSON)
	models := []types.SelectedModel{
		{
			Model: types.Model{
				ID:              "gpt-3.5-turbo",
				ContextLength:   8192,
				MaxOutputLength: 4096,
				InputModalities: []string{"text"},
			},
			Selected: true,
		},
	}

	output, err := g.GenerateFromSelected(models)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)

	modelsMap := result["models"].(map[string]interface{})
	modelConfig := modelsMap["gpt-3.5-turbo"].(map[string]interface{})
	limits := modelConfig["limit"].(map[string]interface{})

	assert.Equal(t, float64(8192), limits["context"])
	assert.Equal(t, float64(4096), limits["output"])
}

func TestGenerator_GenerateFromSelected_UnsupportedFormat(t *testing.T) {
	g := NewGenerator(OutputFormat("xml"))
	models := []types.SelectedModel{
		{
			Model:    types.Model{ID: "gpt-4"},
			Selected: true,
		},
	}

	_, err := g.GenerateFromSelected(models)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format")
}

func TestNewGenerator(t *testing.T) {
	g := NewGenerator(FormatJSON)
	require.NotNil(t, g)
	assert.Equal(t, FormatJSON, g.format)

	g2 := NewGenerator(FormatYAML)
	require.NotNil(t, g2)
	assert.Equal(t, FormatYAML, g2.format)
}
