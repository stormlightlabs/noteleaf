package repo

import (
	"context"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stormlightlabs/noteleaf/internal/models"
)

func TestArticleRepository(t *testing.T) {
	t.Run("CRUD Operations", func(t *testing.T) {
		ctx := context.Background()

		t.Run("Create article", func(t *testing.T) {
			db := CreateTestDB(t)
			repo := NewArticleRepository(db)

			article := CreateSampleArticle()
			id, err := repo.Create(ctx, article)
			AssertNoError(t, err, "Failed to create article")
			AssertNotEqual(t, int64(0), id, "Expected non-zero ID")
			AssertEqual(t, id, article.ID, "Expected article ID to be set correctly")
			AssertFalse(t, article.Created.IsZero(), "Expected Created timestamp to be set")
			AssertFalse(t, article.Modified.IsZero(), "Expected Modified timestamp to be set")
		})

		t.Run("Get article", func(t *testing.T) {
			db := CreateTestDB(t)
			repo := NewArticleRepository(db)

			original := CreateSampleArticle()
			id, err := repo.Create(ctx, original)
			AssertNoError(t, err, "Failed to create article")

			retrieved, err := repo.Get(ctx, id)
			AssertNoError(t, err, "Failed to get article")
			AssertEqual(t, original.ID, retrieved.ID, "ID mismatch")
			AssertEqual(t, original.URL, retrieved.URL, "URL mismatch")
			AssertEqual(t, original.Title, retrieved.Title, "Title mismatch")
			AssertEqual(t, original.Author, retrieved.Author, "Author mismatch")
			AssertEqual(t, original.Date, retrieved.Date, "Date mismatch")
			AssertEqual(t, original.MarkdownPath, retrieved.MarkdownPath, "MarkdownPath mismatch")
			AssertEqual(t, original.HTMLPath, retrieved.HTMLPath, "HTMLPath mismatch")
		})

		t.Run("Update article", func(t *testing.T) {
			db := CreateTestDB(t)
			repo := NewArticleRepository(db)

			article := CreateSampleArticle()
			id, err := repo.Create(ctx, article)
			AssertNoError(t, err, "Failed to create article")

			originalModified := article.Modified
			article.Title = "Updated Title"
			article.Author = "Updated Author"
			article.Date = "2024-01-02"
			article.MarkdownPath = "/updated/path/article.md"
			article.HTMLPath = "/updated/path/article.html"

			err = repo.Update(ctx, article)
			AssertNoError(t, err, "Failed to update article")

			retrieved, err := repo.Get(ctx, id)
			AssertNoError(t, err, "Failed to get updated article")
			AssertEqual(t, "Updated Title", retrieved.Title, "Expected updated title")
			AssertEqual(t, "Updated Author", retrieved.Author, "Expected updated author")
			AssertEqual(t, "2024-01-02", retrieved.Date, "Expected updated date")
			AssertEqual(t, "/updated/path/article.md", retrieved.MarkdownPath, "Expected updated markdown path")
			AssertEqual(t, "/updated/path/article.html", retrieved.HTMLPath, "Expected updated HTML path")
			AssertTrue(t, retrieved.Modified.After(originalModified), "Expected Modified timestamp to be updated")
		})

		t.Run("Delete article", func(t *testing.T) {
			db := CreateTestDB(t)
			repo := NewArticleRepository(db)

			article := CreateSampleArticle()
			id, err := repo.Create(ctx, article)
			AssertNoError(t, err, "Failed to create article")

			err = repo.Delete(ctx, id)
			AssertNoError(t, err, "Failed to delete article")

			_, err = repo.Get(ctx, id)
			AssertError(t, err, "Expected error when getting deleted article")
		})
	})

	t.Run("Validation", func(t *testing.T) {
		db := CreateTestDB(t)
		repo := NewArticleRepository(db)
		ctx := context.Background()

		t.Run("Fails with missing title", func(t *testing.T) {
			article := CreateSampleArticle()
			article.Title = ""
			_, err := repo.Create(ctx, article)
			AssertError(t, err, "Expected error when creating article with empty title")
		})

		t.Run("Fails with missing URL", func(t *testing.T) {
			article := CreateSampleArticle()
			article.URL = ""
			_, err := repo.Create(ctx, article)
			AssertError(t, err, "Expected error when creating article with empty URL")
		})

		t.Run("Fails with duplicate URL", func(t *testing.T) {
			article1 := CreateSampleArticle()
			_, err := repo.Create(ctx, article1)
			AssertNoError(t, err, "Failed to create first article")

			article2 := CreateSampleArticle()
			article2.URL = article1.URL
			_, err = repo.Create(ctx, article2)
			AssertError(t, err, "Expected error when creating article with duplicate URL")
		})

		t.Run("Fails with missing markdown path", func(t *testing.T) {
			article := CreateSampleArticle()
			article.MarkdownPath = ""
			_, err := repo.Create(ctx, article)
			AssertError(t, err, "Expected error when creating article with empty markdown path")
			AssertContains(t, err.Error(), "MarkdownPath", "Expected MarkdownPath validation error")
		})

		t.Run("Fails with missing HTML path", func(t *testing.T) {
			article := CreateSampleArticle()
			article.HTMLPath = ""
			_, err := repo.Create(ctx, article)
			AssertError(t, err, "Expected error when creating article with empty HTML path")
			AssertContains(t, err.Error(), "HTMLPath", "Expected HTMLPath validation error")
		})

		t.Run("Fails with invalid URL format", func(t *testing.T) {
			article := CreateSampleArticle()
			article.URL = "not-a-valid-url"
			_, err := repo.Create(ctx, article)
			AssertError(t, err, "Expected error when creating article with invalid URL format")
			AssertContains(t, err.Error(), "URL", "Expected URL format validation error")
		})

		t.Run("Fails with invalid date format", func(t *testing.T) {
			article := CreateSampleArticle()
			article.Date = "invalid-date"
			_, err := repo.Create(ctx, article)
			AssertError(t, err, "Expected error when creating article with invalid date format")
			AssertContains(t, err.Error(), "Date", "Expected date validation error")
		})

		t.Run("Fails with title too long", func(t *testing.T) {
			article := CreateSampleArticle()
			article.Title = strings.Repeat("a", 501)
			_, err := repo.Create(ctx, article)
			AssertError(t, err, "Expected error when creating article with title too long")
			AssertContains(t, err.Error(), "Title", "Expected title length validation error")
		})

		t.Run("Fails with author too long", func(t *testing.T) {
			article := CreateSampleArticle()
			article.Author = strings.Repeat("a", 201)
			_, err := repo.Create(ctx, article)
			AssertError(t, err, "Expected error when creating article with author too long")
			AssertContains(t, err.Error(), "Author", "Expected author length validation error")
		})

		t.Run("Validates timestamps", func(t *testing.T) {
			article := CreateSampleArticle()
			now := time.Now()
			article.Modified = now
			article.Created = now.Add(time.Hour)
			err := repo.Validate(article)
			AssertError(t, err, "Expected error when created is after modified")
			AssertContains(t, err.Error(), "Created", "Expected timestamp validation error")
		})

		t.Run("Succeeds when created equals modified", func(t *testing.T) {
			article := CreateSampleArticle()
			now := time.Now()
			article.Created = now
			article.Modified = now
			err := repo.Validate(article)
			AssertNoError(t, err, "Expected no error when created equals modified")
		})

		t.Run("Succeeds when created is before modified", func(t *testing.T) {
			article := CreateSampleArticle()
			now := time.Now()
			article.Created = now
			article.Modified = now.Add(time.Hour)
			err := repo.Validate(article)
			AssertNoError(t, err, "Expected no error when created is before modified")
		})

		t.Run("Succeeds with valid optional fields", func(t *testing.T) {
			article := CreateSampleArticle()
			article.Date = "2024-01-01"
			article.Author = "Test Author"
			err := repo.Validate(article)
			AssertNoError(t, err, "Expected no error with valid optional fields")
		})

		t.Run("Succeeds with empty optional fields", func(t *testing.T) {
			article := CreateSampleArticle()
			article.Date = ""
			article.Author = ""
			err := repo.Validate(article)
			AssertNoError(t, err, "Expected no error with empty optional fields")
		})
	})

	t.Run("GetByURL", func(t *testing.T) {
		db := CreateTestDB(t)
		repo := NewArticleRepository(db)
		ctx := context.Background()

		t.Run("Successfully retrieves article by URL", func(t *testing.T) {
			original := CreateSampleArticle()
			_, err := repo.Create(ctx, original)
			AssertNoError(t, err, "Failed to create article")

			retrieved, err := repo.GetByURL(ctx, original.URL)
			AssertNoError(t, err, "Failed to get article by URL")
			AssertEqual(t, original.ID, retrieved.ID, "ID mismatch")
			AssertEqual(t, original.URL, retrieved.URL, "URL mismatch")
			AssertEqual(t, original.Title, retrieved.Title, "Title mismatch")
		})

		t.Run("Fails when URL not found", func(t *testing.T) {
			nonexistent := "https://example.com/nonexistent"
			_, err := repo.GetByURL(ctx, nonexistent)
			AssertError(t, err, "Expected error when getting article by non-existent URL")
			AssertContains(t, err.Error(), "not found", "Expected 'not found' in error message")
		})
	})

	t.Run("List", func(t *testing.T) {
		db := CreateTestDB(t)
		repo := NewArticleRepository(db)
		ctx := context.Background()

		articles := []*models.Article{
			{
				URL:          "https://example.com/article1",
				Title:        "First Article",
				Author:       "John Doe",
				Date:         "2024-01-01",
				MarkdownPath: "/path/article1.md",
				HTMLPath:     "/path/article1.html",
			},
			{
				URL:          "https://example.com/article2",
				Title:        "Second Article",
				Author:       "Jane Smith",
				Date:         "2024-01-02",
				MarkdownPath: "/path/article2.md",
				HTMLPath:     "/path/article2.html",
			},
			{
				URL:          "https://different.com/article3",
				Title:        "Important Article",
				Author:       "John Doe",
				Date:         "2024-01-03",
				MarkdownPath: "/path/article3.md",
				HTMLPath:     "/path/article3.html",
			},
		}

		for _, article := range articles {
			_, err := repo.Create(ctx, article)
			AssertNoError(t, err, "Failed to create test article")
		}

		t.Run("List all articles", func(t *testing.T) {
			results, err := repo.List(ctx, nil)
			AssertNoError(t, err, "Failed to list all articles")
			AssertEqual(t, 3, len(results), "Expected 3 articles")
		})

		t.Run("Filter by title", func(t *testing.T) {
			opts := &ArticleListOptions{Title: "Important"}
			results, err := repo.List(ctx, opts)
			AssertNoError(t, err, "Failed to list articles by title")
			AssertEqual(t, 1, len(results), "Expected 1 article matching title")
			AssertEqual(t, "Important Article", results[0].Title, "Wrong article returned")
		})

		t.Run("Filter by author", func(t *testing.T) {
			opts := &ArticleListOptions{Author: "John Doe"}
			results, err := repo.List(ctx, opts)
			AssertNoError(t, err, "Failed to list articles by author")
			AssertEqual(t, 2, len(results), "Expected 2 articles by John Doe")
		})

		t.Run("Filter by URL", func(t *testing.T) {
			opts := &ArticleListOptions{URL: "different.com"}
			results, err := repo.List(ctx, opts)
			AssertNoError(t, err, "Failed to list articles by URL")
			AssertEqual(t, 1, len(results), "Expected 1 article from different.com")
		})

		t.Run("Filter by date range", func(t *testing.T) {
			opts := &ArticleListOptions{DateFrom: "2024-01-02", DateTo: "2024-01-03"}
			results, err := repo.List(ctx, opts)
			AssertNoError(t, err, "Failed to list articles by date range")
			AssertEqual(t, 2, len(results), "Expected 2 articles in date range")
		})

		t.Run("With limit", func(t *testing.T) {
			opts := &ArticleListOptions{Limit: 2}
			results, err := repo.List(ctx, opts)
			AssertNoError(t, err, "Failed to list articles with limit")
			AssertEqual(t, 2, len(results), "Expected 2 articles due to limit")
		})

		t.Run("With limit and offset", func(t *testing.T) {
			opts := &ArticleListOptions{Limit: 2, Offset: 1}
			results, err := repo.List(ctx, opts)
			AssertNoError(t, err, "Failed to list articles with limit and offset")
			AssertEqual(t, 2, len(results), "Expected 2 articles due to limit")
		})

		t.Run("Multiple filters", func(t *testing.T) {
			opts := &ArticleListOptions{Author: "John Doe", DateFrom: "2024-01-02"}
			results, err := repo.List(ctx, opts)
			AssertNoError(t, err, "Failed to list articles with multiple filters")
			AssertEqual(t, 1, len(results), "Expected 1 article matching all filters")
			AssertEqual(t, "Important Article", results[0].Title, "Wrong article returned")
		})

		t.Run("No results", func(t *testing.T) {
			opts := &ArticleListOptions{Title: "Nonexistent"}
			results, err := repo.List(ctx, opts)
			AssertNoError(t, err, "Failed to list articles")
			AssertEqual(t, 0, len(results), "Expected no articles")
		})
	})

	t.Run("Count", func(t *testing.T) {
		db := CreateTestDB(t)
		repo := NewArticleRepository(db)
		ctx := context.Background()

		articles := []*models.Article{
			CreateSampleArticle(),
			{
				URL:          "https://example.com/article2",
				Title:        "Second Article",
				Author:       "Jane Smith",
				Date:         "2024-01-02",
				MarkdownPath: "/path/article2.md",
				HTMLPath:     "/path/article2.html",
			},
		}

		for _, article := range articles {
			_, err := repo.Create(ctx, article)
			AssertNoError(t, err, "Failed to create test article")
		}

		t.Run("Count all articles", func(t *testing.T) {
			count, err := repo.Count(ctx, nil)
			AssertNoError(t, err, "Failed to count articles")
			AssertEqual(t, int64(2), count, "Expected 2 articles")
		})

		t.Run("Count with filter", func(t *testing.T) {
			opts := &ArticleListOptions{Author: "Test Author"}
			count, err := repo.Count(ctx, opts)
			AssertNoError(t, err, "Failed to count articles with filter")
			AssertEqual(t, int64(1), count, "Expected 1 article by Test Author")
		})

		t.Run("Count with no results", func(t *testing.T) {
			opts := &ArticleListOptions{Title: "Nonexistent"}
			count, err := repo.Count(ctx, opts)
			AssertNoError(t, err, "Failed to count articles")
			AssertEqual(t, int64(0), count, "Expected 0 articles")
		})
	})

	t.Run("Context Cancellation Error Paths", func(t *testing.T) {
		db := CreateTestDB(t)
		repo := NewArticleRepository(db)
		ctx := context.Background()

		article := CreateSampleArticle()
		id, err := repo.Create(ctx, article)
		AssertNoError(t, err, "Failed to create article")

		t.Run("Create with cancelled context", func(t *testing.T) {
			newArticle := CreateSampleArticle()
			_, err := repo.Create(NewCanceledContext(), newArticle)
			AssertCancelledContext(t, err)
		})

		t.Run("Get with cancelled context", func(t *testing.T) {
			_, err := repo.Get(NewCanceledContext(), id)
			AssertCancelledContext(t, err)
		})

		t.Run("GetByURL with cancelled context", func(t *testing.T) {
			_, err := repo.GetByURL(NewCanceledContext(), article.URL)
			AssertCancelledContext(t, err)
		})

		t.Run("Update with cancelled context", func(t *testing.T) {
			article.Title = "Updated"
			err := repo.Update(NewCanceledContext(), article)
			AssertCancelledContext(t, err)
		})

		t.Run("Delete with cancelled context", func(t *testing.T) {
			err := repo.Delete(NewCanceledContext(), id)
			AssertCancelledContext(t, err)
		})

		t.Run("List with cancelled context", func(t *testing.T) {
			_, err := repo.List(NewCanceledContext(), nil)
			AssertCancelledContext(t, err)
		})

		t.Run("Count with cancelled context", func(t *testing.T) {
			_, err := repo.Count(NewCanceledContext(), nil)
			AssertCancelledContext(t, err)
		})
	})

	t.Run("Edge Cases", func(t *testing.T) {
		db := CreateTestDB(t)
		repo := NewArticleRepository(db)
		ctx := context.Background()

		t.Run("Get non-existent article", func(t *testing.T) {
			_, err := repo.Get(ctx, 99999)
			AssertError(t, err, "Expected error for non-existent article")
			AssertContains(t, err.Error(), "not found", "Expected 'not found' in error message")
		})

		t.Run("Update non-existent article", func(t *testing.T) {
			article := CreateSampleArticle()
			article.ID = 99999
			err := repo.Update(ctx, article)
			AssertError(t, err, "Expected error when updating non-existent article")
			AssertContains(t, err.Error(), "not found", "Expected 'not found' in error message")
		})

		t.Run("Delete non-existent article", func(t *testing.T) {
			err := repo.Delete(ctx, 99999)
			AssertError(t, err, "Expected error when deleting non-existent article")
			AssertContains(t, err.Error(), "not found", "Expected 'not found' in error message")
		})

		t.Run("Update validation - remove required title", func(t *testing.T) {
			db := CreateTestDB(t)
			repo := NewArticleRepository(db)

			article := CreateSampleArticle()
			_, err := repo.Create(ctx, article)
			AssertNoError(t, err, "Failed to create article")

			article.Title = ""
			err = repo.Update(ctx, article)
			AssertError(t, err, "Expected error when updating article with empty title")
		})

		t.Run("Update validation - invalid URL format", func(t *testing.T) {
			db := CreateTestDB(t)
			repo := NewArticleRepository(db)

			article := CreateSampleArticle()
			_, err := repo.Create(ctx, article)
			AssertNoError(t, err, "Failed to create article")

			article.URL = "not-a-valid-url"
			err = repo.Update(ctx, article)
			AssertError(t, err, "Expected error when updating article with invalid URL format")
			AssertContains(t, err.Error(), "URL", "Expected URL format validation error")
		})

		t.Run("Update validation - invalid date format", func(t *testing.T) {
			db := CreateTestDB(t)
			repo := NewArticleRepository(db)

			article := CreateSampleArticle()
			_, err := repo.Create(ctx, article)
			AssertNoError(t, err, "Failed to create article")

			article.Date = "invalid-date"
			err = repo.Update(ctx, article)
			AssertError(t, err, "Expected error when updating article with invalid date format")
			AssertContains(t, err.Error(), "Date", "Expected date validation error")
		})

		t.Run("Update validation - title too long", func(t *testing.T) {
			db := CreateTestDB(t)
			repo := NewArticleRepository(db)

			article := CreateSampleArticle()
			_, err := repo.Create(ctx, article)
			AssertNoError(t, err, "Failed to create article")

			article.Title = strings.Repeat("a", 501)
			err = repo.Update(ctx, article)
			AssertError(t, err, "Expected error when updating article with title too long")
			AssertContains(t, err.Error(), "Title", "Expected title length validation error")
		})

		t.Run("Update validation - author too long", func(t *testing.T) {
			db := CreateTestDB(t)
			repo := NewArticleRepository(db)

			article := CreateSampleArticle()
			_, err := repo.Create(ctx, article)
			AssertNoError(t, err, "Failed to create article")

			article.Author = strings.Repeat("a", 201)
			err = repo.Update(ctx, article)
			AssertError(t, err, "Expected error when updating article with author too long")
			AssertContains(t, err.Error(), "Author", "Expected author length validation error")
		})

		t.Run("Update validation - remove markdown path", func(t *testing.T) {
			db := CreateTestDB(t)
			repo := NewArticleRepository(db)

			article := CreateSampleArticle()
			_, err := repo.Create(ctx, article)
			AssertNoError(t, err, "Failed to create article")

			article.MarkdownPath = ""
			err = repo.Update(ctx, article)
			AssertError(t, err, "Expected error when updating article with empty markdown path")
			AssertContains(t, err.Error(), "MarkdownPath", "Expected MarkdownPath validation error")
		})

		t.Run("Update validation - remove HTML path", func(t *testing.T) {
			db := CreateTestDB(t)
			repo := NewArticleRepository(db)

			article := CreateSampleArticle()
			_, err := repo.Create(ctx, article)
			AssertNoError(t, err, "Failed to create article")

			article.HTMLPath = ""
			err = repo.Update(ctx, article)
			AssertError(t, err, "Expected error when updating article with empty HTML path")
			AssertContains(t, err.Error(), "HTMLPath", "Expected HTMLPath validation error")
		})

		t.Run("List with no results", func(t *testing.T) {
			opts := &ArticleListOptions{Author: "NonExistentAuthor"}
			articles, err := repo.List(ctx, opts)
			AssertNoError(t, err, "Should not error when no articles found")
			AssertEqual(t, 0, len(articles), "Expected empty result set")
		})
	})
}
