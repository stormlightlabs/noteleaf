package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stormlightlabs/noteleaf/internal/articles"
	"github.com/stormlightlabs/noteleaf/internal/models"
	"github.com/stormlightlabs/noteleaf/internal/repo"
)

func TestArticleHandler(t *testing.T) {
	t.Run("NewArticleHandler", func(t *testing.T) {
		t.Run("creates handler successfully", func(t *testing.T) {
			helper := NewArticleTestHelper(t)

			if helper.ArticleHandler == nil {
				t.Fatal("Handler should not be nil")
			}

			if helper.db == nil {
				t.Error("Handler database should not be nil")
			}
			if helper.config == nil {
				t.Error("Handler config should not be nil")
			}
			if helper.repos == nil {
				t.Error("Handler repos should not be nil")
			}
			if helper.parser == nil {
				t.Error("Handler parser should not be nil")
			}
		})

		t.Run("handles database initialization error", func(t *testing.T) {
			envHelper := NewEnvironmentTestHelper()
			defer envHelper.RestoreEnv()

			if runtime.GOOS == "windows" {
				envHelper.UnsetEnv("APPDATA")
			} else {
				envHelper.UnsetEnv("XDG_CONFIG_HOME")
				envHelper.UnsetEnv("HOME")
			}

			_, err := NewArticleHandler()
			Expect.AssertError(t, err, "failed to initialize database", "NewArticleHandler should fail when database initialization fails")
		})

	})

	t.Run("Add", func(t *testing.T) {
		t.Run("adds article successfully", func(t *testing.T) {
			helper := NewArticleTestHelper(t)
			ctx := context.Background()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`<html>
					<head><title>Test Article</title></head>
					<body>
						<h1 id="firstHeading">Test Article Title</h1>
						<div class="author">Test Author</div>
						<div class="date">2024-01-01</div>
						<div id="bodyContent">
							<p>This is test content for the article.</p>
						</div>
					</body>
				</html>`))
			}))
			defer server.Close()

			testRule := &articles.ParsingRule{
				Domain: "127.0.0.1",
				Title:  "//h1[@id='firstHeading']",
				Author: "//div[@class='author']",
				Date:   "//div[@class='date']",
				Body:   "//div[@id='bodyContent']",
			}
			helper.AddTestRule("127.0.0.1", testRule)

			err := helper.Add(ctx, server.URL+"/test-article")
			Expect.AssertNoError(t, err, "Add should succeed with valid URL")

			articles, err := helper.repos.Articles.List(ctx, &repo.ArticleListOptions{})
			if err != nil {
				t.Fatalf("Failed to list articles: %v", err)
			}

			if len(articles) != 1 {
				t.Errorf("Expected 1 article, got %d", len(articles))
			}

			article := articles[0]
			if article.Title != "Test Article Title" {
				t.Errorf("Expected title 'Test Article Title', got '%s'", article.Title)
			}
			if article.Author != "Test Author" {
				t.Errorf("Expected author 'Test Author', got '%s'", article.Author)
			}
		})

		t.Run("handles duplicate article", func(t *testing.T) {
			helper := NewArticleTestHelper(t)
			ctx := context.Background()

			duplicateURL := "https://example.com/duplicate"

			existingArticle := &models.Article{
				URL:          duplicateURL,
				Title:        "Existing Article",
				Author:       "Existing Author",
				Date:         "2024-01-01",
				MarkdownPath: "/path/to/existing.md",
				HTMLPath:     "/path/to/existing.html",
				Created:      time.Now(),
				Modified:     time.Now(),
			}

			_, err := helper.repos.Articles.Create(ctx, existingArticle)
			if err != nil {
				t.Fatalf("Failed to create existing article: %v", err)
			}

			err = helper.Add(ctx, duplicateURL)
			Expect.AssertNoError(t, err, "Add should succeed with duplicate URL and return existing")
		})

		t.Run("handles unsupported domain", func(t *testing.T) {
			helper := NewArticleTestHelper(t)
			ctx := context.Background()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("<html><head><title>Test</title></head><body><p>Content</p></body></html>"))
			}))
			defer server.Close()

			err := helper.Add(ctx, server.URL+"/unsupported")
			Expect.AssertError(t, err, "failed to parse article", "Add should fail with unsupported domain")
		})

		t.Run("handles HTTP error", func(t *testing.T) {
			helper := NewArticleTestHelper(t)
			ctx := context.Background()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			}))
			defer server.Close()

			err := helper.Add(ctx, server.URL+"/404")
			Expect.AssertError(t, err, "failed to parse article", "Add should fail with HTTP error")
		})

		t.Run("handles storage directory error", func(t *testing.T) {
			helper := NewArticleTestHelper(t)
			ctx := context.Background()

			envHelper := NewEnvironmentTestHelper()
			defer envHelper.RestoreEnv()

			if runtime.GOOS == "windows" {
				envHelper.UnsetEnv("USERPROFILE")
				envHelper.UnsetEnv("HOMEDRIVE")
				envHelper.UnsetEnv("HOMEPATH")
			} else {
				envHelper.UnsetEnv("HOME")
			}

			err := helper.Add(ctx, "https://example.com/test-article")
			Expect.AssertError(t, err, "failed to get article storage dir", "Add should fail when storage directory cannot be determined")
		})

		t.Run("handles database save error", func(t *testing.T) {
			helper := NewArticleTestHelper(t)
			ctx := context.Background()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`<html>
					<head><title>Test Article</title></head>
					<body>
						<h1 id="firstHeading">Test Article</h1>
						<div id="bodyContent">Test content</div>
					</body>
				</html>`))
			}))
			defer server.Close()

			testRule := &articles.ParsingRule{
				Domain: "127.0.0.1",
				Title:  "//h1[@id='firstHeading']",
				Body:   "//div[@id='bodyContent']",
			}
			helper.AddTestRule("127.0.0.1", testRule)

			helper.db.Exec("DROP TABLE articles")

			err := helper.Add(ctx, server.URL+"/test")
			Expect.AssertError(t, err, "failed to save article to database", "Add should fail when database save fails")
		})
	})

	t.Run("List", func(t *testing.T) {
		t.Run("lists all articles", func(t *testing.T) {
			helper := NewArticleTestHelper(t)
			ctx := context.Background()

			id1 := helper.CreateTestArticle(t, "https://example.com/article1", "First Article", "John Doe", "2024-01-01")
			id2 := helper.CreateTestArticle(t, "https://example.com/article2", "Second Article", "Jane Smith", "2024-01-02")

			err := helper.List(ctx, "", "", 0)
			Expect.AssertNoError(t, err, "List should succeed")

			Expect.AssertArticleExists(t, helper, id1)
			Expect.AssertArticleExists(t, helper, id2)
		})

		t.Run("lists with title filter", func(t *testing.T) {
			helper := NewArticleTestHelper(t)
			ctx := context.Background()

			helper.CreateTestArticle(t, "https://example.com/first", "First Article", "John", "2024-01-01")
			helper.CreateTestArticle(t, "https://example.com/second", "Second Article", "Jane", "2024-01-02")

			err := helper.List(ctx, "First", "", 0)
			Expect.AssertNoError(t, err, "List with title filter should succeed")
		})

		t.Run("lists with author filter", func(t *testing.T) {
			helper := NewArticleTestHelper(t)
			ctx := context.Background()

			helper.CreateTestArticle(t, "https://example.com/john1", "Article by John", "John Doe", "2024-01-01")
			helper.CreateTestArticle(t, "https://example.com/jane1", "Article by Jane", "Jane Smith", "2024-01-02")

			err := helper.List(ctx, "", "John", 0)
			Expect.AssertNoError(t, err, "List with author filter should succeed")
		})

		t.Run("lists with limit", func(t *testing.T) {
			helper := NewArticleTestHelper(t)
			ctx := context.Background()

			helper.CreateTestArticle(t, "https://example.com/1", "Article 1", "Author", "2024-01-01")
			helper.CreateTestArticle(t, "https://example.com/2", "Article 2", "Author", "2024-01-02")
			helper.CreateTestArticle(t, "https://example.com/3", "Article 3", "Author", "2024-01-03")

			err := helper.List(ctx, "", "", 2)
			Expect.AssertNoError(t, err, "List with limit should succeed")
		})

		t.Run("handles empty results", func(t *testing.T) {
			helper := NewArticleTestHelper(t)
			ctx := context.Background()

			err := helper.List(ctx, "nonexistent", "", 0)
			Expect.AssertNoError(t, err, "List with no matches should succeed")
		})

		t.Run("handles database error", func(t *testing.T) {
			helper := NewArticleTestHelper(t)
			ctx := context.Background()

			helper.db.Exec("DROP TABLE articles")

			err := helper.List(ctx, "", "", 0)
			Expect.AssertError(t, err, "failed to list articles", "List should fail when database is corrupted")
		})
	})

	t.Run("View", func(t *testing.T) {
		t.Run("views article successfully", func(t *testing.T) {
			helper := NewArticleTestHelper(t)
			ctx := context.Background()

			id := helper.CreateTestArticle(t, "https://example.com/test", "Test Article", "Test Author", "2024-01-01")

			err := helper.View(ctx, id)
			Expect.AssertNoError(t, err, "View should succeed with valid article ID")
		})

		t.Run("handles non-existent article", func(t *testing.T) {
			helper := NewArticleTestHelper(t)
			ctx := context.Background()

			err := helper.View(ctx, 99999)
			Expect.AssertError(t, err, "failed to get article", "View should fail with non-existent article ID")
		})

		t.Run("handles missing files gracefully", func(t *testing.T) {
			helper := NewArticleTestHelper(t)
			ctx := context.Background()

			article := &models.Article{
				URL:          "https://example.com/missing-files",
				Title:        "Missing Files Article",
				Author:       "Test Author",
				Date:         "2024-01-01",
				MarkdownPath: "/non/existent/path.md",
				HTMLPath:     "/non/existent/path.html",
				Created:      time.Now(),
				Modified:     time.Now(),
			}

			id, err := helper.repos.Articles.Create(ctx, article)
			if err != nil {
				t.Fatalf("Failed to create article with missing files: %v", err)
			}

			err = helper.View(ctx, id)
			Expect.AssertNoError(t, err, "View should succeed even when files are missing")
		})

		t.Run("handles database error", func(t *testing.T) {
			helper := NewArticleTestHelper(t)
			ctx := context.Background()

			helper.db.Exec("DROP TABLE articles")

			err := helper.View(ctx, 1)
			Expect.AssertError(t, err, "failed to get article", "View should fail when database is corrupted")
		})
	})

	t.Run("Remove", func(t *testing.T) {
		t.Run("removes article successfully", func(t *testing.T) {
			helper := NewArticleTestHelper(t)
			ctx := context.Background()

			id := helper.CreateTestArticle(t, "https://example.com/remove", "Remove Test", "Author", "2024-01-01")

			Expect.AssertArticleExists(t, helper, id)

			err := helper.Remove(ctx, id)
			Expect.AssertNoError(t, err, "Remove should succeed")

			Expect.AssertArticleNotExists(t, helper, id)
		})

		t.Run("handles non-existent article", func(t *testing.T) {
			helper := NewArticleTestHelper(t)
			ctx := context.Background()

			err := helper.Remove(ctx, 99999)
			Expect.AssertError(t, err, "failed to get article", "Remove should fail with non-existent article ID")
		})

		t.Run("handles missing files gracefully", func(t *testing.T) {
			helper := NewArticleTestHelper(t)
			ctx := context.Background()

			article := &models.Article{
				URL:          "https://example.com/missing-files",
				Title:        "Missing Files Article",
				Author:       "Test Author",
				Date:         "2024-01-01",
				MarkdownPath: "/non/existent/path.md",
				HTMLPath:     "/non/existent/path.html",
				Created:      time.Now(),
				Modified:     time.Now(),
			}

			id, err := helper.repos.Articles.Create(ctx, article)
			if err != nil {
				t.Fatalf("Failed to create article with missing files: %v", err)
			}

			err = helper.Remove(ctx, id)
			Expect.AssertNoError(t, err, "Remove should succeed even when files don't exist")
		})

		t.Run("handles database error", func(t *testing.T) {
			helper := NewArticleTestHelper(t)
			ctx := context.Background()

			id := helper.CreateTestArticle(t, "https://example.com/db-error", "DB Error Test", "Author", "2024-01-01")

			helper.db.Exec("DROP TABLE articles")

			err := helper.Remove(ctx, id)
			Expect.AssertError(t, err, "failed to get article", "Remove should fail when database is corrupted")
		})
	})

	t.Run("Help", func(t *testing.T) {
		t.Run("shows supported domains", func(t *testing.T) {
			helper := NewArticleTestHelper(t)

			err := helper.Help()
			Expect.AssertNoError(t, err, "Help should succeed")
		})

		t.Run("handles storage directory error", func(t *testing.T) {
			helper := NewArticleTestHelper(t)

			envHelper := NewEnvironmentTestHelper()
			defer envHelper.RestoreEnv()

			if runtime.GOOS == "windows" {
				envHelper.UnsetEnv("USERPROFILE")
				envHelper.UnsetEnv("HOMEDRIVE")
				envHelper.UnsetEnv("HOMEPATH")
			} else {
				envHelper.UnsetEnv("HOME")
			}

			err := helper.Help()
			Expect.AssertError(t, err, "failed to get storage directory", "Help should fail when storage directory cannot be determined")
		})
	})

	t.Run("Close", func(t *testing.T) {
		t.Run("closes successfully", func(t *testing.T) {
			helper := NewArticleTestHelper(t)

			err := helper.Close()
			Expect.AssertNoError(t, err, "Close should succeed")
		})

		t.Run("handles nil database gracefully", func(t *testing.T) {
			helper := NewArticleTestHelper(t)
			helper.db = nil

			err := helper.Close()
			Expect.AssertNoError(t, err, "Close should succeed with nil database")
		})
	})

	t.Run("getStorageDirectory", func(t *testing.T) {
		t.Run("returns storage directory successfully", func(t *testing.T) {
			helper := NewArticleTestHelper(t)

			dir, err := helper.getStorageDirectory()
			Expect.AssertNoError(t, err, "getStorageDirectory should succeed")

			if dir == "" {
				t.Error("Storage directory should not be empty")
			}

			if !strings.Contains(dir, "Documents/Leaf") {
				t.Errorf("Expected storage directory to contain 'Documents/Leaf', got: %s", dir)
			}
		})

		t.Run("handles user home directory error", func(t *testing.T) {
			helper := NewArticleTestHelper(t)

			envHelper := NewEnvironmentTestHelper()
			defer envHelper.RestoreEnv()

			if runtime.GOOS == "windows" {
				envHelper.UnsetEnv("USERPROFILE")
				envHelper.UnsetEnv("HOMEDRIVE")
				envHelper.UnsetEnv("HOMEPATH")
			} else {
				envHelper.UnsetEnv("HOME")
			}

			_, err := helper.getStorageDirectory()
			Expect.AssertError(t, err, "", "getStorageDirectory should fail when home directory cannot be determined")
		})
	})
}

