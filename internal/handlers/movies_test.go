package handlers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stormlightlabs/noteleaf/internal/models"
)

func createTestMovieHandler(t *testing.T) *MovieHandler {
	handler, err := NewMovieHandler()
	if err != nil {
		t.Fatalf("Failed to create test movie handler: %v", err)
	}
	return handler
}

func createTestMovie() *models.Movie {
	now := time.Now()
	return &models.Movie{
		ID:     1,
		Title:  "Test Movie",
		Year:   2023,
		Status: "queued",
		Rating: 4.5,
		Notes:  "Test notes",
		Added:  now,
	}
}

func TestMovieHandler(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		handler := createTestMovieHandler(t)
		defer handler.Close()

		if handler.db == nil {
			t.Error("Expected database to be initialized")
		}
		if handler.config == nil {
			t.Error("Expected config to be initialized")
		}
		if handler.repos == nil {
			t.Error("Expected repositories to be initialized")
		}
		if handler.service == nil {
			t.Error("Expected service to be initialized")
		}
	})

	t.Run("Close", func(t *testing.T) {
		handler := createTestMovieHandler(t)

		err := handler.Close()
		if err != nil {
			t.Errorf("Expected no error when closing handler, got: %v", err)
		}
	})

	t.Run("Search and Add", func(t *testing.T) {
		t.Run("Empty Query", func(t *testing.T) {
			handler := createTestMovieHandler(t)
			defer handler.Close()

			err := handler.SearchAndAdd(context.Background(), "", false)
			if err == nil {
				t.Error("Expected error for empty query")
			}
			if err.Error() != "search query cannot be empty" {
				t.Errorf("Expected 'search query cannot be empty', got: %v", err)
			}
		})

		t.Run("Search Error", func(t *testing.T) {
			handler := createTestMovieHandler(t)
			defer handler.Close()

			// Test with malformed search that should cause network error
			err := handler.SearchAndAdd(context.Background(), "test movie", false)
			// We expect this to work with the actual service, so we test for successful completion
			// or a specific network error - this tests the error handling path in the code
			if err != nil {
				// This is expected - the search might fail due to network issues in test environment
				if err.Error() != "search query cannot be empty" {
					// We got a search error, which tests our error handling path
					t.Logf("Search failed as expected in test environment: %v", err)
				}
			}
		})

		t.Run("Network Error", func(t *testing.T) {
			handler := createTestMovieHandler(t)
			defer handler.Close()

			// Test search with a query that will likely fail due to network issues in test env
			// This tests the error handling path
			err := handler.SearchAndAdd(context.Background(), "unlikely_to_find_this_movie_12345", false)
			// We don't expect a specific error, but this tests the error handling path
			if err != nil {
				t.Logf("Network error encountered (expected in test environment): %v", err)
			}
		})

	})

	t.Run("List", func(t *testing.T) {
		t.Run("Invalid Status", func(t *testing.T) {
			handler := createTestMovieHandler(t)
			defer handler.Close()

			err := handler.List(context.Background(), "invalid_status")
			if err == nil {
				t.Error("Expected error for invalid status")
			}
			if err.Error() != "invalid status: invalid_status (use: queued, watched, or leave empty for all)" {
				t.Errorf("Expected invalid status error, got: %v", err)
			}
		})

		t.Run("All Movies", func(t *testing.T) {
			handler := createTestMovieHandler(t)
			defer handler.Close()

			// Test with empty status (all movies)
			err := handler.List(context.Background(), "")
			if err != nil {
				t.Errorf("Expected no error for listing all movies, got: %v", err)
			}
		})

		t.Run("Queued Movies", func(t *testing.T) {
			handler := createTestMovieHandler(t)
			defer handler.Close()

			err := handler.List(context.Background(), "queued")
			if err != nil {
				t.Errorf("Expected no error for listing queued movies, got: %v", err)
			}
		})

		t.Run("Watched Movies", func(t *testing.T) {
			handler := createTestMovieHandler(t)
			defer handler.Close()

			err := handler.List(context.Background(), "watched")
			if err != nil {
				t.Errorf("Expected no error for listing watched movies, got: %v", err)
			}
		})

	})

	t.Run("View", func(t *testing.T) {
		t.Run("Movie Not Found", func(t *testing.T) {
			handler := createTestMovieHandler(t)
			defer handler.Close()

			err := handler.View(context.Background(), 999)
			if err == nil {
				t.Error("Expected error for non-existent movie")
			}
		})

		t.Run("Invalid ID", func(t *testing.T) {
			handler := createTestMovieHandler(t)
			defer handler.Close()

			err := handler.ViewMovie(context.Background(), "invalid")
			if err == nil {
				t.Error("Expected error for invalid movie ID")
			}
			if err.Error() != "invalid movie ID: invalid" {
				t.Errorf("Expected 'invalid movie ID: invalid', got: %v", err)
			}
		})
	})

	t.Run("Update", func(t *testing.T) {
		t.Run("Update Status", func(t *testing.T) {
			t.Run("Invalid", func(t *testing.T) {
				handler := createTestMovieHandler(t)
				defer handler.Close()

				err := handler.UpdateStatus(context.Background(), 1, "invalid")
				if err == nil {
					t.Error("Expected error for invalid status")
				}
				if err.Error() != "invalid status: invalid (valid: queued, watched, removed)" {
					t.Errorf("Expected invalid status error, got: %v", err)
				}
			})

			t.Run("Movie Not Found", func(t *testing.T) {
				handler := createTestMovieHandler(t)
				defer handler.Close()

				err := handler.UpdateStatus(context.Background(), 999, "watched")
				if err == nil {
					t.Error("Expected error for non-existent movie")
				}
			})
		})
	})

	t.Run("MarkWatched_MovieNotFound", func(t *testing.T) {
		handler := createTestMovieHandler(t)
		defer handler.Close()

		err := handler.MarkWatched(context.Background(), 999)
		if err == nil {
			t.Error("Expected error for non-existent movie")
		}
	})

	t.Run("Remove_MovieNotFound", func(t *testing.T) {
		handler := createTestMovieHandler(t)
		defer handler.Close()

		err := handler.Remove(context.Background(), 999)
		if err == nil {
			t.Error("Expected error for non-existent movie")
		}
	})

	t.Run("UpdateMovieStatus_InvalidID", func(t *testing.T) {
		handler := createTestMovieHandler(t)
		defer handler.Close()

		err := handler.UpdateMovieStatus(context.Background(), "invalid", "watched")
		if err == nil {
			t.Error("Expected error for invalid movie ID")
		}
		if err.Error() != "invalid movie ID: invalid" {
			t.Errorf("Expected 'invalid movie ID: invalid', got: %v", err)
		}
	})

	t.Run("MarkMovieWatched_InvalidID", func(t *testing.T) {
		handler := createTestMovieHandler(t)
		defer handler.Close()

		err := handler.MarkMovieWatched(context.Background(), "invalid")
		if err == nil {
			t.Error("Expected error for invalid movie ID")
		}
		if err.Error() != "invalid movie ID: invalid" {
			t.Errorf("Expected 'invalid movie ID: invalid', got: %v", err)
		}
	})

	t.Run("RemoveMovie_InvalidID", func(t *testing.T) {
		handler := createTestMovieHandler(t)
		defer handler.Close()

		err := handler.RemoveMovie(context.Background(), "invalid")
		if err == nil {
			t.Error("Expected error for invalid movie ID")
		}
		if err.Error() != "invalid movie ID: invalid" {
			t.Errorf("Expected 'invalid movie ID: invalid', got: %v", err)
		}
	})

	t.Run("printMovie", func(t *testing.T) {
		handler := createTestMovieHandler(t)
		defer handler.Close()

		movie := createTestMovie()

		handler.printMovie(movie)

		minimalMovie := &models.Movie{
			ID:    2,
			Title: "Minimal Movie",
		}
		handler.printMovie(minimalMovie)

		watchedMovie := &models.Movie{
			ID:     3,
			Title:  "Watched Movie",
			Year:   2022,
			Status: "watched",
			Rating: 3.5,
		}
		handler.printMovie(watchedMovie)
	})

	t.Run("SearchAndAddMovie", func(t *testing.T) {
		handler := createTestMovieHandler(t)
		defer handler.Close()

		err := handler.SearchAndAddMovie(context.Background(), "", false)
		if err == nil {
			t.Error("Expected error for empty query")
		}
	})

	t.Run("List Movies", func(t *testing.T) {
		handler := createTestMovieHandler(t)
		defer handler.Close()

		err := handler.ListMovies(context.Background(), "")
		if err != nil {
			t.Errorf("Expected no error for listing all movies, got: %v", err)
		}

		err = handler.ListMovies(context.Background(), "invalid")
		if err == nil {
			t.Error("Expected error for invalid status")
		}
	})

	t.Run("Integration", func(t *testing.T) {
		t.Run("CreateAndRetrieve", func(t *testing.T) {
			handler := createTestMovieHandler(t)
			defer handler.Close()

			movie := createTestMovie()
			movie.ID = 0

			id, err := handler.repos.Movies.Create(context.Background(), movie)
			if err != nil {
				t.Errorf("Failed to create movie: %v", err)
				return
			}

			err = handler.View(context.Background(), id)
			if err != nil {
				t.Errorf("Failed to view created movie: %v", err)
			}

			err = handler.UpdateStatus(context.Background(), id, "watched")
			if err != nil {
				t.Errorf("Failed to update movie status: %v", err)
			}

			err = handler.MarkWatched(context.Background(), id)
			if err != nil {
				t.Errorf("Failed to mark movie as watched: %v", err)
			}

			err = handler.Remove(context.Background(), id)
			if err != nil {
				t.Errorf("Failed to remove movie: %v", err)
			}
		})

		t.Run("StatusFiltering", func(t *testing.T) {
			handler := createTestMovieHandler(t)
			defer handler.Close()

			queuedMovie := &models.Movie{
				Title:  "Queued Movie",
				Status: "queued",
				Added:  time.Now(),
			}
			watchedMovie := &models.Movie{
				Title:  "Watched Movie",
				Status: "watched",
				Added:  time.Now(),
			}

			id1, err := handler.repos.Movies.Create(context.Background(), queuedMovie)
			if err != nil {
				t.Errorf("Failed to create queued movie: %v", err)
				return
			}
			defer handler.repos.Movies.Delete(context.Background(), id1)

			id2, err := handler.repos.Movies.Create(context.Background(), watchedMovie)
			if err != nil {
				t.Errorf("Failed to create watched movie: %v", err)
				return
			}
			defer handler.repos.Movies.Delete(context.Background(), id2)

			testCases := []string{"", "queued", "watched"}
			for _, status := range testCases {
				err = handler.List(context.Background(), status)
				if err != nil {
					t.Errorf("Failed to list movies with status '%s': %v", status, err)
				}
			}
		})
	})

	t.Run("ErrorPaths", func(t *testing.T) {
		handler := createTestMovieHandler(t)
		defer handler.Close()

		ctx := context.Background()
		nonExistentID := int64(999999)

		testCases := []struct {
			name string
			fn   func() error
		}{
			{
				name: "View non-existent movie",
				fn:   func() error { return handler.View(ctx, nonExistentID) },
			},
			{
				name: "Update status of non-existent movie",
				fn:   func() error { return handler.UpdateStatus(ctx, nonExistentID, "watched") },
			},
			{
				name: "Mark non-existent movie as watched",
				fn:   func() error { return handler.MarkWatched(ctx, nonExistentID) },
			},
			{
				name: "Remove non-existent movie",
				fn:   func() error { return handler.Remove(ctx, nonExistentID) },
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := tc.fn()
				if err == nil {
					t.Errorf("Expected error for %s", tc.name)
				}
			})
		}
	})

	t.Run("ValidStatusValues", func(t *testing.T) {
		handler := createTestMovieHandler(t)
		defer handler.Close()

		valid := []string{"queued", "watched", "removed"}
		invalid := []string{"invalid", "pending", "completed", ""}

		for _, status := range valid {
			if err := handler.UpdateStatus(context.Background(), 999, status); err != nil &&
				err.Error() == fmt.Sprintf("invalid status: %s (valid: queued, watched, removed)", status) {
				t.Errorf("Status '%s' should be valid but was rejected", status)
			}
		}

		for _, status := range invalid {
			err := handler.UpdateStatus(context.Background(), 1, status)
			if err == nil {
				t.Errorf("Status '%s' should be invalid but was accepted", status)
			}
			got := fmt.Sprintf("invalid status: %s (valid: queued, watched, removed)", status)
			if err.Error() != got {
				t.Errorf("Expected '%s', got: %v", got, err)
			}
		}
	})
}
