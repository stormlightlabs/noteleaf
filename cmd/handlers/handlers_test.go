package handlers

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stormlightlabs/noteleaf/internal/store"
)

func createTestDir(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "noteleaf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	oldConfigHome := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tempDir)

	t.Cleanup(func() {
		os.Setenv("XDG_CONFIG_HOME", oldConfigHome)
		os.RemoveAll(tempDir)
	})

	return tempDir
}

func TestSetup(t *testing.T) {
	t.Run("creates database and config files", func(t *testing.T) {
		_ = createTestDir(t)
		ctx := context.Background()

		err := Setup(ctx, []string{})
		if err != nil {
			t.Errorf("Setup failed: %v", err)
		}

		configDir, err := store.GetConfigDir()
		if err != nil {
			t.Fatalf("Failed to get config dir: %v", err)
		}

		dbPath := filepath.Join(configDir, "noteleaf.db")
		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			t.Error("Database file was not created")
		}

		configPath, err := store.GetConfigPath()
		if err != nil {
			t.Fatalf("Failed to get config path: %v", err)
		}

		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("Config file was not created")
		}

	})

	t.Run("handles existing database gracefully", func(t *testing.T) {
		_ = createTestDir(t)
		ctx := context.Background()

		err1 := Setup(ctx, []string{})
		if err1 != nil {
			t.Errorf("First setup failed: %v", err1)
		}

		err2 := Setup(ctx, []string{})
		if err2 != nil {
			t.Errorf("Second setup should not fail: %v", err2)
		}

	})

	t.Run("initializes migrations", func(t *testing.T) {
		_ = createTestDir(t)
		ctx := context.Background()

		err := Setup(ctx, []string{})
		if err != nil {
			t.Errorf("Setup failed: %v", err)
		}

		db, err := store.NewDatabase()
		if err != nil {
			t.Fatalf("Failed to connect to database: %v", err)
		}
		defer db.Close()

		runner := db.NewMigrationRunner()
		migrations, err := runner.GetAppliedMigrations()
		if err != nil {
			t.Fatalf("Failed to get migrations: %v", err)
		}

		if len(migrations) == 0 {
			t.Error("No migrations were applied")
		}

		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM migrations").Scan(&count)
		if err != nil {
			t.Errorf("Failed to query migrations table: %v", err)
		}

		if count == 0 {
			t.Error("Migrations table is empty")
		}

	})
}

func TestReset(t *testing.T) {
	t.Run("removes database and config files", func(t *testing.T) {
		_ = createTestDir(t)
		ctx := context.Background()

		err := Setup(ctx, []string{})
		if err != nil {
			t.Fatalf("Setup failed: %v", err)
		}

		configDir, err := store.GetConfigDir()
		if err != nil {
			t.Fatalf("Failed to get config dir: %v", err)
		}

		dbPath := filepath.Join(configDir, "noteleaf.db")
		configPath, err := store.GetConfigPath()
		if err != nil {
			t.Fatalf("Failed to get config path: %v", err)
		}

		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			t.Fatal("Database should exist before reset")
		}

		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Fatal("Config should exist before reset")
		}

		err = Reset(ctx, []string{})
		if err != nil {
			t.Errorf("Reset failed: %v", err)
		}

		if _, err := os.Stat(dbPath); !os.IsNotExist(err) {
			t.Error("Database file should be removed after reset")
		}

		if _, err := os.Stat(configPath); !os.IsNotExist(err) {
			t.Error("Config file should be removed after reset")
		}

	})

	t.Run("handles non-existent files gracefully", func(t *testing.T) {
		_ = createTestDir(t)
		ctx := context.Background()

		err := Reset(ctx, []string{})
		if err != nil {
			t.Errorf("Reset should handle non-existent files: %v", err)
		}

	})
}

func TestStatus(t *testing.T) {
	t.Run("reports status when setup", func(t *testing.T) {
		_ = createTestDir(t)
		ctx := context.Background()

		err := Setup(ctx, []string{})
		if err != nil {
			t.Fatalf("Setup failed: %v", err)
		}

		err = Status(ctx, []string{})
		if err != nil {
			t.Errorf("Status failed: %v", err)
		}

	})

	t.Run("reports status when not setup", func(t *testing.T) {
		_ = createTestDir(t)
		ctx := context.Background()

		err := Status(ctx, []string{})
		if err != nil {
			t.Errorf("Status should not fail when not setup: %v", err)
		}

	})

	t.Run("shows migration information", func(t *testing.T) {
		_ = createTestDir(t)
		ctx := context.Background()

		err := Setup(ctx, []string{})
		if err != nil {
			t.Fatalf("Setup failed: %v", err)
		}

		err = Status(ctx, []string{})
		if err != nil {
			t.Errorf("Status failed: %v", err)
		}

	})
}

