package repo

import (
	"context"
	"database/sql"
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
		UUID:        uuid.New().String(),
		Description: "Test task",
		Status:      "pending",
		Priority:    "H",
		Project:     "test-project",
		Tags:        []string{"test", "important"},
		Annotations: []string{"This is a test", "Another annotation"},
	}
}

func TestTaskRepository_CRUD(t *testing.T) {
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
}

func TestTaskRepository_List(t *testing.T) {
	db := createTaskTestDB(t)
	repo := NewTaskRepository(db)
	ctx := context.Background()

	tasks := []*models.Task{
		{UUID: uuid.New().String(), Description: "Task 1", Status: "pending", Project: "proj1"},
		{UUID: uuid.New().String(), Description: "Task 2", Status: "completed", Project: "proj1"},
		{UUID: uuid.New().String(), Description: "Task 3", Status: "pending", Project: "proj2"},
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

		if len(results) != 3 {
			t.Errorf("Expected 3 tasks, got %d", len(results))
		}
	})

	t.Run("List Tasks with Filter", func(t *testing.T) {
		results, err := repo.List(ctx, TaskListOptions{Status: "pending"})
		if err != nil {
			t.Errorf("Failed to list tasks: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("Expected 2 pending tasks, got %d", len(results))
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
}

func TestTaskRepository_SpecialMethods(t *testing.T) {
	db := createTaskTestDB(t)
	repo := NewTaskRepository(db)
	ctx := context.Background()

	task1 := &models.Task{UUID: uuid.New().String(), Description: "Pending task", Status: "pending", Project: "test"}
	task2 := &models.Task{UUID: uuid.New().String(), Description: "Completed task", Status: "completed", Project: "test"}
	task3 := &models.Task{UUID: uuid.New().String(), Description: "Other project", Status: "pending", Project: "other"}

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

		if len(results) != 2 {
			t.Errorf("Expected 2 pending tasks, got %d", len(results))
		}
	})

	t.Run("GetCompleted", func(t *testing.T) {
		results, err := repo.GetCompleted(ctx)
		if err != nil {
			t.Errorf("Failed to get completed tasks: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("Expected 1 completed task, got %d", len(results))
		}
	})

	t.Run("GetByProject", func(t *testing.T) {
		results, err := repo.GetByProject(ctx, "test")
		if err != nil {
			t.Errorf("Failed to get tasks by project: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("Expected 2 tasks in test project, got %d", len(results))
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
}

func TestTaskRepository_Count(t *testing.T) {
	db := createTaskTestDB(t)
	repo := NewTaskRepository(db)
	ctx := context.Background()

	tasks := []*models.Task{
		{UUID: uuid.New().String(), Description: "Test 1", Status: "pending"},
		{UUID: uuid.New().String(), Description: "Test 2", Status: "pending"},
		{UUID: uuid.New().String(), Description: "Test 3", Status: "completed"},
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

		if count != 3 {
			t.Errorf("Expected 3 tasks, got %d", count)
		}
	})

	t.Run("Count pending tasks", func(t *testing.T) {
		count, err := repo.Count(ctx, TaskListOptions{Status: "pending"})
		if err != nil {
			t.Errorf("Failed to count pending tasks: %v", err)
		}

		if count != 2 {
			t.Errorf("Expected 2 pending tasks, got %d", count)
		}
	})

	t.Run("Count completed tasks", func(t *testing.T) {
		count, err := repo.Count(ctx, TaskListOptions{Status: "completed"})
		if err != nil {
			t.Errorf("Failed to count completed tasks: %v", err)
		}

		if count != 1 {
			t.Errorf("Expected 1 completed task, got %d", count)
		}
	})
}
