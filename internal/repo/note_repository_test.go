package repo

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stormlightlabs/noteleaf/internal/models"
)

func createNoteTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create in-memory database: %v", err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	schema := `
		CREATE TABLE IF NOT EXISTS notes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			tags TEXT,
			archived BOOLEAN DEFAULT FALSE,
			created DATETIME DEFAULT CURRENT_TIMESTAMP,
			modified DATETIME DEFAULT CURRENT_TIMESTAMP,
			file_path TEXT
		);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func createSampleNote() *models.Note {
	return &models.Note{
		Title:    "Test Note",
		Content:  "This is test content with **markdown**",
		Tags:     []string{"personal", "work"},
		Archived: false,
		FilePath: "/path/to/note.md",
	}
}

func TestNoteRepository_CRUD(t *testing.T) {
	db := createNoteTestDB(t)
	repo := NewNoteRepository(db)
	ctx := context.Background()

	t.Run("Create Note", func(t *testing.T) {
		note := createSampleNote()

		id, err := repo.Create(ctx, note)
		if err != nil {
			t.Errorf("Failed to create note: %v", err)
		}

		if id == 0 {
			t.Error("Expected non-zero ID")
		}

		if note.ID != id {
			t.Errorf("Expected note ID to be set to %d, got %d", id, note.ID)
		}

		if note.Created.IsZero() {
			t.Error("Expected Created timestamp to be set")
		}
		if note.Modified.IsZero() {
			t.Error("Expected Modified timestamp to be set")
		}
	})

	t.Run("Get Note", func(t *testing.T) {
		original := createSampleNote()
		id, err := repo.Create(ctx, original)
		if err != nil {
			t.Fatalf("Failed to create note: %v", err)
		}

		retrieved, err := repo.Get(ctx, id)
		if err != nil {
			t.Fatalf("Failed to get note: %v", err)
		}

		if retrieved.ID != original.ID {
			t.Errorf("Expected ID %d, got %d", original.ID, retrieved.ID)
		}
		if retrieved.Title != original.Title {
			t.Errorf("Expected title %s, got %s", original.Title, retrieved.Title)
		}
		if retrieved.Content != original.Content {
			t.Errorf("Expected content %s, got %s", original.Content, retrieved.Content)
		}
		if len(retrieved.Tags) != len(original.Tags) {
			t.Errorf("Expected %d tags, got %d", len(original.Tags), len(retrieved.Tags))
		}
		if retrieved.Archived != original.Archived {
			t.Errorf("Expected archived %v, got %v", original.Archived, retrieved.Archived)
		}
		if retrieved.FilePath != original.FilePath {
			t.Errorf("Expected file path %s, got %s", original.FilePath, retrieved.FilePath)
		}
	})

	t.Run("Update Note", func(t *testing.T) {
		note := createSampleNote()
		id, err := repo.Create(ctx, note)
		if err != nil {
			t.Fatalf("Failed to create note: %v", err)
		}

		originalModified := note.Modified

		note.Title = "Updated Title"
		note.Content = "Updated content"
		note.Tags = []string{"updated", "test"}
		note.Archived = true
		note.FilePath = "/new/path/note.md"

		err = repo.Update(ctx, note)
		if err != nil {
			t.Errorf("Failed to update note: %v", err)
		}

		retrieved, err := repo.Get(ctx, id)
		if err != nil {
			t.Fatalf("Failed to get updated note: %v", err)
		}

		if retrieved.Title != "Updated Title" {
			t.Errorf("Expected updated title, got %s", retrieved.Title)
		}
		if retrieved.Content != "Updated content" {
			t.Errorf("Expected updated content, got %s", retrieved.Content)
		}
		if len(retrieved.Tags) != 2 || retrieved.Tags[0] != "updated" || retrieved.Tags[1] != "test" {
			t.Errorf("Expected updated tags, got %v", retrieved.Tags)
		}
		if !retrieved.Archived {
			t.Error("Expected note to be archived")
		}
		if retrieved.FilePath != "/new/path/note.md" {
			t.Errorf("Expected updated file path, got %s", retrieved.FilePath)
		}
		if !retrieved.Modified.After(originalModified) {
			t.Error("Expected Modified timestamp to be updated")
		}
	})

	t.Run("Delete Note", func(t *testing.T) {
		note := createSampleNote()
		id, err := repo.Create(ctx, note)
		if err != nil {
			t.Fatalf("Failed to create note: %v", err)
		}

		err = repo.Delete(ctx, id)
		if err != nil {
			t.Errorf("Failed to delete note: %v", err)
		}

		_, err = repo.Get(ctx, id)
		if err == nil {
			t.Error("Expected error when getting deleted note")
		}
	})
}