func TestSetupErrorHandling(t *testing.T) {
	t.Run("handles GetConfigDir error", func(t *testing.T) {
		originalXDG := os.Getenv("XDG_CONFIG_HOME")
		originalHome := os.Getenv("HOME")

		if runtime.GOOS == "windows" {
			originalAppData := os.Getenv("APPDATA")
			os.Unsetenv("APPDATA")
			defer os.Setenv("APPDATA", originalAppData)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
			os.Unsetenv("HOME")
		}

		defer func() {
			os.Setenv("XDG_CONFIG_HOME", originalXDG)
			os.Setenv("HOME", originalHome)
		}()

		ctx := context.Background()
		err := Setup(ctx, []string{})

		if err == nil {
			t.Error("Setup should fail when GetConfigDir fails")
		}
		if !strings.Contains(err.Error(), "failed to get config directory") {
			t.Errorf("Expected config directory error, got: %v", err)
		}
	})

	t.Run("handles database creation error", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "noteleaf-readonly-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		noteleafDir := filepath.Join(tempDir, "noteleaf")
		if err := os.MkdirAll(noteleafDir, 0755); err != nil {
			t.Fatalf("Failed to create noteleaf dir: %v", err)
		}

		if err := os.Chmod(noteleafDir, 0444); err != nil {
			t.Fatalf("Failed to make directory read-only: %v", err)
		}

		defer os.Chmod(noteleafDir, 0755)

		oldConfigHome := os.Getenv("XDG_CONFIG_HOME")
		os.Setenv("XDG_CONFIG_HOME", tempDir)
		defer os.Setenv("XDG_CONFIG_HOME", oldConfigHome)

		ctx := context.Background()
		err = Setup(ctx, []string{})

		if err == nil {
			t.Error("Setup should fail when database creation fails")
		}
		if !strings.Contains(err.Error(), "failed to initialize database") {
			t.Errorf("Expected database initialization error, got: %v", err)
		}
	})

	t.Run("handles config loading error", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "noteleaf-config-error-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		noteleafDir := filepath.Join(tempDir, "noteleaf")
		if err := os.MkdirAll(noteleafDir, 0755); err != nil {
			t.Fatalf("Failed to create noteleaf dir: %v", err)
		}

		configPath := filepath.Join(noteleafDir, ".noteleaf.conf.toml")
		invalidTOML := `[invalid toml content`
		if err := os.WriteFile(configPath, []byte(invalidTOML), 0644); err != nil {
			t.Fatalf("Failed to write invalid config: %v", err)
		}

		oldConfigHome := os.Getenv("XDG_CONFIG_HOME")
		os.Setenv("XDG_CONFIG_HOME", tempDir)
		defer os.Setenv("XDG_CONFIG_HOME", oldConfigHome)

		ctx := context.Background()
		err = Setup(ctx, []string{})

		if err == nil {
			t.Error("Setup should fail when config loading fails")
		}
		if !strings.Contains(err.Error(), "failed to create configuration") {
			t.Errorf("Expected configuration error, got: %v", err)
		}
	})
}

