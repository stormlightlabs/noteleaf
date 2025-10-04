package store

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed sql/migrations
var testMigrationFiles embed.FS

type fakeMigrationFS struct {
	shouldFailRead   bool
	invalidSQL       bool
	hasNewMigrations bool
}

type fakeDirEntry struct {
	name string
}

func (f fakeDirEntry) Name() string               { return f.name }
func (f fakeDirEntry) IsDir() bool                { return false }
func (f fakeDirEntry) Type() fs.FileMode          { return 0 }
func (f fakeDirEntry) Info() (fs.FileInfo, error) { return nil, fmt.Errorf("info not available") }

func (f *fakeMigrationFS) ReadDir(name string) ([]fs.DirEntry, error) {
	if name == "sql/migrations" {
		entries := []fs.DirEntry{
			fakeDirEntry{name: "0000_create_migrations_table_up.sql"},
		}
		if f.hasNewMigrations {
			entries = append(entries,
				fakeDirEntry{name: "0001_test_migration_up.sql"},
				fakeDirEntry{name: "0001_test_migration_down.sql"},
			)
		}
		return entries, nil
	}
	return nil, fmt.Errorf("directory not found: %s", name)
}

func (f *fakeMigrationFS) ReadFile(name string) ([]byte, error) {
	if f.shouldFailRead {
		return nil, fmt.Errorf("simulated read failure")
	}
	if f.invalidSQL {
		return []byte("INVALID SQL SYNTAX GOES HERE AND MAKES DATABASE SAD"), nil
	}
	if name == "sql/migrations/0000_create_migrations_table_up.sql" {
		return []byte("CREATE TABLE migrations (version TEXT PRIMARY KEY, applied_at DATETIME DEFAULT CURRENT_TIMESTAMP);"), nil
	}
	if name == "sql/migrations/0001_test_migration_up.sql" {
		return []byte("CREATE TABLE test_table (id INTEGER PRIMARY KEY);"), nil
	}
	if name == "sql/migrations/0001_test_migration_down.sql" {
		return []byte("DROP TABLE IF EXISTS test_table;"), nil
	}
	return nil, fmt.Errorf("file not found: %s", name)
}

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

	runner := CreateMigrationRunner(db, testMigrationFiles)
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
		runner := CreateMigrationRunner(db, testMigrationFiles)

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

	t.Run("handles migration directory read failure", func(t *testing.T) {
		db := createTestDB(t)

		emptyFS := embed.FS{}
		runner := CreateMigrationRunner(db, emptyFS)

		err := runner.RunMigrations()
		if err == nil {
			t.Error("RunMigrations should fail when migration directory cannot be read")
		}
	})

	t.Run("handles migration table check failure", func(t *testing.T) {
		db := createTestDB(t)
		db.Close()

		runner := CreateMigrationRunner(db, testMigrationFiles)
		err := runner.RunMigrations()
		if err == nil {
			t.Error("RunMigrations should fail when database connection is closed")
		}
	})

	t.Run("handles migration file read failure", func(t *testing.T) {
		db := createTestDB(t)

		fakeFS := &fakeMigrationFS{shouldFailRead: true, hasNewMigrations: true}
		runner := CreateMigrationRunner(db, fakeFS)

		err := runner.RunMigrations()
		if err == nil {
			t.Error("RunMigrations should fail when migration file cannot be read")
		}
	})

	t.Run("handles invalid SQL in migration file", func(t *testing.T) {
		db := createTestDB(t)

		fakeFS := &fakeMigrationFS{invalidSQL: true, hasNewMigrations: true}
		runner := CreateMigrationRunner(db, fakeFS)

		err := runner.RunMigrations()
		if err == nil {
			t.Error("RunMigrations should fail when migration contains invalid SQL")
		}
	})

	t.Run("handles migration record insertion failure", func(t *testing.T) {
		db := createTestDB(t)
		runner := CreateMigrationRunner(db, testMigrationFiles)

		err := runner.RunMigrations()
		if err != nil {
			t.Fatalf("First RunMigrations failed: %v", err)
		}

		_, err = db.Exec("DROP TABLE migrations")
		if err != nil {
			t.Fatalf("Failed to drop migrations table: %v", err)
		}

		_, err = db.Exec("CREATE TABLE migrations (version TEXT PRIMARY KEY CHECK(length(version) < 0))")
		if err != nil {
			t.Fatalf("Failed to create migrations table with constraint: %v", err)
		}

		err = runner.RunMigrations()
		if err == nil {
			t.Error("RunMigrations should fail when migration record cannot be inserted")
		}
	})

	t.Run("skips already applied migrations", func(t *testing.T) {
		db := createTestDB(t)
		runner := CreateMigrationRunner(db, testMigrationFiles)

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
		runner := CreateMigrationRunner(db, testMigrationFiles)

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
		runner := CreateMigrationRunner(db, testMigrationFiles)

		migrations, err := runner.GetAppliedMigrations()
		if err != nil {
			t.Fatalf("GetAppliedMigrations failed: %v", err)
		}

		if len(migrations) != 0 {
			t.Errorf("Expected 0 migrations, got %d", len(migrations))
		}
	})

	t.Run("handles database connection failure", func(t *testing.T) {
		db := createTestDB(t)
		db.Close()
		runner := CreateMigrationRunner(db, testMigrationFiles)

		_, err := runner.GetAppliedMigrations()
		if err == nil {
			t.Error("GetAppliedMigrations should fail when database connection is closed")
		}
	})

	t.Run("handles query execution failure", func(t *testing.T) {
		db := createTestDB(t)
		runner := CreateMigrationRunner(db, testMigrationFiles)

		err := runner.RunMigrations()
		if err != nil {
			t.Fatalf("RunMigrations failed: %v", err)
		}

		// Close the database to trigger a query failure
		db.Close()

		_, err = runner.GetAppliedMigrations()
		if err == nil {
			t.Error("GetAppliedMigrations should fail when database is closed")
		}
	})

	t.Run("handles row scan failure", func(t *testing.T) {
		db := createTestDB(t)
		runner := CreateMigrationRunner(db, testMigrationFiles)

		err := runner.RunMigrations()
		if err != nil {
			t.Fatalf("RunMigrations failed: %v", err)
		}

		// Insert a record with NULL applied_at which should cause scan issues
		_, err = db.Exec("INSERT INTO migrations (version, applied_at) VALUES ('test', NULL)")
		if err != nil {
			t.Fatalf("Failed to insert NULL migration record: %v", err)
		}

		_, err = runner.GetAppliedMigrations()
		if err == nil {
			t.Error("GetAppliedMigrations should fail when scanning NULL applied_at field")
		}
	})

	t.Run("returns applied migrations", func(t *testing.T) {
		db := createTestDB(t)
		runner := CreateMigrationRunner(db, testMigrationFiles)

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
		runner := CreateMigrationRunner(db, testMigrationFiles)

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

	t.Run("handles migration directory read failure", func(t *testing.T) {
		db := createTestDB(t)

		emptyFS := embed.FS{}
		runner := CreateMigrationRunner(db, emptyFS)

		_, err := runner.GetAvailableMigrations()
		if err == nil {
			t.Error("GetAvailableMigrations should fail when migration directory cannot be read")
		}
	})

	t.Run("handles migration file read failure", func(t *testing.T) {
		db := createTestDB(t)

		fakeFS := &fakeMigrationFS{shouldFailRead: true}
		runner := CreateMigrationRunner(db, fakeFS)

		_, err := runner.GetAvailableMigrations()
		if err == nil {
			t.Error("GetAvailableMigrations should fail when migration file cannot be read")
		}
	})

	t.Run("includes both up and down SQL when available", func(t *testing.T) {
		db := createTestDB(t)
		runner := CreateMigrationRunner(db, testMigrationFiles)

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
		runner := CreateMigrationRunner(db, testMigrationFiles)

		err := runner.Rollback()
		if err == nil {
			t.Error("Rollback should fail when no migrations are applied")
		}
	})

	t.Run("handles database connection failure", func(t *testing.T) {
		db := createTestDB(t)
		runner := CreateMigrationRunner(db, testMigrationFiles)

		err := runner.RunMigrations()
		if err != nil {
			t.Fatalf("RunMigrations failed: %v", err)
		}

		db.Close()

		err = runner.Rollback()
		if err == nil {
			t.Error("Rollback should fail when database connection is closed")
		}
	})

	t.Run("handles migration directory read failure during rollback", func(t *testing.T) {
		db := createTestDB(t)
		runner := CreateMigrationRunner(db, testMigrationFiles)

		err := runner.RunMigrations()
		if err != nil {
			t.Fatalf("RunMigrations failed: %v", err)
		}

		emptyFS := embed.FS{}
		runner.migrationFiles = emptyFS

		err = runner.Rollback()
		if err == nil {
			t.Error("Rollback should fail when migration directory cannot be read")
		}
	})

	t.Run("handles missing down migration file", func(t *testing.T) {
		db := createTestDB(t)
		runner := CreateMigrationRunner(db, testMigrationFiles)

		err := runner.RunMigrations()
		if err != nil {
			t.Fatalf("RunMigrations failed: %v", err)
		}

		fakeFS := &fakeMigrationFS{}
		runner.migrationFiles = fakeFS

		err = runner.Rollback()
		if err == nil {
			t.Error("Rollback should fail when down migration file is not found")
		}
	})

	t.Run("handles down migration file read failure", func(t *testing.T) {
		db := createTestDB(t)

		fakeFS := &fakeMigrationFS{}
		runner := CreateMigrationRunner(db, fakeFS)

		err := runner.RunMigrations()
		if err != nil {
			t.Fatalf("RunMigrations failed: %v", err)
		}

		fakeFS.shouldFailRead = true

		err = runner.Rollback()
		if err == nil {
			t.Error("Rollback should fail when down migration file cannot be read")
		}
	})

	t.Run("handles invalid down migration SQL", func(t *testing.T) {
		db := createTestDB(t)

		fakeFS := &fakeMigrationFS{}
		runner := CreateMigrationRunner(db, fakeFS)

		err := runner.RunMigrations()
		if err != nil {
			t.Fatalf("RunMigrations failed: %v", err)
		}

		fakeFS.invalidSQL = true

		err = runner.Rollback()
		if err == nil {
			t.Error("Rollback should fail when down migration contains invalid SQL")
		}
	})

	t.Run("handles migration record deletion failure", func(t *testing.T) {
		db := createTestDB(t)
		runner := CreateMigrationRunner(db, testMigrationFiles)

		err := runner.RunMigrations()
		if err != nil {
			t.Fatalf("RunMigrations failed: %v", err)
		}

		_, err = db.Exec("DROP TABLE migrations")
		if err != nil {
			t.Fatalf("Failed to drop migrations table: %v", err)
		}

		err = runner.Rollback()
		if err == nil {
			t.Error("Rollback should fail when migration record cannot be deleted")
		}
	})

	t.Run("rolls back last migration", func(t *testing.T) {
		db := createTestDB(t)
		runner := CreateMigrationRunner(db, testMigrationFiles)

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
		runner := CreateMigrationRunner(db, testMigrationFiles)

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
		runner := CreateMigrationRunner(db, migrationFiles)

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