func TestNoteRepository_List(t *testing.T) {
	db := createNoteTestDB(t)
	repo := NewNoteRepository(db)
	ctx := context.Background()

	notes := []*models.Note{
		{Title: "First Note", Content: "Content 1", Tags: []string{"work"}, Archived: false},
		{Title: "Second Note", Content: "Content 2", Tags: []string{"personal"}, Archived: true},
		{Title: "Third Note", Content: "Important content", Tags: []string{"work", "important"}, Archived: false},
	}

	for _, note := range notes {
		_, err := repo.Create(ctx, note)
		if err != nil {
			t.Fatalf("Failed to create test note: %v", err)
		}
	}

	t.Run("List All Notes", func(t *testing.T) {
		results, err := repo.List(ctx, NoteListOptions{})
		if err != nil {
			t.Fatalf("Failed to list notes: %v", err)
		}

		if len(results) != 3 {
			t.Errorf("Expected 3 notes, got %d", len(results))
		}
	})

	t.Run("List Archived Notes Only", func(t *testing.T) {
		archived := true
		results, err := repo.List(ctx, NoteListOptions{Archived: &archived})
		if err != nil {
			t.Fatalf("Failed to list archived notes: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("Expected 1 archived note, got %d", len(results))
		}
		if !results[0].Archived {
			t.Error("Retrieved note should be archived")
		}
	})

	t.Run("List Active Notes Only", func(t *testing.T) {
		archived := false
		results, err := repo.List(ctx, NoteListOptions{Archived: &archived})
		if err != nil {
			t.Fatalf("Failed to list active notes: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("Expected 2 active notes, got %d", len(results))
		}
		for _, note := range results {
			if note.Archived {
				t.Error("Retrieved note should not be archived")
			}
		}
	})

	t.Run("Search by Title", func(t *testing.T) {
		results, err := repo.List(ctx, NoteListOptions{Title: "First"})
		if err != nil {
			t.Fatalf("Failed to search by title: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("Expected 1 note, got %d", len(results))
		}
		if results[0].Title != "First Note" {
			t.Errorf("Expected 'First Note', got %s", results[0].Title)
		}
	})

	t.Run("Search by Content", func(t *testing.T) {
		results, err := repo.List(ctx, NoteListOptions{Content: "Important"})
		if err != nil {
			t.Fatalf("Failed to search by content: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("Expected 1 note, got %d", len(results))
		}
		if results[0].Title != "Third Note" {
			t.Errorf("Expected 'Third Note', got %s", results[0].Title)
		}
	})

	t.Run("Limit and Offset", func(t *testing.T) {
		results, err := repo.List(ctx, NoteListOptions{Limit: 2})
		if err != nil {
			t.Fatalf("Failed to list with limit: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("Expected 2 notes, got %d", len(results))
		}

		results, err = repo.List(ctx, NoteListOptions{Limit: 2, Offset: 1})
		if err != nil {
			t.Fatalf("Failed to list with limit and offset: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("Expected 2 notes with offset, got %d", len(results))
		}
	})
}

func TestNoteRepository_SpecializedMethods(t *testing.T) {
	db := createNoteTestDB(t)
	repo := NewNoteRepository(db)
	ctx := context.Background()

	notes := []*models.Note{
		{Title: "Work Note", Content: "Work content", Tags: []string{"work"}, Archived: false},
		{Title: "Personal Note", Content: "Personal content", Tags: []string{"personal"}, Archived: true},
		{Title: "Important Note", Content: "Important content", Tags: []string{"work", "important"}, Archived: false},
	}

	for _, note := range notes {
		_, err := repo.Create(ctx, note)
		if err != nil {
			t.Fatalf("Failed to create test note: %v", err)
		}
	}

	t.Run("GetByTitle", func(t *testing.T) {
		results, err := repo.GetByTitle(ctx, "Work")
		if err != nil {
			t.Fatalf("Failed to get by title: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("Expected 1 note, got %d", len(results))
		}
		if results[0].Title != "Work Note" {
			t.Errorf("Expected 'Work Note', got %s", results[0].Title)
		}
	})

	t.Run("GetArchived", func(t *testing.T) {
		results, err := repo.GetArchived(ctx)
		if err != nil {
			t.Fatalf("Failed to get archived notes: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("Expected 1 archived note, got %d", len(results))
		}
		if !results[0].Archived {
			t.Error("Retrieved note should be archived")
		}
	})

	t.Run("GetActive", func(t *testing.T) {
		results, err := repo.GetActive(ctx)
		if err != nil {
			t.Fatalf("Failed to get active notes: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("Expected 2 active notes, got %d", len(results))
		}
		for _, note := range results {
			if note.Archived {
				t.Error("Retrieved note should not be archived")
			}
		}
	})

	t.Run("Archive and Unarchive", func(t *testing.T) {
		note := &models.Note{
			Title:    "Test Archive",
			Content:  "Archive test",
			Archived: false,
		}
		id, err := repo.Create(ctx, note)
		if err != nil {
			t.Fatalf("Failed to create note: %v", err)
		}

		err = repo.Archive(ctx, id)
		if err != nil {
			t.Fatalf("Failed to archive note: %v", err)
		}

		retrieved, err := repo.Get(ctx, id)
		if err != nil {
			t.Fatalf("Failed to get note: %v", err)
		}
		if !retrieved.Archived {
			t.Error("Note should be archived")
		}

		err = repo.Unarchive(ctx, id)
		if err != nil {
			t.Fatalf("Failed to unarchive note: %v", err)
		}

		retrieved, err = repo.Get(ctx, id)
		if err != nil {
			t.Fatalf("Failed to get note: %v", err)
		}
		if retrieved.Archived {
			t.Error("Note should not be archived")
		}
	})

	t.Run("SearchContent", func(t *testing.T) {
		results, err := repo.SearchContent(ctx, "Important")
		if err != nil {
			t.Fatalf("Failed to search content: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("Expected 1 note, got %d", len(results))
		}
		if results[0].Title != "Important Note" {
			t.Errorf("Expected 'Important Note', got %s", results[0].Title)
		}
	})

	t.Run("GetRecent", func(t *testing.T) {
		results, err := repo.GetRecent(ctx, 2)
		if err != nil {
			t.Fatalf("Failed to get recent notes: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("Expected 2 notes, got %d", len(results))
		}
	})
}

func TestNoteRepository_TagMethods(t *testing.T) {
	db := createNoteTestDB(t)
	repo := NewNoteRepository(db)
	ctx := context.Background()

	note := &models.Note{
		Title:   "Tag Test Note",
		Content: "Testing tags",
		Tags:    []string{"initial"},
	}
	id, err := repo.Create(ctx, note)
	if err != nil {
		t.Fatalf("Failed to create note: %v", err)
	}

	t.Run("AddTag", func(t *testing.T) {
		err := repo.AddTag(ctx, id, "new-tag")
		if err != nil {
			t.Fatalf("Failed to add tag: %v", err)
		}

		retrieved, err := repo.Get(ctx, id)
		if err != nil {
			t.Fatalf("Failed to get note: %v", err)
		}

		if len(retrieved.Tags) != 2 {
			t.Errorf("Expected 2 tags, got %d", len(retrieved.Tags))
		}

		found := false
		for _, tag := range retrieved.Tags {
			if tag == "new-tag" {
				found = true
				break
			}
		}
		if !found {
			t.Error("New tag not found in note")
		}
	})

	t.Run("AddTag Duplicate", func(t *testing.T) {
		err := repo.AddTag(ctx, id, "new-tag")
		if err != nil {
			t.Fatalf("Failed to add duplicate tag: %v", err)
		}

		retrieved, err := repo.Get(ctx, id)
		if err != nil {
			t.Fatalf("Failed to get note: %v", err)
		}

		if len(retrieved.Tags) != 2 {
			t.Errorf("Expected 2 tags (no duplicate), got %d", len(retrieved.Tags))
		}
	})

	t.Run("RemoveTag", func(t *testing.T) {
		err := repo.RemoveTag(ctx, id, "initial")
		if err != nil {
			t.Fatalf("Failed to remove tag: %v", err)
		}

		retrieved, err := repo.Get(ctx, id)
		if err != nil {
			t.Fatalf("Failed to get note: %v", err)
		}

		if len(retrieved.Tags) != 1 {
			t.Errorf("Expected 1 tag after removal, got %d", len(retrieved.Tags))
		}

		for _, tag := range retrieved.Tags {
			if tag == "initial" {
				t.Error("Removed tag still found in note")
			}
		}
	})

	t.Run("GetByTags", func(t *testing.T) {
		note1 := &models.Note{
			Title:   "Note 1",
			Content: "Content 1",
			Tags:    []string{"work", "urgent"},
		}
		note2 := &models.Note{
			Title:   "Note 2",
			Content: "Content 2",
			Tags:    []string{"personal", "ideas"},
		}
		note3 := &models.Note{
			Title:   "Note 3",
			Content: "Content 3",
			Tags:    []string{"work", "planning"},
		}

		_, err := repo.Create(ctx, note1)
		if err != nil {
			t.Fatalf("Failed to create note1: %v", err)
		}
		_, err = repo.Create(ctx, note2)
		if err != nil {
			t.Fatalf("Failed to create note2: %v", err)
		}
		_, err = repo.Create(ctx, note3)
		if err != nil {
			t.Fatalf("Failed to create note3: %v", err)
		}

		results, err := repo.GetByTags(ctx, []string{"work"})
		if err != nil {
			t.Fatalf("Failed to get notes by tag: %v", err)
		}

		if len(results) < 2 {
			t.Errorf("Expected at least 2 notes with 'work' tag, got %d", len(results))
		}

		results, err = repo.GetByTags(ctx, []string{"nonexistent"})
		if err != nil {
			t.Fatalf("Failed to get notes by nonexistent tag: %v", err)
		}

		if len(results) != 0 {
			t.Errorf("Expected 0 notes with nonexistent tag, got %d", len(results))
		}

		results, err = repo.GetByTags(ctx, []string{})
		if err != nil {
			t.Fatalf("Failed to get notes with empty tags: %v", err)
		}

		if len(results) != 0 {
			t.Errorf("Expected 0 notes with empty tag list, got %d", len(results))
		}
	})
}

func TestNoteRepository_ErrorCases(t *testing.T) {
	db := createNoteTestDB(t)
	repo := NewNoteRepository(db)
	ctx := context.Background()

	t.Run("Get Nonexistent Note", func(t *testing.T) {
		_, err := repo.Get(ctx, 999)
		if err == nil {
			t.Error("Expected error when getting nonexistent note")
		}
	})

	t.Run("Update Nonexistent Note", func(t *testing.T) {
		note := &models.Note{
			ID:      999,
			Title:   "Nonexistent",
			Content: "Should fail",
		}

		err := repo.Update(ctx, note)
		if err == nil {
			t.Error("Expected error when updating nonexistent note")
		}
	})

	t.Run("Delete Nonexistent Note", func(t *testing.T) {
		err := repo.Delete(ctx, 999)
		if err == nil {
			t.Error("Expected error when deleting nonexistent note")
		}
	})

	t.Run("Archive Nonexistent Note", func(t *testing.T) {
		err := repo.Archive(ctx, 999)
		if err == nil {
			t.Error("Expected error when archiving nonexistent note")
		}
	})

	t.Run("AddTag to Nonexistent Note", func(t *testing.T) {
		err := repo.AddTag(ctx, 999, "tag")
		if err == nil {
			t.Error("Expected error when adding tag to nonexistent note")
		}
	})
}

func TestNoteRepository_EdgeCases(t *testing.T) {
	db := createNoteTestDB(t)
	repo := NewNoteRepository(db)
	ctx := context.Background()

	t.Run("Note with Empty Tags", func(t *testing.T) {
		note := &models.Note{
			Title:   "No Tags Note",
			Content: "This note has no tags",
			Tags:    []string{},
		}

		id, err := repo.Create(ctx, note)
		if err != nil {
			t.Fatalf("Failed to create note with empty tags: %v", err)
		}

		retrieved, err := repo.Get(ctx, id)
		if err != nil {
			t.Fatalf("Failed to get note: %v", err)
		}

		if len(retrieved.Tags) != 0 {
			t.Errorf("Expected empty tags slice, got %d tags", len(retrieved.Tags))
		}
	})

	t.Run("Note with Nil Tags", func(t *testing.T) {
		note := &models.Note{
			Title:   "Nil Tags Note",
			Content: "This note has nil tags",
			Tags:    nil,
		}

		id, err := repo.Create(ctx, note)
		if err != nil {
			t.Fatalf("Failed to create note with nil tags: %v", err)
		}

		retrieved, err := repo.Get(ctx, id)
		if err != nil {
			t.Fatalf("Failed to get note: %v", err)
		}

		if retrieved.Tags != nil {
			t.Errorf("Expected nil tags, got %v", retrieved.Tags)
		}
	})

	t.Run("Note with Long Content", func(t *testing.T) {
		longContent := ""
		for i := 0; i < 1000; i++ {
			longContent += "This is a very long content string. "
		}

		note := &models.Note{
			Title:   "Long Content Note",
			Content: longContent,
		}

		id, err := repo.Create(ctx, note)
		if err != nil {
			t.Fatalf("Failed to create note with long content: %v", err)
		}

		retrieved, err := repo.Get(ctx, id)
		if err != nil {
			t.Fatalf("Failed to get note: %v", err)
		}

		if retrieved.Content != longContent {
			t.Error("Long content was not stored/retrieved correctly")
		}
	})
}