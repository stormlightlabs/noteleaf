package store

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed sql/migrations
var migrationFiles embed.FS

// Database wraps sql.DB with application-specific methods
type Database struct {
	*sql.DB
	path string
}

// GetConfigDir returns the appropriate configuration directory based on the OS
var GetConfigDir = func() (string, error) {
	var configDir string

	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("APPDATA environment variable not set")
		}
		configDir = filepath.Join(appData, "noteleaf")
	default:
		xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfigHome == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("failed to get user home directory: %w", err)
			}
			xdgConfigHome = filepath.Join(homeDir, ".config")
		}
		configDir = filepath.Join(xdgConfigHome, "noteleaf")
	}

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return configDir, nil
}

// NewDatabase creates and initializes a new database connection
func NewDatabase() (*Database, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config directory: %w", err)
	}

	dbPath := filepath.Join(configDir, "noteleaf.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	database := &Database{
		DB:   db,
		path: dbPath,
	}

	runner := NewMigrationRunner(db, migrationFiles)
	if err := runner.RunMigrations(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return database, nil
}

// NewMigrationRunnerFromDB creates a new migration runner from a Database instance
func (db *Database) NewMigrationRunner() *MigrationRunner {
	return NewMigrationRunner(db.DB, migrationFiles)
}

// GetPath returns the database file path
func (db *Database) GetPath() string {
	return db.path
}

// Close closes the database connection
func (db *Database) Close() error {
	return db.DB.Close()
}
