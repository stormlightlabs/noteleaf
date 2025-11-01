package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func createTestAuthFormModel(handle string) authFormModel {
	handleInput := textinput.New()
	handleInput.Placeholder = "username.bsky.social"
	handleInput.Width = 40

	passwordInput := textinput.New()
	passwordInput.Placeholder = "App password"
	passwordInput.Width = 40
	passwordInput.EchoMode = textinput.EchoPassword
	passwordInput.EchoCharacter = '•'

	handleLocked := false
	focusIndex := 0

	if handle != "" {
		handleInput.SetValue(handle)
		handleLocked = true
		focusIndex = 1
	}

	return authFormModel{
		handleInput:   handleInput,
		passwordInput: passwordInput,
		focusIndex:    focusIndex,
		keys:          authFormKeys,
		handleLocked:  handleLocked,
	}
}

func TestAuthFormModel(t *testing.T) {
	t.Run("Init", func(t *testing.T) {
		t.Run("focuses handle input when no initial handle", func(t *testing.T) {
			model := createTestAuthFormModel("")

			cmd := model.Init()
			if cmd == nil {
				t.Error("Expected Init to return a focus command")
			}
		})

		t.Run("focuses password input when handle is locked", func(t *testing.T) {
			model := createTestAuthFormModel("test.bsky.social")

			cmd := model.Init()
			if cmd == nil {
				t.Error("Expected Init to return a focus command")
			}

			if !model.handleLocked {
				t.Error("Expected handleLocked to be true")
			}
			if model.focusIndex != 1 {
				t.Errorf("Expected focusIndex to be 1, got %d", model.focusIndex)
			}
		})
	})

	t.Run("Navigation", func(t *testing.T) {
		t.Run("tab moves to next field", func(t *testing.T) {
			model := createTestAuthFormModel("")
			if cmd := model.Init(); cmd != nil {
				model.handleInput.Focus()
			}

			suite := NewTUITestSuite(t, model)
			suite.Start()

			if err := suite.SendKey(tea.KeyTab); err != nil {
				t.Fatalf("Failed to send tab key: %v", err)
			}

			if err := suite.WaitFor(func(m tea.Model) bool {
				if authModel, ok := m.(authFormModel); ok {
					return authModel.focusIndex == 1
				}
				return false
			}, 1*time.Second); err != nil {
				t.Errorf("Expected focusIndex to change to 1: %v", err)
			}
		})

		t.Run("shift+tab moves to previous field", func(t *testing.T) {
			model := createTestAuthFormModel("")
			if cmd := model.Init(); cmd != nil {
				model.handleInput.Focus()
			}
			model.focusIndex = 1

			suite := NewTUITestSuite(t, model)
			suite.Start()

			if err := suite.SendKey(tea.KeyShiftTab); err != nil {
				t.Fatalf("Failed to send shift+tab key: %v", err)
			}

			if err := suite.WaitFor(func(m tea.Model) bool {
				if authModel, ok := m.(authFormModel); ok {
					return authModel.focusIndex == 0
				}
				return false
			}, 1*time.Second); err != nil {
				t.Errorf("Expected focusIndex to change to 0: %v", err)
			}
		})

		t.Run("locked handle prevents navigation to handle field", func(t *testing.T) {
			model := createTestAuthFormModel("test.bsky.social")

			suite := NewTUITestSuite(t, model)
			suite.Start()

			if err := suite.SendKey(tea.KeyShiftTab); err != nil {
				t.Fatalf("Failed to send shift+tab key: %v", err)
			}

			if err := suite.WaitFor(func(m tea.Model) bool {
				if authModel, ok := m.(authFormModel); ok {
					return authModel.focusIndex == 1
				}
				return false
			}, 500*time.Millisecond); err != nil {
				t.Errorf("Expected focusIndex to stay at 1 when handle is locked: %v", err)
			}
		})
	})

	t.Run("Submission", func(t *testing.T) {
		t.Run("enter submits when both fields are filled", func(t *testing.T) {
			model := createTestAuthFormModel("")
			model.handleInput.SetValue("test.bsky.social")
			model.passwordInput.SetValue("test-password")

			suite := NewTUITestSuite(t, model)
			suite.Start()

			if err := suite.SendKey(tea.KeyEnter); err != nil {
				t.Fatalf("Failed to send enter key: %v", err)
			}

			if err := suite.WaitFor(func(m tea.Model) bool {
				if authModel, ok := m.(authFormModel); ok {
					return authModel.submitted
				}
				return false
			}, 1*time.Second); err != nil {
				t.Errorf("Expected model to be submitted: %v", err)
			}
		})

		t.Run("enter does not submit when handle is empty", func(t *testing.T) {
			model := createTestAuthFormModel("")
			model.passwordInput.SetValue("test-password")

			suite := NewTUITestSuite(t, model)
			suite.Start()

			if err := suite.SendKey(tea.KeyEnter); err != nil {
				t.Fatalf("Failed to send enter key: %v", err)
			}

			time.Sleep(100 * time.Millisecond)

			currentModel := suite.GetCurrentModel()
			if authModel, ok := currentModel.(authFormModel); ok {
				if authModel.submitted {
					t.Error("Expected model to not be submitted when handle is empty")
				}
			}
		})

		t.Run("enter does not submit when password is empty", func(t *testing.T) {
			model := createTestAuthFormModel("")
			model.handleInput.SetValue("test.bsky.social")

			suite := NewTUITestSuite(t, model)
			suite.Start()

			if err := suite.SendKey(tea.KeyEnter); err != nil {
				t.Fatalf("Failed to send enter key: %v", err)
			}

			time.Sleep(100 * time.Millisecond)

			currentModel := suite.GetCurrentModel()
			if authModel, ok := currentModel.(authFormModel); ok {
				if authModel.submitted {
					t.Error("Expected model to not be submitted when password is empty")
				}
			}
		})

		t.Run("ctrl+s submits when both fields are filled", func(t *testing.T) {
			model := createTestAuthFormModel("")
			model.handleInput.SetValue("test.bsky.social")
			model.passwordInput.SetValue("test-password")

			suite := NewTUITestSuite(t, model)
			suite.Start()

			if err := suite.SendKeyString("ctrl+s"); err != nil {
				t.Fatalf("Failed to send ctrl+s: %v", err)
			}

			if err := suite.WaitFor(func(m tea.Model) bool {
				if authModel, ok := m.(authFormModel); ok {
					return authModel.submitted
				}
				return false
			}, 1*time.Second); err != nil {
				t.Errorf("Expected model to be submitted: %v", err)
			}
		})
	})

	t.Run("Cancellation", func(t *testing.T) {
		t.Run("esc cancels the form", func(t *testing.T) {
			model := createTestAuthFormModel("")

			suite := NewTUITestSuite(t, model)
			suite.Start()

			if err := suite.SendKey(tea.KeyEsc); err != nil {
				t.Fatalf("Failed to send esc key: %v", err)
			}

			if err := suite.WaitFor(func(m tea.Model) bool {
				if authModel, ok := m.(authFormModel); ok {
					return authModel.canceled
				}
				return false
			}, 1*time.Second); err != nil {
				t.Errorf("Expected model to be canceled: %v", err)
			}
		})

		t.Run("ctrl+c cancels the form", func(t *testing.T) {
			model := createTestAuthFormModel("")

			suite := NewTUITestSuite(t, model)
			suite.Start()

			if err := suite.SendKeyString("ctrl+c"); err != nil {
				t.Fatalf("Failed to send ctrl+c: %v", err)
			}

			if err := suite.WaitFor(func(m tea.Model) bool {
				if authModel, ok := m.(authFormModel); ok {
					return authModel.canceled
				}
				return false
			}, 1*time.Second); err != nil {
				t.Errorf("Expected model to be canceled: %v", err)
			}
		})
	})

	t.Run("View", func(t *testing.T) {
		t.Run("displays handle and password fields", func(t *testing.T) {
			model := createTestAuthFormModel("")

			view := model.View()

			if !strings.Contains(view, "AT Protocol Authentication") {
				t.Error("Expected view to contain title")
			}
			if !strings.Contains(view, "BlueSky Handle:") {
				t.Error("Expected view to contain handle label")
			}
			if !strings.Contains(view, "App Password:") {
				t.Error("Expected view to contain password label")
			}
		})

		t.Run("displays locked status for handle", func(t *testing.T) {
			model := createTestAuthFormModel("test.bsky.social")

			view := model.View()

			if !strings.Contains(view, "test.bsky.social") {
				t.Error("Expected view to contain handle value")
			}
			if !strings.Contains(view, "(locked)") {
				t.Error("Expected view to indicate handle is locked")
			}
		})

		t.Run("displays validation messages when fields are empty", func(t *testing.T) {
			model := createTestAuthFormModel("")

			view := model.View()

			if !strings.Contains(view, "Handle is required") {
				t.Error("Expected view to show handle validation message")
			}
			if !strings.Contains(view, "Password is required") {
				t.Error("Expected view to show password validation message")
			}
		})

		t.Run("displays help text", func(t *testing.T) {
			model := createTestAuthFormModel("")

			view := model.View()

			if !strings.Contains(view, "tab/shift+tab: navigate") {
				t.Error("Expected view to contain navigation help")
			}
			if !strings.Contains(view, "enter/ctrl+s: submit") {
				t.Error("Expected view to contain submit help")
			}
			if !strings.Contains(view, "esc/ctrl+c: cancel") {
				t.Error("Expected view to contain cancel help")
			}
		})
	})

	t.Run("Input handling", func(t *testing.T) {
		t.Run("accepts text input in handle field", func(t *testing.T) {
			model := createTestAuthFormModel("")
			if cmd := model.Init(); cmd != nil {
				model.handleInput.Focus()
			}

			suite := NewTUITestSuite(t, model)
			suite.Start()

			if err := suite.SendKeyString("t"); err != nil {
				t.Fatalf("Failed to send 't' key: %v", err)
			}

			if err := suite.WaitFor(func(m tea.Model) bool {
				if authModel, ok := m.(authFormModel); ok {
					return authModel.handleInput.Value() == "t"
				}
				return false
			}, 1*time.Second); err != nil {
				t.Errorf("Expected handle input to contain 't': %v", err)
			}
		})

		t.Run("does not accept text input in locked handle field", func(t *testing.T) {
			model := createTestAuthFormModel("test.bsky.social")
			originalValue := model.handleInput.Value()

			suite := NewTUITestSuite(t, model)
			suite.Start()

			if err := suite.SendKeyString("x"); err != nil {
				t.Fatalf("Failed to send 'x' key: %v", err)
			}

			time.Sleep(100 * time.Millisecond)

			currentModel := suite.GetCurrentModel()
			if authModel, ok := currentModel.(authFormModel); ok {
				if authModel.handleInput.Value() != originalValue {
					t.Errorf("Expected handle input to remain unchanged, got '%s'", authModel.handleInput.Value())
				}
			}
		})
	})
}

