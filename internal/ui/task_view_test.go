package ui

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stormlightlabs/noteleaf/internal/models"
)

func createMockTask() *models.Task {
	now := time.Now()
	due := now.Add(24 * time.Hour)
	start := now.Add(-2 * time.Hour)

	return &models.Task{
		ID:          1,
		UUID:        "test-uuid-123",
		Description: "Test task description",
		Status:      "pending",
		Priority:    "high",
		Project:     "test-project",
		Tags:        []string{"urgent", "test"},
		Entry:       now.Add(-24 * time.Hour),
		Modified:    now.Add(-1 * time.Hour),
		Due:         &due,
		Start:       &start,
		Annotations: []string{"First annotation", "Second annotation"},
	}
}

func createCompletedMockTask() *models.Task {
	now := time.Now()
	task := createMockTask()
	task.Status = "completed"
	task.End = &now
	return task
}

func TestTaskView(t *testing.T) {
	t.Run("View Options", func(t *testing.T) {
		task := createMockTask()

		t.Run("default options", func(t *testing.T) {
			opts := TaskViewOptions{}
			tv := NewTaskView(task, opts)

			if tv.opts.Output == nil {
				t.Error("Output should default to os.Stdout")
			}
			if tv.opts.Input == nil {
				t.Error("Input should default to os.Stdin")
			}
			if tv.opts.Width != 80 {
				t.Errorf("Width should default to 80, got %d", tv.opts.Width)
			}
			if tv.opts.Height != 24 {
				t.Errorf("Height should default to 24, got %d", tv.opts.Height)
			}
		})

		t.Run("custom options", func(t *testing.T) {
			var buf bytes.Buffer
			opts := TaskViewOptions{
				Output: &buf,
				Static: true,
				Width:  100,
				Height: 30,
			}
			tv := NewTaskView(task, opts)

			if tv.opts.Output != &buf {
				t.Error("Custom output not set")
			}
			if !tv.opts.Static {
				t.Error("Static mode not set")
			}
			if tv.opts.Width != 100 {
				t.Error("Custom width not set")
			}
			if tv.opts.Height != 30 {
				t.Error("Custom height not set")
			}
		})
	})

	t.Run("New", func(t *testing.T) {
		task := createMockTask()

		t.Run("creates task view correctly", func(t *testing.T) {
			opts := TaskViewOptions{Width: 60, Height: 20}
			tv := NewTaskView(task, opts)

			if tv.task != task {
				t.Error("Task not set correctly")
			}
			if tv.opts.Width != 60 {
				t.Error("Width not set correctly")
			}
			if tv.opts.Height != 20 {
				t.Error("Height not set correctly")
			}
		})
	})

	t.Run("Static Mode", func(t *testing.T) {
		t.Run("basic task display", func(t *testing.T) {
			task := createMockTask()
			var buf bytes.Buffer

			tv := NewTaskView(task, TaskViewOptions{
				Output: &buf,
				Static: true,
			})

			err := tv.Show(context.Background())
			if err != nil {
				t.Fatalf("Show failed: %v", err)
			}

			output := buf.String()

			if !strings.Contains(output, "Task 1") {
				t.Error("Task title not displayed")
			}

			if !strings.Contains(output, "test-uuid-123") {
				t.Error("UUID not displayed")
			}
			if !strings.Contains(output, "Test task description") {
				t.Error("Description not displayed")
			}
			if !strings.Contains(output, "Pending") {
				t.Error("Status not displayed with title case")
			}
			if !strings.Contains(output, "High") {
				t.Error("Priority not displayed with title case")
			}
			if !strings.Contains(output, "test-project") {
				t.Error("Project not displayed")
			}
			if !strings.Contains(output, "urgent, test") {
				t.Error("Tags not displayed correctly")
			}

			if !strings.Contains(output, "Dates:") {
				t.Error("Dates section not displayed")
			}
			if !strings.Contains(output, "Created:") {
				t.Error("Created date not displayed")
			}
			if !strings.Contains(output, "Modified:") {
				t.Error("Modified date not displayed")
			}
			if !strings.Contains(output, "Due:") {
				t.Error("Due date not displayed")
			}
			if !strings.Contains(output, "Started:") {
				t.Error("Start date not displayed")
			}

			if !strings.Contains(output, "Annotations:") {
				t.Error("Annotations section not displayed")
			}
			if !strings.Contains(output, "First annotation") {
				t.Error("First annotation not displayed")
			}
			if !strings.Contains(output, "Second annotation") {
				t.Error("Second annotation not displayed")
			}
		})

		t.Run("completed task display", func(t *testing.T) {
			task := createCompletedMockTask()
			var buf bytes.Buffer

			tv := NewTaskView(task, TaskViewOptions{
				Output: &buf,
				Static: true,
			})

			err := tv.Show(context.Background())
			if err != nil {
				t.Fatalf("Show failed: %v", err)
			}

			output := buf.String()

			if !strings.Contains(output, "Completed") {
				t.Error("Completed status not displayed with title case")
			}
			if !strings.Contains(output, "Completed:") {
				t.Error("Completion date not displayed")
			}
		})

		t.Run("minimal task display", func(t *testing.T) {
			now := time.Now()
			task := &models.Task{
				ID:          2,
				UUID:        "minimal-uuid",
				Description: "Minimal task",
				Status:      "pending",
				Entry:       now,
				Modified:    now,
			}

			var buf bytes.Buffer
			tv := NewTaskView(task, TaskViewOptions{
				Output: &buf,
				Static: true,
			})

			err := tv.Show(context.Background())
			if err != nil {
				t.Fatalf("Show failed: %v", err)
			}

			output := buf.String()

			if !strings.Contains(output, "Task 2") {
				t.Error("Task title not displayed")
			}
			if !strings.Contains(output, "minimal-uuid") {
				t.Error("UUID not displayed")
			}
			if !strings.Contains(output, "Minimal task") {
				t.Error("Description not displayed")
			}

			// Should not contain optional fields
			if strings.Contains(output, "Priority:") {
				t.Error("Priority should not be displayed for minimal task")
			}
			if strings.Contains(output, "Project:") {
				t.Error("Project should not be displayed for minimal task")
			}
			if strings.Contains(output, "Tags:") {
				t.Error("Tags should not be displayed for minimal task")
			}
			if strings.Contains(output, "Annotations:") {
				t.Error("Annotations should not be displayed for minimal task")
			}
		})
	})

	t.Run("Format Content", func(t *testing.T) {
		t.Run("formats task content correctly", func(t *testing.T) {
			task := createMockTask()
			content := formatTaskContent(task)

			expectedStrings := []string{
				"UUID: test-uuid-123",
				"Description: Test task description",
				"Status: Pending",
				"Priority: High",
				"Project: test-project",
				"Tags: urgent, test",
				"Dates:",
				"Created:",
				"Modified:",
				"Due:",
				"Started:",
				"Annotations:",
				"1. First annotation",
				"2. Second annotation",
			}

			for _, expected := range expectedStrings {
				if !strings.Contains(content, expected) {
					t.Errorf("Expected content '%s' not found in formatted output", expected)
				}
			}
		})

		t.Run("handles empty optional fields", func(t *testing.T) {
			now := time.Now()
			task := &models.Task{
				ID:          1,
				UUID:        "test-uuid",
				Description: "Test description",
				Status:      "pending",
				Entry:       now,
				Modified:    now,
			}

			content := formatTaskContent(task)

			if strings.Contains(content, "Priority:") {
				t.Error("Priority should not appear when empty")
			}
			if strings.Contains(content, "Project:") {
				t.Error("Project should not appear when empty")
			}
			if strings.Contains(content, "Tags:") {
				t.Error("Tags should not appear when empty")
			}
			if strings.Contains(content, "Annotations:") {
				t.Error("Annotations should not appear when empty")
			}
			if strings.Contains(content, "Due:") {
				t.Error("Due date should not appear when nil")
			}
			if strings.Contains(content, "Started:") {
				t.Error("Start date should not appear when nil")
			}
			if strings.Contains(content, "Completed:") {
				t.Error("End date should not appear when nil")
			}
		})
	})

	t.Run("Model", func(t *testing.T) {
		task := createMockTask()

		t.Run("initial model state", func(t *testing.T) {
			vp := viewport.New(80, 20)
			vp.SetContent(formatTaskContent(task))

			model := taskViewModel{task: task, opts: TaskViewOptions{Width: 80, Height: 24}}

			if model.showingHelp {
				t.Error("Initial showingHelp should be false")
			}
			if model.task != task {
				t.Error("Task not set correctly")
			}
		})

		t.Run("key handling - help toggle", func(t *testing.T) {
			vp := viewport.New(80, 20)
			model := taskViewModel{
				task:     task,
				viewport: vp,
				keys:     taskViewKeys,
				help:     help.New(),
			}

			newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
			if m, ok := newModel.(taskViewModel); ok {
				if !m.showingHelp {
					t.Error("Help key should show help")
				}
			}

			model.showingHelp = true
			newModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
			if m, ok := newModel.(taskViewModel); ok {
				if m.showingHelp {
					t.Error("Help key should exit help when already showing")
				}
			}
		})

		t.Run("key handling - quit and back", func(t *testing.T) {
			vp := viewport.New(80, 20)
			model := taskViewModel{
				task:     task,
				viewport: vp,
				keys:     taskViewKeys,
				help:     help.New(),
			}

			quitKeys := []string{"q", "esc"}
			for _, key := range quitKeys {
				_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
				if cmd == nil {
					t.Errorf("Key %s should return quit command", key)
				}
			}
		})

		t.Run("viewport navigation", func(t *testing.T) {
			vp := viewport.New(80, 20)
			longContent := strings.Repeat("Line of content\n", 50) // Create content longer than viewport
			vp.SetContent(longContent)

			model := taskViewModel{
				task:     task,
				viewport: vp,
				keys:     taskViewKeys,
				help:     help.New(),
			}

			initialOffset := model.viewport.YOffset

			newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
			if m, ok := newModel.(taskViewModel); ok {
				if m.viewport.YOffset <= initialOffset {
					t.Error("Down key should scroll viewport down")
				}
			}

			model.viewport.ScrollDown(5)
			initialOffset = model.viewport.YOffset
			newModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
			if m, ok := newModel.(taskViewModel); ok {
				if m.viewport.YOffset >= initialOffset {
					t.Error("Up key should scroll viewport up")
				}
			}
		})
	})

	t.Run("View Model", func(t *testing.T) {
		task := createMockTask()

		t.Run("normal view", func(t *testing.T) {
			vp := viewport.New(80, 20)
			vp.SetContent(formatTaskContent(task))

			model := taskViewModel{
				task:     task,
				viewport: vp,
				keys:     taskViewKeys,
				help:     help.New(),
			}

			view := model.View()

			if !strings.Contains(view, "Task 1") {
				t.Error("Task title not displayed in view")
			}
			if !strings.Contains(view, "help") {
				t.Error("Help information not displayed")
			}
		})

		t.Run("help view", func(t *testing.T) {
			vp := viewport.New(80, 20)
			model := taskViewModel{
				task:        task,
				viewport:    vp,
				keys:        taskViewKeys,
				help:        help.New(),
				showingHelp: true,
			}

			view := model.View()

			if !strings.Contains(view, "scroll") {
				t.Error("Help view should contain scroll instructions")
			}
		})
	})

	t.Run("Integration", func(t *testing.T) {
		t.Run("creates and displays task view", func(t *testing.T) {
			task := createMockTask()
			var buf bytes.Buffer

			tv := NewTaskView(task, TaskViewOptions{
				Output: &buf,
				Static: true,
				Width:  80,
				Height: 24,
			})

			if tv == nil {
				t.Fatal("NewTaskView returned nil")
			}

			err := tv.Show(context.Background())
			if err != nil {
				t.Fatalf("Show failed: %v", err)
			}

			output := buf.String()
			if len(output) == 0 {
				t.Error("No output generated")
			}

			if !strings.Contains(output, task.Description) {
				t.Error("Task description not displayed")
			}
			if !strings.Contains(output, task.UUID) {
				t.Error("Task UUID not displayed")
			}
		})
	})
}
