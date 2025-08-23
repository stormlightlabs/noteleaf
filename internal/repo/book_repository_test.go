package repo

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stormlightlabs/noteleaf/internal/models"
)

func createBookTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create in-memory database: %v", err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	schema := `
		CREATE TABLE IF NOT EXISTS books (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			author TEXT,
			status TEXT DEFAULT 'queued',
			progress INTEGER DEFAULT 0,
			pages INTEGER,
			rating REAL,
			notes TEXT,
			added DATETIME DEFAULT CURRENT_TIMESTAMP,
			started DATETIME,
			finished DATETIME
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

func createSampleBook() *models.Book {
	return &models.Book{
		Title:    "Test Book",
		Author:   "Test Author",
		Status:   "queued",
		Progress: 25,
		Pages:    300,
		Rating:   4.5,
		Notes:    "Interesting read",
	}
}

func TestBookRepository(t *testing.T) {
	t.Run("CRUD Operations", func(t *testing.T) {
		db := createBookTestDB(t)
		repo := NewBookRepository(db)
		ctx := context.Background()

		t.Run("Create Book", func(t *testing.T) {
			book := createSampleBook()

			id, err := repo.Create(ctx, book)
			if err != nil {
				t.Errorf("Failed to create book: %v", err)
			}

			if id == 0 {
				t.Error("Expected non-zero ID")
			}

			if book.ID != id {
				t.Errorf("Expected book ID to be set to %d, got %d", id, book.ID)
			}

			if book.Added.IsZero() {
				t.Error("Expected Added timestamp to be set")
			}
		})

		t.Run("Get Book", func(t *testing.T) {
			original := createSampleBook()
			id, err := repo.Create(ctx, original)
			if err != nil {
				t.Fatalf("Failed to create book: %v", err)
			}

			retrieved, err := repo.Get(ctx, id)
			if err != nil {
				t.Errorf("Failed to get book: %v", err)
			}

			if retrieved.Title != original.Title {
				t.Errorf("Expected title %s, got %s", original.Title, retrieved.Title)
			}
			if retrieved.Author != original.Author {
				t.Errorf("Expected author %s, got %s", original.Author, retrieved.Author)
			}
			if retrieved.Status != original.Status {
				t.Errorf("Expected status %s, got %s", original.Status, retrieved.Status)
			}
			if retrieved.Progress != original.Progress {
				t.Errorf("Expected progress %d, got %d", original.Progress, retrieved.Progress)
			}
			if retrieved.Pages != original.Pages {
				t.Errorf("Expected pages %d, got %d", original.Pages, retrieved.Pages)
			}
			if retrieved.Rating != original.Rating {
				t.Errorf("Expected rating %f, got %f", original.Rating, retrieved.Rating)
			}
			if retrieved.Notes != original.Notes {
				t.Errorf("Expected notes %s, got %s", original.Notes, retrieved.Notes)
			}
		})

		t.Run("Update Book", func(t *testing.T) {
			book := createSampleBook()
			id, err := repo.Create(ctx, book)
			if err != nil {
				t.Fatalf("Failed to create book: %v", err)
			}

			book.Title = "Updated Book"
			book.Status = "reading"
			book.Progress = 50
			book.Rating = 5.0
			now := time.Now()
			book.Started = &now

			err = repo.Update(ctx, book)
			if err != nil {
				t.Errorf("Failed to update book: %v", err)
			}

			updated, err := repo.Get(ctx, id)
			if err != nil {
				t.Fatalf("Failed to get updated book: %v", err)
			}

			if updated.Title != "Updated Book" {
				t.Errorf("Expected updated title, got %s", updated.Title)
			}
			if updated.Status != "reading" {
				t.Errorf("Expected status reading, got %s", updated.Status)
			}
			if updated.Progress != 50 {
				t.Errorf("Expected progress 50, got %d", updated.Progress)
			}
			if updated.Rating != 5.0 {
				t.Errorf("Expected rating 5.0, got %f", updated.Rating)
			}
			if updated.Started == nil {
				t.Error("Expected started time to be set")
			}
		})

		t.Run("Delete Book", func(t *testing.T) {
			book := createSampleBook()
			id, err := repo.Create(ctx, book)
			if err != nil {
				t.Fatalf("Failed to create book: %v", err)
			}

			err = repo.Delete(ctx, id)
			if err != nil {
				t.Errorf("Failed to delete book: %v", err)
			}

			_, err = repo.Get(ctx, id)
			if err == nil {
				t.Error("Expected error when getting deleted book")
			}
		})
	})

	t.Run("List", func(t *testing.T) {
		db := createBookTestDB(t)
		repo := NewBookRepository(db)
		ctx := context.Background()

		books := []*models.Book{
			{Title: "Book 1", Author: "Author A", Status: "queued", Progress: 0, Rating: 4.0},
			{Title: "Book 2", Author: "Author A", Status: "reading", Progress: 50, Rating: 4.5},
			{Title: "Book 3", Author: "Author B", Status: "finished", Progress: 100, Rating: 5.0},
			{Title: "Book 4", Author: "Author C", Status: "queued", Progress: 0, Rating: 3.5},
		}

		for _, book := range books {
			_, err := repo.Create(ctx, book)
			if err != nil {
				t.Fatalf("Failed to create book: %v", err)
			}
		}

		t.Run("List All Books", func(t *testing.T) {
			results, err := repo.List(ctx, BookListOptions{})
			if err != nil {
				t.Errorf("Failed to list books: %v", err)
			}

			if len(results) != 4 {
				t.Errorf("Expected 4 books, got %d", len(results))
			}
		})

		t.Run("List Books with Status Filter", func(t *testing.T) {
			results, err := repo.List(ctx, BookListOptions{Status: "queued"})
			if err != nil {
				t.Errorf("Failed to list books: %v", err)
			}

			if len(results) != 2 {
				t.Errorf("Expected 2 queued books, got %d", len(results))
			}

			for _, book := range results {
				if book.Status != "queued" {
					t.Errorf("Expected queued status, got %s", book.Status)
				}
			}
		})

		t.Run("List Books by Author", func(t *testing.T) {
			results, err := repo.List(ctx, BookListOptions{Author: "Author A"})
			if err != nil {
				t.Errorf("Failed to list books: %v", err)
			}

			if len(results) != 2 {
				t.Errorf("Expected 2 books by Author A, got %d", len(results))
			}

			for _, book := range results {
				if book.Author != "Author A" {
					t.Errorf("Expected author 'Author A', got %s", book.Author)
				}
			}
		})

		t.Run("List Books with Progress Filter", func(t *testing.T) {
			results, err := repo.List(ctx, BookListOptions{MinProgress: 50})
			if err != nil {
				t.Errorf("Failed to list books: %v", err)
			}

			if len(results) != 2 {
				t.Errorf("Expected 2 books with progress >= 50, got %d", len(results))
			}

			for _, book := range results {
				if book.Progress < 50 {
					t.Errorf("Expected progress >= 50, got %d", book.Progress)
				}
			}
		})

		t.Run("List Books with Rating Filter", func(t *testing.T) {
			results, err := repo.List(ctx, BookListOptions{MinRating: 4.5})
			if err != nil {
				t.Errorf("Failed to list books: %v", err)
			}

			if len(results) != 2 {
				t.Errorf("Expected 2 books with rating >= 4.5, got %d", len(results))
			}

			for _, book := range results {
				if book.Rating < 4.5 {
					t.Errorf("Expected rating >= 4.5, got %f", book.Rating)
				}
			}
		})

		t.Run("List Books with Search", func(t *testing.T) {
			results, err := repo.List(ctx, BookListOptions{Search: "Book 1"})
			if err != nil {
				t.Errorf("Failed to list books: %v", err)
			}

			if len(results) != 1 {
				t.Errorf("Expected 1 book matching search, got %d", len(results))
			}

			if len(results) > 0 && results[0].Title != "Book 1" {
				t.Errorf("Expected 'Book 1', got %s", results[0].Title)
			}
		})

		t.Run("List Books with Limit", func(t *testing.T) {
			results, err := repo.List(ctx, BookListOptions{Limit: 2})
			if err != nil {
				t.Errorf("Failed to list books: %v", err)
			}

			if len(results) != 2 {
				t.Errorf("Expected 2 books due to limit, got %d", len(results))
			}
		})
	})

	t.Run("Special Methods", func(t *testing.T) {
		db := createBookTestDB(t)
		repo := NewBookRepository(db)
		ctx := context.Background()

		book1 := &models.Book{Title: "Queued Book", Author: "Author A", Status: "queued", Progress: 0}
		book2 := &models.Book{Title: "Reading Book", Author: "Author B", Status: "reading", Progress: 45}
		book3 := &models.Book{Title: "Finished Book", Author: "Author C", Status: "finished", Progress: 100}
		book4 := &models.Book{Title: "Another Book", Author: "Author A", Status: "queued", Progress: 0}

		var book1ID int64
		for _, book := range []*models.Book{book1, book2, book3, book4} {
			id, err := repo.Create(ctx, book)
			if err != nil {
				t.Fatalf("Failed to create book: %v", err)
			}
			if book == book1 {
				book1ID = id
			}
		}

		t.Run("GetQueued", func(t *testing.T) {
			results, err := repo.GetQueued(ctx)
			if err != nil {
				t.Errorf("Failed to get queued books: %v", err)
			}

			if len(results) != 2 {
				t.Errorf("Expected 2 queued books, got %d", len(results))
			}

			for _, book := range results {
				if book.Status != "queued" {
					t.Errorf("Expected queued status, got %s", book.Status)
				}
			}
		})

		t.Run("GetReading", func(t *testing.T) {
			results, err := repo.GetReading(ctx)
			if err != nil {
				t.Errorf("Failed to get reading books: %v", err)
			}

			if len(results) != 1 {
				t.Errorf("Expected 1 reading book, got %d", len(results))
			}

			if len(results) > 0 && results[0].Status != "reading" {
				t.Errorf("Expected reading status, got %s", results[0].Status)
			}
		})

		t.Run("GetFinished", func(t *testing.T) {
			results, err := repo.GetFinished(ctx)
			if err != nil {
				t.Errorf("Failed to get finished books: %v", err)
			}

			if len(results) != 1 {
				t.Errorf("Expected 1 finished book, got %d", len(results))
			}

			if len(results) > 0 && results[0].Status != "finished" {
				t.Errorf("Expected finished status, got %s", results[0].Status)
			}
		})

		t.Run("GetByAuthor", func(t *testing.T) {
			results, err := repo.GetByAuthor(ctx, "Author A")
			if err != nil {
				t.Errorf("Failed to get books by author: %v", err)
			}

			if len(results) != 2 {
				t.Errorf("Expected 2 books by Author A, got %d", len(results))
			}

			for _, book := range results {
				if book.Author != "Author A" {
					t.Errorf("Expected author 'Author A', got %s", book.Author)
				}
			}
		})

		t.Run("StartReading", func(t *testing.T) {
			err := repo.StartReading(ctx, book1ID)
			if err != nil {
				t.Errorf("Failed to start reading book: %v", err)
			}

			updated, err := repo.Get(ctx, book1ID)
			if err != nil {
				t.Fatalf("Failed to get updated book: %v", err)
			}

			if updated.Status != "reading" {
				t.Errorf("Expected status to be reading, got %s", updated.Status)
			}

			if updated.Started == nil {
				t.Error("Expected started timestamp to be set")
			}
		})

		t.Run("FinishReading", func(t *testing.T) {
			newBook := &models.Book{Title: "New Book", Status: "reading", Progress: 80}
			id, err := repo.Create(ctx, newBook)
			if err != nil {
				t.Fatalf("Failed to create new book: %v", err)
			}

			err = repo.FinishReading(ctx, id)
			if err != nil {
				t.Errorf("Failed to finish reading book: %v", err)
			}

			updated, err := repo.Get(ctx, id)
			if err != nil {
				t.Fatalf("Failed to get updated book: %v", err)
			}

			if updated.Status != "finished" {
				t.Errorf("Expected status to be finished, got %s", updated.Status)
			}

			if updated.Progress != 100 {
				t.Errorf("Expected progress to be 100, got %d", updated.Progress)
			}

			if updated.Finished == nil {
				t.Error("Expected finished timestamp to be set")
			}
		})

		t.Run("UpdateProgress", func(t *testing.T) {
			newBook := &models.Book{Title: "Progress Book", Status: "queued", Progress: 0}
			id, err := repo.Create(ctx, newBook)
			if err != nil {
				t.Fatalf("Failed to create new book: %v", err)
			}

			err = repo.UpdateProgress(ctx, id, 25)
			if err != nil {
				t.Errorf("Failed to update progress: %v", err)
			}

			updated, err := repo.Get(ctx, id)
			if err != nil {
				t.Fatalf("Failed to get updated book: %v", err)
			}

			if updated.Status != "reading" {
				t.Errorf("Expected status to be reading when progress > 0, got %s", updated.Status)
			}

			if updated.Progress != 25 {
				t.Errorf("Expected progress 25, got %d", updated.Progress)
			}

			if updated.Started == nil {
				t.Error("Expected started timestamp to be set when progress > 0")
			}

			err = repo.UpdateProgress(ctx, id, 100)
			if err != nil {
				t.Errorf("Failed to update progress to 100: %v", err)
			}

			updated, err = repo.Get(ctx, id)
			if err != nil {
				t.Fatalf("Failed to get updated book: %v", err)
			}

			if updated.Status != "finished" {
				t.Errorf("Expected status to be finished when progress = 100, got %s", updated.Status)
			}

			if updated.Progress != 100 {
				t.Errorf("Expected progress 100, got %d", updated.Progress)
			}

			if updated.Finished == nil {
				t.Error("Expected finished timestamp to be set when progress = 100")
			}
		})
	})

	t.Run("Count", func(t *testing.T) {
		db := createBookTestDB(t)
		repo := NewBookRepository(db)
		ctx := context.Background()

		books := []*models.Book{
			{Title: "Book 1", Status: "queued", Progress: 0, Rating: 4.0},
			{Title: "Book 2", Status: "reading", Progress: 50, Rating: 3.5},
			{Title: "Book 3", Status: "finished", Progress: 100, Rating: 5.0},
			{Title: "Book 4", Status: "queued", Progress: 0, Rating: 4.5},
		}

		for _, book := range books {
			_, err := repo.Create(ctx, book)
			if err != nil {
				t.Fatalf("Failed to create book: %v", err)
			}
		}

		t.Run("Count all books", func(t *testing.T) {
			count, err := repo.Count(ctx, BookListOptions{})
			if err != nil {
				t.Errorf("Failed to count books: %v", err)
			}

			if count != 4 {
				t.Errorf("Expected 4 books, got %d", count)
			}
		})

		t.Run("Count queued books", func(t *testing.T) {
			count, err := repo.Count(ctx, BookListOptions{Status: "queued"})
			if err != nil {
				t.Errorf("Failed to count queued books: %v", err)
			}

			if count != 2 {
				t.Errorf("Expected 2 queued books, got %d", count)
			}
		})

		t.Run("Count books by progress", func(t *testing.T) {
			count, err := repo.Count(ctx, BookListOptions{MinProgress: 50})
			if err != nil {
				t.Errorf("Failed to count books with progress >= 50: %v", err)
			}

			if count != 2 {
				t.Errorf("Expected 2 books with progress >= 50, got %d", count)
			}
		})

		t.Run("Count books by rating", func(t *testing.T) {
			count, err := repo.Count(ctx, BookListOptions{MinRating: 4.0})
			if err != nil {
				t.Errorf("Failed to count high-rated books: %v", err)
			}

			if count != 3 {
				t.Errorf("Expected 3 books with rating >= 4.0, got %d", count)
			}
		})
	})
}