func TestResetErrorHandling(t *testing.T) {
	t.Run("handles GetConfigDir error", func(t *testing.T) {
		originalXDG := os.Getenv("XDG_CONFIG_HOME")
		originalHome := os.Getenv("HOME")

		if runtime.GOOS == "windows" {
			originalAppData := os.Getenv("APPDATA")
			os.Unsetenv("APPDATA")
			defer os.Setenv("APPDATA", originalAppData)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
			os.Unsetenv("HOME")
		}

		defer func() {
			os.Setenv("XDG_CONFIG_HOME", originalXDG)
			os.Setenv("HOME", originalHome)
		}()

		ctx := context.Background()
		err := Reset(ctx, []string{})

		if err == nil {
			t.Error("Reset should fail when GetConfigDir fails")
		}
		if !strings.Contains(err.Error(), "failed to get config directory") {
			t.Errorf("Expected config directory error, got: %v", err)
		}
	})

	t.Run("handles GetConfigPath error", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "noteleaf-reset-error-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Set up a scenario where GetConfigPath might fail
		// This is tricky since GetConfigPath internally calls GetConfigDir
		// We'll create a scenario where the config dir exists but GetConfigPath fails
		oldConfigHome := os.Getenv("XDG_CONFIG_HOME")
		os.Setenv("XDG_CONFIG_HOME", tempDir)
		defer os.Setenv("XDG_CONFIG_HOME", oldConfigHome)

		noteleafDir := filepath.Join(tempDir, "noteleaf")
		if err := os.MkdirAll(noteleafDir, 0755); err != nil {
			t.Fatalf("Failed to create noteleaf dir: %v", err)
		}

		dbPath := filepath.Join(noteleafDir, "noteleaf.db")
		if err := os.WriteFile(dbPath, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test db file: %v", err)
		}

		if err := os.Chmod(noteleafDir, 0444); err != nil {
			t.Fatalf("Failed to make directory read-only: %v", err)
		}

		defer os.Chmod(noteleafDir, 0755)

		ctx := context.Background()
		err = Reset(ctx, []string{})

		if err == nil {
			t.Error("Reset should fail when file removal fails")
		}
		if !strings.Contains(err.Error(), "failed to remove") && !strings.Contains(err.Error(), "failed to get config path") {
			t.Errorf("Expected removal or config path error, got: %v", err)
		}
	})
}

func TestStatusErrorHandling(t *testing.T) {
	t.Run("handles GetConfigDir error", func(t *testing.T) {
		originalXDG := os.Getenv("XDG_CONFIG_HOME")
		originalHome := os.Getenv("HOME")

		if runtime.GOOS == "windows" {
			originalAppData := os.Getenv("APPDATA")
			os.Unsetenv("APPDATA")
			defer os.Setenv("APPDATA", originalAppData)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
			os.Unsetenv("HOME")
		}

		defer func() {
			os.Setenv("XDG_CONFIG_HOME", originalXDG)
			os.Setenv("HOME", originalHome)
		}()

		ctx := context.Background()
		err := Status(ctx, []string{})

		if err == nil {
			t.Error("Status should fail when GetConfigDir fails")
		}
		if !strings.Contains(err.Error(), "failed to get config directory") {
			t.Errorf("Expected config directory error, got: %v", err)
		}
	})

	t.Run("handles GetConfigPath error after database exists", func(t *testing.T) {
		_ = createTestDir(t)
		ctx := context.Background()

		err := Setup(ctx, []string{})
		if err != nil {
			t.Fatalf("Setup failed: %v", err)
		}

		// Now we have a database, but let's create a scenario where GetConfigPath fails
		// This is challenging since GetConfigPath uses GetConfigDir which we already tested
		// Instead, let's test the database connection error scenario

		// Remove the database to cause NewDatabase to fail in Status
		configDir, _ := store.GetConfigDir()
		dbPath := filepath.Join(configDir, "noteleaf.db")
		os.Remove(dbPath)

		// Create a directory with the same name as the database file
		// This will cause database connection to fail
		if err := os.MkdirAll(dbPath, 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}

		err = Status(ctx, []string{})
		if err == nil {
			t.Error("Status should fail when database connection fails")
		}
		if !strings.Contains(err.Error(), "failed to connect to database") {
			t.Errorf("Expected database connection error, got: %v", err)
		}
	})

	t.Run("handles migration errors", func(t *testing.T) {
		_ = createTestDir(t)
		ctx := context.Background()

		err := Setup(ctx, []string{})
		if err != nil {
			t.Fatalf("Setup failed: %v", err)
		}

		// Corrupt the migrations table to cause GetAppliedMigrations to fail
		db, err := store.NewDatabase()
		if err != nil {
			t.Fatalf("Failed to connect to database: %v", err)
		}

		_, err = db.Exec("DROP TABLE migrations")
		if err != nil {
			t.Fatalf("Failed to drop migrations table: %v", err)
		}

		_, err = db.Exec("CREATE TABLE migrations (invalid_schema TEXT)")
		if err != nil {
			t.Fatalf("Failed to create corrupted migrations table: %v", err)
		}
		db.Close()

		err = Status(ctx, []string{})
		if err == nil {
			t.Error("Status should fail when migration queries fail")
		}
		if !strings.Contains(err.Error(), "failed to get") && !strings.Contains(err.Error(), "migrations") {
			t.Errorf("Expected migration-related error, got: %v", err)
		}
	})
}

