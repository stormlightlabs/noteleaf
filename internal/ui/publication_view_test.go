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

func createMockNote() *models.Note {
	now := time.Now()
	publishedAt := now.Add(-24 * time.Hour)
	rkey := "test-rkey-123"
	cid := "test-cid-456"

	return &models.Note{
		ID:          1,
		Title:       "Test Publication",
		Content:     "# Test Publication\n\nThis is the content of the test publication.",
		Tags:        []string{"test", "publication"},
		Archived:    false,
		Created:     now.Add(-48 * time.Hour),
		Modified:    now.Add(-1 * time.Hour),
		LeafletRKey: &rkey,
		LeafletCID:  &cid,
		PublishedAt: &publishedAt,
		IsDraft:     false,
	}
}

func createDraftNote() *models.Note {
	note := createMockNote()
	note.IsDraft = true
	note.PublishedAt = nil
	note.LeafletRKey = nil
	note.LeafletCID = nil
	return note
}

func createMinimalNote() *models.Note {
	now := time.Now()
	return &models.Note{
		ID:       2,
		Title:    "Minimal Note",
		Content:  "Simple content without markdown heading.",
		Created:  now,
		Modified: now,
		IsDraft:  true,
	}
}

func TestPublicationView(t *testing.T) {
	t.Run("View Options", func(t *testing.T) {
		note := createMockNote()

		t.Run("default options", func(t *testing.T) {
			opts := PublicationViewOptions{}
			pv := NewPublicationView(note, opts)

			if pv.opts.Output == nil {
				t.Error("Output should default to os.Stdout")
			}
			if pv.opts.Input == nil {
				t.Error("Input should default to os.Stdin")
			}
			if pv.opts.Width != 80 {
				t.Errorf("Width should default to 80, got %d", pv.opts.Width)
			}
			if pv.opts.Height != 24 {
				t.Errorf("Height should default to 24, got %d", pv.opts.Height)
			}
		})

		t.Run("custom options", func(t *testing.T) {
			var buf bytes.Buffer
			opts := PublicationViewOptions{
				Output: &buf,
				Static: true,
				Width:  100,
				Height: 30,
			}
			pv := NewPublicationView(note, opts)

			if pv.opts.Output != &buf {
				t.Error("Custom output not set")
			}
			if !pv.opts.Static {
				t.Error("Static mode not set")
			}
			if pv.opts.Width != 100 {
				t.Error("Custom width not set")
			}
			if pv.opts.Height != 30 {
				t.Error("Custom height not set")
			}
		})
	})

	t.Run("New", func(t *testing.T) {
		note := createMockNote()

		t.Run("creates publication view correctly", func(t *testing.T) {
			opts := PublicationViewOptions{Width: 60, Height: 20}
			pv := NewPublicationView(note, opts)

			if pv.note != note {
				t.Error("Note not set correctly")
			}
			if pv.opts.Width != 60 {
				t.Error("Width not set correctly")
			}
			if pv.opts.Height != 20 {
				t.Error("Height not set correctly")
			}
		})
	})

	t.Run("Static Mode", func(t *testing.T) {
		t.Run("published note display", func(t *testing.T) {
			note := createMockNote()
			var buf bytes.Buffer

			pv := NewPublicationView(note, PublicationViewOptions{
				Output: &buf,
				Static: true,
			})

			err := pv.Show(context.Background())
			if err != nil {
				t.Fatalf("Show failed: %v", err)
			}

			output := buf.String()

			if !strings.Contains(output, "Test") || !strings.Contains(output, "Publication") {
				t.Error("Note title not displayed")
			}
			if !strings.Contains(output, "published") {
				t.Error("Published status not displayed")
			}
			if !strings.Contains(output, "Published:") {
				t.Error("Published date not displayed")
			}
			if !strings.Contains(output, "Modified:") {
				t.Error("Modified date not displayed")
			}
			if !strings.Contains(output, "RKey:") {
				t.Error("RKey not displayed")
			}
			if !strings.Contains(output, "tes") || !strings.Contains(output, "123") {
				t.Error("RKey value not displayed")
			}
			if !strings.Contains(output, "CID:") {
				t.Error("CID not displayed")
			}
			if !strings.Contains(output, "tes") || !strings.Contains(output, "456") {
				t.Error("CID value not displayed")
			}
			if !strings.Contains(output, "This") || !strings.Contains(output, "content") {
				t.Error("Note content not displayed")
			}
		})

		t.Run("draft note display", func(t *testing.T) {
			note := createDraftNote()
			var buf bytes.Buffer

			pv := NewPublicationView(note, PublicationViewOptions{
				Output: &buf,
				Static: true,
			})

			err := pv.Show(context.Background())
			if err != nil {
				t.Fatalf("Show failed: %v", err)
			}

			output := buf.String()

			if !strings.Contains(output, "draft") {
				t.Error("Draft status not displayed")
			}
			if strings.Contains(output, "Published:") {
				t.Error("Published date should not be displayed for draft")
			}
			if strings.Contains(output, "RKey:") {
				t.Error("RKey should not be displayed for draft")
			}
			if strings.Contains(output, "CID:") {
				t.Error("CID should not be displayed for draft")
			}
		})

		t.Run("minimal note display", func(t *testing.T) {
			note := createMinimalNote()
			var buf bytes.Buffer

			pv := NewPublicationView(note, PublicationViewOptions{
				Output: &buf,
				Static: true,
			})

			err := pv.Show(context.Background())
			if err != nil {
				t.Fatalf("Show failed: %v", err)
			}

			output := buf.String()

			if !strings.Contains(output, "Minimal") || !strings.Contains(output, "Note") {
				t.Error("Note title not displayed")
			}
			if !strings.Contains(output, "Simple") || !strings.Contains(output, "content") {
				t.Error("Note content not displayed")
			}
			if !strings.Contains(output, "Modified:") {
				t.Error("Modified date not displayed")
			}
		})
	})

	t.Run("Build Markdown", func(t *testing.T) {
		t.Run("builds markdown for published note", func(t *testing.T) {
			note := createMockNote()
			markdown := buildPublicationMarkdown(note)

			expectedStrings := []string{
				"# Test Publication",
				"**Status:** published",
				"**Published:**",
				"**Modified:**",
				"**RKey:**",
				"**CID:**",
				"---",
				"This is the content",
			}

			for _, expected := range expectedStrings {
				if !strings.Contains(markdown, expected) {
					t.Errorf("Expected markdown '%s' not found in output", expected)
				}
			}
		})

		t.Run("builds markdown for draft note", func(t *testing.T) {
			note := createDraftNote()
			markdown := buildPublicationMarkdown(note)

			if !strings.Contains(markdown, "**Status:** draft") {
				t.Error("Draft status not in markdown")
			}
			if strings.Contains(markdown, "**Published:**") {
				t.Error("Published date should not be in draft markdown")
			}
			if strings.Contains(markdown, "**RKey:**") {
				t.Error("RKey should not be in draft markdown")
			}
			if strings.Contains(markdown, "**CID:**") {
				t.Error("CID should not be in draft markdown")
			}
		})

		t.Run("handles content with markdown heading", func(t *testing.T) {
			note := createMockNote()
			markdown := buildPublicationMarkdown(note)

			headingCount := strings.Count(markdown, "# Test Publication")
			if headingCount != 1 {
				t.Errorf("Expected 1 heading, found %d (content heading should be stripped)", headingCount)
			}
		})

		t.Run("handles content without markdown heading", func(t *testing.T) {
			note := createMinimalNote()
			markdown := buildPublicationMarkdown(note)

			if !strings.Contains(markdown, "Simple content") {
				t.Error("Content without heading should be included as-is")
			}
		})
	})

	t.Run("Format Content", func(t *testing.T) {
		t.Run("formats content with glamour", func(t *testing.T) {
			note := createMockNote()
			content, err := formatPublicationContent(note)

			if err != nil {
				t.Fatalf("formatPublicationContent failed: %v", err)
			}

			if len(content) == 0 {
				t.Error("Formatted content should not be empty")
			}
		})

		t.Run("includes note title in formatted content", func(t *testing.T) {
			note := createMockNote()
			content, err := formatPublicationContent(note)

			if err != nil {
				t.Fatalf("formatPublicationContent failed: %v", err)
			}

			if !strings.Contains(content, "Test") || !strings.Contains(content, "Publication") {
				t.Error("Formatted content should include note title")
			}
		})
	})

	t.Run("Model", func(t *testing.T) {
		note := createMockNote()

		t.Run("initial model state", func(t *testing.T) {
			model := publicationViewModel{
				note: note,
				opts: PublicationViewOptions{Width: 80, Height: 24},
			}

			if model.showingHelp {
				t.Error("Initial showingHelp should be false")
			}
			if model.note != note {
				t.Error("Note not set correctly")
			}
		})

		t.Run("key handling - help toggle", func(t *testing.T) {
			vp := viewport.New(80, 20)
			model := publicationViewModel{
				note:     note,
				viewport: vp,
				keys:     publicationViewKeys,
				help:     help.New(),
				ready:    true,
			}

			newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
			if m, ok := newModel.(publicationViewModel); ok {
				if !m.showingHelp {
					t.Error("Help key should show help")
				}
			}

			model.showingHelp = true
			newModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
			if m, ok := newModel.(publicationViewModel); ok {
				if m.showingHelp {
					t.Error("Help key should exit help when already showing")
				}
			}
		})

		t.Run("key handling - quit and back", func(t *testing.T) {
			vp := viewport.New(80, 20)
			model := publicationViewModel{
				note:     note,
				viewport: vp,
				keys:     publicationViewKeys,
				help:     help.New(),
				ready:    true,
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
			longContent := strings.Repeat("Line of content\n", 50)
			vp.SetContent(longContent)

			model := publicationViewModel{
				note:     note,
				viewport: vp,
				keys:     publicationViewKeys,
				help:     help.New(),
				ready:    true,
			}

			initialOffset := model.viewport.YOffset

			newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
			if m, ok := newModel.(publicationViewModel); ok {
				if m.viewport.YOffset <= initialOffset {
					t.Error("Down key should scroll viewport down")
				}
			}

			model.viewport.ScrollDown(5)
			initialOffset = model.viewport.YOffset
			newModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
			if m, ok := newModel.(publicationViewModel); ok {
				if m.viewport.YOffset >= initialOffset {
					t.Error("Up key should scroll viewport up")
				}
			}
		})

		t.Run("page navigation", func(t *testing.T) {
			vp := viewport.New(80, 20)
			longContent := strings.Repeat("Line of content\n", 100)
			vp.SetContent(longContent)

			model := publicationViewModel{
				note:     note,
				viewport: vp,
				keys:     publicationViewKeys,
				help:     help.New(),
				ready:    true,
			}

			initialOffset := model.viewport.YOffset

			newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")})
			if m, ok := newModel.(publicationViewModel); ok {
				if m.viewport.YOffset <= initialOffset {
					t.Error("Page down key should scroll viewport down")
				}
			}
		})

		t.Run("top and bottom navigation", func(t *testing.T) {
			vp := viewport.New(80, 20)
			longContent := strings.Repeat("Line of content\n", 100)
			vp.SetContent(longContent)

			model := publicationViewModel{
				note:     note,
				viewport: vp,
				keys:     publicationViewKeys,
				help:     help.New(),
				ready:    true,
			}

			model.viewport.ScrollDown(50)

			newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
			if m, ok := newModel.(publicationViewModel); ok {
				if m.viewport.YOffset != 0 {
					t.Error("Top key should scroll to top")
				}
			}

			newModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
			if m, ok := newModel.(publicationViewModel); ok {
				if m.viewport.YOffset == 0 {
					t.Error("Bottom key should scroll to bottom")
				}
			}
		})

		t.Run("window size message handling", func(t *testing.T) {
			vp := viewport.New(80, 20)
			model := publicationViewModel{
				note:     note,
				viewport: vp,
				keys:     publicationViewKeys,
				help:     help.New(),
				opts:     PublicationViewOptions{Static: false},
			}

			newModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
			if m, ok := newModel.(publicationViewModel); ok {
				if m.viewport.Width != 98 {
					t.Errorf("Viewport width should be 98 (100-2), got %d", m.viewport.Width)
				}
				expectedHeight := 30 - 6
				if m.viewport.Height != expectedHeight {
					t.Errorf("Viewport height should be %d, got %d", expectedHeight, m.viewport.Height)
				}
				if !m.ready {
					t.Error("Model should be ready after window size message")
				}
			}
		})

		t.Run("static mode ignores window resize", func(t *testing.T) {
			vp := viewport.New(80, 20)
			model := publicationViewModel{
				note:     note,
				viewport: vp,
				keys:     publicationViewKeys,
				help:     help.New(),
				opts:     PublicationViewOptions{Static: true},
				ready:    true,
			}

			newModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
			if m, ok := newModel.(publicationViewModel); ok {
				if m.viewport.Width != 80 {
					t.Error("Static mode should not resize viewport width")
				}
				if m.viewport.Height != 20 {
					t.Error("Static mode should not resize viewport height")
				}
			}
		})
	})

	t.Run("View Model", func(t *testing.T) {
		note := createMockNote()

		t.Run("normal view with published note", func(t *testing.T) {
			vp := viewport.New(80, 20)
			content, _ := formatPublicationContent(note)
			vp.SetContent(content)

			model := publicationViewModel{
				note:     note,
				viewport: vp,
				keys:     publicationViewKeys,
				help:     help.New(),
				ready:    true,
			}

			view := model.View()

			if !strings.Contains(view, "Test Publication") {
				t.Error("Note title not displayed in view")
			}
			if !strings.Contains(view, "published") {
				t.Error("Published status not displayed in view")
			}
		})

		t.Run("normal view with draft note", func(t *testing.T) {
			draft := createDraftNote()
			vp := viewport.New(80, 20)
			content, _ := formatPublicationContent(draft)
			vp.SetContent(content)

			model := publicationViewModel{
				note:     draft,
				viewport: vp,
				keys:     publicationViewKeys,
				help:     help.New(),
				ready:    true,
			}

			view := model.View()

			if !strings.Contains(view, "draft") {
				t.Error("Draft status not displayed in view")
			}
		})

		t.Run("help view", func(t *testing.T) {
			vp := viewport.New(80, 20)
			model := publicationViewModel{
				note:        note,
				viewport:    vp,
				keys:        publicationViewKeys,
				help:        help.New(),
				showingHelp: true,
				ready:       true,
			}

			view := model.View()

			if !strings.Contains(view, "scroll") {
				t.Error("Help view should contain scroll instructions")
			}
		})

		t.Run("initializing view", func(t *testing.T) {
			vp := viewport.New(80, 20)
			model := publicationViewModel{
				note:     note,
				viewport: vp,
				keys:     publicationViewKeys,
				help:     help.New(),
				ready:    false,
			}

			view := model.View()

			if !strings.Contains(view, "Initializing") {
				t.Error("Not ready state should show initializing message")
			}
		})
	})

	t.Run("Key Bindings", func(t *testing.T) {
		t.Run("short help bindings", func(t *testing.T) {
			bindings := publicationViewKeys.ShortHelp()
			if len(bindings) != 5 {
				t.Errorf("Expected 5 short help bindings, got %d", len(bindings))
			}
		})

		t.Run("full help bindings", func(t *testing.T) {
			bindings := publicationViewKeys.FullHelp()
			if len(bindings) != 3 {
				t.Errorf("Expected 3 rows of full help bindings, got %d", len(bindings))
			}
		})
	})

	t.Run("Integration", func(t *testing.T) {
		t.Run("creates and displays publication view", func(t *testing.T) {
			note := createMockNote()
			var buf bytes.Buffer

			pv := NewPublicationView(note, PublicationViewOptions{
				Output: &buf,
				Static: true,
				Width:  80,
				Height: 24,
			})

			if pv == nil {
				t.Fatal("NewPublicationView returned nil")
			}

			err := pv.Show(context.Background())
			if err != nil {
				t.Fatalf("Show failed: %v", err)
			}

			output := buf.String()
			if len(output) == 0 {
				t.Error("No output generated")
			}

			if !strings.Contains(output, "Test") || !strings.Contains(output, "Publication") {
				t.Error("Note title not displayed")
			}
			if !strings.Contains(output, "This") || !strings.Contains(output, "content") {
				t.Error("Note content not displayed")
			}
		})

		t.Run("creates publication view for draft", func(t *testing.T) {
			draft := createDraftNote()
			var buf bytes.Buffer

			pv := NewPublicationView(draft, PublicationViewOptions{
				Output: &buf,
				Static: true,
			})

			err := pv.Show(context.Background())
			if err != nil {
				t.Fatalf("Show failed: %v", err)
			}

			output := buf.String()
			if !strings.Contains(output, "draft") {
				t.Error("Draft status not displayed")
			}
		})
	})
}
