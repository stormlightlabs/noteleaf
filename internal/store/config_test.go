package store

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("DefaultConfig should not return nil")
	}

	expectedDefaults := map[string]any{
		"DateFormat":   "2006-01-02",
		"ColorScheme":  "default",
		"DefaultView":  "list",
		"AutoArchive":  false,
		"SyncEnabled":  false,
		"ExportFormat": "json",
	}

	if config.DateFormat != expectedDefaults["DateFormat"] {
		t.Errorf("Expected DateFormat %s, got %s", expectedDefaults["DateFormat"], config.DateFormat)
	}
	if config.ColorScheme != expectedDefaults["ColorScheme"] {
		t.Errorf("Expected ColorScheme %s, got %s", expectedDefaults["ColorScheme"], config.ColorScheme)
	}
	if config.DefaultView != expectedDefaults["DefaultView"] {
		t.Errorf("Expected DefaultView %s, got %s", expectedDefaults["DefaultView"], config.DefaultView)
	}
	if config.AutoArchive != expectedDefaults["AutoArchive"] {
		t.Errorf("Expected AutoArchive %v, got %v", expectedDefaults["AutoArchive"], config.AutoArchive)
	}
	if config.SyncEnabled != expectedDefaults["SyncEnabled"] {
		t.Errorf("Expected SyncEnabled %v, got %v", expectedDefaults["SyncEnabled"], config.SyncEnabled)
	}
	if config.ExportFormat != expectedDefaults["ExportFormat"] {
		t.Errorf("Expected ExportFormat %s, got %s", expectedDefaults["ExportFormat"], config.ExportFormat)
	}
}

func TestConfigOperations(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "noteleaf-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalGetConfigDir := GetConfigDir
	GetConfigDir = func() (string, error) {
		return tempDir, nil
	}
	defer func() { GetConfigDir = originalGetConfigDir }()

	t.Run("SaveConfig creates config file", func(t *testing.T) {
		config := DefaultConfig()
		config.ColorScheme = "dark"
		config.AutoArchive = true

		err := SaveConfig(config)
		if err != nil {
			t.Fatalf("SaveConfig failed: %v", err)
		}

		configPath := filepath.Join(tempDir, ".noteleaf.conf.toml")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("Config file should exist after SaveConfig")
		}
	})

	t.Run("LoadConfig reads existing config", func(t *testing.T) {
		config, err := LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig failed: %v", err)
		}

		if config.ColorScheme != "dark" {
			t.Errorf("Expected ColorScheme 'dark', got '%s'", config.ColorScheme)
		}
		if !config.AutoArchive {
			t.Error("Expected AutoArchive to be true")
		}
	})

	t.Run("LoadConfig creates default when file doesn't exist", func(t *testing.T) {
		configPath := filepath.Join(tempDir, ".noteleaf.conf.toml")
		os.Remove(configPath)

		config, err := LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig failed: %v", err)
		}

		if config.ColorScheme != "default" {
			t.Errorf("Expected default ColorScheme 'default', got '%s'", config.ColorScheme)
		}
		if config.AutoArchive {
			t.Error("Expected AutoArchive to be false by default")
		}

		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("Config file should be created when it doesn't exist")
		}
	})

	t.Run("GetConfigPath returns correct path", func(t *testing.T) {
		configPath, err := GetConfigPath()
		if err != nil {
			t.Fatalf("GetConfigPath failed: %v", err)
		}

		expectedPath := filepath.Join(tempDir, ".noteleaf.conf.toml")
		if configPath != expectedPath {
			t.Errorf("Expected config path %s, got %s", expectedPath, configPath)
		}
	})
}

