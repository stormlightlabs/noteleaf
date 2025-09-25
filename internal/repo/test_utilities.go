package repo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jaswdr/faker/v2"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stormlightlabs/noteleaf/internal/models"
)

var fake = faker.New()

// CreateTestDB creates an in-memory SQLite database with the full schema for testing
func CreateTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create in-memory database: %v", err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	// Full schema for all tables
	schema := `
		CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid TEXT UNIQUE NOT NULL,
			description TEXT NOT NULL,
			status TEXT DEFAULT 'pending',
			priority TEXT,
			project TEXT,
			context TEXT,
			tags TEXT,
			due DATETIME,
			entry DATETIME DEFAULT CURRENT_TIMESTAMP,
			modified DATETIME DEFAULT CURRENT_TIMESTAMP,
			end DATETIME,
			start DATETIME,
			annotations TEXT
		);

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

		CREATE TABLE IF NOT EXISTS notes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			content TEXT,
			tags TEXT,
			archived BOOLEAN DEFAULT FALSE,
			created DATETIME DEFAULT CURRENT_TIMESTAMP,
			modified DATETIME DEFAULT CURRENT_TIMESTAMP,
			file_path TEXT
		);

		CREATE TABLE IF NOT EXISTS time_entries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id INTEGER NOT NULL,
			start_time DATETIME NOT NULL,
			end_time DATETIME,
			duration_seconds INTEGER,
			created DATETIME DEFAULT CURRENT_TIMESTAMP,
			modified DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS articles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			url TEXT UNIQUE NOT NULL,
			title TEXT NOT NULL,
			author TEXT,
			date TEXT,
			markdown_path TEXT NOT NULL,
			html_path TEXT NOT NULL,
			created DATETIME DEFAULT CURRENT_TIMESTAMP,
			modified DATETIME DEFAULT CURRENT_TIMESTAMP
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

// Sample data creators
func CreateSampleTask() *models.Task {
	return &models.Task{
		UUID:        uuid.New().String(),
		Description: "Test Task",
		Status:      "pending",
		Priority:    "medium",
		Project:     "test-project",
		Context:     "test-context",
		Tags:        []string{"test", "sample"},
		Entry:       time.Now(),
		Modified:    time.Now(),
	}
}

func CreateSampleBook() *models.Book {
	return &models.Book{
		Title:    "Test Book",
		Author:   "Test Author",
		Status:   "queued",
		Progress: 0,
		Pages:    300,
		Rating:   4.5,
		Notes:    "Great book!",
		Added:    time.Now(),
	}
}

func CreateSampleMovie() *models.Movie {
	return &models.Movie{
		Title:  "Test Movie",
		Year:   2023,
		Status: "queued",
		Rating: 8.5,
		Notes:  "Excellent film",
		Added:  time.Now(),
	}
}

func CreateSampleTVShow() *models.TVShow {
	return &models.TVShow{
		Title:   "Test TV Show",
		Season:  1,
		Episode: 1,
		Status:  "queued",
		Rating:  9.0,
		Notes:   "Amazing series",
		Added:   time.Now(),
	}
}

func CreateSampleNote() *models.Note {
	return &models.Note{
		Title:    "Test Note",
		Content:  "This is a test note content",
		Tags:     []string{"test", "sample"},
		Archived: false,
		Created:  time.Now(),
		Modified: time.Now(),
	}
}

func CreateSampleTimeEntry(taskID int64) *models.TimeEntry {
	startTime := time.Now().Add(-time.Hour)
	return &models.TimeEntry{
		TaskID:          taskID,
		StartTime:       startTime,
		EndTime:         nil,
		DurationSeconds: 0,
		Created:         startTime,
		Modified:        startTime,
	}
}

func CreateSampleArticle() *models.Article {
	return &models.Article{
		URL:          "https://example.com/test-article",
		Title:        "Test Article",
		Author:       "Test Author",
		Date:         "2024-01-01",
		MarkdownPath: "/path/test-article.md",
		HTMLPath:     "/path/test-article.html",
		Created:      time.Now(),
		Modified:     time.Now(),
	}
}

func fakeHTMLFile(f faker.Faker) string {
	original := f.File().AbsoluteFilePath(2)
	split := strings.Split(original, ".")
	split[len(split)-1] = "html"

	return strings.Join(split, ".")
}

func fakeMDFile(f faker.Faker) string {
	original := f.File().AbsoluteFilePath(2)
	split := strings.Split(original, ".")
	split[len(split)-1] = "md"

	return strings.Join(split, ".")
}

func FakeTime(f faker.Faker) time.Time {
	return f.Time().Time(time.Now())
}

func CreateFakeArticle() *models.Article {
	return &models.Article{
		URL:          fake.Internet().URL(),
		Title:        strings.Join(fake.Lorem().Words(3), " "),
		Author:       fmt.Sprintf("%v %v", fake.Person().FirstName(), fake.Person().LastName()),
		Date:         fake.Time().Time(time.Now()).Format("2006-01-02"),
		MarkdownPath: fakeMDFile(fake),
		HTMLPath:     fakeHTMLFile(fake),
		Created:      time.Now(),
		Modified:     time.Now(),
	}
}

func CreateFakeArticles(count int) []*models.Article {
	articles := make([]*models.Article, count)
	for i := range count {
		articles[i] = CreateFakeArticle()
	}

	return articles
}

// Test helpers for common operations
func AssertNoError(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", msg, err)
	}
}

func AssertError(t *testing.T, err error, msg string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: expected error but got none", msg)
	}
}

func AssertEqual[T comparable](t *testing.T, expected, actual T, msg string) {
	t.Helper()
	if expected != actual {
		t.Fatalf("%s: expected %v, got %v", msg, expected, actual)
	}
}

func AssertNotEqual[T comparable](t *testing.T, notExpected, actual T, msg string) {
	t.Helper()
	if notExpected == actual {
		t.Fatalf("%s: expected value to not equal %v", msg, notExpected)
	}
}

func AssertTrue(t *testing.T, condition bool, msg string) {
	t.Helper()
	if !condition {
		t.Fatalf("%s: expected true", msg)
	}
}

func AssertFalse(t *testing.T, condition bool, msg string) {
	t.Helper()
	if condition {
		t.Fatalf("%s: expected false", msg)
	}
}

func AssertContains(t *testing.T, str, substr, msg string) {
	t.Helper()
	if !strings.Contains(str, substr) {
		t.Fatalf("%s: expected string '%s' to contain '%s'", msg, str, substr)
	}
}

// SetupTestData creates sample data in the database and returns the repositories
func SetupTestData(t *testing.T, db *sql.DB) *Repositories {
	ctx := context.Background()
	repos := NewRepositories(db)

	// Create sample tasks
	task1 := CreateSampleTask()
	task1.Description = "Sample Task 1"
	task1.Status = "pending"
	task1.Priority = "high"

	task2 := CreateSampleTask()
	task2.Description = "Sample Task 2"
	task2.Status = "completed"
	task2.Priority = "low"

	id1, err := repos.Tasks.Create(ctx, task1)
	AssertNoError(t, err, "Failed to create sample task 1")
	task1.ID = id1

	id2, err := repos.Tasks.Create(ctx, task2)
	AssertNoError(t, err, "Failed to create sample task 2")
	task2.ID = id2

	// Create sample books
	book1 := CreateSampleBook()
	book1.Title = "Sample Book 1"
	book1.Status = "reading"

	book2 := CreateSampleBook()
	book2.Title = "Sample Book 2"
	book2.Status = "finished"

	bookID1, err := repos.Books.Create(ctx, book1)
	AssertNoError(t, err, "Failed to create sample book 1")
	book1.ID = bookID1

	bookID2, err := repos.Books.Create(ctx, book2)
	AssertNoError(t, err, "Failed to create sample book 2")
	book2.ID = bookID2

	// Create sample movies
	movie1 := CreateSampleMovie()
	movie1.Title = "Sample Movie 1"
	movie1.Status = "queued"

	movie2 := CreateSampleMovie()
	movie2.Title = "Sample Movie 2"
	movie2.Status = "watched"

	movieID1, err := repos.Movies.Create(ctx, movie1)
	AssertNoError(t, err, "Failed to create sample movie 1")
	movie1.ID = movieID1

	movieID2, err := repos.Movies.Create(ctx, movie2)
	AssertNoError(t, err, "Failed to create sample movie 2")
	movie2.ID = movieID2

	// Create sample TV shows
	tv1 := CreateSampleTVShow()
	tv1.Title = "Sample TV Show 1"
	tv1.Status = "queued"

	tv2 := CreateSampleTVShow()
	tv2.Title = "Sample TV Show 2"
	tv2.Status = "watching"

	tvID1, err := repos.TV.Create(ctx, tv1)
	AssertNoError(t, err, "Failed to create sample TV show 1")
	tv1.ID = tvID1

	tvID2, err := repos.TV.Create(ctx, tv2)
	AssertNoError(t, err, "Failed to create sample TV show 2")
	tv2.ID = tvID2

	// Create sample notes
	note1 := CreateSampleNote()
	note1.Title = "Sample Note 1"
	note1.Content = "Content for note 1"

	note2 := CreateSampleNote()
	note2.Title = "Sample Note 2"
	note2.Content = "Content for note 2"
	note2.Archived = true

	noteID1, err := repos.Notes.Create(ctx, note1)
	AssertNoError(t, err, "Failed to create sample note 1")
	note1.ID = noteID1

	noteID2, err := repos.Notes.Create(ctx, note2)
	AssertNoError(t, err, "Failed to create sample note 2")
	note2.ID = noteID2

	return repos
}
