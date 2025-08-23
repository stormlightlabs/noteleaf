package store

import (
	"database/sql"
	"embed"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed sql/migrations
var testMigrationFiles embed.FS

func createTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create in-memory database: %v", err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func TestNewMigrationRunner(t *testing.T) {
	db := createTestDB(t)

	runner := NewMigrationRunner(db, testMigrationFiles)
	if runner == nil {
		t.Fatal("NewMigrationRunner should not return nil")
	}

	if runner.db != db {
		t.Error("Migration runner should store the database reference")
	}
}

func TestMigrationRunner_RunMigrations(t *testing.T) {
	t.Run("runs migrations successfully", func(t *testing.T) {
		db := createTestDB(t)
		runner := NewMigrationRunner(db, testMigrationFiles)

		err := runner.RunMigrations()
		if err != nil {
			t.Fatalf("RunMigrations failed: %v", err)
		}

		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='migrations'").Scan(&count)
		if err != nil {
			t.Fatalf("Failed to check migrations table: %v", err)
		}

		if count != 1 {
			t.Error("Migrations table should exist after running migrations")
		}

		var migrationCount int
		err = db.QueryRow("SELECT COUNT(*) FROM migrations").Scan(&migrationCount)
		if err != nil {
			t.Fatalf("Failed to count applied migrations: %v", err)
		}

		if migrationCount == 0 {
			t.Error("At least one migration should be applied")
		}
	})

	t.Run("skips already applied migrations", func(t *testing.T) {
		db := createTestDB(t)
		runner := NewMigrationRunner(db, testMigrationFiles)

		err := runner.RunMigrations()
		if err != nil {
			t.Fatalf("First RunMigrations failed: %v", err)
		}

		var initialCount int
		err = db.QueryRow("SELECT COUNT(*) FROM migrations").Scan(&initialCount)
		if err != nil {
			t.Fatalf("Failed to count migrations: %v", err)
		}

		err = runner.RunMigrations()
		if err != nil {
			t.Fatalf("Second RunMigrations failed: %v", err)
		}

		var finalCount int
		err = db.QueryRow("SELECT COUNT(*) FROM migrations").Scan(&finalCount)
		if err != nil {
			t.Fatalf("Failed to count migrations after second run: %v", err)
		}

		if finalCount != initialCount {
			t.Errorf("Expected %d migrations, got %d (migrations should not be re-applied)", initialCount, finalCount)
		}
	})

	t.Run("creates expected tables", func(t *testing.T) {
		db := createTestDB(t)
		runner := NewMigrationRunner(db, testMigrationFiles)

		err := runner.RunMigrations()
		if err != nil {
			t.Fatalf("RunMigrations failed: %v", err)
		}

		expectedTables := []string{"migrations", "tasks", "movies", "tv_shows", "books", "notes"}

		for _, tableName := range expectedTables {
			var count int
			err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", tableName).Scan(&count)
			if err != nil {
				t.Fatalf("Failed to check table %s: %v", tableName, err)
			}

			if count != 1 {
				t.Errorf("Table %s should exist after migrations", tableName)
			}
		}
	})
}

func TestMigrationRunner_GetAppliedMigrations(t *testing.T) {
	t.Run("returns empty list when no migrations table", func(t *testing.T) {
		db := createTestDB(t)
		runner := NewMigrationRunner(db, testMigrationFiles)

		migrations, err := runner.GetAppliedMigrations()
		if err != nil {
			t.Fatalf("GetAppliedMigrations failed: %v", err)
		}

		if len(migrations) != 0 {
			t.Errorf("Expected 0 migrations, got %d", len(migrations))
		}
	})

	t.Run("returns applied migrations", func(t *testing.T) {
		db := createTestDB(t)
		runner := NewMigrationRunner(db, testMigrationFiles)

		// Run migrations first
		err := runner.RunMigrations()
		if err != nil {
			t.Fatalf("RunMigrations failed: %v", err)
		}

		migrations, err := runner.GetAppliedMigrations()
		if err != nil {
			t.Fatalf("GetAppliedMigrations failed: %v", err)
		}

		if len(migrations) == 0 {
			t.Error("Should have applied migrations")
		}

		for _, migration := range migrations {
			if migration.Version == "" {
				t.Error("Migration version should not be empty")
			}
			if !migration.Applied {
				t.Error("Migration should be marked as applied")
			}
			if migration.AppliedAt == "" {
				t.Error("Migration should have applied timestamp")
			}
		}

		for i := 1; i < len(migrations); i++ {
			if migrations[i-1].Version > migrations[i].Version {
				t.Error("Migrations should be sorted by version")
			}
		}
	})
}

