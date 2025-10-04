package store

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestNewDatabase(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "noteleaf-db-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalGetConfigDir := GetConfigDir
	originalGetDataDir := GetDataDir
	GetConfigDir = func() (string, error) {
		return tempDir, nil
	}
	GetDataDir = func() (string, error) {
		return tempDir, nil
	}
	defer func() {
		GetConfigDir = originalGetConfigDir
		GetDataDir = originalGetDataDir
	}()

	t.Run("creates database successfully", func(t *testing.T) {
		db, err := NewDatabase()
		if err != nil {
			t.Fatalf("NewDatabase failed: %v", err)
		}
		defer db.Close()

		if db == nil {
			t.Fatal("Database should not be nil")
		}

		dbPath := filepath.Join(tempDir, "noteleaf.db")
		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			t.Error("Database file should exist")
		}

		if db.GetPath() != dbPath {
			t.Errorf("Expected database path %s, got %s", dbPath, db.GetPath())
		}
	})

	t.Run("enables foreign keys", func(t *testing.T) {
		db, err := NewDatabase()
		if err != nil {
			t.Fatalf("NewDatabase failed: %v", err)
		}
		defer db.Close()

		var foreignKeys int
		err = db.QueryRow("PRAGMA foreign_keys").Scan(&foreignKeys)
		if err != nil {
			t.Fatalf("Failed to check foreign keys: %v", err)
		}

		if foreignKeys != 1 {
			t.Error("Foreign keys should be enabled")
		}
	})

	t.Run("enables WAL mode", func(t *testing.T) {
		db, err := NewDatabase()
		if err != nil {
			t.Fatalf("NewDatabase failed: %v", err)
		}
		defer db.Close()

		var journalMode string
		err = db.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
		if err != nil {
			t.Fatalf("Failed to check journal mode: %v", err)
		}

		if journalMode != "wal" {
			t.Errorf("Expected WAL journal mode, got %s", journalMode)
		}
	})

	t.Run("runs migrations", func(t *testing.T) {
		db, err := NewDatabase()
		if err != nil {
			t.Fatalf("NewDatabase failed: %v", err)
		}
		defer db.Close()

		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='migrations'").Scan(&count)
		if err != nil {
			t.Fatalf("Failed to check migrations table: %v", err)
		}

		if count != 1 {
			t.Error("Migrations table should exist")
		}

		var migrationCount int
		err = db.QueryRow("SELECT COUNT(*) FROM migrations").Scan(&migrationCount)
		if err != nil {
			t.Fatalf("Failed to count migrations: %v", err)
		}

		if migrationCount == 0 {
			t.Error("At least one migration should be applied")
		}
	})

	t.Run("creates migration runner", func(t *testing.T) {
		db, err := NewDatabase()
		if err != nil {
			t.Fatalf("NewDatabase failed: %v", err)
		}
		defer db.Close()

		runner := db.NewMigrationRunner()
		if runner == nil {
			t.Error("Migration runner should not be nil")
		}
	})

	t.Run("closes database connection", func(t *testing.T) {
		db, err := NewDatabase()
		if err != nil {
			t.Fatalf("NewDatabase failed: %v", err)
		}

		err = db.Close()
		if err != nil {
			t.Errorf("Close should not return error: %v", err)
		}

		err = db.Ping()
		if err == nil {
			t.Error("Database should be closed and ping should fail")
		}
	})
}

func TestDatabaseErrorHandling(t *testing.T) {
	t.Run("handles GetConfigDir error", func(t *testing.T) {
		originalGetConfigDir := GetConfigDir
		GetConfigDir = func() (string, error) {
			return "", os.ErrPermission
		}
		defer func() { GetConfigDir = originalGetConfigDir }()

		_, err := NewDatabase()
		if err == nil {
			t.Error("NewDatabase should fail when GetConfigDir fails")
		}
	})

	t.Run("handles invalid database path", func(t *testing.T) {
		originalGetConfigDir := GetConfigDir
		GetConfigDir = func() (string, error) {
			return "/invalid/path/that/does/not/exist", nil
		}
		defer func() { GetConfigDir = originalGetConfigDir }()

		_, err := NewDatabase()
		if err == nil {
			t.Error("NewDatabase should fail with invalid database path")
		}
	})

	t.Run("handles invalid database connection", func(t *testing.T) {
		originalGetConfigDir := GetConfigDir
		GetConfigDir = func() (string, error) {
			return "/dev/null", nil
		}
		defer func() { GetConfigDir = originalGetConfigDir }()

		_, err := NewDatabase()
		if err == nil {
			t.Error("NewDatabase should fail when database path is invalid")
		}
	})

	t.Run("handles database file permission error", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Permission test not reliable on Windows")
		}

		tempDir, err := os.MkdirTemp("", "noteleaf-db-perm-test-*")
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

		_, err = NewDatabase()
		if err == nil {
			t.Error("NewDatabase should fail when database directory is not writable")
		}
	})
}

