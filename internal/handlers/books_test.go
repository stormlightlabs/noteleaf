package handlers

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stormlightlabs/noteleaf/internal/models"
	"github.com/stormlightlabs/noteleaf/internal/repo"
	"github.com/stormlightlabs/noteleaf/internal/services"
)

func setupBookTest(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", "noteleaf-book-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	oldNoteleafConfig := os.Getenv("NOTELEAF_CONFIG")
	oldNoteleafDataDir := os.Getenv("NOTELEAF_DATA_DIR")
	os.Setenv("NOTELEAF_CONFIG", filepath.Join(tempDir, ".noteleaf.conf.toml"))
	os.Setenv("NOTELEAF_DATA_DIR", tempDir)

	cleanup := func() {
		os.Setenv("NOTELEAF_CONFIG", oldNoteleafConfig)
		os.Setenv("NOTELEAF_DATA_DIR", oldNoteleafDataDir)
		os.RemoveAll(tempDir)
	}

	ctx := context.Background()
	err = Setup(ctx, []string{})
	if err != nil {
		cleanup()
		t.Fatalf("Failed to setup database: %v", err)
	}

	return tempDir, cleanup
}

func createTestBook(t *testing.T, handler *BookHandler, ctx context.Context) *models.Book {
	t.Helper()
	if handler == nil {
		t.Fatal("handler provided to createTestBook is nil")
	}
	book := &models.Book{
		Title:  "Test Book",
		Author: "Test Author",
		Status: "queued",
		Added:  time.Now(),
	}
	id, err := handler.repos.Books.Create(ctx, book)
	if err != nil {
		t.Fatalf("Failed to create test book: %v", err)
	}
	book.ID = id
	return book
}

