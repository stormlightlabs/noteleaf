package ui

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stormlightlabs/noteleaf/internal/models"
	"github.com/stormlightlabs/noteleaf/internal/repo"
)

// newTestNoteRepo sets up an in-memory SQLite database and returns a [repo.NoteRepository].
func newTestNoteRepo(t *testing.T) (*repo.NoteRepository, func()) {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}

	// Schema based on [models.Note] and observed errors
	createTableSQL := `
	CREATE TABLE notes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		content TEXT,
		tags TEXT,
		created DATETIME NOT NULL,
		modified DATETIME NOT NULL,
		archived BOOLEAN NOT NULL DEFAULT 0,
		file_path TEXT
	);`
	if _, err := db.Exec(createTableSQL); err != nil {
		t.Fatalf("Failed to create notes table: %v", err)
	}

	noteRepo := repo.NewNoteRepository(db)

	cleanup := func() {
		db.Close()
	}

	return noteRepo, cleanup
}

func TestNoteListOptions(t *testing.T) {
	t.Run("default options", func(t *testing.T) {
		repo, cleanup := newTestNoteRepo(t)
		defer cleanup()

		nl := NewNoteList(repo, NoteListOptions{})
		if nl.opts.Static {
			t.Error("Static should default to false")
		}
	})

	t.Run("custom options", func(t *testing.T) {
		repo, cleanup := newTestNoteRepo(t)
		defer cleanup()

		var buf bytes.Buffer
		var in strings.Reader

		opts := NoteListOptions{
			Output: &buf,
			Input:  &in,
			Static: true,
		}

		nl := NewNoteList(repo, opts)

		if !nl.opts.Static {
			t.Error("Static should be true")
		}
		if nl.opts.Output != &buf {
			t.Error("Output should be set to buffer")
		}
		if nl.opts.Input != &in {
			t.Error("Input should be set to reader")
		}
	})
}