func TestConfigPersistence(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "noteleaf-config-persist-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalGetConfigDir := GetConfigDir
	GetConfigDir = func() (string, error) {
		return tempDir, nil
	}
	defer func() { GetConfigDir = originalGetConfigDir }()

	t.Run("config values persist across save/load cycles", func(t *testing.T) {
		originalConfig := &Config{
			DateFormat:      "01/02/2006",
			ColorScheme:     "custom",
			DefaultView:     "kanban",
			DefaultPriority: "high",
			AutoArchive:     true,
			SyncEnabled:     true,
			SyncEndpoint:    "https://api.example.com",
			SyncToken:       "secret-token",
			ExportFormat:    "csv",
			MovieAPIKey:     "movie-key",
			BookAPIKey:      "book-key",
		}

		err := SaveConfig(originalConfig)
		if err != nil {
			t.Fatalf("SaveConfig failed: %v", err)
		}

		loadedConfig, err := LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig failed: %v", err)
		}

		if loadedConfig.DateFormat != originalConfig.DateFormat {
			t.Errorf("DateFormat not preserved: expected %s, got %s", originalConfig.DateFormat, loadedConfig.DateFormat)
		}
		if loadedConfig.ColorScheme != originalConfig.ColorScheme {
			t.Errorf("ColorScheme not preserved: expected %s, got %s", originalConfig.ColorScheme, loadedConfig.ColorScheme)
		}
		if loadedConfig.DefaultView != originalConfig.DefaultView {
			t.Errorf("DefaultView not preserved: expected %s, got %s", originalConfig.DefaultView, loadedConfig.DefaultView)
		}
		if loadedConfig.DefaultPriority != originalConfig.DefaultPriority {
			t.Errorf("DefaultPriority not preserved: expected %s, got %s", originalConfig.DefaultPriority, loadedConfig.DefaultPriority)
		}
		if loadedConfig.AutoArchive != originalConfig.AutoArchive {
			t.Errorf("AutoArchive not preserved: expected %v, got %v", originalConfig.AutoArchive, loadedConfig.AutoArchive)
		}
		if loadedConfig.SyncEnabled != originalConfig.SyncEnabled {
			t.Errorf("SyncEnabled not preserved: expected %v, got %v", originalConfig.SyncEnabled, loadedConfig.SyncEnabled)
		}
		if loadedConfig.SyncEndpoint != originalConfig.SyncEndpoint {
			t.Errorf("SyncEndpoint not preserved: expected %s, got %s", originalConfig.SyncEndpoint, loadedConfig.SyncEndpoint)
		}
		if loadedConfig.SyncToken != originalConfig.SyncToken {
			t.Errorf("SyncToken not preserved: expected %s, got %s", originalConfig.SyncToken, loadedConfig.SyncToken)
		}
		if loadedConfig.ExportFormat != originalConfig.ExportFormat {
			t.Errorf("ExportFormat not preserved: expected %s, got %s", originalConfig.ExportFormat, loadedConfig.ExportFormat)
		}
		if loadedConfig.MovieAPIKey != originalConfig.MovieAPIKey {
			t.Errorf("MovieAPIKey not preserved: expected %s, got %s", originalConfig.MovieAPIKey, loadedConfig.MovieAPIKey)
		}
		if loadedConfig.BookAPIKey != originalConfig.BookAPIKey {
			t.Errorf("BookAPIKey not preserved: expected %s, got %s", originalConfig.BookAPIKey, loadedConfig.BookAPIKey)
		}
	})
}

