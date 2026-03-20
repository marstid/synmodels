// Package opencode provides functionality for managing the opencode CLI configuration.
package opencode

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/marstid/synmodels/internal/config"
)

const (
	// DefaultConfigDir is the default directory for opencode config.
	DefaultConfigDir = "~/.config/opencode"
	// ConfigFilename is the name of the opencode config file.
	ConfigFilename = "opencode.json"
	// BackupSuffix is the suffix added to backup files.
	BackupSuffix = "-backup"
	// EnvConfigPath is the environment variable for custom config path.
	EnvConfigPath = "OPENCODE_CONFIG"
	// EnvAPIBaseURL is the environment variable for custom API base URL.
	EnvAPIBaseURL = "SYN_API"
	// DefaultBaseURL is the default Synthetic API base URL.
	DefaultBaseURL = "https://api.synthetic.new/openai/v1"
)

// ConfigStatus represents the state of the opencode config.
type ConfigStatus int

const (
	// ConfigNotFound indicates the config file doesn't exist.
	ConfigNotFound ConfigStatus = iota
	// ConfigExistsNoProvider indicates config exists but no synthetic provider.
	ConfigExistsNoProvider
	// ConfigExistsWithProvider indicates config exists with synthetic provider.
	ConfigExistsWithProvider
)

// Config represents the root structure of the opencode.json file.
// It uses a map to preserve all fields, not just the ones we know about.
type Config struct {
	data map[string]interface{}
}

// Manager handles reading and writing the opencode configuration.
type Manager struct {
	configPath string
}

// NewManager creates a new ConfigManager with the default config path.
func NewManager() *Manager {
	return NewManagerWithPath("")
}

// NewManagerWithPath creates a new ConfigManager with a custom config path.
// If configPath is empty, it uses the default path (~/.config/opencode/opencode.json).
func NewManagerWithPath(configPath string) *Manager {
	if configPath == "" {
		configPath = getDefaultConfigPath()
	}
	return &Manager{
		configPath: configPath,
	}
}

// getDefaultConfigPath returns the default path to the opencode config file.
// Checks OPENCODE_CONFIG environment variable first, then falls back to default location.
func getDefaultConfigPath() string {
	// Check for environment variable override
	if envPath := os.Getenv(EnvConfigPath); envPath != "" {
		return envPath
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fall back to relative path if we can't get home dir
		return filepath.Join(".config", "opencode", ConfigFilename)
	}
	return filepath.Join(homeDir, ".config", "opencode", ConfigFilename)
}

// expandPath expands a path that may start with ~ to the full home directory path.
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(homeDir, path[2:])
		}
	}
	return path
}

// Read reads the opencode configuration from disk.
// If the file doesn't exist, it returns an empty config.
func (m *Manager) Read() (*Config, error) {
	configPath := expandPath(m.configPath)

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return empty config if file doesn't exist
		return &Config{
			data: make(map[string]interface{}),
		}, nil
	}

	// Read the file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse the JSON into a map to preserve all fields
	var rawData map[string]interface{}
	if err := json.Unmarshal(data, &rawData); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Ensure we have a map (in case JSON was null)
	if rawData == nil {
		rawData = make(map[string]interface{})
	}

	return &Config{data: rawData}, nil
}

// Write writes the configuration to disk with a backup.
func (m *Manager) Write(cfg *Config) error {
	configPath := expandPath(m.configPath)

	// Ensure the directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create backup if the file exists
	if _, err := os.Stat(configPath); err == nil {
		backupPath := configPath + BackupSuffix
		if err := m.createBackup(configPath, backupPath); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
	}

	// Marshal the config with proper formatting
	data, err := json.MarshalIndent(cfg.data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Add trailing newline
	data = append(data, '\n')

	// Write to file atomically
	tempPath := configPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write temp config file: %w", err)
	}

	// Rename temp file to final path
	if err := os.Rename(tempPath, configPath); err != nil {
		// Clean up temp file on error
		os.Remove(tempPath)
		return fmt.Errorf("failed to rename config file: %w", err)
	}

	return nil
}

// createBackup creates a backup of the config file with a timestamp.
func (m *Manager) createBackup(sourcePath, backupPath string) error {
	// Add timestamp to backup path
	timestamp := time.Now().Format("20060102-150405")
	backupPathWithTimestamp := fmt.Sprintf("%s.%s", backupPath, timestamp)

	// Read the source file
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to read source file for backup: %w", err)
	}

	// Write to backup file
	if err := os.WriteFile(backupPathWithTimestamp, data, 0o644); err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}

	return nil
}

