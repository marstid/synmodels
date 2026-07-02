package opencode

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/marstid/synmodels/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	require.NotNil(t, m)
	assert.NotEmpty(t, m.configPath)
	assert.True(t, strings.HasSuffix(m.configPath, ConfigFilename))
}

func TestNewManagerWithPath(t *testing.T) {
	customPath := "/custom/path/config.json"
	m := NewManagerWithPath(customPath)
	require.NotNil(t, m)
	assert.Equal(t, customPath, m.configPath)
}

func TestExpandPath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "path with tilde",
			input:    "~/.config/opencode/config.json",
			expected: filepath.Join(homeDir, ".config/opencode/config.json"),
		},
		{
			name:     "absolute path",
			input:    "/absolute/path/config.json",
			expected: "/absolute/path/config.json",
		},
		{
			name:     "relative path",
			input:    "./relative/path/config.json",
			expected: "./relative/path/config.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandPath(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestManager_Read_ConfigNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent", ConfigFilename)

	m := NewManagerWithPath(configPath)
	cfg, err := m.Read()

	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.data)
	assert.Empty(t, cfg.data)
}

func TestManager_Read_ExistingConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ConfigFilename)

	// Create a test config file with extra fields to verify preservation
	testConfig := map[string]interface{}{
		"$schema":          "./opencode-schema.json",
		"small_model":      "test-small-model",
		"default_provider": "synthetic",
		"mcp": map[string]interface{}{
			"servers": []interface{}{
				map[string]interface{}{
					"name": "test-server",
					"url":  "http://localhost:3000",
				},
			},
		},
		"provider": map[string]interface{}{
			"synthetic": map[string]interface{}{
				"npm":  "@ai-sdk/openai-compatible",
				"name": "Synthetic",
				"options": map[string]interface{}{
					"baseURL": "https://api.synthetic.new/openai/v1",
					"apiKey":  "test-key",
				},
				"models": map[string]interface{}{
					"hf:test-model": map[string]interface{}{
						"name": "test-model",
						"limit": map[string]interface{}{
							"context": float64(8192),
							"output":  float64(4096),
						},
						"modalities": map[string]interface{}{
							"input":  []interface{}{"text"},
							"output": []interface{}{"text"},
						},
					},
				},
			},
		},
	}

	data, err := json.MarshalIndent(testConfig, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(configPath, data, 0o644)
	require.NoError(t, err)

	m := NewManagerWithPath(configPath)
	cfg, err := m.Read()

	require.NoError(t, err)
	assert.NotNil(t, cfg)

	// Verify extra fields are preserved
	assert.Equal(t, "./opencode-schema.json", cfg.data["$schema"])
	assert.Equal(t, "test-small-model", cfg.data["small_model"])
	assert.Equal(t, "synthetic", cfg.data["default_provider"])

	// Verify mcp is preserved
	mcpData, ok := cfg.data["mcp"].(map[string]interface{})
	require.True(t, ok)
	servers, ok := mcpData["servers"].([]interface{})
	require.True(t, ok)
	assert.Len(t, servers, 1)

	// Verify providers are accessible
	providers := cfg.GetProviders()
	assert.Len(t, providers, 1)

	synthetic, ok := providers["synthetic"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "@ai-sdk/openai-compatible", synthetic["npm"])
	assert.Equal(t, "Synthetic", synthetic["name"])
}

func TestManager_Read_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ConfigFilename)

	// Write invalid JSON
	err := os.WriteFile(configPath, []byte("{invalid json}"), 0o644)
	require.NoError(t, err)

	m := NewManagerWithPath(configPath)
	cfg, err := m.Read()

	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "failed to parse config file")
}

func TestManager_Write_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "subdir1", "subdir2", ConfigFilename)

	cfg := &Config{
		data: map[string]interface{}{
			"provider": map[string]interface{}{
				"synthetic": map[string]interface{}{
					"npm":    "@ai-sdk/openai-compatible",
					"name":   "Synthetic",
					"models": map[string]interface{}{},
				},
			},
		},
	}

	m := NewManagerWithPath(configPath)
	err := m.Write(cfg)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(configPath)
	assert.NoError(t, err)

	// Verify content
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var readCfg map[string]interface{}
	err = json.Unmarshal(data, &readCfg)
	require.NoError(t, err)

	provider, ok := readCfg["provider"].(map[string]interface{})
	require.True(t, ok)
	assert.Len(t, provider, 1)
}

