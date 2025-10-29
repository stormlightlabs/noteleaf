package handlers

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/stormlightlabs/noteleaf/internal/repo"
	"github.com/stormlightlabs/noteleaf/internal/store"
	"github.com/stormlightlabs/noteleaf/internal/ui"
	"github.com/stormlightlabs/noteleaf/internal/utils"
)

// SeedHandler handles database seeding operations
type SeedHandler struct {
	db     *store.Database
	config *store.Config
	repos  *repo.Repositories
}

// NewSeedHandler creates a new seed handler
func NewSeedHandler() (*SeedHandler, error) {
	db, err := store.NewDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	config, err := store.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	repos := repo.NewRepositories(db.DB)

	return &SeedHandler{
		db:     db,
		config: config,
		repos:  repos,
	}, nil
}

// Close cleans up resources
func (h *SeedHandler) Close() error {
	if h.db != nil {
		return h.db.Close()
	}
	return nil
}

// Seed populates the database with test data for demonstration and testing
func (h *SeedHandler) Seed(ctx context.Context, force bool) error {
	logger := utils.GetLogger()
	logger.Info("Seeding database with test data")

	if force {
		fmt.Println("Clearing existing data...")
		if err := h.clearAllData(); err != nil {
			return fmt.Errorf("failed to clear existing data: %w", err)
		}
	}

	fmt.Println("Seeding database with test data...")

	tasks := []struct {
		description string
		project     string
		priority    string
		status      string
	}{
		{"Review quarterly report", "work", "high", "pending"},
		{"Plan vacation itinerary", "personal", "medium", "pending"},
		{"Fix bug in user authentication", "development", "high", "pending"},
		{"Read \"Clean Code\" book", "learning", "low", "pending"},
		{"Update project documentation", "work", "medium", "completed"},
	}

	for _, task := range tasks {
		if err := h.seedTask(task.description, task.project, task.priority, task.status); err != nil {
			logger.Warn("Failed to seed task", "description", task.description, "error", err)
		}
	}

	books := []struct {
		title    string
		author   string
		status   string
		progress int
	}{
		{"The Go Programming Language", "Alan Donovan", "reading", 45},
		{"Clean Code", "Robert Martin", "queued", 0},
		{"Design Patterns", "Gang of Four", "finished", 100},
		{"The Pragmatic Programmer", "Andy Hunt", "queued", 0},
		{"Effective Go", "Various", "reading", 75},
	}

	for _, book := range books {
		if err := h.seedBook(book.title, book.author, book.status, book.progress); err != nil {
			logger.Warn("Failed to seed book", "title", book.title, "error", err)
		}
	}

	fmt.Printf("Successfully seeded database with %d tasks and %d books\n", len(tasks), len(books))
	fmt.Printf("\n%s\n", ui.InfoStyle.Render("Example commands to try:"))
	fmt.Printf("  %s\n", ui.SuccessStyle.Render("noteleaf todo list"))
	fmt.Printf("  %s\n", ui.SuccessStyle.Render("noteleaf media book list"))
	fmt.Printf("  %s\n", ui.SuccessStyle.Render("noteleaf todo view 1"))

	return nil
}

func (h *SeedHandler) clearAllData() error {
	queries := []string{
		"DELETE FROM tasks",
		"DELETE FROM books",
		"DELETE FROM notes",
		"DELETE FROM movies",
		"DELETE FROM tv_shows",
		"DELETE FROM sqlite_sequence WHERE name IN ('tasks', 'books', 'notes', 'movies', 'tv_shows')",
	}

	for _, query := range queries {
		if _, err := h.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute %s: %w", query, err)
		}
	}

	return nil
}

func (h *SeedHandler) seedTask(description, project, priority, status string) error {
	uuid := h.generateSimpleUUID()
	query := `INSERT INTO tasks (uuid, description, project, priority, status, entry, modified)
			  VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now'))`
	_, err := h.db.Exec(query, uuid, description, project, priority, status)
	return err
}

func (h *SeedHandler) seedBook(title, author, status string, progress int) error {
	query := `INSERT INTO books (title, author, status, progress, added)
			  VALUES (?, ?, ?, ?, datetime('now'))`
	_, err := h.db.Exec(query, title, author, status, progress)
	return err
}

// generateSimpleUUID creates a simple UUID for seeding (not cryptographically secure, but sufficient for test data)
func (h *SeedHandler) generateSimpleUUID() string {
	now := time.Now()
	randomNum := rand.Intn(10000)
	return fmt.Sprintf("seed-task-%d-%d-%d", now.Unix(), now.UnixNano()%1000000, randomNum)
}
