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
			expectedOutput:     4096,
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
			expectedInput:      []string{"text"},
			expectedOutputMods: []string{"text"},
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
			cfg := GetModelConfig(tt.model)

			assert.Equal(t, tt.expectedName, cfg.Name, "model name should be last part after '/'")
			assert.Equal(t, tt.expectedToolCall, cfg.ToolCall, "tool_call should match")
			assert.Equal(t, tt.expectedContext, cfg.Limit.Context, "context limit should match")
			assert.Equal(t, tt.expectedOutput, cfg.Limit.Output, "output limit should match")
			assert.Equal(t, tt.expectedInput, cfg.Modalities.Input, "input modalities should match")
			assert.Equal(t, tt.expectedOutputMods, cfg.Modalities.Output, "output modalities should match")
		})
	}
}

func TestGetModelConfig_NameExtraction(t *testing.T) {
	t.Run("model without slash has hyphens replaced", func(t *testing.T) {
		model := types.Model{
			ID:              "gpt-4",
			InputModalities: []string{"text"},
			ContextLength:   8192,
		}
		cfg := GetModelConfig(model)
		assert.Equal(t, "gpt 4", cfg.Name)
	})

	t.Run("model with nvidia prefix extracts last part and replaces hyphens", func(t *testing.T) {
		model := types.Model{
			ID:              "nvidia/Kimi-K2.5-NVFP4",
			InputModalities: []string{"text"},
			ContextLength:   8192,
		}
		cfg := GetModelConfig(model)
		assert.Equal(t, "Kimi K2.5 NVFP4", cfg.Name)
	})

	t.Run("model with hf prefix and org extracts last part and replaces hyphens", func(t *testing.T) {
		model := types.Model{
			ID:              "hf:zai-org/GLM-4.7-Flash",
			InputModalities: []string{"text"},
			ContextLength:   8192,
		}
		cfg := GetModelConfig(model)
		assert.Equal(t, "GLM 4.7 Flash", cfg.Name)
	})

	t.Run("model with multiple slashes extracts last part and replaces hyphens", func(t *testing.T) {
		model := types.Model{
			ID:              "hf:org/team/model-name",
			InputModalities: []string{"text"},
			ContextLength:   8192,
		}
		cfg := GetModelConfig(model)
		assert.Equal(t, "model name", cfg.Name)
	})

	t.Run("model with MiniMax format", func(t *testing.T) {
		model := types.Model{
			ID:              "MiniMax-M2.5",
			InputModalities: []string{"text"},
			ContextLength:   8192,
		}
		cfg := GetModelConfig(model)
		assert.Equal(t, "MiniMax M2.5", cfg.Name)
	})

	t.Run("prefers API name field over ID-derived name", func(t *testing.T) {
		model := types.Model{
			ID:              "syn:large:text",
			Name:            "zai-org/GLM-5.2",
			HuggingFaceID:   "zai-org/GLM-5.2",
			InputModalities: []string{"text"},
			ContextLength:   524288,
		}
		cfg := GetModelConfig(model)
		assert.Equal(t, "GLM 5.2", cfg.Name)
	})

	t.Run("falls back to hugging_face_id when no name", func(t *testing.T) {
		model := types.Model{
			ID:              "syn:small:text",
			HuggingFaceID:   "zai-org/GLM-4.7-Flash",
			InputModalities: []string{"text"},
			ContextLength:   196608,
		}
		cfg := GetModelConfig(model)
		assert.Equal(t, "GLM 4.7 Flash", cfg.Name)
	})
}

