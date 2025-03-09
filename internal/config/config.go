package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration
type Config struct {
	AnthropicAPIKey   string   `yaml:"anthropic_api_key"`
	LLMModel          string   `yaml:"llm_model"`
	PreferredCommands []string `yaml:"preferred_commands"`
	ExtraInstructions []string `yaml:"extra_instructions"`
}

// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
	return &Config{
		AnthropicAPIKey:   "",
		LLMModel:          "claude-3-haiku-20240307",
		PreferredCommands: []string{"rg", "fd", "find", "grep", "awk", "sed"},
		ExtraInstructions: []string{
			"Prefer using modern alternatives like ripgrep (rg) instead of grep when available",
			"For Python projects, recommend using uv for package management",
		},
	}
}

// GetConfigPath returns the path to the config file
func GetConfigPath() (string, error) {
	// Try XDG_CONFIG_HOME first
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		// Fall back to HOME/.config
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("could not determine home directory: %w", err)
		}
		configDir = filepath.Join(home, ".config")
	}

	// Ensure the directory exists
	tellConfigDir := filepath.Join(configDir, "tell-llm")
	if err := os.MkdirAll(tellConfigDir, 0755); err != nil {
		return "", fmt.Errorf("could not create config directory: %w", err)
	}

	return filepath.Join(tellConfigDir, "tell.yaml"), nil
}

// Load loads the configuration from disk
func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return default config if file doesn't exist
		return DefaultConfig(), nil
	}

	// Read the file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %w", err)
	}

	// Parse YAML
	config := DefaultConfig()
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("could not parse config file: %w", err)
	}

	return config, nil
}

// Save saves the configuration to disk
func (c *Config) Save() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Marshal to YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("could not marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("could not write config file: %w", err)
	}

	return nil
}

// CreateDefaultConfig creates a default configuration file
func CreateDefaultConfig() error {
	config := DefaultConfig()
	return config.Save()
}