func TestArticleHandlerIntegration(t *testing.T) {
	t.Run("end-to-end workflow", func(t *testing.T) {
		helper := NewArticleTestHelper(t)
		ctx := context.Background()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<html>
				<head><title>Integration Test Article</title></head>
				<body>
					<h1 id="firstHeading">Integration Test Article</h1>
					<div class="author">Integration Author</div>
					<div id="bodyContent">
						<p>Integration test content.</p>
					</div>
				</body>
			</html>`))
		}))
		defer server.Close()

		testRule := &articles.ParsingRule{
			Domain: "127.0.0.1",
			Title:  "//h1[@id='firstHeading']",
			Author: "//div[@class='author']",
			Body:   "//div[@id='bodyContent']",
		}
		helper.AddTestRule("127.0.0.1", testRule)

		err := helper.Add(ctx, server.URL+"/integration-test")
		Expect.AssertNoError(t, err, "Add should succeed in integration test")

		err = helper.List(ctx, "", "", 0)
		Expect.AssertNoError(t, err, "List should succeed in integration test")

		articles, err := helper.repos.Articles.List(ctx, &repo.ArticleListOptions{})
		if err != nil {
			t.Fatalf("Failed to get articles for integration test: %v", err)
		}

		if len(articles) == 0 {
			t.Fatal("Expected at least one article for integration test")
		}

		articleID := articles[0].ID

		err = helper.View(ctx, articleID)
		Expect.AssertNoError(t, err, "View should succeed in integration test")

		err = helper.Help()
		Expect.AssertNoError(t, err, "Help should succeed in integration test")

		err = helper.Remove(ctx, articleID)
		Expect.AssertNoError(t, err, "Remove should succeed in integration test")

		Expect.AssertArticleNotExists(t, helper, articleID)
	})
}
