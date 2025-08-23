package repo

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"stormlightlabs.org/noteleaf/internal/models"
)

func createFullTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create in-memory database: %v", err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	// Create all tables
	schema := `
		-- Tasks table
		CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid TEXT UNIQUE NOT NULL,
			description TEXT NOT NULL,
			status TEXT DEFAULT 'pending',
			priority TEXT,
			project TEXT,
			tags TEXT,
			due DATETIME,
			entry DATETIME DEFAULT CURRENT_TIMESTAMP,
			modified DATETIME DEFAULT CURRENT_TIMESTAMP,
			end DATETIME,
			start DATETIME,
			annotations TEXT
		);

		-- Movies table
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

		-- TV Shows table
		CREATE TABLE IF NOT EXISTS tv_shows (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			season INTEGER,
			episode INTEGER,
			status TEXT DEFAULT 'queued',
			rating REAL,
			notes TEXT,
			added DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_watched DATETIME
		);

		-- Books table
		CREATE TABLE IF NOT EXISTS books (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			author TEXT,
			status TEXT DEFAULT 'queued',
			progress INTEGER DEFAULT 0,
			pages INTEGER,
			rating REAL,
			notes TEXT,
			added DATETIME DEFAULT CURRENT_TIMESTAMP,
			started DATETIME,
			finished DATETIME
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

func TestRepositories(t *testing.T) {
	t.Run("Integration", func(t *testing.T) {
		db := createFullTestDB(t)
		repos := NewRepositories(db)
		ctx := context.Background()

		t.Run("Create all resource types", func(t *testing.T) {
			task := &models.Task{
				UUID:        uuid.New().String(),
				Description: "Integration test task",
				Status:      "pending",
				Project:     "integration",
			}
			taskID, err := repos.Tasks.Create(ctx, task)
			if err != nil {
				t.Errorf("Failed to create task: %v", err)
			}
			if taskID == 0 {
				t.Error("Expected non-zero task ID")
			}

			movie := &models.Movie{
				Title:  "Integration Movie",
				Year:   2023,
				Status: "queued",
				Rating: 8.5,
			}
			movieID, err := repos.Movies.Create(ctx, movie)
			if err != nil {
				t.Errorf("Failed to create movie: %v", err)
			}
			if movieID == 0 {
				t.Error("Expected non-zero movie ID")
			}

			tvShow := &models.TVShow{
				Title:   "Integration Series",
				Season:  1,
				Episode: 1,
				Status:  "queued",
				Rating:  9.0,
			}
			tvID, err := repos.TV.Create(ctx, tvShow)
			if err != nil {
				t.Errorf("Failed to create TV show: %v", err)
			}
			if tvID == 0 {
				t.Error("Expected non-zero TV show ID")
			}

			book := &models.Book{
				Title:    "Integration Book",
				Author:   "Test Author",
				Status:   "queued",
				Progress: 0,
				Pages:    300,
			}
			bookID, err := repos.Books.Create(ctx, book)
			if err != nil {
				t.Errorf("Failed to create book: %v", err)
			}
			if bookID == 0 {
				t.Error("Expected non-zero book ID")
			}
		})

		t.Run("Retrieve all resources", func(t *testing.T) {
			tasks, err := repos.Tasks.List(ctx, TaskListOptions{})
			if err != nil {
				t.Errorf("Failed to list tasks: %v", err)
			}
			if len(tasks) != 1 {
				t.Errorf("Expected 1 task, got %d", len(tasks))
			}

			movies, err := repos.Movies.List(ctx, MovieListOptions{})
			if err != nil {
				t.Errorf("Failed to list movies: %v", err)
			}
			if len(movies) != 1 {
				t.Errorf("Expected 1 movie, got %d", len(movies))
			}

			tvShows, err := repos.TV.List(ctx, TVListOptions{})
			if err != nil {
				t.Errorf("Failed to list TV shows: %v", err)
			}
			if len(tvShows) != 1 {
				t.Errorf("Expected 1 TV show, got %d", len(tvShows))
			}

			books, err := repos.Books.List(ctx, BookListOptions{})
			if err != nil {
				t.Errorf("Failed to list books: %v", err)
			}
			if len(books) != 1 {
				t.Errorf("Expected 1 book, got %d", len(books))
			}
		})

		t.Run("Count all resources", func(t *testing.T) {
			taskCount, err := repos.Tasks.Count(ctx, TaskListOptions{})
			if err != nil {
				t.Errorf("Failed to count tasks: %v", err)
			}
			if taskCount != 1 {
				t.Errorf("Expected 1 task, got %d", taskCount)
			}

			movieCount, err := repos.Movies.Count(ctx, MovieListOptions{})
			if err != nil {
				t.Errorf("Failed to count movies: %v", err)
			}
			if movieCount != 1 {
				t.Errorf("Expected 1 movie, got %d", movieCount)
			}

			tvCount, err := repos.TV.Count(ctx, TVListOptions{})
			if err != nil {
				t.Errorf("Failed to count TV shows: %v", err)
			}
			if tvCount != 1 {
				t.Errorf("Expected 1 TV show, got %d", tvCount)
			}

			bookCount, err := repos.Books.Count(ctx, BookListOptions{})
			if err != nil {
				t.Errorf("Failed to count books: %v", err)
			}
			if bookCount != 1 {
				t.Errorf("Expected 1 book, got %d", bookCount)
			}
		})

		t.Run("Use specialized methods", func(t *testing.T) {
			pendingTasks, err := repos.Tasks.GetPending(ctx)
			if err != nil {
				t.Errorf("Failed to get pending tasks: %v", err)
			}
			if len(pendingTasks) != 1 {
				t.Errorf("Expected 1 pending task, got %d", len(pendingTasks))
			}

			queuedMovies, err := repos.Movies.GetQueued(ctx)
			if err != nil {
				t.Errorf("Failed to get queued movies: %v", err)
			}
			if len(queuedMovies) != 1 {
				t.Errorf("Expected 1 queued movie, got %d", len(queuedMovies))
			}

			queuedTV, err := repos.TV.GetQueued(ctx)
			if err != nil {
				t.Errorf("Failed to get queued TV shows: %v", err)
			}
			if len(queuedTV) != 1 {
				t.Errorf("Expected 1 queued TV show, got %d", len(queuedTV))
			}

			queuedBooks, err := repos.Books.GetQueued(ctx)
			if err != nil {
				t.Errorf("Failed to get queued books: %v", err)
			}
			if len(queuedBooks) != 1 {
				t.Errorf("Expected 1 queued book, got %d", len(queuedBooks))
			}
		})
	})

	t.Run("New", func(t *testing.T) {
		db := createFullTestDB(t)
		repos := NewRepositories(db)

		t.Run("All repositories are initialized", func(t *testing.T) {
			if repos.Tasks == nil {
				t.Error("Tasks repository should be initialized")
			}
			if repos.Movies == nil {
				t.Error("Movies repository should be initialized")
			}
			if repos.TV == nil {
				t.Error("TV repository should be initialized")
			}
			if repos.Books == nil {
				t.Error("Books repository should be initialized")
			}
		})

	})
}
