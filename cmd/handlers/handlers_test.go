package handlers

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"stormlightlabs.org/noteleaf/internal/store"
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