// AddModels replaces all models in the synthetic provider with the provided models.
// If the synthetic provider doesn't exist, it creates one with default values.
// All existing models are removed and replaced with only the newly selected models.
// All other configuration fields (provider metadata, options, etc.) are preserved.
func (m *Manager) AddModels(models map[string]config.ModelConfig) error {
	if len(models) == 0 {
		return fmt.Errorf("no models to add")
	}

	// Read existing config
	cfg, err := m.Read()
	if err != nil {
		return err
	}

	// Ensure provider map exists
	provider, ok := cfg.data["provider"].(map[string]interface{})
	if !ok || provider == nil {
		provider = make(map[string]interface{})
		cfg.data["provider"] = provider
	}

	// Get or create the synthetic provider
	synthetic, ok := provider["synthetic"].(map[string]interface{})
	if !ok || synthetic == nil {
		// Create default synthetic provider
		synthetic = map[string]interface{}{
			"npm":  "@ai-sdk/openai-compatible",
			"name": "Synthetic",
			"options": map[string]interface{}{
				"baseURL": "https://api.synthetic.new/openai/v1",
			},
			"models": make(map[string]interface{}),
		}
		provider["synthetic"] = synthetic
	}

	// Create a new models map with only the selected models (replaces all existing)
	newModelsMap := make(map[string]interface{})
	for modelID, modelCfg := range models {
		newModelsMap[modelID] = modelCfg
	}

	// Replace the entire models section
	synthetic["models"] = newModelsMap

	// Write the updated config
	return m.Write(cfg)
}

// GetConfigPath returns the current config path.
func (m *Manager) GetConfigPath() string {
	return expandPath(m.configPath)
}

// ConfigExists checks if the config file exists.
func (m *Manager) ConfigExists() bool {
	configPath := expandPath(m.configPath)
	_, err := os.Stat(configPath)
	return err == nil
}

// CheckConfigStatus checks if config exists and has synthetic provider.
func (m *Manager) CheckConfigStatus() ConfigStatus {
	if !m.ConfigExists() {
		return ConfigNotFound
	}

	cfg, err := m.Read()
	if err != nil {
		return ConfigNotFound
	}

	provider, ok := cfg.data["provider"].(map[string]interface{})
	if !ok || provider == nil {
		return ConfigExistsNoProvider
	}

	synthetic, ok := provider["synthetic"].(map[string]interface{})
	if !ok || synthetic == nil {
		return ConfigExistsNoProvider
	}

	// Check if synthetic has required fields (npm indicates proper structure)
	if _, hasNPM := synthetic["npm"]; !hasNPM {
		return ConfigExistsNoProvider // Wrong structure
	}

	return ConfigExistsWithProvider
}

// CreateSyntheticProvider creates a new synthetic provider with the given API key.
func (m *Manager) CreateSyntheticProvider(apiKey string) error {
	cfg, err := m.Read()
	if err != nil {
		return err
	}

	// Ensure provider map exists
	provider, ok := cfg.data["provider"].(map[string]interface{})
	if !ok || provider == nil {
		provider = make(map[string]interface{})
		cfg.data["provider"] = provider
	}

	// Get base URL from env or use default
	baseURL := DefaultBaseURL
	if envURL := os.Getenv(EnvAPIBaseURL); envURL != "" {
		baseURL = envURL
	}

	// Create synthetic provider
	synthetic := map[string]interface{}{
		"npm":  "@ai-sdk/openai-compatible",
		"name": "Synthetic",
		"options": map[string]interface{}{
			"apiKey":  apiKey,
			"baseURL": baseURL,
		},
		"models": make(map[string]interface{}),
	}

	provider["synthetic"] = synthetic

	return m.Write(cfg)
}

// GetData returns the underlying data map (for testing purposes).
func (c *Config) GetData() map[string]interface{} {
	return c.data
}

// SetData sets the underlying data map (for testing purposes).
func (c *Config) SetData(data map[string]interface{}) {
	c.data = data
}

// GetProviders returns the provider map for backward compatibility with tests.
func (c *Config) GetProviders() map[string]interface{} {
	if provider, ok := c.data["provider"].(map[string]interface{}); ok && provider != nil {
		return provider
	}
	return make(map[string]interface{})
}