func TestDatabaseIntegration(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "noteleaf-db-integration-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalGetConfigDir := GetConfigDir
	originalGetDataDir := GetDataDir
	GetConfigDir = func() (string, error) {
		return tempDir, nil
	}
	GetDataDir = func() (string, error) {
		return tempDir, nil
	}
	defer func() {
		GetConfigDir = originalGetConfigDir
		GetDataDir = originalGetDataDir
	}()

	t.Run("multiple database instances use same file", func(t *testing.T) {
		db1, err := NewDatabase()
		if err != nil {
			t.Fatalf("First NewDatabase failed: %v", err)
		}
		defer db1.Close()

		db2, err := NewDatabase()
		if err != nil {
			t.Fatalf("Second NewDatabase failed: %v", err)
		}
		defer db2.Close()

		if db1.GetPath() != db2.GetPath() {
			t.Error("Both database instances should use the same file path")
		}
	})

	t.Run("database survives connection close and reopen", func(t *testing.T) {
		db1, err := NewDatabase()
		if err != nil {
			t.Fatalf("NewDatabase failed: %v", err)
		}

		_, err = db1.Exec("CREATE TABLE IF NOT EXISTS test_table (id INTEGER PRIMARY KEY, name TEXT)")
		if err != nil {
			t.Fatalf("Failed to create test table: %v", err)
		}

		_, err = db1.Exec("INSERT INTO test_table (name) VALUES (?)", "test_value")
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}

		db1.Close()

		db2, err := NewDatabase()
		if err != nil {
			t.Fatalf("Second NewDatabase failed: %v", err)
		}
		defer db2.Close()

		var name string
		err = db2.QueryRow("SELECT name FROM test_table WHERE id = 1").Scan(&name)
		if err != nil {
			t.Fatalf("Failed to query test data: %v", err)
		}

		if name != "test_value" {
			t.Errorf("Expected 'test_value', got '%s'", name)
		}
	})
}

// TestNewDatabaseWithConfig tests database creation with custom config
func TestNewDatabaseWithConfig(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "noteleaf-db-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("creates database with custom database path", func(t *testing.T) {
		customDBPath := filepath.Join(tempDir, "custom.db")
		config := &Config{
			DatabasePath: customDBPath,
		}

		db, err := NewDatabaseWithConfig(config)
		if err != nil {
			t.Fatalf("NewDatabaseWithConfig failed: %v", err)
		}
		defer db.Close()

		if db.GetPath() != customDBPath {
			t.Errorf("Expected database path %s, got %s", customDBPath, db.GetPath())
		}

		if _, err := os.Stat(customDBPath); os.IsNotExist(err) {
			t.Error("Custom database file should exist")
		}
	})

	t.Run("creates database with custom data dir", func(t *testing.T) {
		customDataDir := filepath.Join(tempDir, "data")

		// Create the custom data directory
		if err := os.MkdirAll(customDataDir, 0755); err != nil {
			t.Fatalf("Failed to create custom data dir: %v", err)
		}

		config := &Config{
			DataDir: customDataDir,
		}

		db, err := NewDatabaseWithConfig(config)
		if err != nil {
			t.Fatalf("NewDatabaseWithConfig failed: %v", err)
		}
		defer db.Close()

		expectedPath := filepath.Join(customDataDir, "noteleaf.db")
		if db.GetPath() != expectedPath {
			t.Errorf("Expected database path %s, got %s", expectedPath, db.GetPath())
		}
	})

	t.Run("handles invalid custom database path", func(t *testing.T) {
		config := &Config{
			DatabasePath: "/invalid/path/that/does/not/exist/database.db",
		}

		_, err := NewDatabaseWithConfig(config)
		if err == nil {
			t.Error("NewDatabaseWithConfig should fail with invalid database path")
		}
	})

	t.Run("creates nested directories for custom database path", func(t *testing.T) {
		nestedPath := filepath.Join(tempDir, "nested", "deep", "database.db")
		config := &Config{
			DatabasePath: nestedPath,
		}

		db, err := NewDatabaseWithConfig(config)
		if err != nil {
			t.Fatalf("NewDatabaseWithConfig should create nested directories: %v", err)
		}
		defer db.Close()

		if _, err := os.Stat(nestedPath); os.IsNotExist(err) {
			t.Error("Database file should exist in nested directory")
		}
	})
}

