package store

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

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
func GetConfigDir() (string, error) {
	var configDir string

	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("APPDATA environment variable not set")
		}
		configDir = filepath.Join(appData, "noteleaf")
	default: // Unix-like systems (Linux, macOS, BSD, etc.)
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

	// Enable foreign keys and WAL mode for better performance
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

	// Run migrations
	if err := database.runMigrations(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return database, nil
}

// runMigrations applies all pending migrations
func (db *Database) runMigrations() error {
	// Get all migration files
	entries, err := migrationFiles.ReadDir("../sql/migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Filter and sort up migrations
	var upMigrations []string
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), "_up.sql") {
			upMigrations = append(upMigrations, entry.Name())
		}
	}
	sort.Strings(upMigrations)

	// Apply migrations in order
	for _, migrationFile := range upMigrations {
		version := extractVersionFromFilename(migrationFile)

		// Check if migration already applied
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='migrations'").Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check migrations table: %w", err)
		}

		// If migrations table doesn't exist and this isn't migration 0000, skip
		if count == 0 && version != "0000" {
			continue
		}

		// Check if this specific migration was applied (only if migrations table exists)
		if count > 0 {
			var applied int
			err = db.QueryRow("SELECT COUNT(*) FROM migrations WHERE version = ?", version).Scan(&applied)
			if err != nil {
				return fmt.Errorf("failed to check migration %s: %w", version, err)
			}
			if applied > 0 {
				continue // Skip already applied migration
			}
		}

		// Read and execute migration
		content, err := migrationFiles.ReadFile("../sql/migrations/" + migrationFile)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", migrationFile, err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", migrationFile, err)
		}

		// Record migration as applied (only if migrations table exists)
		if count > 0 || version == "0000" {
			if _, err := db.Exec("INSERT INTO migrations (version) VALUES (?)", version); err != nil {
				return fmt.Errorf("failed to record migration %s: %w", version, err)
			}
		}
	}

	return nil
}

// extractVersionFromFilename extracts the 4-digit version from a migration filename
func extractVersionFromFilename(filename string) string {
	parts := strings.Split(filename, "_")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// GetPath returns the database file path
func (db *Database) GetPath() string {
	return db.path
}

// Close closes the database connection
func (db *Database) Close() error {
	return db.DB.Close()
}
