package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stormlightlabs/noteleaf/internal/models"
	"github.com/stormlightlabs/noteleaf/internal/repo"
	"github.com/stormlightlabs/noteleaf/internal/services"
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

		t.Run("Context Cancellation During Search", func(t *testing.T) {
			handler := createTestMovieHandler(t)
			defer handler.Close()

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			err := handler.SearchAndAdd(ctx, "test movie", false)
			if err == nil {
				t.Error("Expected error for cancelled context")
			}
		})

		t.Run("Search Service Error", func(t *testing.T) {
			handler := createTestMovieHandler(t)
			defer handler.Close()

			mockFetcher := &MockMediaFetcher{
				ShouldError:  true,
				ErrorMessage: "network error",
			}

			handler.service = CreateTestMovieService(mockFetcher)

			err := handler.SearchAndAdd(context.Background(), "test movie", false)
			if err == nil {
				t.Error("Expected error when search service fails")
			}

			if !strings.Contains(err.Error(), "search failed") {
				t.Errorf("Expected search failure error, got: %v", err)
			}
		})

		t.Run("Empty Search Results", func(t *testing.T) {
			handler := createTestMovieHandler(t)
			defer handler.Close()

			mockFetcher := &MockMediaFetcher{SearchResults: []services.Media{}}

			handler.service = CreateTestMovieService(mockFetcher)

			if err := handler.SearchAndAdd(context.Background(), "nonexistent movie", false); err != nil {
				t.Errorf("Expected no error for empty results, got: %v", err)
			}
		})

		t.Run("Search Results with No Movies", func(t *testing.T) {
			handler := createTestMovieHandler(t)
			defer handler.Close()

			mockFetcher := &MockMediaFetcher{
				SearchResults: []services.Media{
					{Title: "Test TV Show", Link: "/tv/test_show", Type: "tv"},
				},
			}

			handler.service = CreateTestMovieService(mockFetcher)

			if err := handler.SearchAndAdd(context.Background(), "tv show", false); err != nil {
				t.Errorf("Expected no error for TV-only results, got: %v", err)
			}
		})

		t.Run("Interactive Mode Path", func(t *testing.T) {
			// Skip interactive mode test to prevent hanging in CI/test environments
			// TODO: Interactive mode uses TUI components that require terminal interaction
			t.Skip("Interactive mode requires terminal interaction, skipping to prevent hanging")
		})

		t.Run("successful search and add with user selection", func(t *testing.T) {
			t.Skip()
			handler := createTestMovieHandler(t)
			defer handler.Close()

			mockFetcher := &MockMediaFetcher{
				SearchResults: []services.Media{
					{Title: "Test Movie 1", Link: "/m/test_movie_1", Type: "movie", CriticScore: "85%"},
					{Title: "Test Movie 2", Link: "/m/test_movie_2", Type: "movie", CriticScore: "72%"},
				},
			}

			handler.service = CreateTestMovieService(mockFetcher)
			handler.SetInputReader(MenuSelection(1))

			if err := handler.SearchAndAdd(context.Background(), "test movie", false); err != nil {
				t.Errorf("Expected successful search and add, got error: %v", err)
			}

			movies, err := handler.repos.Movies.List(context.Background(), repo.MovieListOptions{})
			if err != nil {
				t.Fatalf("Failed to list movies: %v", err)
			}
			if len(movies) != 1 {
				t.Errorf("Expected 1 movie in database, got %d", len(movies))
			}
			if len(movies) > 0 && movies[0].Title != "Test Movie 1" {
				t.Errorf("Expected movie title 'Test Movie 1', got '%s'", movies[0].Title)
			}
		})

		t.Run("successful search with user cancellation", func(t *testing.T) {
			t.Skip()
			handler := createTestMovieHandler(t)
			defer handler.Close()

			mockFetcher := &MockMediaFetcher{
				SearchResults: []services.Media{
					{Title: "Another Movie", Link: "/m/another_movie", Type: "movie", CriticScore: "90%"},
				},
			}

			handler.service = CreateTestMovieService(mockFetcher)
			handler.SetInputReader(MenuCancel())

			err := handler.SearchAndAdd(context.Background(), "another movie", false)
			if err != nil {
				t.Errorf("Expected no error on cancellation, got: %v", err)
			}

			movies, err := handler.repos.Movies.List(context.Background(), repo.MovieListOptions{})
			if err != nil {
				t.Fatalf("Failed to list movies: %v", err)
			}

			expected := 1
			if len(movies) != expected {
				t.Errorf("Expected %d movies in database after cancellation, got %d", expected, len(movies))
			}
		})

		t.Run("invalid user choice", func(t *testing.T) {
			handler := createTestMovieHandler(t)
			defer handler.Close()

			mockFetcher := &MockMediaFetcher{
				SearchResults: []services.Media{
					{Title: "Choice Test Movie", Link: "/m/choice_test", Type: "movie", CriticScore: "75%"},
				},
			}

			handler.service = CreateTestMovieService(mockFetcher)
			handler.SetInputReader(MenuSelection(5))

			err := handler.SearchAndAdd(context.Background(), "choice test", false)
			if err == nil {
				t.Error("Expected error for invalid choice")
			}
			if err != nil && !strings.Contains(err.Error(), "invalid choice") {
				t.Errorf("Expected 'invalid choice' error, got: %v", err)
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

		err := handler.MarkWatched(context.Background(), "999")
		if err == nil {
			t.Error("Expected error for non-existent movie")
		}
	})

	t.Run("RemoveMovie_MovieNotFound", func(t *testing.T) {
		handler := createTestMovieHandler(t)
		defer handler.Close()

		err := handler.Remove(context.Background(), "999")
		if err == nil {
			t.Error("Expected error for non-existent movie")
		}
	})

	t.Run("MarkWatched_InvalidID", func(t *testing.T) {
		handler := createTestMovieHandler(t)
		defer handler.Close()

		err := handler.MarkWatched(context.Background(), "invalid")
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

		err := handler.Remove(context.Background(), "invalid")
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

			err = handler.MarkWatched(context.Background(), strconv.Itoa(int(id)))
			if err != nil {
				t.Errorf("Failed to mark movie as watched: %v", err)
			}

			err = handler.Remove(context.Background(), strconv.Itoa(int(id)))
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

			tt := []string{"", "queued", "watched"}
			for _, tc := range tt {
				err = handler.List(context.Background(), tc)
				if err != nil {
					t.Errorf("Failed to list movies with status '%s': %v", tc, err)
				}
			}
		})
	})

	t.Run("ErrorPaths", func(t *testing.T) {
		handler := createTestMovieHandler(t)
		defer handler.Close()

		ctx := context.Background()
		nonExistentID := int64(999999)

		tt := []struct {
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
				fn:   func() error { return handler.MarkWatched(ctx, strconv.Itoa(int(nonExistentID))) },
			},
			{
				name: "Remove non-existent movie",
				fn:   func() error { return handler.Remove(ctx, strconv.Itoa(int(nonExistentID))) },
			},
		}

		for _, tc := range tt {
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