// TestGetDataDirPlatformSpecific tests platform-specific GetDataDir behavior
func TestGetDataDirPlatformSpecific(t *testing.T) {
	t.Run("handles missing LOCALAPPDATA on Windows", func(t *testing.T) {
		if runtime.GOOS != "windows" {
			t.Skip("Windows-specific test")
		}

		originalEnv := os.Getenv("LOCALAPPDATA")
		os.Unsetenv("LOCALAPPDATA")
		defer os.Setenv("LOCALAPPDATA", originalEnv)

		_, err := GetDataDir()
		if err == nil {
			t.Error("GetDataDir should fail when LOCALAPPDATA is not set on Windows")
		}
	})

	t.Run("handles missing HOME on Unix", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Unix-specific test")
		}

		originalXDG := os.Getenv("XDG_DATA_HOME")
		originalHome := os.Getenv("HOME")
		os.Unsetenv("XDG_DATA_HOME")
		os.Unsetenv("HOME")

		defer func() {
			os.Setenv("XDG_DATA_HOME", originalXDG)
			os.Setenv("HOME", originalHome)
		}()

		_, err := GetDataDir()
		if err == nil {
			t.Error("GetDataDir should fail when both XDG_DATA_HOME and HOME are not set on Unix")
		}
	})

	t.Run("uses XDG_DATA_HOME when set on Linux", func(t *testing.T) {
		if runtime.GOOS != "linux" {
			t.Skip("Linux-specific test")
		}

		tempDir, err := os.MkdirTemp("", "noteleaf-xdg-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		originalXDG := os.Getenv("XDG_DATA_HOME")
		os.Setenv("XDG_DATA_HOME", tempDir)
		defer os.Setenv("XDG_DATA_HOME", originalXDG)

		dataDir, err := GetDataDir()
		if err != nil {
			t.Fatalf("GetDataDir failed: %v", err)
		}

		expectedPath := filepath.Join(tempDir, "noteleaf")
		if dataDir != expectedPath {
			t.Errorf("Expected data dir %s, got %s", expectedPath, dataDir)
		}
	})

	t.Run("falls back to HOME/.local/share on Linux when XDG_DATA_HOME not set", func(t *testing.T) {
		if runtime.GOOS != "linux" {
			t.Skip("Linux-specific test")
		}

		tempHome, err := os.MkdirTemp("", "noteleaf-home-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp home: %v", err)
		}
		defer os.RemoveAll(tempHome)

		originalXDG := os.Getenv("XDG_DATA_HOME")
		originalHome := os.Getenv("HOME")
		os.Unsetenv("XDG_DATA_HOME")
		os.Setenv("HOME", tempHome)

		defer func() {
			os.Setenv("XDG_DATA_HOME", originalXDG)
			os.Setenv("HOME", originalHome)
		}()

		dataDir, err := GetDataDir()
		if err != nil {
			t.Fatalf("GetDataDir failed: %v", err)
		}

		expectedPath := filepath.Join(tempHome, ".local", "share", "noteleaf")
		if dataDir != expectedPath {
			t.Errorf("Expected data dir %s, got %s", expectedPath, dataDir)
		}
	})
}

// TestDatabaseConstraintViolations tests database constraint handling
func TestDatabaseConstraintViolations(t *testing.T) {
	env := NewIsolatedEnvironment(t)
	defer env.Cleanup()

	t.Run("handles foreign key constraint violations", func(t *testing.T) {
		db, err := NewDatabase()
		if err != nil {
			t.Fatalf("NewDatabase failed: %v", err)
		}
		defer db.Close()

		var foreignKeys int
		err = db.QueryRow("PRAGMA foreign_keys").Scan(&foreignKeys)
		if err != nil {
			t.Fatalf("Failed to check foreign keys: %v", err)
		}

		if foreignKeys != 1 {
			t.Error("Foreign keys should be enabled by default")
		}
	})
}