func TestStaticList(t *testing.T) {
	ctx := context.Background()

	t.Run("no notes", func(t *testing.T) {
		repo, cleanup := newTestNoteRepo(t)
		defer cleanup()

		var buf bytes.Buffer
		nl := NewNoteList(repo, NoteListOptions{
			Output: &buf,
			Static: true,
		})

		err := nl.Browse(ctx)
		if err != nil {
			t.Fatalf("Browse failed: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "No notes found") {
			t.Errorf("Expected 'No notes found', got %q", output)
		}
	})

	t.Run("with notes", func(t *testing.T) {
		repo, cleanup := newTestNoteRepo(t)
		defer cleanup()

		// Create some instances of [models.Note]
		note1 := &models.Note{Title: "Test Note 1", Content: "Content 1", Tags: []string{"t1"}, Created: time.Now(), Modified: time.Now()}
		note2 := &models.Note{Title: "Test Note 2", Content: "Content 2", Tags: []string{"t2"}, Created: time.Now(), Modified: time.Now()}
		if _, err := repo.Create(ctx, note1); err != nil {
			t.Fatalf("Failed to create note: %v", err)
		}
		if _, err := repo.Create(ctx, note2); err != nil {
			t.Fatalf("Failed to create note: %v", err)
		}

		var buf bytes.Buffer
		nl := NewNoteList(repo, NoteListOptions{
			Output: &buf,
			Static: true,
		})

		err := nl.Browse(ctx)
		if err != nil {
			t.Fatalf("Browse failed: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Test Note 1") {
			t.Error("Output does not contain first note")
		}
		if !strings.Contains(output, "Test Note 2") {
			t.Error("Output does not contain second note")
		}
		if !strings.Contains(output, "t1") {
			t.Error("Output does not contain first note's tag")
		}
	})

	t.Run("with archived notes", func(t *testing.T) {
		repo, cleanup := newTestNoteRepo(t)
		defer cleanup()

		note1 := &models.Note{Title: "Active Note", Created: time.Now(), Modified: time.Now()}
		note2 := &models.Note{Title: "Archived Note", Archived: true, Created: time.Now(), Modified: time.Now()}
		if _, err := repo.Create(ctx, note1); err != nil {
			t.Fatalf("Failed to create note: %v", err)
		}
		if _, err := repo.Create(ctx, note2); err != nil {
			t.Fatalf("Failed to create note: %v", err)
		}

		var buf bytes.Buffer
		nl := NewNoteList(repo, NoteListOptions{
			Output:       &buf,
			Static:       true,
			ShowArchived: false,
		})

		// Test with ShowArchived: false (default behavior)
		err := nl.Browse(ctx)
		if err != nil {
			t.Fatalf("Browse failed: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Active Note") {
			t.Error("Expected to see active note")
		}
		if strings.Contains(output, "Archived Note") {
			t.Error("Did not expect to see archived note")
		}

		buf.Reset()
		nl.opts.ShowArchived = true
		err = nl.Browse(ctx)
		if err != nil {
			t.Fatalf("Browse failed: %v", err)
		}

		output = buf.String()
		if !strings.Contains(output, "Active Note") {
			t.Error("Expected to see active note")
		}
		if !strings.Contains(output, "Archived Note") {
			t.Error("Expected to see archived note")
		}
	})
}

func TestNoteListModelView(t *testing.T) {
	repo, cleanup := newTestNoteRepo(t)
	defer cleanup()

	opts := NoteListOptions{}

	t.Run("initial view", func(t *testing.T) {
		model := noteListModel{
			repo:  repo,
			opts:  opts,
			notes: []*models.Note{},
		}
		view := model.View()
		if !strings.Contains(view, "No notes found") {
			t.Errorf("Expected 'No notes found', got %q", view)
		}
	})

	t.Run("error view", func(t *testing.T) {
		model := noteListModel{
			repo: repo,
			opts: opts,
			err:  errors.New("test error"),
		}
		view := model.View()
		if !strings.Contains(view, "Error: test error") {
			t.Errorf("Expected error message, got %q", view)
		}
	})

	t.Run("with notes view", func(t *testing.T) {
		note := &models.Note{ID: 1, Title: "My Test Note", Tags: []string{"testing"}, Modified: time.Now()}
		model := noteListModel{
			repo:     repo,
			opts:     opts,
			notes:    []*models.Note{note},
			selected: 0,
		}
		view := model.View()
		if !strings.Contains(view, "My Test Note") {
			t.Error("Expected to see note title")
		}
		if !strings.Contains(view, ">") {
			t.Error("Expected to see selection indicator")
		}
	})

	t.Run("viewing note", func(t *testing.T) {
		model := noteListModel{
			repo:        repo,
			opts:        opts,
			viewing:     true,
			viewContent: "## My Note Content",
		}
		view := model.View()
		if !strings.Contains(view, "## My Note Content") {
			t.Error("Expected to see note content")
		}
		if !strings.Contains(view, "Press q/esc/backspace to return to list") {
			t.Error("Expected to see exit instructions")
		}
	})
}

func TestFormatNoteForView(t *testing.T) {
	note := &models.Note{
		Title:    "My Note Title",
		Content:  "This is the content.",
		Tags:     []string{"go", "testing"},
		Created:  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		Modified: time.Date(2023, 1, 1, 13, 0, 0, 0, time.UTC),
	}

	m := noteListModel{}
	formatted := m.formatNoteForView(note)

	if !strings.Contains(formatted, "# My Note Title") {
		t.Error("Expected title markdown")
	}
	if !strings.Contains(formatted, "**Tags:** `go`, `testing`") {
		t.Error("Expected tags markdown")
	}
	if !strings.Contains(formatted, "**Created:** 2023-01-01 12:00") {
		t.Error("Expected created timestamp")
	}
	if !strings.Contains(formatted, "**Modified:** 2023-01-01 13:00") {
		t.Error("Expected modified timestamp")
	}
	if !strings.Contains(formatted, "This is the content.") {
		t.Error("Expected content")
	}
}
