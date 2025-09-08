package handlers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stormlightlabs/noteleaf/internal/models"
	"github.com/stormlightlabs/noteleaf/internal/store"
)

// HandlerTestHelper wraps NoteHandler with test-specific functionality
type HandlerTestHelper struct {
	*NoteHandler
	tempDir string
	cleanup func()
}

// NewHandlerTestHelper creates a NoteHandler with isolated test database
func NewHandlerTestHelper(t *testing.T) *HandlerTestHelper {
	tempDir, err := os.MkdirTemp("", "noteleaf-handler-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	oldConfigHome := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tempDir)

	cleanup := func() {
		os.Setenv("XDG_CONFIG_HOME", oldConfigHome)
		os.RemoveAll(tempDir)
	}

	ctx := context.Background()
	err = Setup(ctx, []string{})
	if err != nil {
		cleanup()
		t.Fatalf("Failed to setup database: %v", err)
	}

	handler, err := NewNoteHandler()
	if err != nil {
		cleanup()
		t.Fatalf("Failed to create note handler: %v", err)
	}

	testHandler := &HandlerTestHelper{
		NoteHandler: handler,
		tempDir:     tempDir,
		cleanup:     cleanup,
	}

	t.Cleanup(func() {
		testHandler.Close()
		testHandler.cleanup()
	})

	return testHandler
}

// CreateTestNote creates a test note and returns its ID
func (th *HandlerTestHelper) CreateTestNote(t *testing.T, title, content string, tags []string) int64 {
	ctx := context.Background()
	note := &models.Note{
		Title:    title,
		Content:  content,
		Tags:     tags,
		Created:  time.Now(),
		Modified: time.Now(),
	}

	id, err := th.repos.Notes.Create(ctx, note)
	if err != nil {
		t.Fatalf("Failed to create test note: %v", err)
	}
	return id
}

// CreateTestFile creates a temporary markdown file with content
func (th *HandlerTestHelper) CreateTestFile(t *testing.T, filename, content string) string {
	filePath := filepath.Join(th.tempDir, filename)
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	return filePath
}

// MockEditor provides a mock editor function for testing
type MockEditor struct {
	shouldFail     bool
	failureMsg     string
	contentToWrite string
	deleteFile     bool
	makeReadOnly   bool
}

// NewMockEditor creates a mock editor with default success behavior
func NewMockEditor() *MockEditor {
	return &MockEditor{
		contentToWrite: `# Test Note

Test content here.

<!-- Tags: test -->`,
	}
}

// WithFailure configures the mock editor to fail
func (me *MockEditor) WithFailure(msg string) *MockEditor {
	me.shouldFail = true
	me.failureMsg = msg
	return me
}

// WithContent configures the content the mock editor will write
func (me *MockEditor) WithContent(content string) *MockEditor {
	me.contentToWrite = content
	return me
}

// WithFileDeleted configures the mock editor to delete the temp file
func (me *MockEditor) WithFileDeleted() *MockEditor {
	me.deleteFile = true
	return me
}

// WithReadOnly configures the mock editor to make the file read-only
func (me *MockEditor) WithReadOnly() *MockEditor {
	me.makeReadOnly = true
	return me
}

// GetEditorFunc returns the editor function for use with NoteHandler
func (me *MockEditor) GetEditorFunc() editorFunc {
	return func(editor, filePath string) error {
		if me.shouldFail {
			return fmt.Errorf("%s", me.failureMsg)
		}

		if me.deleteFile {
			return os.Remove(filePath)
		}

		if me.makeReadOnly {
			os.Chmod(filePath, 0444)
			return nil
		}

		return os.WriteFile(filePath, []byte(me.contentToWrite), 0644)
	}
}

// DatabaseTestHelper provides database testing utilities
type DatabaseTestHelper struct {
	originalDB *store.Database
	handler    *HandlerTestHelper
}

// NewDatabaseTestHelper creates a helper for database error testing
func NewDatabaseTestHelper(handler *HandlerTestHelper) *DatabaseTestHelper {
	return &DatabaseTestHelper{
		originalDB: handler.db,
		handler:    handler,
	}
}

// CloseDatabase closes the database connection
func (dth *DatabaseTestHelper) CloseDatabase() {
	dth.handler.db.Close()
}

// RestoreDatabase restores the original database connection
func (dth *DatabaseTestHelper) RestoreDatabase(t *testing.T) {
	var err error
	dth.handler.db, err = store.NewDatabase()
	if err != nil {
		t.Fatalf("Failed to restore database: %v", err)
	}
}

// DropNotesTable drops the notes table to simulate database errors
func (dth *DatabaseTestHelper) DropNotesTable() {
	dth.handler.db.Exec("DROP TABLE notes")
}

// CreateCorruptedDatabase creates a new database with corrupted schema
func (dth *DatabaseTestHelper) CreateCorruptedDatabase(t *testing.T) {
	dth.CloseDatabase()

	db, err := store.NewDatabase()
	if err != nil {
		t.Fatalf("Failed to create corrupted database: %v", err)
	}

	// Drop the notes table to simulate corruption
	db.Exec("DROP TABLE notes")
	dth.handler.db = db
}

// AssertionHelpers provides test assertion utilities
type AssertionHelpers struct{}

// AssertError checks that an error occurred and optionally contains expected text
func (ah AssertionHelpers) AssertError(t *testing.T, err error, expectedSubstring string, msg string) {
	t.Helper()
	if err == nil {
		t.Errorf("%s: expected error but got none", msg)
		return
	}
	if expectedSubstring != "" && !containsString(err.Error(), expectedSubstring) {
		t.Errorf("%s: expected error containing %q, got: %v", msg, expectedSubstring, err)
	}
}

// AssertNoError checks that no error occurred
func (ah AssertionHelpers) AssertNoError(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Errorf("%s: unexpected error: %v", msg, err)
	}
}

