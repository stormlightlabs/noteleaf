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