func TestManager_Write_CreatesBackup(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ConfigFilename)

	// Write initial config with extra fields
	initialConfig := map[string]interface{}{
		"$schema":     "./schema.json",
		"small_model": "original-model",
		"provider": map[string]interface{}{
			"synthetic": map[string]interface{}{
				"models": map[string]interface{}{},
			},
		},
	}
	initialData, err := json.MarshalIndent(initialConfig, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(configPath, initialData, 0o644)
	require.NoError(t, err)

	// Write new config
	newConfig := &Config{
		data: map[string]interface{}{
			"provider": map[string]interface{}{
				"synthetic": map[string]interface{}{
					"npm":    "@ai-sdk/openai-compatible",
					"name":   "Synthetic",
					"models": map[string]interface{}{},
				},
			},
		},
	}

	m := NewManagerWithPath(configPath)
	err = m.Write(newConfig)
	require.NoError(t, err)

	// Check that backup was created
	entries, err := os.ReadDir(tmpDir)
	require.NoError(t, err)

	var backupFound bool
	for _, entry := range entries {
		if strings.Contains(entry.Name(), BackupSuffix) {
			backupFound = true
			// Verify backup content still has original fields
			backupData, err := os.ReadFile(filepath.Join(tmpDir, entry.Name()))
			require.NoError(t, err)
			var backupConfig map[string]interface{}
			err = json.Unmarshal(backupData, &backupConfig)
			require.NoError(t, err)
			assert.Equal(t, "./schema.json", backupConfig["$schema"])
			assert.Equal(t, "original-model", backupConfig["small_model"])
			break
		}
	}
	assert.True(t, backupFound, "Backup file should be created")
}

func TestManager_Write_PreservesExistingStructure(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ConfigFilename)

	// Create initial config with multiple providers and extra fields
	initialConfig := map[string]interface{}{
		"$schema":          "./opencode-schema.json",
		"small_model":      "test-small",
		"default_provider": "openai",
		"mcp": map[string]interface{}{
			"servers": []interface{}{
				map[string]interface{}{
					"name": "server1",
					"url":  "http://localhost:3000",
				},
			},
		},
		"settings": map[string]interface{}{
			"temperature": 0.7,
			"max_tokens":  2048,
		},
		"provider": map[string]interface{}{
			"openai": map[string]interface{}{
				"npm":  "@ai-sdk/openai",
				"name": "OpenAI",
				"options": map[string]interface{}{
					"apiKey": "sk-test",
				},
				"models": map[string]interface{}{},
			},
			"anthropic": map[string]interface{}{
				"npm":  "@ai-sdk/anthropic",
				"name": "Anthropic",
				"options": map[string]interface{}{
					"apiKey": "sk-ant",
				},
				"models": map[string]interface{}{},
			},
		},
	}

	initialData, err := json.MarshalIndent(initialConfig, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(configPath, initialData, 0o644)
	require.NoError(t, err)

	// Add models to synthetic provider
	m := NewManagerWithPath(configPath)
	models := map[string]config.ModelConfig{
		"hf:new-model": {
			Name:  "new-model",
			Limit: config.ModelLimits{Context: 8192, Output: 4096},
			Modalities: config.ModelModalities{
				Input:  []string{"text"},
				Output: []string{"text"},
			},
		},
	}

	err = m.AddModels(models)
	require.NoError(t, err)

	// Read the file as raw JSON to verify all fields are preserved
	resultData, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(resultData, &result)
	require.NoError(t, err)

	// Verify all extra fields are preserved
	assert.Equal(t, "./opencode-schema.json", result["$schema"], "$schema should be preserved")
	assert.Equal(t, "test-small", result["small_model"], "small_model should be preserved")
	assert.Equal(t, "openai", result["default_provider"], "default_provider should be preserved")

	// Verify mcp is preserved
	mcpResult, ok := result["mcp"].(map[string]interface{})
	require.True(t, ok, "mcp should be preserved")
	servers, ok := mcpResult["servers"].([]interface{})
	require.True(t, ok)
	assert.Len(t, servers, 1, "mcp.servers should be preserved")

	// Verify settings are preserved
	settings, ok := result["settings"].(map[string]interface{})
	require.True(t, ok, "settings should be preserved")
	assert.Equal(t, float64(0.7), settings["temperature"], "settings.temperature should be preserved")
	assert.Equal(t, float64(2048), settings["max_tokens"], "settings.max_tokens should be preserved")

	// Verify provider map is preserved
	provider, ok := result["provider"].(map[string]interface{})
	require.True(t, ok)
	assert.Len(t, provider, 3, "should have 3 providers")

	// Verify original providers still exist
	assert.Contains(t, provider, "openai", "openai provider should be preserved")
	assert.Contains(t, provider, "anthropic", "anthropic provider should be preserved")
	assert.Contains(t, provider, "synthetic", "synthetic provider should be added")

	// Verify the new model was added
	synthetic, ok := provider["synthetic"].(map[string]interface{})
	require.True(t, ok)
	syntheticModels, ok := synthetic["models"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, syntheticModels, "hf:new-model", "new model should be added")
}

