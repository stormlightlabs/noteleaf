package repo

import (
	"context"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stormlightlabs/noteleaf/internal/models"
)

func newUUID() string {
	return uuid.New().String()
}

func TestTaskRepository(t *testing.T) {
	db := CreateTestDB(t)
	repo := NewTaskRepository(db)
	ctx := context.Background()

	t.Run("Create Task", func(t *testing.T) {
		task := CreateSampleTask()

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

		t.Run("Errors", func(t *testing.T) {
			t.Run("when called with duplicate UUID", func(t *testing.T) {
				task1 := CreateSampleTask()
				task1.UUID = "duplicate-test-uuid"

				_, err := repo.Create(ctx, task1)
				if err != nil {
					t.Fatalf("Failed to create first task: %v", err)
				}

				task2 := CreateSampleTask()
				task2.UUID = "duplicate-test-uuid"

				_, err = repo.Create(ctx, task2)
				if err == nil {
					t.Error("Expected error when creating task with duplicate UUID")
				}
			})

			t.Run("when called with context cancellation", func(t *testing.T) {
				task := CreateSampleTask()
				_, err := repo.Create(NewCanceledContext(), task)
				AssertCancelledContext(t, err)
			})
		})
	})

	t.Run("Get Task", func(t *testing.T) {
		original := CreateSampleTask()
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
		if retrieved.Context != original.Context {
			t.Errorf("Expected context %s, got %s", original.Context, retrieved.Context)
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
		task := CreateSampleTask()
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

		t.Run("Update Task Error Cases", func(t *testing.T) {
			t.Run("when called with context cancellation", func(t *testing.T) {
				task := CreateSampleTask()
				_, err := repo.Create(ctx, task)
				AssertNoError(t, err, "Failed to create task")

				task.Description = "Updated"
				err = repo.Update(NewCanceledContext(), task)
				AssertCancelledContext(t, err)
			})
		})
	})

	t.Run("Delete Task", func(t *testing.T) {
		task := CreateSampleTask()
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

		t.Run("GetByUUID with invalid UUID", func(t *testing.T) {
			_, err := repo.GetByUUID(ctx, "invalid-uuid-format")
			if err == nil {
				t.Error("Expected error with invalid UUID format")
			}
		})

		t.Run("GetByUUID with non-existent UUID", func(t *testing.T) {
			nonExistentUUID := newUUID()
			_, err := repo.GetByUUID(ctx, nonExistentUUID)
			if err == nil {
				t.Error("Expected error with non-existent UUID")
			}
		})

		t.Run("GetByUUID with context cancellation", func(t *testing.T) {
			_, err := repo.GetByUUID(NewCanceledContext(), task1.UUID)
			AssertCancelledContext(t, err)
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

		t.Run("Count with context cancellation", func(t *testing.T) {
			_, err := repo.Count(NewCanceledContext(), TaskListOptions{})
			AssertCancelledContext(t, err)
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

	t.Run("Recurrence Fields", func(t *testing.T) {
		task := CreateSampleTask()
		task.Recur = "FREQ=DAILY"
		until := time.Now().Add(7 * 24 * time.Hour)
		task.Until = &until
		parent := newUUID()
		task.ParentUUID = &parent

		id, err := repo.Create(ctx, task)
		if err != nil {
			t.Fatalf("failed to create task with recurrence: %v", err)
		}

		retrieved, err := repo.Get(ctx, id)
		if err != nil {
			t.Fatalf("failed to get task with recurrence: %v", err)
		}

		if retrieved.Recur != "FREQ=DAILY" {
			t.Errorf("expected Recur=FREQ=DAILY, got %s", retrieved.Recur)
		}
		if retrieved.Until == nil || !retrieved.Until.Equal(until) {
			t.Errorf("expected Until=%v, got %v", until, retrieved.Until)
		}
		if retrieved.ParentUUID == nil || *retrieved.ParentUUID != parent {
			t.Errorf("expected ParentUUID=%s, got %v", parent, retrieved.ParentUUID)
		}
	})

	t.Run("Dependencies", func(t *testing.T) {
		parent := CreateSampleTask()
		child := CreateSampleTask()

		_, err := repo.Create(ctx, parent)
		if err != nil {
			t.Fatalf("failed to create parent: %v", err)
		}
		_, err = repo.Create(ctx, child)
		if err != nil {
			t.Fatalf("failed to create child: %v", err)
		}

		if err := repo.AddDependency(ctx, child.UUID, parent.UUID); err != nil {
			t.Fatalf("failed to add dependency: %v", err)
		}

		deps, err := repo.GetDependencies(ctx, child.UUID)
		if err != nil {
			t.Fatalf("failed to get dependencies: %v", err)
		}
		if len(deps) != 1 || deps[0] != parent.UUID {
			t.Errorf("expected child to depend on parent=%s, got %v", parent.UUID, deps)
		}

		dependents, err := repo.GetDependents(ctx, parent.UUID)
		if err != nil {
			t.Fatalf("failed to get dependents: %v", err)
		}
		if len(dependents) != 1 || dependents[0].UUID != child.UUID {
			t.Errorf("expected dependent to be child=%s, got %v", child.UUID, dependents)
		}

		if err := repo.RemoveDependency(ctx, child.UUID, parent.UUID); err != nil {
			t.Fatalf("failed to remove dependency: %v", err)
		}
		deps, _ = repo.GetDependencies(ctx, child.UUID)
		if len(deps) != 0 {
			t.Errorf("expected dependencies to be cleared, got %v", deps)
		}

		if err := repo.AddDependency(ctx, child.UUID, parent.UUID); err != nil {
			t.Fatalf("failed to re-add dependency: %v", err)
		}
		if err := repo.ClearDependencies(ctx, child.UUID); err != nil {
			t.Fatalf("failed to clear dependencies: %v", err)
		}
		deps, _ = repo.GetDependencies(ctx, child.UUID)
		if len(deps) != 0 {
			t.Errorf("expected no dependencies after clear, got %v", deps)
		}
	})

	t.Run("Error Paths", func(t *testing.T) {
		t.Run("Create fails on MarshalTags error", func(t *testing.T) {
			orig := marshalTaskTags
			marshalTaskTags = func(t *models.Task) (string, error) {
				return "", fmt.Errorf("marshal fail")
			}
			defer func() { marshalTaskTags = orig }()

			_, err := repo.Create(ctx, CreateSampleTask())
			AssertError(t, err, "expected MarshalTags error")
			AssertContains(t, err.Error(), "failed to marshal tags", "error message")
		})

		t.Run("Create fails on MarshalAnnotations error", func(t *testing.T) {
			orig := marshalTaskAnnotations
			marshalTaskAnnotations = func(t *models.Task) (string, error) {
				return "", fmt.Errorf("marshal fail")
			}
			defer func() { marshalTaskAnnotations = orig }()

			_, err := repo.Create(ctx, CreateSampleTask())
			AssertError(t, err, "expected MarshalAnnotations error")
			AssertContains(t, err.Error(), "failed to marshal annotations", "error message")
		})

		t.Run("Update fails on MarshalTags error", func(t *testing.T) {
			task := CreateSampleTask()
			id, err := repo.Create(ctx, task)
			AssertNoError(t, err, "create should succeed")

			orig := marshalTaskTags
			marshalTaskTags = func(t *models.Task) (string, error) {
				return "", fmt.Errorf("marshal fail")
			}
			defer func() { marshalTaskTags = orig }()

			task.ID = id
			err = repo.Update(ctx, task)
			AssertError(t, err, "expected MarshalTags error")
			AssertContains(t, err.Error(), "failed to marshal tags", "error message")
		})

		t.Run("Update fails on MarshalAnnotations error", func(t *testing.T) {
			task := CreateSampleTask()
			id, err := repo.Create(ctx, task)
			AssertNoError(t, err, "create should succeed")

			orig := marshalTaskAnnotations
			marshalTaskAnnotations = func(t *models.Task) (string, error) {
				return "", fmt.Errorf("marshal fail")
			}
			defer func() { marshalTaskAnnotations = orig }()

			task.ID = id
			err = repo.Update(ctx, task)
			AssertError(t, err, "expected MarshalAnnotations error")
			AssertContains(t, err.Error(), "failed to marshal annotations", "error message")
		})

		t.Run("Get fails on UnmarshalTags error", func(t *testing.T) {
			task := CreateSampleTask()
			task.Tags = []string{"test"}
			id, err := repo.Create(ctx, task)
			AssertNoError(t, err, "create should succeed")

			orig := unmarshalTaskTags
			unmarshalTaskTags = func(t *models.Task, s string) error {
				return fmt.Errorf("unmarshal fail")
			}
			defer func() { unmarshalTaskTags = orig }()

			_, err = repo.Get(ctx, id)
			AssertError(t, err, "expected UnmarshalTags error")
			AssertContains(t, err.Error(), "failed to unmarshal tags", "error message")
		})

		t.Run("Get fails on UnmarshalAnnotations error", func(t *testing.T) {
			task := CreateSampleTask()
			task.Annotations = []string{"test"}
			id, err := repo.Create(ctx, task)
			AssertNoError(t, err, "create should succeed")

			orig := unmarshalTaskAnnotations
			unmarshalTaskAnnotations = func(t *models.Task, s string) error {
				return fmt.Errorf("unmarshal fail")
			}
			defer func() { unmarshalTaskAnnotations = orig }()

			_, err = repo.Get(ctx, id)
			AssertError(t, err, "expected UnmarshalAnnotations error")
			AssertContains(t, err.Error(), "failed to unmarshal annotations", "error message")
		})

		t.Run("GetByUUID fails on UnmarshalTags error", func(t *testing.T) {
			task := CreateSampleTask()
			task.Tags = []string{"test"}
			_, err := repo.Create(ctx, task)
			AssertNoError(t, err, "create should succeed")

			orig := unmarshalTaskTags
			unmarshalTaskTags = func(t *models.Task, s string) error {
				return fmt.Errorf("unmarshal fail")
			}
			defer func() { unmarshalTaskTags = orig }()

			_, err = repo.GetByUUID(ctx, task.UUID)
			AssertError(t, err, "expected UnmarshalTags error")
			AssertContains(t, err.Error(), "failed to unmarshal tags", "error message")
		})

		t.Run("GetByUUID fails on UnmarshalAnnotations error", func(t *testing.T) {
			task := CreateSampleTask()
			task.Annotations = []string{"test"}
			_, err := repo.Create(ctx, task)
			AssertNoError(t, err, "create should succeed")

			orig := unmarshalTaskAnnotations
			unmarshalTaskAnnotations = func(t *models.Task, s string) error {
				return fmt.Errorf("unmarshal fail")
			}
			defer func() { unmarshalTaskAnnotations = orig }()

			_, err = repo.GetByUUID(ctx, task.UUID)
			AssertError(t, err, "expected UnmarshalAnnotations error")
			AssertContains(t, err.Error(), "failed to unmarshal annotations", "error message")
		})
	})

	t.Run("GetContexts", func(t *testing.T) {

		task1 := CreateSampleTask()
		task1.Context = "work"
		_, err := repo.Create(ctx, task1)
		if err != nil {
			t.Fatalf("Failed to create task1: %v", err)
		}

		task2 := CreateSampleTask()
		task2.Context = "home"
		_, err = repo.Create(ctx, task2)
		if err != nil {
			t.Fatalf("Failed to create task2: %v", err)
		}

		task3 := CreateSampleTask()
		task3.Context = "work"
		_, err = repo.Create(ctx, task3)
		if err != nil {
			t.Fatalf("Failed to create task3: %v", err)
		}

		task4 := CreateSampleTask()
		task4.Context = ""
		_, err = repo.Create(ctx, task4)
		if err != nil {
			t.Fatalf("Failed to create task4: %v", err)
		}

		contexts, err := repo.GetContexts(ctx)
		if err != nil {
			t.Fatalf("Failed to get contexts: %v", err)
		}

		if len(contexts) < 2 {
			t.Errorf("Expected at least 2 contexts, got %d", len(contexts))
		}

		expectedCounts := map[string]int{
			"home":         1,
			"work":         2,
			"test-context": 14,
		}

		for _, context := range contexts {
			expected, exists := expectedCounts[context.Name]
			if !exists {
				t.Errorf("Unexpected context: %s", context.Name)
			}
			if context.TaskCount < expected {
				t.Errorf("Expected at least %d tasks for context %s, got %d", expected, context.Name, context.TaskCount)
			}
		}
	})

	t.Run("GetByContext", func(t *testing.T) {
		task1 := NewTaskBuilder().WithContext("work").WithDescription("Work task 1").Build()
		_, err := repo.Create(ctx, task1)
		AssertNoError(t, err, "Failed to create task1")

		task2 := NewTaskBuilder().WithContext("home").WithDescription("Home task 1").Build()
		_, err = repo.Create(ctx, task2)
		AssertNoError(t, err, "Failed to create task2")

		task3 := NewTaskBuilder().WithContext("work").WithDescription("Work task 2").Build()
		_, err = repo.Create(ctx, task3)
		AssertNoError(t, err, "Failed to create task3")

		workTasks, err := repo.GetByContext(ctx, "work")
		if err != nil {
			t.Fatalf("Failed to get tasks by context: %v", err)
		}

		if len(workTasks) < 2 {
			t.Errorf("Expected at least 2 work tasks, got %d", len(workTasks))
		}

		for _, task := range workTasks {
			if task.Context != "work" {
				t.Errorf("Expected context 'work', got '%s'", task.Context)
			}
		}

		homeTasks, err := repo.GetByContext(ctx, "home")
		if err != nil {
			t.Fatalf("Failed to get tasks by context: %v", err)
		}
		if len(homeTasks) < 1 {
			t.Errorf("Expected at least 1 home task, got %d", len(homeTasks))
		}
		if homeTasks[0].Context != "home" {
			t.Errorf("Expected context 'home', got '%s'", homeTasks[0].Context)
		}
	})

	t.Run("GetBlockedTasks", func(t *testing.T) {
		blocker := CreateSampleTask()
		blocker.Description = "Blocker task"
		_, err := repo.Create(ctx, blocker)
		AssertNoError(t, err, "create blocker should succeed")

		blocked1 := CreateSampleTask()
		blocked1.Description = "Blocked task 1"
		blocked1.DependsOn = []string{blocker.UUID}
		_, err = repo.Create(ctx, blocked1)
		AssertNoError(t, err, "create blocked1 should succeed")

		blocked2 := CreateSampleTask()
		blocked2.Description = "Blocked task 2"
		blocked2.DependsOn = []string{blocker.UUID}
		_, err = repo.Create(ctx, blocked2)
		AssertNoError(t, err, "create blocked2 should succeed")

		independent := CreateSampleTask()
		independent.Description = "Independent task"
		_, err = repo.Create(ctx, independent)
		AssertNoError(t, err, "create independent should succeed")

		blockedTasks, err := repo.GetBlockedTasks(ctx, blocker.UUID)
		AssertNoError(t, err, "GetBlockedTasks should succeed")
		AssertEqual(t, 2, len(blockedTasks), "should find 2 blocked tasks")

		for _, task := range blockedTasks {
			AssertTrue(t, slices.Contains(task.DependsOn, blocker.UUID), "task should depend on blocker")
		}

		emptyBlocked, err := repo.GetBlockedTasks(ctx, independent.UUID)
		AssertNoError(t, err, "GetBlockedTasks for independent should succeed")
		AssertEqual(t, 0, len(emptyBlocked), "independent task should not block anything")
	})
}
