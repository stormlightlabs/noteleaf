package handlers

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stormlightlabs/noteleaf/internal/models"
)

func setupTimeTrackingTestHandler(t *testing.T) (*TaskHandler, func()) {
	tempDir := t.TempDir()
	os.Setenv("NOTELEAF_CONFIG_DIR", tempDir)

	handler, err := NewTaskHandler()
	if err != nil {
		t.Fatalf("Failed to create test handler: %v", err)
	}

	cleanup := func() {
		handler.Close()
		os.Unsetenv("NOTELEAF_CONFIG_DIR")
	}

	return handler, cleanup
}

func createTimeTrackingTestTask(t *testing.T, handler *TaskHandler) *models.Task {
	ctx := context.Background()
	task := &models.Task{
		UUID:        fmt.Sprintf("test-time-uuid-%d", time.Now().UnixNano()),
		Description: "Test Time Tracking Task",
		Status:      "pending",
	}

	id, err := handler.repos.Tasks.Create(ctx, task)
	if err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}
	task.ID = id
	return task
}

func TestTimeTracking(t *testing.T) {
	t.Run("Start", func(t *testing.T) {
		handler, cleanup := setupTimeTrackingTestHandler(t)
		defer cleanup()

		ctx := context.Background()
		task := createTimeTrackingTestTask(t, handler)

		t.Run("starts time tracking by ID", func(t *testing.T) {
			err := handler.Start(ctx, fmt.Sprintf("%d", task.ID), "Working on tests")

			if err != nil {
				t.Fatalf("Failed to start time tracking: %v", err)
			}

			active, err := handler.repos.TimeEntries.GetActiveByTaskID(ctx, task.ID)
			if err != nil {
				t.Fatalf("Failed to get active time entry: %v", err)
			}

			if active.Description != "Working on tests" {
				t.Errorf("Expected description 'Working on tests', got %q", active.Description)
			}
			if !active.IsActive() {
				t.Error("Expected time entry to be active")
			}
		})

		t.Run("starts time tracking by UUID", func(t *testing.T) {
			err := handler.Stop(ctx, task.UUID)
			if err != nil {
				t.Fatalf("Failed to stop previous tracking: %v", err)
			}

			err = handler.Start(ctx, task.UUID, "Working via UUID")

			if err != nil {
				t.Fatalf("Failed to start time tracking by UUID: %v", err)
			}

			active, err := handler.repos.TimeEntries.GetActiveByTaskID(ctx, task.ID)
			if err != nil {
				t.Fatalf("Failed to get active time entry: %v", err)
			}

			if active.Description != "Working via UUID" {
				t.Errorf("Expected description 'Working via UUID', got %q", active.Description)
			}
		})

		t.Run("handles already started task gracefully", func(t *testing.T) {
			err := handler.Start(ctx, fmt.Sprintf("%d", task.ID), "Another attempt")

			if err != nil {
				t.Fatalf("Expected graceful handling of already started task, got error: %v", err)
			}
		})

		t.Run("fails with non-existent task", func(t *testing.T) {
			err := handler.Start(ctx, "99999", "Non-existent task")

			if err == nil {
				t.Error("Expected error for non-existent task")
			}
			if !strings.Contains(err.Error(), "failed to find task") {
				t.Errorf("Expected 'failed to find task' error, got: %v", err)
			}
		})
	})

	t.Run("Stop", func(t *testing.T) {
		handler, cleanup := setupTimeTrackingTestHandler(t)
		defer cleanup()

		ctx := context.Background()
		task := createTimeTrackingTestTask(t, handler)

		t.Run("stops active time tracking", func(t *testing.T) {
			err := handler.Start(ctx, fmt.Sprintf("%d", task.ID), "Test work")
			if err != nil {
				t.Fatalf("Failed to start time tracking: %v", err)
			}

			time.Sleep(1010 * time.Millisecond)

			err = handler.Stop(ctx, fmt.Sprintf("%d", task.ID))

			if err != nil {
				t.Fatalf("Failed to stop time tracking: %v", err)
			}

			_, err = handler.repos.TimeEntries.GetActiveByTaskID(ctx, task.ID)
			if err.Error() != "sql: no rows in result set" {
				t.Errorf("Expected no active time entry after stopping, got: %v", err)
			}
		})

		t.Run("handles no active tracking gracefully", func(t *testing.T) {
			err := handler.Stop(ctx, fmt.Sprintf("%d", task.ID))

			if err != nil {
				t.Fatalf("Expected graceful handling of no active tracking, got error: %v", err)
			}
		})

		t.Run("stops by UUID", func(t *testing.T) {
			err := handler.Start(ctx, task.UUID, "UUID test")
			if err != nil {
				t.Fatalf("Failed to start time tracking: %v", err)
			}

			time.Sleep(1010 * time.Millisecond)

			err = handler.Stop(ctx, task.UUID)

			if err != nil {
				t.Fatalf("Failed to stop time tracking by UUID: %v", err)
			}
		})

		t.Run("fails with non-existent task", func(t *testing.T) {
			err := handler.Stop(ctx, "99999")

			if err == nil {
				t.Error("Expected error for non-existent task")
			}
			if !strings.Contains(err.Error(), "failed to find task") {
				t.Errorf("Expected 'failed to find task' error, got: %v", err)
			}
		})
	})

	t.Run("Timesheet", func(t *testing.T) {
		handler, cleanup := setupTimeTrackingTestHandler(t)
		defer cleanup()

		ctx := context.Background()
		task1 := createTimeTrackingTestTask(t, handler)

		task2 := &models.Task{
			UUID:        fmt.Sprintf("test-time-uuid-2-%d", time.Now().UnixNano()),
			Description: "Second Time Tracking Task",
			Status:      "pending",
		}
		id2, err := handler.repos.Tasks.Create(ctx, task2)
		if err != nil {
			t.Fatalf("Failed to create second test task: %v", err)
		}
		task2.ID = id2

		setupTimeEntries := func() {
			entry1, _ := handler.repos.TimeEntries.Start(ctx, task1.ID, "First session")
			time.Sleep(1010 * time.Millisecond)
			handler.repos.TimeEntries.Stop(ctx, entry1.ID)

			entry2, _ := handler.repos.TimeEntries.Start(ctx, task2.ID, "Second task work")
			time.Sleep(1010 * time.Millisecond)
			handler.repos.TimeEntries.Stop(ctx, entry2.ID)

			handler.repos.TimeEntries.Start(ctx, task1.ID, "Active work")
		}

		t.Run("shows general timesheet", func(t *testing.T) {
			setupTimeEntries()

			err := handler.Timesheet(ctx, 7, "")

			if err != nil {
				t.Fatalf("Failed to generate timesheet: %v", err)
			}
		})

		t.Run("shows task-specific timesheet", func(t *testing.T) {
			err := handler.Timesheet(ctx, 7, fmt.Sprintf("%d", task1.ID))

			if err != nil {
				t.Fatalf("Failed to generate task timesheet: %v", err)
			}
		})

		t.Run("shows task-specific timesheet by UUID", func(t *testing.T) {
			err := handler.Timesheet(ctx, 7, task1.UUID)

			if err != nil {
				t.Fatalf("Failed to generate task timesheet by UUID: %v", err)
			}
		})

		t.Run("handles empty timesheet gracefully", func(t *testing.T) {
			task3 := &models.Task{
				UUID:        fmt.Sprintf("test-empty-uuid-%d", time.Now().UnixNano()),
				Description: "Empty Task",
				Status:      "pending",
			}
			id3, err := handler.repos.Tasks.Create(ctx, task3)
			if err != nil {
				t.Fatalf("Failed to create empty test task: %v", err)
			}

			err = handler.Timesheet(ctx, 7, fmt.Sprintf("%d", id3))

			if err != nil {
				t.Fatalf("Failed to handle empty timesheet: %v", err)
			}
		})

		t.Run("fails with non-existent task", func(t *testing.T) {
			err := handler.Timesheet(ctx, 7, "99999")

			if err == nil {
				t.Error("Expected error for non-existent task")
			}
			if !strings.Contains(err.Error(), "failed to find task") {
				t.Errorf("Expected 'failed to find task' error, got: %v", err)
			}
		})
	})

	t.Run("TestFormatDuration", func(t *testing.T) {
		tests := []struct {
			duration time.Duration
			expected string
		}{
			{30 * time.Second, "30s"},
			{90 * time.Second, "2m"},
			{30 * time.Minute, "30m"},
			{90 * time.Minute, "1.5h"},
			{2 * time.Hour, "2.0h"},
			{25 * time.Hour, "1d 1.0h"},
			{48 * time.Hour, "2d"},
			{72 * time.Hour, "3d"},
		}

		for _, test := range tests {
			result := formatDuration(test.duration)
			if result != test.expected {
				t.Errorf("formatDuration(%v) = %q, expected %q", test.duration, result, test.expected)
			}
		}
	})

	t.Run("TestTimeEntryMethods", func(t *testing.T) {
		now := time.Now()

		t.Run("IsActive returns true for entry without end time", func(t *testing.T) {
			entry := &models.TimeEntry{
				StartTime: now,
				EndTime:   nil,
			}

			if !entry.IsActive() {
				t.Error("Expected entry to be active")
			}
		})

		t.Run("IsActive returns false for entry with end time", func(t *testing.T) {
			endTime := now.Add(time.Hour)
			entry := &models.TimeEntry{
				StartTime: now,
				EndTime:   &endTime,
			}

			if entry.IsActive() {
				t.Error("Expected entry to not be active")
			}
		})

		t.Run("Stop sets end time and calculates duration", func(t *testing.T) {
			entry := &models.TimeEntry{
				StartTime: now.Add(-time.Second), // Start 1 second ago
				EndTime:   nil,
			}

			entry.Stop()

			if entry.EndTime == nil {
				t.Error("Expected EndTime to be set after stopping")
			}
			if entry.DurationSeconds <= 0 {
				t.Error("Expected duration to be calculated and greater than 0")
			}
			if entry.IsActive() {
				t.Error("Expected entry to not be active after stopping")
			}
		})

		t.Run("GetDuration returns calculated duration for completed entry", func(t *testing.T) {
			start := now
			end := now.Add(2 * time.Hour)
			entry := &models.TimeEntry{
				StartTime:       start,
				EndTime:         &end,
				DurationSeconds: int64((2 * time.Hour).Seconds()),
			}

			duration := entry.GetDuration()
			expected := 2 * time.Hour

			if duration != expected {
				t.Errorf("Expected duration %v, got %v", expected, duration)
			}
		})

		t.Run("GetDuration returns live duration for active entry", func(t *testing.T) {
			start := time.Now().Add(-time.Minute)
			entry := &models.TimeEntry{
				StartTime: start,
				EndTime:   nil,
			}

			duration := entry.GetDuration()

			if duration < 59*time.Second || duration > 61*time.Second {
				t.Errorf("Expected duration around 1 minute, got %v", duration)
			}
		})
	})
}