func TestManager_AddModels_NewProvider(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ConfigFilename)

	m := NewManagerWithPath(configPath)

	models := map[string]config.ModelConfig{
		"hf:model1": {
			Name:  "model1",
			Limit: config.ModelLimits{Context: 8192, Output: 4096},
			Modalities: config.ModelModalities{
				Input:  []string{"text"},
				Output: []string{"text"},
			},
		},
		"hf:model2": {
			Name:  "model2",
			Limit: config.ModelLimits{Context: 191488, Output: 655360},
			Modalities: config.ModelModalities{
				Input:  []string{"text", "image"},
				Output: []string{"text"},
			},
		},
	}

	err := m.AddModels(models)
	require.NoError(t, err)

	// Verify config was written
	cfg, err := m.Read()
	require.NoError(t, err)

	providers := cfg.GetProviders()
	synthetic, exists := providers["synthetic"].(map[string]interface{})
	require.True(t, exists)
	assert.Equal(t, "@ai-sdk/openai-compatible", synthetic["npm"])
	assert.Equal(t, "Synthetic", synthetic["name"])

	options, ok := synthetic["options"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "https://api.synthetic.new/openai/v1", options["baseURL"])

	modelsMap, ok := synthetic["models"].(map[string]interface{})
	require.True(t, ok)
	assert.Len(t, modelsMap, 2)

	// Verify models
	model1, exists := modelsMap["hf:model1"].(map[string]interface{})
	require.True(t, exists)
	assert.Equal(t, "model1", model1["name"])

	model2, exists := modelsMap["hf:model2"].(map[string]interface{})
	require.True(t, exists)
	assert.Equal(t, "model2", model2["name"])
}

func TestManager_AddModels_ExistingProvider(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ConfigFilename)

	// Create initial config with synthetic provider and extra fields
	initialConfig := map[string]interface{}{
		"$schema":     "./opencode-schema.json",
		"small_model": "original-small",
		"mcp": map[string]interface{}{
			"servers": []interface{}{},
		},
		"provider": map[string]interface{}{
			"synthetic": map[string]interface{}{
				"npm":  "@ai-sdk/openai-compatible",
				"name": "Synthetic",
				"options": map[string]interface{}{
					"baseURL": "https://api.synthetic.new/openai/v1",
					"apiKey":  "existing-key",
				},
				"models": map[string]interface{}{
					"hf:existing-model": map[string]interface{}{
						"name": "existing-model",
						"limit": map[string]interface{}{
							"context": float64(4096),
							"output":  float64(2048),
						},
						"modalities": map[string]interface{}{
							"input":  []interface{}{"text"},
							"output": []interface{}{"text"},
						},
					},
				},
			},
		},
	}

	m := NewManagerWithPath(configPath)
	cfg := &Config{data: initialConfig}
	err := m.Write(cfg)
	require.NoError(t, err)

	// Add new models (one with same ID should overwrite)
	newModels := map[string]config.ModelConfig{
		"hf:existing-model": {
			Name:  "existing-model-updated",
			Limit: config.ModelLimits{Context: 8192, Output: 4096},
			Modalities: config.ModelModalities{
				Input:  []string{"text", "image"},
				Output: []string{"text"},
			},
		},
		"hf:new-model": {
			Name:  "new-model",
			Limit: config.ModelLimits{Context: 191488, Output: 655360},
			Modalities: config.ModelModalities{
				Input:  []string{"text"},
				Output: []string{"text"},
			},
		},
	}

	err = m.AddModels(newModels)
	require.NoError(t, err)

	// Verify by reading raw JSON
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	// Verify extra fields are preserved
	assert.Equal(t, "./opencode-schema.json", result["$schema"], "$schema should be preserved")
	assert.Equal(t, "original-small", result["small_model"], "small_model should be preserved")
	assert.Contains(t, result, "mcp", "mcp should be preserved")

	// Verify provider data
	provider := result["provider"].(map[string]interface{})
	synthetic := provider["synthetic"].(map[string]interface{})
	modelsMap := synthetic["models"].(map[string]interface{})

	assert.Len(t, modelsMap, 2)

	// Verify existing model was overwritten
	existingModel := modelsMap["hf:existing-model"].(map[string]interface{})
	assert.Equal(t, "existing-model-updated", existingModel["name"])
	existingLimit := existingModel["limit"].(map[string]interface{})
	assert.Equal(t, float64(8192), existingLimit["context"])
	existingModalities := existingModel["modalities"].(map[string]interface{})
	assert.Equal(t, []interface{}{"text", "image"}, existingModalities["input"])

	// Verify API key was preserved
	options := synthetic["options"].(map[string]interface{})
	assert.Equal(t, "existing-key", options["apiKey"])
}

