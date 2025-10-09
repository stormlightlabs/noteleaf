package ui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stormlightlabs/noteleaf/internal/models"
	"github.com/stormlightlabs/noteleaf/internal/shared"
)

func TestInteractiveTUIBehavior(t *testing.T) {
	t.Run("Priority Mode Switching with TUI Framework", func(t *testing.T) {
		task := &models.Task{ID: 1, Priority: models.PriorityHigh}
		model := createTestTaskEditModel(task)

		suite := NewTUITestSuite(t, model, WithInitialSize(80, 24))
		suite.Start()

		if err := suite.SendKeyString("p"); err != nil {
			t.Fatalf("Failed to send 'p' key: %v", err)
		}

		if err := suite.WaitFor(func(m tea.Model) bool {
			if taskModel, ok := m.(taskEditModel); ok {
				return taskModel.mode == priorityPicker
			}
			return false
		}, 1*time.Second); err != nil {
			t.Fatalf("Failed to enter priority picker mode: %v", err)
		}

		if err := suite.SendKeyString("m"); err != nil {
			t.Fatalf("Failed to send 'm' key: %v", err)
		}

		if err := suite.WaitFor(func(m tea.Model) bool {
			if taskModel, ok := m.(taskEditModel); ok {
				return taskModel.priorityMode == priorityModeNumeric
			}
			return false
		}, 1*time.Second); err != nil {
			t.Fatalf("Failed to switch to numeric priority mode: %v", err)
		}

		if err := suite.WaitForView("Numeric", 1*time.Second); err != nil {
			t.Errorf("Expected view to contain 'Numeric': %v", err)
		}

		if err := suite.SendKeyString("m"); err != nil {
			t.Fatalf("Failed to send second 'm' key: %v", err)
		}

		if err := suite.WaitFor(func(m tea.Model) bool {
			if taskModel, ok := m.(taskEditModel); ok {
				return taskModel.priorityMode == priorityModeLegacy
			}
			return false
		}, 1*time.Second); err != nil {
			t.Fatalf("Failed to switch to legacy priority mode: %v", err)
		}

		if err := suite.WaitForView("Legacy", 1*time.Second); err != nil {
			t.Errorf("Expected view to contain 'Legacy': %v", err)
		}
	})

	t.Run("Keyboard Navigation with TUI Framework", func(t *testing.T) {
		task := &models.Task{ID: 1, Status: models.StatusTodo}
		model := createTestTaskEditModel(task)

		suite := NewTUITestSuite(t, model)
		suite.Start()

		if err := suite.SendKeyString("s"); err != nil {
			t.Fatalf("Failed to send 's' key: %v", err)
		}

		if err := suite.WaitFor(func(m tea.Model) bool {
			if taskModel, ok := m.(taskEditModel); ok {
				return taskModel.mode == statusPicker
			}
			return false
		}, 1*time.Second); err != nil {
			t.Fatalf("Failed to enter status picker mode: %v", err)
		}

		if err := suite.SendKey(tea.KeyDown); err != nil {
			t.Fatalf("Failed to send down arrow: %v", err)
		}

		if err := suite.WaitFor(func(m tea.Model) bool {
			if taskModel, ok := m.(taskEditModel); ok {
				return taskModel.statusIndex == 1
			}
			return false
		}, 1*time.Second); err != nil {
			t.Fatalf("Status index should have changed to 1: %v", err)
		}

		if err := suite.SendKey(tea.KeyEsc); err != nil {
			t.Fatalf("Failed to send escape key: %v", err)
		}

		if err := suite.WaitFor(func(m tea.Model) bool {
			if taskModel, ok := m.(taskEditModel); ok {
				return taskModel.mode == fieldNavigation
			}
			return false
		}, 1*time.Second); err != nil {
			t.Fatalf("Should have returned to field navigation mode: %v", err)
		}
	})

	t.Run("Window Resize Handling", func(t *testing.T) {
		task := &models.Task{ID: 1}
		model := createTestTaskEditModel(task)

		suite := NewTUITestSuite(t, model)
		suite.Start()

		resizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
		if err := suite.SendMessage(resizeMsg); err != nil {
			t.Fatalf("Failed to send window resize message: %v", err)
		}

		if err := suite.WaitFor(func(m tea.Model) bool {
			if taskModel, ok := m.(taskEditModel); ok {
				return taskModel.opts.Width == 120
			}
			return false
		}, 1*time.Second); err != nil {
			t.Fatalf("Window width should have been updated to 120: %v", err)
		}
	})

	t.Run("Complex Key Sequence with TUI Framework", func(t *testing.T) {
		task := &models.Task{ID: 1, Description: "Test", Project: "TestProject"}
		model := createTestTaskEditModel(task)

		suite := NewTUITestSuite(t, model)
		suite.Start()

		keySequence := []KeyWithTiming{
			{KeyType: tea.KeyDown, Delay: 50 * time.Millisecond},
			{KeyType: tea.KeyDown, Delay: 50 * time.Millisecond},
			{KeyType: tea.KeyDown, Delay: 50 * time.Millisecond},
			{KeyType: tea.KeyEnter, Delay: 100 * time.Millisecond},
		}

		if err := suite.SimulateKeySequence(keySequence); err != nil {
			t.Fatalf("Failed to simulate key sequence: %v", err)
		}

		if err := suite.WaitFor(func(m tea.Model) bool {
			if taskModel, ok := m.(taskEditModel); ok {
				return taskModel.mode == textInput && taskModel.currentField == 3 // Project field
			}
			return false
		}, 2*time.Second); err != nil {
			t.Fatalf("Should have entered text input mode for project field: %v", err)
		}

		if err := suite.SendKeyString(" Updated"); err != nil {
			t.Fatalf("Failed to send text: %v", err)
		}

		if err := suite.SendKey(tea.KeyEnter); err != nil {
			t.Fatalf("Failed to send enter key: %v", err)
		}

		if err := suite.WaitFor(func(m tea.Model) bool {
			if taskModel, ok := m.(taskEditModel); ok {
				return taskModel.mode == fieldNavigation &&
					taskModel.task.Project == "TestProject Updated"
			}
			return false
		}, 1*time.Second); err != nil {
			t.Fatalf("Project should have been updated: %v", err)
		}
	})
}

