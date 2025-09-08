package repo

import (
	"context"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stormlightlabs/noteleaf/internal/models"
)

func TestTVRepository(t *testing.T) {
	t.Run("CRUD Operations", func(t *testing.T) {
		db := CreateTestDB(t)
		repo := NewTVRepository(db)
		ctx := context.Background()

		t.Run("Create TV Show", func(t *testing.T) {
			tvShow := CreateSampleTVShow()

			id, err := repo.Create(ctx, tvShow)
			if err != nil {
				t.Errorf("Failed to create TV show: %v", err)
			}

			if id == 0 {
				t.Error("Expected non-zero ID")
			}

			if tvShow.ID != id {
				t.Errorf("Expected TV show ID to be set to %d, got %d", id, tvShow.ID)
			}

			if tvShow.Added.IsZero() {
				t.Error("Expected Added timestamp to be set")
			}
		})

		t.Run("Get TV Show", func(t *testing.T) {
			original := CreateSampleTVShow()
			id, err := repo.Create(ctx, original)
			if err != nil {
				t.Fatalf("Failed to create TV show: %v", err)
			}

			retrieved, err := repo.Get(ctx, id)
			if err != nil {
				t.Errorf("Failed to get TV show: %v", err)
			}

			if retrieved.Title != original.Title {
				t.Errorf("Expected title %s, got %s", original.Title, retrieved.Title)
			}
			if retrieved.Season != original.Season {
				t.Errorf("Expected season %d, got %d", original.Season, retrieved.Season)
			}
			if retrieved.Episode != original.Episode {
				t.Errorf("Expected episode %d, got %d", original.Episode, retrieved.Episode)
			}
			if retrieved.Status != original.Status {
				t.Errorf("Expected status %s, got %s", original.Status, retrieved.Status)
			}
			if retrieved.Rating != original.Rating {
				t.Errorf("Expected rating %f, got %f", original.Rating, retrieved.Rating)
			}
			if retrieved.Notes != original.Notes {
				t.Errorf("Expected notes %s, got %s", original.Notes, retrieved.Notes)
			}
		})

		t.Run("Update TV Show", func(t *testing.T) {
			tvShow := CreateSampleTVShow()
			id, err := repo.Create(ctx, tvShow)
			if err != nil {
				t.Fatalf("Failed to create TV show: %v", err)
			}

			tvShow.Title = "Updated Show"
			tvShow.Season = 2
			tvShow.Episode = 5
			tvShow.Status = "watching"
			tvShow.Rating = 9.5
			now := time.Now()
			tvShow.LastWatched = &now

			err = repo.Update(ctx, tvShow)
			if err != nil {
				t.Errorf("Failed to update TV show: %v", err)
			}

			updated, err := repo.Get(ctx, id)
			if err != nil {
				t.Fatalf("Failed to get updated TV show: %v", err)
			}

			if updated.Title != "Updated Show" {
				t.Errorf("Expected updated title, got %s", updated.Title)
			}
			if updated.Season != 2 {
				t.Errorf("Expected season 2, got %d", updated.Season)
			}
			if updated.Episode != 5 {
				t.Errorf("Expected episode 5, got %d", updated.Episode)
			}
			if updated.Status != "watching" {
				t.Errorf("Expected status watching, got %s", updated.Status)
			}
			if updated.Rating != 9.5 {
				t.Errorf("Expected rating 9.5, got %f", updated.Rating)
			}
			if updated.LastWatched == nil {
				t.Error("Expected last watched time to be set")
			}
		})

		t.Run("Delete TV Show", func(t *testing.T) {
			tvShow := CreateSampleTVShow()
			id, err := repo.Create(ctx, tvShow)
			if err != nil {
				t.Fatalf("Failed to create TV show: %v", err)
			}

			err = repo.Delete(ctx, id)
			if err != nil {
				t.Errorf("Failed to delete TV show: %v", err)
			}

			_, err = repo.Get(ctx, id)
			if err == nil {
				t.Error("Expected error when getting deleted TV show")
			}
		})
	})

	t.Run("List", func(t *testing.T) {
		db := CreateTestDB(t)
		repo := NewTVRepository(db)
		ctx := context.Background()

		tvShows := []*models.TVShow{
			{Title: "Show A", Season: 1, Episode: 1, Status: "queued", Rating: 8.0},
			{Title: "Show A", Season: 1, Episode: 2, Status: "watching", Rating: 8.5},
			{Title: "Show B", Season: 1, Episode: 1, Status: "queued", Rating: 9.0},
			{Title: "Show B", Season: 2, Episode: 1, Status: "watched", Rating: 9.2},
		}

		for _, tvShow := range tvShows {
			_, err := repo.Create(ctx, tvShow)
			if err != nil {
				t.Fatalf("Failed to create TV show: %v", err)
			}
		}

		t.Run("List All TV Shows", func(t *testing.T) {
			results, err := repo.List(ctx, TVListOptions{})
			if err != nil {
				t.Errorf("Failed to list TV shows: %v", err)
			}

			if len(results) != 4 {
				t.Errorf("Expected 4 TV shows, got %d", len(results))
			}
		})

		t.Run("List TV Shows with Status Filter", func(t *testing.T) {
			results, err := repo.List(ctx, TVListOptions{Status: "queued"})
			if err != nil {
				t.Errorf("Failed to list TV shows: %v", err)
			}

			if len(results) != 2 {
				t.Errorf("Expected 2 queued TV shows, got %d", len(results))
			}

			for _, tvShow := range results {
				if tvShow.Status != "queued" {
					t.Errorf("Expected queued status, got %s", tvShow.Status)
				}
			}
		})

		t.Run("List TV Shows by Title", func(t *testing.T) {
			results, err := repo.List(ctx, TVListOptions{Title: "Show A"})
			if err != nil {
				t.Errorf("Failed to list TV shows: %v", err)
			}

			if len(results) != 2 {
				t.Errorf("Expected 2 episodes of Show A, got %d", len(results))
			}

			for _, tvShow := range results {
				if tvShow.Title != "Show A" {
					t.Errorf("Expected title 'Show A', got %s", tvShow.Title)
				}
			}
		})

		t.Run("List TV Shows by Season", func(t *testing.T) {
			results, err := repo.List(ctx, TVListOptions{Title: "Show B", Season: 1})
			if err != nil {
				t.Errorf("Failed to list TV shows: %v", err)
			}

			if len(results) != 1 {
				t.Errorf("Expected 1 episode of Show B season 1, got %d", len(results))
			}

			if len(results) > 0 {
				if results[0].Title != "Show B" || results[0].Season != 1 {
					t.Errorf("Expected Show B season 1, got %s season %d", results[0].Title, results[0].Season)
				}
			}
		})

		t.Run("List TV Shows with Rating Filter", func(t *testing.T) {
			results, err := repo.List(ctx, TVListOptions{MinRating: 9.0})
			if err != nil {
				t.Errorf("Failed to list TV shows: %v", err)
			}

			if len(results) != 2 {
				t.Errorf("Expected 2 TV shows with rating >= 9.0, got %d", len(results))
			}

			for _, tvShow := range results {
				if tvShow.Rating < 9.0 {
					t.Errorf("Expected rating >= 9.0, got %f", tvShow.Rating)
				}
			}
		})

		t.Run("List TV Shows with Search", func(t *testing.T) {
			results, err := repo.List(ctx, TVListOptions{Search: "Show A"})
			if err != nil {
				t.Errorf("Failed to list TV shows: %v", err)
			}

			if len(results) != 2 {
				t.Errorf("Expected 2 TV shows matching search, got %d", len(results))
			}

			for _, tvShow := range results {
				if tvShow.Title != "Show A" {
					t.Errorf("Expected 'Show A', got %s", tvShow.Title)
				}
			}
		})

		t.Run("List TV Shows with Limit", func(t *testing.T) {
			results, err := repo.List(ctx, TVListOptions{Limit: 2})
			if err != nil {
				t.Errorf("Failed to list TV shows: %v", err)
			}

			if len(results) != 2 {
				t.Errorf("Expected 2 TV shows due to limit, got %d", len(results))
			}
		})
	})

	t.Run("Special Methods", func(t *testing.T) {
		db := CreateTestDB(t)
		repo := NewTVRepository(db)
		ctx := context.Background()

		tvShow1 := &models.TVShow{Title: "Queued Show", Status: "queued", Rating: 8.0}
		tvShow2 := &models.TVShow{Title: "Watching Show", Status: "watching", Rating: 9.0}
		tvShow3 := &models.TVShow{Title: "Watched Show", Status: "watched", Rating: 8.5}
		tvShow4 := &models.TVShow{Title: "Test Series", Season: 1, Episode: 1, Status: "queued"}
		tvShow5 := &models.TVShow{Title: "Test Series", Season: 1, Episode: 2, Status: "queued"}
		tvShow6 := &models.TVShow{Title: "Test Series", Season: 2, Episode: 1, Status: "queued"}

		var tvShow1ID int64
		for _, tvShow := range []*models.TVShow{tvShow1, tvShow2, tvShow3, tvShow4, tvShow5, tvShow6} {
			id, err := repo.Create(ctx, tvShow)
			if err != nil {
				t.Fatalf("Failed to create TV show: %v", err)
			}
			if tvShow == tvShow1 {
				tvShow1ID = id
			}
		}

		t.Run("GetQueued", func(t *testing.T) {
			results, err := repo.GetQueued(ctx)
			if err != nil {
				t.Errorf("Failed to get queued TV shows: %v", err)
			}

			if len(results) != 4 {
				t.Errorf("Expected 4 queued TV shows, got %d", len(results))
			}

			for _, tvShow := range results {
				if tvShow.Status != "queued" {
					t.Errorf("Expected queued status, got %s", tvShow.Status)
				}
			}
		})

		t.Run("GetWatching", func(t *testing.T) {
			results, err := repo.GetWatching(ctx)
			if err != nil {
				t.Errorf("Failed to get watching TV shows: %v", err)
			}

			if len(results) != 1 {
				t.Errorf("Expected 1 watching TV show, got %d", len(results))
			}

			if len(results) > 0 && results[0].Status != "watching" {
				t.Errorf("Expected watching status, got %s", results[0].Status)
			}
		})

		t.Run("GetWatched", func(t *testing.T) {
			results, err := repo.GetWatched(ctx)
			if err != nil {
				t.Errorf("Failed to get watched TV shows: %v", err)
			}

			if len(results) != 1 {
				t.Errorf("Expected 1 watched TV show, got %d", len(results))
			}

			if len(results) > 0 && results[0].Status != "watched" {
				t.Errorf("Expected watched status, got %s", results[0].Status)
			}
		})

		t.Run("GetByTitle", func(t *testing.T) {
			results, err := repo.GetByTitle(ctx, "Test Series")
			if err != nil {
				t.Errorf("Failed to get TV shows by title: %v", err)
			}

			if len(results) != 3 {
				t.Errorf("Expected 3 episodes of Test Series, got %d", len(results))
			}

			for _, tvShow := range results {
				if tvShow.Title != "Test Series" {
					t.Errorf("Expected title 'Test Series', got %s", tvShow.Title)
				}
			}
		})

		t.Run("GetBySeason", func(t *testing.T) {
			results, err := repo.GetBySeason(ctx, "Test Series", 1)
			if err != nil {
				t.Errorf("Failed to get TV shows by season: %v", err)
			}

			if len(results) != 2 {
				t.Errorf("Expected 2 episodes of Test Series season 1, got %d", len(results))
			}

			for _, tvShow := range results {
				if tvShow.Title != "Test Series" || tvShow.Season != 1 {
					t.Errorf("Expected Test Series season 1, got %s season %d", tvShow.Title, tvShow.Season)
				}
			}
		})

		t.Run("MarkWatched", func(t *testing.T) {
			err := repo.MarkWatched(ctx, tvShow1ID)
			if err != nil {
				t.Errorf("Failed to mark TV show as watched: %v", err)
			}

			updated, err := repo.Get(ctx, tvShow1ID)
			if err != nil {
				t.Fatalf("Failed to get updated TV show: %v", err)
			}

			if updated.Status != "watched" {
				t.Errorf("Expected status to be watched, got %s", updated.Status)
			}

			if updated.LastWatched == nil {
				t.Error("Expected last watched timestamp to be set")
			}
		})

		t.Run("StartWatching", func(t *testing.T) {
			newShow := &models.TVShow{Title: "New Show", Status: "queued"}
			id, err := repo.Create(ctx, newShow)
			if err != nil {
				t.Fatalf("Failed to create new TV show: %v", err)
			}

			err = repo.StartWatching(ctx, id)
			if err != nil {
				t.Errorf("Failed to start watching TV show: %v", err)
			}

			updated, err := repo.Get(ctx, id)
			if err != nil {
				t.Fatalf("Failed to get updated TV show: %v", err)
			}

			if updated.Status != "watching" {
				t.Errorf("Expected status to be watching, got %s", updated.Status)
			}

			if updated.LastWatched == nil {
				t.Error("Expected last watched timestamp to be set")
			}
		})
	})

	t.Run("Count", func(t *testing.T) {
		db := CreateTestDB(t)
		repo := NewTVRepository(db)
		ctx := context.Background()

		tvShows := []*models.TVShow{
			{Title: "Show 1", Status: "queued", Rating: 8.0},
			{Title: "Show 2", Status: "watching", Rating: 7.0},
			{Title: "Show 3", Status: "watched", Rating: 9.0},
			{Title: "Show 4", Status: "queued", Rating: 8.5},
		}

		for _, tvShow := range tvShows {
			_, err := repo.Create(ctx, tvShow)
			if err != nil {
				t.Fatalf("Failed to create TV show: %v", err)
			}
		}

		t.Run("Count all TV shows", func(t *testing.T) {
			count, err := repo.Count(ctx, TVListOptions{})
			if err != nil {
				t.Errorf("Failed to count TV shows: %v", err)
			}

			if count != 4 {
				t.Errorf("Expected 4 TV shows, got %d", count)
			}
		})

		t.Run("Count queued TV shows", func(t *testing.T) {
			count, err := repo.Count(ctx, TVListOptions{Status: "queued"})
			if err != nil {
				t.Errorf("Failed to count queued TV shows: %v", err)
			}

			if count != 2 {
				t.Errorf("Expected 2 queued TV shows, got %d", count)
			}
		})

		t.Run("Count TV shows by rating", func(t *testing.T) {
			count, err := repo.Count(ctx, TVListOptions{MinRating: 8.0})
			if err != nil {
				t.Errorf("Failed to count high-rated TV shows: %v", err)
			}

			if count != 3 {
				t.Errorf("Expected 3 TV shows with rating >= 8.0, got %d", count)
			}
		})
	})
}
