package store

import (
	"database/sql"
	"embed"
	"fmt"
	"sort"
	"strings"
)

// Migration represents a single database migration
type Migration struct {
	Version   string
	Name      string
	UpSQL     string
	DownSQL   string
	Applied   bool
	AppliedAt string
}

// MigrationRunner handles database migrations
type MigrationRunner struct {
	db             *sql.DB
	migrationFiles embed.FS
}

// NewMigrationRunner creates a new migration runner
func NewMigrationRunner(db *sql.DB, files embed.FS) *MigrationRunner {
	return &MigrationRunner{
		db:             db,
		migrationFiles: files,
	}
}

// RunMigrations applies all pending migrations
func (mr *MigrationRunner) RunMigrations() error {
	entries, err := mr.migrationFiles.ReadDir("sql/migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var upMigrations []string
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), "_up.sql") {
			upMigrations = append(upMigrations, entry.Name())
		}
	}
	sort.Strings(upMigrations)

	for _, migrationFile := range upMigrations {
		version := extractVersionFromFilename(migrationFile)

		var count int
		err := mr.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='migrations'").Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check migrations table: %w", err)
		}

		if count == 0 && version != "0000" {
			continue
		}

		if count > 0 {
			var applied int
			err = mr.db.QueryRow("SELECT COUNT(*) FROM migrations WHERE version = ?", version).Scan(&applied)
			if err != nil {
				return fmt.Errorf("failed to check migration %s: %w", version, err)
			}
			if applied > 0 {
				continue
			}
		}

		content, err := mr.migrationFiles.ReadFile("sql/migrations/" + migrationFile)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", migrationFile, err)
		}

		if _, err := mr.db.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", migrationFile, err)
		}

		if count > 0 || version == "0000" {
			if _, err := mr.db.Exec("INSERT INTO migrations (version) VALUES (?)", version); err != nil {
				return fmt.Errorf("failed to record migration %s: %w", version, err)
			}
		}
	}

	return nil
}

// GetAppliedMigrations returns a list of all applied migrations
func (mr *MigrationRunner) GetAppliedMigrations() ([]Migration, error) {
	var count int
	err := mr.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='migrations'").Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to check migrations table: %w", err)
	}

	if count == 0 {
		return []Migration{}, nil
	}

	rows, err := mr.db.Query("SELECT version, applied_at FROM migrations ORDER BY version")
	if err != nil {
		return nil, fmt.Errorf("failed to query migrations: %w", err)
	}
	defer rows.Close()

	var migrations []Migration
	for rows.Next() {
		var m Migration
		if err := rows.Scan(&m.Version, &m.AppliedAt); err != nil {
			return nil, fmt.Errorf("failed to scan migration: %w", err)
		}
		m.Applied = true
		migrations = append(migrations, m)
	}

	return migrations, nil
}

// GetAvailableMigrations returns all available migrations from embedded files
func (mr *MigrationRunner) GetAvailableMigrations() ([]Migration, error) {
	entries, err := mr.migrationFiles.ReadDir("sql/migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	migrationMap := make(map[string]*Migration)

	for _, entry := range entries {
		version := extractVersionFromFilename(entry.Name())
		if version == "" {
			continue
		}

		if migrationMap[version] == nil {
			migrationMap[version] = &Migration{
				Version: version,
				Name:    extractNameFromFilename(entry.Name()),
			}
		}

		content, err := mr.migrationFiles.ReadFile("sql/migrations/" + entry.Name())
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", entry.Name(), err)
		}

		if strings.HasSuffix(entry.Name(), "_up.sql") {
			migrationMap[version].UpSQL = string(content)
		} else if strings.HasSuffix(entry.Name(), "_down.sql") {
			migrationMap[version].DownSQL = string(content)
		}
	}

	var migrations []Migration
	for _, m := range migrationMap {
		migrations = append(migrations, *m)
	}
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// Rollback rolls back the last applied migration
func (mr *MigrationRunner) Rollback() error {
	var version string
	err := mr.db.QueryRow("SELECT version FROM migrations ORDER BY version DESC LIMIT 1").Scan(&version)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("no migrations to rollback")
		}
		return fmt.Errorf("failed to get last migration: %w", err)
	}

	entries, err := mr.migrationFiles.ReadDir("sql/migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var downContent []byte
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), version) && strings.HasSuffix(entry.Name(), "_down.sql") {
			downContent, err = mr.migrationFiles.ReadFile("sql/migrations/" + entry.Name())
			if err != nil {
				return fmt.Errorf("failed to read down migration: %w", err)
			}
			break
		}
	}

	if downContent == nil {
		return fmt.Errorf("down migration not found for version %s", version)
	}

	if _, err := mr.db.Exec(string(downContent)); err != nil {
		return fmt.Errorf("failed to execute down migration: %w", err)
	}

	if _, err := mr.db.Exec("DELETE FROM migrations WHERE version = ?", version); err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
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

func extractNameFromFilename(filename string) string {
	parts := strings.Split(filename, "_")
	if len(parts) < 3 {
		return ""
	}

	name := strings.Join(parts[1:len(parts)-1], "_")
	return strings.TrimSuffix(name, "_up")
}
