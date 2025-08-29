package repo

import (
	"context"
	"database/sql"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stormlightlabs/noteleaf/internal/models"
)

func createTaskTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create in-memory database: %v", err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	schema := `
		CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid TEXT UNIQUE NOT NULL,
			description TEXT NOT NULL,
			status TEXT DEFAULT 'pending',
			priority TEXT,
			project TEXT,
			tags TEXT,
			due DATETIME,
			entry DATETIME DEFAULT CURRENT_TIMESTAMP,
			modified DATETIME DEFAULT CURRENT_TIMESTAMP,
			end DATETIME,
			start DATETIME,
			annotations TEXT
		);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func createSampleTask() *models.Task {
	return &models.Task{
		UUID:        newUUID(),
		Description: "Test task",
		Status:      "pending",
		Priority:    "H",
		Project:     "test-project",
		Tags:        []string{"test", "important"},
		Annotations: []string{"This is a test", "Another annotation"},
	}
}

func newUUID() string {
	return uuid.New().String()
}

func TestTaskRepository(t *testing.T) {
	db := createTaskTestDB(t)
	repo := NewTaskRepository(db)
	ctx := context.Background()

	t.Run("Create Task", func(t *testing.T) {
		task := createSampleTask()

		id, err := repo.Create(ctx, task)
		if err != nil {
			t.Errorf("Failed to create task: %v", err)
		}

		if id == 0 {
			t.Error("Expected non-zero ID")
		}

		if task.ID != id {
			t.Errorf("Expected task ID to be set to %d, got %d", id, task.ID)
		}

		if task.Entry.IsZero() {
			t.Error("Expected Entry timestamp to be set")
		}
		if task.Modified.IsZero() {
			t.Error("Expected Modified timestamp to be set")
		}
	})

	t.Run("Get Task", func(t *testing.T) {
		original := createSampleTask()
		id, err := repo.Create(ctx, original)
		if err != nil {
			t.Fatalf("Failed to create task: %v", err)
		}

		retrieved, err := repo.Get(ctx, id)
		if err != nil {
			t.Errorf("Failed to get task: %v", err)
		}

		if retrieved.UUID != original.UUID {
			t.Errorf("Expected UUID %s, got %s", original.UUID, retrieved.UUID)
		}
		if retrieved.Description != original.Description {
			t.Errorf("Expected description %s, got %s", original.Description, retrieved.Description)
		}
		if retrieved.Status != original.Status {
			t.Errorf("Expected status %s, got %s", original.Status, retrieved.Status)
		}
		if retrieved.Priority != original.Priority {
			t.Errorf("Expected priority %s, got %s", original.Priority, retrieved.Priority)
		}
		if retrieved.Project != original.Project {
			t.Errorf("Expected project %s, got %s", original.Project, retrieved.Project)
		}

		if len(retrieved.Tags) != len(original.Tags) {
			t.Errorf("Expected %d tags, got %d", len(original.Tags), len(retrieved.Tags))
		}
		for i, tag := range original.Tags {
			if i < len(retrieved.Tags) && retrieved.Tags[i] != tag {
				t.Errorf("Expected tag %s, got %s", tag, retrieved.Tags[i])
			}
		}

		if len(retrieved.Annotations) != len(original.Annotations) {
			t.Errorf("Expected %d annotations, got %d", len(original.Annotations), len(retrieved.Annotations))
		}
	})

	t.Run("Update Task", func(t *testing.T) {
		task := createSampleTask()
		id, err := repo.Create(ctx, task)
		if err != nil {
			t.Fatalf("Failed to create task: %v", err)
		}

		task.Description = "Updated description"
		task.Status = "completed"
		task.Priority = "L"
		now := time.Now()
		task.End = &now

		err = repo.Update(ctx, task)
		if err != nil {
			t.Errorf("Failed to update task: %v", err)
		}

		updated, err := repo.Get(ctx, id)
		if err != nil {
			t.Fatalf("Failed to get updated task: %v", err)
		}

		if updated.Description != "Updated description" {
			t.Errorf("Expected updated description, got %s", updated.Description)
		}
		if updated.Status != "completed" {
			t.Errorf("Expected status completed, got %s", updated.Status)
		}
		if updated.Priority != "L" {
			t.Errorf("Expected priority L, got %s", updated.Priority)
		}
		if updated.End == nil {
			t.Error("Expected end time to be set")
		}
	})

	t.Run("Delete Task", func(t *testing.T) {
		task := createSampleTask()
		id, err := repo.Create(ctx, task)
		if err != nil {
			t.Fatalf("Failed to create task: %v", err)
		}

		err = repo.Delete(ctx, id)
		if err != nil {
			t.Errorf("Failed to delete task: %v", err)
		}

		_, err = repo.Get(ctx, id)
		if err == nil {
			t.Error("Expected error when getting deleted task")
		}
	})

	t.Run("List", func(t *testing.T) {
		tasks := []*models.Task{
			{UUID: newUUID(), Description: "Task 1", Status: "pending", Project: "proj1"},
			{UUID: newUUID(), Description: "Task 2", Status: "completed", Project: "proj1"},
			{UUID: newUUID(), Description: "Task 3", Status: "pending", Project: "proj2"},
		}

		for _, task := range tasks {
			_, err := repo.Create(ctx, task)
			if err != nil {
				t.Fatalf("Failed to create task: %v", err)
			}
		}

		t.Run("List All Tasks", func(t *testing.T) {
			results, err := repo.List(ctx, TaskListOptions{})
			if err != nil {
				t.Errorf("Failed to list tasks: %v", err)
			}

			if len(results) < 3 {
				t.Errorf("Expected at least 3 tasks, got %d", len(results))
			}
		})

		t.Run("List Tasks with Filter", func(t *testing.T) {
			results, err := repo.List(ctx, TaskListOptions{Status: "pending"})
			if err != nil {
				t.Errorf("Failed to list tasks: %v", err)
			}

			if len(results) < 2 {
				t.Errorf("Expected at least 2 pending tasks, got %d", len(results))
			}

			for _, task := range results {
				if task.Status != "pending" {
					t.Errorf("Expected pending status, got %s", task.Status)
				}
			}
		})

		t.Run("List Tasks with Limit", func(t *testing.T) {
			results, err := repo.List(ctx, TaskListOptions{Limit: 2})
			if err != nil {
				t.Errorf("Failed to list tasks: %v", err)
			}

			if len(results) != 2 {
				t.Errorf("Expected 2 tasks due to limit, got %d", len(results))
			}
		})

		t.Run("List Tasks with Search", func(t *testing.T) {
			results, err := repo.List(ctx, TaskListOptions{Search: "Task 1"})
			if err != nil {
				t.Errorf("Failed to list tasks: %v", err)
			}

			if len(results) != 1 {
				t.Errorf("Expected 1 task matching search, got %d", len(results))
			}

			if len(results) > 0 && results[0].Description != "Task 1" {
				t.Errorf("Expected 'Task 1', got %s", results[0].Description)
			}
		})
	})

	t.Run("Special Methods", func(t *testing.T) {
		task1 := &models.Task{UUID: newUUID(), Description: "Pending task", Status: "pending", Project: "test"}
		task2 := &models.Task{UUID: newUUID(), Description: "Completed task", Status: "completed", Project: "test"}
		task3 := &models.Task{UUID: newUUID(), Description: "Other project", Status: "pending", Project: "other"}

		for _, task := range []*models.Task{task1, task2, task3} {
			_, err := repo.Create(ctx, task)
			if err != nil {
				t.Fatalf("Failed to create task: %v", err)
			}
		}

		t.Run("GetPending", func(t *testing.T) {
			results, err := repo.GetPending(ctx)
			if err != nil {
				t.Errorf("Failed to get pending tasks: %v", err)
			}

			if len(results) < 2 {
				t.Errorf("Expected at least 2 pending tasks, got %d", len(results))
			}
		})

		t.Run("GetCompleted", func(t *testing.T) {
			results, err := repo.GetCompleted(ctx)
			if err != nil {
				t.Errorf("Failed to get completed tasks: %v", err)
			}

			if len(results) < 1 {
				t.Errorf("Expected at least 1 completed task, got %d", len(results))
			}
		})

		t.Run("GetByProject", func(t *testing.T) {
			results, err := repo.GetByProject(ctx, "test")
			if err != nil {
				t.Errorf("Failed to get tasks by project: %v", err)
			}

			if len(results) < 2 {
				t.Errorf("Expected at least 2 tasks in test project, got %d", len(results))
			}

			for _, task := range results {
				if task.Project != "test" {
					t.Errorf("Expected project 'test', got %s", task.Project)
				}
			}
		})

		t.Run("GetByUUID", func(t *testing.T) {
			result, err := repo.GetByUUID(ctx, task1.UUID)
			if err != nil {
				t.Errorf("Failed to get task by UUID: %v", err)
			}

			if result.UUID != task1.UUID {
				t.Errorf("Expected UUID %s, got %s", task1.UUID, result.UUID)
			}
			if result.Description != task1.Description {
				t.Errorf("Expected description %s, got %s", task1.Description, result.Description)
			}
		})
	})

	t.Run("Count", func(t *testing.T) {
		tasks := []*models.Task{
			{UUID: newUUID(), Description: "Test 1", Status: "pending"},
			{UUID: newUUID(), Description: "Test 2", Status: "pending"},
			{UUID: newUUID(), Description: "Test 3", Status: "completed"},
		}

		for _, task := range tasks {
			_, err := repo.Create(ctx, task)
			if err != nil {
				t.Fatalf("Failed to create task: %v", err)
			}
		}

		t.Run("Count all tasks", func(t *testing.T) {
			count, err := repo.Count(ctx, TaskListOptions{})
			if err != nil {
				t.Errorf("Failed to count tasks: %v", err)
			}

			if count < 3 {
				t.Errorf("Expected at least 3 tasks, got %d", count)
			}
		})

		t.Run("Count pending tasks", func(t *testing.T) {
			count, err := repo.Count(ctx, TaskListOptions{Status: "pending"})
			if err != nil {
				t.Errorf("Failed to count pending tasks: %v", err)
			}

			if count < 2 {
				t.Errorf("Expected at least 2 pending tasks, got %d", count)
			}
		})

		t.Run("Count completed tasks", func(t *testing.T) {
			count, err := repo.Count(ctx, TaskListOptions{Status: "completed"})
			if err != nil {
				t.Errorf("Failed to count completed tasks: %v", err)
			}

			if count < 1 {
				t.Errorf("Expected at least 1 completed task, got %d", count)
			}
		})
	})

	t.Run("Projects & Tags", func(t *testing.T) {
		tasks := []*models.Task{
			{UUID: newUUID(), Description: "Task 1", Status: "pending", Project: "web-app", Tags: []string{"frontend", "urgent"}},
			{UUID: newUUID(), Description: "Task 2", Status: "pending", Project: "web-app", Tags: []string{"backend", "database"}},
			{UUID: newUUID(), Description: "Task 3", Status: "completed", Project: "mobile-app", Tags: []string{"frontend", "ios"}},
			{UUID: newUUID(), Description: "Task 4", Status: "pending", Project: "mobile-app", Tags: []string{"android", "urgent"}},
			{UUID: newUUID(), Description: "Task 5", Status: "pending", Project: "", Tags: []string{"documentation"}},
		}

		for _, task := range tasks {
			_, err := repo.Create(ctx, task)
			if err != nil {
				t.Fatalf("Failed to create task: %v", err)
			}
		}

		t.Run("GetProjects", func(t *testing.T) {
			projects, err := repo.GetProjects(ctx)
			if err != nil {
				t.Errorf("Failed to get projects: %v", err)
			}

			expectedProjectCount := 0
			projectCounts := make(map[string]int)

			for _, project := range projects {
				if project.Name != "" {
					expectedProjectCount++
					projectCounts[project.Name] = project.TaskCount
				}
			}

			if expectedProjectCount < 2 {
				t.Errorf("Expected at least 2 projects, got %d", expectedProjectCount)
			}

			if count, exists := projectCounts["web-app"]; exists {
				if count < 2 {
					t.Errorf("Expected at least 2 tasks for web-app project, got %d", count)
				}
			} else {
				t.Error("Expected web-app project to exist")
			}

			if count, exists := projectCounts["mobile-app"]; exists {
				if count < 2 {
					t.Errorf("Expected at least 2 tasks for mobile-app project, got %d", count)
				}
			} else {
				t.Error("Expected mobile-app project to exist")
			}
		})

		t.Run("GetTags", func(t *testing.T) {
			tags, err := repo.GetTags(ctx)
			if err != nil {
				t.Errorf("Failed to get tags: %v", err)
			}

			tagCounts := make(map[string]int)
			for _, tag := range tags {
				tagCounts[tag.Name] = tag.TaskCount
			}

			expectedMinCounts := map[string]int{
				"android": 1, "backend": 1, "database": 1, "documentation": 1,
				"frontend": 2, "ios": 1, "urgent": 2,
			}

			for expectedTag, minCount := range expectedMinCounts {
				if count, exists := tagCounts[expectedTag]; exists {
					if count < minCount {
						t.Errorf("Expected at least %d tasks for tag %s, got %d", minCount, expectedTag, count)
					}
				} else {
					t.Errorf("Expected tag %s to exist", expectedTag)
				}
			}

			if len(tags) < len(expectedMinCounts) {
				t.Errorf("Expected at least %d tags, got %d", len(expectedMinCounts), len(tags))
			}
		})

		t.Run("GetTasksByTag", func(t *testing.T) {
			frontend, err := repo.GetTasksByTag(ctx, "frontend")
			if err != nil {
				t.Errorf("Failed to get tasks by tag: %v", err)
			}

			if len(frontend) < 2 {
				t.Errorf("Expected at least 2 tasks with frontend tag, got %d", len(frontend))
			}

			for _, task := range frontend {
				if !slices.Contains(task.Tags, "frontend") {
					t.Errorf("Task %s should have frontend tag", task.Description)
				}
			}

			urgent, err := repo.GetTasksByTag(ctx, "urgent")
			if err != nil {
				t.Errorf("Failed to get tasks by tag: %v", err)
			}

			if len(urgent) < 2 {
				t.Errorf("Expected at least 2 tasks with urgent tag, got %d", len(urgent))
			}

			for _, task := range urgent {
				if !slices.Contains(task.Tags, "urgent") {
					t.Errorf("Task %s should have urgent tag", task.Description)
				}
			}

			nonexistent, err := repo.GetTasksByTag(ctx, "nonexistent")
			if err != nil {
				t.Errorf("Failed to get tasks by nonexistent tag: %v", err)
			}

			if len(nonexistent) != 0 {
				t.Errorf("Expected 0 tasks with nonexistent tag, got %d", len(nonexistent))
			}
		})
	})

	t.Run("New Status Tracking Methods", func(t *testing.T) {
		statusTasks := []*models.Task{
			{UUID: newUUID(), Description: "Todo task", Status: models.StatusTodo, Project: "test"},
			{UUID: newUUID(), Description: "In progress task", Status: models.StatusInProgress, Project: "test"},
			{UUID: newUUID(), Description: "Blocked task", Status: models.StatusBlocked, Project: "test"},
			{UUID: newUUID(), Description: "Done task", Status: models.StatusDone, Project: "test"},
			{UUID: newUUID(), Description: "Abandoned task", Status: models.StatusAbandoned, Project: "test"},
		}

		for _, task := range statusTasks {
			_, err := repo.Create(ctx, task)
			if err != nil {
				t.Fatalf("Failed to create task: %v", err)
			}
		}

		t.Run("GetTodo", func(t *testing.T) {
			results, err := repo.GetTodo(ctx)
			if err != nil {
				t.Errorf("Failed to get todo tasks: %v", err)
			}

			if len(results) < 1 {
				t.Errorf("Expected at least 1 todo task, got %d", len(results))
			}

			for _, task := range results {
				if task.Status != models.StatusTodo {
					t.Errorf("Expected status %s, got %s", models.StatusTodo, task.Status)
				}
			}
		})

		t.Run("GetInProgress", func(t *testing.T) {
			results, err := repo.GetInProgress(ctx)
			if err != nil {
				t.Errorf("Failed to get in-progress tasks: %v", err)
			}

			if len(results) < 1 {
				t.Errorf("Expected at least 1 in-progress task, got %d", len(results))
			}

			for _, task := range results {
				if task.Status != models.StatusInProgress {
					t.Errorf("Expected status %s, got %s", models.StatusInProgress, task.Status)
				}
			}
		})

		t.Run("GetBlocked", func(t *testing.T) {
			results, err := repo.GetBlocked(ctx)
			if err != nil {
				t.Errorf("Failed to get blocked tasks: %v", err)
			}

			if len(results) < 1 {
				t.Errorf("Expected at least 1 blocked task, got %d", len(results))
			}

			for _, task := range results {
				if task.Status != models.StatusBlocked {
					t.Errorf("Expected status %s, got %s", models.StatusBlocked, task.Status)
				}
			}
		})

		t.Run("GetDone", func(t *testing.T) {
			results, err := repo.GetDone(ctx)
			if err != nil {
				t.Errorf("Failed to get done tasks: %v", err)
			}

			if len(results) < 1 {
				t.Errorf("Expected at least 1 done task, got %d", len(results))
			}

			for _, task := range results {
				if task.Status != models.StatusDone {
					t.Errorf("Expected status %s, got %s", models.StatusDone, task.Status)
				}
			}
		})

		t.Run("GetAbandoned", func(t *testing.T) {
			results, err := repo.GetAbandoned(ctx)
			if err != nil {
				t.Errorf("Failed to get abandoned tasks: %v", err)
			}

			if len(results) < 1 {
				t.Errorf("Expected at least 1 abandoned task, got %d", len(results))
			}

			for _, task := range results {
				if task.Status != models.StatusAbandoned {
					t.Errorf("Expected status %s, got %s", models.StatusAbandoned, task.Status)
				}
			}
		})
	})

	t.Run("Priority System Methods", func(t *testing.T) {
		priorityTasks := []*models.Task{
			{UUID: newUUID(), Description: "High priority task", Status: "pending", Priority: models.PriorityHigh},
			{UUID: newUUID(), Description: "Medium priority task", Status: "pending", Priority: models.PriorityMedium},
			{UUID: newUUID(), Description: "Low priority task", Status: "pending", Priority: models.PriorityLow},
			{UUID: newUUID(), Description: "Numeric 5 priority", Status: "pending", Priority: "5"},
			{UUID: newUUID(), Description: "Numeric 3 priority", Status: "pending", Priority: "3"},
			{UUID: newUUID(), Description: "Numeric 1 priority", Status: "pending", Priority: "1"},
			{UUID: newUUID(), Description: "Legacy A priority", Status: "pending", Priority: "A"},
			{UUID: newUUID(), Description: "Legacy B priority", Status: "pending", Priority: "B"},
			{UUID: newUUID(), Description: "No priority task", Status: "pending", Priority: ""},
		}

		for _, task := range priorityTasks {
			_, err := repo.Create(ctx, task)
			if err != nil {
				t.Fatalf("Failed to create task: %v", err)
			}
		}

		t.Run("GetHighPriority", func(t *testing.T) {
			results, err := repo.GetHighPriority(ctx)
			if err != nil {
				t.Errorf("Failed to get high priority tasks: %v", err)
			}

			if len(results) < 1 {
				t.Errorf("Expected at least 1 high priority task, got %d", len(results))
			}

			for _, task := range results {
				if task.Priority != models.PriorityHigh {
					t.Errorf("Expected priority %s, got %s", models.PriorityHigh, task.Priority)
				}
			}
		})

		t.Run("GetMediumPriority", func(t *testing.T) {
			results, err := repo.GetMediumPriority(ctx)
			if err != nil {
				t.Errorf("Failed to get medium priority tasks: %v", err)
			}

			if len(results) < 1 {
				t.Errorf("Expected at least 1 medium priority task, got %d", len(results))
			}

			for _, task := range results {
				if task.Priority != models.PriorityMedium {
					t.Errorf("Expected priority %s, got %s", models.PriorityMedium, task.Priority)
				}
			}
		})

		t.Run("GetLowPriority", func(t *testing.T) {
			results, err := repo.GetLowPriority(ctx)
			if err != nil {
				t.Errorf("Failed to get low priority tasks: %v", err)
			}

			if len(results) < 1 {
				t.Errorf("Expected at least 1 low priority task, got %d", len(results))
			}

			for _, task := range results {
				if task.Priority != models.PriorityLow {
					t.Errorf("Expected priority %s, got %s", models.PriorityLow, task.Priority)
				}
			}
		})

		t.Run("GetByPriority", func(t *testing.T) {
			// Test numeric priority
			results, err := repo.GetByPriority(ctx, "5")
			if err != nil {
				t.Errorf("Failed to get tasks by priority 5: %v", err)
			}

			if len(results) < 1 {
				t.Errorf("Expected at least 1 task with priority 5, got %d", len(results))
			}

			for _, task := range results {
				if task.Priority != "5" {
					t.Errorf("Expected priority 5, got %s", task.Priority)
				}
			}

			// Test legacy priority
			results, err = repo.GetByPriority(ctx, "A")
			if err != nil {
				t.Errorf("Failed to get tasks by priority A: %v", err)
			}

			if len(results) < 1 {
				t.Errorf("Expected at least 1 task with priority A, got %d", len(results))
			}

			for _, task := range results {
				if task.Priority != "A" {
					t.Errorf("Expected priority A, got %s", task.Priority)
				}
			}

			// Test empty priority - create a specific task with no priority for this test
			noPriorityTask := &models.Task{
				UUID:        newUUID(),
				Description: "No priority task for test",
				Status:      "pending",
				Priority:    "",
			}
			_, err = repo.Create(ctx, noPriorityTask)
			if err != nil {
				t.Fatalf("Failed to create no priority task: %v", err)
			}

			results, err = repo.GetByPriority(ctx, "")
			if err != nil {
				t.Errorf("Failed to get tasks with no priority: %v", err)
			}

			if len(results) < 1 {
				t.Errorf("Expected at least 1 task with no priority, got %d", len(results))
			}

			for _, task := range results {
				if task.Priority != "" {
					t.Errorf("Expected empty priority, got %s", task.Priority)
				}
			}
		})
	})

	t.Run("Summary Methods", func(t *testing.T) {
		summaryTasks := []*models.Task{
			{UUID: newUUID(), Description: "Summary task 1", Status: models.StatusTodo, Priority: models.PriorityHigh},
			{UUID: newUUID(), Description: "Summary task 2", Status: models.StatusTodo, Priority: models.PriorityMedium},
			{UUID: newUUID(), Description: "Summary task 3", Status: models.StatusInProgress, Priority: models.PriorityHigh},
			{UUID: newUUID(), Description: "Summary task 4", Status: models.StatusDone, Priority: models.PriorityLow},
			{UUID: newUUID(), Description: "Summary task 5", Status: models.StatusBlocked, Priority: ""},
			{UUID: newUUID(), Description: "Summary task 6", Status: models.StatusAbandoned, Priority: "5"},
		}

		for _, task := range summaryTasks {
			_, err := repo.Create(ctx, task)
			if err != nil {
				t.Fatalf("Failed to create summary task: %v", err)
			}
		}

		t.Run("GetStatusSummary", func(t *testing.T) {
			summary, err := repo.GetStatusSummary(ctx)
			if err != nil {
				t.Errorf("Failed to get status summary: %v", err)
			}

			if len(summary) == 0 {
				t.Error("Expected non-empty status summary")
			}

			// Check that we have expected statuses with counts
			expectedStatuses := []string{
				models.StatusTodo, models.StatusInProgress, models.StatusDone,
				models.StatusBlocked, models.StatusAbandoned,
			}

			for _, status := range expectedStatuses {
				if count, exists := summary[status]; exists {
					if count < 1 {
						t.Errorf("Expected at least 1 task with status %s, got %d", status, count)
					}
				} else {
					t.Errorf("Expected status %s in summary", status)
				}
			}

			if todoCount := summary[models.StatusTodo]; todoCount < 2 {
				t.Errorf("Expected at least 2 todo tasks, got %d", todoCount)
			}
		})

		t.Run("GetPrioritySummary", func(t *testing.T) {
			summary, err := repo.GetPrioritySummary(ctx)
			if err != nil {
				t.Errorf("Failed to get priority summary: %v", err)
			}

			if len(summary) == 0 {
				t.Error("Expected non-empty priority summary")
			}

			expectedPriorities := []string{models.PriorityHigh, models.PriorityMedium, models.PriorityLow, "5"}

			for _, priority := range expectedPriorities {
				if count, exists := summary[priority]; exists {
					if count < 1 {
						t.Errorf("Expected at least 1 task with priority %s, got %d", priority, count)
					}
				} else {
					t.Errorf("Expected priority %s in summary", priority)
				}
			}

			if noPriorityCount, exists := summary["No Priority"]; exists {
				if noPriorityCount < 1 {
					t.Errorf("Expected at least 1 task with no priority, got %d", noPriorityCount)
				}
			} else {
				t.Error("Expected 'No Priority' group in summary")
			}

			if highCount := summary[models.PriorityHigh]; highCount < 2 {
				t.Errorf("Expected at least 2 high priority tasks, got %d", highCount)
			}
		})
	})
}
