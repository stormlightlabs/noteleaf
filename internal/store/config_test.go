package store

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
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
		switch runtime.GOOS {
		case "windows":
			envVar = "APPDATA"
			originalEnv = os.Getenv("APPDATA")
			os.Setenv("APPDATA", tempDir)
		default:
			envVar = "XDG_CONFIG_HOME"
			originalEnv = os.Getenv("XDG_CONFIG_HOME")
			os.Setenv("XDG_CONFIG_HOME", tempDir)
		}
		defer os.Setenv(envVar, originalEnv)

		configDir, err := GetConfigDir()
		if err != nil {
			t.Fatalf("GetConfigDir failed: %v", err)
		}

		expectedPath := filepath.Join(tempDir, "noteleaf")
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
		var envVar string
		switch runtime.GOOS {
		case "windows":
			envVar = "APPDATA"
			originalEnv = os.Getenv("APPDATA")
			os.Setenv("APPDATA", tempParent)
		default:
			envVar = "XDG_CONFIG_HOME"
			originalEnv = os.Getenv("XDG_CONFIG_HOME")
			os.Setenv("XDG_CONFIG_HOME", tempParent)
		}
		defer os.Setenv(envVar, originalEnv)

		_, err = GetConfigDir()
		if err == nil {
			t.Error("GetConfigDir should fail when directory creation is not permitted")
		}
	})
}