func TestConfigErrorHandling(t *testing.T) {
	t.Run("LoadConfig handles invalid TOML", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "noteleaf-config-error-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		originalGetConfigDir := GetConfigDir
		GetConfigDir = func() (string, error) {
			return tempDir, nil
		}
		defer func() { GetConfigDir = originalGetConfigDir }()

		configPath := filepath.Join(tempDir, ".noteleaf.conf.toml")
		invalidTOML := `[invalid toml content`
		err = os.WriteFile(configPath, []byte(invalidTOML), 0644)
		if err != nil {
			t.Fatalf("Failed to write invalid TOML: %v", err)
		}

		_, err = LoadConfig()
		if err == nil {
			t.Error("LoadConfig should fail with invalid TOML")
		}
	})

	t.Run("LoadConfig handles file read permission error", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Permission test not reliable on Windows")
		}

		tempDir, err := os.MkdirTemp("", "noteleaf-config-perm-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		originalGetConfigDir := GetConfigDir
		GetConfigDir = func() (string, error) {
			return tempDir, nil
		}
		defer func() { GetConfigDir = originalGetConfigDir }()

		configPath := filepath.Join(tempDir, ".noteleaf.conf.toml")
		validTOML := `color_scheme = "dark"`
		err = os.WriteFile(configPath, []byte(validTOML), 0644)
		if err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		err = os.Chmod(configPath, 0000)
		if err != nil {
			t.Fatalf("Failed to change file permissions: %v", err)
		}
		defer os.Chmod(configPath, 0644)

		_, err = LoadConfig()
		if err == nil {
			t.Error("LoadConfig should fail when config file is not readable")
		}
	})

	t.Run("LoadConfig handles GetConfigDir error", func(t *testing.T) {
		originalGetConfigDir := GetConfigDir
		GetConfigDir = func() (string, error) {
			return "", os.ErrPermission
		}
		defer func() { GetConfigDir = originalGetConfigDir }()

		_, err := LoadConfig()
		if err == nil {
			t.Error("LoadConfig should fail when GetConfigDir fails")
		}
	})

	t.Run("LoadConfig handles SaveConfig failure when creating default", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "noteleaf-config-save-fail-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		_ = filepath.Join(tempDir, ".noteleaf.conf.toml")

		callCount := 0
		originalGetConfigDir := GetConfigDir
		GetConfigDir = func() (string, error) {
			callCount++
			if callCount == 1 {
				return tempDir, nil
			}
			return "", os.ErrPermission
		}
		defer func() { GetConfigDir = originalGetConfigDir }()

		_, err = LoadConfig()
		if err == nil {
			t.Error("LoadConfig should fail when SaveConfig fails during default config creation")
		}
	})

	t.Run("SaveConfig handles directory creation failure", func(t *testing.T) {
		originalGetConfigDir := GetConfigDir
		GetConfigDir = func() (string, error) {
			return "/invalid/path/that/cannot/be/created", nil
		}
		defer func() { GetConfigDir = originalGetConfigDir }()

		config := DefaultConfig()
		err := SaveConfig(config)
		if err == nil {
			t.Error("SaveConfig should fail when config directory cannot be accessed")
		}
	})

	t.Run("SaveConfig handles file write permission error", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Permission test not reliable on Windows")
		}

		tempDir, err := os.MkdirTemp("", "noteleaf-config-write-perm-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		originalGetConfigDir := GetConfigDir
		GetConfigDir = func() (string, error) {
			return tempDir, nil
		}
		defer func() { GetConfigDir = originalGetConfigDir }()

		err = os.Chmod(tempDir, 0555)
		if err != nil {
			t.Fatalf("Failed to change directory permissions: %v", err)
		}
		defer os.Chmod(tempDir, 0755)

		config := DefaultConfig()
		err = SaveConfig(config)
		if err == nil {
			t.Error("SaveConfig should fail when directory is not writable")
		}
	})

	t.Run("GetConfigPath handles GetConfigDir error", func(t *testing.T) {
		originalGetConfigDir := GetConfigDir
		GetConfigDir = func() (string, error) {
			return "", os.ErrPermission
		}
		defer func() { GetConfigDir = originalGetConfigDir }()

		_, err := GetConfigPath()
		if err == nil {
			t.Error("GetConfigPath should fail when GetConfigDir fails")
		}
	})
}

