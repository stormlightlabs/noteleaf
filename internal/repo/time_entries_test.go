package repo

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stormlightlabs/noteleaf/internal/models"
	"github.com/stormlightlabs/noteleaf/internal/store"
)

func setupTimeEntryTestDB(t *testing.T) (*sql.DB, *TimeEntryRepository, *TaskRepository, func()) {
	os.Setenv("NOTELEAF_CONFIG_DIR", t.TempDir())

	db, err := store.NewDatabase()
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	timeRepo := NewTimeEntryRepository(db.DB)
	taskRepo := NewTaskRepository(db.DB)

	cleanup := func() {
		db.Close()
		os.Unsetenv("NOTELEAF_CONFIG_DIR")
	}

	return db.DB, timeRepo, taskRepo, cleanup
}

func createTestTask(t *testing.T, taskRepo *TaskRepository) *models.Task {
	ctx := context.Background()
	task := &models.Task{
		UUID:        fmt.Sprintf("test-uuid-%d", time.Now().UnixNano()),
		Description: "Test Task",
		Status:      "pending",
	}

	id, err := taskRepo.Create(ctx, task)
	if err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}
	task.ID = id
	return task
}

func TestTimeEntryRepository_Start(t *testing.T) {
	_, repo, taskRepo, cleanup := setupTimeEntryTestDB(t)
	defer cleanup()

	ctx := context.Background()
	task := createTestTask(t, taskRepo)

	t.Run("starts time tracking successfully", func(t *testing.T) {
		description := "Working on feature"
		entry, err := repo.Start(ctx, task.ID, description)

		if err != nil {
			t.Fatalf("Failed to start time tracking: %v", err)
		}

		if entry.ID == 0 {
			t.Error("Expected entry to have an ID")
		}
		if entry.TaskID != task.ID {
			t.Errorf("Expected TaskID %d, got %d", task.ID, entry.TaskID)
		}
		if entry.Description != description {
			t.Errorf("Expected description %q, got %q", description, entry.Description)
		}
		if entry.EndTime != nil {
			t.Error("Expected EndTime to be nil for active entry")
		}
		if !entry.IsActive() {
			t.Error("Expected entry to be active")
		}
	})

	t.Run("prevents starting already active task", func(t *testing.T) {
		_, err := repo.Start(ctx, task.ID, "Another attempt")

		if err == nil {
			t.Error("Expected error when starting already active task")
		}
		if err.Error() != "task already has an active time entry" {
			t.Errorf("Expected specific error message, got: %v", err)
		}
	})
}

func TestTimeEntryRepository_Stop(t *testing.T) {
	_, repo, taskRepo, cleanup := setupTimeEntryTestDB(t)
	defer cleanup()

	ctx := context.Background()
	task := createTestTask(t, taskRepo)

	entry, err := repo.Start(ctx, task.ID, "Test work")
	if err != nil {
		t.Fatalf("Failed to start time tracking: %v", err)
	}

	time.Sleep(1010 * time.Millisecond)

	t.Run("stops active time entry", func(t *testing.T) {
		stoppedEntry, err := repo.Stop(ctx, entry.ID)

		if err != nil {
			t.Fatalf("Failed to stop time tracking: %v", err)
		}

		if stoppedEntry.EndTime == nil {
			t.Error("Expected EndTime to be set")
		}
		if stoppedEntry.DurationSeconds <= 0 {
			t.Error("Expected duration to be greater than 0")
		}
		if stoppedEntry.IsActive() {
			t.Error("Expected entry to not be active after stopping")
		}
	})

	t.Run("fails to stop already stopped entry", func(t *testing.T) {
		_, err := repo.Stop(ctx, entry.ID)

		if err == nil {
			t.Error("Expected error when stopping already stopped entry")
		}
		if err.Error() != "time entry is not active" {
			t.Errorf("Expected specific error message, got: %v", err)
		}
	})
}

func TestTimeEntryRepository_StopActiveByTaskID(t *testing.T) {
	_, repo, taskRepo, cleanup := setupTimeEntryTestDB(t)
	defer cleanup()

	ctx := context.Background()
	task := createTestTask(t, taskRepo)

	t.Run("stops active entry by task ID", func(t *testing.T) {
		_, err := repo.Start(ctx, task.ID, "Test work")
		if err != nil {
			t.Fatalf("Failed to start time tracking: %v", err)
		}

		stoppedEntry, err := repo.StopActiveByTaskID(ctx, task.ID)

		if err != nil {
			t.Fatalf("Failed to stop time tracking by task ID: %v", err)
		}

		if stoppedEntry.EndTime == nil {
			t.Error("Expected EndTime to be set")
		}
		if stoppedEntry.IsActive() {
			t.Error("Expected entry to not be active")
		}
	})

	t.Run("fails when no active entry exists", func(t *testing.T) {
		_, err := repo.StopActiveByTaskID(ctx, task.ID)

		if err == nil {
			t.Error("Expected error when no active entry exists")
		}
		if err.Error() != "no active time entry found for task" {
			t.Errorf("Expected specific error message, got: %v", err)
		}
	})
}

