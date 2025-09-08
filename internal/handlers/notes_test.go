package handlers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stormlightlabs/noteleaf/internal/models"
	"github.com/stormlightlabs/noteleaf/internal/store"
)

func setupNoteTest(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", "noteleaf-note-test-*")
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

	return tempDir, cleanup
}

func createTestMarkdownFile(t *testing.T, dir, filename, content string) string {
	filePath := filepath.Join(dir, filename)
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	return filePath
}

func TestNoteHandler(t *testing.T) {
	tempDir, cleanup := setupNoteTest(t)
	defer cleanup()

	handler, err := NewNoteHandler()
	if err != nil {
		t.Fatalf("Failed to create note handler: %v", err)
	}
	defer handler.Close()

	t.Run("New", func(t *testing.T) {
		t.Run("creates handler successfully", func(t *testing.T) {
			testHandler, err := NewNoteHandler()
			if err != nil {
				t.Fatalf("NewNoteHandler failed: %v", err)
			}
			if testHandler == nil {
				t.Fatal("Handler should not be nil")
			}
			defer testHandler.Close()

			if testHandler.db == nil {
				t.Error("Handler database should not be nil")
			}
			if testHandler.config == nil {
				t.Error("Handler config should not be nil")
			}
			if testHandler.repos == nil {
				t.Error("Handler repos should not be nil")
			}
		})

		t.Run("handles database initialization error", func(t *testing.T) {
			originalXDG := os.Getenv("XDG_CONFIG_HOME")
			originalHome := os.Getenv("HOME")

			if runtime.GOOS == "windows" {
				originalAppData := os.Getenv("APPDATA")
				os.Unsetenv("APPDATA")
				defer os.Setenv("APPDATA", originalAppData)
			} else {
				os.Unsetenv("XDG_CONFIG_HOME")
				os.Unsetenv("HOME")
			}

			defer func() {
				os.Setenv("XDG_CONFIG_HOME", originalXDG)
				os.Setenv("HOME", originalHome)
			}()

			_, err := NewNoteHandler()
			if err == nil {
				t.Error("NewNoteHandler should fail when database initialization fails")
			}
			if !strings.Contains(err.Error(), "failed to initialize database") {
				t.Errorf("Expected database error, got: %v", err)
			}
		})
	})

	t.Run("Create", func(t *testing.T) {
		ctx := context.Background()

		t.Run("creates note from title only", func(t *testing.T) {
			err := handler.Create(ctx, "Test Note 1", "", "", false)
			if err != nil {
				t.Errorf("Create failed: %v", err)
			}
		})

		t.Run("creates note from title and content", func(t *testing.T) {
			err := handler.Create(ctx, "Test Note 2", "This is test content", "", false)
			if err != nil {
				t.Errorf("Create failed: %v", err)
			}
		})

		t.Run("creates note from markdown file", func(t *testing.T) {
			content := `# My Test Note
<!-- tags: personal, work -->

This is the content of my note.`
			filePath := createTestMarkdownFile(t, tempDir, "test.md", content)

			err := handler.Create(ctx, "", "", filePath, false)
			if err != nil {
				t.Errorf("Create from file failed: %v", err)
			}
		})

		t.Run("handles non-existent file", func(t *testing.T) {
			err := handler.Create(ctx, "", "", "/non/existent/file.md", false)
			if err == nil {
				t.Error("Create should fail with non-existent file")
			}
		})
	})

	t.Run("Edit", func(t *testing.T) {
		ctx := context.Background()

		t.Run("handles non-existent note", func(t *testing.T) {
			err := handler.Edit(ctx, 999)
			if err == nil {
				t.Error("Edit should fail with non-existent note ID")
			}
			if !strings.Contains(err.Error(), "failed to get note") && !strings.Contains(err.Error(), "failed to find note") {
				t.Errorf("Expected note not found error, got: %v", err)
			}
		})

		t.Run("handles no editor configured", func(t *testing.T) {
			originalEditor := os.Getenv("EDITOR")
			originalPath := os.Getenv("PATH")
			os.Setenv("EDITOR", "")
			os.Setenv("PATH", "")
			defer func() {
				os.Setenv("EDITOR", originalEditor)
				os.Setenv("PATH", originalPath)
			}()

			err := handler.Edit(ctx, 1)
			if err == nil {
				t.Error("Edit should fail when no editor is configured")
			}
			if !strings.Contains(err.Error(), "no editor configured") && !strings.Contains(err.Error(), "failed to open editor") {
				t.Errorf("Expected no editor error, got: %v", err)
			}
		})

		t.Run("handles database connection error", func(t *testing.T) {
			handler.db.Close()
			defer func() {
				var err error
				handler.db, err = store.NewDatabase()
				if err != nil {
					t.Fatalf("Failed to reconnect to database: %v", err)
				}
			}()

			err := handler.Edit(ctx, 1)
			if err == nil {
				t.Error("Edit should fail when database is closed")
			}
			if !strings.Contains(err.Error(), "failed to get note") {
				t.Errorf("Expected database error, got: %v", err)
			}
		})

		t.Run("handles temp file creation error", func(t *testing.T) {
			testHandler, err := NewNoteHandler()
			if err != nil {
				t.Fatalf("Failed to create test handler: %v", err)
			}
			defer testHandler.Close()

			err = testHandler.Create(ctx, "Temp File Test Note", "Test content", "", false)
			if err != nil {
				t.Fatalf("Failed to create test note: %v", err)
			}

			originalTempDir := os.Getenv("TMPDIR")
			os.Setenv("TMPDIR", "/non/existent/path")
			defer os.Setenv("TMPDIR", originalTempDir)

			err = testHandler.Edit(ctx, 1)
			if err == nil {
				t.Error("Edit should fail when temp file creation fails")
			}
			if !strings.Contains(err.Error(), "failed to create temporary file") {
				t.Errorf("Expected temp file error, got: %v", err)
			}
		})

		t.Run("handles editor failure", func(t *testing.T) {
			testHandler, err := NewNoteHandler()
			if err != nil {
				t.Fatalf("Failed to create test handler: %v", err)
			}
			defer testHandler.Close()

			err = testHandler.Create(ctx, "Editor Failure Test Note", "Test content", "", false)
			if err != nil {
				t.Fatalf("Failed to create test note: %v", err)
			}

			mockEditor := func(editor, filePath string) error {
				return fmt.Errorf("editor process failed")
			}
			testHandler.openInEditorFunc = mockEditor

			err = testHandler.Edit(ctx, 1)
			if err == nil {
				t.Error("Edit should fail when editor fails")
			}
			if !strings.Contains(err.Error(), "failed to open editor") {
				t.Errorf("Expected editor error, got: %v", err)
			}
		})

		t.Run("handles temp file write error", func(t *testing.T) {
			originalHandler := handler.openInEditorFunc
			defer func() { handler.openInEditorFunc = originalHandler }()

			mockEditor := func(editor, filePath string) error {
				return os.Chmod(filePath, 0444)
			}
			handler.openInEditorFunc = mockEditor

			err := handler.Edit(ctx, 1)
			if err == nil {
				t.Error("Edit should handle temp file write issues")
			}
		})

		t.Run("handles file read error after editing", func(t *testing.T) {
			testHandler, err := NewNoteHandler()
			if err != nil {
				t.Fatalf("Failed to create test handler: %v", err)
			}
			defer testHandler.Close()

			err = testHandler.Create(ctx, "File Read Error Test Note", "Test content", "", false)
			if err != nil {
				t.Fatalf("Failed to create test note: %v", err)
			}

			mockEditor := func(editor, filePath string) error {
				return os.Remove(filePath)
			}
			testHandler.openInEditorFunc = mockEditor

			err = testHandler.Edit(ctx, 1)
			if err == nil {
				t.Error("Edit should fail when temp file is deleted")
			}
			if !strings.Contains(err.Error(), "failed to read edited content") {
				t.Errorf("Expected file read error, got: %v", err)
			}
		})

		t.Run("handles database update error", func(t *testing.T) {
			handler := NewHandlerTestHelper(t)
			id := handler.CreateTestNote(t, "Database Update Error Test Note", "Test content", nil)

			dbHelper := NewDatabaseTestHelper(handler)
			dbHelper.DropNotesTable()

			mockEditor := NewMockEditor().WithContent(`# Modified Note

Modified content here.`)
			handler.openInEditorFunc = mockEditor.GetEditorFunc()

			err := handler.Edit(ctx, id)
			Expect.AssertError(t, err, "failed to get note", "Edit should fail when database is corrupted")
		})

		t.Run("handles validation error - corrupted note content", func(t *testing.T) {
			handler := NewHandlerTestHelper(t)
			id := handler.CreateTestNote(t, "Corrupted Content Test Note", "Test content", nil)

			invalidContent := string([]byte{0, 1, 2, 255, 254, 253})
			mockEditor := NewMockEditor().WithContent(invalidContent)
			handler.openInEditorFunc = mockEditor.GetEditorFunc()

			err := handler.Edit(ctx, id)
			if err != nil && !strings.Contains(err.Error(), "failed to update note") {
				t.Errorf("Edit should handle corrupted content gracefully, got: %v", err)
			}
		})

		t.Run("handles validation error - empty note after edit", func(t *testing.T) {
			mockEditor := func(editor, filePath string) error {
				return os.WriteFile(filePath, []byte(""), 0644)
			}
			handler.openInEditorFunc = mockEditor

			err := handler.Edit(ctx, 1)
			if err != nil {
				t.Logf("Edit with empty content handled: %v", err)
			}
		})

		t.Run("handles database constraint violations", func(t *testing.T) {
			db, dbErr := store.NewDatabase()
			if dbErr != nil {
				t.Fatalf("Failed to get new database: %v", dbErr)
			}
			defer db.Close()

			_, execErr := db.Exec(`ALTER TABLE notes ADD CONSTRAINT test_constraint
				CHECK (length(title) > 0 AND length(title) < 5)`)
			if execErr != nil {
				t.Skipf("Could not add constraint for test: %v", execErr)
			}

			handler.db.Close()
			handler.db = db

			mockEditor := func(editor, filePath string) error {
				content := `# This Title Is Way Too Long For The Test Constraint

Content here.`
				return os.WriteFile(filePath, []byte(content), 0644)
			}
			handler.openInEditorFunc = mockEditor

			err := handler.Edit(ctx, 1)
			if err == nil {
				t.Error("Edit should fail with constraint violation")
			}
			if !strings.Contains(err.Error(), "failed to update note") {
				t.Errorf("Expected constraint violation error, got: %v", err)
			}
		})

		t.Run("handles database transaction rollback", func(t *testing.T) {
			handler.db.Close()
			var dbErr error
			handler.db, dbErr = store.NewDatabase()
			if dbErr != nil {
				t.Fatalf("Failed to reconnect: %v", dbErr)
			}

			handler.db.Exec("BEGIN TRANSACTION")
			handler.db.Exec("UPDATE notes SET title = 'locked' WHERE id = 1")

			db2, err2 := store.NewDatabase()
			if err2 != nil {
				t.Fatalf("Failed to create second connection: %v", err2)
			}
			defer db2.Close()

			oldDB := handler.db
			handler.db = db2

			mockEditor := func(editor, filePath string) error {
				content := `# Modified Title

Modified content.`
				return os.WriteFile(filePath, []byte(content), 0644)
			}
			handler.openInEditorFunc = mockEditor

			err := handler.Edit(ctx, 1)

			oldDB.Exec("ROLLBACK")
			handler.db = oldDB

			if err == nil {
				t.Log("Edit succeeded despite transaction conflict")
			}
		})

		t.Run("handles successful edit", func(t *testing.T) {
			handler := NewHandlerTestHelper(t)
			id := handler.CreateTestNote(t, "Edit Test Note", "Original content", nil)

			mockEditor := NewMockEditor().WithContent(`# Modified Edit Test Note

This is the modified content.

<!-- Tags: modified, test -->`)
			handler.openInEditorFunc = mockEditor.GetEditorFunc()

			err := handler.Edit(ctx, id)
			Expect.AssertNoError(t, err, "Edit should succeed")
		})
	})

	t.Run("Edit Errors", func(t *testing.T) {
		t.Run("API Failures", func(t *testing.T) {
			t.Run("handles non-existent note", func(t *testing.T) {
				handler := NewHandlerTestHelper(t)
				ctx := context.Background()

				err := handler.Edit(ctx, 999)
				Expect.AssertError(t, err, "failed to get note", "Edit should fail with non-existent note ID")
			})

			t.Run("handles no editor configured", func(t *testing.T) {
				handler := NewHandlerTestHelper(t)
				ctx := context.Background()

				noteID := handler.CreateTestNote(t, "Test Note", "Test content", nil)

				envHelper := NewEnvironmentTestHelper()
				defer envHelper.RestoreEnv()

				envHelper.UnsetEnv("EDITOR")
				envHelper.SetEnv("PATH", "")

				err := handler.Edit(ctx, noteID)
				Expect.AssertError(t, err, "failed to open editor", "Edit should fail when no editor is configured")
			})

			t.Run("handles temp file creation error", func(t *testing.T) {
				handler := NewHandlerTestHelper(t)
				ctx := context.Background()

				noteID := handler.CreateTestNote(t, "Test Note", "Test content", nil)

				envHelper := NewEnvironmentTestHelper()
				defer envHelper.RestoreEnv()

				envHelper.SetEnv("TMPDIR", "/non/existent/path")

				err := handler.Edit(ctx, noteID)
				Expect.AssertError(t, err, "failed to create temporary file", "Edit should fail when temp file creation fails")
			})

			t.Run("handles editor failure", func(t *testing.T) {
				handler := NewHandlerTestHelper(t)
				ctx := context.Background()

				noteID := handler.CreateTestNote(t, "Test Note", "Test content", nil)

				mockEditor := NewMockEditor().WithFailure("editor process failed")
				handler.openInEditorFunc = mockEditor.GetEditorFunc()

				err := handler.Edit(ctx, noteID)
				Expect.AssertError(t, err, "failed to open editor", "Edit should fail when editor fails")
			})

			t.Run("handles file read error after editing", func(t *testing.T) {
				handler := NewHandlerTestHelper(t)
				ctx := context.Background()

				noteID := handler.CreateTestNote(t, "Test Note", "Test content", nil)

				mockEditor := NewMockEditor().WithFileDeleted()
				handler.openInEditorFunc = mockEditor.GetEditorFunc()

				err := handler.Edit(ctx, noteID)
				Expect.AssertError(t, err, "failed to read edited content", "Edit should fail when temp file is deleted")
			})
		})

		t.Run("Database Errors", func(t *testing.T) {
			t.Run("handles database connection error", func(t *testing.T) {
				handler := NewHandlerTestHelper(t)
				ctx := context.Background()

				noteID := handler.CreateTestNote(t, "Test Note", "Test content", nil)

				dbHelper := NewDatabaseTestHelper(handler)
				dbHelper.CloseDatabase()

				err := handler.Edit(ctx, noteID)
				Expect.AssertError(t, err, "failed to get note", "Edit should fail when database is closed")
			})

			t.Run("handles database update error", func(t *testing.T) {
				handler := NewHandlerTestHelper(t)
				ctx := context.Background()

				noteID := handler.CreateTestNote(t, "Test Note", "Test content", nil)

				dbHelper := NewDatabaseTestHelper(handler)
				dbHelper.DropNotesTable()

				mockEditor := NewMockEditor().WithContent(`# Modified Note

Modified content here.`)
				handler.openInEditorFunc = mockEditor.GetEditorFunc()

				err := handler.Edit(ctx, noteID)
				Expect.AssertError(t, err, "failed to get note", "Edit should fail when database table is missing")
			})

			t.Run("handles database constraint violations", func(t *testing.T) {
				handler := NewHandlerTestHelper(t)
				ctx := context.Background()

				noteID := handler.CreateTestNote(t, "Test Note", "Test content", nil)

				_, execErr := handler.db.Exec(`ALTER TABLE notes ADD CONSTRAINT test_constraint
				CHECK (length(title) > 0 AND length(title) < 5)`)
				if execErr != nil {
					t.Skipf("Could not add constraint for test: %v", execErr)
				}

				mockEditor := NewMockEditor().WithContent(`# This Title Is Way Too Long For The Test Constraint

Content here.`)
				handler.openInEditorFunc = mockEditor.GetEditorFunc()

				err := handler.Edit(ctx, noteID)
				Expect.AssertError(t, err, "failed to update note", "Edit should fail with constraint violation")
			})
		})

		t.Run("Validation Errors", func(t *testing.T) {
			t.Run("handles corrupted note content", func(t *testing.T) {
				handler := NewHandlerTestHelper(t)
				ctx := context.Background()

				noteID := handler.CreateTestNote(t, "Test Note", "Test content", nil)

				invalidContent := string([]byte{0, 1, 2, 255, 254, 253})
				mockEditor := NewMockEditor().WithContent(invalidContent)
				handler.openInEditorFunc = mockEditor.GetEditorFunc()

				err := handler.Edit(ctx, noteID)
				if err != nil && !strings.Contains(err.Error(), "failed to update note") {
					t.Errorf("Edit should handle corrupted content gracefully, got: %v", err)
				}
			})

			t.Run("handles empty note after edit", func(t *testing.T) {
				handler := NewHandlerTestHelper(t)
				ctx := context.Background()

				noteID := handler.CreateTestNote(t, "Test Note", "Test content", nil)

				mockEditor := NewMockEditor().WithContent("")
				handler.openInEditorFunc = mockEditor.GetEditorFunc()

				err := handler.Edit(ctx, noteID)
				if err != nil {
					t.Logf("Edit with empty content handled: %v", err)
				}
			})

			t.Run("handles very large content", func(t *testing.T) {
				handler := NewHandlerTestHelper(t)
				ctx := context.Background()

				noteID := handler.CreateTestNote(t, "Test Note", "Test content", nil)

				largeContent := fmt.Sprintf("# Large Note\n\n%s", strings.Repeat("Large content ", 70000))
				mockEditor := NewMockEditor().WithContent(largeContent)
				handler.openInEditorFunc = mockEditor.GetEditorFunc()

				err := handler.Edit(ctx, noteID)
				if err != nil {
					t.Logf("Edit with large content handled: %v", err)
				} else {
					t.Log("Edit succeeded with large content")
				}
			})
		})

		t.Run("Success Cases", func(t *testing.T) {
			t.Run("handles successful edit with title and tags", func(t *testing.T) {
				handler := NewHandlerTestHelper(t)
				ctx := context.Background()

				noteID := handler.CreateTestNote(t, "Original Note", "Original content", []string{"original"})

				mockEditor := NewMockEditor().WithContent(`# Modified Note

This is the modified content.

<!-- Tags: modified, test -->`)
				handler.openInEditorFunc = mockEditor.GetEditorFunc()

				err := handler.Edit(ctx, noteID)
				Expect.AssertNoError(t, err, "Edit should succeed")

				Expect.AssertNoteExists(t, handler, noteID)
			})

			t.Run("handles no changes made", func(t *testing.T) {
				handler := NewHandlerTestHelper(t)
				ctx := context.Background()

				noteID := handler.CreateTestNote(t, "Test Note", "Test content", nil)

				originalContent := handler.formatNoteForEdit(&models.Note{
					ID:      noteID,
					Title:   "Test Note",
					Content: "Test content",
					Tags:    nil,
				})
				mockEditor := NewMockEditor().WithContent(originalContent)
				handler.openInEditorFunc = mockEditor.GetEditorFunc()

				err := handler.Edit(ctx, noteID)
				Expect.AssertNoError(t, err, "Edit should succeed even with no changes")
			})

			t.Run("handles content without title", func(t *testing.T) {
				handler := NewHandlerTestHelper(t)
				ctx := context.Background()

				noteID := handler.CreateTestNote(t, "Original Title", "Original content", nil)

				mockEditor := NewMockEditor().WithContent("Just some content without a title")
				handler.openInEditorFunc = mockEditor.GetEditorFunc()

				err := handler.Edit(ctx, noteID)
				Expect.AssertNoError(t, err, "Edit should succeed without title")
			})
		})
	})

	t.Run("Read/View", func(t *testing.T) {
		ctx := context.Background()

		t.Run("views note successfully", func(t *testing.T) {
			testHandler, err := NewNoteHandler()
			if err != nil {
				t.Fatalf("Failed to create test handler: %v", err)
			}
			defer testHandler.Close()

			err = testHandler.Create(ctx, "View Test Note", "Test content for viewing", "", false)
			if err != nil {
				t.Fatalf("Failed to create test note: %v", err)
			}

			err = testHandler.View(ctx, 1)
			if err != nil {
				t.Errorf("View should succeed: %v", err)
			}
		})

		t.Run("handles non-existent note", func(t *testing.T) {
			err := handler.View(ctx, 999)
			if err == nil {
				t.Error("View should fail with non-existent note ID")
			}
			if !strings.Contains(err.Error(), "failed to get note") && !strings.Contains(err.Error(), "failed to find note") {
				t.Errorf("Expected note not found error, got: %v", err)
			}
		})

	})

	t.Run("List", func(t *testing.T) {
		ctx := context.Background()

		t.Run("lists with archived filter", func(t *testing.T) {
			testHandler, err := NewNoteHandler()
			if err != nil {
				t.Fatalf("Failed to create test handler: %v", err)
			}
			defer testHandler.Close()

			err = testHandler.List(ctx, true, true, nil)
			if err != nil {
				t.Errorf("List with archived filter should succeed: %v", err)
			}
		})

		t.Run("lists with tag filter", func(t *testing.T) {
			testHandler, err := NewNoteHandler()
			if err != nil {
				t.Fatalf("Failed to create test handler: %v", err)
			}
			defer testHandler.Close()

			err = testHandler.List(ctx, true, false, []string{"work", "personal"})
			if err != nil {
				t.Errorf("List with tag filter should succeed: %v", err)
			}
		})

		t.Run("handles empty note list", func(t *testing.T) {
			_, emptyCleanup := setupNoteTest(t)
			defer emptyCleanup()

			emptyHandler, err := NewNoteHandler()
			if err != nil {
				t.Fatalf("Failed to create empty handler: %v", err)
			}
			defer emptyHandler.Close()

			err = emptyHandler.List(ctx, true, false, nil)
			if err != nil {
				t.Errorf("ListStatic should succeed with empty list: %v", err)
			}
		})
	})

	t.Run("Delete", func(t *testing.T) {
		ctx := context.Background()

		t.Run("handles non-existent note", func(t *testing.T) {
			testHandler, err := NewNoteHandler()
			if err != nil {
				t.Fatalf("Failed to create test handler: %v", err)
			}
			defer testHandler.Close()

			err = testHandler.Delete(ctx, 999)
			if err == nil {
				t.Error("Delete should fail with non-existent note ID")
			}
			if !strings.Contains(err.Error(), "failed to get note") && !strings.Contains(err.Error(), "failed to find note") {
				t.Errorf("Expected note not found error, got: %v", err)
			}
		})

		t.Run("deletes note successfully", func(t *testing.T) {
			testHandler, err := NewNoteHandler()
			if err != nil {
				t.Fatalf("Failed to create test handler: %v", err)
			}
			defer testHandler.Close()

			err = testHandler.Create(ctx, "Note to Delete", "This will be deleted", "", false)
			if err != nil {
				t.Fatalf("Failed to create test note: %v", err)
			}

			err = testHandler.Delete(ctx, 1)
			if err != nil {
				t.Errorf("Delete should succeed: %v", err)
			}

			err = testHandler.View(ctx, 1)
			if err == nil {
				t.Error("Note should be gone after deletion")
			}
		})

		t.Run("deletes note with file path", func(t *testing.T) {
			testTempDir, testCleanup := setupNoteTest(t)
			defer testCleanup()

			testHandler, err := NewNoteHandler()
			if err != nil {
				t.Fatalf("Failed to create test handler: %v", err)
			}
			defer testHandler.Close()

			filePath := createTestMarkdownFile(t, testTempDir, "delete-test.md", "# Test Note\n\nTest content")

			err = testHandler.Create(ctx, "", "", filePath, false)
			if err != nil {
				t.Fatalf("Failed to create test note from file: %v", err)
			}

			err = testHandler.Delete(ctx, 1)
			if err != nil {
				t.Errorf("Delete should succeed: %v", err)
			}

			err = testHandler.View(ctx, 1)
			if err == nil {
				t.Error("Note should be gone after deletion")
			}
		})
	})

	t.Run("Close", func(t *testing.T) {
		testHandler, err := NewNoteHandler()
		if err != nil {
			t.Fatalf("Failed to create test handler: %v", err)
		}

		err = testHandler.Close()
		if err != nil {
			t.Errorf("Close should succeed: %v", err)
		}
	})

	t.Run("Helper Methods", func(t *testing.T) {
		t.Run("parseNoteContent", func(t *testing.T) {
			tests := []struct {
				name            string
				content         string
				expectedTitle   string
				expectedContent string
				expectedTags    []string
			}{
				{
					name:            "note with title and tags",
					content:         "# My Note\n<!-- tags: work, personal -->\n\nContent here",
					expectedTitle:   "My Note",
					expectedContent: "# My Note\n<!-- tags: work, personal -->\n\nContent here",
					expectedTags:    nil,
				},
				{
					name:            "note without title",
					content:         "Just some content without title",
					expectedTitle:   "",
					expectedContent: "Just some content without title",
					expectedTags:    nil,
				},
				{
					name:            "note without tags",
					content:         "# Title Only\n\nContent here",
					expectedTitle:   "Title Only",
					expectedContent: "# Title Only\n\nContent here",
					expectedTags:    nil,
				},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					title, content, tags := handler.parseNoteContent(tt.content)
					if title != tt.expectedTitle {
						t.Errorf("Expected title %q, got %q", tt.expectedTitle, title)
					}
					if content != tt.expectedContent {
						t.Errorf("Expected content %q, got %q", tt.expectedContent, content)
					}
					if len(tags) != len(tt.expectedTags) {
						t.Errorf("Expected %d tags, got %d", len(tt.expectedTags), len(tags))
					}
					for i, tag := range tt.expectedTags {
						if i < len(tags) && tags[i] != tag {
							t.Errorf("Expected tag %q, got %q", tag, tags[i])
						}
					}
				})
			}
		})

		t.Run("getEditor", func(t *testing.T) {
			originalEditor := os.Getenv("EDITOR")
			defer os.Setenv("EDITOR", originalEditor)

			t.Run("uses EDITOR environment variable", func(t *testing.T) {
				os.Setenv("EDITOR", "test-editor")
				editor := handler.getEditor()
				if editor != "test-editor" {
					t.Errorf("Expected 'test-editor', got %q", editor)
				}
			})

			t.Run("finds available editor", func(t *testing.T) {
				os.Unsetenv("EDITOR")
				editor := handler.getEditor()
				if editor == "" {
					t.Skip("No editors available in PATH")
				}
			})

			t.Run("returns empty when no editor available", func(t *testing.T) {
				os.Unsetenv("EDITOR")
				originalPath := os.Getenv("PATH")
				os.Setenv("PATH", "")
				defer os.Setenv("PATH", originalPath)

				editor := handler.getEditor()
				if editor != "" {
					t.Errorf("Expected empty editor, got %q", editor)
				}
			})
		})
	})

	t.Run("CreateInteractive", func(t *testing.T) {
		ctx := context.Background()

		t.Run("creates note successfully", func(t *testing.T) {
			handler := NewHandlerTestHelper(t)
			mockEditor := NewMockEditor().WithContent(`# Test Interactive Note

This is content from the interactive editor.

<!-- Tags: interactive, test -->`)
			handler.openInEditorFunc = mockEditor.GetEditorFunc()

			err := handler.createInteractive(ctx)
			Expect.AssertNoError(t, err, "createInteractive should succeed")
		})

		t.Run("handles cancelled note creation", func(t *testing.T) {
			handler := NewHandlerTestHelper(t)
			mockEditor := NewMockEditor().WithContent("") // Empty content simulates cancellation
			handler.openInEditorFunc = mockEditor.GetEditorFunc()

			err := handler.createInteractive(ctx)
			Expect.AssertNoError(t, err, "createInteractive should succeed even when cancelled")
		})

		t.Run("handles editor error", func(t *testing.T) {
			handler := NewHandlerTestHelper(t)
			mockEditor := NewMockEditor().WithFailure("editor failed to open")
			handler.openInEditorFunc = mockEditor.GetEditorFunc()

			err := handler.createInteractive(ctx)
			Expect.AssertError(t, err, "failed to open editor", "createInteractive should fail when editor fails")
		})

		t.Run("handles no editor configured", func(t *testing.T) {
			handler := NewHandlerTestHelper(t)
			envHelper := NewEnvironmentTestHelper()
			defer envHelper.RestoreEnv()

			envHelper.UnsetEnv("EDITOR")
			envHelper.SetEnv("PATH", "")

			err := handler.createInteractive(ctx)
			Expect.AssertError(t, err, "no editor configured", "createInteractive should fail when no editor is configured")
		})

		t.Run("handles file read error after editing", func(t *testing.T) {
			handler := NewHandlerTestHelper(t)
			mockEditor := NewMockEditor().WithFileDeleted()
			handler.openInEditorFunc = mockEditor.GetEditorFunc()

			err := handler.createInteractive(ctx)
			Expect.AssertError(t, err, "failed to read edited content", "createInteractive should fail when temp file is deleted")
		})
	})
}
