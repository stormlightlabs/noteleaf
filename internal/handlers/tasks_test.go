package handlers

import (
	"bytes"
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
	"github.com/stormlightlabs/noteleaf/internal/ui"
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

		handler, err := NewTaskHandler()
		if err != nil {
			t.Fatalf("Failed to create handler: %v", err)
		}
		defer handler.Close()

		t.Run("creates task successfully", func(t *testing.T) {
			ctx := context.Background()
			args := []string{"Buy groceries", "and", "cook dinner"}

			err := handler.Create(ctx, args, "", "", "", "", []string{})
			if err != nil {
				t.Errorf("CreateTask failed: %v", err)
			}

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

			err := handler.Create(ctx, args, "", "", "", "", []string{})
			if err == nil {
				t.Error("Expected error for empty description")
			}

			if !strings.Contains(err.Error(), "task description required") {
				t.Errorf("Expected error about required description, got: %v", err)
			}
		})

		t.Run("creates task with flags", func(t *testing.T) {
			ctx := context.Background()
			args := []string{"Task", "with", "flags"}
			priority := "A"
			project := "test-project"
			due := "2024-12-31"
			tags := []string{"urgent", "work"}

			err := handler.Create(ctx, args, priority, project, "test-context", due, tags)
			if err != nil {
				t.Errorf("CreateTask with flags failed: %v", err)
			}

			tasks, err := handler.repos.Tasks.GetPending(ctx)
			if err != nil {
				t.Fatalf("Failed to get pending tasks: %v", err)
			}

			if len(tasks) < 1 {
				t.Errorf("Expected at least 1 task, got %d", len(tasks))
			}

			var task *models.Task
			for _, t := range tasks {
				if t.Description == "Task with flags" {
					task = t
					break
				}
			}

			if task == nil {
				t.Fatal("Could not find created task")
			}

			if task.Priority != priority {
				t.Errorf("Expected priority '%s', got '%s'", priority, task.Priority)
			}

			if task.Project != project {
				t.Errorf("Expected project '%s', got '%s'", project, task.Project)
			}

			if task.Due == nil {
				t.Error("Expected due date to be set")
			} else if task.Due.Format("2006-01-02") != due {
				t.Errorf("Expected due date '%s', got '%s'", due, task.Due.Format("2006-01-02"))
			}

			if len(task.Tags) != len(tags) {
				t.Errorf("Expected %d tags, got %d", len(tags), len(task.Tags))
			} else {
				for i, tag := range tags {
					if task.Tags[i] != tag {
						t.Errorf("Expected tag '%s' at index %d, got '%s'", tag, i, task.Tags[i])
					}
				}
			}
		})

		t.Run("fails with invalid due date format", func(t *testing.T) {
			ctx := context.Background()
			args := []string{"Task", "with", "invalid", "date"}
			invalidDue := "invalid-date"

			err := handler.Create(ctx, args, "", "", "", invalidDue, []string{})
			if err == nil {
				t.Error("Expected error for invalid due date format")
			}

			if !strings.Contains(err.Error(), "invalid due date format") {
				t.Errorf("Expected error about invalid date format, got: %v", err)
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
			err := handler.List(ctx, true, false, "", "", "", "")
			if err != nil {
				t.Errorf("ListTasks failed: %v", err)
			}
		})

		t.Run("filters by status (static mode)", func(t *testing.T) {
			err := handler.List(ctx, true, false, "completed", "", "", "")
			if err != nil {
				t.Errorf("ListTasks with status filter failed: %v", err)
			}
		})

		t.Run("filters by priority (static mode)", func(t *testing.T) {
			err := handler.List(ctx, true, false, "", "A", "", "")
			if err != nil {
				t.Errorf("ListTasks with priority filter failed: %v", err)
			}
		})

		t.Run("filters by project (static mode)", func(t *testing.T) {
			err := handler.List(ctx, true, false, "", "", "work", "")
			if err != nil {
				t.Errorf("ListTasks with project filter failed: %v", err)
			}
		})

		t.Run("show all tasks (static mode)", func(t *testing.T) {
			err := handler.List(ctx, true, true, "", "", "", "")
			if err != nil {
				t.Errorf("ListTasks with show all failed: %v", err)
			}
		})
	})

	t.Run("Update", func(t *testing.T) {
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
			Description: "Original description",
			Status:      "pending",
		}
		id, err := handler.repos.Tasks.Create(ctx, task)
		if err != nil {
			t.Fatalf("Failed to create task: %v", err)
		}

		t.Run("updates task by ID", func(t *testing.T) {
			taskID := strconv.FormatInt(id, 10)

			err := handler.Update(ctx, taskID, "Updated description", "", "", "", "", "", []string{}, []string{})
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
			taskID := task.UUID

			err := handler.Update(ctx, taskID, "", "completed", "", "", "", "", []string{}, []string{})
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
			taskID := strconv.FormatInt(id, 10)

			err := handler.Update(ctx, taskID, "Multiple updates", "", "B", "test", "office", "2024-12-31", []string{}, []string{})
			if err != nil {
				t.Errorf("UpdateTask with multiple fields failed: %v", err)
			}

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
			taskID := strconv.FormatInt(id, 10)

			err := handler.Update(ctx, taskID, "", "", "", "", "", "", []string{"work", "urgent"}, []string{})
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

			taskID = strconv.FormatInt(id, 10)

			err = handler.Update(ctx, taskID, "", "", "", "", "", "", []string{}, []string{"urgent"})
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
			err := handler.Update(ctx, "", "", "", "", "", "", "", []string{}, []string{})
			if err == nil {
				t.Error("Expected error for missing task ID")
			}

			if !strings.Contains(err.Error(), "failed to find task") {
				t.Errorf("Expected error about task not found, got: %v", err)
			}
		})

		t.Run("fails with invalid task ID", func(t *testing.T) {
			taskID := "99999"

			err := handler.Update(ctx, taskID, "test", "", "", "", "", "", []string{}, []string{})
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

			err := handler.Delete(ctx, args)
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

			err = handler.Delete(ctx, args)
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

			err := handler.Delete(ctx, args)
			if err == nil {
				t.Error("Expected error for missing task ID")
			}

			if !strings.Contains(err.Error(), "task ID required") {
				t.Errorf("Expected error about required task ID, got: %v", err)
			}
		})

		t.Run("fails with invalid task ID", func(t *testing.T) {
			args := []string{"99999"}

			err := handler.Delete(ctx, args)
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

			err := handler.View(ctx, args, "detailed", false, false)
			if err != nil {
				t.Errorf("ViewTask failed: %v", err)
			}
		})

		t.Run("views task by UUID", func(t *testing.T) {
			args := []string{task.UUID}

			err := handler.View(ctx, args, "detailed", false, false)
			if err != nil {
				t.Errorf("ViewTask by UUID failed: %v", err)
			}
		})

		t.Run("fails with missing task ID", func(t *testing.T) {
			args := []string{}

			err := handler.View(ctx, args, "detailed", false, false)
			if err == nil {
				t.Error("Expected error for missing task ID")
			}

			if !strings.Contains(err.Error(), "task ID required") {
				t.Errorf("Expected error about required task ID, got: %v", err)
			}
		})

		t.Run("fails with invalid task ID", func(t *testing.T) {
			args := []string{"99999"}

			err := handler.View(ctx, args, "detailed", false, false)
			if err == nil {
				t.Error("Expected error for invalid task ID")
			}

			if !strings.Contains(err.Error(), "failed to find task") {
				t.Errorf("Expected error about task not found, got: %v", err)
			}
		})

		t.Run("uses brief format", func(t *testing.T) {
			args := []string{strconv.FormatInt(id, 10)}
			err := handler.View(ctx, args, "brief", false, false)
			if err != nil {
				t.Errorf("ViewTask with brief format failed: %v", err)
			}
		})

		t.Run("hides metadata", func(t *testing.T) {
			args := []string{strconv.FormatInt(id, 10)}
			err := handler.View(ctx, args, "detailed", false, true)
			if err != nil {
				t.Errorf("ViewTask with no-metadata failed: %v", err)
			}
		})

		t.Run("outputs JSON", func(t *testing.T) {
			args := []string{strconv.FormatInt(id, 10)}
			err := handler.View(ctx, args, "detailed", true, false)
			if err != nil {
				t.Errorf("ViewTask with JSON output failed: %v", err)
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

			err := handler.Done(ctx, args)
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

			err = handler.Done(ctx, args)
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

			err = handler.Done(ctx, args)
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

			err := handler.Done(ctx, args)
			if err == nil {
				t.Error("Expected error for missing task ID")
			}

			if !strings.Contains(err.Error(), "task ID required") {
				t.Errorf("Expected error about required task ID, got: %v", err)
			}
		})

		t.Run("fails with invalid task ID", func(t *testing.T) {
			args := []string{"99999"}

			err := handler.Done(ctx, args)
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

			handler.printTaskDetail(task, false)
		})
	})

	t.Run("ListProjects", func(t *testing.T) {
		_, cleanup := setupTaskTest(t)
		defer cleanup()

		ctx := context.Background()

		handler, err := NewTaskHandler()
		if err != nil {
			t.Fatalf("Failed to create handler: %v", err)
		}
		defer handler.Close()

		tasks := []*models.Task{
			{UUID: uuid.New().String(), Description: "Task 1", Status: "pending", Project: "web-app"},
			{UUID: uuid.New().String(), Description: "Task 2", Status: "completed", Project: "web-app"},
			{UUID: uuid.New().String(), Description: "Task 3", Status: "pending", Project: "mobile-app"},
			{UUID: uuid.New().String(), Description: "Task 4", Status: "pending", Project: ""},
		}

		for _, task := range tasks {
			_, err := handler.repos.Tasks.Create(ctx, task)
			if err != nil {
				t.Fatalf("Failed to create task: %v", err)
			}
		}

		t.Run("lists projects successfully", func(t *testing.T) {
			err := handler.ListProjects(ctx, true)
			if err != nil {
				t.Errorf("ListProjects failed: %v", err)
			}
		})

		t.Run("returns no projects when none exist", func(t *testing.T) {
			_, cleanup2 := setupTaskTest(t)
			defer cleanup2()

			err := handler.ListProjects(ctx, true)
			if err != nil {
				t.Errorf("ListProjects with no projects failed: %v", err)
			}
		})
	})

	t.Run("ListTags", func(t *testing.T) {
		_, cleanup := setupTaskTest(t)
		defer cleanup()

		ctx := context.Background()

		handler, err := NewTaskHandler()
		if err != nil {
			t.Fatalf("Failed to create handler: %v", err)
		}
		defer handler.Close()

		tasks := []*models.Task{
			{UUID: uuid.New().String(), Description: "Task 1", Status: "pending", Tags: []string{"frontend", "urgent"}},
			{UUID: uuid.New().String(), Description: "Task 2", Status: "completed", Tags: []string{"backend", "database"}},
			{UUID: uuid.New().String(), Description: "Task 3", Status: "pending", Tags: []string{"frontend", "ios"}},
			{UUID: uuid.New().String(), Description: "Task 4", Status: "pending", Tags: []string{}},
		}

		for _, task := range tasks {
			_, err := handler.repos.Tasks.Create(ctx, task)
			if err != nil {
				t.Fatalf("Failed to create task: %v", err)
			}
		}

		t.Run("lists tags successfully", func(t *testing.T) {
			err := handler.ListTags(ctx, true)
			if err != nil {
				t.Errorf("ListTags failed: %v", err)
			}
		})

		t.Run("returns no tags when none exist", func(t *testing.T) {
			_, cleanup2 := setupTaskTest(t)
			defer cleanup2()

			err := handler.ListTags(ctx, true)
			if err != nil {
				t.Errorf("ListTags with no tags failed: %v", err)
			}
		})
	})

	t.Run("Pluralize", func(t *testing.T) {
		t.Run("returns empty string for singular", func(t *testing.T) {
			result := pluralize(1)
			if result != "" {
				t.Errorf("Expected empty string for 1, got '%s'", result)
			}
		})

		t.Run("returns 's' for plural", func(t *testing.T) {
			result := pluralize(0)
			if result != "s" {
				t.Errorf("Expected 's' for 0, got '%s'", result)
			}

			result = pluralize(2)
			if result != "s" {
				t.Errorf("Expected 's' for 2, got '%s'", result)
			}

			result = pluralize(10)
			if result != "s" {
				t.Errorf("Expected 's' for 10, got '%s'", result)
			}
		})
	})

	t.Run("InteractiveComponentsStatic", func(t *testing.T) {
		_, cleanup := setupTaskTest(t)
		defer cleanup()

		handler, err := NewTaskHandler()
		if err != nil {
			t.Fatalf("Failed to create task handler: %v", err)
		}
		defer handler.Close()

		ctx := context.Background()

		err = handler.Create(ctx, []string{"Test", "Task", "1"}, "high", "test-project", "test-context", "", []string{"tag1"})
		if err != nil {
			t.Fatalf("Failed to create test task: %v", err)
		}

		err = handler.Create(ctx, []string{"Test", "Task", "2"}, "medium", "test-project", "test-context", "", []string{"tag2"})
		if err != nil {
			t.Fatalf("Failed to create test task: %v", err)
		}

		t.Run("taskListStaticMode", func(t *testing.T) {
			var output bytes.Buffer

			t.Run("lists all tasks", func(t *testing.T) {
				output.Reset()
				taskTable := ui.NewTaskListFromTable(handler.repos.Tasks, &output, os.Stdin, true, true, "", "", "")
				err := taskTable.Browse(ctx)
				if err != nil {
					t.Errorf("Static task list should succeed: %v", err)
				}
				if !strings.Contains(output.String(), "Test Task 1") {
					t.Error("Output should contain Test Task 1")
				}
				if !strings.Contains(output.String(), "Test Task 2") {
					t.Error("Output should contain Test Task 2")
				}
			})

			t.Run("filters by status", func(t *testing.T) {
				output.Reset()
				taskTable := ui.NewTaskListFromTable(handler.repos.Tasks, &output, os.Stdin, true, false, "pending", "", "")
				err := taskTable.Browse(ctx)
				if err != nil {
					t.Errorf("Static task list with status filter should succeed: %v", err)
				}
			})

			t.Run("filters by priority", func(t *testing.T) {
				output.Reset()
				taskTable := ui.NewTaskListFromTable(handler.repos.Tasks, &output, os.Stdin, true, false, "", "high", "")
				err := taskTable.Browse(ctx)
				if err != nil {
					t.Errorf("Static task list with priority filter should succeed: %v", err)
				}
			})

			t.Run("filters by project", func(t *testing.T) {
				output.Reset()
				taskTable := ui.NewTaskListFromTable(handler.repos.Tasks, &output, os.Stdin, true, false, "", "", "test-project")
				err := taskTable.Browse(ctx)
				if err != nil {
					t.Errorf("Static task list with project filter should succeed: %v", err)
				}
			})
		})

		t.Run("projectListStaticMode", func(t *testing.T) {
			var output bytes.Buffer

			t.Run("lists projects", func(t *testing.T) {
				output.Reset()
				projectTable := ui.NewProjectListFromTable(handler.repos.Tasks, &output, os.Stdin, true)
				err := projectTable.Browse(ctx)
				if err != nil {
					t.Errorf("Static project list should succeed: %v", err)
				}
				if !strings.Contains(output.String(), "test-project") {
					t.Error("Output should contain test-project")
				}
			})
		})

		t.Run("tagListStaticMode", func(t *testing.T) {
			var output bytes.Buffer

			t.Run("lists tags", func(t *testing.T) {
				output.Reset()
				tagTable := ui.NewTagListFromTable(handler.repos.Tasks, &output, os.Stdin, true)
				err := tagTable.Browse(ctx)
				if err != nil {
					t.Errorf("Static tag list should succeed: %v", err)
				}
				if !strings.Contains(output.String(), "tag1") {
					t.Error("Output should contain tag1")
				}
			})
		})

		t.Run("contextListStaticMode", func(t *testing.T) {
			oldStdout := os.Stdout
			defer func() { os.Stdout = oldStdout }()

			r, w, _ := os.Pipe()
			os.Stdout = w

			outputChan := make(chan string, 1)
			go func() {
				var buf bytes.Buffer
				buf.ReadFrom(r)
				outputChan <- buf.String()
			}()

			t.Run("lists contexts with tasks", func(t *testing.T) {
				err := handler.listContextsStatic(ctx, false)
				w.Close()
				capturedOutput := <-outputChan

				if err != nil {
					t.Errorf("listContextsStatic should succeed: %v", err)
				}
				if !strings.Contains(capturedOutput, "test-context") {
					t.Error("Output should contain 'test-context' context")
				}
			})

			r, w, _ = os.Pipe()
			os.Stdout = w
			go func() {
				var buf bytes.Buffer
				buf.ReadFrom(r)
				outputChan <- buf.String()
			}()

			t.Run("lists contexts with todo.txt format", func(t *testing.T) {
				err := handler.listContextsStatic(ctx, true)
				w.Close()
				capturedOutput := <-outputChan

				if err != nil {
					t.Errorf("listContextsStatic with todoTxt should succeed: %v", err)
				}
				if !strings.Contains(capturedOutput, "@test-context") {
					t.Error("Output should contain '@test-context' in todo.txt format")
				}
			})

			t.Run("handles no contexts", func(t *testing.T) {
				_, cleanup2 := setupTaskTest(t)
				defer cleanup2()

				handler2, err := NewTaskHandler()
				if err != nil {
					t.Fatalf("Failed to create handler: %v", err)
				}
				defer handler2.Close()

				r, w, _ = os.Pipe()
				os.Stdout = w
				go func() {
					var buf bytes.Buffer
					buf.ReadFrom(r)
					outputChan <- buf.String()
				}()

				err = handler2.listContextsStatic(ctx, false)
				w.Close()
				capturedOutput := <-outputChan

				if err != nil {
					t.Errorf("listContextsStatic with no contexts should succeed: %v", err)
				}
				if !strings.Contains(capturedOutput, "No contexts found") {
					t.Error("Output should contain 'No contexts found'")
				}
			})

			t.Run("handles repository error", func(t *testing.T) {
				cancelCtx, cancel := context.WithCancel(ctx)
				cancel()

				err := handler.listContextsStatic(cancelCtx, false)
				if err == nil {
					t.Error("Expected error with cancelled context")
				}
				if !strings.Contains(err.Error(), "failed to list tasks for contexts") {
					t.Errorf("Expected specific error message, got: %v", err)
				}
			})

			t.Run("counts tasks per context correctly", func(t *testing.T) {
				r, w, _ = os.Pipe()
				os.Stdout = w
				go func() {
					var buf bytes.Buffer
					buf.ReadFrom(r)
					outputChan <- buf.String()
				}()

				err := handler.listContextsStatic(ctx, false)
				w.Close()
				capturedOutput := <-outputChan

				if err != nil {
					t.Errorf("listContextsStatic should succeed: %v", err)
				}

				if !strings.Contains(capturedOutput, "test-context (2 tasks)") {
					t.Error("Output should show correct count for test-context context")
				}
			})
		})
	})
}
