package repo

import (
	"context"
	"testing"
	"time"

	"github.com/stormlightlabs/noteleaf/internal/documents"
	"github.com/stormlightlabs/noteleaf/internal/shared"
)

func CreateSampleDocument() *documents.Document {
	return &documents.Document{
		Title:     "Test Document",
		Body:      "This is test content for searching",
		CreatedAt: time.Now(),
		DocKind:   int64(documents.NoteDoc),
	}
}

func TestDocumentRepository(t *testing.T) {
	db := CreateTestDB(t)
	repo := NewDocumentRepository(db)
	ctx := context.Background()

	t.Run("Create", func(t *testing.T) {
		t.Run("creates document successfully", func(t *testing.T) {
			doc := CreateSampleDocument()
			id, err := repo.Create(ctx, doc)

			shared.AssertNoError(t, err, "create should succeed")
			shared.AssertTrue(t, id > 0, "id should be positive")
			shared.AssertEqual(t, id, doc.ID, "document ID should be set")
		})

		t.Run("returns error with cancelled context", func(t *testing.T) {
			doc := CreateSampleDocument()
			canceledCtx := NewCanceledContext()

			_, err := repo.Create(canceledCtx, doc)
			AssertCancelledContext(t, err)
		})
	})

	t.Run("Get", func(t *testing.T) {
		t.Run("retrieves existing document", func(t *testing.T) {
			doc := CreateSampleDocument()
			id, err := repo.Create(ctx, doc)
			shared.AssertNoError(t, err, "create should succeed")

			retrieved, err := repo.Get(ctx, id)
			shared.AssertNoError(t, err, "get should succeed")
			shared.AssertEqual(t, doc.Title, retrieved.Title, "title should match")
			shared.AssertEqual(t, doc.Body, retrieved.Body, "body should match")
			shared.AssertEqual(t, doc.DocKind, retrieved.DocKind, "doc kind should match")
		})

		t.Run("returns error for non-existent document", func(t *testing.T) {
			_, err := repo.Get(ctx, 99999)
			shared.AssertError(t, err, "should return error for non-existent document")
			shared.AssertContains(t, err.Error(), "not found", "error should mention not found")
		})

		t.Run("returns error with cancelled context", func(t *testing.T) {
			canceledCtx := NewCanceledContext()
			_, err := repo.Get(canceledCtx, 1)
			AssertCancelledContext(t, err)
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("deletes existing document", func(t *testing.T) {
			doc := CreateSampleDocument()
			id, err := repo.Create(ctx, doc)
			shared.AssertNoError(t, err, "create should succeed")

			err = repo.Delete(ctx, id)
			shared.AssertNoError(t, err, "delete should succeed")

			_, err = repo.Get(ctx, id)
			shared.AssertError(t, err, "get after delete should fail")
		})

		t.Run("returns error for non-existent document", func(t *testing.T) {
			err := repo.Delete(ctx, 99999)
			shared.AssertError(t, err, "should return error for non-existent document")
		})

		t.Run("returns error with cancelled context", func(t *testing.T) {
			canceledCtx := NewCanceledContext()
			err := repo.Delete(canceledCtx, 1)
			AssertCancelledContext(t, err)
		})
	})

	t.Run("List", func(t *testing.T) {
		t.Run("returns all documents", func(t *testing.T) {
			db := CreateTestDB(t)
			repo := NewDocumentRepository(db)

			doc1 := CreateSampleDocument()
			doc1.Title = "First"
			doc2 := CreateSampleDocument()
			doc2.Title = "Second"

			_, err := repo.Create(ctx, doc1)
			shared.AssertNoError(t, err, "create doc1 should succeed")
			_, err = repo.Create(ctx, doc2)
			shared.AssertNoError(t, err, "create doc2 should succeed")

			docs, err := repo.List(ctx)
			shared.AssertNoError(t, err, "list should succeed")
			shared.AssertEqual(t, 2, len(docs), "should return 2 documents")
		})

		t.Run("returns empty list when no documents exist", func(t *testing.T) {
			db := CreateTestDB(t)
			repo := NewDocumentRepository(db)

			docs, err := repo.List(ctx)
			shared.AssertNoError(t, err, "list should succeed")
			shared.AssertEqual(t, 0, len(docs), "should return empty list")
		})

		t.Run("returns error with cancelled context", func(t *testing.T) {
			canceledCtx := NewCanceledContext()
			_, err := repo.List(canceledCtx)
			AssertCancelledContext(t, err)
		})
	})

	t.Run("ListByKind", func(t *testing.T) {
		t.Run("filters documents by kind", func(t *testing.T) {
			db := CreateTestDB(t)
			repo := NewDocumentRepository(db)

			noteDoc := CreateSampleDocument()
			noteDoc.DocKind = int64(documents.NoteDoc)
			articleDoc := CreateSampleDocument()
			articleDoc.DocKind = int64(documents.ArticleDoc)

			_, err := repo.Create(ctx, noteDoc)
			shared.AssertNoError(t, err, "create note should succeed")
			_, err = repo.Create(ctx, articleDoc)
			shared.AssertNoError(t, err, "create article should succeed")

			notes, err := repo.ListByKind(ctx, documents.NoteDoc)
			shared.AssertNoError(t, err, "list by kind should succeed")
			shared.AssertEqual(t, 1, len(notes), "should return 1 note")
			shared.AssertEqual(t, int64(documents.NoteDoc), notes[0].DocKind, "should be note kind")
		})

		t.Run("returns error with cancelled context", func(t *testing.T) {
			canceledCtx := NewCanceledContext()
			_, err := repo.ListByKind(canceledCtx, documents.NoteDoc)
			AssertCancelledContext(t, err)
		})
	})

	t.Run("DeleteAll", func(t *testing.T) {
		t.Run("removes all documents", func(t *testing.T) {
			db := CreateTestDB(t)
			repo := NewDocumentRepository(db)

			_, err := repo.Create(ctx, CreateSampleDocument())
			shared.AssertNoError(t, err, "create should succeed")
			_, err = repo.Create(ctx, CreateSampleDocument())
			shared.AssertNoError(t, err, "create should succeed")

			err = repo.DeleteAll(ctx)
			shared.AssertNoError(t, err, "delete all should succeed")

			docs, err := repo.List(ctx)
			shared.AssertNoError(t, err, "list should succeed")
			shared.AssertEqual(t, 0, len(docs), "should have no documents after delete all")
		})

		t.Run("returns error with cancelled context", func(t *testing.T) {
			canceledCtx := NewCanceledContext()
			err := repo.DeleteAll(canceledCtx)
			AssertCancelledContext(t, err)
		})
	})

	t.Run("RebuildFromNotes", func(t *testing.T) {
		t.Run("creates documents from notes", func(t *testing.T) {
			db := CreateTestDB(t)
			repo := NewDocumentRepository(db)
			noteRepo := NewNoteRepository(db)

			note1 := CreateSampleNote()
			note1.Title = "Note 1"
			note2 := CreateSampleNote()
			note2.Title = "Note 2"

			_, err := noteRepo.Create(ctx, note1)
			shared.AssertNoError(t, err, "create note1 should succeed")
			_, err = noteRepo.Create(ctx, note2)
			shared.AssertNoError(t, err, "create note2 should succeed")

			err = repo.RebuildFromNotes(ctx, noteRepo)
			shared.AssertNoError(t, err, "rebuild should succeed")

			docs, err := repo.List(ctx)
			shared.AssertNoError(t, err, "list should succeed")
			shared.AssertEqual(t, 2, len(docs), "should have 2 documents from notes")
		})

		t.Run("returns error with cancelled context", func(t *testing.T) {
			db := CreateTestDB(t)
			repo := NewDocumentRepository(db)
			noteRepo := NewNoteRepository(db)
			canceledCtx := NewCanceledContext()

			err := repo.RebuildFromNotes(canceledCtx, noteRepo)
			AssertCancelledContext(t, err)
		})
	})

	t.Run("BuildIndex", func(t *testing.T) {
		t.Run("creates search index from documents", func(t *testing.T) {
			db := CreateTestDB(t)
			repo := NewDocumentRepository(db)

			doc1 := CreateSampleDocument()
			doc1.Title = "Go Programming"
			doc1.Body = "Learn Go language"
			doc2 := CreateSampleDocument()
			doc2.Title = "Python Guide"
			doc2.Body = "Python tutorial"

			_, err := repo.Create(ctx, doc1)
			shared.AssertNoError(t, err, "create doc1 should succeed")
			_, err = repo.Create(ctx, doc2)
			shared.AssertNoError(t, err, "create doc2 should succeed")

			idx, err := repo.BuildIndex(ctx)
			shared.AssertNoError(t, err, "build index should succeed")
			shared.AssertNotNil(t, idx, "index should not be nil")
			shared.AssertEqual(t, 2, idx.NumDocs, "index should contain 2 documents")
		})

		t.Run("handles empty document set", func(t *testing.T) {
			db := CreateTestDB(t)
			repo := NewDocumentRepository(db)

			idx, err := repo.BuildIndex(ctx)
			shared.AssertNoError(t, err, "build index should succeed with empty set")
			shared.AssertEqual(t, 0, idx.NumDocs, "index should be empty")
		})

		t.Run("returns error with cancelled context", func(t *testing.T) {
			canceledCtx := NewCanceledContext()
			_, err := repo.BuildIndex(canceledCtx)
			AssertCancelledContext(t, err)
		})
	})
}

func TestSearchEngine(t *testing.T) {
	db := CreateTestDB(t)
	docRepo := NewDocumentRepository(db)
	ctx := context.Background()

	doc1 := CreateSampleDocument()
	doc1.Title = "Go Programming"
	doc1.Body = "Learn Go programming language with examples"
	doc2 := CreateSampleDocument()
	doc2.Title = "Python Tutorial"
	doc2.Body = "Python is a versatile programming language"
	doc3 := CreateSampleDocument()
	doc3.Title = "Go Advanced"
	doc3.Body = "Advanced Go concepts and patterns"

	_, err := docRepo.Create(ctx, doc1)
	shared.AssertNoError(t, err, "create doc1 should succeed")
	_, err = docRepo.Create(ctx, doc2)
	shared.AssertNoError(t, err, "create doc2 should succeed")
	_, err = docRepo.Create(ctx, doc3)
	shared.AssertNoError(t, err, "create doc3 should succeed")

	t.Run("Rebuild", func(t *testing.T) {
		t.Run("builds search index", func(t *testing.T) {
			engine := NewSearchEngine(docRepo)
			err := engine.Rebuild(ctx)
			shared.AssertNoError(t, err, "rebuild should succeed")
			shared.AssertNotNil(t, engine.index, "index should be set after rebuild")
		})

		t.Run("returns error with cancelled context", func(t *testing.T) {
			engine := NewSearchEngine(docRepo)
			canceledCtx := NewCanceledContext()
			err := engine.Rebuild(canceledCtx)
			AssertCancelledContext(t, err)
		})
	})

	t.Run("Search", func(t *testing.T) {
		t.Run("returns error when index not initialized", func(t *testing.T) {
			engine := NewSearchEngine(docRepo)
			_, err := engine.Search(ctx, "go", 10)
			shared.AssertError(t, err, "should error when index not initialized")
			shared.AssertContains(t, err.Error(), "not initialized", "error should mention not initialized")
		})

		t.Run("finds matching documents", func(t *testing.T) {
			engine := NewSearchEngine(docRepo)
			err := engine.Rebuild(ctx)
			shared.AssertNoError(t, err, "rebuild should succeed")

			docs, err := engine.Search(ctx, "go", 10)
			shared.AssertNoError(t, err, "search should succeed")
			shared.AssertTrue(t, len(docs) >= 2, "should find at least 2 documents with 'go'")
		})

		t.Run("returns empty results for non-matching query", func(t *testing.T) {
			engine := NewSearchEngine(docRepo)
			err := engine.Rebuild(ctx)
			shared.AssertNoError(t, err, "rebuild should succeed")

			docs, err := engine.Search(ctx, "rust", 10)
			shared.AssertNoError(t, err, "search should succeed")
			shared.AssertEqual(t, 0, len(docs), "should return no results for non-matching query")
		})

		t.Run("respects limit parameter", func(t *testing.T) {
			engine := NewSearchEngine(docRepo)
			err := engine.Rebuild(ctx)
			shared.AssertNoError(t, err, "rebuild should succeed")

			docs, err := engine.Search(ctx, "programming", 1)
			shared.AssertNoError(t, err, "search should succeed")
			shared.AssertTrue(t, len(docs) <= 1, "should respect limit parameter")
		})

		t.Run("returns error with cancelled context", func(t *testing.T) {
			engine := NewSearchEngine(docRepo)
			err := engine.Rebuild(ctx)
			shared.AssertNoError(t, err, "rebuild should succeed")

			canceledCtx := NewCanceledContext()
			_, err = engine.Search(canceledCtx, "go", 10)
			AssertCancelledContext(t, err)
		})
	})

	t.Run("SearchWithScores", func(t *testing.T) {
		t.Run("returns results with scores", func(t *testing.T) {
			engine := NewSearchEngine(docRepo)
			err := engine.Rebuild(ctx)
			shared.AssertNoError(t, err, "rebuild should succeed")

			results, docs, err := engine.SearchWithScores(ctx, "go", 10)
			shared.AssertNoError(t, err, "search should succeed")
			shared.AssertEqual(t, len(results), len(docs), "results and docs should have same length")
			shared.AssertTrue(t, len(results) >= 2, "should find at least 2 results")

			for _, result := range results {
				shared.AssertTrue(t, result.Score > 0, "score should be positive")
			}
		})

		t.Run("returns error when index not initialized", func(t *testing.T) {
			engine := NewSearchEngine(docRepo)
			_, _, err := engine.SearchWithScores(ctx, "go", 10)
			shared.AssertError(t, err, "should error when index not initialized")
		})

		t.Run("returns error with cancelled context", func(t *testing.T) {
			engine := NewSearchEngine(docRepo)
			err := engine.Rebuild(ctx)
			shared.AssertNoError(t, err, "rebuild should succeed")

			canceledCtx := NewCanceledContext()
			_, _, err = engine.SearchWithScores(canceledCtx, "go", 10)
			AssertCancelledContext(t, err)
		})
	})
}
