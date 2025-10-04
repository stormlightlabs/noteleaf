package handlers

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stormlightlabs/noteleaf/internal/store"
)

func TestConfigHandlerGet(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "noteleaf-config-handler-get-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set up environment
	customConfigPath := filepath.Join(tempDir, "test-config.toml")
	originalEnv := os.Getenv("NOTELEAF_CONFIG")
	os.Setenv("NOTELEAF_CONFIG", customConfigPath)
	defer os.Setenv("NOTELEAF_CONFIG", originalEnv)

	// Create a test config
	config := store.DefaultConfig()
	config.ColorScheme = "test-scheme"
	config.Editor = "vim"
	if err := store.SaveConfig(config); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	t.Run("Get all config values", func(t *testing.T) {
		handler, err := NewConfigHandler()
		if err != nil {
			t.Fatalf("Failed to create handler: %v", err)
		}

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err = handler.Get("")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		if !strings.Contains(output, "color_scheme") {
			t.Error("Output should contain color_scheme")
		}
		if !strings.Contains(output, "test-scheme") {
			t.Error("Output should contain test-scheme value")
		}
	})

	t.Run("Get specific config value", func(t *testing.T) {
		handler, err := NewConfigHandler()
		if err != nil {
			t.Fatalf("Failed to create handler: %v", err)
		}

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err = handler.Get("editor")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		if !strings.Contains(output, "editor = vim") {
			t.Errorf("Output should contain 'editor = vim', got: %s", output)
		}
	})

	t.Run("Get unknown config key", func(t *testing.T) {
		handler, err := NewConfigHandler()
		if err != nil {
			t.Fatalf("Failed to create handler: %v", err)
		}

		err = handler.Get("nonexistent_key")
		if err == nil {
			t.Error("Get should fail for unknown key")
		}
		if !strings.Contains(err.Error(), "unknown config key") {
			t.Errorf("Error should mention unknown config key, got: %v", err)
		}
	})
}

func TestConfigHandlerSet(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "noteleaf-config-handler-set-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set up environment
	customConfigPath := filepath.Join(tempDir, "test-config.toml")
	originalEnv := os.Getenv("NOTELEAF_CONFIG")
	os.Setenv("NOTELEAF_CONFIG", customConfigPath)
	defer os.Setenv("NOTELEAF_CONFIG", originalEnv)

	t.Run("Set string config value", func(t *testing.T) {
		handler, err := NewConfigHandler()
		if err != nil {
			t.Fatalf("Failed to create handler: %v", err)
		}

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err = handler.Set("editor", "emacs")
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		if !strings.Contains(output, "Set editor = emacs") {
			t.Errorf("Output should confirm setting, got: %s", output)
		}

		// Verify it was actually saved
		loadedConfig, err := store.LoadConfig()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		if loadedConfig.Editor != "emacs" {
			t.Errorf("Expected editor 'emacs', got '%s'", loadedConfig.Editor)
		}
	})

	t.Run("Set boolean config value", func(t *testing.T) {
		handler, err := NewConfigHandler()
		if err != nil {
			t.Fatalf("Failed to create handler: %v", err)
		}

		err = handler.Set("auto_archive", "true")
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		// Verify it was actually saved
		loadedConfig, err := store.LoadConfig()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		if !loadedConfig.AutoArchive {
			t.Error("Expected auto_archive to be true")
		}
	})

	t.Run("Set boolean config value with various formats", func(t *testing.T) {
		testCases := []struct {
			value    string
			expected bool
		}{
			{"true", true},
			{"1", true},
			{"yes", true},
			{"false", false},
			{"0", false},
			{"no", false},
		}

		for _, tc := range testCases {
			handler, err := NewConfigHandler()
			if err != nil {
				t.Fatalf("Failed to create handler: %v", err)
			}

			err = handler.Set("sync_enabled", tc.value)
			if err != nil {
				t.Fatalf("Set failed for value '%s': %v", tc.value, err)
			}

			loadedConfig, err := store.LoadConfig()
			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}

			if loadedConfig.SyncEnabled != tc.expected {
				t.Errorf("For value '%s', expected sync_enabled %v, got %v", tc.value, tc.expected, loadedConfig.SyncEnabled)
			}
		}
	})

	t.Run("Set unknown config key", func(t *testing.T) {
		handler, err := NewConfigHandler()
		if err != nil {
			t.Fatalf("Failed to create handler: %v", err)
		}

		err = handler.Set("nonexistent_key", "value")
		if err == nil {
			t.Error("Set should fail for unknown key")
		}
		if !strings.Contains(err.Error(), "unknown config key") {
			t.Errorf("Error should mention unknown config key, got: %v", err)
		}
	})
}

func TestConfigHandlerPath(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "noteleaf-config-handler-path-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	customConfigPath := filepath.Join(tempDir, "my-config.toml")
	originalEnv := os.Getenv("NOTELEAF_CONFIG")
	os.Setenv("NOTELEAF_CONFIG", customConfigPath)
	defer os.Setenv("NOTELEAF_CONFIG", originalEnv)

	t.Run("Path returns correct config file path", func(t *testing.T) {
		handler, err := NewConfigHandler()
		if err != nil {
			t.Fatalf("Failed to create handler: %v", err)
		}

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err = handler.Path()
		if err != nil {
			t.Fatalf("Path failed: %v", err)
		}

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := strings.TrimSpace(buf.String())

		if output != customConfigPath {
			t.Errorf("Expected path '%s', got '%s'", customConfigPath, output)
		}
	})
}

func TestConfigHandlerReset(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "noteleaf-config-handler-reset-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	customConfigPath := filepath.Join(tempDir, "test-config.toml")
	originalEnv := os.Getenv("NOTELEAF_CONFIG")
	os.Setenv("NOTELEAF_CONFIG", customConfigPath)
	defer os.Setenv("NOTELEAF_CONFIG", originalEnv)

	t.Run("Reset restores default config", func(t *testing.T) {
		// First, modify the config
		config := store.DefaultConfig()
		config.ColorScheme = "custom"
		config.AutoArchive = true
		config.Editor = "emacs"
		if err := store.SaveConfig(config); err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		handler, err := NewConfigHandler()
		if err != nil {
			t.Fatalf("Failed to create handler: %v", err)
		}

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err = handler.Reset()
		if err != nil {
			t.Fatalf("Reset failed: %v", err)
		}

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		if !strings.Contains(output, "reset to defaults") {
			t.Errorf("Output should confirm reset, got: %s", output)
		}

		// Verify config was reset
		loadedConfig, err := store.LoadConfig()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		defaultConfig := store.DefaultConfig()
		if loadedConfig.ColorScheme != defaultConfig.ColorScheme {
			t.Errorf("ColorScheme should be reset to default '%s', got '%s'", defaultConfig.ColorScheme, loadedConfig.ColorScheme)
		}
		if loadedConfig.AutoArchive != defaultConfig.AutoArchive {
			t.Errorf("AutoArchive should be reset to default %v, got %v", defaultConfig.AutoArchive, loadedConfig.AutoArchive)
		}
		if loadedConfig.Editor != defaultConfig.Editor {
			t.Errorf("Editor should be reset to default '%s', got '%s'", defaultConfig.Editor, loadedConfig.Editor)
		}
	})
}
