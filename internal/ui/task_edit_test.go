package ui

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stormlightlabs/noteleaf/internal/models"
	"github.com/stormlightlabs/noteleaf/internal/repo"
)

type mockTaskRepo struct {
	tasks   map[int64]*models.Task
	updated []*models.Task
}

func (m *mockTaskRepo) List(ctx context.Context, opts repo.TaskListOptions) ([]*models.Task, error) {
	var result []*models.Task
	for _, task := range m.tasks {
		result = append(result, task)
	}
	return result, nil
}

func (m *mockTaskRepo) Update(ctx context.Context, task *models.Task) error {
	m.updated = append(m.updated, task)
	if existing, exists := m.tasks[task.ID]; exists {
		*existing = *task
	}
	return nil
}

func createTestTaskEditModel(task *models.Task) taskEditModel {
	now := time.Now()
	if task.Entry.IsZero() {
		task.Entry = now
	}
	if task.Modified.IsZero() {
		task.Modified = now
	}

	repo := &mockTaskRepo{tasks: map[int64]*models.Task{task.ID: task}}

	model := taskEditModel{
		task:         task,
		originalTask: task,
		repo:         repo,
		opts:         TaskEditOptions{Output: &bytes.Buffer{}, Width: 80, Height: 24},
		keys:         taskEditKeys,
		help:         help.New(),

		mode:         fieldNavigation,
		currentField: 0,
		priorityMode: priorityModeText,

		fields: []string{"Description", "Status", "Priority", "Project"},
	}

	model.descInput = textinput.New()
	model.descInput.SetValue(task.Description)
	model.projectInput = textinput.New()
	model.projectInput.SetValue(task.Project)

	for i, status := range statusOptions {
		if status == task.Status {
			model.statusIndex = i
			break
		}
	}

	model.updatePriorityIndex()

	return model
}

func TestTaskEditor(t *testing.T) {
	t.Run("Creation", func(t *testing.T) {
		task := &models.Task{
			ID:          1,
			Description: "Test task",
			Status:      models.StatusTodo,
			Priority:    models.PriorityHigh,
			Project:     "test-project",
		}

		repo := &mockTaskRepo{tasks: map[int64]*models.Task{1: task}}
		editor := NewTaskEditor(task, repo, TaskEditOptions{Width: 80, Height: 24})

		if editor.task != task {
			t.Error("Task should be set correctly")
		}

		if editor.repo != repo {
			t.Error("Repository should be set correctly")
		}

		if editor.opts.Width != 80 {
			t.Errorf("Expected width 80, got %d", editor.opts.Width)
		}
	})

	t.Run("Default Options", func(t *testing.T) {
		task := &models.Task{ID: 1}
		repo := &mockTaskRepo{}
		editor := NewTaskEditor(task, repo, TaskEditOptions{})

		if editor.opts.Width != 80 {
			t.Errorf("Expected default width 80, got %d", editor.opts.Width)
		}

		if editor.opts.Height != 24 {
			t.Errorf("Expected default height 24, got %d", editor.opts.Height)
		}
	})
}

