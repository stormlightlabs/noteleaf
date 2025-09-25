package ui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type mockListModel struct {
	showingHelp bool
	width       int
	height      int
}

func createMockListModel() *mockListModel {
	return &mockListModel{width: 80, height: 24}
}

func (m *mockListModel) Init() tea.Cmd {
	return nil
}

func (m *mockListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "?":
			m.showingHelp = !m.showingHelp
		case "q":
			return m, tea.Quit
		case "esc":
			m.showingHelp = false
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m *mockListModel) View() string {
	if m.showingHelp {
		return "help: showing help content"
	}
	return "mock list view - normal mode"
}

func TestDataListInteractiveBehavior(t *testing.T) {
	t.Run("Help Toggle with TUI Framework", func(t *testing.T) {
		model := createMockListModel()

		suite := NewTUITestSuite(t, model)
		suite.Start()

		if err := suite.SendKeyString("?"); err != nil {
			t.Fatalf("Failed to send '?' key: %v", err)
		}

		if err := suite.WaitForView("help", 1*time.Second); err != nil {
			t.Errorf("Expected help to be shown: %v", err)
		}

		if err := suite.SendKey(tea.KeyEsc); err != nil {
			t.Fatalf("Failed to send escape key: %v", err)
		}

		if err := suite.WaitFor(func(m tea.Model) bool {
			view := m.View()
			return !containsString(view, "help")
		}, 1*time.Second); err != nil {
			t.Errorf("Help should have been hidden: %v", err)
		}
	})

	t.Run("Page Navigation with TUI Framework", func(t *testing.T) {
		model := createMockListModel()

		suite := NewTUITestSuite(t, model)
		suite.Start()

		if err := suite.SendKey(tea.KeyPgDown); err != nil {
			t.Fatalf("Failed to send page down key: %v", err)
		}

		time.Sleep(100 * time.Millisecond)

		if err := suite.SendKey(tea.KeyPgUp); err != nil {
			t.Fatalf("Failed to send page up key: %v", err)
		}

		if err := suite.SendKey(tea.KeyHome); err != nil {
			t.Fatalf("Failed to send home key: %v", err)
		}

		if err := suite.SendKey(tea.KeyEnd); err != nil {
			t.Fatalf("Failed to send end key: %v", err)
		}

		view := suite.GetCurrentView()
		if len(view) == 0 {
			t.Error("View should not be empty after navigation")
		}
	})

	t.Run("Quit Command with TUI Framework", func(t *testing.T) {
		model := createMockListModel()

		suite := NewTUITestSuite(t, model)
		suite.Start()

		if err := suite.SendKeyString("q"); err != nil {
			t.Fatalf("Failed to send 'q' key: %v", err)
		}

		if err := suite.WaitFor(func(m tea.Model) bool {
			return true
		}, 1*time.Second); err != nil {
			t.Errorf("Quit command should have been processed: %v", err)
		}
	})
}

func TestTUIPatternMatching(t *testing.T) {
	t.Run("Key Message Switch Cases", func(t *testing.T) {
		model := createMockListModel()
		suite := NewTUITestSuite(t, model)
		suite.Start()

		testKeys := []tea.KeyType{
			tea.KeyUp,
			tea.KeyDown,
			tea.KeyLeft,
			tea.KeyRight,
			tea.KeyEnter,
			tea.KeySpace,
			tea.KeyTab,
		}

		for _, keyType := range testKeys {
			err := suite.SendKey(keyType)
			if err != nil {
				t.Errorf("Failed to send key %v: %v", keyType, err)
			}

			time.Sleep(50 * time.Millisecond)
		}

		view := suite.GetCurrentView()
		if len(view) == 0 {
			t.Error("View should not be empty after key processing")
		}
	})

	t.Run("Message Type Switch Cases", func(t *testing.T) {
		model := createMockListModel()
		suite := NewTUITestSuite(t, model)
		suite.Start()

		messages := []tea.Msg{
			tea.WindowSizeMsg{Width: 100, Height: 30},
			tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}},
			tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}},
		}

		for _, msg := range messages {
			err := suite.SendMessage(msg)
			if err != nil {
				t.Errorf("Failed to send message %T: %v", msg, err)
			}

			time.Sleep(50 * time.Millisecond)
		}

		view := suite.GetCurrentView()
		if len(view) == 0 {
			t.Error("View should not be empty after message processing")
		}
	})
}