func TestGetConfigDir(t *testing.T) {
	t.Run("returns correct directory based on OS", func(t *testing.T) {
		configDir, err := GetConfigDir()
		if err != nil {
			t.Fatalf("GetConfigDir failed: %v", err)
		}

		if configDir == "" {
			t.Error("Config directory should not be empty")
		}

		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			t.Error("Config directory should be created if it doesn't exist")
		}

		if filepath.Base(configDir) != "noteleaf" {
			t.Errorf("Config directory should end with 'noteleaf', got: %s", configDir)
		}
	})

	t.Run("creates directory if it doesn't exist", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "noteleaf-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		var originalEnv string
		var envVar string
		var expectedPath string
		switch runtime.GOOS {
		case "windows":
			envVar = "APPDATA"
			originalEnv = os.Getenv("APPDATA")
			os.Setenv("APPDATA", tempDir)
			expectedPath = filepath.Join(tempDir, "noteleaf")
		case "darwin":
			envVar = "HOME"
			originalEnv = os.Getenv("HOME")
			os.Setenv("HOME", tempDir)
			expectedPath = filepath.Join(tempDir, "Library", "Application Support", "noteleaf")
		default:
			envVar = "XDG_CONFIG_HOME"
			originalEnv = os.Getenv("XDG_CONFIG_HOME")
			os.Setenv("XDG_CONFIG_HOME", tempDir)
			expectedPath = filepath.Join(tempDir, "noteleaf")
		}
		defer os.Setenv(envVar, originalEnv)

		configDir, err := GetConfigDir()
		if err != nil {
			t.Fatalf("GetConfigDir failed: %v", err)
		}

		if configDir != expectedPath {
			t.Errorf("Expected config dir %s, got %s", expectedPath, configDir)
		}

		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			t.Error("Config directory should be created")
		}
	})

	t.Run("handles missing environment variables", func(t *testing.T) {
		switch runtime.GOOS {
		case "windows":
			originalAppData := os.Getenv("APPDATA")
			os.Unsetenv("APPDATA")
			defer os.Setenv("APPDATA", originalAppData)

			_, err := GetConfigDir()
			if err == nil {
				t.Error("GetConfigDir should fail when APPDATA is not set on Windows")
			}
		case "darwin":
			originalHome := os.Getenv("HOME")

			tempHome, err := os.MkdirTemp("", "noteleaf-home-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp home: %v", err)
			}
			defer os.RemoveAll(tempHome)
			os.Setenv("HOME", tempHome)
			defer os.Setenv("HOME", originalHome)

			configDir, err := GetConfigDir()
			if err != nil {
				t.Fatalf("GetConfigDir should work with HOME on macOS: %v", err)
			}

			expectedPath := filepath.Join(tempHome, "Library", "Application Support", "noteleaf")
			if configDir != expectedPath {
				t.Errorf("Expected config dir %s, got %s", expectedPath, configDir)
			}
		default:
			originalXDG := os.Getenv("XDG_CONFIG_HOME")
			originalHome := os.Getenv("HOME")
			os.Unsetenv("XDG_CONFIG_HOME")

			tempHome, err := os.MkdirTemp("", "noteleaf-home-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp home: %v", err)
			}
			defer os.RemoveAll(tempHome)
			os.Setenv("HOME", tempHome)

			defer func() {
				os.Setenv("XDG_CONFIG_HOME", originalXDG)
				os.Setenv("HOME", originalHome)
			}()

			configDir, err := GetConfigDir()
			if err != nil {
				t.Fatalf("GetConfigDir should work with HOME fallback: %v", err)
			}

			expectedPath := filepath.Join(tempHome, ".config", "noteleaf")
			if configDir != expectedPath {
				t.Errorf("Expected config dir %s, got %s", expectedPath, configDir)
			}
		}
	})

	t.Run("handles HOME directory lookup failure on Unix", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("HOME directory test not applicable on Windows")
		}

		originalXDG := os.Getenv("XDG_CONFIG_HOME")
		originalHome := os.Getenv("HOME")
		os.Unsetenv("XDG_CONFIG_HOME")
		os.Unsetenv("HOME")

		defer func() {
			os.Setenv("XDG_CONFIG_HOME", originalXDG)
			os.Setenv("HOME", originalHome)
		}()

		_, err := GetConfigDir()
		if err == nil {
			t.Error("GetConfigDir should fail when both XDG_CONFIG_HOME and HOME are not available")
		}
	})

	t.Run("handles directory creation permission failure", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Permission test not reliable on Windows")
		}

		if runtime.GOOS == "darwin" {
			t.Skip("Permission test not reliable on macOS with nested Library/Application Support paths")
		}

		tempParent, err := os.MkdirTemp("", "noteleaf-parent-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp parent directory: %v", err)
		}
		defer os.RemoveAll(tempParent)

		err = os.Chmod(tempParent, 0555)
		if err != nil {
			t.Fatalf("Failed to change parent directory permissions: %v", err)
		}
		defer os.Chmod(tempParent, 0755)

		var originalEnv string
		envVar := "XDG_CONFIG_HOME"
		originalEnv = os.Getenv("XDG_CONFIG_HOME")
		os.Setenv("XDG_CONFIG_HOME", tempParent)
		defer os.Setenv(envVar, originalEnv)

		_, err = GetConfigDir()
		if err == nil {
			t.Error("GetConfigDir should fail when directory creation is not permitted")
		}
	})
}