func TestTaskEditModel(t *testing.T) {
	t.Run("Init", func(t *testing.T) {
		task := &models.Task{
			ID:          1,
			Description: "Test task",
			Status:      models.StatusInProgress,
			Priority:    models.PriorityMedium,
		}

		model := createTestTaskEditModel(task)
		cmd := model.Init()
		if cmd == nil {
			t.Error("Init should return a command")
		}
	})

	t.Run("Field Navigation", func(t *testing.T) {
		task := &models.Task{ID: 1, Description: "Test task", Status: models.StatusTodo}
		model := createTestTaskEditModel(task)

		if model.currentField != 0 {
			t.Errorf("Expected initial field 0, got %d", model.currentField)
		}

		msg := tea.KeyMsg{Type: tea.KeyDown}
		updatedModel, _ := model.Update(msg)
		model = updatedModel.(taskEditModel)

		if model.currentField != 1 {
			t.Errorf("Expected field 1 after down, got %d", model.currentField)
		}

		msg = tea.KeyMsg{Type: tea.KeyUp}
		updatedModel, _ = model.Update(msg)
		model = updatedModel.(taskEditModel)

		if model.currentField != 0 {
			t.Errorf("Expected field 0 after up, got %d", model.currentField)
		}
	})

	t.Run("Status Picker", func(t *testing.T) {
		task := &models.Task{ID: 1, Description: "Test task", Status: models.StatusTodo}
		model := createTestTaskEditModel(task)
		model.currentField = 1

		msg := tea.KeyMsg{Type: tea.KeyEnter}
		updatedModel, _ := model.Update(msg)
		model = updatedModel.(taskEditModel)

		if model.mode != statusPicker {
			t.Errorf("Expected statusPicker mode, got %d", model.mode)
		}

		msg = tea.KeyMsg{Type: tea.KeyDown}
		updatedModel, _ = model.Update(msg)
		model = updatedModel.(taskEditModel)

		if model.statusIndex != 1 {
			t.Errorf("Expected status index 1, got %d", model.statusIndex)
		}

		msg = tea.KeyMsg{Type: tea.KeyEnter}
		updatedModel, _ = model.Update(msg)
		model = updatedModel.(taskEditModel)

		if model.task.Status != statusOptions[1] {
			t.Errorf("Expected status %s, got %s", statusOptions[1], model.task.Status)
		}

		if model.mode != fieldNavigation {
			t.Errorf("Expected fieldNavigation mode after selection, got %d", model.mode)
		}
	})

	t.Run("Priority Picker", func(t *testing.T) {
		task := &models.Task{ID: 1, Description: "Test task", Priority: ""}
		model := createTestTaskEditModel(task)
		model.currentField = 2

		msg := tea.KeyMsg{Type: tea.KeyEnter}
		updatedModel, _ := model.Update(msg)
		model = updatedModel.(taskEditModel)

		if model.mode != priorityPicker {
			t.Errorf("Expected priorityPicker mode, got %d", model.mode)
		}

		msg = tea.KeyMsg{Type: tea.KeyDown}
		updatedModel, _ = model.Update(msg)
		model = updatedModel.(taskEditModel)

		if model.priorityIndex != 1 {
			t.Errorf("Expected priority index 1, got %d", model.priorityIndex)
		}

		msg = tea.KeyMsg{Type: tea.KeyEnter}
		updatedModel, _ = model.Update(msg)
		model = updatedModel.(taskEditModel)

		expectedPriority := textPriorityOptions[1]
		if model.task.Priority != expectedPriority {
			t.Errorf("Expected priority %s, got %s", expectedPriority, model.task.Priority)
		}
	})

	t.Run("Priority Mode Switch", func(t *testing.T) {
		task := &models.Task{ID: 1, Priority: models.PriorityHigh}
		model := createTestTaskEditModel(task)
		model.mode = priorityPicker

		if model.priorityMode != priorityModeText {
			t.Errorf("Expected text priority mode initially, got %d", model.priorityMode)
		}

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
		updatedModel, _ := model.Update(msg)
		model = updatedModel.(taskEditModel)

		if model.priorityMode != priorityModeNumeric {
			t.Errorf("Expected numeric priority mode, got %d", model.priorityMode)
		}

		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
		updatedModel, _ = model.Update(msg)
		model = updatedModel.(taskEditModel)

		if model.priorityMode != priorityModeLegacy {
			t.Errorf("Expected legacy priority mode, got %d", model.priorityMode)
		}

		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
		updatedModel, _ = model.Update(msg)
		model = updatedModel.(taskEditModel)

		if model.priorityMode != priorityModeText {
			t.Errorf("Expected text priority mode after full cycle, got %d", model.priorityMode)
		}
	})

	t.Run("TextInput", func(t *testing.T) {
		task := &models.Task{ID: 1, Description: "Original description", Project: "original-project"}

		model := createTestTaskEditModel(task)
		model.currentField = 0

		msg := tea.KeyMsg{Type: tea.KeyEnter}
		updatedModel, _ := model.Update(msg)
		model = updatedModel.(taskEditModel)

		if model.mode != textInput {
			t.Errorf("Expected textInput mode, got %d", model.mode)
		}

		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'X'}}
		updatedModel, _ = model.Update(msg)
		model = updatedModel.(taskEditModel)

		msg = tea.KeyMsg{Type: tea.KeyEnter}
		updatedModel, _ = model.Update(msg)
		model = updatedModel.(taskEditModel)

		if model.mode != fieldNavigation {
			t.Errorf("Expected fieldNavigation mode after text input, got %d", model.mode)
		}

		expected := "Original descriptionX"
		if model.task.Description != expected {
			t.Errorf("Expected description %s, got %s", expected, model.task.Description)
		}
	})

	t.Run("Help", func(t *testing.T) {
		task := &models.Task{ID: 1}
		model := createTestTaskEditModel(task)

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
		updatedModel, _ := model.Update(msg)
		model = updatedModel.(taskEditModel)

		if !model.showingHelp {
			t.Error("Expected help to be shown")
		}

		msg = tea.KeyMsg{Type: tea.KeyEsc}
		updatedModel, _ = model.Update(msg)
		model = updatedModel.(taskEditModel)

		if model.showingHelp {
			t.Error("Expected help to be hidden")
		}
	})

	t.Run("Save", func(t *testing.T) {
		task := &models.Task{ID: 1}
		model := createTestTaskEditModel(task)
		msg := tea.KeyMsg{Type: tea.KeyCtrlS}
		updatedModel, cmd := model.Update(msg)
		model = updatedModel.(taskEditModel)

		if !model.saved {
			t.Error("Expected saved flag to be set")
		}

		if cmd == nil {
			t.Error("Expected quit command after save")
		}
	})

	t.Run("Cancel", func(t *testing.T) {
		task := &models.Task{ID: 1}
		model := createTestTaskEditModel(task)
		msg := tea.KeyMsg{Type: tea.KeyCtrlC}
		updatedModel, cmd := model.Update(msg)
		model = updatedModel.(taskEditModel)

		if !model.cancelled {
			t.Error("Expected cancelled flag to be set")
		}

		if cmd == nil {
			t.Error("Expected quit command after cancel")
		}
	})

	t.Run("View", func(t *testing.T) {
		task := &models.Task{
			ID:          1,
			Description: "Test task",
			Status:      models.StatusTodo,
			Priority:    models.PriorityHigh,
			Project:     "test-project",
		}

		model := createTestTaskEditModel(task)
		view := model.View()

		if !strings.Contains(view, "Edit Task") {
			t.Error("View should contain title")
		}

		if !strings.Contains(view, "Test task") {
			t.Error("View should contain task description")
		}

		if !strings.Contains(view, "test-project") {
			t.Error("View should contain project")
		}
	})

	t.Run("Status Picker View", func(t *testing.T) {
		task := &models.Task{ID: 1, Status: models.StatusTodo}
		model := createTestTaskEditModel(task)
		model.mode = statusPicker

		view := model.View()

		if !strings.Contains(view, "Select Status:") {
			t.Error("Status picker should show selection prompt")
		}

		for _, status := range statusOptions {
			if !strings.Contains(view, status) {
				t.Errorf("Status picker should contain %s", status)
			}
		}
	})

	t.Run("Priority Picker View", func(t *testing.T) {
		task := &models.Task{ID: 1, Priority: ""}
		model := createTestTaskEditModel(task)
		model.mode = priorityPicker
		model.priorityMode = priorityModeText

		view := model.View()

		if !strings.Contains(view, "Select Priority") {
			t.Error("Priority picker should show selection prompt")
		}

		if !strings.Contains(view, "Text") {
			t.Error("Priority picker should show current mode")
		}
	})

	t.Run("KeyBindings", func(t *testing.T) {
		keyMap := taskEditKeys

		if keyMap.Up.Keys()[0] != "up" {
			t.Error("Up key binding should be defined")
		}

		if keyMap.StatusEdit.Keys()[0] != "s" {
			t.Error("Status edit key binding should be 's'")
		}

		if keyMap.Priority.Keys()[0] != "p" {
			t.Error("Priority key binding should be 'p'")
		}

		if keyMap.PriorityMode.Keys()[0] != "m" {
			t.Error("Priority mode key binding should be 'm'")
		}
	})
}

