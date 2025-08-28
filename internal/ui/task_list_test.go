package ui

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stormlightlabs/noteleaf/internal/models"
	"github.com/stormlightlabs/noteleaf/internal/repo"
)

// MockTaskRepository implements TaskRepository interface for testing
type MockTaskRepository struct {
	tasks       []*models.Task
	listError   error
	updateError error
}

func (m *MockTaskRepository) List(ctx context.Context, opts repo.TaskListOptions) ([]*models.Task, error) {
	if m.listError != nil {
		return nil, m.listError
	}

	var filteredTasks []*models.Task
	for _, task := range m.tasks {
		// Apply filters
		if opts.Status != "" && task.Status != opts.Status {
			continue
		}
		if opts.Priority != "" && task.Priority != opts.Priority {
			continue
		}
		if opts.Project != "" && task.Project != opts.Project {
			continue
		}
		if opts.Search != "" && !strings.Contains(strings.ToLower(task.Description), strings.ToLower(opts.Search)) {
			continue
		}
		filteredTasks = append(filteredTasks, task)
	}

	// Apply limit
	if opts.Limit > 0 && len(filteredTasks) > opts.Limit {
		filteredTasks = filteredTasks[:opts.Limit]
	}

	return filteredTasks, nil
}

func (m *MockTaskRepository) Update(ctx context.Context, task *models.Task) error {
	if m.updateError != nil {
		return m.updateError
	}
	// Update the task in our mock data
	for i, t := range m.tasks {
		if t.ID == task.ID {
			m.tasks[i] = task
			break
		}
	}
	return nil
}

// Create mock tasks for testing
func createMockTasks() []*models.Task {
	now := time.Now()
	return []*models.Task{
		{
			ID:          1,
			UUID:        "uuid-1",
			Description: "Review quarterly report",
			Status:      "pending",
			Priority:    "high",
			Project:     "work",
			Tags:        []string{"urgent", "business"},
			Entry:       now.Add(-24 * time.Hour),
			Modified:    now.Add(-12 * time.Hour),
		},
		{
			ID:          2,
			UUID:        "uuid-2",
			Description: "Plan vacation itinerary",
			Status:      "pending",
			Priority:    "medium",
			Project:     "personal",
			Tags:        []string{"travel"},
			Entry:       now.Add(-48 * time.Hour),
			Modified:    now.Add(-6 * time.Hour),
		},
		{
			ID:          3,
			UUID:        "uuid-3",
			Description: "Fix authentication bug",
			Status:      "completed",
			Priority:    "high",
			Project:     "development",
			Tags:        []string{"bug", "security"},
			Entry:       now.Add(-72 * time.Hour),
			Modified:    now.Add(-1 * time.Hour),
			End:         &now,
		},
		{
			ID:          4,
			UUID:        "uuid-4",
			Description: "Read Clean Code book",
			Status:      "pending",
			Priority:    "low",
			Project:     "learning",
			Tags:        []string{"books", "development"},
			Entry:       now.Add(-96 * time.Hour),
			Modified:    now.Add(-3 * time.Hour),
		},
	}
}

func TestTaskListOptions(t *testing.T) {
	t.Run("default options", func(t *testing.T) {
		opts := TaskListOptions{}
		if opts.Static {
			t.Error("Static should default to false")
		}
		if opts.ShowAll {
			t.Error("ShowAll should default to false")
		}
	})

	t.Run("custom options", func(t *testing.T) {
		var buf bytes.Buffer
		opts := TaskListOptions{
			Output:   &buf,
			Static:   true,
			ShowAll:  true,
			Status:   "pending",
			Priority: "high",
			Project:  "work",
		}

		if !opts.Static {
			t.Error("Static should be enabled")
		}
		if !opts.ShowAll {
			t.Error("ShowAll should be enabled")
		}
		if opts.Output != &buf {
			t.Error("Output should be set to buffer")
		}
		if opts.Status != "pending" {
			t.Error("Status filter not set correctly")
		}
		if opts.Priority != "high" {
			t.Error("Priority filter not set correctly")
		}
		if opts.Project != "work" {
			t.Error("Project filter not set correctly")
		}
	})
}

