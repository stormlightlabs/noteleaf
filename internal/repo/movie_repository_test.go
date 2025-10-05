package repo

import (
	"context"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stormlightlabs/noteleaf/internal/models"
)

func TestMovieRepository(t *testing.T) {
	t.Run("CRUD Operations", func(t *testing.T) {
		db := CreateTestDB(t)
		repo := NewMovieRepository(db)
		ctx := context.Background()

		t.Run("Create Movie", func(t *testing.T) {
			movie := CreateSampleMovie()

			id, err := repo.Create(ctx, movie)
			AssertNoError(t, err, "Failed to create movie")
			AssertNotEqual(t, int64(0), id, "Expected non-zero ID")
			AssertEqual(t, id, movie.ID, "Expected movie ID to be set correctly")
			AssertFalse(t, movie.Added.IsZero(), "Expected Added timestamp to be set")
		})

		t.Run("Get Movie", func(t *testing.T) {
			original := CreateSampleMovie()
			id, err := repo.Create(ctx, original)
			AssertNoError(t, err, "Failed to create movie")

			retrieved, err := repo.Get(ctx, id)
			AssertNoError(t, err, "Failed to get movie")

			AssertEqual(t, original.Title, retrieved.Title, "Title mismatch")
			AssertEqual(t, original.Year, retrieved.Year, "Year mismatch")
			AssertEqual(t, original.Status, retrieved.Status, "Status mismatch")
			AssertEqual(t, original.Rating, retrieved.Rating, "Rating mismatch")
			AssertEqual(t, original.Notes, retrieved.Notes, "Notes mismatch")
		})

		t.Run("Update Movie", func(t *testing.T) {
			movie := CreateSampleMovie()
			id, err := repo.Create(ctx, movie)
			AssertNoError(t, err, "Failed to create movie")

			movie.Title = "Updated Movie"
			movie.Status = "watched"
			movie.Rating = 9.0
			now := time.Now()
			movie.Watched = &now

			err = repo.Update(ctx, movie)
			AssertNoError(t, err, "Failed to update movie")

			updated, err := repo.Get(ctx, id)
			AssertNoError(t, err, "Failed to get updated movie")

			AssertEqual(t, "Updated Movie", updated.Title, "Expected updated title")
			AssertEqual(t, "watched", updated.Status, "Expected watched status")
			AssertEqual(t, 9.0, updated.Rating, "Expected rating 9.0")
			AssertTrue(t, updated.Watched != nil, "Expected watched time to be set")
		})

		t.Run("Delete Movie", func(t *testing.T) {
			movie := CreateSampleMovie()
			id, err := repo.Create(ctx, movie)
			AssertNoError(t, err, "Failed to create movie")

			err = repo.Delete(ctx, id)
			AssertNoError(t, err, "Failed to delete movie")

			_, err = repo.Get(ctx, id)
			AssertError(t, err, "Expected error when getting deleted movie")
		})
	})

	t.Run("List", func(t *testing.T) {
		db := CreateTestDB(t)
		repo := NewMovieRepository(db)
		ctx := context.Background()

		movies := []*models.Movie{
			{Title: "Movie 1", Year: 2020, Status: "queued", Rating: 8.0},
			{Title: "Movie 2", Year: 2021, Status: "watched", Rating: 7.5},
			{Title: "Movie 3", Year: 2022, Status: "queued", Rating: 9.0},
		}

		for _, movie := range movies {
			_, err := repo.Create(ctx, movie)
			AssertNoError(t, err, "Failed to create movie")
		}

		t.Run("List All Movies", func(t *testing.T) {
			results, err := repo.List(ctx, MovieListOptions{})
			AssertNoError(t, err, "Failed to list movies")
			AssertEqual(t, 3, len(results), "Expected 3 movies")
		})

		t.Run("List Movies with Status Filter", func(t *testing.T) {
			results, err := repo.List(ctx, MovieListOptions{Status: "queued"})
			AssertNoError(t, err, "Failed to list movies")
			AssertEqual(t, 2, len(results), "Expected 2 queued movies")

			for _, movie := range results {
				AssertEqual(t, "queued", movie.Status, "Expected queued status")
			}
		})

		t.Run("List Movies with Year Filter", func(t *testing.T) {
			results, err := repo.List(ctx, MovieListOptions{Year: 2021})
			AssertNoError(t, err, "Failed to list movies")
			AssertEqual(t, 1, len(results), "Expected 1 movie from 2021")

			if len(results) > 0 {
				AssertEqual(t, 2021, results[0].Year, "Expected year 2021")
			}
		})

		t.Run("List Movies with Rating Filter", func(t *testing.T) {
			results, err := repo.List(ctx, MovieListOptions{MinRating: 8.0})
			AssertNoError(t, err, "Failed to list movies")
			AssertEqual(t, 2, len(results), "Expected 2 movies with rating >= 8.0")

			for _, movie := range results {
				AssertTrue(t, movie.Rating >= 8.0, "Expected rating >= 8.0")
			}
		})

		t.Run("List Movies with Search", func(t *testing.T) {
			results, err := repo.List(ctx, MovieListOptions{Search: "Movie 1"})
			AssertNoError(t, err, "Failed to list movies")
			AssertEqual(t, 1, len(results), "Expected 1 movie matching search")

			if len(results) > 0 {
				AssertEqual(t, "Movie 1", results[0].Title, "Expected 'Movie 1'")
			}
		})

		t.Run("List Movies with Limit", func(t *testing.T) {
			results, err := repo.List(ctx, MovieListOptions{Limit: 2})
			AssertNoError(t, err, "Failed to list movies")
			AssertEqual(t, 2, len(results), "Expected 2 movies due to limit")
		})
	})

	t.Run("Special Methods", func(t *testing.T) {
		db := CreateTestDB(t)
		repo := NewMovieRepository(db)
		ctx := context.Background()

		movie1 := &models.Movie{Title: "Queued Movie", Status: "queued", Rating: 8.0}
		movie2 := &models.Movie{Title: "Watched Movie", Status: "watched", Rating: 9.0}
		movie3 := &models.Movie{Title: "Another Queued", Status: "queued", Rating: 7.0}

		var movie1ID int64
		for _, movie := range []*models.Movie{movie1, movie2, movie3} {
			id, err := repo.Create(ctx, movie)
			AssertNoError(t, err, "Failed to create movie")
			if movie == movie1 {
				movie1ID = id
			}
		}

		t.Run("GetQueued", func(t *testing.T) {
			results, err := repo.GetQueued(ctx)
			AssertNoError(t, err, "Failed to get queued movies")
			AssertEqual(t, 2, len(results), "Expected 2 queued movies")

			for _, movie := range results {
				AssertEqual(t, "queued", movie.Status, "Expected queued status")
			}
		})

		t.Run("GetWatched", func(t *testing.T) {
			results, err := repo.GetWatched(ctx)
			AssertNoError(t, err, "Failed to get watched movies")
			AssertEqual(t, 1, len(results), "Expected 1 watched movie")

			if len(results) > 0 {
				AssertEqual(t, "watched", results[0].Status, "Expected watched status")
			}
		})

		t.Run("MarkWatched", func(t *testing.T) {
			err := repo.MarkWatched(ctx, movie1ID)
			AssertNoError(t, err, "Failed to mark movie as watched")

			updated, err := repo.Get(ctx, movie1ID)
			AssertNoError(t, err, "Failed to get updated movie")

			AssertEqual(t, "watched", updated.Status, "Expected status to be watched")
			AssertTrue(t, updated.Watched != nil, "Expected watched timestamp to be set")
		})
	})

	t.Run("Count", func(t *testing.T) {
		db := CreateTestDB(t)
		repo := NewMovieRepository(db)
		ctx := context.Background()

		movies := []*models.Movie{
			{Title: "Movie 1", Status: "queued", Rating: 8.0},
			{Title: "Movie 2", Status: "watched", Rating: 7.0},
			{Title: "Movie 3", Status: "queued", Rating: 9.0},
		}

		for _, movie := range movies {
			_, err := repo.Create(ctx, movie)
			AssertNoError(t, err, "Failed to create movie")
		}

		t.Run("Count all movies", func(t *testing.T) {
			count, err := repo.Count(ctx, MovieListOptions{})
			AssertNoError(t, err, "Failed to count movies")
			AssertEqual(t, int64(3), count, "Expected 3 movies")
		})

		t.Run("Count queued movies", func(t *testing.T) {
			count, err := repo.Count(ctx, MovieListOptions{Status: "queued"})
			AssertNoError(t, err, "Failed to count queued movies")
			AssertEqual(t, int64(2), count, "Expected 2 queued movies")
		})

		t.Run("Count movies by rating", func(t *testing.T) {
			count, err := repo.Count(ctx, MovieListOptions{MinRating: 8.0})
			AssertNoError(t, err, "Failed to count high-rated movies")
			AssertEqual(t, int64(2), count, "Expected 2 movies with rating >= 8.0")
		})

		t.Run("Count with context cancellation", func(t *testing.T) {
			_, err := repo.Count(NewCanceledContext(), MovieListOptions{})
			AssertError(t, err, "Expected error with cancelled context")
		})
	})

	t.Run("Context Cancellation Error Paths", func(t *testing.T) {
		db := CreateTestDB(t)
		repo := NewMovieRepository(db)
		ctx := context.Background()

		movie := NewMovieBuilder().WithTitle("Test Movie").WithYear(2023).Build()
		id, err := repo.Create(ctx, movie)
		AssertNoError(t, err, "Failed to create movie")

		t.Run("Create with cancelled context", func(t *testing.T) {
			newMovie := NewMovieBuilder().WithTitle("Cancelled").Build()
			_, err := repo.Create(NewCanceledContext(), newMovie)
			AssertError(t, err, "Expected error with cancelled context")
		})

		t.Run("Get with cancelled context", func(t *testing.T) {
			_, err := repo.Get(NewCanceledContext(), id)
			AssertError(t, err, "Expected error with cancelled context")
		})

		t.Run("Update with cancelled context", func(t *testing.T) {
			movie.Title = "Updated"
			err := repo.Update(NewCanceledContext(), movie)
			AssertError(t, err, "Expected error with cancelled context")
		})

		t.Run("Delete with cancelled context", func(t *testing.T) {
			err := repo.Delete(NewCanceledContext(), id)
			AssertError(t, err, "Expected error with cancelled context")
		})

		t.Run("List with cancelled context", func(t *testing.T) {
			_, err := repo.List(NewCanceledContext(), MovieListOptions{})
			AssertError(t, err, "Expected error with cancelled context")
		})

		t.Run("GetQueued with cancelled context", func(t *testing.T) {
			_, err := repo.GetQueued(NewCanceledContext())
			AssertError(t, err, "Expected error with cancelled context")
		})

		t.Run("GetWatched with cancelled context", func(t *testing.T) {
			_, err := repo.GetWatched(NewCanceledContext())
			AssertError(t, err, "Expected error with cancelled context")
		})

		t.Run("MarkWatched with cancelled context", func(t *testing.T) {
			err := repo.MarkWatched(NewCanceledContext(), id)
			AssertError(t, err, "Expected error with cancelled context")
		})
	})

	t.Run("Edge Cases", func(t *testing.T) {
		db := CreateTestDB(t)
		repo := NewMovieRepository(db)
		ctx := context.Background()

		t.Run("Get non-existent movie", func(t *testing.T) {
			_, err := repo.Get(ctx, 99999)
			AssertError(t, err, "Expected error for non-existent movie")
		})

		t.Run("Update non-existent movie succeeds with no rows affected", func(t *testing.T) {
			movie := NewMovieBuilder().WithTitle("Non-existent").Build()
			movie.ID = 99999
			err := repo.Update(ctx, movie)
			AssertNoError(t, err, "Update should not error when no rows affected")
		})

		t.Run("Delete non-existent movie succeeds with no rows affected", func(t *testing.T) {
			err := repo.Delete(ctx, 99999)
			AssertNoError(t, err, "Delete should not error when no rows affected")
		})

		t.Run("MarkWatched non-existent movie", func(t *testing.T) {
			err := repo.MarkWatched(ctx, 99999)
			AssertError(t, err, "Expected error for non-existent movie")
		})

		t.Run("List with no results", func(t *testing.T) {
			movies, err := repo.List(ctx, MovieListOptions{Year: 1900})
			AssertNoError(t, err, "Should not error when no movies found")
			AssertEqual(t, 0, len(movies), "Expected empty result set")
		})
	})
}