func TestTUIFrameworkFeatures(t *testing.T) {
	t.Run("Output Capture", func(t *testing.T) {
		task := &models.Task{ID: 1, Description: "Test Output"}
		model := createTestTaskEditModel(task)

		suite := NewTUITestSuite(t, model)
		suite.Start()

		time.Sleep(100 * time.Millisecond)

		view := suite.GetCurrentView()
		if len(view) == 0 {
			t.Error("View should not be empty")
		}

		if !shared.ContainsString(view, "Test Output") {
			t.Error("View should contain task description")
		}
	})

	t.Run("Timeout Handling", func(t *testing.T) {
		task := &models.Task{ID: 1}
		model := createTestTaskEditModel(task)

		suite := NewTUITestSuite(t, model, WithTimeout(100*time.Millisecond))
		suite.Start()

		if err := suite.WaitFor(func(m tea.Model) bool {
			return false
		}, 50*time.Millisecond); err == nil {
			t.Error("Expected timeout error")
		}
	})

	t.Run("Multiple Assertions", func(t *testing.T) {
		task := &models.Task{ID: 1, Description: "Test Task", Status: models.StatusTodo}
		model := createTestTaskEditModel(task)

		suite := NewTUITestSuite(t, model)
		suite.Start()

		Expect.AssertViewContains(t, suite, "Test Task", "View should contain task description")
		Expect.AssertViewContains(t, suite, models.StatusTodo, "View should contain status")
		Expect.AssertModelState(t, suite, func(m tea.Model) bool {
			if taskModel, ok := m.(taskEditModel); ok {
				return taskModel.mode == fieldNavigation
			}
			return false
		}, "Model should be in field navigation mode")
	})
}
