package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func withTempDirs(t *testing.T) string {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "noteleaf-db-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(tempDir) })

	origConfig, origData := GetConfigDir, GetDataDir
	GetConfigDir = func() (string, error) { return tempDir, nil }
	GetDataDir = func() (string, error) { return tempDir, nil }
	t.Cleanup(func() {
		GetConfigDir, GetDataDir = origConfig, origData
	})

	return tempDir
}

func TestNewDatabase(t *testing.T) {
	tempDir := withTempDirs(t)

	t.Run("creates database file", func(t *testing.T) {
		db, err := NewDatabase()
		if err != nil {
			t.Fatalf("NewDatabase failed: %v", err)
		}
		defer db.Close()

		dbPath := filepath.Join(tempDir, "noteleaf.db")
		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			t.Errorf("expected database file at %s", dbPath)
		}
	})

	t.Run("foreign keys enabled", func(t *testing.T) {
		db, _ := NewDatabase()
		defer db.Close()

		var fk int
		if err := db.QueryRow("PRAGMA foreign_keys").Scan(&fk); err != nil {
			t.Fatalf("query failed: %v", err)
		}
		if fk != 1 {
			t.Errorf("expected foreign_keys=1, got %d", fk)
		}
	})

	t.Run("WAL enabled", func(t *testing.T) {
		db, _ := NewDatabase()
		defer db.Close()

		var mode string
		if err := db.QueryRow("PRAGMA journal_mode").Scan(&mode); err != nil {
			t.Fatalf("query failed: %v", err)
		}
		if strings.ToLower(mode) != "wal" {
			t.Errorf("expected wal, got %s", mode)
		}
	})
}

func TestNewDatabase_ErrorPaths(t *testing.T) {
	t.Run("sql.Open fails", func(t *testing.T) {
		orig := sqlOpen
		sqlOpen = func(driver, dsn string) (*sql.DB, error) {
			return nil, fmt.Errorf("boom")
		}
		t.Cleanup(func() { sqlOpen = orig })

		_, err := NewDatabase()
		if err == nil || !strings.Contains(err.Error(), "failed to open database") {
			t.Errorf("expected open error, got %v", err)
		}
	})

	t.Run("foreign_keys pragma fails", func(t *testing.T) {
		orig := pragmaExec
		pragmaExec = func(db *sql.DB, stmt string) (sql.Result, error) {
			if strings.Contains(stmt, "foreign_keys") {
				return nil, fmt.Errorf("fk fail")
			}
			return orig(db, stmt)
		}
		t.Cleanup(func() { pragmaExec = orig })

		_, err := NewDatabase()
		if err == nil || !strings.Contains(err.Error(), "failed to enable foreign keys") {
			t.Errorf("expected foreign key error, got %v", err)
		}
	})

	t.Run("WAL pragma fails", func(t *testing.T) {
		orig := pragmaExec
		pragmaExec = func(db *sql.DB, stmt string) (sql.Result, error) {
			if strings.Contains(stmt, "journal_mode") {
				return nil, fmt.Errorf("wal fail")
			}
			return orig(db, stmt)
		}
		t.Cleanup(func() { pragmaExec = orig })

		_, err := NewDatabase()
		if err == nil || !strings.Contains(err.Error(), "failed to enable WAL mode") {
			t.Errorf("expected WAL error, got %v", err)
		}
	})

	t.Run("migration runner fails", func(t *testing.T) {
		orig := createMigrationRunner
		createMigrationRunner = func(db *sql.DB, fs FileSystem) *MigrationRunner {
			return &MigrationRunner{runFn: func() error { return fmt.Errorf("migration fail") }}
		}
		t.Cleanup(func() { createMigrationRunner = orig })

		_, err := NewDatabase()
		if err == nil || !strings.Contains(err.Error(), "failed to run migrations") {
			t.Errorf("expected migration error, got %v", err)
		}
	})

	t.Run("permission denied on config dir", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("not reliable on Windows")
		}
		dir := withTempDirs(t)
		os.Chmod(dir, 0555) // make read-only
		defer os.Chmod(dir, 0755)

		GetConfigDir = func() (string, error) { return dir, nil }

		_, err := NewDatabase()
		if err == nil {
			t.Error("expected mkdir fail due to permission denied")
		}
	})
}