func TestUpdatePriorityIndex(t *testing.T) {
	testCases := []struct {
		priority    string
		mode        priorityMode
		expectedIdx int
	}{
		{models.PriorityHigh, priorityModeText, 3},
		{models.PriorityMedium, priorityModeText, 2},
		{models.PriorityLow, priorityModeText, 1},
		{"", priorityModeText, 0},
		{"3", priorityModeNumeric, 3},
		{"A", priorityModeLegacy, 1},
		{"unknown", priorityModeText, 0},
	}

	for _, tc := range testCases {
		task := &models.Task{ID: 1, Priority: tc.priority}
		model := createTestTaskEditModel(task)
		model.priorityMode = tc.mode
		model.updatePriorityIndex()

		if model.priorityIndex != tc.expectedIdx {
			t.Errorf("Priority %s in mode %d should have index %d, got %d",
				tc.priority, tc.mode, tc.expectedIdx, model.priorityIndex)
		}
	}
}

func TestRenderStatusField(t *testing.T) {
	task := &models.Task{ID: 1, Status: models.StatusInProgress}
	model := createTestTaskEditModel(task)

	result := model.renderStatusField()
	if !strings.Contains(result, models.StatusInProgress) {
		t.Error("Status field should contain the status")
	}

	model.mode = statusPicker
	result = model.renderStatusField()
	if !strings.Contains(result, models.StatusTodo) || !strings.Contains(result, models.StatusDone) {
		t.Error("Status picker should show status legend")
	}
}

