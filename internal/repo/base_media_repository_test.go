package repo

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stormlightlabs/noteleaf/internal/models"
	"github.com/stormlightlabs/noteleaf/internal/shared"
)

func TestBaseMediaRepository(t *testing.T) {
	ctx := context.Background()

	t.Run("Books", func(t *testing.T) {
		t.Run("Create and Get", func(t *testing.T) {
			db := CreateTestDB(t)
			repo := NewBookRepository(db)

			book := &models.Book{
				Title:  "Test Book",
				Author: "Test Author",
				Status: "queued",
			}

			id, err := repo.Create(ctx, book)
			shared.AssertNoError(t, err, "Failed to create book")
			shared.AssertNotEqual(t, int64(0), id, "Expected non-zero ID")

			retrieved, err := repo.Get(ctx, id)
			shared.AssertNoError(t, err, "Failed to get book")
			shared.AssertEqual(t, book.Title, retrieved.Title, "Title mismatch")
			shared.AssertEqual(t, book.Author, retrieved.Author, "Author mismatch")
			shared.AssertEqual(t, book.Status, retrieved.Status, "Status mismatch")
		})

		t.Run("Update", func(t *testing.T) {
			db := CreateTestDB(t)
			repo := NewBookRepository(db)

			book := &models.Book{
				Title:  "Original Title",
				Author: "Original Author",
				Status: "queued",
			}

			id, err := repo.Create(ctx, book)
			shared.AssertNoError(t, err, "Failed to create book")

			book.Title = "Updated Title"
			book.Author = "Updated Author"
			book.Status = "reading"

			err = repo.Update(ctx, book)
			shared.AssertNoError(t, err, "Failed to update book")

			retrieved, err := repo.Get(ctx, id)
			shared.AssertNoError(t, err, "Failed to get updated book")
			shared.AssertEqual(t, "Updated Title", retrieved.Title, "Title not updated")
			shared.AssertEqual(t, "Updated Author", retrieved.Author, "Author not updated")
			shared.AssertEqual(t, "reading", retrieved.Status, "Status not updated")
		})

		t.Run("Delete", func(t *testing.T) {
			db := CreateTestDB(t)
			repo := NewBookRepository(db)

			book := &models.Book{
				Title:  "To Delete",
				Status: "queued",
			}

			id, err := repo.Create(ctx, book)
			shared.AssertNoError(t, err, "Failed to create book")

			err = repo.Delete(ctx, id)
			shared.AssertNoError(t, err, "Failed to delete book")

			_, err = repo.Get(ctx, id)
			shared.AssertError(t, err, "Expected error when getting deleted book")
		})

		t.Run("Get non-existent", func(t *testing.T) {
			db := CreateTestDB(t)
			repo := NewBookRepository(db)

			_, err := repo.Get(ctx, 9999)
			shared.AssertError(t, err, "Expected error for non-existent book")
			shared.AssertContains(t, err.Error(), "not found", "Error should mention 'not found'")
		})

		t.Run("ListQuery with multiple books", func(t *testing.T) {
			db := CreateTestDB(t)
			repo := NewBookRepository(db)

			books := []*models.Book{
				{Title: "Book 1", Author: "Author A", Status: "queued"},
				{Title: "Book 2", Author: "Author B", Status: "reading"},
				{Title: "Book 3", Author: "Author A", Status: "finished"},
			}

			for _, book := range books {
				_, err := repo.Create(ctx, book)
				shared.AssertNoError(t, err, "Failed to create book")
			}

			allBooks, err := repo.List(ctx, BookListOptions{})
			shared.AssertNoError(t, err, "Failed to list books")
			if len(allBooks) != 3 {
				t.Errorf("Expected 3 books, got %d", len(allBooks))
			}
		})

		t.Run("CountQuery", func(t *testing.T) {
			db := CreateTestDB(t)
			repo := NewBookRepository(db)

			for i := 0; i < 5; i++ {
				book := &models.Book{
					Title:  "Book",
					Status: "queued",
				}
				_, err := repo.Create(ctx, book)
				shared.AssertNoError(t, err, "Failed to create book")
			}

			count, err := repo.Count(ctx, BookListOptions{})
			shared.AssertNoError(t, err, "Failed to count books")
			if count != 5 {
				t.Errorf("Expected count of 5, got %d", count)
			}
		})
	})

	t.Run("Movies", func(t *testing.T) {
		t.Run("Create and Get", func(t *testing.T) {
			db := CreateTestDB(t)
			repo := NewMovieRepository(db)

			movie := &models.Movie{
				Title:  "Test Movie",
				Year:   2023,
				Status: "queued",
			}

			id, err := repo.Create(ctx, movie)
			shared.AssertNoError(t, err, "Failed to create movie")
			shared.AssertNotEqual(t, int64(0), id, "Expected non-zero ID")

			retrieved, err := repo.Get(ctx, id)
			shared.AssertNoError(t, err, "Failed to get movie")
			shared.AssertEqual(t, movie.Title, retrieved.Title, "Title mismatch")
			shared.AssertEqual(t, movie.Year, retrieved.Year, "Year mismatch")
			shared.AssertEqual(t, movie.Status, retrieved.Status, "Status mismatch")
		})

		t.Run("Update", func(t *testing.T) {
			db := CreateTestDB(t)
			repo := NewMovieRepository(db)

			movie := &models.Movie{
				Title:  "Original Movie",
				Year:   2020,
				Status: "queued",
			}

			id, err := repo.Create(ctx, movie)
			shared.AssertNoError(t, err, "Failed to create movie")

			movie.Title = "Updated Movie"
			movie.Year = 2023
			movie.Status = "watched"

			err = repo.Update(ctx, movie)
			shared.AssertNoError(t, err, "Failed to update movie")

			retrieved, err := repo.Get(ctx, id)
			shared.AssertNoError(t, err, "Failed to get updated movie")
			shared.AssertEqual(t, "Updated Movie", retrieved.Title, "Title not updated")
			shared.AssertEqual(t, 2023, retrieved.Year, "Year not updated")
			shared.AssertEqual(t, "watched", retrieved.Status, "Status not updated")
		})

		t.Run("Delete", func(t *testing.T) {
			db := CreateTestDB(t)
			repo := NewMovieRepository(db)

			movie := &models.Movie{
				Title:  "To Delete",
				Status: "queued",
			}

			id, err := repo.Create(ctx, movie)
			shared.AssertNoError(t, err, "Failed to create movie")

			err = repo.Delete(ctx, id)
			shared.AssertNoError(t, err, "Failed to delete movie")

			_, err = repo.Get(ctx, id)
			shared.AssertError(t, err, "Expected error when getting deleted movie")
		})
	})

	t.Run("TV Shows", func(t *testing.T) {
		t.Run("Create and Get", func(t *testing.T) {
			db := CreateTestDB(t)
			repo := NewTVRepository(db)

			show := &models.TVShow{
				Title:   "Test Show",
				Season:  1,
				Episode: 1,
				Status:  "queued",
			}

			id, err := repo.Create(ctx, show)
			shared.AssertNoError(t, err, "Failed to create TV show")
			shared.AssertNotEqual(t, int64(0), id, "Expected non-zero ID")

			retrieved, err := repo.Get(ctx, id)
			shared.AssertNoError(t, err, "Failed to get TV show")
			shared.AssertEqual(t, show.Title, retrieved.Title, "Title mismatch")
			shared.AssertEqual(t, show.Season, retrieved.Season, "Season mismatch")
			shared.AssertEqual(t, show.Episode, retrieved.Episode, "Episode mismatch")
			shared.AssertEqual(t, show.Status, retrieved.Status, "Status mismatch")
		})

		t.Run("Update", func(t *testing.T) {
			db := CreateTestDB(t)
			repo := NewTVRepository(db)

			show := &models.TVShow{
				Title:   "Original Show",
				Season:  1,
				Episode: 1,
				Status:  "queued",
			}

			id, err := repo.Create(ctx, show)
			shared.AssertNoError(t, err, "Failed to create TV show")

			show.Title = "Updated Show"
			show.Season = 2
			show.Episode = 5
			show.Status = "watching"

			err = repo.Update(ctx, show)
			shared.AssertNoError(t, err, "Failed to update TV show")

			retrieved, err := repo.Get(ctx, id)
			shared.AssertNoError(t, err, "Failed to get updated TV show")
			shared.AssertEqual(t, "Updated Show", retrieved.Title, "Title not updated")
			shared.AssertEqual(t, 2, retrieved.Season, "Season not updated")
			shared.AssertEqual(t, 5, retrieved.Episode, "Episode not updated")
			shared.AssertEqual(t, "watching", retrieved.Status, "Status not updated")
		})

		t.Run("Delete", func(t *testing.T) {
			db := CreateTestDB(t)
			repo := NewTVRepository(db)

			show := &models.TVShow{
				Title:  "To Delete",
				Status: "queued",
			}

			id, err := repo.Create(ctx, show)
			shared.AssertNoError(t, err, "Failed to create TV show")

			err = repo.Delete(ctx, id)
			shared.AssertNoError(t, err, "Failed to delete TV show")

			_, err = repo.Get(ctx, id)
			shared.AssertError(t, err, "Expected error when getting deleted TV show")
		})
	})

	t.Run("Edge Cases", func(t *testing.T) {
		t.Run("buildPlaceholders", func(t *testing.T) {
			emptyResult := buildPlaceholders([]any{})
			if emptyResult != "" {
				t.Errorf("Expected empty string for empty values, got '%s'", emptyResult)
			}

			singleResult := buildPlaceholders([]any{1})
			if singleResult != "?" {
				t.Errorf("Expected '?' for single value, got '%s'", singleResult)
			}

			multipleResult := buildPlaceholders([]any{1, 2, 3})
			if multipleResult != "?,?,?" {
				t.Errorf("Expected '?,?,?' for three values, got '%s'", multipleResult)
			}

			largeResult := buildPlaceholders(make([]any, 10))
			expected := "?,?,?,?,?,?,?,?,?,?"
			if largeResult != expected {
				t.Errorf("Expected '%s' for ten values, got '%s'", expected, largeResult)
			}
		})
	})
}
