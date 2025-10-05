package repo

import (
	"context"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stormlightlabs/noteleaf/internal/models"
)

func TestBookRepository(t *testing.T) {
	t.Run("CRUD Operations", func(t *testing.T) {
		db := CreateTestDB(t)
		repo := NewBookRepository(db)
		ctx := context.Background()

		t.Run("Create Book", func(t *testing.T) {
			book := CreateSampleBook()

			id, err := repo.Create(ctx, book)
			AssertNoError(t, err, "Failed to create book")
			AssertNotEqual(t, int64(0), id, "Expected non-zero ID")
			AssertEqual(t, id, book.ID, "Expected book ID to be set correctly")
			AssertFalse(t, book.Added.IsZero(), "Expected Added timestamp to be set")
		})

		t.Run("Get Book", func(t *testing.T) {
			original := CreateSampleBook()
			id, err := repo.Create(ctx, original)
			AssertNoError(t, err, "Failed to create book")

			retrieved, err := repo.Get(ctx, id)
			AssertNoError(t, err, "Failed to get book")

			AssertEqual(t, original.Title, retrieved.Title, "Title mismatch")
			AssertEqual(t, original.Author, retrieved.Author, "Author mismatch")
			AssertEqual(t, original.Status, retrieved.Status, "Status mismatch")
			AssertEqual(t, original.Progress, retrieved.Progress, "Progress mismatch")
			AssertEqual(t, original.Pages, retrieved.Pages, "Pages mismatch")
			AssertEqual(t, original.Rating, retrieved.Rating, "Rating mismatch")
			AssertEqual(t, original.Notes, retrieved.Notes, "Notes mismatch")
		})

		t.Run("Update Book", func(t *testing.T) {
			book := CreateSampleBook()
			id, err := repo.Create(ctx, book)
			AssertNoError(t, err, "Failed to create book")

			book.Title = "Updated Book"
			book.Status = "reading"
			book.Progress = 50
			book.Rating = 5.0
			now := time.Now()
			book.Started = &now

			err = repo.Update(ctx, book)
			AssertNoError(t, err, "Failed to update book")

			updated, err := repo.Get(ctx, id)
			AssertNoError(t, err, "Failed to get updated book")

			AssertEqual(t, "Updated Book", updated.Title, "Expected updated title")
			AssertEqual(t, "reading", updated.Status, "Expected reading status")
			AssertEqual(t, 50, updated.Progress, "Expected progress 50")
			AssertEqual(t, 5.0, updated.Rating, "Expected rating 5.0")
			AssertTrue(t, updated.Started != nil, "Expected started time to be set")
		})

		t.Run("Delete Book", func(t *testing.T) {
			book := CreateSampleBook()
			id, err := repo.Create(ctx, book)
			AssertNoError(t, err, "Failed to create book")

			err = repo.Delete(ctx, id)
			AssertNoError(t, err, "Failed to delete book")

			_, err = repo.Get(ctx, id)
			AssertError(t, err, "Expected error when getting deleted book")
		})
	})

	t.Run("List", func(t *testing.T) {
		db := CreateTestDB(t)
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
			AssertNoError(t, err, "Failed to create book")
		}

		t.Run("List All Books", func(t *testing.T) {
			results, err := repo.List(ctx, BookListOptions{})
			AssertNoError(t, err, "Failed to list books")
			AssertEqual(t, 4, len(results), "Expected 4 books")
		})

		t.Run("List Books with Status Filter", func(t *testing.T) {
			results, err := repo.List(ctx, BookListOptions{Status: "queued"})
			AssertNoError(t, err, "Failed to list books")
			AssertEqual(t, 2, len(results), "Expected 2 queued books")

			for _, book := range results {
				AssertEqual(t, "queued", book.Status, "Expected queued status")
			}
		})

		t.Run("List Books by Author", func(t *testing.T) {
			results, err := repo.List(ctx, BookListOptions{Author: "Author A"})
			AssertNoError(t, err, "Failed to list books")
			AssertEqual(t, 2, len(results), "Expected 2 books by Author A")

			for _, book := range results {
				AssertEqual(t, "Author A", book.Author, "Expected author 'Author A'")
			}
		})

		t.Run("List Books with Progress Filter", func(t *testing.T) {
			results, err := repo.List(ctx, BookListOptions{MinProgress: 50})
			AssertNoError(t, err, "Failed to list books")
			AssertEqual(t, 2, len(results), "Expected 2 books with progress >= 50")

			for _, book := range results {
				AssertTrue(t, book.Progress >= 50, "Expected progress >= 50")
			}
		})

		t.Run("List Books with Rating Filter", func(t *testing.T) {
			results, err := repo.List(ctx, BookListOptions{MinRating: 4.5})
			AssertNoError(t, err, "Failed to list books")
			AssertEqual(t, 2, len(results), "Expected 2 books with rating >= 4.5")

			for _, book := range results {
				AssertTrue(t, book.Rating >= 4.5, "Expected rating >= 4.5")
			}
		})

		t.Run("List Books with Search", func(t *testing.T) {
			results, err := repo.List(ctx, BookListOptions{Search: "Book 1"})
			AssertNoError(t, err, "Failed to list books")
			AssertEqual(t, 1, len(results), "Expected 1 book matching search")

			if len(results) > 0 {
				AssertEqual(t, "Book 1", results[0].Title, "Expected 'Book 1'")
			}
		})

		t.Run("List Books with Limit", func(t *testing.T) {
			results, err := repo.List(ctx, BookListOptions{Limit: 2})
			AssertNoError(t, err, "Failed to list books")
			AssertEqual(t, 2, len(results), "Expected 2 books due to limit")
		})
	})

	t.Run("Special Methods", func(t *testing.T) {
		db := CreateTestDB(t)
		repo := NewBookRepository(db)
		ctx := context.Background()

		book1 := &models.Book{Title: "Queued Book", Author: "Author A", Status: "queued", Progress: 0}
		book2 := &models.Book{Title: "Reading Book", Author: "Author B", Status: "reading", Progress: 45}
		book3 := &models.Book{Title: "Finished Book", Author: "Author C", Status: "finished", Progress: 100}
		book4 := &models.Book{Title: "Another Book", Author: "Author A", Status: "queued", Progress: 0}

		var book1ID int64
		for _, book := range []*models.Book{book1, book2, book3, book4} {
			id, err := repo.Create(ctx, book)
			AssertNoError(t, err, "Failed to create book")
			if book == book1 {
				book1ID = id
			}
		}

		t.Run("GetQueued", func(t *testing.T) {
			results, err := repo.GetQueued(ctx)
			AssertNoError(t, err, "Failed to get queued books")
			AssertEqual(t, 2, len(results), "Expected 2 queued books")

			for _, book := range results {
				AssertEqual(t, "queued", book.Status, "Expected queued status")
			}
		})

		t.Run("GetReading", func(t *testing.T) {
			results, err := repo.GetReading(ctx)
			AssertNoError(t, err, "Failed to get reading books")
			AssertEqual(t, 1, len(results), "Expected 1 reading book")

			if len(results) > 0 {
				AssertEqual(t, "reading", results[0].Status, "Expected reading status")
			}
		})

		t.Run("GetFinished", func(t *testing.T) {
			results, err := repo.GetFinished(ctx)
			AssertNoError(t, err, "Failed to get finished books")
			AssertEqual(t, 1, len(results), "Expected 1 finished book")

			if len(results) > 0 {
				AssertEqual(t, "finished", results[0].Status, "Expected finished status")
			}
		})

		t.Run("GetByAuthor", func(t *testing.T) {
			results, err := repo.GetByAuthor(ctx, "Author A")
			AssertNoError(t, err, "Failed to get books by author")
			AssertEqual(t, 2, len(results), "Expected 2 books by Author A")

			for _, book := range results {
				AssertEqual(t, "Author A", book.Author, "Expected author 'Author A'")
			}
		})

		t.Run("StartReading", func(t *testing.T) {
			err := repo.StartReading(ctx, book1ID)
			AssertNoError(t, err, "Failed to start reading book")

			updated, err := repo.Get(ctx, book1ID)
			AssertNoError(t, err, "Failed to get updated book")

			AssertEqual(t, "reading", updated.Status, "Expected status to be reading")
			AssertTrue(t, updated.Started != nil, "Expected started timestamp to be set")
		})

		t.Run("FinishReading", func(t *testing.T) {
			newBook := &models.Book{Title: "New Book", Status: "reading", Progress: 80}
			id, err := repo.Create(ctx, newBook)
			AssertNoError(t, err, "Failed to create new book")

			err = repo.FinishReading(ctx, id)
			AssertNoError(t, err, "Failed to finish reading book")

			updated, err := repo.Get(ctx, id)
			AssertNoError(t, err, "Failed to get updated book")

			AssertEqual(t, "finished", updated.Status, "Expected status to be finished")
			AssertEqual(t, 100, updated.Progress, "Expected progress to be 100")
			AssertTrue(t, updated.Finished != nil, "Expected finished timestamp to be set")
		})

		t.Run("UpdateProgress", func(t *testing.T) {
			newBook := &models.Book{Title: "Progress Book", Status: "queued", Progress: 0}
			id, err := repo.Create(ctx, newBook)
			AssertNoError(t, err, "Failed to create new book")

			err = repo.UpdateProgress(ctx, id, 25)
			AssertNoError(t, err, "Failed to update progress")

			updated, err := repo.Get(ctx, id)
			AssertNoError(t, err, "Failed to get updated book")

			AssertEqual(t, "reading", updated.Status, "Expected status to be reading when progress > 0")
			AssertEqual(t, 25, updated.Progress, "Expected progress 25")
			AssertTrue(t, updated.Started != nil, "Expected started timestamp to be set when progress > 0")

			err = repo.UpdateProgress(ctx, id, 100)
			AssertNoError(t, err, "Failed to update progress to 100")

			updated, err = repo.Get(ctx, id)
			AssertNoError(t, err, "Failed to get updated book")

			AssertEqual(t, "finished", updated.Status, "Expected status to be finished when progress = 100")
			AssertEqual(t, 100, updated.Progress, "Expected progress 100")
			AssertTrue(t, updated.Finished != nil, "Expected finished timestamp to be set when progress = 100")
		})
	})

	t.Run("Count", func(t *testing.T) {
		db := CreateTestDB(t)
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
			AssertNoError(t, err, "Failed to create book")
		}

		t.Run("Count all books", func(t *testing.T) {
			count, err := repo.Count(ctx, BookListOptions{})
			AssertNoError(t, err, "Failed to count books")
			AssertEqual(t, int64(4), count, "Expected 4 books")
		})

		t.Run("Count queued books", func(t *testing.T) {
			count, err := repo.Count(ctx, BookListOptions{Status: "queued"})
			AssertNoError(t, err, "Failed to count queued books")
			AssertEqual(t, int64(2), count, "Expected 2 queued books")
		})

		t.Run("Count books by progress", func(t *testing.T) {
			count, err := repo.Count(ctx, BookListOptions{MinProgress: 50})
			AssertNoError(t, err, "Failed to count books with progress >= 50")
			AssertEqual(t, int64(2), count, "Expected 2 books with progress >= 50")
		})

		t.Run("Count books by rating", func(t *testing.T) {
			count, err := repo.Count(ctx, BookListOptions{MinRating: 4.0})
			AssertNoError(t, err, "Failed to count high-rated books")
			AssertEqual(t, int64(3), count, "Expected 3 books with rating >= 4.0")
		})

		t.Run("Count with context cancellation", func(t *testing.T) {
			_, err := repo.Count(NewCanceledContext(), BookListOptions{})
			AssertError(t, err, "Expected error with cancelled context")
		})
	})

	t.Run("Context Cancellation Error Paths", func(t *testing.T) {
		db := CreateTestDB(t)
		repo := NewBookRepository(db)
		ctx := context.Background()

		book := NewBookBuilder().WithTitle("Test Book").WithAuthor("Test Author").Build()
		id, err := repo.Create(ctx, book)
		AssertNoError(t, err, "Failed to create book")

		t.Run("Create with cancelled context", func(t *testing.T) {
			newBook := NewBookBuilder().WithTitle("Cancelled").Build()
			_, err := repo.Create(NewCanceledContext(), newBook)
			AssertError(t, err, "Expected error with cancelled context")
		})

		t.Run("Get with cancelled context", func(t *testing.T) {
			_, err := repo.Get(NewCanceledContext(), id)
			AssertError(t, err, "Expected error with cancelled context")
		})

		t.Run("Update with cancelled context", func(t *testing.T) {
			book.Title = "Updated"
			err := repo.Update(NewCanceledContext(), book)
			AssertError(t, err, "Expected error with cancelled context")
		})

		t.Run("Delete with cancelled context", func(t *testing.T) {
			err := repo.Delete(NewCanceledContext(), id)
			AssertError(t, err, "Expected error with cancelled context")
		})

		t.Run("List with cancelled context", func(t *testing.T) {
			_, err := repo.List(NewCanceledContext(), BookListOptions{})
			AssertError(t, err, "Expected error with cancelled context")
		})

		t.Run("GetQueued with cancelled context", func(t *testing.T) {
			_, err := repo.GetQueued(NewCanceledContext())
			AssertError(t, err, "Expected error with cancelled context")
		})

		t.Run("GetReading with cancelled context", func(t *testing.T) {
			_, err := repo.GetReading(NewCanceledContext())
			AssertError(t, err, "Expected error with cancelled context")
		})

		t.Run("GetFinished with cancelled context", func(t *testing.T) {
			_, err := repo.GetFinished(NewCanceledContext())
			AssertError(t, err, "Expected error with cancelled context")
		})

		t.Run("GetByAuthor with cancelled context", func(t *testing.T) {
			_, err := repo.GetByAuthor(NewCanceledContext(), "Test Author")
			AssertError(t, err, "Expected error with cancelled context")
		})

		t.Run("StartReading with cancelled context", func(t *testing.T) {
			err := repo.StartReading(NewCanceledContext(), id)
			AssertError(t, err, "Expected error with cancelled context")
		})

		t.Run("FinishReading with cancelled context", func(t *testing.T) {
			err := repo.FinishReading(NewCanceledContext(), id)
			AssertError(t, err, "Expected error with cancelled context")
		})

		t.Run("UpdateProgress with cancelled context", func(t *testing.T) {
			err := repo.UpdateProgress(NewCanceledContext(), id, 50)
			AssertError(t, err, "Expected error with cancelled context")
		})
	})

	t.Run("Edge Cases", func(t *testing.T) {
		db := CreateTestDB(t)
		repo := NewBookRepository(db)
		ctx := context.Background()

		t.Run("Get non-existent book", func(t *testing.T) {
			_, err := repo.Get(ctx, 99999)
			AssertError(t, err, "Expected error for non-existent book")
		})

		t.Run("Update non-existent book succeeds with no rows affected", func(t *testing.T) {
			book := NewBookBuilder().WithTitle("Non-existent").Build()
			book.ID = 99999
			err := repo.Update(ctx, book)
			AssertNoError(t, err, "Update should not error when no rows affected")
		})

		t.Run("Delete non-existent book succeeds with no rows affected", func(t *testing.T) {
			err := repo.Delete(ctx, 99999)
			AssertNoError(t, err, "Delete should not error when no rows affected")
		})

		t.Run("StartReading non-existent book", func(t *testing.T) {
			err := repo.StartReading(ctx, 99999)
			AssertError(t, err, "Expected error for non-existent book")
		})

		t.Run("FinishReading non-existent book", func(t *testing.T) {
			err := repo.FinishReading(ctx, 99999)
			AssertError(t, err, "Expected error for non-existent book")
		})

		t.Run("UpdateProgress non-existent book", func(t *testing.T) {
			err := repo.UpdateProgress(ctx, 99999, 50)
			AssertError(t, err, "Expected error for non-existent book")
		})

		t.Run("GetByAuthor with no results", func(t *testing.T) {
			books, err := repo.GetByAuthor(ctx, "NonExistentAuthor")
			AssertNoError(t, err, "Should not error when no books found")
			AssertEqual(t, 0, len(books), "Expected empty result set")
		})
	})
}
