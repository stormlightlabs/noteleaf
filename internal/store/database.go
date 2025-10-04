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

// Database wraps [sql.DB] with application-specific methods
type Database struct {
	*sql.DB
	path string
}

// GetConfigDir returns the appropriate configuration directory based on [runtime.GOOS]
var GetConfigDir = func() (string, error) {
	var configDir string

	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("APPDATA environment variable not set")
		}
		configDir = filepath.Join(appData, "noteleaf")
	case "darwin":
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
		configDir = filepath.Join(homeDir, "Library", "Application Support", "noteleaf")
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

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return configDir, nil
}

// GetDataDir returns the appropriate data directory based on [runtime.GOOS] or NOTELEAF_DATA_DIR
var GetDataDir = func() (string, error) {
	if envDataDir := os.Getenv("NOTELEAF_DATA_DIR"); envDataDir != "" {
		if err := os.MkdirAll(envDataDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create data directory: %w", err)
		}
		return envDataDir, nil
	}

	var dataDir string

	switch runtime.GOOS {
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			return "", fmt.Errorf("LOCALAPPDATA environment variable not set")
		}
		dataDir = filepath.Join(localAppData, "noteleaf")
	case "darwin":
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
		dataDir = filepath.Join(homeDir, "Library", "Application Support", "noteleaf")
	default:
		xdgDataHome := os.Getenv("XDG_DATA_HOME")
		if xdgDataHome == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("failed to get user home directory: %w", err)
			}
			xdgDataHome = filepath.Join(homeDir, ".local", "share")
		}
		dataDir = filepath.Join(xdgDataHome, "noteleaf")
	}

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create data directory: %w", err)
	}

	return dataDir, nil
}

// NewDatabase creates and initializes a new database connection
var NewDatabase = func() (*Database, error) {
	return NewDatabaseWithConfig(nil)
}

// NewDatabaseWithConfig creates and initializes a new database connection using the provided config
func NewDatabaseWithConfig(config *Config) (*Database, error) {
	if config == nil {
		var err error
		config, err = LoadConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	}

	var dbPath string
	if config.DatabasePath != "" {
		dbPath = config.DatabasePath
		dbDir := filepath.Dir(dbPath)
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}
	} else if config.DataDir != "" {
		dbPath = filepath.Join(config.DataDir, "noteleaf.db")
	} else {
		dataDir, err := GetDataDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get data directory: %w", err)
		}
		dbPath = filepath.Join(dataDir, "noteleaf.db")
	}

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

	database := &Database{DB: db, path: dbPath}
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