func TestNewTaskList(t *testing.T) {
	repo := &MockTaskRepository{tasks: createMockTasks()}

	t.Run("with default options", func(t *testing.T) {
		tl := NewTaskList(repo, TaskListOptions{})
		if tl == nil {
			t.Fatal("NewTaskList returned nil")
		}
		if tl.repo != repo {
			t.Error("Repository not set correctly")
		}
		if tl.opts.Output == nil {
			t.Error("Output should default to os.Stdout")
		}
		if tl.opts.Input == nil {
			t.Error("Input should default to os.Stdin")
		}
	})

	t.Run("with custom options", func(t *testing.T) {
		var buf bytes.Buffer
		opts := TaskListOptions{
			Output:   &buf,
			Static:   true,
			ShowAll:  true,
			Priority: "high",
		}
		tl := NewTaskList(repo, opts)
		if tl.opts.Output != &buf {
			t.Error("Custom output not set")
		}
		if !tl.opts.Static {
			t.Error("Static mode not set")
		}
		if tl.opts.Priority != "high" {
			t.Error("Priority filter not set")
		}
	})
}

func TestTaskListStaticMode(t *testing.T) {
	t.Run("successful static list", func(t *testing.T) {
		repo := &MockTaskRepository{tasks: createMockTasks()}
		var buf bytes.Buffer

		tl := NewTaskList(repo, TaskListOptions{
			Output: &buf,
			Static: true,
		})

		err := tl.Browse(context.Background())
		if err != nil {
			t.Fatalf("Browse failed: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Tasks (pending only)") {
			t.Error("Title not displayed correctly")
		}
		if !strings.Contains(output, "Review quarterly report") {
			t.Error("First task not displayed")
		}
		if !strings.Contains(output, "Plan vacation itinerary") {
			t.Error("Second task not displayed")
		}
		// Should not show completed task by default
		if strings.Contains(output, "Fix authentication bug") {
			t.Error("Completed task should not be shown by default")
		}
	})

	t.Run("static list with all tasks", func(t *testing.T) {
		repo := &MockTaskRepository{tasks: createMockTasks()}
		var buf bytes.Buffer

		tl := NewTaskList(repo, TaskListOptions{
			Output:  &buf,
			Static:  true,
			ShowAll: true,
		})

		err := tl.Browse(context.Background())
		if err != nil {
			t.Fatalf("Browse failed: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Tasks (showing all)") {
			t.Error("All tasks title not displayed correctly")
		}
		if !strings.Contains(output, "Fix authentication bug") {
			t.Error("Completed task should be shown with --all")
		}
	})

	t.Run("static list with filters", func(t *testing.T) {
		repo := &MockTaskRepository{tasks: createMockTasks()}
		var buf bytes.Buffer

		tl := NewTaskList(repo, TaskListOptions{
			Output:   &buf,
			Static:   true,
			ShowAll:  true,
			Priority: "high",
		})

		err := tl.Browse(context.Background())
		if err != nil {
			t.Fatalf("Browse failed: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Review quarterly report") {
			t.Error("High priority task not displayed")
		}
		if !strings.Contains(output, "Fix authentication bug") {
			t.Error("High priority completed task not displayed")
		}
		if strings.Contains(output, "Plan vacation itinerary") {
			t.Error("Medium priority task should not be displayed")
		}
	})

	t.Run("static list with no results", func(t *testing.T) {
		repo := &MockTaskRepository{tasks: []*models.Task{}}
		var buf bytes.Buffer

		tl := NewTaskList(repo, TaskListOptions{
			Output: &buf,
			Static: true,
		})

		err := tl.Browse(context.Background())
		if err != nil {
			t.Fatalf("Browse failed: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "No tasks found") {
			t.Error("No tasks message not displayed")
		}
	})

	t.Run("static list with repository error", func(t *testing.T) {
		repo := &MockTaskRepository{
			listError: errors.New("database error"),
		}
		var buf bytes.Buffer

		tl := NewTaskList(repo, TaskListOptions{
			Output: &buf,
			Static: true,
		})

		err := tl.Browse(context.Background())
		if err == nil {
			t.Fatal("Expected error, got nil")
		}

		output := buf.String()
		if !strings.Contains(output, "Error: database error") {
			t.Error("Error message not displayed")
		}
	})
}

func TestTaskListModel(t *testing.T) {
	repo := &MockTaskRepository{tasks: createMockTasks()}

	t.Run("initial model state", func(t *testing.T) {
		model := taskListModel{
			opts: TaskListOptions{ShowAll: false},
		}

		if model.selected != 0 {
			t.Error("Initial selected should be 0")
		}
		if model.viewing {
			t.Error("Initial viewing should be false")
		}
		if model.showAll {
			t.Error("Initial showAll should be false")
		}
	})

	t.Run("load tasks command", func(t *testing.T) {
		model := taskListModel{
			keys: keys,
			help: help.New(),
			repo: repo,
			opts: TaskListOptions{ShowAll: false},
		}

		cmd := model.loadTasks()
		if cmd == nil {
			t.Fatal("loadTasks should return a command")
		}

		msg := cmd()
		switch msg := msg.(type) {
		case tasksLoadedMsg:
			tasks := []*models.Task(msg)
			if len(tasks) != 3 { // Only pending tasks
				t.Errorf("Expected 3 pending tasks, got %d", len(tasks))
			}
		case errorTaskMsg:
			t.Fatalf("Unexpected error: %v", error(msg))
		default:
			t.Fatalf("Unexpected message type: %T", msg)
		}
	})

	t.Run("load all tasks", func(t *testing.T) {
		model := taskListModel{
			keys:    keys,
			repo:    repo,
			opts:    TaskListOptions{ShowAll: true},
			showAll: true,
		}

		cmd := model.loadTasks()
		msg := cmd()

		switch msg := msg.(type) {
		case tasksLoadedMsg:
			tasks := []*models.Task(msg)
			if len(tasks) != 4 { // All tasks
				t.Errorf("Expected 4 tasks, got %d", len(tasks))
			}
		case errorTaskMsg:
			t.Fatalf("Unexpected error: %v", error(msg))
		}
	})

	t.Run("view task command", func(t *testing.T) {
		model := taskListModel{
			keys: keys,
			help: help.New(),
			repo: repo,
			opts: TaskListOptions{},
		}

		task := createMockTasks()[0]
		cmd := model.viewTask(task)
		if cmd == nil {
			t.Fatal("viewTask should return a command")
		}

		msg := cmd()
		switch msg := msg.(type) {
		case taskViewMsg:
			content := string(msg)
			if !strings.Contains(content, "# Task 1") {
				t.Error("Task title not in view content")
			}
			if !strings.Contains(content, "Review quarterly report") {
				t.Error("Task description not in view content")
			}
			if !strings.Contains(content, "**Status:** pending") {
				t.Error("Task status not in view content")
			}
			if !strings.Contains(content, "**Priority:** high") {
				t.Error("Task priority not in view content")
			}
		default:
			t.Fatalf("Unexpected message type: %T", msg)
		}
	})

	t.Run("mark done command", func(t *testing.T) {
		model := taskListModel{
			keys: keys,
			help: help.New(),
			repo: repo,
			opts: TaskListOptions{},
		}

		task := createMockTasks()[0] // Pending task
		cmd := model.markDone(task)
		if cmd == nil {
			t.Fatal("markDone should return a command")
		}

		msg := cmd()
		switch msg := msg.(type) {
		case tasksLoadedMsg:
			// Success - tasks reloaded
		case errorTaskMsg:
			err := error(msg)
			if !strings.Contains(err.Error(), "completed") {
				t.Fatalf("Unexpected error: %v", err)
			}
		default:
			t.Fatalf("Unexpected message type: %T", msg)
		}
	})

	t.Run("mark done already completed task", func(t *testing.T) {
		model := taskListModel{
			keys: keys,
			help: help.New(),
			repo: repo,
			opts: TaskListOptions{},
		}

		task := createMockTasks()[2]
		cmd := model.markDone(task)
		msg := cmd()

		switch msg := msg.(type) {
		case errorTaskMsg:
			err := error(msg)
			if !strings.Contains(err.Error(), "already completed") {
				t.Errorf("Expected 'already completed' error, got: %v", err)
			}
		default:
			t.Fatalf("Expected errorTaskMsg for already completed task, got: %T", msg)
		}
	})
}

func TestTaskListModelKeyHandling(t *testing.T) {
	repo := &MockTaskRepository{tasks: createMockTasks()}

	t.Run("quit commands", func(t *testing.T) {
		model := taskListModel{
			keys:  keys,
			repo:  repo,
			tasks: createMockTasks()[:2], // First 2 tasks
			opts:  TaskListOptions{},
		}

		quitKeys := []string{"ctrl+c", "q"}
		for _, key := range quitKeys {
			newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
			if cmd == nil {
				t.Errorf("Key %s should return quit command", key)
			}
			_ = newModel // Model should be returned
		}
	})

	t.Run("navigation keys", func(t *testing.T) {
		model := taskListModel{
			keys:     keys,
			repo:     repo,
			tasks:    createMockTasks()[:3], // First 3 tasks
			selected: 1,                     // Start in middle
			opts:     TaskListOptions{},
		}

		upKeys := []string{"up", "k"}
		for _, key := range upKeys {
			testModel := model
			newModel, _ := testModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
			if m, ok := newModel.(taskListModel); ok {
				if m.selected != 0 {
					t.Errorf("Key %s should move selection up to 0, got %d", key, m.selected)
				}
			}
		}

		downKeys := []string{"down", "j"}
		for _, key := range downKeys {
			testModel := model
			newModel, _ := testModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
			if m, ok := newModel.(taskListModel); ok {
				if m.selected != 2 {
					t.Errorf("Key %s should move selection down to 2, got %d", key, m.selected)
				}
			}
		}
	})

	t.Run("view task keys", func(t *testing.T) {
		model := taskListModel{
			keys:     keys,
			repo:     repo,
			tasks:    createMockTasks()[:2],
			selected: 0,
			opts:     TaskListOptions{},
		}

		viewKeys := []string{"enter", "v"}
		for _, key := range viewKeys {
			testModel := model
			_, cmd := testModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
			if cmd == nil {
				t.Errorf("Key %s should return view command", key)
			}
		}
	})

	t.Run("number shortcuts", func(t *testing.T) {
		model := taskListModel{
			keys:  keys,
			repo:  repo,
			tasks: createMockTasks()[:4],
			opts:  TaskListOptions{},
		}

		for i := 1; i <= 4; i++ {
			testModel := model
			key := fmt.Sprintf("%d", i)
			newModel, _ := testModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
			if m, ok := newModel.(taskListModel); ok {
				expectedIndex := i - 1
				if m.selected != expectedIndex {
					t.Errorf("Number key %s should select index %d, got %d", key, expectedIndex, m.selected)
				}
			}
		}
	})

	t.Run("toggle all/pending", func(t *testing.T) {
		model := taskListModel{
			keys:    keys,
			repo:    repo,
			tasks:   createMockTasks()[:2],
			showAll: false,
			opts:    TaskListOptions{},
		}

		newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
		if m, ok := newModel.(taskListModel); ok {
			if !m.showAll {
				t.Error("Key 'a' should toggle showAll to true")
			}
		}
		if cmd == nil {
			t.Error("Toggle all should trigger task reload")
		}
	})

	t.Run("mark done key", func(t *testing.T) {
		model := taskListModel{
			keys:     keys,
			repo:     repo,
			tasks:    createMockTasks()[:2],
			selected: 0,
			opts:     TaskListOptions{},
		}

		_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
		if cmd == nil {
			t.Error("Key 'd' should return mark done command")
		}
	})

	t.Run("refresh key", func(t *testing.T) {
		model := taskListModel{
			keys:  keys,
			repo:  repo,
			tasks: createMockTasks()[:2],
			opts:  TaskListOptions{},
		}

		_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
		if cmd == nil {
			t.Error("Key 'r' should return refresh command")
		}
	})

	t.Run("viewing mode navigation", func(t *testing.T) {
		model := taskListModel{
			keys:        keys,
			repo:        repo,
			tasks:       createMockTasks()[:2],
			viewing:     true,
			viewContent: "Test content",
			opts:        TaskListOptions{},
		}

		exitKeys := []string{"q", "esc", "backspace"}
		for _, key := range exitKeys {
			testModel := model
			newModel, _ := testModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
			if m, ok := newModel.(taskListModel); ok {
				if m.viewing {
					t.Errorf("Key %s should exit viewing mode", key)
				}
				if m.viewContent != "" {
					t.Errorf("Key %s should clear view content", key)
				}
			}
		}
	})
}

func TestTaskListModelView(t *testing.T) {
	repo := &MockTaskRepository{tasks: createMockTasks()}

	t.Run("viewing mode", func(t *testing.T) {
		model := taskListModel{
			keys:        keys,
			repo:        repo,
			viewing:     true,
			viewContent: "# Task Details\nTest content here",
			opts:        TaskListOptions{},
		}

		view := model.View()
		if !strings.Contains(view, "# Task Details") {
			t.Error("View content not displayed in viewing mode")
		}
		if !strings.Contains(view, "Press q/esc/backspace to return to list") {
			t.Error("Return instructions not displayed")
		}
	})

	t.Run("error state", func(t *testing.T) {
		model := taskListModel{
			keys: keys,
			help: help.New(),
			repo: repo,
			err:  errors.New("test error"),
			opts: TaskListOptions{},
		}

		view := model.View()
		if !strings.Contains(view, "Error: test error") {
			t.Error("Error message not displayed")
		}
	})

	t.Run("no tasks", func(t *testing.T) {
		model := taskListModel{
			keys:  keys,
			repo:  repo,
			tasks: []*models.Task{},
			opts:  TaskListOptions{},
		}

		view := model.View()
		if !strings.Contains(view, "No tasks found") {
			t.Error("No tasks message not displayed")
		}
		if !strings.Contains(view, "Press r to refresh, q to quit") {
			t.Error("Help text not displayed")
		}
	})

	t.Run("with tasks", func(t *testing.T) {
		tasks := createMockTasks()[:2] // First 2 tasks
		model := taskListModel{
			keys:     keys,
			repo:     repo,
			tasks:    tasks,
			selected: 0,
			showAll:  false,
			opts:     TaskListOptions{},
		}

		view := model.View()
		if !strings.Contains(view, "Tasks (pending only)") {
			t.Error("Title not displayed correctly")
		}
		if !strings.Contains(view, "Review quarterly report") {
			t.Error("First task not displayed")
		}
		if !strings.Contains(view, "Plan vacation itinerary") {
			t.Error("Second task not displayed")
		}
		if !strings.Contains(view, "help") {
			t.Error("Help instructions not displayed")
		}
	})

	t.Run("show all mode", func(t *testing.T) {
		model := taskListModel{
			keys:    keys,
			repo:    repo,
			tasks:   createMockTasks(),
			showAll: true,
			opts:    TaskListOptions{ShowAll: true},
		}

		view := model.View()
		if !strings.Contains(view, "Tasks (showing all)") {
			t.Error("Show all title not displayed correctly")
		}
	})

	t.Run("selected task highlighting", func(t *testing.T) {
		tasks := createMockTasks()[:2]
		model := taskListModel{

			keys:     keys,
			repo:     repo,
			tasks:    tasks,
			selected: 0,
			opts:     TaskListOptions{},
		}

		view := model.View()
		if !strings.Contains(view, " > 1   ") {
			t.Error("Selected task not highlighted with '>' prefix")
		}
	})
}

func TestTaskListModelUpdate(t *testing.T) {
	repo := &MockTaskRepository{tasks: createMockTasks()}

	t.Run("tasks loaded message", func(t *testing.T) {
		model := taskListModel{
			keys: keys,
			help: help.New(),
			repo: repo,
			opts: TaskListOptions{},
		}

		tasks := createMockTasks()[:2]
		newModel, _ := model.Update(tasksLoadedMsg(tasks))

		if m, ok := newModel.(taskListModel); ok {
			if len(m.tasks) != 2 {
				t.Errorf("Expected 2 tasks, got %d", len(m.tasks))
			}
			if m.tasks[0].Description != "Review quarterly report" {
				t.Error("Tasks not loaded correctly")
			}
		}
	})

	t.Run("task view message", func(t *testing.T) {
		model := taskListModel{
			keys: keys,
			help: help.New(),
			repo: repo,
			opts: TaskListOptions{},
		}

		content := "# Task Details\nTest content"
		newModel, _ := model.Update(taskViewMsg(content))

		if m, ok := newModel.(taskListModel); ok {
			if !m.viewing {
				t.Error("Viewing mode not activated")
			}
			if m.viewContent != content {
				t.Error("View content not set correctly")
			}
		}
	})

	t.Run("error message", func(t *testing.T) {
		model := taskListModel{
			keys: keys,
			help: help.New(),
			repo: repo,
			opts: TaskListOptions{},
		}

		testErr := errors.New("test error")
		newModel, _ := model.Update(errorTaskMsg(testErr))

		if m, ok := newModel.(taskListModel); ok {
			if m.err == nil {
				t.Error("Error not set")
			}
			if m.err.Error() != "test error" {
				t.Errorf("Expected 'test error', got %v", m.err)
			}
		}
	})

	t.Run("selected index bounds", func(t *testing.T) {
		model := taskListModel{
			keys:     keys,
			repo:     repo,
			tasks:    createMockTasks()[:2],
			selected: 5,
			opts:     TaskListOptions{},
		}

		newTasks := createMockTasks()[:1]
		newModel, _ := model.Update(tasksLoadedMsg(newTasks))

		if m, ok := newModel.(taskListModel); ok {
			if m.selected >= len(m.tasks) {
				t.Error("Selected index should be adjusted to bounds")
			}
			if m.selected != 0 {
				t.Errorf("Expected selected to be 0, got %d", m.selected)
			}
		}
	})
}

func TestTaskListRepositoryError(t *testing.T) {
	t.Run("list error in loadTasks", func(t *testing.T) {
		repo := &MockTaskRepository{
			listError: errors.New("database connection failed"),
		}

		model := taskListModel{
			keys: keys,
			help: help.New(),
			repo: repo,
			opts: TaskListOptions{},
		}

		cmd := model.loadTasks()
		msg := cmd()

		switch msg := msg.(type) {
		case errorTaskMsg:
			err := error(msg)
			if !strings.Contains(err.Error(), "database connection failed") {
				t.Errorf("Expected database error, got: %v", err)
			}
		default:
			t.Fatalf("Expected errorTaskMsg, got: %T", msg)
		}
	})

	t.Run("update error in markDone", func(t *testing.T) {
		repo := &MockTaskRepository{
			tasks:       createMockTasks(),
			updateError: errors.New("update failed"),
		}

		model := taskListModel{
			keys: keys,
			help: help.New(),
			repo: repo,
			opts: TaskListOptions{},
		}

		task := createMockTasks()[0]
		cmd := model.markDone(task)
		msg := cmd()

		switch msg := msg.(type) {
		case errorTaskMsg:
			err := error(msg)
			if !strings.Contains(err.Error(), "failed to mark task done") {
				t.Errorf("Expected mark done error, got: %v", err)
			}
		default:
			t.Fatalf("Expected errorTaskMsg, got: %T", msg)
		}
	})
}
