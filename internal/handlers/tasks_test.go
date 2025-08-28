package handlers

import (
	"context"
	"os"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stormlightlabs/noteleaf/internal/models"
)

func setupTaskTest(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", "noteleaf-task-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	oldConfigHome := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tempDir)

	cleanup := func() {
		os.Setenv("XDG_CONFIG_HOME", oldConfigHome)
		os.RemoveAll(tempDir)
	}

	ctx := context.Background()
	err = Setup(ctx, []string{})
	if err != nil {
		cleanup()
		t.Fatalf("Failed to setup database: %v", err)
	}

	return tempDir, cleanup
}

func TestTaskHandler(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		t.Run("creates handler successfully", func(t *testing.T) {
			_, cleanup := setupTaskTest(t)
			defer cleanup()

			handler, err := NewTaskHandler()
			if err != nil {
				t.Fatalf("NewTaskHandler failed: %v", err)
			}
			if handler == nil {
				t.Fatal("Handler should not be nil")
			}
			defer handler.Close()

			if handler.db == nil {
				t.Error("Handler database should not be nil")
			}
			if handler.config == nil {
				t.Error("Handler config should not be nil")
			}
			if handler.repos == nil {
				t.Error("Handler repos should not be nil")
			}
		})

		t.Run("handles database initialization error", func(t *testing.T) {
			originalXDG := os.Getenv("XDG_CONFIG_HOME")
			originalHome := os.Getenv("HOME")

			if runtime.GOOS == "windows" {
				originalAppData := os.Getenv("APPDATA")
				os.Unsetenv("APPDATA")
				defer os.Setenv("APPDATA", originalAppData)
			} else {
				os.Unsetenv("XDG_CONFIG_HOME")
				os.Unsetenv("HOME")
				defer os.Setenv("XDG_CONFIG_HOME", originalXDG)
				defer os.Setenv("HOME", originalHome)
			}

			handler, err := NewTaskHandler()
			if err == nil {
				if handler != nil {
					handler.Close()
				}
				t.Error("Expected error when database initialization fails")
			}
		})
	})

	t.Run("Create", func(t *testing.T) {
		_, cleanup := setupTaskTest(t)
		defer cleanup()

		t.Run("creates task successfully", func(t *testing.T) {
			ctx := context.Background()
			args := []string{"Buy groceries", "and", "cook dinner"}

			err := CreateTask(ctx, args)
			if err != nil {
				t.Errorf("CreateTask failed: %v", err)
			}

			handler, err := NewTaskHandler()
			if err != nil {
				t.Fatalf("Failed to create handler: %v", err)
			}
			defer handler.Close()

			tasks, err := handler.repos.Tasks.GetPending(ctx)
			if err != nil {
				t.Fatalf("Failed to get pending tasks: %v", err)
			}

			if len(tasks) != 1 {
				t.Errorf("Expected 1 task, got %d", len(tasks))
			}

			task := tasks[0]
			expectedDesc := "Buy groceries and cook dinner"
			if task.Description != expectedDesc {
				t.Errorf("Expected description '%s', got '%s'", expectedDesc, task.Description)
			}

			if task.Status != "pending" {
				t.Errorf("Expected status 'pending', got '%s'", task.Status)
			}

			if task.UUID == "" {
				t.Error("Task should have a UUID")
			}
		})

		t.Run("fails with empty description", func(t *testing.T) {
			ctx := context.Background()
			args := []string{}

			err := CreateTask(ctx, args)
			if err == nil {
				t.Error("Expected error for empty description")
			}

			if !strings.Contains(err.Error(), "task description required") {
				t.Errorf("Expected error about required description, got: %v", err)
			}
		})
	})

	t.Run("List", func(t *testing.T) {
		_, cleanup := setupTaskTest(t)
		defer cleanup()

		ctx := context.Background()

		handler, err := NewTaskHandler()
		if err != nil {
			t.Fatalf("Failed to create handler: %v", err)
		}
		defer handler.Close()

		task1 := &models.Task{
			UUID:        uuid.New().String(),
			Description: "Task 1",
			Status:      "pending",
			Priority:    "A",
			Project:     "work",
		}
		_, err = handler.repos.Tasks.Create(ctx, task1)
		if err != nil {
			t.Fatalf("Failed to create task1: %v", err)
		}

		task2 := &models.Task{
			UUID:        uuid.New().String(),
			Description: "Task 2",
			Status:      "completed",
		}
		_, err = handler.repos.Tasks.Create(ctx, task2)
		if err != nil {
			t.Fatalf("Failed to create task2: %v", err)
		}

		t.Run("lists pending tasks by default (static mode)", func(t *testing.T) {
			err := ListTasks(ctx, true, false, "", "", "")
			if err != nil {
				t.Errorf("ListTasks failed: %v", err)
			}
		})

		t.Run("filters by status (static mode)", func(t *testing.T) {
			err := ListTasks(ctx, true, false, "completed", "", "")
			if err != nil {
				t.Errorf("ListTasks with status filter failed: %v", err)
			}
		})

		t.Run("filters by priority (static mode)", func(t *testing.T) {
			err := ListTasks(ctx, true, false, "", "A", "")
			if err != nil {
				t.Errorf("ListTasks with priority filter failed: %v", err)
			}
		})

		t.Run("filters by project (static mode)", func(t *testing.T) {
			err := ListTasks(ctx, true, false, "", "", "work")
			if err != nil {
				t.Errorf("ListTasks with project filter failed: %v", err)
			}
		})

		t.Run("show all tasks (static mode)", func(t *testing.T) {
			err := ListTasks(ctx, true, true, "", "", "")
			if err != nil {
				t.Errorf("ListTasks with show all failed: %v", err)
			}
		})
	})

	t.Run("Update", func(t *testing.T) {
		_, cleanup := setupTaskTest(t)
		defer cleanup()

		ctx := context.Background()

		// Create test task
		handler, err := NewTaskHandler()
		if err != nil {
			t.Fatalf("Failed to create handler: %v", err)
		}
		defer handler.Close()

		task := &models.Task{
			UUID:        uuid.New().String(),
			Description: "Original description",
			Status:      "pending",
		}
		id, err := handler.repos.Tasks.Create(ctx, task)
		if err != nil {
			t.Fatalf("Failed to create task: %v", err)
		}

		t.Run("updates task by ID", func(t *testing.T) {
			args := []string{strconv.FormatInt(id, 10), "--description", "Updated description"}

			err := UpdateTask(ctx, args)
			if err != nil {
				t.Errorf("UpdateTask failed: %v", err)
			}

			updatedTask, err := handler.repos.Tasks.Get(ctx, id)
			if err != nil {
				t.Fatalf("Failed to get updated task: %v", err)
			}

			if updatedTask.Description != "Updated description" {
				t.Errorf("Expected description 'Updated description', got '%s'", updatedTask.Description)
			}
		})

		t.Run("updates task by UUID", func(t *testing.T) {
			args := []string{task.UUID, "--status", "completed"}

			err := UpdateTask(ctx, args)
			if err != nil {
				t.Errorf("UpdateTask by UUID failed: %v", err)
			}

			updatedTask, err := handler.repos.Tasks.GetByUUID(ctx, task.UUID)
			if err != nil {
				t.Fatalf("Failed to get updated task by UUID: %v", err)
			}

			if updatedTask.Status != "completed" {
				t.Errorf("Expected status 'completed', got '%s'", updatedTask.Status)
			}
		})

		t.Run("updates multiple fields", func(t *testing.T) {
			args := []string{
				strconv.FormatInt(id, 10),
				"--description", "Multiple updates",
				"--priority", "B",
				"--project", "test",
				"--due", "2024-12-31",
			}

			err := UpdateTask(ctx, args)
			if err != nil {
				t.Errorf("UpdateTask with multiple fields failed: %v", err)
			}

			// Verify all updates
			updatedTask, err := handler.repos.Tasks.Get(ctx, id)
			if err != nil {
				t.Fatalf("Failed to get updated task: %v", err)
			}

			if updatedTask.Description != "Multiple updates" {
				t.Errorf("Expected description 'Multiple updates', got '%s'", updatedTask.Description)
			}
			if updatedTask.Priority != "B" {
				t.Errorf("Expected priority 'B', got '%s'", updatedTask.Priority)
			}
			if updatedTask.Project != "test" {
				t.Errorf("Expected project 'test', got '%s'", updatedTask.Project)
			}
			if updatedTask.Due == nil {
				t.Error("Expected due date to be set")
			}
		})

		t.Run("adds and removes tags", func(t *testing.T) {
			args := []string{
				strconv.FormatInt(id, 10),
				"--add-tag=work",
				"--add-tag=urgent",
			}

			err := UpdateTask(ctx, args)
			if err != nil {
				t.Errorf("UpdateTask with add tags failed: %v", err)
			}

			updatedTask, err := handler.repos.Tasks.Get(ctx, id)
			if err != nil {
				t.Fatalf("Failed to get updated task: %v", err)
			}

			if len(updatedTask.Tags) != 2 {
				t.Errorf("Expected 2 tags, got %d", len(updatedTask.Tags))
			}

			args = []string{
				strconv.FormatInt(id, 10),
				"--remove-tag=urgent",
			}

			err = UpdateTask(ctx, args)
			if err != nil {
				t.Errorf("UpdateTask with remove tag failed: %v", err)
			}

			updatedTask, err = handler.repos.Tasks.Get(ctx, id)
			if err != nil {
				t.Fatalf("Failed to get updated task: %v", err)
			}

			if len(updatedTask.Tags) != 1 {
				t.Errorf("Expected 1 tag after removal, got %d", len(updatedTask.Tags))
			}

			if updatedTask.Tags[0] != "work" {
				t.Errorf("Expected remaining tag 'work', got '%s'", updatedTask.Tags[0])
			}
		})

		t.Run("fails with missing task ID", func(t *testing.T) {
			args := []string{}

			err := UpdateTask(ctx, args)
			if err == nil {
				t.Error("Expected error for missing task ID")
			}

			if !strings.Contains(err.Error(), "task ID required") {
				t.Errorf("Expected error about required task ID, got: %v", err)
			}
		})

		t.Run("fails with invalid task ID", func(t *testing.T) {
			args := []string{"99999", "--description", "test"}

			err := UpdateTask(ctx, args)
			if err == nil {
				t.Error("Expected error for invalid task ID")
			}

			if !strings.Contains(err.Error(), "failed to find task") {
				t.Errorf("Expected error about task not found, got: %v", err)
			}
		})
	})

	t.Run("Delete", func(t *testing.T) {
		_, cleanup := setupTaskTest(t)
		defer cleanup()

		ctx := context.Background()

		handler, err := NewTaskHandler()
		if err != nil {
			t.Fatalf("Failed to create handler: %v", err)
		}
		defer handler.Close()

		task := &models.Task{
			UUID:        uuid.New().String(),
			Description: "Task to delete",
			Status:      "pending",
		}
		id, err := handler.repos.Tasks.Create(ctx, task)
		if err != nil {
			t.Fatalf("Failed to create task: %v", err)
		}

		t.Run("deletes task by ID", func(t *testing.T) {
			args := []string{strconv.FormatInt(id, 10)}

			err := DeleteTask(ctx, args)
			if err != nil {
				t.Errorf("DeleteTask failed: %v", err)
			}

			_, err = handler.repos.Tasks.Get(ctx, id)
			if err == nil {
				t.Error("Expected error when getting deleted task")
			}
		})

		t.Run("deletes task by UUID", func(t *testing.T) {
			task2 := &models.Task{
				UUID:        uuid.New().String(),
				Description: "Task to delete by UUID",
				Status:      "pending",
			}
			_, err := handler.repos.Tasks.Create(ctx, task2)
			if err != nil {
				t.Fatalf("Failed to create task2: %v", err)
			}

			args := []string{task2.UUID}

			err = DeleteTask(ctx, args)
			if err != nil {
				t.Errorf("DeleteTask by UUID failed: %v", err)
			}

			_, err = handler.repos.Tasks.GetByUUID(ctx, task2.UUID)
			if err == nil {
				t.Error("Expected error when getting deleted task by UUID")
			}
		})

		t.Run("fails with missing task ID", func(t *testing.T) {
			args := []string{}

			err := DeleteTask(ctx, args)
			if err == nil {
				t.Error("Expected error for missing task ID")
			}

			if !strings.Contains(err.Error(), "task ID required") {
				t.Errorf("Expected error about required task ID, got: %v", err)
			}
		})

		t.Run("fails with invalid task ID", func(t *testing.T) {
			args := []string{"99999"}

			err := DeleteTask(ctx, args)
			if err == nil {
				t.Error("Expected error for invalid task ID")
			}

			if !strings.Contains(err.Error(), "failed to find task") {
				t.Errorf("Expected error about task not found, got: %v", err)
			}
		})
	})

	t.Run("View", func(t *testing.T) {
		_, cleanup := setupTaskTest(t)
		defer cleanup()

		ctx := context.Background()

		handler, err := NewTaskHandler()
		if err != nil {
			t.Fatalf("Failed to create handler: %v", err)
		}
		defer handler.Close()

		now := time.Now()
		task := &models.Task{
			UUID:        uuid.New().String(),
			Description: "Task to view",
			Status:      "pending",
			Priority:    "A",
			Project:     "test",
			Tags:        []string{"work", "important"},
			Entry:       now,
			Modified:    now,
		}
		id, err := handler.repos.Tasks.Create(ctx, task)
		if err != nil {
			t.Fatalf("Failed to create task: %v", err)
		}

		t.Run("views task by ID", func(t *testing.T) {
			args := []string{strconv.FormatInt(id, 10)}

			err := ViewTask(ctx, args)
			if err != nil {
				t.Errorf("ViewTask failed: %v", err)
			}
		})

		t.Run("views task by UUID", func(t *testing.T) {
			args := []string{task.UUID}

			err := ViewTask(ctx, args)
			if err != nil {
				t.Errorf("ViewTask by UUID failed: %v", err)
			}
		})

		t.Run("fails with missing task ID", func(t *testing.T) {
			args := []string{}

			err := ViewTask(ctx, args)
			if err == nil {
				t.Error("Expected error for missing task ID")
			}

			if !strings.Contains(err.Error(), "task ID required") {
				t.Errorf("Expected error about required task ID, got: %v", err)
			}
		})

		t.Run("fails with invalid task ID", func(t *testing.T) {
			args := []string{"99999"}

			err := ViewTask(ctx, args)
			if err == nil {
				t.Error("Expected error for invalid task ID")
			}

			if !strings.Contains(err.Error(), "failed to find task") {
				t.Errorf("Expected error about task not found, got: %v", err)
			}
		})
	})

	t.Run("Done", func(t *testing.T) {
		_, cleanup := setupTaskTest(t)
		defer cleanup()

		ctx := context.Background()

		handler, err := NewTaskHandler()
		if err != nil {
			t.Fatalf("Failed to create handler: %v", err)
		}
		defer handler.Close()

		task := &models.Task{
			UUID:        uuid.New().String(),
			Description: "Task to complete",
			Status:      "pending",
		}
		id, err := handler.repos.Tasks.Create(ctx, task)
		if err != nil {
			t.Fatalf("Failed to create task: %v", err)
		}

		t.Run("marks task as done by ID", func(t *testing.T) {
			args := []string{strconv.FormatInt(id, 10)}

			err := DoneTask(ctx, args)
			if err != nil {
				t.Errorf("DoneTask failed: %v", err)
			}

			completedTask, err := handler.repos.Tasks.Get(ctx, id)
			if err != nil {
				t.Fatalf("Failed to get completed task: %v", err)
			}

			if completedTask.Status != "completed" {
				t.Errorf("Expected status 'completed', got '%s'", completedTask.Status)
			}

			if completedTask.End == nil {
				t.Error("Expected end time to be set")
			}
		})

		t.Run("handles already completed task", func(t *testing.T) {
			task2 := &models.Task{
				UUID:        uuid.New().String(),
				Description: "Already completed task",
				Status:      "completed",
			}
			id2, err := handler.repos.Tasks.Create(ctx, task2)
			if err != nil {
				t.Fatalf("Failed to create task2: %v", err)
			}

			args := []string{strconv.FormatInt(id2, 10)}

			err = DoneTask(ctx, args)
			if err != nil {
				t.Errorf("DoneTask on completed task failed: %v", err)
			}
		})

		t.Run("marks task as done by UUID", func(t *testing.T) {
			task3 := &models.Task{
				UUID:        uuid.New().String(),
				Description: "Task to complete by UUID",
				Status:      "pending",
			}
			_, err := handler.repos.Tasks.Create(ctx, task3)
			if err != nil {
				t.Fatalf("Failed to create task3: %v", err)
			}

			args := []string{task3.UUID}

			err = DoneTask(ctx, args)
			if err != nil {
				t.Errorf("DoneTask by UUID failed: %v", err)
			}

			completedTask, err := handler.repos.Tasks.GetByUUID(ctx, task3.UUID)
			if err != nil {
				t.Fatalf("Failed to get completed task by UUID: %v", err)
			}

			if completedTask.Status != "completed" {
				t.Errorf("Expected status 'completed', got '%s'", completedTask.Status)
			}

			if completedTask.End == nil {
				t.Error("Expected end time to be set")
			}
		})

		t.Run("fails with missing task ID", func(t *testing.T) {
			args := []string{}

			err := DoneTask(ctx, args)
			if err == nil {
				t.Error("Expected error for missing task ID")
			}

			if !strings.Contains(err.Error(), "task ID required") {
				t.Errorf("Expected error about required task ID, got: %v", err)
			}
		})

		t.Run("fails with invalid task ID", func(t *testing.T) {
			args := []string{"99999"}

			err := DoneTask(ctx, args)
			if err == nil {
				t.Error("Expected error for invalid task ID")
			}

			if !strings.Contains(err.Error(), "failed to find task") {
				t.Errorf("Expected error about task not found, got: %v", err)
			}
		})
	})

	t.Run("Helper", func(t *testing.T) {

		t.Run("removeString function", func(t *testing.T) {
			slice := []string{"a", "b", "c", "b"}
			result := removeString(slice, "b")

			if len(result) != 2 {
				t.Errorf("Expected 2 items after removing 'b', got %d", len(result))
			}

			if slices.Contains(result, "b") {
				t.Error("Expected 'b' to be removed from slice")
			}

			if !slices.Contains(result, "a") || !slices.Contains(result, "c") {
				t.Error("Expected 'a' and 'c' to remain in slice")
			}
		})
	})

	t.Run("Print", func(t *testing.T) {
		_, cleanup := setupTaskTest(t)
		defer cleanup()

		handler, err := NewTaskHandler()
		if err != nil {
			t.Fatalf("Failed to create handler: %v", err)
		}
		defer handler.Close()

		now := time.Now()
		due := now.Add(24 * time.Hour)

		task := &models.Task{
			ID:          1,
			UUID:        uuid.New().String(),
			Description: "Test task",
			Status:      "pending",
			Priority:    "A",
			Project:     "test",
			Tags:        []string{"work", "urgent"},
			Due:         &due,
			Entry:       now,
			Modified:    now,
		}

		t.Run("printTask doesn't panic", func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("printTask panicked: %v", r)
				}
			}()

			handler.printTask(task)
		})

		t.Run("printTaskDetail doesn't panic", func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("printTaskDetail panicked: %v", r)
				}
			}()

			handler.printTaskDetail(task)
		})
	})
}
