package repo

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stormlightlabs/noteleaf/internal/models"
	"github.com/stormlightlabs/noteleaf/internal/shared"
)

func createTestTask(t *testing.T, db *sql.DB) *models.Task {
	t.Helper()
	ctx := context.Background()
	taskRepo := NewTaskRepository(db)
	task := &models.Task{
		UUID:        fmt.Sprintf("test-uuid-%d", time.Now().UnixNano()),
		Description: "Test Task",
		Status:      "pending",
	}

	id, err := taskRepo.Create(ctx, task)
	shared.AssertNoError(t, err, "Failed to create test task")
	task.ID = id
	return task
}

func TestTimeEntryRepository(t *testing.T) {
	t.Run("CRUD Operations", func(t *testing.T) {
		db := CreateTestDB(t)
		repo := NewTimeEntryRepository(db)
		ctx := context.Background()
		task := createTestTask(t, db)

		t.Run("Start time tracking", func(t *testing.T) {
			description := "Working on feature"
			entry, err := repo.Start(ctx, task.ID, description)

			shared.AssertNoError(t, err, "Failed to start time tracking")
			shared.AssertNotEqual(t, int64(0), entry.ID, "Expected non-zero entry ID")
			shared.AssertEqual(t, task.ID, entry.TaskID, "Expected TaskID to match")
			shared.AssertEqual(t, description, entry.Description, "Expected description to match")
			shared.AssertTrue(t, entry.EndTime == nil, "Expected EndTime to be nil for active entry")
			shared.AssertTrue(t, entry.IsActive(), "Expected entry to be active")
		})

		t.Run("Prevent starting already active task", func(t *testing.T) {
			_, err := repo.Start(ctx, task.ID, "Another attempt")

			shared.AssertError(t, err, "Expected error when starting already active task")
			shared.AssertContains(t, err.Error(), "task already has an active time entry", "Expected specific error message")
		})

		t.Run("Stop active time entry", func(t *testing.T) {
			db := CreateTestDB(t)
			repo := NewTimeEntryRepository(db)
			task := createTestTask(t, db)

			entry, err := repo.Start(ctx, task.ID, "Test work")
			shared.AssertNoError(t, err, "Failed to start time tracking")

			time.Sleep(1010 * time.Millisecond)

			stoppedEntry, err := repo.Stop(ctx, entry.ID)
			shared.AssertNoError(t, err, "Failed to stop time tracking")
			shared.AssertTrue(t, stoppedEntry.EndTime != nil, "Expected EndTime to be set")
			shared.AssertGreaterThan(t, stoppedEntry.DurationSeconds, int64(0), "Expected duration > 0")
			shared.AssertFalse(t, stoppedEntry.IsActive(), "Expected entry to not be active after stopping")
		})

		t.Run("Fail to stop already stopped entry", func(t *testing.T) {
			db := CreateTestDB(t)
			repo := NewTimeEntryRepository(db)
			task := createTestTask(t, db)

			entry, err := repo.Start(ctx, task.ID, "Test work")
			shared.AssertNoError(t, err, "Failed to start time tracking")

			time.Sleep(1010 * time.Millisecond)
			_, err = repo.Stop(ctx, entry.ID)
			shared.AssertNoError(t, err, "Failed to stop time tracking")

			_, err = repo.Stop(ctx, entry.ID)
			shared.AssertError(t, err, "Expected error when stopping already stopped entry")
			shared.AssertContains(t, err.Error(), "time entry is not active", "Expected specific error message")
		})

		t.Run("Get time entry", func(t *testing.T) {
			db := CreateTestDB(t)
			repo := NewTimeEntryRepository(db)
			task := createTestTask(t, db)

			original, err := repo.Start(ctx, task.ID, "Test entry")
			shared.AssertNoError(t, err, "Failed to start time tracking")

			retrieved, err := repo.Get(ctx, original.ID)
			shared.AssertNoError(t, err, "Failed to get time entry")
			shared.AssertEqual(t, original.ID, retrieved.ID, "ID mismatch")
			shared.AssertEqual(t, original.TaskID, retrieved.TaskID, "TaskID mismatch")
			shared.AssertEqual(t, original.Description, retrieved.Description, "Description mismatch")
		})

		t.Run("Delete time entry", func(t *testing.T) {
			db := CreateTestDB(t)
			repo := NewTimeEntryRepository(db)
			task := createTestTask(t, db)

			entry, err := repo.Start(ctx, task.ID, "To be deleted")
			shared.AssertNoError(t, err, "Failed to create entry")

			err = repo.Delete(ctx, entry.ID)
			shared.AssertNoError(t, err, "Failed to delete entry")

			_, err = repo.Get(ctx, entry.ID)
			shared.AssertError(t, err, "Expected error when getting deleted entry")
			shared.AssertEqual(t, sql.ErrNoRows, err, "Expected sql.ErrNoRows")
		})
	})

	t.Run("Query Methods", func(t *testing.T) {
		db := CreateTestDB(t)
		repo := NewTimeEntryRepository(db)
		ctx := context.Background()
		task := createTestTask(t, db)

		t.Run("GetActiveByTaskID returns error when no active entry", func(t *testing.T) {
			_, err := repo.GetActiveByTaskID(ctx, task.ID)
			shared.AssertError(t, err, "Expected error when no active entry exists")
			shared.AssertEqual(t, sql.ErrNoRows, err, "Expected sql.ErrNoRows")
		})

		t.Run("GetActiveByTaskID returns active entry", func(t *testing.T) {
			startedEntry, err := repo.Start(ctx, task.ID, "Test work")
			shared.AssertNoError(t, err, "Failed to start time tracking")

			activeEntry, err := repo.GetActiveByTaskID(ctx, task.ID)
			shared.AssertNoError(t, err, "Failed to get active entry")
			shared.AssertEqual(t, startedEntry.ID, activeEntry.ID, "Expected entry IDs to match")
			shared.AssertTrue(t, activeEntry.IsActive(), "Expected entry to be active")
		})

		t.Run("StopActiveByTaskID stops active entry", func(t *testing.T) {
			db := CreateTestDB(t)
			repo := NewTimeEntryRepository(db)
			task := createTestTask(t, db)

			_, err := repo.Start(ctx, task.ID, "Test work")
			shared.AssertNoError(t, err, "Failed to start time tracking")

			stoppedEntry, err := repo.StopActiveByTaskID(ctx, task.ID)
			shared.AssertNoError(t, err, "Failed to stop time tracking by task ID")
			shared.AssertTrue(t, stoppedEntry.EndTime != nil, "Expected EndTime to be set")
			shared.AssertFalse(t, stoppedEntry.IsActive(), "Expected entry to not be active")
		})

		t.Run("StopActiveByTaskID fails when no active entry", func(t *testing.T) {
			db := CreateTestDB(t)
			repo := NewTimeEntryRepository(db)
			task := createTestTask(t, db)

			_, err := repo.StopActiveByTaskID(ctx, task.ID)
			shared.AssertError(t, err, "Expected error when no active entry exists")
			shared.AssertContains(t, err.Error(), "no active time entry found for task", "Expected specific error message")
		})

		t.Run("GetByTaskID returns empty when no entries", func(t *testing.T) {
			db := CreateTestDB(t)
			repo := NewTimeEntryRepository(db)
			task := createTestTask(t, db)

			entries, err := repo.GetByTaskID(ctx, task.ID)
			shared.AssertNoError(t, err, "Failed to get entries")
			shared.AssertEqual(t, 0, len(entries), "Expected 0 entries")
		})

		t.Run("GetByTaskID returns all entries for task", func(t *testing.T) {
			db := CreateTestDB(t)
			repo := NewTimeEntryRepository(db)
			task := createTestTask(t, db)

			_, err := repo.Start(ctx, task.ID, "First session")
			shared.AssertNoError(t, err, "Failed to start first session")

			_, err = repo.StopActiveByTaskID(ctx, task.ID)
			shared.AssertNoError(t, err, "Failed to stop first session")

			_, err = repo.Start(ctx, task.ID, "Second session")
			shared.AssertNoError(t, err, "Failed to start second session")

			entries, err := repo.GetByTaskID(ctx, task.ID)
			shared.AssertNoError(t, err, "Failed to get entries")
			shared.AssertEqual(t, 2, len(entries), "Expected 2 entries")
			shared.AssertEqual(t, "Second session", entries[0].Description, "Expected newest entry first")
			shared.AssertEqual(t, "First session", entries[1].Description, "Expected oldest entry second")
		})

		t.Run("GetTotalTimeByTaskID returns zero when no entries", func(t *testing.T) {
			db := CreateTestDB(t)
			repo := NewTimeEntryRepository(db)
			task := createTestTask(t, db)

			duration, err := repo.GetTotalTimeByTaskID(ctx, task.ID)
			shared.AssertNoError(t, err, "Failed to get total time")
			shared.AssertEqual(t, time.Duration(0), duration, "Expected 0 duration")
		})

		t.Run("GetTotalTimeByTaskID calculates total including active entries", func(t *testing.T) {
			db := CreateTestDB(t)
			repo := NewTimeEntryRepository(db)
			task := createTestTask(t, db)

			entry1, err := repo.Start(ctx, task.ID, "Completed work")
			shared.AssertNoError(t, err, "Failed to start first entry")

			time.Sleep(1010 * time.Millisecond)
			_, err = repo.Stop(ctx, entry1.ID)
			shared.AssertNoError(t, err, "Failed to stop first entry")

			_, err = repo.Start(ctx, task.ID, "Active work")
			shared.AssertNoError(t, err, "Failed to start second entry")

			time.Sleep(1010 * time.Millisecond)

			totalTime, err := repo.GetTotalTimeByTaskID(ctx, task.ID)
			shared.AssertNoError(t, err, "Failed to get total time")
			shared.AssertTrue(t, totalTime > 0, "Expected total time > 0")
			shared.AssertTrue(t, totalTime >= 2*time.Second, "Expected total time >= 2s")
		})
	})

	t.Run("GetByDateRange", func(t *testing.T) {
		db := CreateTestDB(t)
		repo := NewTimeEntryRepository(db)
		ctx := context.Background()

		t.Run("Returns empty when no entries in range", func(t *testing.T) {
			start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
			end := time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)

			entries, err := repo.GetByDateRange(ctx, start, end)
			shared.AssertNoError(t, err, "Failed to get entries by date range")
			shared.AssertEqual(t, 0, len(entries), "Expected 0 entries")
		})

		t.Run("Returns entries within date range", func(t *testing.T) {
			task := createTestTask(t, db)

			entry, err := repo.Start(ctx, task.ID, "Test entry")
			shared.AssertNoError(t, err, "Failed to start entry")

			_, err = repo.Stop(ctx, entry.ID)
			shared.AssertNoError(t, err, "Failed to stop entry")

			now := time.Now()
			start := now.Add(-time.Hour)
			end := now.Add(time.Hour)

			entries, err := repo.GetByDateRange(ctx, start, end)
			shared.AssertNoError(t, err, "Failed to get entries by date range")

			found := false
			for _, e := range entries {
				if e.Description == "Test entry" {
					found = true
					break
				}
			}
			shared.AssertTrue(t, found, "Expected to find 'Test entry' in results")
		})

		t.Run("Respects date range boundaries", func(t *testing.T) {
			task := createTestTask(t, db)

			entry, err := repo.Start(ctx, task.ID, "Boundary test")
			shared.AssertNoError(t, err, "Failed to start entry")

			_, err = repo.Stop(ctx, entry.ID)
			shared.AssertNoError(t, err, "Failed to stop entry")

			start := time.Now().Add(time.Hour)
			end := time.Now().Add(2 * time.Hour)

			entries, err := repo.GetByDateRange(ctx, start, end)
			shared.AssertNoError(t, err, "Failed to get entries by date range")

			for _, e := range entries {
				if e.Description == "Boundary test" {
					t.Error("Should not find 'Boundary test' in future date range")
				}
			}
		})

		t.Run("Handles invalid date range", func(t *testing.T) {
			start := time.Now()
			end := time.Now().AddDate(0, 0, -1)

			entries, err := repo.GetByDateRange(ctx, start, end)
			shared.AssertNoError(t, err, "Should not error with invalid date range")
			shared.AssertEqual(t, 0, len(entries), "Expected 0 entries with invalid range")
		})
	})

	t.Run("Context Cancellation Error Paths", func(t *testing.T) {
		db := CreateTestDB(t)
		repo := NewTimeEntryRepository(db)
		ctx := context.Background()
		task := createTestTask(t, db)

		entry, err := repo.Start(ctx, task.ID, "Test entry")
		shared.AssertNoError(t, err, "Failed to create entry")

		t.Run("Start with cancelled context", func(t *testing.T) {
			db := CreateTestDB(t)
			repo := NewTimeEntryRepository(db)
			task := createTestTask(t, db)

			_, err := repo.Start(NewCanceledContext(), task.ID, "Cancelled")
			AssertCancelledContext(t, err)
		})

		t.Run("Get with cancelled context", func(t *testing.T) {
			_, err := repo.Get(NewCanceledContext(), entry.ID)
			AssertCancelledContext(t, err)
		})

		t.Run("Stop with cancelled context", func(t *testing.T) {
			_, err := repo.Stop(NewCanceledContext(), entry.ID)
			AssertCancelledContext(t, err)
		})

		t.Run("GetActiveByTaskID with cancelled context", func(t *testing.T) {
			_, err := repo.GetActiveByTaskID(NewCanceledContext(), task.ID)
			AssertCancelledContext(t, err)
		})

		t.Run("StopActiveByTaskID with cancelled context", func(t *testing.T) {
			_, err := repo.StopActiveByTaskID(NewCanceledContext(), task.ID)
			AssertCancelledContext(t, err)
		})

		t.Run("GetByTaskID with cancelled context", func(t *testing.T) {
			_, err := repo.GetByTaskID(NewCanceledContext(), task.ID)
			AssertCancelledContext(t, err)
		})

		t.Run("GetTotalTimeByTaskID with cancelled context", func(t *testing.T) {
			_, err := repo.GetTotalTimeByTaskID(NewCanceledContext(), task.ID)
			AssertCancelledContext(t, err)
		})

		t.Run("Delete with cancelled context", func(t *testing.T) {
			err := repo.Delete(NewCanceledContext(), entry.ID)
			AssertCancelledContext(t, err)
		})

		t.Run("GetByDateRange with cancelled context", func(t *testing.T) {
			start := time.Now().AddDate(0, 0, -1)
			end := time.Now()

			_, err := repo.GetByDateRange(NewCanceledContext(), start, end)
			AssertCancelledContext(t, err)
		})
	})

	t.Run("Edge Cases", func(t *testing.T) {
		db := CreateTestDB(t)
		repo := NewTimeEntryRepository(db)
		ctx := context.Background()

		t.Run("Get non-existent entry", func(t *testing.T) {
			_, err := repo.Get(ctx, 99999)
			shared.AssertError(t, err, "Expected error for non-existent entry")
			shared.AssertEqual(t, sql.ErrNoRows, err, "Expected sql.ErrNoRows")
		})

		t.Run("Stop non-existent entry", func(t *testing.T) {
			_, err := repo.Stop(ctx, 99999)
			shared.AssertError(t, err, "Expected error for non-existent entry")
		})

		t.Run("Delete non-existent entry", func(t *testing.T) {
			err := repo.Delete(ctx, 99999)
			shared.AssertError(t, err, "Expected error for non-existent entry")
			shared.AssertContains(t, err.Error(), "time entry not found", "Expected specific error message")
		})

		t.Run("Start with non-existent task", func(t *testing.T) {
			_, err := repo.Start(ctx, 99999, "Test")
			shared.AssertError(t, err, "Expected error for non-existent task")
		})

		t.Run("GetActiveByTaskID with no results", func(t *testing.T) {
			task := createTestTask(t, db)
			_, err := repo.GetActiveByTaskID(ctx, task.ID)
			shared.AssertError(t, err, "Expected error when no active entry")
			shared.AssertEqual(t, sql.ErrNoRows, err, "Expected sql.ErrNoRows")
		})

		t.Run("GetByTaskID with no results", func(t *testing.T) {
			task := createTestTask(t, db)
			entries, err := repo.GetByTaskID(ctx, task.ID)
			shared.AssertNoError(t, err, "Should not error when no entries found")
			shared.AssertEqual(t, 0, len(entries), "Expected empty result set")
		})
	})
}