func TestEnvironmentVariableOverrides(t *testing.T) {
	t.Run("NOTELEAF_CONFIG overrides default config path for LoadConfig", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "noteleaf-env-config-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		customConfigPath := filepath.Join(tempDir, "custom-config.toml")
		originalEnv := os.Getenv("NOTELEAF_CONFIG")
		os.Setenv("NOTELEAF_CONFIG", customConfigPath)
		defer os.Setenv("NOTELEAF_CONFIG", originalEnv)

		// Create a custom config
		customConfig := DefaultConfig()
		customConfig.ColorScheme = "custom-env-test"
		if err := SaveConfig(customConfig); err != nil {
			t.Fatalf("Failed to save custom config: %v", err)
		}

		// Load config should use the custom path
		loadedConfig, err := LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig failed: %v", err)
		}

		if loadedConfig.ColorScheme != "custom-env-test" {
			t.Errorf("Expected ColorScheme 'custom-env-test', got '%s'", loadedConfig.ColorScheme)
		}
	})

	t.Run("NOTELEAF_CONFIG overrides default config path for SaveConfig", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "noteleaf-env-save-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		customConfigPath := filepath.Join(tempDir, "subdir", "config.toml")
		originalEnv := os.Getenv("NOTELEAF_CONFIG")
		os.Setenv("NOTELEAF_CONFIG", customConfigPath)
		defer os.Setenv("NOTELEAF_CONFIG", originalEnv)

		config := DefaultConfig()
		config.DefaultView = "kanban-env"
		if err := SaveConfig(config); err != nil {
			t.Fatalf("SaveConfig failed: %v", err)
		}

		// Verify the file was created at the custom path
		if _, err := os.Stat(customConfigPath); os.IsNotExist(err) {
			t.Error("Config file should be created at custom NOTELEAF_CONFIG path")
		}

		// Verify the content
		data, err := os.ReadFile(customConfigPath)
		if err != nil {
			t.Fatalf("Failed to read config file: %v", err)
		}

		loadedConfig := DefaultConfig()
		if err := toml.Unmarshal(data, loadedConfig); err != nil {
			t.Fatalf("Failed to parse config: %v", err)
		}

		if loadedConfig.DefaultView != "kanban-env" {
			t.Errorf("Expected DefaultView 'kanban-env', got '%s'", loadedConfig.DefaultView)
		}
	})

	t.Run("NOTELEAF_CONFIG overrides default config path for GetConfigPath", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "noteleaf-env-path-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		customConfigPath := filepath.Join(tempDir, "my-config.toml")
		originalEnv := os.Getenv("NOTELEAF_CONFIG")
		os.Setenv("NOTELEAF_CONFIG", customConfigPath)
		defer os.Setenv("NOTELEAF_CONFIG", originalEnv)

		path, err := GetConfigPath()
		if err != nil {
			t.Fatalf("GetConfigPath failed: %v", err)
		}

		if path != customConfigPath {
			t.Errorf("Expected config path '%s', got '%s'", customConfigPath, path)
		}
	})

	t.Run("NOTELEAF_CONFIG creates parent directories if needed", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "noteleaf-env-mkdir-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		customConfigPath := filepath.Join(tempDir, "nested", "deep", "config.toml")
		originalEnv := os.Getenv("NOTELEAF_CONFIG")
		os.Setenv("NOTELEAF_CONFIG", customConfigPath)
		defer os.Setenv("NOTELEAF_CONFIG", originalEnv)

		config := DefaultConfig()
		if err := SaveConfig(config); err != nil {
			t.Fatalf("SaveConfig should create parent directories: %v", err)
		}

		if _, err := os.Stat(customConfigPath); os.IsNotExist(err) {
			t.Error("Config file should be created with parent directories")
		}
	})
}

