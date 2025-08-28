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

// Mock tag repository for testing
type mockTagRepository struct {
	tags []repo.TagSummary
	err  error
}

func (m *mockTagRepository) GetTags(ctx context.Context) ([]repo.TagSummary, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.tags, nil
}

func TestTagList(t *testing.T) {
	t.Run("NewTagList", func(t *testing.T) {
		t.Run("creates tag list successfully", func(t *testing.T) {
			mockRepo := &mockTagRepository{}
			opts := TagListOptions{}

			tagList := NewTagList(mockRepo, opts)

			if tagList == nil {
				t.Fatal("TagList should not be nil")
			}
			if tagList.repo != mockRepo {
				t.Error("TagList repo should be set correctly")
			}
		})

		t.Run("sets default options", func(t *testing.T) {
			mockRepo := &mockTagRepository{}
			opts := TagListOptions{}

			tagList := NewTagList(mockRepo, opts)

			if tagList.opts.Output == nil {
				t.Error("Default output should be set")
			}
			if tagList.opts.Input == nil {
				t.Error("Default input should be set")
			}
		})

		t.Run("preserves custom options", func(t *testing.T) {
			mockRepo := &mockTagRepository{}
			output := &bytes.Buffer{}
			input := strings.NewReader("")
			opts := TagListOptions{
				Output: output,
				Input:  input,
				Static: true,
			}

			tagList := NewTagList(mockRepo, opts)

			if tagList.opts.Output != output {
				t.Error("Custom output should be preserved")
			}
			if tagList.opts.Input != input {
				t.Error("Custom input should be preserved")
			}
			if !tagList.opts.Static {
				t.Error("Static option should be preserved")
			}
		})
	})

	t.Run("StaticList", func(t *testing.T) {
		t.Run("displays tags correctly", func(t *testing.T) {
			tags := []repo.TagSummary{
				{Name: "frontend", TaskCount: 5},
				{Name: "backend", TaskCount: 3},
				{Name: "urgent", TaskCount: 1},
			}
			mockRepo := &mockTagRepository{tags: tags}
			output := &bytes.Buffer{}
			opts := TagListOptions{Output: output, Static: true}

			tagList := NewTagList(mockRepo, opts)
			err := tagList.Browse(context.Background())

			if err != nil {
				t.Errorf("Browse should not return error: %v", err)
			}

			result := output.String()
			if !strings.Contains(result, "Tags") {
				t.Error("Output should contain title")
			}
			if !strings.Contains(result, "frontend") {
				t.Error("Output should contain frontend tag")
			}
			if !strings.Contains(result, "backend") {
				t.Error("Output should contain backend tag")
			}
			if !strings.Contains(result, "5 tasks") {
				t.Error("Output should show correct task count for frontend")
			}
			if !strings.Contains(result, "1 task") {
				t.Error("Output should show singular task for urgent")
			}
		})

		t.Run("handles empty tag list", func(t *testing.T) {
			mockRepo := &mockTagRepository{tags: []repo.TagSummary{}}
			output := &bytes.Buffer{}
			opts := TagListOptions{Output: output, Static: true}

			tagList := NewTagList(mockRepo, opts)
			err := tagList.Browse(context.Background())

			if err != nil {
				t.Errorf("Browse should not return error: %v", err)
			}

			result := output.String()
			if !strings.Contains(result, "No tags found") {
				t.Error("Output should indicate no tags found")
			}
		})

		t.Run("handles repository errors", func(t *testing.T) {
			mockRepo := &mockTagRepository{err: fmt.Errorf("database error")}
			output := &bytes.Buffer{}
			opts := TagListOptions{Output: output, Static: true}

			tagList := NewTagList(mockRepo, opts)
			err := tagList.Browse(context.Background())

			if err == nil {
				t.Error("Browse should return error when repository fails")
			}

			result := output.String()
			if !strings.Contains(result, "Error:") {
				t.Error("Output should contain error message")
			}
		})

		t.Run("truncates long tag names", func(t *testing.T) {
			tags := []repo.TagSummary{
				{Name: "this-is-a-very-long-tag-name-that-should-be-truncated", TaskCount: 2},
			}
			mockRepo := &mockTagRepository{tags: tags}
			output := &bytes.Buffer{}
			opts := TagListOptions{Output: output, Static: true}

			tagList := NewTagList(mockRepo, opts)
			err := tagList.Browse(context.Background())

			if err != nil {
				t.Errorf("Browse should not return error: %v", err)
			}

			result := output.String()
			if !strings.Contains(result, "...") {
				t.Error("Output should truncate long tag names")
			}
		})
	})

	t.Run("TagListModel", func(t *testing.T) {
		t.Run("initializes correctly", func(t *testing.T) {
			model := tagListModel{
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
			tags := []repo.TagSummary{
				{Name: "tag1", TaskCount: 1},
				{Name: "tag2", TaskCount: 2},
				{Name: "tag3", TaskCount: 3},
			}

			model := tagListModel{
				tags:     tags,
				selected: 1,
				keys:     tagKeys,
			}

			// Test down key
			downMsg := tea.KeyMsg{Type: tea.KeyDown}
			updatedModel, _ := model.Update(downMsg)
			if updatedModel.(tagListModel).selected != 2 {
				t.Error("Down key should move selection down")
			}

			// Test up key
			upMsg := tea.KeyMsg{Type: tea.KeyUp}
			updatedModel, _ = updatedModel.Update(upMsg)
			if updatedModel.(tagListModel).selected != 1 {
				t.Error("Up key should move selection up")
			}

			// Test boundary conditions
			model.selected = 0
			updatedModel, _ = model.Update(upMsg)
			if updatedModel.(tagListModel).selected != 0 {
				t.Error("Up key should not move selection below 0")
			}

			model.selected = len(tags) - 1
			updatedModel, _ = model.Update(downMsg)
			if updatedModel.(tagListModel).selected != len(tags)-1 {
				t.Error("Down key should not move selection beyond list length")
			}
		})

		t.Run("handles number key selection", func(t *testing.T) {
			tags := []repo.TagSummary{
				{Name: "tag1", TaskCount: 1},
				{Name: "tag2", TaskCount: 2},
				{Name: "tag3", TaskCount: 3},
			}

			model := tagListModel{
				tags:     tags,
				selected: 0,
				keys:     tagKeys,
			}

			// Test number key 3 (index 2)
			keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}}
			updatedModel, _ := model.Update(keyMsg)
			if updatedModel.(tagListModel).selected != 2 {
				t.Error("Number key 3 should select index 2")
			}
		})

		t.Run("handles help toggle", func(t *testing.T) {
			model := tagListModel{
				keys: tagKeys,
			}

			helpMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
			updatedModel, _ := model.Update(helpMsg)
			if !updatedModel.(tagListModel).showingHelp {
				t.Error("Help key should show help")
			}

			updatedModel, _ = updatedModel.Update(helpMsg)
			if updatedModel.(tagListModel).showingHelp {
				t.Error("Help key should hide help when already showing")
			}
		})

		t.Run("handles tags loaded message", func(t *testing.T) {
			tags := []repo.TagSummary{
				{Name: "new-tag", TaskCount: 5},
			}

			model := tagListModel{
				selected: 5, // Invalid selection
			}

			msg := tagsLoadedMsg(tags)
			updatedModel, _ := model.Update(msg)
			resultModel := updatedModel.(tagListModel)

			if len(resultModel.tags) != 1 {
				t.Error("Tags should be loaded correctly")
			}
			if resultModel.selected != 0 {
				t.Error("Selection should be reset to valid range")
			}
		})

		t.Run("handles error message", func(t *testing.T) {
			model := tagListModel{}

			errorMsg := errorTagMsg(fmt.Errorf("test error"))
			updatedModel, _ := model.Update(errorMsg)
			resultModel := updatedModel.(tagListModel)

			if resultModel.err == nil {
				t.Error("Error should be set")
			}
			if resultModel.err.Error() != "test error" {
				t.Errorf("Expected 'test error', got '%s'", resultModel.err.Error())
			}
		})

		t.Run("view renders correctly", func(t *testing.T) {
			tags := []repo.TagSummary{
				{Name: "frontend", TaskCount: 5},
				{Name: "urgent", TaskCount: 1},
			}

			model := tagListModel{
				tags:     tags,
				selected: 0,
				keys:     tagKeys,
			}

			view := model.View()
			if !strings.Contains(view, "Tags") {
				t.Error("View should contain title")
			}
			if !strings.Contains(view, "frontend") {
				t.Error("View should contain tag names")
			}
			if !strings.Contains(view, "5 tasks") {
				t.Error("View should show task counts")
			}
			if !strings.Contains(view, "1 task") {
				t.Error("View should show singular task count")
			}
		})

		t.Run("view handles empty state", func(t *testing.T) {
			model := tagListModel{
				tags: []repo.TagSummary{},
			}

			view := model.View()
			if !strings.Contains(view, "No tags found") {
				t.Error("View should show empty state message")
			}
		})

		t.Run("view handles error state", func(t *testing.T) {
			model := tagListModel{
				err: fmt.Errorf("test error"),
			}

			view := model.View()
			if !strings.Contains(view, "Error:") {
				t.Error("View should show error message")
			}
		})
	})
}