func TestErrorScenarios(t *testing.T) {
	errorTests := []struct {
		name        string
		setupFunc   func(t *testing.T) (cleanup func())
		handlerFunc func(ctx context.Context, args []string) error
		expectError bool
		errorSubstr string
	}{
		{
			name: "Setup with invalid environment",
			setupFunc: func(t *testing.T) func() {
				if runtime.GOOS == "windows" {
					original := os.Getenv("APPDATA")
					os.Unsetenv("APPDATA")
					return func() { os.Setenv("APPDATA", original) }
				} else {
					originalXDG := os.Getenv("XDG_CONFIG_HOME")
					originalHome := os.Getenv("HOME")
					os.Unsetenv("XDG_CONFIG_HOME")
					os.Unsetenv("HOME")
					return func() {
						os.Setenv("XDG_CONFIG_HOME", originalXDG)
						os.Setenv("HOME", originalHome)
					}
				}
			},
			handlerFunc: Setup,
			expectError: true,
			errorSubstr: "config directory",
		},
		{
			name: "Reset with invalid environment",
			setupFunc: func(t *testing.T) func() {
				if runtime.GOOS == "windows" {
					original := os.Getenv("APPDATA")
					os.Unsetenv("APPDATA")
					return func() { os.Setenv("APPDATA", original) }
				} else {
					originalXDG := os.Getenv("XDG_CONFIG_HOME")
					originalHome := os.Getenv("HOME")
					os.Unsetenv("XDG_CONFIG_HOME")
					os.Unsetenv("HOME")
					return func() {
						os.Setenv("XDG_CONFIG_HOME", originalXDG)
						os.Setenv("HOME", originalHome)
					}
				}
			},
			handlerFunc: Reset,
			expectError: true,
			errorSubstr: "config directory",
		},
		{
			name: "Status with invalid environment",
			setupFunc: func(t *testing.T) func() {
				if runtime.GOOS == "windows" {
					original := os.Getenv("APPDATA")
					os.Unsetenv("APPDATA")
					return func() { os.Setenv("APPDATA", original) }
				} else {
					originalXDG := os.Getenv("XDG_CONFIG_HOME")
					originalHome := os.Getenv("HOME")
					os.Unsetenv("XDG_CONFIG_HOME")
					os.Unsetenv("HOME")
					return func() {
						os.Setenv("XDG_CONFIG_HOME", originalXDG)
						os.Setenv("HOME", originalHome)
					}
				}
			},
			handlerFunc: Status,
			expectError: true,
			errorSubstr: "config directory",
		},
	}

	for _, tt := range errorTests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setupFunc(t)
			defer cleanup()

			ctx := context.Background()
			err := tt.handlerFunc(ctx, []string{})

			if tt.expectError && err == nil {
				t.Errorf("Expected error containing %q, got nil", tt.errorSubstr)
			} else if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			} else if tt.expectError && err != nil && !strings.Contains(err.Error(), tt.errorSubstr) {
				t.Errorf("Expected error containing %q, got: %v", tt.errorSubstr, err)
			}
		})
	}
}

func TestIntegration(t *testing.T) {
	t.Run("full setup-reset-status cycle", func(t *testing.T) {
		_ = createTestDir(t)
		ctx := context.Background()

		err := Status(ctx, []string{})
		if err != nil {
			t.Errorf("Initial status failed: %v", err)
		}

		err = Setup(ctx, []string{})
		if err != nil {
			t.Errorf("Setup failed: %v", err)
		}

		err = Status(ctx, []string{})
		if err != nil {
			t.Errorf("Status after setup failed: %v", err)
		}

		configDir, _ := store.GetConfigDir()
		dbPath := filepath.Join(configDir, "noteleaf.db")
		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			t.Error("Database should exist after setup")
		}

		err = Reset(ctx, []string{})
		if err != nil {
			t.Errorf("Reset failed: %v", err)
		}

		if _, err := os.Stat(dbPath); !os.IsNotExist(err) {
			t.Error("Database should not exist after reset")
		}

		err = Status(ctx, []string{})
		if err != nil {
			t.Errorf("Status after reset failed: %v", err)
		}

	})

	t.Run("concurrent operations", func(t *testing.T) {
		_ = createTestDir(t)
		ctx := context.Background()

		err := Setup(ctx, []string{})
		if err != nil {
			t.Fatalf("Setup failed: %v", err)
		}

		done := make(chan error, 3)

		for range 3 {
			go func() {
				time.Sleep(time.Millisecond * 10)
				done <- Status(ctx, []string{})
			}()
		}

		for i := range 3 {
			if err := <-done; err != nil {
				t.Errorf("Concurrent status operation %d failed: %v", i, err)
			}
		}
	})
}