func TestBookHandler(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		t.Run("creates handler successfully", func(t *testing.T) {
			_, cleanup := setupBookTest(t)
			defer cleanup()

			handler, err := NewBookHandler()
			if err != nil {
				t.Fatalf("NewBookHandler failed: %v", err)
			}
			if handler == nil {
				t.Fatal("Handler should not be nil")
			}
			defer handler.Close()

			if handler.db == nil {
				t.Error("Handler database should not be nil")
			}
			if handler.config == nil {
				t.Error("Handler config should not be nil")
			}
			if handler.repos == nil {
				t.Error("Handler repos should not be nil")
			}
			if handler.service == nil {
				t.Error("Handler service should not be nil")
			}
		})

		t.Run("handles database initialization error", func(t *testing.T) {
			originalXDG := os.Getenv("XDG_CONFIG_HOME")
			originalHome := os.Getenv("HOME")

			os.Unsetenv("XDG_CONFIG_HOME")
			os.Unsetenv("HOME")
			defer func() {
				os.Setenv("XDG_CONFIG_HOME", originalXDG)
				os.Setenv("HOME", originalHome)
			}()

			handler, err := NewBookHandler()
			if err == nil {
				if handler != nil {
					handler.Close()
				}
				t.Error("Expected error when database initialization fails")
			}
		})
	})

	t.Run("BookHandler instance methods", func(t *testing.T) {
		_, cleanup := setupBookTest(t)
		defer cleanup()

		handler, err := NewBookHandler()
		if err != nil {
			t.Fatalf("Failed to create handler: %v", err)
		}
		defer handler.Close()

		t.Run("Search & Add", func(t *testing.T) {
			ctx := context.Background()
			t.Run("fails with empty args", func(t *testing.T) {
				args := []string{}
				query := strings.Join(args, " ")
				err := handler.SearchAndAdd(ctx, query, false)
				if err == nil {
					t.Error("Expected error for empty args")
				}
			})

			t.Run("context cancellation during search", func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				query := strings.Join([]string{"test", "book"}, " ")
				if err := handler.SearchAndAdd(ctx, query, false); err == nil {
					t.Error("Expected error for cancelled context")
				}
			})

			t.Run("handles HTTP error responses", func(t *testing.T) {
				mockServer := HTTPErrorMockServer(500, "Internal Server Error")
				defer mockServer.Close()

				handler.service = services.NewBookService(mockServer.URL())

				err := handler.SearchAndAdd(ctx, strings.Join([]string{"test", "book"}, " "), false)
				if err == nil {
					t.Error("Expected error for HTTP 500")
				}

				if !strings.Contains(err.Error(), "search failed") {
					t.Errorf("Expected search failure error, got: %v", err)
				}
			})

			t.Run("handles malformed JSON response", func(t *testing.T) {
				mockServer := InvalidJSONMockServer()
				defer mockServer.Close()

				handler.service = services.NewBookService(mockServer.URL())

				err := handler.SearchAndAdd(ctx, strings.Join([]string{"test", "book"}, " "), false)
				if err == nil {
					t.Error("Expected error for malformed JSON")
				}

				if !strings.Contains(err.Error(), "search failed") {
					t.Errorf("Expected search failure error, got: %v", err)
				}
			})

			t.Run("handles empty search results", func(t *testing.T) {
				emptyResponse := services.OpenLibrarySearchResponse{
					NumFound: 0, Start: 0, Docs: []services.OpenLibrarySearchDoc{},
				}

				mockServer := JSONMockServer(emptyResponse)
				defer mockServer.Close()

				handler.service = services.NewBookService(mockServer.URL())

				args := []string{"nonexistent", "book"}
				query := strings.Join(args, " ")
				err := handler.SearchAndAdd(ctx, query, false)
				if err != nil {
					t.Errorf("Expected no error for empty results, got: %v", err)
				}
			})

			t.Run("handles network timeouts", func(t *testing.T) {
				mockServer := TimeoutMockServer(5 * time.Second)
				defer mockServer.Close()

				handler.service = services.NewBookService(mockServer.URL())

				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				defer cancel()

				if err := handler.SearchAndAdd(ctx, strings.Join([]string{"test", "book"}, " "), false); err == nil {
					t.Error("Expected error for timeout")
				}
			})

			t.Run("handles context cancellation", func(t *testing.T) {
				mockBooks := []MockBook{
					{Key: "/works/OL123456W", Title: "Test Book", Authors: []string{"Author"}, Year: 2020},
				}
				mockResponse := MockOpenLibraryResponse(mockBooks)
				mockServer := JSONMockServer(mockResponse)
				defer mockServer.Close()

				handler.service = services.NewBookService(mockServer.URL())

				ctx, cancel := context.WithCancel(context.Background())
				cancel()

				if err := handler.SearchAndAdd(ctx, strings.Join([]string{"test", "book"}, " "), false); err == nil {
					t.Error("Expected error for cancelled context")
				}
			})

			t.Run("handles interactive mode", func(t *testing.T) {
				_, cleanup := setupBookTest(t)
				defer cleanup()

				handler, err := NewBookHandler()
				if err != nil {
					t.Fatalf("Failed to create handler: %v", err)
				}
				defer handler.Close()

				ctx := context.Background()
				if _, err = handler.repos.Books.Create(ctx, &models.Book{
					Title: "Test Book 1", Author: "Test Author 1", Status: "queued",
				}); err != nil {
					t.Fatalf("Failed to create test book: %v", err)
				}

				if _, err = handler.repos.Books.Create(ctx, &models.Book{
					Title: "Test Book 2", Author: "Test Author 2", Status: "reading",
				}); err != nil {
					t.Fatalf("Failed to create test book: %v", err)
				}

				if err = TestBookInteractiveList(t, handler, ""); err != nil {
					t.Errorf("Interactive book list test failed: %v", err)
				}
			})

			t.Run("interactive mode path", func(t *testing.T) {
				_, cleanup := setupBookTest(t)
				defer cleanup()

				handler, err := NewBookHandler()
				if err != nil {
					t.Fatalf("Failed to create handler: %v", err)
				}
				defer handler.Close()

				ctx := context.Background()
				if _, err = handler.repos.Books.Create(ctx,
					&models.Book{
						Title: "Interactive Test Book", Author: "Interactive Author", Status: "finished",
					}); err != nil {
					t.Fatalf("Failed to create test book: %v", err)
				}

				if err = TestBookInteractiveList(t, handler, "completed"); err != nil {
					t.Errorf("Interactive book list test with status filter failed: %v", err)
				}
			})

			t.Run("successful search and add with user selection", func(t *testing.T) {
				mockBooks := []MockBook{
					{Key: "/works/OL123W", Title: "Test Book 1", Authors: []string{"Author 1"}, Year: 2020, Editions: 5, CoverID: 123},
					{Key: "/works/OL456W", Title: "Test Book 2", Authors: []string{"Author 2"}, Year: 2021, Editions: 3, CoverID: 456},
				}
				mockResponse := MockOpenLibraryResponse(mockBooks)
				mockServer := JSONMockServer(mockResponse)
				defer mockServer.Close()

				handler.service = services.NewBookService(mockServer.URL())
				handler.SetInputReader(MenuSelection(1))

				args := []string{"test", "search"}
				query := strings.Join(args, " ")
				err := handler.SearchAndAdd(ctx, query, false)
				if err != nil {
					t.Errorf("Expected successful search and add, got error: %v", err)
				}

				books, err := handler.repos.Books.List(ctx, repo.BookListOptions{})
				if err != nil {
					t.Fatalf("Failed to list books: %v", err)
				}
				if len(books) != 1 {
					t.Errorf("Expected 1 book in database, got %d", len(books))
				}
				if len(books) > 0 && books[0].Title != "Test Book 1" {
					t.Errorf("Expected book title 'Test Book 1', got '%s'", books[0].Title)
				}
			})

			t.Run("successful search with user cancellation", func(t *testing.T) {
				mockBooks := []MockBook{
					{Key: "/works/OL789W", Title: "Another Book", Authors: []string{"Another Author"}, Year: 2022},
				}
				mockResponse := MockOpenLibraryResponse(mockBooks)
				mockServer := JSONMockServer(mockResponse)
				defer mockServer.Close()

				handler.service = services.NewBookService(mockServer.URL())
				handler.SetInputReader(MenuCancel())

				args := []string{"another", "search"}
				query := strings.Join(args, " ")
				err := handler.SearchAndAdd(ctx, query, false)
				if err != nil {
					t.Errorf("Expected no error on cancellation, got: %v", err)
				}

				books, err := handler.repos.Books.List(ctx, repo.BookListOptions{})
				if err != nil {
					t.Fatalf("Failed to list books: %v", err)
				}
				expected := 1
				if len(books) != expected {
					t.Errorf("Expected %d books in database after cancellation, got %d", expected, len(books))
				}
			})

			t.Run("invalid user choice", func(t *testing.T) {
				mockBooks := []MockBook{
					{Key: "/works/OL999W", Title: "Choice Test Book", Authors: []string{"Choice Author"}, Year: 2023},
				}
				mockResponse := MockOpenLibraryResponse(mockBooks)
				mockServer := JSONMockServer(mockResponse)
				defer mockServer.Close()

				handler.service = services.NewBookService(mockServer.URL())
				handler.SetInputReader(MenuSelection(5))

				args := []string{"choice", "test"}
				query := strings.Join(args, " ")
				err := handler.SearchAndAdd(ctx, query, false)
				if err == nil {
					t.Error("Expected error for invalid choice")
				}
				if err != nil && !strings.Contains(err.Error(), "invalid choice") {
					t.Errorf("Expected 'invalid choice' error, got: %v", err)
				}
			})

		})

		t.Run("List", func(t *testing.T) {
			ctx := context.Background()
			_ = createTestBook(t, handler, ctx)

			book2 := &models.Book{
				Title:  "Reading Book",
				Author: "Reading Author",
				Status: "reading",
				Added:  time.Now(),
			}
			id2, err := handler.repos.Books.Create(ctx, book2)
			if err != nil {
				t.Fatalf("Failed to create book2: %v", err)
			}
			book2.ID = id2

			book3 := &models.Book{
				Title:  "Finished Book",
				Author: "Finished Author",
				Status: "finished",
				Added:  time.Now(),
			}
			id3, err := handler.repos.Books.Create(ctx, book3)
			if err != nil {
				t.Fatalf("Failed to create book3: %v", err)
			}
			book3.ID = id3

			t.Run("lists queued books by default", func(t *testing.T) {
				err := handler.List(ctx, "queued")
				if err != nil {
					t.Errorf("ListBooks failed: %v", err)
				}
			})

			t.Run("filters by status - all", func(t *testing.T) {
				err := handler.List(ctx, "")
				if err != nil {
					t.Errorf("ListBooks with status all failed: %v", err)
				}
			})

			t.Run("filters by status - reading", func(t *testing.T) {
				err := handler.List(ctx, "reading")
				if err != nil {
					t.Errorf("ListBooks with status reading failed: %v", err)
				}
			})

			t.Run("filters by status - finished", func(t *testing.T) {
				err := handler.List(ctx, "finished")
				if err != nil {
					t.Errorf("ListBooks with status finished failed: %v", err)
				}
			})

			t.Run("filters by status - queued", func(t *testing.T) {
				err := handler.List(ctx, "queued")
				if err != nil {
					t.Errorf("ListBooks with status queued failed: %v", err)
				}
			})

			t.Run("handles various flag formats", func(t *testing.T) {
				statusVariants := map[string]string{
					"--all": "", "-a": "",
					"--reading": "reading", "-r": "reading",
					"--finished": "finished", "-f": "finished",
					"--queued": "queued", "-q": "queued",
				}

				for flag, status := range statusVariants {
					if err := handler.List(ctx, status); err != nil {
						t.Errorf("ListBooks with flag %s (status %s) failed: %v", flag, status, err)
					}
				}
			})
		})

		t.Run("Update", func(t *testing.T) {
			t.Run("Update status", func(t *testing.T) {
				ctx := context.Background()
				book := createTestBook(t, handler, ctx)

				t.Run("updates book status successfully", func(t *testing.T) {
					if err := handler.UpdateStatus(ctx, strconv.FormatInt(book.ID, 10), "reading"); err != nil {
						t.Errorf("UpdateBookStatusByID failed: %v", err)
					}

					updated, err := handler.repos.Books.Get(ctx, book.ID)
					if err != nil {
						t.Fatalf("Failed to get updated book: %v", err)
					}

					if updated.Status != "reading" {
						t.Errorf("Expected status 'reading', got '%s'", updated.Status)
					}

					if updated.Started == nil {
						t.Error("Expected started time to be set")
					}
				})

				t.Run("updates to finished status", func(t *testing.T) {
					err := handler.UpdateStatus(ctx, strconv.FormatInt(book.ID, 10), "finished")
					if err != nil {
						t.Errorf("UpdateBookStatusByID failed: %v", err)
					}

					updated, err := handler.repos.Books.Get(ctx, book.ID)
					if err != nil {
						t.Fatalf("Failed to get updated book: %v", err)
					}

					if updated.Status != "finished" {
						t.Errorf("Expected status 'finished', got '%s'", updated.Status)
					}

					if updated.Finished == nil {
						t.Error("Expected finished time to be set")
					}

					if updated.Progress != 100 {
						t.Errorf("Expected progress 100, got %d", updated.Progress)
					}
				})

				t.Run("fails with invalid book ID", func(t *testing.T) {
					err := handler.UpdateStatus(ctx, "invalid-id", "reading")
					if err == nil {
						t.Error("Expected error for invalid book ID")
					}

					if !strings.Contains(err.Error(), "invalid book ID") {
						t.Errorf("Expected invalid book ID error, got: %v", err)
					}
				})

				t.Run("fails with invalid status", func(t *testing.T) {
					err := handler.UpdateStatus(ctx, strconv.FormatInt(book.ID, 10), "invalid-status")
					if err == nil {
						t.Error("Expected error for invalid status")
					}

					if !strings.Contains(err.Error(), "invalid status") {
						t.Errorf("Expected invalid status error, got: %v", err)
					}
				})

				t.Run("fails with non-existent book ID", func(t *testing.T) {
					err := handler.UpdateStatus(ctx, "99999", "reading")
					if err == nil {
						t.Error("Expected error for non-existent book ID")
					}

					if !strings.Contains(err.Error(), "failed to get book") {
						t.Errorf("Expected book not found error, got: %v", err)
					}
				})

				t.Run("validates all status options", func(t *testing.T) {
					validStatuses := []string{"queued", "reading", "finished", "removed"}

					for _, status := range validStatuses {
						if err := handler.UpdateStatus(ctx, strconv.FormatInt(book.ID, 10), status); err != nil {
							t.Errorf("UpdateBookStatusByID with status %s failed: %v", status, err)
						}
					}
				})
			})

			t.Run("progress", func(t *testing.T) {
				_, cleanup := setupBookTest(t)
				defer cleanup()

				ctx := context.Background()

				handler, err := NewBookHandler()
				if err != nil {
					t.Fatalf("Failed to create handler: %v", err)
				}
				defer handler.Close()

				book := createTestBook(t, handler, ctx)

				t.Run("updates progress successfully", func(t *testing.T) {
					err := handler.UpdateProgress(ctx, strconv.FormatInt(book.ID, 10), 50)
					if err != nil {
						t.Errorf("UpdateBookProgressByID failed: %v", err)
					}

					updated, err := handler.repos.Books.Get(ctx, book.ID)
					if err != nil {
						t.Fatalf("Failed to get updated book: %v", err)
					}

					if updated.Progress != 50 {
						t.Errorf("Expected progress 50, got %d", updated.Progress)
					}

					if updated.Status != "reading" {
						t.Errorf("Expected status 'reading', got '%s'", updated.Status)
					}

					if updated.Started == nil {
						t.Error("Expected started time to be set")
					}
				})

				t.Run("auto-completes book at 100%", func(t *testing.T) {
					err := handler.UpdateProgress(ctx, strconv.FormatInt(book.ID, 10), 100)
					if err != nil {
						t.Errorf("UpdateBookProgressByID failed: %v", err)
					}

					updated, err := handler.repos.Books.Get(ctx, book.ID)
					if err != nil {
						t.Fatalf("Failed to get updated book: %v", err)
					}

					if updated.Progress != 100 {
						t.Errorf("Expected progress 100, got %d", updated.Progress)
					}

					if updated.Status != "finished" {
						t.Errorf("Expected status 'finished', got '%s'", updated.Status)
					}

					if updated.Finished == nil {
						t.Error("Expected finished time to be set")
					}
				})

				t.Run("resets to queued at 0%", func(t *testing.T) {
					book.Status = "reading"
					now := time.Now()
					book.Started = &now
					handler.repos.Books.Update(ctx, book)

					if err := handler.UpdateProgress(ctx, strconv.FormatInt(book.ID, 10), 0); err != nil {
						t.Errorf("UpdateBookProgressByID failed: %v", err)
					}

					updated, err := handler.repos.Books.Get(ctx, book.ID)
					if err != nil {
						t.Fatalf("Failed to get updated book: %v", err)
					}

					if updated.Progress != 0 {
						t.Errorf("Expected progress 0, got %d", updated.Progress)
					}

					if updated.Status != "queued" {
						t.Errorf("Expected status 'queued', got '%s'", updated.Status)
					}

					if updated.Started != nil {
						t.Error("Expected started time to be nil")
					}
				})

				t.Run("fails with invalid book ID", func(t *testing.T) {
					err := handler.UpdateProgress(ctx, "invalid-id", 50)
					if err == nil {
						t.Error("Expected error for invalid book ID")
					}

					if !strings.Contains(err.Error(), "invalid book ID") {
						t.Errorf("Expected invalid book ID error, got: %v", err)
					}
				})

				t.Run("fails with progress out of range", func(t *testing.T) {
					tt := []int{-1, 101, 150}

					for _, progress := range tt {
						err := handler.UpdateProgress(ctx, strconv.FormatInt(book.ID, 10), progress)
						if err == nil {
							t.Errorf("Expected error for progress %d", progress)
						}

						if !strings.Contains(err.Error(), "progress must be between 0 and 100") {
							t.Errorf("Expected range error for progress %d, got: %v", progress, err)
						}
					}
				})

				t.Run("fails with non-existent book ID", func(t *testing.T) {
					err := handler.UpdateProgress(ctx, "99999", 50)
					if err == nil {
						t.Error("Expected error for non-existent book ID")
					}

					if !strings.Contains(err.Error(), "failed to get book") {
						t.Errorf("Expected book not found error, got: %v", err)
					}
				})
			})
		})
	})

	t.Run("Close", func(t *testing.T) {
		t.Run("closes handler resources", func(t *testing.T) {
			_, cleanup := setupBookTest(t)
			defer cleanup()

			handler, err := NewBookHandler()
			if err != nil {
				t.Fatalf("NewBookHandler failed: %v", err)
			}

			err = handler.Close()
			if err != nil {
				t.Errorf("Close failed: %v", err)
			}
		})

		t.Run("handles service close gracefully", func(t *testing.T) {
			_, cleanup := setupBookTest(t)
			defer cleanup()

			handler, err := NewBookHandler()
			if err != nil {
				t.Fatalf("NewBookHandler failed: %v", err)
			}

			if err = handler.Close(); err != nil {
				t.Errorf("Close should succeed: %v", err)
			}
		})
	})

	t.Run("Print", func(t *testing.T) {
		_, cleanup := setupBookTest(t)
		defer cleanup()

		handler, err := NewBookHandler()
		if err != nil {
			t.Fatalf("Failed to create handler: %v", err)
		}
		defer handler.Close()

		now := time.Now()
		book := &models.Book{
			ID:       1,
			Title:    "Test Book",
			Author:   "Test Author",
			Status:   "reading",
			Progress: 75,
			Rating:   4.5,
			Notes:    "This is a test note that is longer than 80 characters to test the truncation functionality in the print method",
			Added:    now,
		}

		t.Run("printBook doesn't panic", func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("printBook panicked: %v", r)
				}
			}()

			handler.printBook(book)
		})

		t.Run("handles book with minimal fields", func(t *testing.T) {
			minimalBook := &models.Book{
				ID:     2,
				Title:  "Minimal Book",
				Status: "queued",
				Added:  now,
			}

			defer func() {
				if r := recover(); r != nil {
					t.Errorf("printBook panicked with minimal book: %v", r)
				}
			}()

			handler.printBook(minimalBook)
		})
	})

	t.Run("Error handling", func(t *testing.T) {
		t.Run("handler creation fails with invalid environment", func(t *testing.T) {
			originalXDG := os.Getenv("XDG_CONFIG_HOME")
			originalHome := os.Getenv("HOME")

			os.Unsetenv("XDG_CONFIG_HOME")
			os.Unsetenv("HOME")
			defer func() {
				os.Setenv("XDG_CONFIG_HOME", originalXDG)
				os.Setenv("HOME", originalHome)
			}()

			handler, err := NewBookHandler()
			if err == nil {
				if handler != nil {
					handler.Close()
				}
				t.Error("Expected error when environment is invalid")
			}
		})

	})

	t.Run("Integration", func(t *testing.T) {
		t.Run("full book lifecycle", func(t *testing.T) {
			_, cleanup := setupBookTest(t)
			defer cleanup()

			ctx := context.Background()

			handler, err := NewBookHandler()
			if err != nil {
				t.Fatalf("Failed to create handler: %v", err)
			}
			defer handler.Close()

			book := createTestBook(t, handler, ctx)

			if book.Status != "queued" {
				t.Errorf("Expected initial status 'queued', got '%s'", book.Status)
			}

			err = handler.UpdateProgress(ctx, strconv.FormatInt(book.ID, 10), 25)
			if err != nil {
				t.Errorf("Failed to update progress: %v", err)
			}

			updated, err := handler.repos.Books.Get(ctx, book.ID)
			if err != nil {
				t.Fatalf("Failed to get updated book: %v", err)
			}

			if updated.Status != "reading" {
				t.Errorf("Expected status 'reading', got '%s'", updated.Status)
			}

			err = handler.UpdateProgress(ctx, strconv.FormatInt(book.ID, 10), 100)
			if err != nil {
				t.Errorf("Failed to complete book: %v", err)
			}

			completedBook, err := handler.repos.Books.Get(ctx, book.ID)
			if err != nil {
				t.Fatalf("Failed to get completed book: %v", err)
			}

			if completedBook.Status != "finished" {
				t.Errorf("Expected status 'finished', got '%s'", completedBook.Status)
			}

			if completedBook.Progress != 100 {
				t.Errorf("Expected progress 100, got %d", completedBook.Progress)
			}

			if completedBook.Finished == nil {
				t.Error("Expected finished time to be set")
			}
		})

		t.Run("concurrent book operations", func(t *testing.T) {
			_, cleanup := setupBookTest(t)
			defer cleanup()

			ctx := context.Background()

			handler, err := NewBookHandler()
			if err != nil {
				t.Fatalf("Failed to create handler: %v", err)
			}
			defer handler.Close()

			book := createTestBook(t, handler, ctx)

			done := make(chan error, 3)

			go func() {
				time.Sleep(time.Millisecond * 10)
				done <- handler.List(ctx, "")
			}()

			go func() {
				time.Sleep(time.Millisecond * 15)
				done <- handler.UpdateProgress(ctx, strconv.FormatInt(book.ID, 10), 50)
			}()

			go func() {
				time.Sleep(time.Millisecond * 20)
				done <- handler.UpdateStatus(ctx, strconv.FormatInt(book.ID, 10), "finished")
			}()

			for i := range 3 {
				if err := <-done; err != nil {
					t.Errorf("Concurrent operation %d failed: %v", i, err)
				}
			}
		})
	})
}
