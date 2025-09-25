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
	db := CreateTestDB(t)
	repo := NewArticleRepository(db)
	ctx := context.Background()
	articles := CreateFakeArticles(10)

	t.Run("CRUD Operations", func(t *testing.T) {
		t.Run("Create", func(t *testing.T) {
			t.Run("successfully creates an article", func(t *testing.T) {
				article := articles[0]
				id, err := repo.Create(ctx, article)
				AssertNoError(t, err, "Failed to create article")
				AssertNotEqual(t, int64(0), id, "Expected non-zero ID")
				AssertEqual(t, id, article.ID, "Expected article ID to be set correctly")
				AssertFalse(t, article.Created.IsZero(), "Expected Created timestamp to be set")
				AssertFalse(t, article.Modified.IsZero(), "Expected Modified timestamp to be set")
			})

			t.Run("Fails with missing title", func(t *testing.T) {
				article := articles[1]
				article.Title = ""
				_, err := repo.Create(ctx, article)
				AssertError(t, err, "Expected error when creating article with empty title")
			})

			t.Run("Fails with missing URL", func(t *testing.T) {
				article := articles[2]
				article.URL = ""
				_, err := repo.Create(ctx, article)
				AssertError(t, err, "Expected error when creating article with empty URL")
			})

			t.Run("Fails with duplicate URL", func(t *testing.T) {
				article1 := articles[0]
				article2 := articles[1]
				article2.URL = article1.URL
				_, err := repo.Create(ctx, article2)
				AssertError(t, err, "Expected error when creating article with duplicate URL")
			})

			t.Run("Fails with missing markdown path", func(t *testing.T) {
				article := articles[3]
				article.MarkdownPath = ""
				_, err := repo.Create(ctx, article)
				AssertError(t, err, "Expected error when creating article with empty markdown path")
				AssertContains(t, err.Error(), "MarkdownPath", "Expected MarkdownPath validation error")
			})

			t.Run("Fails with missing HTML path", func(t *testing.T) {
				article := articles[4]
				article.HTMLPath = ""
				_, err := repo.Create(ctx, article)

				AssertError(t, err, "Expected error when creating article with empty HTML path")
				AssertContains(t, err.Error(), "HTMLPath", "Expected HTMLPath validation error")
			})

			t.Run("Fails with invalid URL format", func(t *testing.T) {
				article := articles[5]
				article.URL = "not-a-valid-url"
				_, err := repo.Create(ctx, article)
				AssertError(t, err, "Expected error when creating article with invalid URL format")
				AssertContains(t, err.Error(), "URL", "Expected URL format validation error")
			})

			t.Run("Fails with invalid date format", func(t *testing.T) {
				article := articles[6]
				article.Date = "invalid-date"
				_, err := repo.Create(ctx, article)
				AssertError(t, err, "Expected error when creating article with invalid date format")
				AssertContains(t, err.Error(), "Date", "Expected date validation error")
			})

			t.Run("Fails with title too long", func(t *testing.T) {
				article := articles[7]
				article.Title = strings.Repeat("a", 501)
				_, err := repo.Create(ctx, article)
				AssertError(t, err, "Expected error when creating article with title too long")
				AssertContains(t, err.Error(), "Title", "Expected title length validation error")
			})

			t.Run("Fails with author too long", func(t *testing.T) {
				article := articles[8]
				article.Author = strings.Repeat("a", 201)
				_, err := repo.Create(ctx, article)
				AssertError(t, err, "Expected error when creating article with author too long")
				AssertContains(t, err.Error(), "Author", "Expected author length validation error")
			})
		})

		t.Run("Get", func(t *testing.T) {
			t.Run("successfully retrieves an article", func(t *testing.T) {
				original := CreateFakeArticle()
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

			t.Run("Fails when ID isn't found", func(t *testing.T) {
				nonExistentID := int64(99999)
				_, err := repo.Get(ctx, nonExistentID)
				AssertError(t, err, "Expected error when getting non-existent article")
				AssertContains(t, err.Error(), "not found", "Expected 'not found' in error message")
			})
		})

		t.Run("Update", func(t *testing.T) {
			t.Run("successfully updates an article", func(t *testing.T) {
				article := CreateFakeArticle()
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

			t.Run("Fails when ID isn't found", func(t *testing.T) {
				article := CreateFakeArticle()
				article.ID = 99999
				err := repo.Update(ctx, article)
				AssertError(t, err, "Expected error when updating non-existent article")
				AssertContains(t, err.Error(), "not found", "Expected 'not found' in error message")
			})

			t.Run("Fails when trying to remove required value", func(t *testing.T) {
				article := CreateFakeArticle()
				_, err := repo.Create(ctx, article)
				AssertNoError(t, err, "Failed to create article")

				article.Title = ""
				err = repo.Update(ctx, article)
				AssertError(t, err, "Expected error when updating article with empty title")
			})

			t.Run("Fails when setting invalid URL format", func(t *testing.T) {
				article := CreateFakeArticle()
				_, err := repo.Create(ctx, article)
				AssertNoError(t, err, "Failed to create article")

				article.URL = "not-a-valid-url"
				err = repo.Update(ctx, article)
				AssertError(t, err, "Expected error when updating article with invalid URL format")
				AssertContains(t, err.Error(), "URL", "Expected URL format validation error")
			})

			t.Run("Fails when setting invalid date format", func(t *testing.T) {
				article := CreateFakeArticle()
				_, err := repo.Create(ctx, article)
				AssertNoError(t, err, "Failed to create article")

				article.Date = "invalid-date"
				err = repo.Update(ctx, article)
				AssertError(t, err, "Expected error when updating article with invalid date format")
				AssertContains(t, err.Error(), "Date", "Expected date validation error")
			})

			t.Run("Fails when setting title too long", func(t *testing.T) {
				article := CreateFakeArticle()
				_, err := repo.Create(ctx, article)
				AssertNoError(t, err, "Failed to create article")

				article.Title = strings.Repeat("a", 501)
				err = repo.Update(ctx, article)
				AssertError(t, err, "Expected error when updating article with title too long")
				AssertContains(t, err.Error(), "Title", "Expected title length validation error")
			})

			t.Run("Fails when setting author too long", func(t *testing.T) {
				article := CreateFakeArticle()
				_, err := repo.Create(ctx, article)
				AssertNoError(t, err, "Failed to create article")

				article.Author = strings.Repeat("a", 201)
				err = repo.Update(ctx, article)
				AssertError(t, err, "Expected error when updating article with author too long")
				AssertContains(t, err.Error(), "Author", "Expected author length validation error")
			})

			t.Run("Fails when removing markdown path", func(t *testing.T) {
				article := CreateFakeArticle()
				_, err := repo.Create(ctx, article)
				AssertNoError(t, err, "Failed to create article")

				article.MarkdownPath = ""
				err = repo.Update(ctx, article)
				AssertError(t, err, "Expected error when updating article with empty markdown path")
				AssertContains(t, err.Error(), "MarkdownPath", "Expected MarkdownPath validation error")
			})

			t.Run("Fails when removing HTML path", func(t *testing.T) {
				article := CreateFakeArticle()
				_, err := repo.Create(ctx, article)
				AssertNoError(t, err, "Failed to create article")

				article.HTMLPath = ""
				err = repo.Update(ctx, article)
				AssertError(t, err, "Expected error when updating article with empty HTML path")
				AssertContains(t, err.Error(), "HTMLPath", "Expected HTMLPath validation error")
			})
		})

		t.Run("Delete", func(t *testing.T) {
			t.Run("successfully removes an article", func(t *testing.T) {
				article := CreateFakeArticle()
				id, err := repo.Create(ctx, article)
				AssertNoError(t, err, "Failed to create article")

				err = repo.Delete(ctx, id)
				AssertNoError(t, err, "Failed to delete article")

				_, err = repo.Get(ctx, id)
				AssertError(t, err, "Expected error when getting deleted article")
			})

			t.Run("Fails when ID isn't found", func(t *testing.T) {
				nonexistent := int64(99999)
				err := repo.Delete(ctx, nonexistent)
				AssertError(t, err, "Expected error when deleting non-existent article")
				AssertContains(t, err.Error(), "not found", "Expected 'not found' in error message")
			})
		})
	})

	t.Run("GetByURL", func(t *testing.T) {
		db := CreateTestDB(t)
		repo := NewArticleRepository(db)
		ctx := context.Background()

		t.Run("successfully retrieves an article by URL", func(t *testing.T) {
			original := CreateFakeArticle()
			_, err := repo.Create(ctx, original)
			AssertNoError(t, err, "Failed to create article")

			retrieved, err := repo.GetByURL(ctx, original.URL)
			AssertNoError(t, err, "Failed to get article by URL")
			AssertEqual(t, original.ID, retrieved.ID, "ID mismatch")
			AssertEqual(t, original.URL, retrieved.URL, "URL mismatch")
			AssertEqual(t, original.Title, retrieved.Title, "Title mismatch")
		})

		t.Run("Fails when URL isn't found", func(t *testing.T) {
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

		t.Run("All articles", func(t *testing.T) {
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

		articles := []*models.Article{CreateSampleArticle(), {
			URL:          "https://example.com/article2",
			Title:        "Second Article",
			Author:       "Jane Smith",
			Date:         "2024-01-02",
			MarkdownPath: "/path/article2.md",
			HTMLPath:     "/path/article2.html"},
		}

		for _, article := range articles {
			_, err := repo.Create(ctx, article)
			AssertNoError(t, err, "Failed to create test article")
		}

		t.Run("Count all", func(t *testing.T) {
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

		t.Run("Count with context cancellation", func(t *testing.T) {
			cancelCtx, cancel := context.WithCancel(ctx)
			cancel()

			_, err := repo.Count(cancelCtx, nil)
			if err == nil {
				t.Error("Expected error with cancelled context")
			}
		})
	})

	t.Run("Validate", func(t *testing.T) {
		repo := NewArticleRepository(CreateTestDB(t))

		t.Run("successfully validates a valid article", func(t *testing.T) {
			article := CreateSampleArticle()
			err := repo.Validate(article)
			AssertNoError(t, err, "Expected no validation errors for valid article")
		})

		t.Run("fails with missing required URL", func(t *testing.T) {
			article := CreateSampleArticle()
			article.URL = ""
			err := repo.Validate(article)
			AssertError(t, err, "Expected error for missing URL")
			AssertContains(t, err.Error(), "URL", "Expected URL validation error")
		})

		t.Run("fails with missing required title", func(t *testing.T) {
			article := CreateSampleArticle()
			article.Title = ""
			err := repo.Validate(article)
			AssertError(t, err, "Expected error for missing title")
			AssertContains(t, err.Error(), "Title", "Expected title validation error")
		})

		t.Run("fails with missing markdown path", func(t *testing.T) {
			article := CreateSampleArticle()
			article.MarkdownPath = ""
			err := repo.Validate(article)
			AssertError(t, err, "Expected error for missing markdown path")
			AssertContains(t, err.Error(), "MarkdownPath", "Expected markdown path validation error")
		})

		t.Run("fails with missing HTML path", func(t *testing.T) {
			article := CreateSampleArticle()
			article.HTMLPath = ""
			err := repo.Validate(article)
			AssertError(t, err, "Expected error for missing HTML path")
			AssertContains(t, err.Error(), "HTMLPath", "Expected HTML path validation error")
		})

		t.Run("fails with invalid URL format", func(t *testing.T) {
			article := CreateSampleArticle()
			article.URL = "not-a-valid-url"
			err := repo.Validate(article)
			AssertError(t, err, "Expected error for invalid URL format")
			AssertContains(t, err.Error(), "URL", "Expected URL format validation error")
		})

		t.Run("fails with invalid date format", func(t *testing.T) {
			article := CreateSampleArticle()
			article.Date = "invalid-date"
			err := repo.Validate(article)
			AssertError(t, err, "Expected error for invalid date format")
			AssertContains(t, err.Error(), "Date", "Expected date validation error")
		})

		t.Run("fails with title too long", func(t *testing.T) {
			article := CreateSampleArticle()
			article.Title = strings.Repeat("a", 501)
			err := repo.Validate(article)
			AssertError(t, err, "Expected error for title too long")
			AssertContains(t, err.Error(), "Title", "Expected title length validation error")
		})

		t.Run("fails with author too long", func(t *testing.T) {
			article := CreateSampleArticle()
			article.Author = strings.Repeat("a", 201)
			err := repo.Validate(article)
			AssertError(t, err, "Expected error for author too long")
			AssertContains(t, err.Error(), "Author", "Expected author length validation error")
		})

		t.Run("fails when created is after modified", func(t *testing.T) {
			article := CreateSampleArticle()
			now := time.Now()
			article.Modified = now
			article.Created = now.Add(time.Hour)
			err := repo.Validate(article)
			AssertError(t, err, "Expected error when created is after modified")
			AssertContains(t, err.Error(), "Created", "Expected timestamp validation error")
		})

		t.Run("succeeds when created equals modified", func(t *testing.T) {
			article := CreateSampleArticle()
			now := time.Now()
			article.Created = now
			article.Modified = now
			err := repo.Validate(article)
			AssertNoError(t, err, "Expected no error when created equals modified")
		})

		t.Run("succeeds when created is before modified", func(t *testing.T) {
			article := CreateSampleArticle()
			now := time.Now()
			article.Created = now
			article.Modified = now.Add(time.Hour)
			err := repo.Validate(article)
			AssertNoError(t, err, "Expected no error when created is before modified")
		})

		t.Run("succeeds with valid optional fields", func(t *testing.T) {
			article := CreateSampleArticle()
			article.Date = "2024-01-01"
			article.Author = "Test Author"
			err := repo.Validate(article)
			AssertNoError(t, err, "Expected no error with valid optional fields")
		})

		t.Run("succeeds with empty optional fields", func(t *testing.T) {
			article := CreateSampleArticle()
			article.Date = ""
			article.Author = ""
			err := repo.Validate(article)
			AssertNoError(t, err, "Expected no error with empty optional fields")
		})
	})
}