func TestGetDataDir(t *testing.T) {
	t.Run("NOTELEAF_DATA_DIR overrides default data directory", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "noteleaf-data-dir-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		customDataDir := filepath.Join(tempDir, "my-data")
		originalEnv := os.Getenv("NOTELEAF_DATA_DIR")
		os.Setenv("NOTELEAF_DATA_DIR", customDataDir)
		defer os.Setenv("NOTELEAF_DATA_DIR", originalEnv)

		dataDir, err := GetDataDir()
		if err != nil {
			t.Fatalf("GetDataDir failed: %v", err)
		}

		if dataDir != customDataDir {
			t.Errorf("Expected data dir '%s', got '%s'", customDataDir, dataDir)
		}

		// Verify directory was created
		if _, err := os.Stat(customDataDir); os.IsNotExist(err) {
			t.Error("Data directory should be created")
		}
	})

	t.Run("GetDataDir returns correct directory based on OS", func(t *testing.T) {
		// Temporarily unset NOTELEAF_DATA_DIR
		originalEnv := os.Getenv("NOTELEAF_DATA_DIR")
		os.Unsetenv("NOTELEAF_DATA_DIR")
		defer os.Setenv("NOTELEAF_DATA_DIR", originalEnv)

		dataDir, err := GetDataDir()
		if err != nil {
			t.Fatalf("GetDataDir failed: %v", err)
		}

		if dataDir == "" {
			t.Error("Data directory should not be empty")
		}

		if filepath.Base(dataDir) != "noteleaf" {
			t.Errorf("Data directory should end with 'noteleaf', got: %s", dataDir)
		}
	})

	t.Run("GetDataDir handles NOTELEAF_DATA_DIR with nested path", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "noteleaf-nested-data-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		customDataDir := filepath.Join(tempDir, "level1", "level2", "data")
		originalEnv := os.Getenv("NOTELEAF_DATA_DIR")
		os.Setenv("NOTELEAF_DATA_DIR", customDataDir)
		defer os.Setenv("NOTELEAF_DATA_DIR", originalEnv)

		dataDir, err := GetDataDir()
		if err != nil {
			t.Fatalf("GetDataDir should create nested directories: %v", err)
		}

		if dataDir != customDataDir {
			t.Errorf("Expected data dir '%s', got '%s'", customDataDir, dataDir)
		}

		// Verify nested directories were created
		if _, err := os.Stat(customDataDir); os.IsNotExist(err) {
			t.Error("Nested data directories should be created")
		}
	})

	t.Run("GetDataDir uses platform-specific defaults", func(t *testing.T) {
		// Temporarily unset NOTELEAF_DATA_DIR
		originalEnv := os.Getenv("NOTELEAF_DATA_DIR")
		os.Unsetenv("NOTELEAF_DATA_DIR")
		defer os.Setenv("NOTELEAF_DATA_DIR", originalEnv)

		// Create temporary environment for testing
		tempHome, err := os.MkdirTemp("", "noteleaf-home-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp home: %v", err)
		}
		defer os.RemoveAll(tempHome)

		var envVar, originalValue string
		switch runtime.GOOS {
		case "windows":
			envVar = "LOCALAPPDATA"
			originalValue = os.Getenv("LOCALAPPDATA")
			os.Setenv("LOCALAPPDATA", tempHome)
		case "darwin":
			envVar = "HOME"
			originalValue = os.Getenv("HOME")
			os.Setenv("HOME", tempHome)
		default:
			envVar = "XDG_DATA_HOME"
			originalValue = os.Getenv("XDG_DATA_HOME")
			os.Setenv("XDG_DATA_HOME", tempHome)
		}
		defer os.Setenv(envVar, originalValue)

		dataDir, err := GetDataDir()
		if err != nil {
			t.Fatalf("GetDataDir failed: %v", err)
		}

		// Verify the path contains our temp directory
		if !strings.Contains(dataDir, tempHome) {
			t.Errorf("Data directory should be under temp home, got: %s", dataDir)
		}
	})
}
