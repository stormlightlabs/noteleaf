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
	"github.com/stormlightlabs/noteleaf/internal/store"
)

var fake = faker.New()

// CreateTestDB creates an in-memory SQLite database with the full schema for testing
func CreateTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create in-memory database: %v", err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	// if _, err := db.Exec(testSchema); err != nil {
	// 	t.Fatalf("Failed to create schema: %v", err)
	// }

	mr := store.NewMigrationRunner(&store.Database{DB: db})
	if err := mr.RunMigrations(); err != nil {
		t.Errorf("failed to run migrations %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

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

func AssertNil(t *testing.T, value interface{}, msg string) {
	t.Helper()
	if value != nil {
		t.Fatalf("%s: expected nil, got %v", msg, value)
	}
}

func AssertNotNil(t *testing.T, value interface{}, msg string) {
	t.Helper()
	if value == nil {
		t.Fatalf("%s: expected non-nil value", msg)
	}
}

func AssertGreaterThan[T interface{ int | int64 | float64 }](t *testing.T, actual, threshold T, msg string) {
	t.Helper()
	if actual <= threshold {
		t.Fatalf("%s: expected %v > %v", msg, actual, threshold)
	}
}

func AssertLessThan[T interface{ int | int64 | float64 }](t *testing.T, actual, threshold T, msg string) {
	t.Helper()
	if actual >= threshold {
		t.Fatalf("%s: expected %v < %v", msg, actual, threshold)
	}
}

func AssertStringContains(t *testing.T, str, substr, msg string) {
	t.Helper()
	if !strings.Contains(str, substr) {
		t.Fatalf("%s: expected string to contain '%s', got '%s'", msg, substr, str)
	}
}

// NewCanceledContext returns a pre-canceled context for testing error conditions
func NewCanceledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

// TaskBuilder provides a fluent interface for building test tasks
type TaskBuilder struct {
	task *models.Task
}

// NewTaskBuilder creates a new TaskBuilder with sensible defaults
func NewTaskBuilder() *TaskBuilder {
	return &TaskBuilder{
		task: &models.Task{
			UUID:     uuid.New().String(),
			Status:   "pending",
			Entry:    time.Now(),
			Modified: time.Now(),
		},
	}
}

func (b *TaskBuilder) WithUUID(uuid string) *TaskBuilder {
	b.task.UUID = uuid
	return b
}

func (b *TaskBuilder) WithDescription(desc string) *TaskBuilder {
	b.task.Description = desc
	return b
}

func (b *TaskBuilder) WithStatus(status string) *TaskBuilder {
	b.task.Status = status
	return b
}

func (b *TaskBuilder) WithPriority(priority string) *TaskBuilder {
	b.task.Priority = priority
	return b
}

func (b *TaskBuilder) WithProject(project string) *TaskBuilder {
	b.task.Project = project
	return b
}

func (b *TaskBuilder) WithContext(ctx string) *TaskBuilder {
	b.task.Context = ctx
	return b
}

func (b *TaskBuilder) WithTags(tags []string) *TaskBuilder {
	b.task.Tags = tags
	return b
}

func (b *TaskBuilder) WithDue(due time.Time) *TaskBuilder {
	b.task.Due = &due
	return b
}

func (b *TaskBuilder) WithEnd(end time.Time) *TaskBuilder {
	b.task.End = &end
	return b
}

func (b *TaskBuilder) WithRecur(recur string) *TaskBuilder {
	b.task.Recur = models.RRule(recur)
	return b
}

func (b *TaskBuilder) WithDependsOn(deps []string) *TaskBuilder {
	b.task.DependsOn = deps
	return b
}

func (b *TaskBuilder) Build() *models.Task {
	return b.task
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