func TestNewAuthForm(t *testing.T) {
	t.Run("creates form with default options", func(t *testing.T) {
		form := NewAuthForm("", AuthFormOptions{})

		if form == nil {
			t.Fatal("Expected form to be created")
		}
		if form.opts.Width != 80 {
			t.Errorf("Expected default width 80, got %d", form.opts.Width)
		}
		if form.opts.Height != 24 {
			t.Errorf("Expected default height 24, got %d", form.opts.Height)
		}
		if form.opts.Output == nil {
			t.Error("Expected output to be set to default")
		}
		if form.opts.Input == nil {
			t.Error("Expected input to be set to default")
		}
	})

	t.Run("creates form with initial handle", func(t *testing.T) {
		handle := "test.bsky.social"
		form := NewAuthForm(handle, AuthFormOptions{})

		if form.initialHandle != handle {
			t.Errorf("Expected initialHandle '%s', got '%s'", handle, form.initialHandle)
		}
	})

	t.Run("creates form with custom options", func(t *testing.T) {
		opts := AuthFormOptions{
			Width:  100,
			Height: 30,
		}
		form := NewAuthForm("", opts)

		if form.opts.Width != 100 {
			t.Errorf("Expected width 100, got %d", form.opts.Width)
		}
		if form.opts.Height != 30 {
			t.Errorf("Expected height 30, got %d", form.opts.Height)
		}
	})
}