func TestTimeEntryRepository_GetActiveByTaskID(t *testing.T) {
	_, repo, taskRepo, cleanup := setupTimeEntryTestDB(t)
	defer cleanup()

	ctx := context.Background()
	task := createTestTask(t, taskRepo)

	t.Run("returns nil when no active entry exists", func(t *testing.T) {
		_, err := repo.GetActiveByTaskID(ctx, task.ID)

		if err != sql.ErrNoRows {
			t.Errorf("Expected sql.ErrNoRows, got: %v", err)
		}
	})

	t.Run("returns active entry when one exists", func(t *testing.T) {
		startedEntry, err := repo.Start(ctx, task.ID, "Test work")
		if err != nil {
			t.Fatalf("Failed to start time tracking: %v", err)
		}

		activeEntry, err := repo.GetActiveByTaskID(ctx, task.ID)

		if err != nil {
			t.Fatalf("Failed to get active entry: %v", err)
		}

		if activeEntry.ID != startedEntry.ID {
			t.Errorf("Expected entry ID %d, got %d", startedEntry.ID, activeEntry.ID)
		}
		if !activeEntry.IsActive() {
			t.Error("Expected entry to be active")
		}
	})
}

func TestTimeEntryRepository_GetByTaskID(t *testing.T) {
	_, repo, taskRepo, cleanup := setupTimeEntryTestDB(t)
	defer cleanup()

	ctx := context.Background()
	task := createTestTask(t, taskRepo)

	t.Run("returns empty slice when no entries exist", func(t *testing.T) {
		entries, err := repo.GetByTaskID(ctx, task.ID)

		if err != nil {
			t.Fatalf("Failed to get entries: %v", err)
		}

		if len(entries) != 0 {
			t.Errorf("Expected 0 entries, got %d", len(entries))
		}
	})

	t.Run("returns all entries for task", func(t *testing.T) {
		_, err := repo.Start(ctx, task.ID, "First session")
		if err != nil {
			t.Fatalf("Failed to start first session: %v", err)
		}

		_, err = repo.StopActiveByTaskID(ctx, task.ID)
		if err != nil {
			t.Fatalf("Failed to stop first session: %v", err)
		}

		_, err = repo.Start(ctx, task.ID, "Second session")
		if err != nil {
			t.Fatalf("Failed to start second session: %v", err)
		}

		entries, err := repo.GetByTaskID(ctx, task.ID)

		if err != nil {
			t.Fatalf("Failed to get entries: %v", err)
		}

		if len(entries) != 2 {
			t.Errorf("Expected 2 entries, got %d", len(entries))
		}

		if entries[0].Description != "Second session" {
			t.Errorf("Expected first entry to be 'Second session', got %q", entries[0].Description)
		}
		if entries[1].Description != "First session" {
			t.Errorf("Expected second entry to be 'First session', got %q", entries[1].Description)
		}
	})
}

func TestTimeEntryRepository_GetTotalTimeByTaskID(t *testing.T) {
	_, repo, taskRepo, cleanup := setupTimeEntryTestDB(t)
	defer cleanup()

	ctx := context.Background()
	task := createTestTask(t, taskRepo)

	t.Run("returns zero duration when no entries exist", func(t *testing.T) {
		duration, err := repo.GetTotalTimeByTaskID(ctx, task.ID)

		if err != nil {
			t.Fatalf("Failed to get total time: %v", err)
		}

		if duration != 0 {
			t.Errorf("Expected 0 duration, got %v", duration)
		}
	})

	t.Run("calculates total time including active entries", func(t *testing.T) {
		entry1, err := repo.Start(ctx, task.ID, "Completed work")
		if err != nil {
			t.Fatalf("Failed to start first entry: %v", err)
		}

		time.Sleep(1010 * time.Millisecond)
		_, err = repo.Stop(ctx, entry1.ID)
		if err != nil {
			t.Fatalf("Failed to stop first entry: %v", err)
		}

		_, err = repo.Start(ctx, task.ID, "Active work")
		if err != nil {
			t.Fatalf("Failed to start second entry: %v", err)
		}

		time.Sleep(1010 * time.Millisecond)

		totalTime, err := repo.GetTotalTimeByTaskID(ctx, task.ID)

		if err != nil {
			t.Fatalf("Failed to get total time: %v", err)
		}

		if totalTime <= 0 {
			t.Error("Expected total time to be greater than 0")
		}

		if totalTime < 2*time.Second {
			t.Errorf("Expected total time to be at least 2s, got %v", totalTime)
		}
	})
}

func TestTimeEntryRepository_Delete(t *testing.T) {
	_, repo, taskRepo, cleanup := setupTimeEntryTestDB(t)
	defer cleanup()

	ctx := context.Background()
	task := createTestTask(t, taskRepo)

	t.Run("deletes existing entry", func(t *testing.T) {
		entry, err := repo.Start(ctx, task.ID, "To be deleted")
		if err != nil {
			t.Fatalf("Failed to create entry: %v", err)
		}

		err = repo.Delete(ctx, entry.ID)

		if err != nil {
			t.Fatalf("Failed to delete entry: %v", err)
		}

		_, err = repo.Get(ctx, entry.ID)
		if err != sql.ErrNoRows {
			t.Errorf("Expected entry to be deleted, but got: %v", err)
		}
	})

	t.Run("fails to delete non-existent entry", func(t *testing.T) {
		err := repo.Delete(ctx, 99999)

		if err == nil {
			t.Error("Expected error when deleting non-existent entry")
		}
		if err.Error() != "time entry not found" {
			t.Errorf("Expected specific error message, got: %v", err)
		}
	})
}
