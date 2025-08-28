package ui

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stormlightlabs/noteleaf/internal/repo"
)

// Mock project repository for testing
type mockProjectRepository struct {
	projects []repo.ProjectSummary
	err      error
}

func (m *mockProjectRepository) GetProjects(ctx context.Context) ([]repo.ProjectSummary, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.projects, nil
}

func TestProjectList(t *testing.T) {
	t.Run("NewProjectList", func(t *testing.T) {
		t.Run("creates project list successfully", func(t *testing.T) {
			mockRepo := &mockProjectRepository{}
			opts := ProjectListOptions{}

			projectList := NewProjectList(mockRepo, opts)

			if projectList == nil {
				t.Fatal("ProjectList should not be nil")
			}
			if projectList.repo != mockRepo {
				t.Error("ProjectList repo should be set correctly")
			}
		})

		t.Run("sets default options", func(t *testing.T) {
			mockRepo := &mockProjectRepository{}
			opts := ProjectListOptions{}

			projectList := NewProjectList(mockRepo, opts)

			if projectList.opts.Output == nil {
				t.Error("Default output should be set")
			}
			if projectList.opts.Input == nil {
				t.Error("Default input should be set")
			}
		})

		t.Run("preserves custom options", func(t *testing.T) {
			mockRepo := &mockProjectRepository{}
			output := &bytes.Buffer{}
			input := strings.NewReader("")
			opts := ProjectListOptions{
				Output: output,
				Input:  input,
				Static: true,
			}

			projectList := NewProjectList(mockRepo, opts)

			if projectList.opts.Output != output {
				t.Error("Custom output should be preserved")
			}
			if projectList.opts.Input != input {
				t.Error("Custom input should be preserved")
			}
			if !projectList.opts.Static {
				t.Error("Static option should be preserved")
			}
		})
	})

	t.Run("StaticList", func(t *testing.T) {
		t.Run("displays projects correctly", func(t *testing.T) {
			projects := []repo.ProjectSummary{
				{Name: "web-app", TaskCount: 5},
				{Name: "mobile-app", TaskCount: 3},
				{Name: "documentation", TaskCount: 1},
			}
			mockRepo := &mockProjectRepository{projects: projects}
			output := &bytes.Buffer{}
			opts := ProjectListOptions{Output: output, Static: true}

			projectList := NewProjectList(mockRepo, opts)
			err := projectList.Browse(context.Background())

			if err != nil {
				t.Errorf("Browse should not return error: %v", err)
			}

			result := output.String()
			if !strings.Contains(result, "Projects") {
				t.Error("Output should contain title")
			}
			if !strings.Contains(result, "web-app") {
				t.Error("Output should contain web-app project")
			}
			if !strings.Contains(result, "mobile-app") {
				t.Error("Output should contain mobile-app project")
			}
			if !strings.Contains(result, "5 tasks") {
				t.Error("Output should show correct task count for web-app")
			}
			if !strings.Contains(result, "1 task") {
				t.Error("Output should show singular task for documentation")
			}
		})

		t.Run("handles empty project list", func(t *testing.T) {
			mockRepo := &mockProjectRepository{projects: []repo.ProjectSummary{}}
			output := &bytes.Buffer{}
			opts := ProjectListOptions{Output: output, Static: true}

			projectList := NewProjectList(mockRepo, opts)
			err := projectList.Browse(context.Background())

			if err != nil {
				t.Errorf("Browse should not return error: %v", err)
			}

			result := output.String()
			if !strings.Contains(result, "No projects found") {
				t.Error("Output should indicate no projects found")
			}
		})

		t.Run("handles repository errors", func(t *testing.T) {
			mockRepo := &mockProjectRepository{err: fmt.Errorf("database error")}
			output := &bytes.Buffer{}
			opts := ProjectListOptions{Output: output, Static: true}

			projectList := NewProjectList(mockRepo, opts)
			err := projectList.Browse(context.Background())

			if err == nil {
				t.Error("Browse should return error when repository fails")
			}

			result := output.String()
			if !strings.Contains(result, "Error:") {
				t.Error("Output should contain error message")
			}
		})

		t.Run("truncates long project names", func(t *testing.T) {
			projects := []repo.ProjectSummary{
				{Name: "this-is-a-very-long-project-name-that-should-be-truncated", TaskCount: 2},
			}
			mockRepo := &mockProjectRepository{projects: projects}
			output := &bytes.Buffer{}
			opts := ProjectListOptions{Output: output, Static: true}

			projectList := NewProjectList(mockRepo, opts)
			err := projectList.Browse(context.Background())

			if err != nil {
				t.Errorf("Browse should not return error: %v", err)
			}

			result := output.String()
			if !strings.Contains(result, "...") {
				t.Error("Output should truncate long project names")
			}
		})
	})

	t.Run("ProjectListModel", func(t *testing.T) {
		t.Run("initializes correctly", func(t *testing.T) {
			model := projectListModel{
				selected:    0,
				showingHelp: false,
			}

			if model.selected != 0 {
				t.Error("Initial selection should be 0")
			}
			if model.showingHelp {
				t.Error("Should not be showing help initially")
			}
		})

		t.Run("handles key navigation", func(t *testing.T) {
			projects := []repo.ProjectSummary{
				{Name: "project1", TaskCount: 1},
				{Name: "project2", TaskCount: 2},
				{Name: "project3", TaskCount: 3},
			}

			model := projectListModel{
				projects: projects,
				selected: 1,
				keys:     projectKeys,
			}

			// Test down key
			downMsg := tea.KeyMsg{Type: tea.KeyDown}
			updatedModel, _ := model.Update(downMsg)
			if updatedModel.(projectListModel).selected != 2 {
				t.Error("Down key should move selection down")
			}

			// Test up key
			upMsg := tea.KeyMsg{Type: tea.KeyUp}
			updatedModel, _ = updatedModel.Update(upMsg)
			if updatedModel.(projectListModel).selected != 1 {
				t.Error("Up key should move selection up")
			}

			// Test boundary conditions
			model.selected = 0
			updatedModel, _ = model.Update(upMsg)
			if updatedModel.(projectListModel).selected != 0 {
				t.Error("Up key should not move selection below 0")
			}

			model.selected = len(projects) - 1
			updatedModel, _ = model.Update(downMsg)
			if updatedModel.(projectListModel).selected != len(projects)-1 {
				t.Error("Down key should not move selection beyond list length")
			}
		})

		t.Run("handles number key selection", func(t *testing.T) {
			projects := []repo.ProjectSummary{
				{Name: "project1", TaskCount: 1},
				{Name: "project2", TaskCount: 2},
				{Name: "project3", TaskCount: 3},
			}

			model := projectListModel{
				projects: projects,
				selected: 0,
				keys:     projectKeys,
			}

			// Test number key 3 (index 2)
			keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}}
			updatedModel, _ := model.Update(keyMsg)
			if updatedModel.(projectListModel).selected != 2 {
				t.Error("Number key 3 should select index 2")
			}
		})

		t.Run("handles help toggle", func(t *testing.T) {
			model := projectListModel{
				keys: projectKeys,
			}

			// Toggle help on
			helpMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
			updatedModel, _ := model.Update(helpMsg)
			if !updatedModel.(projectListModel).showingHelp {
				t.Error("Help key should show help")
			}

			// Toggle help off
			updatedModel, _ = updatedModel.Update(helpMsg)
			if updatedModel.(projectListModel).showingHelp {
				t.Error("Help key should hide help when already showing")
			}
		})

		t.Run("handles projects loaded message", func(t *testing.T) {
			projects := []repo.ProjectSummary{
				{Name: "new-project", TaskCount: 5},
			}

			model := projectListModel{
				selected: 5, // Invalid selection
			}

			msg := projectsLoadedMsg(projects)
			updatedModel, _ := model.Update(msg)
			resultModel := updatedModel.(projectListModel)

			if len(resultModel.projects) != 1 {
				t.Error("Projects should be loaded correctly")
			}
			if resultModel.selected != 0 {
				t.Error("Selection should be reset to valid range")
			}
		})

		t.Run("handles error message", func(t *testing.T) {
			model := projectListModel{}

			errorMsg := errorProjectMsg(fmt.Errorf("test error"))
			updatedModel, _ := model.Update(errorMsg)
			resultModel := updatedModel.(projectListModel)

			if resultModel.err == nil {
				t.Error("Error should be set")
			}
			if resultModel.err.Error() != "test error" {
				t.Errorf("Expected 'test error', got '%s'", resultModel.err.Error())
			}
		})

		t.Run("view renders correctly", func(t *testing.T) {
			projects := []repo.ProjectSummary{
				{Name: "web-app", TaskCount: 5},
				{Name: "mobile-app", TaskCount: 1},
			}

			model := projectListModel{
				projects: projects,
				selected: 0,
				keys:     projectKeys,
			}

			view := model.View()
			if !strings.Contains(view, "Projects") {
				t.Error("View should contain title")
			}
			if !strings.Contains(view, "web-app") {
				t.Error("View should contain project names")
			}
			if !strings.Contains(view, "5 tasks") {
				t.Error("View should show task counts")
			}
			if !strings.Contains(view, "1 task") {
				t.Error("View should show singular task count")
			}
		})

		t.Run("view handles empty state", func(t *testing.T) {
			model := projectListModel{
				projects: []repo.ProjectSummary{},
			}

			view := model.View()
			if !strings.Contains(view, "No projects found") {
				t.Error("View should show empty state message")
			}
		})

		t.Run("view handles error state", func(t *testing.T) {
			model := projectListModel{
				err: fmt.Errorf("test error"),
			}

			view := model.View()
			if !strings.Contains(view, "Error:") {
				t.Error("View should show error message")
			}
		})
	})

	t.Run("PluralizeCount", func(t *testing.T) {
		t.Run("returns empty string for singular", func(t *testing.T) {
			result := pluralizeCount(1)
			if result != "" {
				t.Errorf("Expected empty string for 1, got '%s'", result)
			}
		})

		t.Run("returns 's' for plural", func(t *testing.T) {
			testCases := []int{0, 2, 10, 100}
			for _, count := range testCases {
				result := pluralizeCount(count)
				if result != "s" {
					t.Errorf("Expected 's' for %d, got '%s'", count, result)
				}
			}
		})
	})
}