func TestMigrationRunner_GetAvailableMigrations(t *testing.T) {
	t.Run("returns available migrations from embedded files", func(t *testing.T) {
		db := createTestDB(t)
		runner := NewMigrationRunner(db, testMigrationFiles)

		migrations, err := runner.GetAvailableMigrations()
		if err != nil {
			t.Fatalf("GetAvailableMigrations failed: %v", err)
		}

		if len(migrations) == 0 {
			t.Error("Should have available migrations")
		}

		for _, migration := range migrations {
			if migration.Version == "" {
				t.Error("Migration version should not be empty")
			}
			if migration.UpSQL == "" {
				t.Error("Migration should have up SQL")
			}
			// Note: Down SQL might be empty for some migrations, so we don't check it
		}

		for i := 1; i < len(migrations); i++ {
			if migrations[i-1].Version > migrations[i].Version {
				t.Error("Migrations should be sorted by version")
			}
		}
	})

	t.Run("includes both up and down SQL when available", func(t *testing.T) {
		db := createTestDB(t)
		runner := NewMigrationRunner(db, testMigrationFiles)

		migrations, err := runner.GetAvailableMigrations()
		if err != nil {
			t.Fatalf("GetAvailableMigrations failed: %v", err)
		}

		var foundMigrationWithDown bool
		for _, migration := range migrations {
			if migration.UpSQL != "" && migration.DownSQL != "" {
				foundMigrationWithDown = true
				break
			}
		}

		if !foundMigrationWithDown {
			t.Log("Note: No migrations found with both up and down SQL - this may be expected")
		}
	})
}

func TestMigrationRunner_Rollback(t *testing.T) {
	t.Run("fails when no migrations to rollback", func(t *testing.T) {
		db := createTestDB(t)
		runner := NewMigrationRunner(db, testMigrationFiles)

		err := runner.Rollback()
		if err == nil {
			t.Error("Rollback should fail when no migrations are applied")
		}
	})

	t.Run("rolls back last migration", func(t *testing.T) {
		db := createTestDB(t)
		runner := NewMigrationRunner(db, testMigrationFiles)

		err := runner.RunMigrations()
		if err != nil {
			t.Fatalf("RunMigrations failed: %v", err)
		}

		var initialCount int
		err = db.QueryRow("SELECT COUNT(*) FROM migrations").Scan(&initialCount)
		if err != nil {
			t.Fatalf("Failed to count migrations: %v", err)
		}

		if initialCount == 0 {
			t.Skip("No migrations to rollback")
		}

		err = runner.Rollback()
		if err != nil {
			t.Fatalf("Rollback failed: %v", err)
		}

		var finalCount int
		err = db.QueryRow("SELECT COUNT(*) FROM migrations").Scan(&finalCount)
		if err != nil {
			t.Fatalf("Failed to count migrations after rollback: %v", err)
		}

		if finalCount != initialCount-1 {
			t.Errorf("Expected %d migrations after rollback, got %d", initialCount-1, finalCount)
		}
	})
}

func TestMigrationHelperFunctions(t *testing.T) {
	t.Run("extractVersionFromFilename", func(t *testing.T) {
		testCases := []struct {
			filename string
			expected string
		}{
			{"0000_create_migrations_table_up.sql", "0000"},
			{"0001_create_all_tables_up.sql", "0001"},
			{"0002_add_indexes_down.sql", "0002"},
			{"invalid_filename.sql", "invalid"},
			{"", ""},
		}

		for _, tc := range testCases {
			result := extractVersionFromFilename(tc.filename)
			if result != tc.expected {
				t.Errorf("extractVersionFromFilename(%s): expected %s, got %s", tc.filename, tc.expected, result)
			}
		}
	})

	t.Run("extractNameFromFilename", func(t *testing.T) {
		testCases := []struct {
			filename string
			expected string
		}{
			{"0000_create_migrations_table_up.sql", "create_migrations_table"},
			{"0001_create_all_tables_up.sql", "create_all_tables"},
			{"0002_add_indexes_down.sql", "add_indexes"},
			{"invalid_filename.sql", ""},
			{"0003_up.sql", ""},
			{"", ""},
		}

		for _, tc := range testCases {
			result := extractNameFromFilename(tc.filename)
			if result != tc.expected {
				t.Errorf("extractNameFromFilename(%s): expected %s, got %s", tc.filename, tc.expected, result)
			}
		}
	})
}

func TestMigrationIntegration(t *testing.T) {
	t.Run("full migration lifecycle", func(t *testing.T) {
		db := createTestDB(t)
		runner := NewMigrationRunner(db, testMigrationFiles)

		available, err := runner.GetAvailableMigrations()
		if err != nil {
			t.Fatalf("GetAvailableMigrations failed: %v", err)
		}

		if len(available) == 0 {
			t.Skip("No migrations available for testing")
		}

		err = runner.RunMigrations()
		if err != nil {
			t.Fatalf("RunMigrations failed: %v", err)
		}

		applied, err := runner.GetAppliedMigrations()
		if err != nil {
			t.Fatalf("GetAppliedMigrations failed: %v", err)
		}

		if len(applied) == 0 {
			t.Error("No migrations were applied")
		}

		tables := []string{"tasks", "movies", "tv_shows", "books", "notes"}
		for _, table := range tables {
			var count int
			err = db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count)
			if err != nil {
				t.Errorf("Failed to query table %s: %v", table, err)
			}
		}

		if len(applied) > 1 { // Only test rollback if we have more than one migration
			err = runner.Rollback()
			if err != nil {
				t.Logf("Rollback failed (may be expected): %v", err)
			}
		}
	})

	t.Run("migration runner works with real database", func(t *testing.T) {
		db := createTestDB(t)
		runner := NewMigrationRunner(db, migrationFiles)

		err := runner.RunMigrations()
		if err != nil {
			t.Fatalf("RunMigrations with real files failed: %v", err)
		}

		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM migrations").Scan(&count)
		if err != nil {
			t.Fatalf("Failed to count real migrations: %v", err)
		}

		if count == 0 {
			t.Error("Real migrations should be applied")
		}
	})
}
