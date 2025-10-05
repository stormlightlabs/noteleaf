package handlers

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stormlightlabs/noteleaf/internal/repo"
	"github.com/stormlightlabs/noteleaf/internal/store"
)

// HandlerTestSuite provides reusable test infrastructure for handlers
type HandlerTestSuite struct {
	t       *testing.T
	tempDir string
	ctx     context.Context
	cleanup func()
}

// NewHandlerTestSuite creates a new test suite with isolated environment
func NewHandlerTestSuite(t *testing.T) *HandlerTestSuite {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "noteleaf-handler-suite-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	oldNoteleafConfig := os.Getenv("NOTELEAF_CONFIG")
	oldNoteleafDataDir := os.Getenv("NOTELEAF_DATA_DIR")
	os.Setenv("NOTELEAF_CONFIG", filepath.Join(tempDir, ".noteleaf.conf.toml"))
	os.Setenv("NOTELEAF_DATA_DIR", tempDir)

	cleanup := func() {
		os.Setenv("NOTELEAF_CONFIG", oldNoteleafConfig)
		os.Setenv("NOTELEAF_DATA_DIR", oldNoteleafDataDir)
		os.RemoveAll(tempDir)
	}

	ctx := context.Background()
	if err := Setup(ctx, []string{}); err != nil {
		cleanup()
		t.Fatalf("Failed to setup database: %v", err)
	}

	suite := &HandlerTestSuite{
		t:       t,
		tempDir: tempDir,
		ctx:     ctx,
		cleanup: cleanup,
	}

	t.Cleanup(suite.Cleanup)

	return suite
}

// Context returns the test context
func (s *HandlerTestSuite) Context() context.Context {
	return s.ctx
}

// TempDir returns the temporary directory for this test
func (s *HandlerTestSuite) TempDir() string {
	return s.tempDir
}

// Cleanup performs test cleanup
func (s *HandlerTestSuite) Cleanup() {
	if s.cleanup != nil {
		s.cleanup()
	}
}

// AssertNoError fails the test if err is not nil
func (s *HandlerTestSuite) AssertNoError(err error, msg string) {
	s.t.Helper()
	if err != nil {
		s.t.Fatalf("%s: unexpected error: %v", msg, err)
	}
}

// AssertError fails the test if err is nil
func (s *HandlerTestSuite) AssertError(err error, msg string) {
	s.t.Helper()
	if err == nil {
		s.t.Fatalf("%s: expected error but got nil", msg)
	}
}

// CreateTestDatabase creates an isolated test database
func (s *HandlerTestSuite) CreateTestDatabase() (*store.Database, *repo.Repositories, error) {
	db, err := store.NewDatabase()
	if err != nil {
		return nil, nil, err
	}

	repos := repo.NewRepositories(db.DB)
	return db, repos, nil
}

// HandlerLifecycleTest tests basic handler lifecycle (create, close)
//
// This is a reusable test pattern for any handler implementing Closeable
func HandlerLifecycleTest[H Closeable](t *testing.T, createHandler func() (H, error)) {
	t.Helper()

	t.Run("creates handler successfully", func(t *testing.T) {
		handler, err := createHandler()
		if err != nil {
			t.Fatalf("Failed to create handler: %v", err)
		}

		if err := handler.Close(); err != nil {
			t.Errorf("Close failed: %v", err)
		}
	})

	t.Run("handles close gracefully", func(t *testing.T) {
		handler, err := createHandler()
		if err != nil {
			t.Fatalf("Failed to create handler: %v", err)
		}

		if err := handler.Close(); err != nil {
			t.Errorf("First close should succeed: %v", err)
		}

		_ = handler.Close()
	})
}

// ViewableTest tests View functionality for handlers
func ViewableTest[H Viewable](t *testing.T, handler H, validID, invalidID int64) {
	t.Helper()
	ctx := context.Background()

	t.Run("views valid item", func(t *testing.T) {
		err := handler.View(ctx, validID)
		if err != nil {
			t.Errorf("View with valid ID should succeed: %v", err)
		}
	})

	t.Run("fails with invalid ID", func(t *testing.T) {
		err := handler.View(ctx, invalidID)
		if err == nil {
			t.Error("View with invalid ID should fail")
		}
	})
}