func TestRenderPriorityField(t *testing.T) {
	task := &models.Task{ID: 1, Priority: models.PriorityMedium}
	model := createTestTaskEditModel(task)
	result := model.renderPriorityField()
	if !strings.Contains(result, models.PriorityMedium) {
		t.Error("Priority field should contain the priority")
	}

	model.mode = priorityPicker
	model.priorityMode = priorityModeNumeric
	result = model.renderPriorityField()
	if !strings.Contains(result, "Numeric") {
		t.Error("Priority picker should show current mode")
	}
}

// TestUncoveredPriorityModes tests all priority mode switch cases
func TestUncoveredPriorityModes(t *testing.T) {
	t.Run("Priority Mode Display Strings", func(t *testing.T) {
		task := &models.Task{ID: 1, Priority: models.PriorityHigh}
		model := createTestTaskEditModel(task)
		model.mode = priorityPicker

		// Test numeric mode
		model.priorityMode = priorityModeNumeric
		result := model.renderPriorityPicker()
		if !strings.Contains(result, "Numeric") {
			t.Error("Priority picker should show Numeric mode")
		}

		// Test legacy mode
		model.priorityMode = priorityModeLegacy
		result = model.renderPriorityPicker()
		if !strings.Contains(result, "Legacy") {
			t.Error("Priority picker should show Legacy mode")
		}
	})

	t.Run("Priority Display Type Switches", func(t *testing.T) {
		tests := []struct {
			name         string
			priority     string
			expectedType string
		}{
			{"Numeric Priority", "3", "numeric"},
			{"Legacy Priority A", "A", "legacy"},
			{"Legacy Priority E", "E", "legacy"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				task := &models.Task{ID: 1, Priority: tt.priority}

				displayType := GetPriorityDisplayType(task.Priority)
				if displayType != tt.expectedType {
					t.Errorf("Expected display type %s for priority %s, got %s (task:%v)", tt.expectedType, tt.priority, displayType, task)
				}
			})
		}
	})
}

