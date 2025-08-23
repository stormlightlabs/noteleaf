package repo

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"stormlightlabs.org/noteleaf/internal/models"
)

func createMovieTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create in-memory database: %v", err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	schema := `
		CREATE TABLE IF NOT EXISTS movies (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			year INTEGER,
			status TEXT DEFAULT 'queued',
			rating REAL,
			notes TEXT,
			added DATETIME DEFAULT CURRENT_TIMESTAMP,
			watched DATETIME
		);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func createSampleMovie() *models.Movie {
	return &models.Movie{
		Title:  "Test Movie",
		Year:   2023,
		Status: "queued",
		Rating: 8.5,
		Notes:  "Great movie to watch",
	}
}

func TestMovieRepository(t *testing.T) {
	t.Run("CRUD Operations", func(t *testing.T) {
		db := createMovieTestDB(t)
		repo := NewMovieRepository(db)
		ctx := context.Background()

		t.Run("Create Movie", func(t *testing.T) {
			movie := createSampleMovie()

			id, err := repo.Create(ctx, movie)
			if err != nil {
				t.Errorf("Failed to create movie: %v", err)
			}

			if id == 0 {
				t.Error("Expected non-zero ID")
			}

			if movie.ID != id {
				t.Errorf("Expected movie ID to be set to %d, got %d", id, movie.ID)
			}

			if movie.Added.IsZero() {
				t.Error("Expected Added timestamp to be set")
			}
		})

		t.Run("Get Movie", func(t *testing.T) {
			original := createSampleMovie()
			id, err := repo.Create(ctx, original)
			if err != nil {
				t.Fatalf("Failed to create movie: %v", err)
			}

			retrieved, err := repo.Get(ctx, id)
			if err != nil {
				t.Errorf("Failed to get movie: %v", err)
			}

			if retrieved.Title != original.Title {
				t.Errorf("Expected title %s, got %s", original.Title, retrieved.Title)
			}
			if retrieved.Year != original.Year {
				t.Errorf("Expected year %d, got %d", original.Year, retrieved.Year)
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

		t.Run("Update Movie", func(t *testing.T) {
			movie := createSampleMovie()
			id, err := repo.Create(ctx, movie)
			if err != nil {
				t.Fatalf("Failed to create movie: %v", err)
			}

			movie.Title = "Updated Movie"
			movie.Status = "watched"
			movie.Rating = 9.0
			now := time.Now()
			movie.Watched = &now

			err = repo.Update(ctx, movie)
			if err != nil {
				t.Errorf("Failed to update movie: %v", err)
			}

			updated, err := repo.Get(ctx, id)
			if err != nil {
				t.Fatalf("Failed to get updated movie: %v", err)
			}

			if updated.Title != "Updated Movie" {
				t.Errorf("Expected updated title, got %s", updated.Title)
			}
			if updated.Status != "watched" {
				t.Errorf("Expected status watched, got %s", updated.Status)
			}
			if updated.Rating != 9.0 {
				t.Errorf("Expected rating 9.0, got %f", updated.Rating)
			}
			if updated.Watched == nil {
				t.Error("Expected watched time to be set")
			}
		})

		t.Run("Delete Movie", func(t *testing.T) {
			movie := createSampleMovie()
			id, err := repo.Create(ctx, movie)
			if err != nil {
				t.Fatalf("Failed to create movie: %v", err)
			}

			err = repo.Delete(ctx, id)
			if err != nil {
				t.Errorf("Failed to delete movie: %v", err)
			}

			_, err = repo.Get(ctx, id)
			if err == nil {
				t.Error("Expected error when getting deleted movie")
			}
		})
	})

	t.Run("List", func(t *testing.T) {
		db := createMovieTestDB(t)
		repo := NewMovieRepository(db)
		ctx := context.Background()

		movies := []*models.Movie{
			{Title: "Movie 1", Year: 2020, Status: "queued", Rating: 8.0},
			{Title: "Movie 2", Year: 2021, Status: "watched", Rating: 7.5},
			{Title: "Movie 3", Year: 2022, Status: "queued", Rating: 9.0},
		}

		for _, movie := range movies {
			_, err := repo.Create(ctx, movie)
			if err != nil {
				t.Fatalf("Failed to create movie: %v", err)
			}
		}

		t.Run("List All Movies", func(t *testing.T) {
			results, err := repo.List(ctx, MovieListOptions{})
			if err != nil {
				t.Errorf("Failed to list movies: %v", err)
			}

			if len(results) != 3 {
				t.Errorf("Expected 3 movies, got %d", len(results))
			}
		})

		t.Run("List Movies with Status Filter", func(t *testing.T) {
			results, err := repo.List(ctx, MovieListOptions{Status: "queued"})
			if err != nil {
				t.Errorf("Failed to list movies: %v", err)
			}

			if len(results) != 2 {
				t.Errorf("Expected 2 queued movies, got %d", len(results))
			}

			for _, movie := range results {
				if movie.Status != "queued" {
					t.Errorf("Expected queued status, got %s", movie.Status)
				}
			}
		})

		t.Run("List Movies with Year Filter", func(t *testing.T) {
			results, err := repo.List(ctx, MovieListOptions{Year: 2021})
			if err != nil {
				t.Errorf("Failed to list movies: %v", err)
			}

			if len(results) != 1 {
				t.Errorf("Expected 1 movie from 2021, got %d", len(results))
			}

			if len(results) > 0 && results[0].Year != 2021 {
				t.Errorf("Expected year 2021, got %d", results[0].Year)
			}
		})

		t.Run("List Movies with Rating Filter", func(t *testing.T) {
			results, err := repo.List(ctx, MovieListOptions{MinRating: 8.0})
			if err != nil {
				t.Errorf("Failed to list movies: %v", err)
			}

			if len(results) != 2 {
				t.Errorf("Expected 2 movies with rating >= 8.0, got %d", len(results))
			}

			for _, movie := range results {
				if movie.Rating < 8.0 {
					t.Errorf("Expected rating >= 8.0, got %f", movie.Rating)
				}
			}
		})

		t.Run("List Movies with Search", func(t *testing.T) {
			results, err := repo.List(ctx, MovieListOptions{Search: "Movie 1"})
			if err != nil {
				t.Errorf("Failed to list movies: %v", err)
			}

			if len(results) != 1 {
				t.Errorf("Expected 1 movie matching search, got %d", len(results))
			}

			if len(results) > 0 && results[0].Title != "Movie 1" {
				t.Errorf("Expected 'Movie 1', got %s", results[0].Title)
			}
		})

		t.Run("List Movies with Limit", func(t *testing.T) {
			results, err := repo.List(ctx, MovieListOptions{Limit: 2})
			if err != nil {
				t.Errorf("Failed to list movies: %v", err)
			}

			if len(results) != 2 {
				t.Errorf("Expected 2 movies due to limit, got %d", len(results))
			}
		})
	})

	t.Run("Special Methods", func(t *testing.T) {
		db := createMovieTestDB(t)
		repo := NewMovieRepository(db)
		ctx := context.Background()

		movie1 := &models.Movie{Title: "Queued Movie", Status: "queued", Rating: 8.0}
		movie2 := &models.Movie{Title: "Watched Movie", Status: "watched", Rating: 9.0}
		movie3 := &models.Movie{Title: "Another Queued", Status: "queued", Rating: 7.0}

		var movie1ID int64
		for _, movie := range []*models.Movie{movie1, movie2, movie3} {
			id, err := repo.Create(ctx, movie)
			if err != nil {
				t.Fatalf("Failed to create movie: %v", err)
			}
			if movie == movie1 {
				movie1ID = id
			}
		}

		t.Run("GetQueued", func(t *testing.T) {
			results, err := repo.GetQueued(ctx)
			if err != nil {
				t.Errorf("Failed to get queued movies: %v", err)
			}

			if len(results) != 2 {
				t.Errorf("Expected 2 queued movies, got %d", len(results))
			}

			for _, movie := range results {
				if movie.Status != "queued" {
					t.Errorf("Expected queued status, got %s", movie.Status)
				}
			}
		})

		t.Run("GetWatched", func(t *testing.T) {
			results, err := repo.GetWatched(ctx)
			if err != nil {
				t.Errorf("Failed to get watched movies: %v", err)
			}

			if len(results) != 1 {
				t.Errorf("Expected 1 watched movie, got %d", len(results))
			}

			if len(results) > 0 && results[0].Status != "watched" {
				t.Errorf("Expected watched status, got %s", results[0].Status)
			}
		})

		t.Run("MarkWatched", func(t *testing.T) {
			err := repo.MarkWatched(ctx, movie1ID)
			if err != nil {
				t.Errorf("Failed to mark movie as watched: %v", err)
			}

			updated, err := repo.Get(ctx, movie1ID)
			if err != nil {
				t.Fatalf("Failed to get updated movie: %v", err)
			}

			if updated.Status != "watched" {
				t.Errorf("Expected status to be watched, got %s", updated.Status)
			}

			if updated.Watched == nil {
				t.Error("Expected watched timestamp to be set")
			}
		})
	})

	t.Run("Count", func(t *testing.T) {
		db := createMovieTestDB(t)
		repo := NewMovieRepository(db)
		ctx := context.Background()

		movies := []*models.Movie{
			{Title: "Movie 1", Status: "queued", Rating: 8.0},
			{Title: "Movie 2", Status: "watched", Rating: 7.0},
			{Title: "Movie 3", Status: "queued", Rating: 9.0},
		}

		for _, movie := range movies {
			_, err := repo.Create(ctx, movie)
			if err != nil {
				t.Fatalf("Failed to create movie: %v", err)
			}
		}

		t.Run("Count all movies", func(t *testing.T) {
			count, err := repo.Count(ctx, MovieListOptions{})
			if err != nil {
				t.Errorf("Failed to count movies: %v", err)
			}

			if count != 3 {
				t.Errorf("Expected 3 movies, got %d", count)
			}
		})

		t.Run("Count queued movies", func(t *testing.T) {
			count, err := repo.Count(ctx, MovieListOptions{Status: "queued"})
			if err != nil {
				t.Errorf("Failed to count queued movies: %v", err)
			}

			if count != 2 {
				t.Errorf("Expected 2 queued movies, got %d", count)
			}
		})

		t.Run("Count movies by rating", func(t *testing.T) {
			count, err := repo.Count(ctx, MovieListOptions{MinRating: 8.0})
			if err != nil {
				t.Errorf("Failed to count high-rated movies: %v", err)
			}

			if count != 2 {
				t.Errorf("Expected 2 movies with rating >= 8.0, got %d", count)
			}
		})
	})
}