// AssertNoteExists checks that a note exists in the database
func (ah AssertionHelpers) AssertNoteExists(t *testing.T, handler *HandlerTestHelper, id int64) {
	t.Helper()
	ctx := context.Background()
	_, err := handler.repos.Notes.Get(ctx, id)
	if err != nil {
		t.Errorf("Note %d should exist but got error: %v", id, err)
	}
}

// AssertNoteNotExists checks that a note does not exist in the database
func (ah AssertionHelpers) AssertNoteNotExists(t *testing.T, handler *HandlerTestHelper, id int64) {
	t.Helper()
	ctx := context.Background()
	_, err := handler.repos.Notes.Get(ctx, id)
	if err == nil {
		t.Errorf("Note %d should not exist but was found", id)
	}
}

// EnvironmentTestHelper provides environment manipulation utilities for testing
type EnvironmentTestHelper struct {
	originalVars map[string]string
}

// NewEnvironmentTestHelper creates a new environment test helper
func NewEnvironmentTestHelper() *EnvironmentTestHelper {
	return &EnvironmentTestHelper{
		originalVars: make(map[string]string),
	}
}

// SetEnv sets an environment variable and remembers the original value
func (eth *EnvironmentTestHelper) SetEnv(key, value string) {
	if _, exists := eth.originalVars[key]; !exists {
		eth.originalVars[key] = os.Getenv(key)
	}
	os.Setenv(key, value)
}

// UnsetEnv unsets an environment variable and remembers the original value
func (eth *EnvironmentTestHelper) UnsetEnv(key string) {
	if _, exists := eth.originalVars[key]; !exists {
		eth.originalVars[key] = os.Getenv(key)
	}
	os.Unsetenv(key)
}

// RestoreEnv restores all modified environment variables
func (eth *EnvironmentTestHelper) RestoreEnv() {
	for key, value := range eth.originalVars {
		if value == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, value)
		}
	}
}

// Helper function to check if string contains substring (case-insensitive)
func containsString(haystack, needle string) bool {
	if needle == "" {
		return true
	}
	return len(haystack) >= len(needle) &&
		haystack[len(haystack)-len(needle):] == needle ||
		haystack[:len(needle)] == needle ||
		(len(haystack) > len(needle) &&
			func() bool {
				for i := 1; i <= len(haystack)-len(needle); i++ {
					if haystack[i:i+len(needle)] == needle {
						return true
					}
				}
				return false
			}())
}

// FileTestHelper provides file manipulation utilities for testing
type FileTestHelper struct {
	tempFiles []string
}

// NewFileTestHelper creates a new file test helper
func NewFileTestHelper() *FileTestHelper {
	return &FileTestHelper{}
}

// CreateTempFile creates a temporary file with content
func (fth *FileTestHelper) CreateTempFile(t *testing.T, pattern, content string) string {
	file, err := os.CreateTemp("", pattern)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	if content != "" {
		if _, err := file.WriteString(content); err != nil {
			file.Close()
			os.Remove(file.Name())
			t.Fatalf("Failed to write temp file content: %v", err)
		}
	}

	fileName := file.Name()
	file.Close()

	fth.tempFiles = append(fth.tempFiles, fileName)
	return fileName
}

// Cleanup removes all temporary files created by this helper
func (fth *FileTestHelper) Cleanup() {
	for _, file := range fth.tempFiles {
		os.Remove(file)
	}
	fth.tempFiles = nil
}

var Expect = AssertionHelpers{}