func TestManager_AddModels_EmptyModels(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ConfigFilename)

	m := NewManagerWithPath(configPath)
	err := m.AddModels(map[string]config.ModelConfig{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no models to add")
}

func TestManager_AddModels_PreservesAllExtraFields(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ConfigFilename)

	// Create a realistic opencode.json with many fields
	initialConfig := map[string]interface{}{
		"$schema":                 "./opencode-schema.json",
		"small_model":             "gpt-4o-mini",
		"medium_model":            "gpt-4o",
		"large_model":             "gpt-4o",
		"default_provider":        "synthetic",
		"include_commit_message":  true,
		"include_current_changes": true,
		"enable_editor":           true,
		"enable_wobbly_windows":   false,
		"code_theme":              "github-dark",
		"custom": map[string]interface{}{
			"user_name": "test-user",
		},
		"mcp": map[string]interface{}{
			"servers": []interface{}{
				map[string]interface{}{
					"name": "filesystem",
					"url":  "http://localhost:3001",
				},
				map[string]interface{}{
					"name": "github",
					"url":  "http://localhost:3002",
				},
			},
		},
		"settings": map[string]interface{}{
			"temperature":    0.7,
			"max_tokens":     4096,
			"context_window": 128000,
		},
		"provider": map[string]interface{}{
			"openai": map[string]interface{}{
				"npm":  "@ai-sdk/openai",
				"name": "OpenAI",
				"options": map[string]interface{}{
					"apiKey": "sk-xxx",
				},
				"models": map[string]interface{}{},
			},
		},
	}

	data, err := json.MarshalIndent(initialConfig, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(configPath, data, 0o644)
	require.NoError(t, err)

	// Add models
	m := NewManagerWithPath(configPath)
	models := map[string]config.ModelConfig{
		"hf:test-model": {
			Name:  "test-model",
			Limit: config.ModelLimits{Context: 8192, Output: 4096},
			Modalities: config.ModelModalities{
				Input:  []string{"text"},
				Output: []string{"text"},
			},
		},
	}

	err = m.AddModels(models)
	require.NoError(t, err)

	// Read back and verify all fields
	resultData, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(resultData, &result)
	require.NoError(t, err)

	// Verify every single field is preserved
	assert.Equal(t, "./opencode-schema.json", result["$schema"], "$schema should be preserved")
	assert.Equal(t, "gpt-4o-mini", result["small_model"], "small_model should be preserved")
	assert.Equal(t, "gpt-4o", result["medium_model"], "medium_model should be preserved")
	assert.Equal(t, "gpt-4o", result["large_model"], "large_model should be preserved")
	assert.Equal(t, "synthetic", result["default_provider"], "default_provider should be preserved")
	assert.Equal(t, true, result["include_commit_message"], "include_commit_message should be preserved")
	assert.Equal(t, true, result["include_current_changes"], "include_current_changes should be preserved")
	assert.Equal(t, true, result["enable_editor"], "enable_editor should be preserved")
	assert.Equal(t, false, result["enable_wobbly_windows"], "enable_wobbly_windows should be preserved")
	assert.Equal(t, "github-dark", result["code_theme"], "code_theme should be preserved")

	// Verify custom object is preserved
	custom, ok := result["custom"].(map[string]interface{})
	require.True(t, ok, "custom should be preserved")
	assert.Equal(t, "test-user", custom["user_name"], "custom.user_name should be preserved")

	// Verify mcp is preserved
	mcp, ok := result["mcp"].(map[string]interface{})
	require.True(t, ok, "mcp should be preserved")
	servers, ok := mcp["servers"].([]interface{})
	require.True(t, ok)
	assert.Len(t, servers, 2, "mcp.servers should have 2 items")

	// Verify settings are preserved
	settings, ok := result["settings"].(map[string]interface{})
	require.True(t, ok, "settings should be preserved")
	assert.Equal(t, float64(0.7), settings["temperature"], "settings.temperature should be preserved")
	assert.Equal(t, float64(4096), settings["max_tokens"], "settings.max_tokens should be preserved")
	assert.Equal(t, float64(128000), settings["context_window"], "settings.context_window should be preserved")

	// Verify provider map is preserved and synthetic was added
	provider, ok := result["provider"].(map[string]interface{})
	require.True(t, ok)
	assert.Len(t, provider, 2, "should have 2 providers (openai + synthetic)")
	assert.Contains(t, provider, "openai", "openai provider should be preserved")
	assert.Contains(t, provider, "synthetic", "synthetic provider should be added")

	// Verify the new model was added to synthetic
	synthetic := provider["synthetic"].(map[string]interface{})
	syntheticModels := synthetic["models"].(map[string]interface{})
	assert.Contains(t, syntheticModels, "hf:test-model", "new model should be in synthetic provider")
}

func TestManager_GetConfigPath(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ConfigFilename)

	m := NewManagerWithPath("~/" + strings.TrimPrefix(configPath, os.TempDir()))
	path := m.GetConfigPath()

	// Should expand the path
	assert.NotContains(t, path, "~")
}

