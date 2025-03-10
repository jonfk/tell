package config

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

func EditConfig() {
	slog.Info("Opening config file in editor")

	configPath, err := GetConfigPath()
	if err != nil {
		slog.Error("Failed to get config path", "error", err)
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Create the config file if it doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		slog.Info("Config file doesn't exist, creating default")
		if err := CreateDefaultConfig(); err != nil {
			slog.Error("Failed to create default config", "error", err)
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}

	// Get the editor from environment variables
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = "vi" // Default to vi if no editor is specified
	}

	slog.Info("Opening config with editor", "editor", editor, "path", configPath)

	// Create command to open the editor
	editorCmd := exec.Command(editor, configPath)
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr

	if err := editorCmd.Run(); err != nil {
		slog.Error("Failed to open editor", "error", err)
		fmt.Fprintf(os.Stderr, "Error opening editor: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Configuration saved at %s\n", configPath)
}

func InitConfig() {
	slog.Info("Initializing default configuration")

	if err := CreateDefaultConfig(); err != nil {
		slog.Error("Failed to create default configuration", "error", err)
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	configPath, _ := GetConfigPath()
	fmt.Printf("Created default configuration at %s\n", configPath)
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
		slog.Info("Config file not found, using defaults", "path", configPath)
		return DefaultConfig(), nil
	}

	// Read the file
	data, err := os.ReadFile(configPath)
	if err != nil {
		slog.Error("Failed to read config file", "path", configPath, "error", err)
		return nil, fmt.Errorf("could not read config file: %w", err)
	}

	// Parse YAML
	config := DefaultConfig()
	if err := yaml.Unmarshal(data, config); err != nil {
		slog.Error("Failed to parse config file", "path", configPath, "error", err)
		return nil, fmt.Errorf("could not parse config file: %w", err)
	}

	slog.Debug("Loaded configuration",
		"path", configPath,
		"model", config.LLMModel,
		"preferredCommandsCount", len(config.PreferredCommands))

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
		slog.Error("Failed to marshal config to YAML", "error", err)
		return fmt.Errorf("could not marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		slog.Error("Failed to write config file", "path", configPath, "error", err)
		return fmt.Errorf("could not write config file: %w", err)
	}

	slog.Info("Saved configuration", "path", configPath)
	return nil
}

// String returns a string representation of the config with sensitive information truncated
func (c *Config) String() string {
	var sb strings.Builder

	sb.WriteString("Configuration:\n")

	// Truncate API key for security
	apiKey := c.AnthropicAPIKey
	if apiKey != "" {
		// Show only first 4 and last 4 characters
		if len(apiKey) > 8 {
			apiKey = apiKey[:4] + "..." + apiKey[len(apiKey)-4:]
		} else {
			apiKey = "****"
		}
	} else {
		apiKey = "<not set>"
	}

	sb.WriteString(fmt.Sprintf("  Anthropic API Key: %s\n", apiKey))
	sb.WriteString(fmt.Sprintf("  LLM Model: %s\n", c.LLMModel))

	sb.WriteString("  Preferred Commands:\n")
	for _, cmd := range c.PreferredCommands {
		sb.WriteString(fmt.Sprintf("    - %s\n", cmd))
	}

	sb.WriteString("  Extra Instructions:\n")
	for _, instr := range c.ExtraInstructions {
		sb.WriteString(fmt.Sprintf("    - %s\n", instr))
	}

	return sb.String()
}

// CreateDefaultConfig creates a default configuration file
func CreateDefaultConfig() error {
	config := DefaultConfig()
	slog.Info("Creating default configuration")
	return config.Save()
}