func TestUncoveredKeyboardNavigation(t *testing.T) {
	t.Run("Status Edit Key Binding", func(t *testing.T) {
		task := &models.Task{ID: 1, Status: models.StatusTodo}
		model := createTestTaskEditModel(task)

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
		updatedModel, _ := model.Update(msg)
		model = updatedModel.(taskEditModel)

		if model.mode != statusPicker {
			t.Error("Expected status picker mode when 's' key is pressed")
		}
	})

	t.Run("Priority Edit Key Binding", func(t *testing.T) {
		task := &models.Task{ID: 1}
		model := createTestTaskEditModel(task)

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}}
		updatedModel, _ := model.Update(msg)
		model = updatedModel.(taskEditModel)

		if model.mode != priorityPicker {
			t.Error("Expected priority picker mode when 'p' key is pressed")
		}
	})

	t.Run("Window Resize Handling", func(t *testing.T) {
		task := &models.Task{ID: 1}
		model := createTestTaskEditModel(task)

		msg := tea.WindowSizeMsg{Width: 120, Height: 40}
		updatedModel, _ := model.Update(msg)
		model = updatedModel.(taskEditModel)

		if model.opts.Width != 120 {
			t.Errorf("Expected width to be updated to 120, got %d", model.opts.Width)
		}
	})

	t.Run("Escape Key in Status Picker", func(t *testing.T) {
		task := &models.Task{ID: 1}
		model := createTestTaskEditModel(task)
		model.mode = statusPicker

		msg := tea.KeyMsg{Type: tea.KeyEsc}
		updatedModel, _ := model.Update(msg)
		model = updatedModel.(taskEditModel)

		if model.mode != fieldNavigation {
			t.Error("Expected to return to field navigation when escape is pressed in status picker")
		}
	})

	t.Run("Escape Key in Priority Picker", func(t *testing.T) {
		task := &models.Task{ID: 1}
		model := createTestTaskEditModel(task)
		model.mode = priorityPicker

		msg := tea.KeyMsg{Type: tea.KeyEsc}
		updatedModel, _ := model.Update(msg)
		model = updatedModel.(taskEditModel)

		if model.mode != fieldNavigation {
			t.Error("Expected to return to field navigation when escape is pressed in priority picker")
		}
	})

	t.Run("Navigation Keys in Status Picker", func(t *testing.T) {
		task := &models.Task{ID: 1}
		model := createTestTaskEditModel(task)
		model.mode = statusPicker
		model.statusIndex = 0

		// Test down/right navigation
		msg := tea.KeyMsg{Type: tea.KeyDown}
		updatedModel, _ := model.Update(msg)
		model = updatedModel.(taskEditModel)

		if model.statusIndex != 1 {
			t.Errorf("Expected status index to be 1, got %d", model.statusIndex)
		}

		// Test up/left navigation
		msg = tea.KeyMsg{Type: tea.KeyUp}
		updatedModel, _ = model.Update(msg)
		model = updatedModel.(taskEditModel)

		if model.statusIndex != 0 {
			t.Errorf("Expected status index to be 0, got %d", model.statusIndex)
		}
	})

	t.Run("Navigation Keys in Priority Picker", func(t *testing.T) {
		task := &models.Task{ID: 1}
		model := createTestTaskEditModel(task)
		model.mode = priorityPicker
		model.priorityIndex = 0

		// Test down/right navigation
		msg := tea.KeyMsg{Type: tea.KeyDown}
		updatedModel, _ := model.Update(msg)
		model = updatedModel.(taskEditModel)

		if model.priorityIndex != 1 {
			t.Errorf("Expected priority index to be 1, got %d", model.priorityIndex)
		}

		// Test up/left navigation
		msg = tea.KeyMsg{Type: tea.KeyUp}
		updatedModel, _ = model.Update(msg)
		model = updatedModel.(taskEditModel)

		if model.priorityIndex != 0 {
			t.Errorf("Expected priority index to be 0, got %d", model.priorityIndex)
		}
	})
}

func TestUncoveredFieldSwitches(t *testing.T) {
	t.Run("Field Entry Switch Cases", func(t *testing.T) {
		task := &models.Task{ID: 1, Description: "Test", Project: "TestProject"}
		model := createTestTaskEditModel(task)

		model.currentField = 3
		model.mode = textInput

		msg := tea.KeyMsg{Type: tea.KeyEnter}
		updatedModel, _ := model.Update(msg)
		model = updatedModel.(taskEditModel)

		if model.mode != fieldNavigation {
			t.Error("Expected to return to field navigation after entering project field")
		}
	})

	t.Run("Field Update Switch Cases", func(t *testing.T) {
		task := &models.Task{ID: 1, Description: "Test", Project: "TestProject"}
		model := createTestTaskEditModel(task)
		model.currentField = 3
		model.mode = textInput

		model.projectInput.SetValue("Updated Project")

		msg := tea.KeyMsg{Type: tea.KeyEnter}
		updatedModel, _ := model.Update(msg)
		model = updatedModel.(taskEditModel)

		if model.task.Project != "Updated Project" {
			t.Errorf("Expected project to be updated to 'Updated Project', got %s", model.task.Project)
		}
	})
}

func TestTaskFieldAccessors(t *testing.T) {
	t.Run("Task Field Value Extraction", func(t *testing.T) {
		now := time.Now()
		task := &models.Task{
			ID:          1,
			Description: "Test",
			Tags:        []string{"tag1", "tag2"},
			Due:         &now,
			Entry:       now,
			Start:       &now,
			End:         &now,
		}

		tests := []struct {
			field    string
			expected any
		}{
			{"due", task.Due},
			{"entry", task.Entry},
			{"start", task.Start},
			{"end", task.End},
		}

		for _, tt := range tests {
			t.Run(tt.field, func(t *testing.T) {
				result := getTaskFieldValue(task, tt.field)
				if result != tt.expected {
					t.Errorf("Expected %v for field %s, got %v", tt.expected, tt.field, result)
				}
			})
		}
	})
}

func getTaskFieldValue(task *models.Task, field string) any {
	switch field {
	case "description":
		return task.Description
	case "status":
		return task.Status
	case "priority":
		return task.Priority
	case "project":
		return task.Project
	case "tags":
		return task.Tags
	case "due":
		return task.Due
	case "entry":
		return task.Entry
	case "start":
		return task.Start
	case "end":
		return task.End
	default:
		return nil
	}
}
