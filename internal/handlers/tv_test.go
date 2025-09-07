package handlers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stormlightlabs/noteleaf/internal/models"
)

func createTestTVHandler(t *testing.T) *TVHandler {
	handler, err := NewTVHandler()
	if err != nil {
		t.Fatalf("Failed to create test TV handler: %v", err)
	}
	return handler
}

func createTestTVShow() *models.TVShow {
	now := time.Now()
	return &models.TVShow{
		ID:     1,
		Title:  "Test TV Show",
		Season: 1,
		Status: "queued",
		Rating: 4.5,
		Notes:  "Test notes",
		Added:  now,
	}
}

func TestTVHandler(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		handler := createTestTVHandler(t)
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
		handler := createTestTVHandler(t)

		err := handler.Close()
		if err != nil {
			t.Errorf("Expected no error when closing handler, got: %v", err)
		}
	})

	t.Run("Search and Add", func(t *testing.T) {
		t.Run("Empty Query", func(t *testing.T) {
			handler := createTestTVHandler(t)
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
			handler := createTestTVHandler(t)
			defer handler.Close()

			err := handler.SearchAndAdd(context.Background(), "test show", false)
			if err != nil {
				t.Logf("Search failed as expected in test environment: %v", err)
			}
		})

		t.Run("Network Error", func(t *testing.T) {
			handler := createTestTVHandler(t)
			defer handler.Close()

			err := handler.SearchAndAdd(context.Background(), "unlikely_to_find_this_show_12345", false)
			if err != nil {
				t.Logf("Network error encountered (expected in test environment): %v", err)
			}
		})
	})

	t.Run("List", func(t *testing.T) {
		t.Run("Invalid Status", func(t *testing.T) {
			handler := createTestTVHandler(t)
			defer handler.Close()

			err := handler.List(context.Background(), "invalid_status")
			if err == nil {
				t.Error("Expected error for invalid status")
			}
			if err.Error() != "invalid status: invalid_status (use: queued, watching, watched, or leave empty for all)" {
				t.Errorf("Expected invalid status error, got: %v", err)
			}
		})

		t.Run("All Shows", func(t *testing.T) {
			handler := createTestTVHandler(t)
			defer handler.Close()

			err := handler.List(context.Background(), "")
			if err != nil {
				t.Errorf("Expected no error for listing all TV shows, got: %v", err)
			}
		})

		t.Run("Queued Shows", func(t *testing.T) {
			handler := createTestTVHandler(t)
			defer handler.Close()

			err := handler.List(context.Background(), "queued")
			if err != nil {
				t.Errorf("Expected no error for listing queued TV shows, got: %v", err)
			}
		})

		t.Run("Watching Shows", func(t *testing.T) {
			handler := createTestTVHandler(t)
			defer handler.Close()

			err := handler.List(context.Background(), "watching")
			if err != nil {
				t.Errorf("Expected no error for listing watching TV shows, got: %v", err)
			}
		})

		t.Run("Watched Shows", func(t *testing.T) {
			handler := createTestTVHandler(t)
			defer handler.Close()

			err := handler.List(context.Background(), "watched")
			if err != nil {
				t.Errorf("Expected no error for listing watched TV shows, got: %v", err)
			}
		})
	})

	t.Run("View", func(t *testing.T) {
		t.Run("Show Not Found", func(t *testing.T) {
			handler := createTestTVHandler(t)
			defer handler.Close()

			err := handler.View(context.Background(), 999)
			if err == nil {
				t.Error("Expected error for non-existent TV show")
			}
		})

		t.Run("Invalid ID", func(t *testing.T) {
			handler := createTestTVHandler(t)
			defer handler.Close()

			err := handler.ViewTVShow(context.Background(), "invalid")
			if err == nil {
				t.Error("Expected error for invalid TV show ID")
			}
			if err.Error() != "invalid TV show ID: invalid" {
				t.Errorf("Expected 'invalid TV show ID: invalid', got: %v", err)
			}
		})
	})

	t.Run("Update", func(t *testing.T) {
		t.Run("Update Status", func(t *testing.T) {
			t.Run("Invalid", func(t *testing.T) {
				handler := createTestTVHandler(t)
				defer handler.Close()

				err := handler.UpdateStatus(context.Background(), 1, "invalid")
				if err == nil {
					t.Error("Expected error for invalid status")
				}
				if err.Error() != "invalid status: invalid (valid: queued, watching, watched, removed)" {
					t.Errorf("Expected invalid status error, got: %v", err)
				}
			})

			t.Run("Show Not Found", func(t *testing.T) {
				handler := createTestTVHandler(t)
				defer handler.Close()

				err := handler.UpdateStatus(context.Background(), 999, "watched")
				if err == nil {
					t.Error("Expected error for non-existent TV show")
				}
			})
		})
	})

	t.Run("MarkWatching_ShowNotFound", func(t *testing.T) {
		handler := createTestTVHandler(t)
		defer handler.Close()

		err := handler.MarkWatching(context.Background(), 999)
		if err == nil {
			t.Error("Expected error for non-existent TV show")
		}
	})

	t.Run("MarkWatched_ShowNotFound", func(t *testing.T) {
		handler := createTestTVHandler(t)
		defer handler.Close()

		err := handler.MarkWatched(context.Background(), 999)
		if err == nil {
			t.Error("Expected error for non-existent TV show")
		}
	})

	t.Run("Remove_ShowNotFound", func(t *testing.T) {
		handler := createTestTVHandler(t)
		defer handler.Close()

		err := handler.Remove(context.Background(), 999)
		if err == nil {
			t.Error("Expected error for non-existent TV show")
		}
	})

	t.Run("UpdateTVShowStatus_InvalidID", func(t *testing.T) {
		handler := createTestTVHandler(t)
		defer handler.Close()

		err := handler.UpdateTVShowStatus(context.Background(), "invalid", "watched")
		if err == nil {
			t.Error("Expected error for invalid TV show ID")
		}
		if err.Error() != "invalid TV show ID: invalid" {
			t.Errorf("Expected 'invalid TV show ID: invalid', got: %v", err)
		}
	})

	t.Run("MarkTVShowWatching_InvalidID", func(t *testing.T) {
		handler := createTestTVHandler(t)
		defer handler.Close()

		err := handler.MarkTVShowWatching(context.Background(), "invalid")
		if err == nil {
			t.Error("Expected error for invalid TV show ID")
		}
		if err.Error() != "invalid TV show ID: invalid" {
			t.Errorf("Expected 'invalid TV show ID: invalid', got: %v", err)
		}
	})

	t.Run("MarkTVShowWatched_InvalidID", func(t *testing.T) {
		handler := createTestTVHandler(t)
		defer handler.Close()

		err := handler.MarkTVShowWatched(context.Background(), "invalid")
		if err == nil {
			t.Error("Expected error for invalid TV show ID")
		}
		if err.Error() != "invalid TV show ID: invalid" {
			t.Errorf("Expected 'invalid TV show ID: invalid', got: %v", err)
		}
	})

	t.Run("RemoveTVShow_InvalidID", func(t *testing.T) {
		handler := createTestTVHandler(t)
		defer handler.Close()

		err := handler.RemoveTVShow(context.Background(), "invalid")
		if err == nil {
			t.Error("Expected error for invalid TV show ID")
		}
		if err.Error() != "invalid TV show ID: invalid" {
			t.Errorf("Expected 'invalid TV show ID: invalid', got: %v", err)
		}
	})

	t.Run("printTVShow", func(t *testing.T) {
		handler := createTestTVHandler(t)
		defer handler.Close()

		show := createTestTVShow()

		handler.printTVShow(show)

		minimalShow := &models.TVShow{
			ID:    2,
			Title: "Minimal Show",
		}
		handler.printTVShow(minimalShow)

		watchedShow := &models.TVShow{
			ID:      3,
			Title:   "Watched Show",
			Season:  2,
			Episode: 5,
			Status:  "watched",
			Rating:  3.5,
		}
		handler.printTVShow(watchedShow)
	})

	t.Run("SearchAndAddTV", func(t *testing.T) {
		handler := createTestTVHandler(t)
		defer handler.Close()

		err := handler.SearchAndAddTV(context.Background(), "", false)
		if err == nil {
			t.Error("Expected error for empty query")
		}
	})

	t.Run("List TV Shows", func(t *testing.T) {
		handler := createTestTVHandler(t)
		defer handler.Close()

		err := handler.ListTVShows(context.Background(), "")
		if err != nil {
			t.Errorf("Expected no error for listing all TV shows, got: %v", err)
		}

		err = handler.ListTVShows(context.Background(), "invalid")
		if err == nil {
			t.Error("Expected error for invalid status")
		}
	})

	t.Run("Integration", func(t *testing.T) {
		t.Run("CreateAndRetrieve", func(t *testing.T) {
			handler := createTestTVHandler(t)
			defer handler.Close()

			show := createTestTVShow()
			show.ID = 0

			id, err := handler.repos.TV.Create(context.Background(), show)
			if err != nil {
				t.Errorf("Failed to create TV show: %v", err)
				return
			}

			err = handler.View(context.Background(), id)
			if err != nil {
				t.Errorf("Failed to view created TV show: %v", err)
			}

			err = handler.UpdateStatus(context.Background(), id, "watching")
			if err != nil {
				t.Errorf("Failed to update TV show status: %v", err)
			}

			err = handler.MarkWatched(context.Background(), id)
			if err != nil {
				t.Errorf("Failed to mark TV show as watched: %v", err)
			}

			err = handler.MarkWatching(context.Background(), id)
			if err != nil {
				t.Errorf("Failed to mark TV show as watching: %v", err)
			}

			err = handler.Remove(context.Background(), id)
			if err != nil {
				t.Errorf("Failed to remove TV show: %v", err)
			}
		})

		t.Run("StatusFiltering", func(t *testing.T) {
			handler := createTestTVHandler(t)
			defer handler.Close()

			queuedShow := &models.TVShow{
				Title:  "Queued Show",
				Status: "queued",
				Added:  time.Now(),
			}
			watchingShow := &models.TVShow{
				Title:  "Watching Show",
				Status: "watching",
				Added:  time.Now(),
			}
			watchedShow := &models.TVShow{
				Title:  "Watched Show",
				Status: "watched",
				Added:  time.Now(),
			}

			id1, err := handler.repos.TV.Create(context.Background(), queuedShow)
			if err != nil {
				t.Errorf("Failed to create queued show: %v", err)
				return
			}
			defer handler.repos.TV.Delete(context.Background(), id1)

			id2, err := handler.repos.TV.Create(context.Background(), watchingShow)
			if err != nil {
				t.Errorf("Failed to create watching show: %v", err)
				return
			}
			defer handler.repos.TV.Delete(context.Background(), id2)

			id3, err := handler.repos.TV.Create(context.Background(), watchedShow)
			if err != nil {
				t.Errorf("Failed to create watched show: %v", err)
				return
			}
			defer handler.repos.TV.Delete(context.Background(), id3)

			testCases := []string{"", "queued", "watching", "watched"}
			for _, status := range testCases {
				err = handler.List(context.Background(), status)
				if err != nil {
					t.Errorf("Failed to list TV shows with status '%s': %v", status, err)
				}
			}
		})
	})

	t.Run("ErrorPaths", func(t *testing.T) {
		handler := createTestTVHandler(t)
		defer handler.Close()

		ctx := context.Background()
		nonExistentID := int64(999999)

		testCases := []struct {
			name string
			fn   func() error
		}{
			{
				name: "View non-existent show",
				fn:   func() error { return handler.View(ctx, nonExistentID) },
			},
			{
				name: "Update status of non-existent show",
				fn:   func() error { return handler.UpdateStatus(ctx, nonExistentID, "watched") },
			},
			{
				name: "Mark non-existent show as watching",
				fn:   func() error { return handler.MarkWatching(ctx, nonExistentID) },
			},
			{
				name: "Mark non-existent show as watched",
				fn:   func() error { return handler.MarkWatched(ctx, nonExistentID) },
			},
			{
				name: "Remove non-existent show",
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
		handler := createTestTVHandler(t)
		defer handler.Close()

		valid := []string{"queued", "watching", "watched", "removed"}
		invalid := []string{"invalid", "pending", "completed", ""}

		for _, status := range valid {
			if err := handler.UpdateStatus(context.Background(), 999, status); err != nil &&
				err.Error() == fmt.Sprintf("invalid status: %s (valid: queued, watching, watched, removed)", status) {
				t.Errorf("Status '%s' should be valid but was rejected", status)
			}
		}

		for _, status := range invalid {
			err := handler.UpdateStatus(context.Background(), 1, status)
			if err == nil {
				t.Errorf("Status '%s' should be invalid but was accepted", status)
			}
			got := fmt.Sprintf("invalid status: %s (valid: queued, watching, watched, removed)", status)
			if err.Error() != got {
				t.Errorf("Expected '%s', got: %v", got, err)
			}
		}
	})
}
