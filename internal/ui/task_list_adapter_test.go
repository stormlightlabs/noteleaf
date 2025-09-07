package ui

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stormlightlabs/noteleaf/internal/models"
	"github.com/stormlightlabs/noteleaf/internal/repo"
)

type mockTaskRepository struct {
	tasks []*models.Task
	err   error
}

func (m *mockTaskRepository) List(ctx context.Context, options repo.TaskListOptions) ([]*models.Task, error) {
	if m.err != nil {
		return nil, m.err
	}

	var filtered []*models.Task
	for _, task := range m.tasks {
		if options.Status == "pending" && task.Status == "completed" {
			continue
		} else if options.Status != "" && options.Status != "pending" && task.Status != options.Status {
			continue
		}

		if options.Priority != "" && task.Priority != options.Priority {
			continue
		}

		if options.Project != "" && task.Project != options.Project {
			continue
		}

		filtered = append(filtered, task)

		if options.Limit > 0 && len(filtered) >= options.Limit {
			break
		}
	}

	return filtered, nil
}

func (m *mockTaskRepository) Update(ctx context.Context, task *models.Task) error {
	if m.err != nil {
		return m.err
	}
	for i, t := range m.tasks {
		if t.ID == task.ID {
			m.tasks[i] = task
			break
		}
	}
	return nil
}

