package store

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config holds application configuration
type Config struct {
	DatabasePath    string `toml:"database_path,omitempty"`
	DataDir         string `toml:"data_dir,omitempty"`
	DateFormat      string `toml:"date_format"`
	ColorScheme     string `toml:"color_scheme"`
	DefaultView     string `toml:"default_view"`
	DefaultPriority string `toml:"default_priority,omitempty"`
	Editor          string `toml:"editor,omitempty"`
	ArticlesDir     string `toml:"articles_dir,omitempty"`
	NotesDir        string `toml:"notes_dir,omitempty"`
	AutoArchive     bool   `toml:"auto_archive"`
	SyncEnabled     bool   `toml:"sync_enabled"`
	SyncEndpoint    string `toml:"sync_endpoint,omitempty"`
	SyncToken       string `toml:"sync_token,omitempty"`
	ExportFormat    string `toml:"export_format"`
	MovieAPIKey     string `toml:"movie_api_key,omitempty"`
	BookAPIKey      string `toml:"book_api_key,omitempty"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		DateFormat:   "2006-01-02",
		ColorScheme:  "default",
		DefaultView:  "list",
		AutoArchive:  false,
		SyncEnabled:  false,
		ExportFormat: "json",
	}
}

// LoadConfig loads configuration from the config directory or NOTELEAF_CONFIG path
func LoadConfig() (*Config, error) {
	var configPath string

	// Check for NOTELEAF_CONFIG environment variable
	if envConfigPath := os.Getenv("NOTELEAF_CONFIG"); envConfigPath != "" {
		configPath = envConfigPath
	} else {
		configDir, err := GetConfigDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get config directory: %w", err)
		}
		configPath = filepath.Join(configDir, ".noteleaf.conf.toml")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		config := DefaultConfig()
		if err := SaveConfig(config); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		return config, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := DefaultConfig()
	if err := toml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// SaveConfig saves the configuration to the config directory or NOTELEAF_CONFIG path
func SaveConfig(config *Config) error {
	var configPath string

	// Check for NOTELEAF_CONFIG environment variable
	if envConfigPath := os.Getenv("NOTELEAF_CONFIG"); envConfigPath != "" {
		configPath = envConfigPath
		// Ensure the directory exists for custom config path
		configDir := filepath.Dir(configPath)
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
	} else {
		configDir, err := GetConfigDir()
		if err != nil {
			return fmt.Errorf("failed to get config directory: %w", err)
		}
		configPath = filepath.Join(configDir, ".noteleaf.conf.toml")
	}

	data, err := toml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetConfigPath returns the path to the configuration file
func GetConfigPath() (string, error) {
	// Check for NOTELEAF_CONFIG environment variable
	if envConfigPath := os.Getenv("NOTELEAF_CONFIG"); envConfigPath != "" {
		return envConfigPath, nil
	}

	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, ".noteleaf.conf.toml"), nil
}
