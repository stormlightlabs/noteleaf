package repo

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stormlightlabs/noteleaf/internal/models"
)

func TestNoteRepository(t *testing.T) {
	t.Run("CRUD Operations", func(t *testing.T) {
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

	t.Run("Special Methods", func(t *testing.T) {
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
		AssertNoError(t, err, "Failed to create note")

		t.Run("AddTag", func(t *testing.T) {
			err := repo.AddTag(ctx, id, "new-tag")
			AssertNoError(t, err, "Failed to add tag")

			retrieved, err := repo.Get(ctx, id)
			AssertNoError(t, err, "Failed to get note")

			AssertEqual(t, 2, len(retrieved.Tags), "Expected 2 tags")

			found := false
			for _, tag := range retrieved.Tags {
				if tag == "new-tag" {
					found = true
					break
				}
			}
			AssertTrue(t, found, "New tag not found in note")
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
			AssertNoError(t, err, "Failed to create note1")
			_, err = repo.Create(ctx, note2)
			AssertNoError(t, err, "Failed to create note2")
			_, err = repo.Create(ctx, note3)
			AssertNoError(t, err, "Failed to create note3")

			results, err := repo.GetByTags(ctx, []string{"work"})
			AssertNoError(t, err, "Failed to get notes by tag")
			AssertTrue(t, len(results) >= 2, "Expected at least 2 notes with 'work' tag")

			results, err = repo.GetByTags(ctx, []string{"nonexistent"})
			AssertNoError(t, err, "Failed to get notes by nonexistent tag")
			AssertEqual(t, 0, len(results), "Expected 0 notes with nonexistent tag")

			results, err = repo.GetByTags(ctx, []string{})
			AssertNoError(t, err, "Failed to get notes with empty tags")
			AssertEqual(t, 0, len(results), "Expected 0 notes with empty tag list")
		})
	})

	t.Run("Context Cancellation Error Paths", func(t *testing.T) {
		db := CreateTestDB(t)
		repo := NewNoteRepository(db)
		ctx := context.Background()

		note := NewNoteBuilder().WithTitle("Test Note").WithContent("Test content").Build()
		id, err := repo.Create(ctx, note)
		AssertNoError(t, err, "Failed to create note")

		t.Run("Create with cancelled context", func(t *testing.T) {
			newNote := NewNoteBuilder().WithTitle("Cancelled").Build()
			_, err := repo.Create(NewCanceledContext(), newNote)
			AssertError(t, err, "Expected error with cancelled context")
		})

		t.Run("Get with cancelled context", func(t *testing.T) {
			_, err := repo.Get(NewCanceledContext(), id)
			AssertError(t, err, "Expected error with cancelled context")
		})

		t.Run("Update with cancelled context", func(t *testing.T) {
			note.Title = "Updated"
			err := repo.Update(NewCanceledContext(), note)
			AssertError(t, err, "Expected error with cancelled context")
		})

		t.Run("Delete with cancelled context", func(t *testing.T) {
			err := repo.Delete(NewCanceledContext(), id)
			AssertError(t, err, "Expected error with cancelled context")
		})

		t.Run("List with cancelled context", func(t *testing.T) {
			_, err := repo.List(NewCanceledContext(), NoteListOptions{})
			AssertError(t, err, "Expected error with cancelled context")
		})

		t.Run("GetByTitle with cancelled context", func(t *testing.T) {
			_, err := repo.GetByTitle(NewCanceledContext(), "Test")
			AssertError(t, err, "Expected error with cancelled context")
		})

		t.Run("GetArchived with cancelled context", func(t *testing.T) {
			_, err := repo.GetArchived(NewCanceledContext())
			AssertError(t, err, "Expected error with cancelled context")
		})

		t.Run("GetActive with cancelled context", func(t *testing.T) {
			_, err := repo.GetActive(NewCanceledContext())
			AssertError(t, err, "Expected error with cancelled context")
		})

		t.Run("Archive with cancelled context", func(t *testing.T) {
			err := repo.Archive(NewCanceledContext(), id)
			AssertError(t, err, "Expected error with cancelled context")
		})

		t.Run("Unarchive with cancelled context", func(t *testing.T) {
			err := repo.Unarchive(NewCanceledContext(), id)
			AssertError(t, err, "Expected error with cancelled context")
		})

		t.Run("SearchContent with cancelled context", func(t *testing.T) {
			_, err := repo.SearchContent(NewCanceledContext(), "test")
			AssertError(t, err, "Expected error with cancelled context")
		})

		t.Run("GetRecent with cancelled context", func(t *testing.T) {
			_, err := repo.GetRecent(NewCanceledContext(), 10)
			AssertError(t, err, "Expected error with cancelled context")
		})

		t.Run("AddTag with cancelled context", func(t *testing.T) {
			err := repo.AddTag(NewCanceledContext(), id, "tag")
			AssertError(t, err, "Expected error with cancelled context")
		})

		t.Run("RemoveTag with cancelled context", func(t *testing.T) {
			err := repo.RemoveTag(NewCanceledContext(), id, "tag")
			AssertError(t, err, "Expected error with cancelled context")
		})

		t.Run("GetByTags with cancelled context", func(t *testing.T) {
			_, err := repo.GetByTags(NewCanceledContext(), []string{"tag"})
			AssertError(t, err, "Expected error with cancelled context")
		})
	})

	t.Run("Edge Cases", func(t *testing.T) {
		db := CreateTestDB(t)
		repo := NewNoteRepository(db)
		ctx := context.Background()

		t.Run("Get non-existent note", func(t *testing.T) {
			_, err := repo.Get(ctx, 99999)
			AssertError(t, err, "Expected error for non-existent note")
		})

		t.Run("Update non-existent note", func(t *testing.T) {
			note := &models.Note{
				ID:      99999,
				Title:   "Nonexistent",
				Content: "Should fail",
			}

			err := repo.Update(ctx, note)
			AssertError(t, err, "Expected error when updating non-existent note")
		})

		t.Run("Delete non-existent note", func(t *testing.T) {
			err := repo.Delete(ctx, 99999)
			AssertError(t, err, "Expected error when deleting non-existent note")
		})

		t.Run("Archive non-existent note", func(t *testing.T) {
			err := repo.Archive(ctx, 99999)
			AssertError(t, err, "Expected error when archiving non-existent note")
		})

		t.Run("AddTag to non-existent note", func(t *testing.T) {
			err := repo.AddTag(ctx, 99999, "tag")
			AssertError(t, err, "Expected error when adding tag to non-existent note")
		})

		t.Run("Note with empty tags", func(t *testing.T) {
			note := &models.Note{
				Title:   "No Tags Note",
				Content: "This note has no tags",
				Tags:    []string{},
			}

			id, err := repo.Create(ctx, note)
			AssertNoError(t, err, "Failed to create note with empty tags")

			retrieved, err := repo.Get(ctx, id)
			AssertNoError(t, err, "Failed to get note")

			AssertEqual(t, 0, len(retrieved.Tags), "Expected empty tags slice")
		})

		t.Run("Note with nil tags", func(t *testing.T) {
			note := &models.Note{
				Title:   "Nil Tags Note",
				Content: "This note has nil tags",
				Tags:    nil,
			}

			id, err := repo.Create(ctx, note)
			AssertNoError(t, err, "Failed to create note with nil tags")

			retrieved, err := repo.Get(ctx, id)
			AssertNoError(t, err, "Failed to get note")

			AssertEqual(t, 0, len(retrieved.Tags), "Expected empty tags")
		})

		t.Run("Note with long content", func(t *testing.T) {
			longContent := ""
			for i := 0; i < 1000; i++ {
				longContent += "This is a very long content string. "
			}

			note := &models.Note{
				Title:   "Long Content Note",
				Content: longContent,
			}

			id, err := repo.Create(ctx, note)
			AssertNoError(t, err, "Failed to create note with long content")

			retrieved, err := repo.Get(ctx, id)
			AssertNoError(t, err, "Failed to get note")

			AssertEqual(t, longContent, retrieved.Content, "Long content was not stored/retrieved correctly")
		})

		t.Run("List with no results", func(t *testing.T) {
			notes, err := repo.List(ctx, NoteListOptions{Title: "NonexistentTitle"})
			AssertNoError(t, err, "Should not error when no notes found")
			AssertEqual(t, 0, len(notes), "Expected empty result set")
		})
	})
}