func TestGetModelConfig_CaseInsensitivity(t *testing.T) {
	t.Run("model with image input", func(t *testing.T) {
		model := types.Model{
			ID:              "MODEL-VISION-LARGE",
			InputModalities: []string{"text", "image"},
			ContextLength:   191488,
		}
		cfg := GetModelConfig(model)
		assert.Equal(t, 191488, cfg.Limit.Context)
		assert.Contains(t, cfg.Modalities.Input, "image")
		assert.Contains(t, cfg.Modalities.Input, "text")
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
		cfg := GetModelConfig(model)
		assert.Equal(t, 128000, cfg.Limit.Context)
		assert.Equal(t, 4096, cfg.Limit.Output)
		assert.Contains(t, cfg.Modalities.Input, "image")
		assert.True(t, cfg.ToolCall)
	})

	t.Run("tools feature matched case-insensitively", func(t *testing.T) {
		model := types.Model{
			ID:                "test-model",
			InputModalities:   []string{"text"},
			ContextLength:     8192,
			SupportedFeatures: []string{"TOOLS"},
		}
		cfg := GetModelConfig(model)
		assert.True(t, cfg.ToolCall)
	})

	t.Run("reasoning feature detected", func(t *testing.T) {
		model := types.Model{
			ID:                "hf:zai-org/GLM-5.2",
			HuggingFaceID:     "zai-org/GLM-5.2",
			InputModalities:   []string{"text"},
			ContextLength:     524288,
			MaxOutputLength:   65536,
			SupportedFeatures: []string{"tools", "json_mode", "structured_outputs", "reasoning"},
		}
		cfg := GetModelConfig(model)
		assert.True(t, cfg.Reasoning)
		assert.True(t, cfg.ToolCall)
	})
}

func TestGetModelConfig_Cost(t *testing.T) {
	t.Run("pricing strings parsed to numeric cost", func(t *testing.T) {
		model := types.Model{
			ID:              "hf:zai-org/GLM-5.2",
			HuggingFaceID:   "zai-org/GLM-5.2",
			InputModalities: []string{"text"},
			ContextLength:   524288,
			MaxOutputLength: 65536,
			Pricing: types.Pricing{
				Prompt:           "$0.0000014",
				Completion:       "$0.0000044",
				InputCacheReads:  "$0.0000014",
				InputCacheWrites: "0",
			},
		}
		cfg := GetModelConfig(model)
		require.NotNil(t, cfg.Cost)
		assert.Equal(t, 0.0000014, cfg.Cost.Input)
		assert.Equal(t, 0.0000044, cfg.Cost.Output)
		assert.Equal(t, 0.0000014, cfg.Cost.CacheRead)
		assert.Equal(t, float64(0), cfg.Cost.CacheWrite)
	})

	t.Run("zero pricing omitted", func(t *testing.T) {
		model := types.Model{
			ID:              "free-model",
			InputModalities: []string{"text"},
			ContextLength:   8192,
			Pricing: types.Pricing{
				Prompt:     "0",
				Completion: "0",
			},
		}
		cfg := GetModelConfig(model)
		assert.Nil(t, cfg.Cost)
	})
}

