package handlers

import (
	"context"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stormlightlabs/noteleaf/internal/models"
)

func setupBookTest(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", "noteleaf-book-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	oldConfigHome := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tempDir)

	cleanup := func() {
		os.Setenv("XDG_CONFIG_HOME", oldConfigHome)
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
				err := handler.SearchAndAddBook(ctx, args, false)
				if err == nil {
					t.Error("Expected error for empty args")
				}

				if !strings.Contains(err.Error(), "usage: book add") {
					t.Errorf("Expected usage error, got: %v", err)
				}
			})

			t.Run("handles empty search", func(t *testing.T) {
				args := []string{""}
				err := handler.SearchAndAddBook(ctx, args, false)
				if err != nil && !strings.Contains(err.Error(), "No books found") {
					t.Errorf("Expected no error or 'No books found', got: %v", err)
				}
			})

			t.Run("with options", func(t *testing.T) {
				ctx := context.Background()
				t.Run("fails with empty args", func(t *testing.T) {
					args := []string{}
					err := handler.SearchAndAddBook(ctx, args, false)
					if err == nil {
						t.Error("Expected error for empty args")
					}

					if !strings.Contains(err.Error(), "usage: book add") {
						t.Errorf("Expected usage error, got: %v", err)
					}
				})

				t.Run("handles search service errors", func(t *testing.T) {
					args := []string{"test", "book"}
					err := handler.SearchAndAddBook(ctx, args, false)
					if err == nil {
						t.Error("Expected error due to mocked service")
					}
					if strings.Contains(err.Error(), "usage:") {
						t.Error("Should not show usage error for valid args")
					}
				})
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
				err := handler.ListBooks(ctx, "queued")
				if err != nil {
					t.Errorf("ListBooks failed: %v", err)
				}
			})

			t.Run("filters by status - all", func(t *testing.T) {
				err := handler.ListBooks(ctx, "")
				if err != nil {
					t.Errorf("ListBooks with status all failed: %v", err)
				}
			})

			t.Run("filters by status - reading", func(t *testing.T) {
				err := handler.ListBooks(ctx, "reading")
				if err != nil {
					t.Errorf("ListBooks with status reading failed: %v", err)
				}
			})

			t.Run("filters by status - finished", func(t *testing.T) {
				err := handler.ListBooks(ctx, "finished")
				if err != nil {
					t.Errorf("ListBooks with status finished failed: %v", err)
				}
			})

			t.Run("filters by status - queued", func(t *testing.T) {
				err := handler.ListBooks(ctx, "queued")
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
					err := handler.ListBooks(ctx, status)
					if err != nil {
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
					err := handler.UpdateBookStatusByID(ctx, strconv.FormatInt(book.ID, 10), "reading")
					if err != nil {
						t.Errorf("UpdateBookStatusByID failed: %v", err)
					}

					updatedBook, err := handler.repos.Books.Get(ctx, book.ID)
					if err != nil {
						t.Fatalf("Failed to get updated book: %v", err)
					}

					if updatedBook.Status != "reading" {
						t.Errorf("Expected status 'reading', got '%s'", updatedBook.Status)
					}

					if updatedBook.Started == nil {
						t.Error("Expected started time to be set")
					}
				})

				t.Run("updates to finished status", func(t *testing.T) {
					err := handler.UpdateBookStatusByID(ctx, strconv.FormatInt(book.ID, 10), "finished")
					if err != nil {
						t.Errorf("UpdateBookStatusByID failed: %v", err)
					}

					updatedBook, err := handler.repos.Books.Get(ctx, book.ID)
					if err != nil {
						t.Fatalf("Failed to get updated book: %v", err)
					}

					if updatedBook.Status != "finished" {
						t.Errorf("Expected status 'finished', got '%s'", updatedBook.Status)
					}

					if updatedBook.Finished == nil {
						t.Error("Expected finished time to be set")
					}

					if updatedBook.Progress != 100 {
						t.Errorf("Expected progress 100, got %d", updatedBook.Progress)
					}
				})

				t.Run("fails with invalid book ID", func(t *testing.T) {
					err := handler.UpdateBookStatusByID(ctx, "invalid-id", "reading")
					if err == nil {
						t.Error("Expected error for invalid book ID")
					}

					if !strings.Contains(err.Error(), "invalid book ID") {
						t.Errorf("Expected invalid book ID error, got: %v", err)
					}
				})

				t.Run("fails with invalid status", func(t *testing.T) {
					err := handler.UpdateBookStatusByID(ctx, strconv.FormatInt(book.ID, 10), "invalid-status")
					if err == nil {
						t.Error("Expected error for invalid status")
					}

					if !strings.Contains(err.Error(), "invalid status") {
						t.Errorf("Expected invalid status error, got: %v", err)
					}
				})

				t.Run("fails with non-existent book ID", func(t *testing.T) {
					err := handler.UpdateBookStatusByID(ctx, "99999", "reading")
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
						err := handler.UpdateBookStatusByID(ctx, strconv.FormatInt(book.ID, 10), status)
						if err != nil {
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
					err := handler.UpdateBookProgressByID(ctx, strconv.FormatInt(book.ID, 10), 50)
					if err != nil {
						t.Errorf("UpdateBookProgressByID failed: %v", err)
					}

					updatedBook, err := handler.repos.Books.Get(ctx, book.ID)
					if err != nil {
						t.Fatalf("Failed to get updated book: %v", err)
					}

					if updatedBook.Progress != 50 {
						t.Errorf("Expected progress 50, got %d", updatedBook.Progress)
					}

					if updatedBook.Status != "reading" {
						t.Errorf("Expected status 'reading', got '%s'", updatedBook.Status)
					}

					if updatedBook.Started == nil {
						t.Error("Expected started time to be set")
					}
				})

				t.Run("auto-completes book at 100%", func(t *testing.T) {
					err := handler.UpdateBookProgressByID(ctx, strconv.FormatInt(book.ID, 10), 100)
					if err != nil {
						t.Errorf("UpdateBookProgressByID failed: %v", err)
					}

					updatedBook, err := handler.repos.Books.Get(ctx, book.ID)
					if err != nil {
						t.Fatalf("Failed to get updated book: %v", err)
					}

					if updatedBook.Progress != 100 {
						t.Errorf("Expected progress 100, got %d", updatedBook.Progress)
					}

					if updatedBook.Status != "finished" {
						t.Errorf("Expected status 'finished', got '%s'", updatedBook.Status)
					}

					if updatedBook.Finished == nil {
						t.Error("Expected finished time to be set")
					}
				})

				t.Run("resets to queued at 0%", func(t *testing.T) {
					book.Status = "reading"
					now := time.Now()
					book.Started = &now
					handler.repos.Books.Update(ctx, book)

					err := handler.UpdateBookProgressByID(ctx, strconv.FormatInt(book.ID, 10), 0)
					if err != nil {
						t.Errorf("UpdateBookProgressByID failed: %v", err)
					}

					updatedBook, err := handler.repos.Books.Get(ctx, book.ID)
					if err != nil {
						t.Fatalf("Failed to get updated book: %v", err)
					}

					if updatedBook.Progress != 0 {
						t.Errorf("Expected progress 0, got %d", updatedBook.Progress)
					}

					if updatedBook.Status != "queued" {
						t.Errorf("Expected status 'queued', got '%s'", updatedBook.Status)
					}

					if updatedBook.Started != nil {
						t.Error("Expected started time to be nil")
					}
				})

				t.Run("fails with invalid book ID", func(t *testing.T) {
					err := handler.UpdateBookProgressByID(ctx, "invalid-id", 50)
					if err == nil {
						t.Error("Expected error for invalid book ID")
					}

					if !strings.Contains(err.Error(), "invalid book ID") {
						t.Errorf("Expected invalid book ID error, got: %v", err)
					}
				})

				t.Run("fails with progress out of range", func(t *testing.T) {
					testCases := []int{-1, 101, 150}

					for _, progress := range testCases {
						err := handler.UpdateBookProgressByID(ctx, strconv.FormatInt(book.ID, 10), progress)
						if err == nil {
							t.Errorf("Expected error for progress %d", progress)
						}

						if !strings.Contains(err.Error(), "progress must be between 0 and 100") {
							t.Errorf("Expected range error for progress %d, got: %v", progress, err)
						}
					}
				})

				t.Run("fails with non-existent book ID", func(t *testing.T) {
					err := handler.UpdateBookProgressByID(ctx, "99999", 50)
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

			err = handler.UpdateBookProgressByID(ctx, strconv.FormatInt(book.ID, 10), 25)
			if err != nil {
				t.Errorf("Failed to update progress: %v", err)
			}

			updatedBook, err := handler.repos.Books.Get(ctx, book.ID)
			if err != nil {
				t.Fatalf("Failed to get updated book: %v", err)
			}

			if updatedBook.Status != "reading" {
				t.Errorf("Expected status 'reading', got '%s'", updatedBook.Status)
			}

			err = handler.UpdateBookProgressByID(ctx, strconv.FormatInt(book.ID, 10), 100)
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
				done <- handler.ListBooks(ctx, "")
			}()

			go func() {
				time.Sleep(time.Millisecond * 15)
				done <- handler.UpdateBookProgressByID(ctx, strconv.FormatInt(book.ID, 10), 50)
			}()

			go func() {
				time.Sleep(time.Millisecond * 20)
				done <- handler.UpdateBookStatusByID(ctx, strconv.FormatInt(book.ID, 10), "finished")
			}()

			for i := range 3 {
				if err := <-done; err != nil {
					t.Errorf("Concurrent operation %d failed: %v", i, err)
				}
			}
		})
	})
}
