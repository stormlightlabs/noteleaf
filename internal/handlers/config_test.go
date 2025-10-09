package handlers

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stormlightlabs/noteleaf/internal/shared"
	"github.com/stormlightlabs/noteleaf/internal/store"
)

func TestConfigHandler(t *testing.T) {
	t.Run("Get", func(t *testing.T) {
		tempDir, cleanup := shared.CreateTempDir("noteleaf-config-handler-get-test-*", t)
		defer cleanup()

		customConfigPath := filepath.Join(tempDir, "test-config.toml")
		originalEnv := os.Getenv("NOTELEAF_CONFIG")
		os.Setenv("NOTELEAF_CONFIG", customConfigPath)
		defer os.Setenv("NOTELEAF_CONFIG", originalEnv)

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
	})

	t.Run("Set", func(t *testing.T) {
		tempDir, cleanup := shared.CreateTempDir("noteleaf-config-handler-set-test-*", t)
		defer cleanup()

		customConfigPath := filepath.Join(tempDir, "test-config.toml")
		originalEnv := os.Getenv("NOTELEAF_CONFIG")
		os.Setenv("NOTELEAF_CONFIG", customConfigPath)
		defer os.Setenv("NOTELEAF_CONFIG", originalEnv)

		t.Run("Set string config value", func(t *testing.T) {
			handler, err := NewConfigHandler()
			if err != nil {
				t.Fatalf("Failed to create handler: %v", err)
			}

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

			loadedConfig, err := store.LoadConfig()
			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}

			if !loadedConfig.AutoArchive {
				t.Error("Expected auto_archive to be true")
			}
		})

		t.Run("Set boolean config value with various formats", func(t *testing.T) {
			tc := []struct {
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

			for _, tt := range tc {
				handler, err := NewConfigHandler()
				if err != nil {
					t.Fatalf("Failed to create handler: %v", err)
				}

				err = handler.Set("sync_enabled", tt.value)
				if err != nil {
					t.Fatalf("Set failed for value '%s': %v", tt.value, err)
				}

				loadedConfig, err := store.LoadConfig()
				if err != nil {
					t.Fatalf("Failed to load config: %v", err)
				}

				if loadedConfig.SyncEnabled != tt.expected {
					t.Errorf("For value '%s', expected sync_enabled %v, got %v", tt.value, tt.expected, loadedConfig.SyncEnabled)
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
	})
	t.Run("Path", func(t *testing.T) {
		tempDir, cleanup := shared.CreateTempDir("noteleaf-config-handler-path-test-*", t)
		defer cleanup()

		customConfigPath := filepath.Join(tempDir, "my-config.toml")
		originalEnv := os.Getenv("NOTELEAF_CONFIG")
		os.Setenv("NOTELEAF_CONFIG", customConfigPath)
		defer os.Setenv("NOTELEAF_CONFIG", originalEnv)

		t.Run("Path returns correct config file path", func(t *testing.T) {
			handler, err := NewConfigHandler()
			if err != nil {
				t.Fatalf("Failed to create handler: %v", err)
			}

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
	})

	t.Run("Reset", func(t *testing.T) {
		tempDir, cleanup := shared.CreateTempDir("noteleaf-config-handler-reset-test-*", t)
		defer cleanup()

		customConfigPath := filepath.Join(tempDir, "test-config.toml")
		originalEnv := os.Getenv("NOTELEAF_CONFIG")
		os.Setenv("NOTELEAF_CONFIG", customConfigPath)
		defer os.Setenv("NOTELEAF_CONFIG", originalEnv)

		t.Run("Reset restores default config", func(t *testing.T) {
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
	})
}
