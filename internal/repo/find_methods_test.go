package repo

import (
	"context"
	"testing"
)

func TestFindMethods(t *testing.T) {
	db := CreateTestDB(t)
	repos := SetupTestData(t, db)
	ctx := context.Background()

	t.Run("TaskRepository Find", func(t *testing.T) {
		t.Run("finds tasks by status", func(t *testing.T) {
			options := TaskListOptions{
				Status: "pending",
			}
			tasks, err := repos.Tasks.Find(ctx, options)
			AssertNoError(t, err, "Find should succeed")
			AssertTrue(t, len(tasks) >= 1, "Should find at least one pending task")
			for _, task := range tasks {
				AssertEqual(t, "pending", task.Status, "All returned tasks should be pending")
			}
		})

		t.Run("finds tasks by priority", func(t *testing.T) {
			options := TaskListOptions{
				Priority: "high",
			}
			tasks, err := repos.Tasks.Find(ctx, options)
			AssertNoError(t, err, "Find should succeed")
			AssertTrue(t, len(tasks) >= 1, "Should find at least one high priority task")
			for _, task := range tasks {
				AssertEqual(t, "high", task.Priority, "All returned tasks should be high priority")
			}
		})

		t.Run("finds tasks by project", func(t *testing.T) {
			options := TaskListOptions{
				Project: "test-project",
			}
			tasks, err := repos.Tasks.Find(ctx, options)
			AssertNoError(t, err, "Find should succeed")
			AssertTrue(t, len(tasks) >= 1, "Should find tasks in test-project")
			for _, task := range tasks {
				AssertEqual(t, "test-project", task.Project, "All returned tasks should be in test-project")
			}
		})

		t.Run("finds tasks by context", func(t *testing.T) {
			options := TaskListOptions{
				Context: "test-context",
			}
			tasks, err := repos.Tasks.Find(ctx, options)
			AssertNoError(t, err, "Find should succeed")
			AssertTrue(t, len(tasks) >= 1, "Should find tasks in test-context")
			for _, task := range tasks {
				AssertEqual(t, "test-context", task.Context, "All returned tasks should be in test-context")
			}
		})

		t.Run("finds tasks by multiple criteria", func(t *testing.T) {
			options := TaskListOptions{
				Status:   "pending",
				Priority: "high",
				Project:  "test-project",
			}
			tasks, err := repos.Tasks.Find(ctx, options)
			AssertNoError(t, err, "Find should succeed")
			for _, task := range tasks {
				AssertEqual(t, "pending", task.Status, "Task should be pending")
				AssertEqual(t, "high", task.Priority, "Task should be high priority")
				AssertEqual(t, "test-project", task.Project, "Task should be in test-project")
			}
		})

		t.Run("returns empty for non-matching criteria", func(t *testing.T) {
			options := TaskListOptions{
				Status: "non-existent-status",
			}
			tasks, err := repos.Tasks.Find(ctx, options)
			AssertNoError(t, err, "Find should succeed even with no results")
			AssertEqual(t, 0, len(tasks), "Should find no tasks")
		})

		t.Run("returns all tasks with empty options", func(t *testing.T) {
			options := TaskListOptions{}
			tasks, err := repos.Tasks.Find(ctx, options)
			AssertNoError(t, err, "Find should succeed with empty options")
			AssertTrue(t, len(tasks) >= 2, "Should return all tasks for empty options")
		})
	})

	t.Run("BookRepository Find", func(t *testing.T) {
		t.Run("finds books by status", func(t *testing.T) {
			options := BookListOptions{
				Status: "reading",
			}
			books, err := repos.Books.Find(ctx, options)
			AssertNoError(t, err, "Find should succeed")
			AssertTrue(t, len(books) >= 1, "Should find at least one book being read")
			for _, book := range books {
				AssertEqual(t, "reading", book.Status, "All returned books should be reading")
			}
		})

		t.Run("finds books by author", func(t *testing.T) {
			options := BookListOptions{
				Author: "Test Author",
			}
			books, err := repos.Books.Find(ctx, options)
			AssertNoError(t, err, "Find should succeed")
			AssertTrue(t, len(books) >= 1, "Should find at least one book by Test Author")
			for _, book := range books {
				AssertEqual(t, "Test Author", book.Author, "All returned books should be by Test Author")
			}
		})

		t.Run("finds books by minimum progress", func(t *testing.T) {
			options := BookListOptions{
				MinProgress: 0,
			}
			books, err := repos.Books.Find(ctx, options)
			AssertNoError(t, err, "Find should succeed")
			AssertTrue(t, len(books) >= 1, "Should find books with progress >= 0")
			for _, book := range books {
				AssertTrue(t, book.Progress >= 0, "All returned books should have progress >= 0")
			}
		})

		t.Run("finds books by multiple criteria", func(t *testing.T) {
			options := BookListOptions{
				Status:      "reading",
				Author:      "Test Author",
				MinProgress: 0,
			}
			books, err := repos.Books.Find(ctx, options)
			AssertNoError(t, err, "Find should succeed")
			for _, book := range books {
				AssertEqual(t, "reading", book.Status, "Book should be reading")
				AssertEqual(t, "Test Author", book.Author, "Book should be by Test Author")
				AssertTrue(t, book.Progress >= 0, "Book should have progress >= 0")
			}
		})

		t.Run("returns empty for non-matching criteria", func(t *testing.T) {
			options := BookListOptions{
				Status: "non-existent-status",
			}
			books, err := repos.Books.Find(ctx, options)
			AssertNoError(t, err, "Find should succeed even with no results")
			AssertEqual(t, 0, len(books), "Should find no books")
		})

		t.Run("returns all books with empty options", func(t *testing.T) {
			options := BookListOptions{}
			books, err := repos.Books.Find(ctx, options)
			AssertNoError(t, err, "Find should succeed with empty options")
			AssertTrue(t, len(books) >= 2, "Should return all books for empty options")
		})
	})

	t.Run("MovieRepository Find", func(t *testing.T) {
		t.Run("finds movies by status", func(t *testing.T) {
			options := MovieListOptions{
				Status: "watched",
			}
			movies, err := repos.Movies.Find(ctx, options)
			AssertNoError(t, err, "Find should succeed")
			AssertTrue(t, len(movies) >= 1, "Should find at least one watched movie")
			for _, movie := range movies {
				AssertEqual(t, "watched", movie.Status, "All returned movies should be watched")
			}
		})

		t.Run("finds movies by year", func(t *testing.T) {
			options := MovieListOptions{
				Year: 2023,
			}
			movies, err := repos.Movies.Find(ctx, options)
			AssertNoError(t, err, "Find should succeed")
			AssertTrue(t, len(movies) >= 1, "Should find movies from 2023")
			for _, movie := range movies {
				AssertEqual(t, 2023, movie.Year, "Movie should be from 2023")
			}
		})

		t.Run("finds movies by minimum rating", func(t *testing.T) {
			options := MovieListOptions{
				MinRating: 0.0,
			}
			movies, err := repos.Movies.Find(ctx, options)
			AssertNoError(t, err, "Find should succeed")
			AssertTrue(t, len(movies) >= 1, "Should find movies with rating >= 0")
			for _, movie := range movies {
				AssertTrue(t, movie.Rating >= 0.0, "Movie rating should be >= 0")
			}
		})

		t.Run("finds movies by multiple criteria", func(t *testing.T) {
			options := MovieListOptions{
				Status:    "watched",
				Year:      2023,
				MinRating: 0.0,
			}
			movies, err := repos.Movies.Find(ctx, options)
			AssertNoError(t, err, "Find should succeed")
			for _, movie := range movies {
				AssertEqual(t, "watched", movie.Status, "Movie should be watched")
				AssertEqual(t, 2023, movie.Year, "Movie should be from 2023")
				AssertTrue(t, movie.Rating >= 0.0, "Movie rating should be >= 0")
			}
		})

		t.Run("returns empty for non-matching criteria", func(t *testing.T) {
			options := MovieListOptions{
				Status: "non-existent-status",
			}
			movies, err := repos.Movies.Find(ctx, options)
			AssertNoError(t, err, "Find should succeed even with no results")
			AssertEqual(t, 0, len(movies), "Should find no movies")
		})

		t.Run("returns all movies with empty options", func(t *testing.T) {
			options := MovieListOptions{}
			movies, err := repos.Movies.Find(ctx, options)
			AssertNoError(t, err, "Find should succeed with empty options")
			AssertTrue(t, len(movies) >= 2, "Should return all movies for empty options")
		})
	})

	t.Run("TVRepository Find", func(t *testing.T) {
		t.Run("finds TV shows by status", func(t *testing.T) {
			options := TVListOptions{
				Status: "watching",
			}
			shows, err := repos.TV.Find(ctx, options)
			AssertNoError(t, err, "Find should succeed")
			AssertTrue(t, len(shows) >= 1, "Should find at least one TV show being watched")
			for _, show := range shows {
				AssertEqual(t, "watching", show.Status, "All returned shows should be watching")
			}
		})

		t.Run("finds TV shows by season", func(t *testing.T) {
			options := TVListOptions{
				Season: 1,
			}
			shows, err := repos.TV.Find(ctx, options)
			AssertNoError(t, err, "Find should succeed")
			AssertTrue(t, len(shows) >= 1, "Should find TV shows with season 1")
			for _, show := range shows {
				AssertEqual(t, 1, show.Season, "All returned shows should be season 1")
			}
		})

		t.Run("finds TV shows by minimum rating", func(t *testing.T) {
			options := TVListOptions{
				MinRating: 0.0,
			}
			shows, err := repos.TV.Find(ctx, options)
			AssertNoError(t, err, "Find should succeed")
			AssertTrue(t, len(shows) >= 1, "Should find TV shows with rating >= 0")
			for _, show := range shows {
				AssertTrue(t, show.Rating >= 0.0, "Show rating should be >= 0")
			}
		})

		t.Run("finds TV shows by multiple criteria", func(t *testing.T) {
			options := TVListOptions{
				Status:    "watching",
				Season:    1,
				MinRating: 0.0,
			}
			shows, err := repos.TV.Find(ctx, options)
			AssertNoError(t, err, "Find should succeed")
			for _, show := range shows {
				AssertEqual(t, "watching", show.Status, "Show should be watching")
				AssertEqual(t, 1, show.Season, "Show should be season 1")
				AssertTrue(t, show.Rating >= 0.0, "Show rating should be >= 0")
			}
		})

		t.Run("returns empty for non-matching criteria", func(t *testing.T) {
			options := TVListOptions{
				Status: "non-existent-status",
			}
			shows, err := repos.TV.Find(ctx, options)
			AssertNoError(t, err, "Find should succeed even with no results")
			AssertEqual(t, 0, len(shows), "Should find no TV shows")
		})

		t.Run("returns all TV shows with empty options", func(t *testing.T) {
			options := TVListOptions{}
			shows, err := repos.TV.Find(ctx, options)
			AssertNoError(t, err, "Find should succeed with empty options")
			AssertTrue(t, len(shows) >= 2, "Should return all TV shows for empty options")
		})
	})
}