func TestGetModelConfig_Variants(t *testing.T) {
	t.Run("GLM-5.2 gets variants from registry", func(t *testing.T) {
		model := types.Model{
			ID:                "hf:zai-org/GLM-5.2",
			HuggingFaceID:     "zai-org/GLM-5.2",
			InputModalities:   []string{"text"},
			ContextLength:     524288,
			MaxOutputLength:   65536,
			SupportedFeatures: []string{"tools", "json_mode", "structured_outputs", "reasoning"},
		}
		cfg := GetModelConfig(model)
		assert.True(t, cfg.Reasoning)
		require.NotNil(t, cfg.Variants)
		assert.Contains(t, cfg.Variants, "none")
		assert.Contains(t, cfg.Variants, "high")
		assert.Contains(t, cfg.Variants, "max")
		assert.Equal(t, "none", cfg.Variants["none"]["reasoningEffort"])
		assert.Equal(t, "high", cfg.Variants["high"]["reasoningEffort"])
		assert.Equal(t, "xhigh", cfg.Variants["max"]["reasoningEffort"])
	})

	t.Run("syn-style ID resolves variants via hugging_face_id", func(t *testing.T) {
		model := types.Model{
			ID:                "syn:large:text",
			HuggingFaceID:     "zai-org/GLM-5.2",
			Name:              "syn:large:text",
			InputModalities:   []string{"text"},
			ContextLength:     524288,
			MaxOutputLength:   65536,
			SupportedFeatures: []string{"tools", "reasoning"},
		}
		cfg := GetModelConfig(model)
		assert.True(t, cfg.Reasoning)
		require.NotNil(t, cfg.Variants)
		assert.Contains(t, cfg.Variants, "none")
		assert.Contains(t, cfg.Variants, "max")
	})

	t.Run("non-registered model gets no variants", func(t *testing.T) {
		model := types.Model{
			ID:                "gpt-4",
			InputModalities:   []string{"text"},
			ContextLength:     8192,
			SupportedFeatures: []string{"tools"},
		}
		cfg := GetModelConfig(model)
		assert.Nil(t, cfg.Variants)
		assert.False(t, cfg.Reasoning)
	})
}

func TestGetModelConfig_ContextDefault(t *testing.T) {
	t.Run("zero context length gets default 8192", func(t *testing.T) {
		model := types.Model{
			ID:              "test-model",
			InputModalities: []string{"text"},
			ContextLength:   0,
			MaxOutputLength: 4096,
		}
		cfg := GetModelConfig(model)
		assert.Equal(t, 8192, cfg.Limit.Context)
		assert.Equal(t, 4096, cfg.Limit.Output)
	})
}

func TestOpencodeModelKey(t *testing.T) {
	tests := []struct {
		name     string
		model    types.Model
		expected string
	}{
		{
			name:     "hf prefix used as-is",
			model:    types.Model{ID: "hf:zai-org/GLM-5.2"},
			expected: "hf:zai-org/GLM-5.2",
		},
		{
			name:     "syn ID with hugging_face_id gets hf prefix",
			model:    types.Model{ID: "syn:large:text", HuggingFaceID: "zai-org/GLM-5.2"},
			expected: "hf:zai-org/GLM-5.2",
		},
		{
			name:     "bare ID without hugging_face_id stays as-is",
			model:    types.Model{ID: "gpt-4"},
			expected: "gpt-4",
		},
		{
			name:     "org/model without hf prefix and no hugging_face_id stays as-is",
			model:    types.Model{ID: "nvidia/Kimi-K2.5-NVFP4"},
			expected: "nvidia/Kimi-K2.5-NVFP4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, OpencodeModelKey(tt.model))
		})
	}
}

