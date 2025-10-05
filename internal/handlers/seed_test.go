package handlers

import (
	"context"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/stormlightlabs/noteleaf/internal/store"
)

// setupSeedTest removed - use NewHandlerTestSuite(t) instead

func countRecords(t *testing.T, db *store.Database, table string) int {
	t.Helper()

	var count int
	query := "SELECT COUNT(*) FROM " + table
	err := db.QueryRow(query).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count records in %s: %v", table, err)
	}
	return count
}

func getTaskRecord(t *testing.T, db *store.Database, id int) (description, project, priority, status string) {
	t.Helper()

	query := "SELECT description, project, priority, status FROM tasks WHERE id = ?"
	err := db.QueryRow(query, id).Scan(&description, &project, &priority, &status)
	if err != nil {
		t.Fatalf("Failed to get task record: %v", err)
	}
	return description, project, priority, status
}

func getBookRecord(t *testing.T, db *store.Database, id int) (title, author, status string, progress int) {
	t.Helper()

	query := "SELECT title, author, status, progress FROM books WHERE id = ?"
	err := db.QueryRow(query, id).Scan(&title, &author, &status, &progress)
	if err != nil {
		t.Fatalf("Failed to get book record: %v", err)
	}
	return title, author, status, progress
}

func TestSeedHandler(t *testing.T) {
	_ = NewHandlerTestSuite(t)

	handler, err := NewSeedHandler()
	if err != nil {
		t.Fatalf("Failed to create seed handler: %v", err)
	}
	defer handler.Close()

	ctx := context.Background()

	t.Run("New", func(t *testing.T) {
		t.Run("creates handler successfully", func(t *testing.T) {
			testHandler, err := NewSeedHandler()
			if err != nil {
				t.Fatalf("NewSeedHandler failed: %v", err)
			}
			if testHandler == nil {
				t.Fatal("Handler should not be nil")
			}
			defer testHandler.Close()

			if testHandler.db == nil {
				t.Error("Handler database should not be nil")
			}
			if testHandler.config == nil {
				t.Error("Handler config should not be nil")
			}
			if testHandler.repos == nil {
				t.Error("Handler repos should not be nil")
			}
		})

		t.Run("handles database initialization error", func(t *testing.T) {
			originalXDG := os.Getenv("XDG_CONFIG_HOME")
			originalHome := os.Getenv("HOME")
			originalNoteleafConfig := os.Getenv("NOTELEAF_CONFIG")
			originalNoteleafDataDir := os.Getenv("NOTELEAF_DATA_DIR")

			if runtime.GOOS == "windows" {
				originalAppData := os.Getenv("APPDATA")
				os.Unsetenv("APPDATA")
				defer os.Setenv("APPDATA", originalAppData)
			} else {
				os.Unsetenv("XDG_CONFIG_HOME")
				os.Unsetenv("HOME")
			}
			os.Unsetenv("NOTELEAF_CONFIG")
			os.Unsetenv("NOTELEAF_DATA_DIR")

			defer func() {
				os.Setenv("XDG_CONFIG_HOME", originalXDG)
				os.Setenv("HOME", originalHome)
				os.Setenv("NOTELEAF_CONFIG", originalNoteleafConfig)
				os.Setenv("NOTELEAF_DATA_DIR", originalNoteleafDataDir)
			}()

			_, err := NewSeedHandler()
			if err == nil {
				t.Error("NewSeedHandler should fail when database initialization fails")
			}
			if !strings.Contains(err.Error(), "failed to initialize database") {
				t.Errorf("Expected database error, got: %v", err)
			}
		})
	})

	t.Run("Seed", func(t *testing.T) {
		t.Run("seeds database with test data", func(t *testing.T) {
			err := handler.Seed(ctx, false)
			if err != nil {
				t.Fatalf("Seed failed: %v", err)
			}

			taskCount := countRecords(t, handler.db, "tasks")
			if taskCount != 5 {
				t.Errorf("Expected 5 tasks, got %d", taskCount)
			}

			bookCount := countRecords(t, handler.db, "books")
			if bookCount != 5 {
				t.Errorf("Expected 5 books, got %d", bookCount)
			}

			desc, proj, prio, status := getTaskRecord(t, handler.db, 1)
			if desc != "Review quarterly report" {
				t.Errorf("Expected 'Review quarterly report', got '%s'", desc)
			}
			if proj != "work" {
				t.Errorf("Expected 'work' project, got '%s'", proj)
			}
			if prio != "high" {
				t.Errorf("Expected 'high' priority, got '%s'", prio)
			}
			if status != "pending" {
				t.Errorf("Expected 'pending' status, got '%s'", status)
			}

			title, author, bookStatus, progress := getBookRecord(t, handler.db, 1)
			if title != "The Go Programming Language" {
				t.Errorf("Expected 'The Go Programming Language', got '%s'", title)
			}
			if author != "Alan Donovan" {
				t.Errorf("Expected 'Alan Donovan', got '%s'", author)
			}
			if bookStatus != "reading" {
				t.Errorf("Expected 'reading' status, got '%s'", bookStatus)
			}
			if progress != 45 {
				t.Errorf("Expected 45%% progress, got %d", progress)
			}
		})

		t.Run("seeds without force flag preserves existing data", func(t *testing.T) {
			err := handler.Seed(ctx, false)
			if err != nil {
				t.Fatalf("First seed failed: %v", err)
			}

			initialTaskCount := countRecords(t, handler.db, "tasks")
			initialBookCount := countRecords(t, handler.db, "books")

			err = handler.Seed(ctx, false)
			if err != nil {
				t.Fatalf("Second seed failed: %v", err)
			}

			finalTaskCount := countRecords(t, handler.db, "tasks")
			finalBookCount := countRecords(t, handler.db, "books")

			expectedTasks := initialTaskCount + 5
			expectedBooks := initialBookCount + 5

			if finalTaskCount != expectedTasks {
				t.Errorf("Expected %d tasks after second seed, got %d", expectedTasks, finalTaskCount)
			}
			if finalBookCount != expectedBooks {
				t.Errorf("Expected %d books after second seed, got %d", expectedBooks, finalBookCount)
			}
		})

		t.Run("force flag clears existing data before seeding", func(t *testing.T) {
			err := handler.Seed(ctx, false)
			if err != nil {
				t.Fatalf("Initial seed failed: %v", err)
			}

			if countRecords(t, handler.db, "tasks") == 0 {
				t.Fatal("No tasks found after initial seed")
			}
			if countRecords(t, handler.db, "books") == 0 {
				t.Fatal("No books found after initial seed")
			}

			err = handler.Seed(ctx, true)
			if err != nil {
				t.Fatalf("Force seed failed: %v", err)
			}

			taskCount := countRecords(t, handler.db, "tasks")
			bookCount := countRecords(t, handler.db, "books")

			if taskCount != 5 {
				t.Errorf("Expected exactly 5 tasks after force seed, got %d", taskCount)
			}
			if bookCount != 5 {
				t.Errorf("Expected exactly 5 books after force seed, got %d", bookCount)
			}

			_, _, _, _ = getTaskRecord(t, handler.db, 1) // Should not error
			_, _, _, _ = getBookRecord(t, handler.db, 1) // Should not error
		})
	})

	t.Run("Close", func(t *testing.T) {
		testHandler, err := NewSeedHandler()
		if err != nil {
			t.Fatalf("Failed to create test handler: %v", err)
		}

		if err = testHandler.Close(); err != nil {
			t.Errorf("Close should succeed: %v", err)
		}
	})
}