// InputReaderTest tests SetInputReader functionality
func InputReaderTest[H InputReader](t *testing.T, handler H) {
	t.Helper()

	t.Run("sets input reader", func(t *testing.T) {
		sim := NewInputSimulator("test")
		handler.SetInputReader(sim)
		// If we get here without panic, the test passes
	})

	t.Run("handles nil reader", func(t *testing.T) {
		handler.SetInputReader(nil)
		// If we get here without panic, the test passes
	})
}

// MediaHandlerTestSuite provides specialized test utilities for media handlers
type MediaHandlerTestSuite struct {
	*HandlerTestSuite
}

// NewMediaHandlerTestSuite creates a test suite for media handlers
func NewMediaHandlerTestSuite(t *testing.T) *MediaHandlerTestSuite {
	return &MediaHandlerTestSuite{
		HandlerTestSuite: NewHandlerTestSuite(t),
	}
}

// TestSearchAndAdd is a reusable test pattern for SearchAndAdd operations
func (s *MediaHandlerTestSuite) TestSearchAndAdd(handler MediaHandler, query string, shouldSucceed bool) {
	s.t.Helper()

	err := handler.SearchAndAdd(s.ctx, query, false)
	if shouldSucceed && err != nil {
		s.t.Errorf("SearchAndAdd should succeed: %v", err)
	}
	if !shouldSucceed && err == nil {
		s.t.Error("SearchAndAdd should fail")
	}
}

// TestList is a reusable test pattern for List operations
func (s *MediaHandlerTestSuite) TestList(handler MediaHandler, status string) {
	s.t.Helper()

	err := handler.List(s.ctx, status)
	if err != nil {
		s.t.Errorf("List should succeed: %v", err)
	}
}

// TestUpdateStatus is a reusable test pattern for UpdateStatus operations
func (s *MediaHandlerTestSuite) TestUpdateStatus(handler MediaHandler, id, status string, shouldSucceed bool) {
	s.t.Helper()

	err := handler.UpdateStatus(s.ctx, id, status)
	if shouldSucceed && err != nil {
		s.t.Errorf("UpdateStatus should succeed: %v", err)
	}
	if !shouldSucceed && err == nil {
		s.t.Error("UpdateStatus should fail")
	}
}

// TestRemove is a reusable test pattern for Remove operations
func (s *MediaHandlerTestSuite) TestRemove(handler MediaHandler, id string, shouldSucceed bool) {
	s.t.Helper()

	err := handler.Remove(s.ctx, id)
	if shouldSucceed && err != nil {
		s.t.Errorf("Remove should succeed: %v", err)
	}
	if !shouldSucceed && err == nil {
		s.t.Error("Remove should fail")
	}
}

// CreateHandler creates a handler with automatic cleanup
func CreateHandler[H Closeable](t *testing.T, factory func() (H, error)) H {
	t.Helper()

	handler, err := factory()
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	t.Cleanup(func() {
		handler.Close()
	})

	return handler
}

// AssertExists verifies that an item exists using a getter
func AssertExists[T any](t *testing.T, getter func(context.Context, int64) (*T, error), id int64, itemType string) {
	t.Helper()

	ctx := context.Background()
	_, err := getter(ctx, id)
	if err != nil {
		t.Errorf("%s %d should exist but got error: %v", itemType, id, err)
	}
}

// AssertNotExists verifies that an item does not exist using a getter
func AssertNotExists[T any](t *testing.T, getter func(context.Context, int64) (*T, error), id int64, itemType string) {
	t.Helper()

	ctx := context.Background()
	_, err := getter(ctx, id)
	if err == nil {
		t.Errorf("%s %d should not exist but was found", itemType, id)
	}
}