func TestGetConfigDir_AllBranches(t *testing.T) {
	tmp := t.TempDir()

	t.Run("windows success", func(t *testing.T) {
		origGOOS := getRuntime
		getRuntime = func() string { return "windows" }
		defer func() { getRuntime = origGOOS }()

		os.Setenv("APPDATA", tmp)
		defer os.Unsetenv("APPDATA")

		dir, err := GetConfigDir()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := filepath.Join(tmp, "noteleaf")
		if dir != expected {
			t.Errorf("expected %s, got %s", expected, dir)
		}
	})

	t.Run("windows missing APPDATA", func(t *testing.T) {
		origGOOS := getRuntime
		getRuntime = func() string { return "windows" }
		defer func() { getRuntime = origGOOS }()

		os.Unsetenv("APPDATA")

		_, err := GetConfigDir()
		if err == nil || !strings.Contains(err.Error(), "APPDATA") {
			t.Errorf("expected APPDATA error, got %v", err)
		}
	})

	t.Run("darwin success", func(t *testing.T) {
		origGOOS, origHome := getRuntime, getHomeDir
		getRuntime = func() string { return "darwin" }
		getHomeDir = func() (string, error) { return tmp, nil }
		defer func() {
			getRuntime = origGOOS
			getHomeDir = origHome
		}()

		dir, err := GetConfigDir()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := filepath.Join(tmp, "Library", "Application Support", "noteleaf")
		if dir != expected {
			t.Errorf("expected %s, got %s", expected, dir)
		}
	})

	t.Run("linux default with XDG_CONFIG_HOME unset", func(t *testing.T) {
		origGOOS, origHome := getRuntime, getHomeDir
		getRuntime = func() string { return "linux" }
		getHomeDir = func() (string, error) { return tmp, nil }
		defer func() {
			getRuntime = origGOOS
			getHomeDir = origHome
		}()

		os.Unsetenv("XDG_CONFIG_HOME")

		dir, err := GetConfigDir()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := filepath.Join(tmp, ".config", "noteleaf")
		if dir != expected {
			t.Errorf("expected %s, got %s", expected, dir)
		}
	})
}

func TestGetDataDir_AllBranches(t *testing.T) {
	tmp := t.TempDir()

	t.Run("NOTELEAF_DATA_DIR overrides", func(t *testing.T) {
		os.Setenv("NOTELEAF_DATA_DIR", tmp)
		defer os.Unsetenv("NOTELEAF_DATA_DIR")

		dir, err := GetDataDir()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if dir != tmp {
			t.Errorf("expected %s, got %s", tmp, dir)
		}
	})

	t.Run("windows success", func(t *testing.T) {
		origGOOS := getRuntime
		getRuntime = func() string { return "windows" }
		defer func() { getRuntime = origGOOS }()

		os.Setenv("LOCALAPPDATA", tmp)
		defer os.Unsetenv("LOCALAPPDATA")

		dir, err := GetDataDir()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := filepath.Join(tmp, "noteleaf")
		if dir != expected {
			t.Errorf("expected %s, got %s", expected, dir)
		}
	})

	t.Run("windows missing LOCALAPPDATA", func(t *testing.T) {
		origGOOS := getRuntime
		getRuntime = func() string { return "windows" }
		defer func() { getRuntime = origGOOS }()

		os.Unsetenv("LOCALAPPDATA")

		_, err := GetDataDir()
		if err == nil || !strings.Contains(err.Error(), "LOCALAPPDATA") {
			t.Errorf("expected LOCALAPPDATA error, got %v", err)
		}
	})

	t.Run("darwin success", func(t *testing.T) {
		origGOOS, origHome := getRuntime, getHomeDir
		getRuntime = func() string { return "darwin" }
		getHomeDir = func() (string, error) { return tmp, nil }
		defer func() {
			getRuntime = origGOOS
			getHomeDir = origHome
		}()

		dir, err := GetDataDir()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := filepath.Join(tmp, "Library", "Application Support", "noteleaf")
		if dir != expected {
			t.Errorf("expected %s, got %s", expected, dir)
		}
	})

	t.Run("linux default with XDG_DATA_HOME unset", func(t *testing.T) {
		origGOOS, origHome := getRuntime, getHomeDir
		getRuntime = func() string { return "linux" }
		getHomeDir = func() (string, error) { return tmp, nil }
		defer func() {
			getRuntime = origGOOS
			getHomeDir = origHome
		}()

		os.Unsetenv("XDG_DATA_HOME")

		dir, err := GetDataDir()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := filepath.Join(tmp, ".local", "share", "noteleaf")
		if dir != expected {
			t.Errorf("expected %s, got %s", expected, dir)
		}
	})
}
