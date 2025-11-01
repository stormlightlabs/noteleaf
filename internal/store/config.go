package store

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/stormlightlabs/noteleaf/internal/shared"
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

	ATProtoDID        string `toml:"atproto_did,omitempty"`
	ATProtoHandle     string `toml:"atproto_handle,omitempty"`
	ATProtoAccessJWT  string `toml:"atproto_access_jwt,omitempty"`
	ATProtoRefreshJWT string `toml:"atproto_refresh_jwt,omitempty"`
	ATProtoPDSURL     string `toml:"atproto_pds_url,omitempty"`
	ATProtoExpiresAt  string `toml:"atproto_expires_at,omitempty"` // ISO8601 timestamp
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
var LoadConfig = func() (*Config, error) {
	var configPath string

	if envConfigPath := os.Getenv("NOTELEAF_CONFIG"); envConfigPath != "" {
		configPath = envConfigPath
	} else {
		configDir, err := GetConfigDir()
		if err != nil {
			return nil, shared.ConfigError("failed to get config directory", err)
		}
		configPath = filepath.Join(configDir, ".noteleaf.conf.toml")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		config := DefaultConfig()
		if err := SaveConfig(config); err != nil {
			return nil, shared.ConfigError("failed to create default config", err)
		}
		return config, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, shared.ConfigError("failed to read config file", err)
	}

	config := DefaultConfig()
	if err := toml.Unmarshal(data, config); err != nil {
		return nil, shared.ConfigError("failed to parse config file", err)
	}

	return config, nil
}

// SaveConfig saves the configuration to the config directory or NOTELEAF_CONFIG path
func SaveConfig(config *Config) error {
	var configPath string

	if envConfigPath := os.Getenv("NOTELEAF_CONFIG"); envConfigPath != "" {
		configPath = envConfigPath
		configDir := filepath.Dir(configPath)
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return shared.ConfigError("failed to create config directory", err)
		}
	} else {
		configDir, err := GetConfigDir()
		if err != nil {
			return shared.ConfigError("failed to get config directory", err)
		}
		configPath = filepath.Join(configDir, ".noteleaf.conf.toml")
	}

	data, err := toml.Marshal(config)
	if err != nil {
		return shared.ConfigError("failed to marshal config", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return shared.ConfigError("failed to write config file", err)
	}

	return nil
}

// GetConfigPath returns the path to the configuration file
func GetConfigPath() (string, error) {
	if envConfigPath := os.Getenv("NOTELEAF_CONFIG"); envConfigPath != "" {
		return envConfigPath, nil
	}

	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, ".noteleaf.conf.toml"), nil
}
