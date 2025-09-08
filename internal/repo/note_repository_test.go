package repo

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stormlightlabs/noteleaf/internal/models"
)

func TestNoteRepository(t *testing.T) {
	t.Run("CRUD", func(t *testing.T) {
		db := CreateTestDB(t)
		repo := NewNoteRepository(db)
		ctx := context.Background()

		t.Run("Create Note", func(t *testing.T) {
			note := CreateSampleNote()

			id, err := repo.Create(ctx, note)
			AssertNoError(t, err, "Failed to create note")
			AssertNotEqual(t, int64(0), id, "Expected non-zero ID")
			AssertEqual(t, id, note.ID, "Expected note ID to be set correctly")
			AssertFalse(t, note.Created.IsZero(), "Expected Created timestamp to be set")
			AssertFalse(t, note.Modified.IsZero(), "Expected Modified timestamp to be set")
		})

		t.Run("Get Note", func(t *testing.T) {
			original := CreateSampleNote()
			id, err := repo.Create(ctx, original)
			AssertNoError(t, err, "Failed to create note")

			retrieved, err := repo.Get(ctx, id)
			AssertNoError(t, err, "Failed to get note")

			AssertEqual(t, original.ID, retrieved.ID, "ID mismatch")
			AssertEqual(t, original.Title, retrieved.Title, "Title mismatch")
			AssertEqual(t, original.Content, retrieved.Content, "Content mismatch")
			AssertEqual(t, len(original.Tags), len(retrieved.Tags), "Tags length mismatch")
			AssertEqual(t, original.Archived, retrieved.Archived, "Archived mismatch")
			AssertEqual(t, original.FilePath, retrieved.FilePath, "FilePath mismatch")
		})

		t.Run("Update Note", func(t *testing.T) {
			note := CreateSampleNote()
			id, err := repo.Create(ctx, note)
			AssertNoError(t, err, "Failed to create note")

			originalModified := note.Modified

			note.Title = "Updated Title"
			note.Content = "Updated content"
			note.Tags = []string{"updated", "test"}
			note.Archived = true
			note.FilePath = "/new/path/note.md"

			err = repo.Update(ctx, note)
			AssertNoError(t, err, "Failed to update note")

			retrieved, err := repo.Get(ctx, id)
			AssertNoError(t, err, "Failed to get updated note")

			AssertEqual(t, "Updated Title", retrieved.Title, "Expected updated title")
			AssertEqual(t, "Updated content", retrieved.Content, "Expected updated content")
			AssertEqual(t, 2, len(retrieved.Tags), "Expected 2 tags")
			if len(retrieved.Tags) >= 2 {
				AssertEqual(t, "updated", retrieved.Tags[0], "Expected first tag to be 'updated'")
				AssertEqual(t, "test", retrieved.Tags[1], "Expected second tag to be 'test'")
			}
			AssertTrue(t, retrieved.Archived, "Expected note to be archived")
			AssertEqual(t, "/new/path/note.md", retrieved.FilePath, "Expected updated file path")
			AssertTrue(t, retrieved.Modified.After(originalModified), "Expected Modified timestamp to be updated")
		})

		t.Run("Delete Note", func(t *testing.T) {
			note := CreateSampleNote()
			id, err := repo.Create(ctx, note)
			AssertNoError(t, err, "Failed to create note")

			err = repo.Delete(ctx, id)
			AssertNoError(t, err, "Failed to delete note")

			_, err = repo.Get(ctx, id)
			AssertError(t, err, "Expected error when getting deleted note")
		})
	})

	t.Run("List", func(t *testing.T) {
		db := CreateTestDB(t)
		repo := NewNoteRepository(db)
		ctx := context.Background()

		notes := []*models.Note{
			{Title: "First Note", Content: "Content 1", Tags: []string{"work"}, Archived: false},
			{Title: "Second Note", Content: "Content 2", Tags: []string{"personal"}, Archived: true},
			{Title: "Third Note", Content: "Important content", Tags: []string{"work", "important"}, Archived: false},
		}

		for _, note := range notes {
			_, err := repo.Create(ctx, note)
			AssertNoError(t, err, "Failed to create test note")
		}

		t.Run("List All Notes", func(t *testing.T) {
			results, err := repo.List(ctx, NoteListOptions{})
			AssertNoError(t, err, "Failed to list notes")
			AssertEqual(t, 3, len(results), "Expected 3 notes")
		})

		t.Run("List Archived Notes Only", func(t *testing.T) {
			archived := true
			results, err := repo.List(ctx, NoteListOptions{Archived: &archived})
			AssertNoError(t, err, "Failed to list archived notes")
			AssertEqual(t, 1, len(results), "Expected 1 archived note")
			if len(results) > 0 {
				AssertTrue(t, results[0].Archived, "Retrieved note should be archived")
			}
		})

		t.Run("List Active Notes Only", func(t *testing.T) {
			archived := false
			results, err := repo.List(ctx, NoteListOptions{Archived: &archived})
			AssertNoError(t, err, "Failed to list active notes")
			AssertEqual(t, 2, len(results), "Expected 2 active notes")
			for _, note := range results {
				AssertFalse(t, note.Archived, "Retrieved note should not be archived")
			}
		})

		t.Run("Search by Title", func(t *testing.T) {
			results, err := repo.List(ctx, NoteListOptions{Title: "First"})
			AssertNoError(t, err, "Failed to search by title")
			AssertEqual(t, 1, len(results), "Expected 1 note")
			if len(results) > 0 {
				AssertEqual(t, "First Note", results[0].Title, "Expected 'First Note'")
			}
		})

		t.Run("Search by Content", func(t *testing.T) {
			results, err := repo.List(ctx, NoteListOptions{Content: "Important"})
			AssertNoError(t, err, "Failed to search by content")
			AssertEqual(t, 1, len(results), "Expected 1 note")
			if len(results) > 0 {
				AssertEqual(t, "Third Note", results[0].Title, "Expected 'Third Note'")
			}
		})

		t.Run("Limit and Offset", func(t *testing.T) {
			results, err := repo.List(ctx, NoteListOptions{Limit: 2})
			AssertNoError(t, err, "Failed to list with limit")
			AssertEqual(t, 2, len(results), "Expected 2 notes")

			results, err = repo.List(ctx, NoteListOptions{Limit: 2, Offset: 1})
			AssertNoError(t, err, "Failed to list with limit and offset")
			AssertEqual(t, 2, len(results), "Expected 2 notes with offset")
		})
	})

	t.Run("Specialized Methods", func(t *testing.T) {
		db := CreateTestDB(t)
		repo := NewNoteRepository(db)
		ctx := context.Background()

		notes := []*models.Note{
			{Title: "Work Note", Content: "Work content", Tags: []string{"work"}, Archived: false},
			{Title: "Personal Note", Content: "Personal content", Tags: []string{"personal"}, Archived: true},
			{Title: "Important Note", Content: "Important content", Tags: []string{"work", "important"}, Archived: false},
		}

		for _, note := range notes {
			_, err := repo.Create(ctx, note)
			AssertNoError(t, err, "Failed to create test note")
		}

		t.Run("GetByTitle", func(t *testing.T) {
			results, err := repo.GetByTitle(ctx, "Work")
			AssertNoError(t, err, "Failed to get by title")
			AssertEqual(t, 1, len(results), "Expected 1 note")
			if len(results) > 0 {
				AssertEqual(t, "Work Note", results[0].Title, "Expected 'Work Note'")
			}
		})

		t.Run("GetArchived", func(t *testing.T) {
			results, err := repo.GetArchived(ctx)
			AssertNoError(t, err, "Failed to get archived notes")
			AssertEqual(t, 1, len(results), "Expected 1 archived note")
			if len(results) > 0 {
				AssertTrue(t, results[0].Archived, "Retrieved note should be archived")
			}
		})

		t.Run("GetActive", func(t *testing.T) {
			results, err := repo.GetActive(ctx)
			AssertNoError(t, err, "Failed to get active notes")
			AssertEqual(t, 2, len(results), "Expected 2 active notes")
			for _, note := range results {
				AssertFalse(t, note.Archived, "Retrieved note should not be archived")
			}
		})

		t.Run("Archive and Unarchive", func(t *testing.T) {
			note := &models.Note{
				Title:    "Test Archive",
				Content:  "Archive test",
				Archived: false,
			}
			id, err := repo.Create(ctx, note)
			AssertNoError(t, err, "Failed to create note")

			err = repo.Archive(ctx, id)
			AssertNoError(t, err, "Failed to archive note")

			retrieved, err := repo.Get(ctx, id)
			AssertNoError(t, err, "Failed to get note")
			AssertTrue(t, retrieved.Archived, "Note should be archived")

			err = repo.Unarchive(ctx, id)
			AssertNoError(t, err, "Failed to unarchive note")

			retrieved, err = repo.Get(ctx, id)
			AssertNoError(t, err, "Failed to get note")
			AssertFalse(t, retrieved.Archived, "Note should not be archived")
		})

		t.Run("SearchContent", func(t *testing.T) {
			results, err := repo.SearchContent(ctx, "Important")
			AssertNoError(t, err, "Failed to search content")
			AssertEqual(t, 1, len(results), "Expected 1 note")
			if len(results) > 0 {
				AssertEqual(t, "Important Note", results[0].Title, "Expected 'Important Note'")
			}
		})

		t.Run("GetRecent", func(t *testing.T) {
			results, err := repo.GetRecent(ctx, 2)
			AssertNoError(t, err, "Failed to get recent notes")
			AssertEqual(t, 2, len(results), "Expected 2 notes")
		})
	})

	t.Run("Tag Methods", func(t *testing.T) {
		db := CreateTestDB(t)
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
			AssertNoError(t, err, "Failed to add duplicate tag")

			retrieved, err := repo.Get(ctx, id)
			AssertNoError(t, err, "Failed to get note")

			AssertEqual(t, 2, len(retrieved.Tags), "Expected 2 tags (no duplicate)")
		})

		t.Run("RemoveTag", func(t *testing.T) {
			err := repo.RemoveTag(ctx, id, "initial")
			AssertNoError(t, err, "Failed to remove tag")

			retrieved, err := repo.Get(ctx, id)
			AssertNoError(t, err, "Failed to get note")

			AssertEqual(t, 1, len(retrieved.Tags), "Expected 1 tag after removal")

			for _, tag := range retrieved.Tags {
				AssertNotEqual(t, "initial", tag, "Removed tag still found in note")
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
	})

	t.Run("Error Cases", func(t *testing.T) {
		db := CreateTestDB(t)
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
	})

	t.Run("Edge Cases", func(t *testing.T) {
		db := CreateTestDB(t)
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
	})
}