// modelsFromJSON navigates the provider-wrapped JSON structure to extract the models map.
func modelsFromJSON(t *testing.T, output string) map[string]interface{} {
	t.Helper()
	var result map[string]interface{}
	err := json.Unmarshal([]byte(output), &result)
	require.NoError(t, err, "output should be valid JSON")

	provider, ok := result["provider"].(map[string]interface{})
	require.True(t, ok, "should have provider key")

	synthetic, ok := provider["synthetic"].(map[string]interface{})
	require.True(t, ok, "should have synthetic provider")

	modelsMap, ok := synthetic["models"].(map[string]interface{})
	require.True(t, ok, "models should be a map")

	return modelsMap
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
				modelsMap := modelsFromJSON(t, output)

				selectedCount := 0
				for _, m := range tt.models {
					if m.Selected {
						selectedCount++
					}
				}
				assert.Len(t, modelsMap, selectedCount, "should have correct number of models")

				for _, m := range tt.models {
					if m.Selected {
						modelKey := OpencodeModelKey(m.Model)
						modelConfig, exists := modelsMap[modelKey]
						require.True(t, exists, "model %s should exist in output", modelKey)
						assert.NotNil(t, modelConfig)

						configMap, ok := modelConfig.(map[string]interface{})
						require.True(t, ok, "model config should be a map")
						assert.Contains(t, configMap, "name")
						assert.Contains(t, configMap, "tool_call")
						assert.Contains(t, configMap, "limit")
						assert.Contains(t, configMap, "modalities")
					}
				}

			case FormatYAML:
				assert.Contains(t, output, "models:")
				for _, m := range tt.models {
					if m.Selected {
						modelKey := OpencodeModelKey(m.Model)
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

	modelsMap := modelsFromJSON(t, output)

	visionConfig, ok := modelsMap["vision-model"].(map[string]interface{})
	require.True(t, ok)

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

	modelsMap := modelsFromJSON(t, output)
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

	modelsMap := modelsFromJSON(t, output)
	modelConfig := modelsMap["gpt-3.5-turbo"].(map[string]interface{})
	limits := modelConfig["limit"].(map[string]interface{})

	assert.Equal(t, float64(8192), limits["context"])
	assert.Equal(t, float64(4096), limits["output"])
}

func TestGenerator_GenerateFromSelected_CostInOutput(t *testing.T) {
	g := NewGenerator(FormatJSON)
	models := []types.SelectedModel{
		{
			Model: types.Model{
				ID:              "hf:zai-org/GLM-5.2",
				HuggingFaceID:   "zai-org/GLM-5.2",
				InputModalities: []string{"text"},
				ContextLength:   524288,
				MaxOutputLength: 65536,
				Pricing: types.Pricing{
					Prompt:     "$0.0000014",
					Completion: "$0.0000044",
				},
			},
			Selected: true,
		},
	}

	output, err := g.GenerateFromSelected(models)
	require.NoError(t, err)

	modelsMap := modelsFromJSON(t, output)
	modelConfig := modelsMap["hf:zai-org/GLM-5.2"].(map[string]interface{})

	cost, ok := modelConfig["cost"].(map[string]interface{})
	require.True(t, ok, "should have cost field")
	assert.Equal(t, 0.0000014, cost["input"])
	assert.Equal(t, 0.0000044, cost["output"])

	_, hasPricing := modelConfig["pricing"]
	assert.False(t, hasPricing, "should not have legacy pricing field")
}

func TestGenerator_GenerateFromSelected_VariantsInOutput(t *testing.T) {
	g := NewGenerator(FormatJSON)
	models := []types.SelectedModel{
		{
			Model: types.Model{
				ID:                "hf:zai-org/GLM-5.2",
				HuggingFaceID:     "zai-org/GLM-5.2",
				InputModalities:   []string{"text"},
				ContextLength:     524288,
				MaxOutputLength:   65536,
				SupportedFeatures: []string{"tools", "reasoning"},
			},
			Selected: true,
		},
	}

	output, err := g.GenerateFromSelected(models)
	require.NoError(t, err)

	modelsMap := modelsFromJSON(t, output)
	modelConfig := modelsMap["hf:zai-org/GLM-5.2"].(map[string]interface{})

	assert.Equal(t, true, modelConfig["reasoning"])

	variants, ok := modelConfig["variants"].(map[string]interface{})
	require.True(t, ok, "should have variants field")
	assert.Contains(t, variants, "none")
	assert.Contains(t, variants, "high")
	assert.Contains(t, variants, "max")
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
	assert.Equal(t, defaultBaseURL, g.baseURL)

	g2 := NewGenerator(FormatYAML)
	require.NotNil(t, g2)
	assert.Equal(t, FormatYAML, g2.format)
}

func TestNewGenerator_WithBaseURL(t *testing.T) {
	g := NewGenerator(FormatJSON).WithBaseURL("https://custom.example.com/v1")
	assert.Equal(t, "https://custom.example.com/v1", g.baseURL)

	g2 := NewGenerator(FormatJSON).WithBaseURL("")
	assert.Equal(t, defaultBaseURL, g2.baseURL)
}
