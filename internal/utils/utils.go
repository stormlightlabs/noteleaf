package utils

import (
	"context"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/stormlightlabs/noteleaf/internal/models"
	"github.com/stormlightlabs/noteleaf/internal/repo"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Logger is the global application logger
var Logger *log.Logger

// NewLogger creates a new logger with the specified level and format
func NewLogger(level string, format string) *log.Logger {
	logger := log.New(os.Stderr)

	switch strings.ToLower(level) {
	case "debug":
		logger.SetLevel(log.DebugLevel)
	case "info":
		logger.SetLevel(log.InfoLevel)
	case "warn", "warning":
		logger.SetLevel(log.WarnLevel)
	case "error":
		logger.SetLevel(log.ErrorLevel)
	default:
		logger.SetLevel(log.InfoLevel)
	}

	// Set format
	if format == "json" {
		logger.SetFormatter(log.JSONFormatter)
	} else {
		logger.SetFormatter(log.TextFormatter)
		logger.SetReportTimestamp(true)
	}

	return logger
}

// GetLogger returns the global logger, creating a default one if it doesn't exist
func GetLogger() *log.Logger {
	if Logger == nil {
		Logger = NewLogger("info", "text")
	}
	return Logger
}

func Titlecase(s string) string {
	return cases.Title(language.Und, cases.NoLower).String(s)
}

// TestTaskRepository interface for dependency injection in tests
type TestTaskRepository interface {
	List(ctx context.Context, options repo.TaskListOptions) ([]*models.Task, error)
	Update(ctx context.Context, task *models.Task) error
}

// TestBookRepository interface for dependency injection in tests
type TestBookRepository interface {
	List(ctx context.Context, options repo.BookListOptions) ([]*models.Book, error)
}

// TestNoteRepository interface for dependency injection in tests
type TestNoteRepository interface {
	List(ctx context.Context, options repo.NoteListOptions) ([]*models.Note, error)
}