func TestTaskAdapter(t *testing.T) {
	now := time.Now()
	task := &models.Task{
		ID:          1,
		UUID:        "test-uuid-123",
		Description: "Test task",
		Status:      "todo",
		Priority:    "high",
		Project:     "work",
		Tags:        []string{"urgent", "review"},
		Entry:       now,
		Modified:    now.Add(time.Hour),
		Annotations: []string{"First note", "Second note"},
	}

	t.Run("TaskRecord", func(t *testing.T) {
		record := &TaskRecord{Task: task}

		t.Run("GetField", func(t *testing.T) {
			tests := []struct {
				field    string
				expected any
				name     string
			}{
				{"id", int64(1), "should return task ID"},
				{"uuid", "test-uuid-123", "should return task UUID"},
				{"description", "Test task", "should return task description"},
				{"status", "todo", "should return task status"},
				{"priority", "high", "should return task priority"},
				{"project", "work", "should return task project"},
				{"tags", []string{"urgent", "review"}, "should return task tags"},
				{"entry", now, "should return entry time"},
				{"modified", now.Add(time.Hour), "should return modified time"},
				{"annotations", []string{"First note", "Second note"}, "should return annotations"},
				{"unknown", "", "should return empty string for unknown field"},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					result := record.GetField(tt.field)

					switch expected := tt.expected.(type) {
					case []string:
						resultSlice, ok := result.([]string)
						if !ok || len(resultSlice) != len(expected) {
							t.Errorf("GetField(%q) = %v, want %v", tt.field, result, tt.expected)
							return
						}
						for i, item := range expected {
							if resultSlice[i] != item {
								t.Errorf("GetField(%q) = %v, want %v", tt.field, result, tt.expected)
								return
							}
						}
					case time.Time:
						if resultTime, ok := result.(time.Time); !ok || !resultTime.Equal(expected) {
							t.Errorf("GetField(%q) = %v, want %v", tt.field, result, tt.expected)
						}
					default:
						if result != tt.expected {
							t.Errorf("GetField(%q) = %v, want %v", tt.field, result, tt.expected)
						}
					}
				})
			}
		})

		t.Run("Model interface", func(t *testing.T) {
			if record.GetID() != 1 {
				t.Errorf("GetID() = %d, want 1", record.GetID())
			}

			if record.GetTableName() != "tasks" {
				t.Errorf("GetTableName() = %q, want 'tasks'", record.GetTableName())
			}
		})
	})

	t.Run("TaskDataSource", func(t *testing.T) {
		tasks := []*models.Task{
			{
				ID:          1,
				Description: "Todo task",
				Status:      "todo",
				Priority:    "high",
				Project:     "work",
				Entry:       now,
				Modified:    now,
			},
			{
				ID:          2,
				Description: "In progress task",
				Status:      "in-progress",
				Priority:    "medium",
				Project:     "personal",
				Entry:       now,
				Modified:    now,
			},
			{
				ID:          3,
				Description: "Completed task",
				Status:      "completed",
				Priority:    "low",
				Project:     "work",
				Entry:       now,
				Modified:    now,
			},
		}

		t.Run("Load", func(t *testing.T) {
			repo := &mockTaskRepository{tasks: tasks}
			source := &TaskDataSource{repo: repo, showAll: true}

			records, err := source.Load(context.Background(), DataOptions{})
			if err != nil {
				t.Fatalf("Load() failed: %v", err)
			}

			if len(records) != 3 {
				t.Errorf("Load() returned %d records, want 3", len(records))
			}

			if records[0].GetField("description") != "Todo task" {
				t.Errorf("First record description = %v, want 'Todo task'", records[0].GetField("description"))
			}
		})

		t.Run("Load with pending filter", func(t *testing.T) {
			repo := &mockTaskRepository{tasks: tasks}
			source := &TaskDataSource{repo: repo, showAll: false}

			records, err := source.Load(context.Background(), DataOptions{})
			if err != nil {
				t.Fatalf("Load() failed: %v", err)
			}

			if len(records) != 2 {
				t.Errorf("Load() with pending filter returned %d records, want 2", len(records))
			}
		})

		t.Run("Load with status filter", func(t *testing.T) {
			repo := &mockTaskRepository{tasks: tasks}
			source := &TaskDataSource{repo: repo, status: "completed"}

			records, err := source.Load(context.Background(), DataOptions{})
			if err != nil {
				t.Fatalf("Load() failed: %v", err)
			}

			if len(records) != 1 {
				t.Errorf("Load() with status filter returned %d records, want 1", len(records))
			}
			if records[0].GetField("status") != "completed" {
				t.Errorf("Filtered record status = %v, want 'completed'", records[0].GetField("status"))
			}
		})

		t.Run("Load with priority filter", func(t *testing.T) {
			repo := &mockTaskRepository{tasks: tasks}
			source := &TaskDataSource{repo: repo, priority: "high", showAll: true}

			records, err := source.Load(context.Background(), DataOptions{})
			if err != nil {
				t.Fatalf("Load() failed: %v", err)
			}

			if len(records) != 1 {
				t.Errorf("Load() with priority filter returned %d records, want 1", len(records))
			}
			if records[0].GetField("priority") != "high" {
				t.Errorf("Filtered record priority = %v, want 'high'", records[0].GetField("priority"))
			}
		})

		t.Run("Load with project filter", func(t *testing.T) {
			repo := &mockTaskRepository{tasks: tasks}
			source := &TaskDataSource{repo: repo, project: "work", showAll: true}

			records, err := source.Load(context.Background(), DataOptions{})
			if err != nil {
				t.Fatalf("Load() failed: %v", err)
			}

			if len(records) != 2 {
				t.Errorf("Load() with project filter returned %d records, want 2", len(records))
			}
			if records[0].GetField("project") != "work" {
				t.Errorf("Filtered record project = %v, want 'work'", records[0].GetField("project"))
			}
		})

		t.Run("Load error", func(t *testing.T) {
			testErr := fmt.Errorf("test error")
			repo := &mockTaskRepository{err: testErr}
			source := &TaskDataSource{repo: repo}

			_, err := source.Load(context.Background(), DataOptions{})
			if err != testErr {
				t.Errorf("Load() error = %v, want %v", err, testErr)
			}
		})

		t.Run("Count", func(t *testing.T) {
			repo := &mockTaskRepository{tasks: tasks}
			source := &TaskDataSource{repo: repo, showAll: true}

			count, err := source.Count(context.Background(), DataOptions{})
			if err != nil {
				t.Fatalf("Count() failed: %v", err)
			}

			if count != 3 {
				t.Errorf("Count() = %d, want 3", count)
			}
		})

		t.Run("Count error", func(t *testing.T) {
			testErr := fmt.Errorf("test error")
			repo := &mockTaskRepository{err: testErr}
			source := &TaskDataSource{repo: repo}

			_, err := source.Count(context.Background(), DataOptions{})
			if err != testErr {
				t.Errorf("Count() error = %v, want %v", err, testErr)
			}
		})
	})

	t.Run("NewTaskDataTable", func(t *testing.T) {
		repo := &mockTaskRepository{
			tasks: []*models.Task{
				{
					ID:          1,
					Description: "Test task",
					Status:      "todo",
					Priority:    "high",
					Project:     "work",
					Entry:       time.Now(),
					Modified:    time.Now(),
				},
			},
		}

		opts := DataTableOptions{
			Output: &bytes.Buffer{},
			Input:  strings.NewReader("q\n"),
			Static: true,
		}

		table := NewTaskDataTable(repo, opts, false, "", "", "")
		if table == nil {
			t.Fatal("NewTaskDataTable() returned nil")
		}

		err := table.Browse(context.Background())
		if err != nil {
			t.Errorf("Browse() failed: %v", err)
		}
	})

	t.Run("NewTaskListFromTable", func(t *testing.T) {
		repo := &mockTaskRepository{
			tasks: []*models.Task{
				{
					ID:          1,
					Description: "Test task",
					Status:      "todo",
					Priority:    "high",
					Project:     "work",
					Entry:       time.Now(),
					Modified:    time.Now(),
				},
			},
		}

		output := &bytes.Buffer{}
		input := strings.NewReader("q\n")

		table := NewTaskListFromTable(repo, output, input, true, false, "", "", "")
		if table == nil {
			t.Fatal("NewTaskListFromTable() returned nil")
		}

		err := table.Browse(context.Background())
		if err != nil {
			t.Errorf("Browse() failed: %v", err)
		}

		outputStr := output.String()
		if !strings.Contains(outputStr, "Tasks") {
			t.Error("Output should contain 'Tasks' title")
		}
		if !strings.Contains(outputStr, "Test task") {
			t.Error("Output should contain task description")
		}
		if !strings.Contains(outputStr, "pending only") {
			t.Error("Output should contain 'pending only' filter status")
		}
	})

	t.Run("Format Task For View", func(t *testing.T) {
		now := time.Now()
		due := now.Add(24 * time.Hour)
		start := now.Add(-time.Hour)
		end := now.Add(time.Hour)

		task := &models.Task{
			ID:          1,
			UUID:        "test-uuid-123",
			Description: "Test task description",
			Status:      "in-progress",
			Priority:    "high",
			Project:     "work",
			Tags:        []string{"urgent", "review"},
			Due:         &due,
			Entry:       now,
			Modified:    now.Add(30 * time.Minute),
			Start:       &start,
			End:         &end,
			Annotations: []string{"First note", "Second note"},
		}

		result := formatTaskForView(task)

		if !strings.Contains(result, "Task 1") {
			t.Error("Formatted view should contain task ID in title")
		}
		if !strings.Contains(result, "test-uuid-123") {
			t.Error("Formatted view should contain UUID")
		}
		if !strings.Contains(result, "Test task description") {
			t.Error("Formatted view should contain description")
		}
		if !strings.Contains(result, "in-progress") {
			t.Error("Formatted view should contain status")
		}
		if !strings.Contains(result, "high") {
			t.Error("Formatted view should contain priority")
		}
		if !strings.Contains(result, "work") {
			t.Error("Formatted view should contain project")
		}
		if !strings.Contains(result, "urgent, review") {
			t.Error("Formatted view should contain tags")
		}
		if !strings.Contains(result, "Due:") {
			t.Error("Formatted view should contain due date")
		}
		if !strings.Contains(result, "Started:") {
			t.Error("Formatted view should contain start time")
		}
		if !strings.Contains(result, "Completed:") {
			t.Error("Formatted view should contain end time")
		}
		if !strings.Contains(result, "First note") {
			t.Error("Formatted view should contain annotations")
		}
	})

	t.Run("Format Priority Field", func(t *testing.T) {
		tests := []struct {
			priority string
			name     string
			contains string // What the result should contain
		}{
			{"", "empty priority", "-"},
			{"high", "high priority", "High"},
			{"urgent", "urgent priority", "Urgent"},
			{"medium", "medium priority", "Medium"},
			{"low", "low priority", "Low"},
			{"unknown", "unknown priority", "Unknown"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := formatPriorityField(tt.priority)
				if !strings.Contains(result, tt.contains) {
					t.Errorf("formatPriorityField(%q) = %q, should contain %q", tt.priority, result, tt.contains)
				}
			})
		}
	})
}