func TestManager_ConfigExists(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ConfigFilename)

	m := NewManagerWithPath(configPath)

	// Should return false when file doesn't exist
	assert.False(t, m.ConfigExists())

	// Create the file
	err := os.WriteFile(configPath, []byte("{}"), 0o644)
	require.NoError(t, err)

	// Should return true when file exists
	assert.True(t, m.ConfigExists())
}

func TestManager_Read_NilProvidersMap(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ConfigFilename)

	// Write config with null provider and extra fields
	configWithNullProvider := map[string]interface{}{
		"$schema":     "./schema.json",
		"small_model": "test",
		"provider":    nil,
	}
	data, err := json.MarshalIndent(configWithNullProvider, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(configPath, data, 0o644)
	require.NoError(t, err)

	m := NewManagerWithPath(configPath)
	cfg, err := m.Read()

	require.NoError(t, err)
	// Extra fields should still be preserved
	assert.Equal(t, "./schema.json", cfg.data["$schema"])
	assert.Equal(t, "test", cfg.data["small_model"])
	// provider would be nil, GetProviders returns empty map
	assert.Empty(t, cfg.GetProviders())
}

func TestManager_Write_BackupPermissions(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ConfigFilename)

	// Write initial config
	initialConfig := map[string]interface{}{
		"$schema": "./schema.json",
		"provider": map[string]interface{}{
			"synthetic": map[string]interface{}{
				"models": map[string]interface{}{},
			},
		},
	}
	initialData, err := json.MarshalIndent(initialConfig, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(configPath, initialData, 0o644)
	require.NoError(t, err)

	// Write new config to trigger a backup
	newConfig := &Config{
		data: map[string]interface{}{
			"provider": map[string]interface{}{
				"synthetic": map[string]interface{}{
					"npm":    "@ai-sdk/openai-compatible",
					"name":   "Synthetic",
					"models": map[string]interface{}{},
				},
			},
		},
	}

	m := NewManagerWithPath(configPath)
	err = m.Write(newConfig)
	require.NoError(t, err)

	// Find the backup file and check its permissions
	entries, err := os.ReadDir(tmpDir)
	require.NoError(t, err)

	var backupFound bool
	for _, entry := range entries {
		if strings.Contains(entry.Name(), BackupSuffix) {
			backupFound = true
			info, err := os.Stat(filepath.Join(tmpDir, entry.Name()))
			require.NoError(t, err)
			assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
			break
		}
	}
	assert.True(t, backupFound, "Backup file should be created")
}

func TestManager_CheckConfigStatus_UnreadableConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ConfigFilename)

	// Write invalid JSON to the config file
	err := os.WriteFile(configPath, []byte("{invalid json}"), 0o644)
	require.NoError(t, err)

	m := NewManagerWithPath(configPath)
	status := m.CheckConfigStatus()

	assert.Equal(t, ConfigUnreadable, status)
	assert.NotEqual(t, ConfigExistsNoProvider, status)
}
