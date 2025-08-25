package handlers

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
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
	t.Run("New", func(t *testing.T) {
		t.Run("creates handler successfully", func(t *testing.T) {
			_, cleanup := setupNoteTest(t)
			defer cleanup()

			handler, err := NewNoteHandler()
			if err != nil {
				t.Fatalf("NewNoteHandler failed: %v", err)
			}
			if handler == nil {
				t.Fatal("Handler should not be nil")
			}
			defer handler.Close()

			if handler.db == nil {
				t.Error("Handler database should not be nil")
			}
			if handler.config == nil {
				t.Error("Handler config should not be nil")
			}
			if handler.repos == nil {
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

	t.Run("parse content", func(t *testing.T) {
		handler := &NoteHandler{}

		testCases := []struct {
			name            string
			input           string
			expectedTitle   string
			expectedContent string
			expectedTags    []string
		}{
			{
				name: "note with title and tags",
				input: `# My Test Note

This is the content.

<!-- Tags: personal, work, important -->`,
				expectedTitle: "My Test Note",
				expectedContent: `# My Test Note

This is the content.

<!-- Tags: personal, work, important -->`,
				expectedTags: []string{"personal", "work", "important"},
			},
			{
				name: "note without title",
				input: `Just some content here.

No title heading.

<!-- Tags: test -->`,
				expectedTitle: "",
				expectedContent: `Just some content here.

No title heading.

<!-- Tags: test -->`,
				expectedTags: []string{"test"},
			},
			{
				name: "note without tags",
				input: `# Title Only

Content without tags.`,
				expectedTitle: "Title Only",
				expectedContent: `# Title Only

Content without tags.`,
				expectedTags: nil,
			},
			{
				name: "empty tags comment",
				input: `# Test Note

Content here.

<!-- Tags: -->`,
				expectedTitle: "Test Note",
				expectedContent: `# Test Note

Content here.

<!-- Tags: -->`,
				expectedTags: nil,
			},
			{
				name: "malformed tags comment",
				input: `# Test Note

Content here.

<!-- Tags: tag1, , tag2,, tag3 -->`,
				expectedTitle: "Test Note",
				expectedContent: `# Test Note

Content here.

<!-- Tags: tag1, , tag2,, tag3 -->`,
				expectedTags: []string{"tag1", "tag2", "tag3"},
			},
			{
				name: "multiple headings",
				input: `## Secondary Heading

# Main Title

Content here.`,
				expectedTitle: "Main Title",
				expectedContent: `## Secondary Heading

# Main Title

Content here.`,
				expectedTags: nil,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				title, content, tags := handler.parseNoteContent(tc.input)

				if title != tc.expectedTitle {
					t.Errorf("Expected title %q, got %q", tc.expectedTitle, title)
				}

				if content != tc.expectedContent {
					t.Errorf("Expected content %q, got %q", tc.expectedContent, content)
				}

				if len(tags) != len(tc.expectedTags) {
					t.Errorf("Expected %d tags, got %d", len(tc.expectedTags), len(tags))
				}

				for i, expectedTag := range tc.expectedTags {
					if i >= len(tags) || tags[i] != expectedTag {
						t.Errorf("Expected tag %q at position %d, got %q", expectedTag, i, tags[i])
					}
				}
			})
		}
	})

	t.Run("IsFile helper", func(t *testing.T) {
		testCases := []struct {
			name     string
			input    string
			expected bool
		}{
			{"file with extension", "test.md", true},
			{"file with multiple extensions", "test.tar.gz", true},
			{"path with slash", "/path/to/file", true},
			{"path with backslash", "path\\to\\file", true},
			{"relative path", "./file", true},
			{"just text", "hello", false},
			{"empty string", "", false},
		}

		tempDir, err := os.MkdirTemp("", "isfile-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		existingFile := filepath.Join(tempDir, "existing")
		err = os.WriteFile(existingFile, []byte("test"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		testCases = append(testCases, struct {
			name     string
			input    string
			expected bool
		}{"existing file without extension", existingFile, true})

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := isFile(tc.input)
				if result != tc.expected {
					t.Errorf("isFile(%q) = %v, expected %v", tc.input, result, tc.expected)
				}
			})
		}
	})

	t.Run("getEditor", func(t *testing.T) {
		handler := &NoteHandler{}

		t.Run("uses EDITOR environment variable", func(t *testing.T) {
			originalEditor := os.Getenv("EDITOR")
			os.Setenv("EDITOR", "test-editor")
			defer os.Setenv("EDITOR", originalEditor)

			editor := handler.getEditor()
			if editor != "test-editor" {
				t.Errorf("Expected 'test-editor', got %q", editor)
			}
		})

		t.Run("finds available editor", func(t *testing.T) {
			originalEditor := os.Getenv("EDITOR")
			os.Unsetenv("EDITOR")
			defer os.Setenv("EDITOR", originalEditor)

			editor := handler.getEditor()
			if editor == "" {
				t.Skip("No common editors found on system, skipping test")
			}
		})

		t.Run("returns empty when no editor available", func(t *testing.T) {
			originalEditor := os.Getenv("EDITOR")
			originalPath := os.Getenv("PATH")

			os.Unsetenv("EDITOR")
			os.Setenv("PATH", "")

			defer func() {
				os.Setenv("EDITOR", originalEditor)
				os.Setenv("PATH", originalPath)
			}()

			editor := handler.getEditor()
			if editor != "" {
				t.Errorf("Expected empty string when no editor available, got %q", editor)
			}
		})
	})

	t.Run("Create Errors", func(t *testing.T) {
		errorTests := []struct {
			name        string
			setupFunc   func(t *testing.T) (cleanup func())
			args        []string
			expectError bool
			errorSubstr string
		}{
			{
				name: "database initialization error",
				setupFunc: func(t *testing.T) func() {
					if runtime.GOOS == "windows" {
						original := os.Getenv("APPDATA")
						os.Unsetenv("APPDATA")
						return func() { os.Setenv("APPDATA", original) }
					} else {
						originalXDG := os.Getenv("XDG_CONFIG_HOME")
						originalHome := os.Getenv("HOME")
						os.Unsetenv("XDG_CONFIG_HOME")
						os.Unsetenv("HOME")
						return func() {
							os.Setenv("XDG_CONFIG_HOME", originalXDG)
							os.Setenv("HOME", originalHome)
						}
					}
				},
				args:        []string{"Test Note"},
				expectError: true,
				errorSubstr: "failed to initialize database",
			},
			{
				name: "note creation in database fails",
				setupFunc: func(t *testing.T) func() {
					tempDir, cleanup := setupNoteTest(t)

					configDir := filepath.Join(tempDir, "noteleaf")
					dbPath := filepath.Join(configDir, "noteleaf.db")

					err := os.WriteFile(dbPath, []byte("invalid sqlite content"), 0644)
					if err != nil {
						t.Fatalf("Failed to corrupt database: %v", err)
					}

					return cleanup
				},
				args:        []string{"Test Note"},
				expectError: true,
				errorSubstr: "failed to initialize database",
			},
		}

		for _, tt := range errorTests {
			t.Run(tt.name, func(t *testing.T) {
				cleanup := tt.setupFunc(t)
				defer cleanup()

				oldStdin := os.Stdin
				r, w, _ := os.Pipe()
				os.Stdin = r
				defer func() { os.Stdin = oldStdin }()

				go func() {
					w.WriteString("n\n")
					w.Close()
				}()

				ctx := context.Background()
				err := Create(ctx, tt.args)

				if tt.expectError && err == nil {
					t.Errorf("Expected error containing %q, got nil", tt.errorSubstr)
				} else if !tt.expectError && err != nil {
					t.Errorf("Expected no error, got: %v", err)
				} else if tt.expectError && err != nil && !strings.Contains(err.Error(), tt.errorSubstr) {
					t.Errorf("Expected error containing %q, got: %v", tt.errorSubstr, err)
				}
			})
		}
	})

	t.Run("Create (args)", func(t *testing.T) {
		t.Run("creates note from title only", func(t *testing.T) {
			_, cleanup := setupNoteTest(t)
			defer cleanup()

			oldStdin := os.Stdin
			r, w, _ := os.Pipe()
			os.Stdin = r
			defer func() { os.Stdin = oldStdin }()

			go func() {
				w.WriteString("n\n")
				w.Close()
			}()

			ctx := context.Background()
			err := Create(ctx, []string{"Test Note"})
			if err != nil {
				t.Errorf("Create failed: %v", err)
			}
		})

		t.Run("creates note from title and content", func(t *testing.T) {
			_, cleanup := setupNoteTest(t)
			defer cleanup()

			oldStdin := os.Stdin
			r, w, _ := os.Pipe()
			os.Stdin = r
			defer func() { os.Stdin = oldStdin }()

			go func() {
				w.WriteString("n\n")
				w.Close()
			}()

			ctx := context.Background()
			err := Create(ctx, []string{"Test Note", "This", "is", "test", "content"})
			if err != nil {
				t.Errorf("Create failed: %v", err)
			}
		})

		t.Run("handles database connection error", func(t *testing.T) {
			tempDir, cleanup := setupNoteTest(t)
			defer cleanup()

			configDir := filepath.Join(tempDir, "noteleaf")
			dbPath := filepath.Join(configDir, "noteleaf.db")
			os.Remove(dbPath)

			os.MkdirAll(dbPath, 0755)
			defer os.RemoveAll(dbPath)

			ctx := context.Background()
			err := Create(ctx, []string{"Test Note"})
			if err == nil {
				t.Error("Create should fail when database is inaccessible")
			}
		})

		t.Run("New is alias for Create", func(t *testing.T) {
			_, cleanup := setupNoteTest(t)
			defer cleanup()

			oldStdin := os.Stdin
			r, w, _ := os.Pipe()
			os.Stdin = r
			defer func() { os.Stdin = oldStdin }()

			go func() {
				w.WriteString("n\n")
				w.Close()
			}()

			ctx := context.Background()
			err := New(ctx, []string{"Test Note via New"})
			if err != nil {
				t.Errorf("New failed: %v", err)
			}
		})
	})

	t.Run("Create from file", func(t *testing.T) {
		t.Run("creates note from markdown file", func(t *testing.T) {
			tempDir, cleanup := setupNoteTest(t)
			defer cleanup()

			content := `# My Test Note

This is the content of my test note.

## Section 2

More content here.

<!-- Tags: personal, work -->`

			filePath := createTestMarkdownFile(t, tempDir, "test.md", content)

			ctx := context.Background()
			err := Create(ctx, []string{filePath})
			if err != nil {
				t.Errorf("Create from file failed: %v", err)
			}
		})

		t.Run("handles non-existent file", func(t *testing.T) {
			_, cleanup := setupNoteTest(t)
			defer cleanup()

			ctx := context.Background()
			err := Create(ctx, []string{"/non/existent/file.md"})
			if err == nil {
				t.Error("Create should fail for non-existent file")
			}
			if !strings.Contains(err.Error(), "file does not exist") {
				t.Errorf("Expected file not found error, got: %v", err)
			}
		})

		t.Run("handles empty file", func(t *testing.T) {
			tempDir, cleanup := setupNoteTest(t)
			defer cleanup()

			filePath := createTestMarkdownFile(t, tempDir, "empty.md", "")

			ctx := context.Background()
			err := Create(ctx, []string{filePath})
			if err == nil {
				t.Error("Create should fail for empty file")
			}
			if !strings.Contains(err.Error(), "file is empty") {
				t.Errorf("Expected empty file error, got: %v", err)
			}
		})

		t.Run("handles whitespace-only file", func(t *testing.T) {
			tempDir, cleanup := setupNoteTest(t)
			defer cleanup()

			filePath := createTestMarkdownFile(t, tempDir, "whitespace.md", "   \n\t  \n  ")

			ctx := context.Background()
			err := Create(ctx, []string{filePath})
			if err == nil {
				t.Error("Create should fail for whitespace-only file")
			}
			if !strings.Contains(err.Error(), "file is empty") {
				t.Errorf("Expected empty file error, got: %v", err)
			}
		})

		t.Run("creates note without title in file", func(t *testing.T) {
			tempDir, cleanup := setupNoteTest(t)
			defer cleanup()

			content := `This note has no title heading.

Just some content here.`

			filePath := createTestMarkdownFile(t, tempDir, "notitle.md", content)

			ctx := context.Background()
			err := Create(ctx, []string{filePath})
			if err != nil {
				t.Errorf("Create from file without title failed: %v", err)
			}
		})

		t.Run("handles file read error", func(t *testing.T) {
			tempDir, cleanup := setupNoteTest(t)
			defer cleanup()

			filePath := createTestMarkdownFile(t, tempDir, "unreadable.md", "test content")
			err := os.Chmod(filePath, 0000)
			if err != nil {
				t.Fatalf("Failed to make file unreadable: %v", err)
			}
			defer os.Chmod(filePath, 0644)

			ctx := context.Background()
			err = Create(ctx, []string{filePath})
			if err == nil {
				t.Error("Create should fail for unreadable file")
			}
			if !strings.Contains(err.Error(), "failed to read file") {
				t.Errorf("Expected file read error, got: %v", err)
			}
		})
	})

	t.Run("Interactive Create", func(t *testing.T) {
		t.Run("handles no editor configured", func(t *testing.T) {
			_, cleanup := setupNoteTest(t)
			defer cleanup()

			originalEditor := os.Getenv("EDITOR")
			originalPath := os.Getenv("PATH")
			os.Unsetenv("EDITOR")
			os.Setenv("PATH", "")
			defer func() {
				os.Setenv("EDITOR", originalEditor)
				os.Setenv("PATH", originalPath)
			}()

			ctx := context.Background()
			err := Create(ctx, []string{})
			if err == nil {
				t.Error("Create should fail when no editor is configured")
			}
			if !strings.Contains(err.Error(), "no editor configured") {
				t.Errorf("Expected no editor error, got: %v", err)
			}
		})

		t.Run("handles editor command failure", func(t *testing.T) {
			_, cleanup := setupNoteTest(t)
			defer cleanup()

			originalEditor := os.Getenv("EDITOR")
			os.Setenv("EDITOR", "nonexistent-editor-12345")
			defer os.Setenv("EDITOR", originalEditor)

			ctx := context.Background()
			err := Create(ctx, []string{})
			if err == nil {
				t.Error("Create should fail when editor command fails")
			}
			if !strings.Contains(err.Error(), "failed to open editor") {
				t.Errorf("Expected editor failure error, got: %v", err)
			}
		})

		t.Run("creates note successfully with mocked editor", func(t *testing.T) {
			_, cleanup := setupNoteTest(t)
			defer cleanup()

			originalEditor := os.Getenv("EDITOR")
			os.Setenv("EDITOR", "test-editor")
			defer os.Setenv("EDITOR", originalEditor)

			handler, err := NewNoteHandler()
			if err != nil {
				t.Fatalf("NewNoteHandler failed: %v", err)
			}
			defer handler.Close()

			handler.openInEditorFunc = func(editor, filePath string) error {
				content := `# Test Note

This is edited content.

<!-- Tags: test, created -->`
				return os.WriteFile(filePath, []byte(content), 0644)
			}

			ctx := context.Background()
			err = handler.createInteractive(ctx)
			if err != nil {
				t.Errorf("Interactive create failed: %v", err)
			}
		})

		t.Run("handles editor cancellation", func(t *testing.T) {
			_, cleanup := setupNoteTest(t)
			defer cleanup()

			originalEditor := os.Getenv("EDITOR")
			os.Setenv("EDITOR", "test-editor")
			defer os.Setenv("EDITOR", originalEditor)

			handler, err := NewNoteHandler()
			if err != nil {
				t.Fatalf("NewNoteHandler failed: %v", err)
			}
			defer handler.Close()

			handler.openInEditorFunc = func(editor, filePath string) error {
				return nil
			}

			ctx := context.Background()
			err = handler.createInteractive(ctx)
			if err != nil {
				t.Errorf("Interactive create should handle cancellation gracefully: %v", err)
			}
		})
	})

	t.Run("Close", func(t *testing.T) {
		_, cleanup := setupNoteTest(t)
		defer cleanup()

		handler, err := NewNoteHandler()
		if err != nil {
			t.Fatalf("NewNoteHandler failed: %v", err)
		}

		err = handler.Close()
		if err != nil {
			t.Errorf("Close should not return error: %v", err)
		}

		handler.db = nil
		err = handler.Close()
		if err != nil {
			t.Errorf("Close should handle nil database gracefully: %v", err)
		}
	})

	t.Run("Edit", func(t *testing.T) {
		t.Run("validates argument count", func(t *testing.T) {
			_, cleanup := setupNoteTest(t)
			defer cleanup()

			ctx := context.Background()

			err := Edit(ctx, []string{})
			if err == nil {
				t.Error("Edit should fail with no arguments")
			}
			if !strings.Contains(err.Error(), "edit requires exactly one argument") {
				t.Errorf("Expected argument count error, got: %v", err)
			}

			err = Edit(ctx, []string{"1", "2"})
			if err == nil {
				t.Error("Edit should fail with too many arguments")
			}
			if !strings.Contains(err.Error(), "edit requires exactly one argument") {
				t.Errorf("Expected argument count error, got: %v", err)
			}
		})

		t.Run("validates note ID format", func(t *testing.T) {
			_, cleanup := setupNoteTest(t)
			defer cleanup()

			ctx := context.Background()

			err := Edit(ctx, []string{"invalid"})
			if err == nil {
				t.Error("Edit should fail with invalid note ID")
			}
			if !strings.Contains(err.Error(), "invalid note ID") {
				t.Errorf("Expected invalid ID error, got: %v", err)
			}

			err = Edit(ctx, []string{"-1"})
			if err == nil {
				t.Error("Edit should fail with negative note ID")
			}

			if !strings.Contains(err.Error(), "failed to get note") {
				t.Errorf("Expected note not found error for negative ID, got: %v", err)
			}
		})

		t.Run("handles non-existent note", func(t *testing.T) {
			_, cleanup := setupNoteTest(t)
			defer cleanup()

			ctx := context.Background()

			err := Edit(ctx, []string{"999"})
			if err == nil {
				t.Error("Edit should fail with non-existent note ID")
			}
			if !strings.Contains(err.Error(), "failed to get note") {
				t.Errorf("Expected note not found error, got: %v", err)
			}
		})

		t.Run("handles no editor configured", func(t *testing.T) {
			_, cleanup := setupNoteTest(t)
			defer cleanup()

			originalEditor := os.Getenv("EDITOR")
			originalPath := os.Getenv("PATH")
			os.Setenv("EDITOR", "")
			os.Setenv("PATH", "")
			defer func() {
				os.Setenv("EDITOR", originalEditor)
				os.Setenv("PATH", originalPath)
			}()

			ctx := context.Background()

			err := Create(ctx, []string{"Test Note", "Test content"})
			if err != nil {
				t.Fatalf("Failed to create test note: %v", err)
			}

			err = Edit(ctx, []string{"1"})
			if err == nil {
				t.Error("Edit should fail when no editor is configured")
			}

			if !strings.Contains(err.Error(), "no editor configured") && !strings.Contains(err.Error(), "failed to open editor") {
				t.Errorf("Expected no editor or editor failure error, got: %v", err)
			}
		})

		t.Run("handles editor command failure", func(t *testing.T) {
			_, cleanup := setupNoteTest(t)
			defer cleanup()

			originalEditor := os.Getenv("EDITOR")
			os.Setenv("EDITOR", "nonexistent-editor-12345")
			defer os.Setenv("EDITOR", originalEditor)

			ctx := context.Background()

			err := Create(ctx, []string{"Test Note", "Test content"})
			if err != nil {
				t.Fatalf("Failed to create test note: %v", err)
			}

			err = Edit(ctx, []string{"1"})
			if err == nil {
				t.Error("Edit should fail when editor command fails")
			}
			if !strings.Contains(err.Error(), "failed to open editor") {
				t.Errorf("Expected editor failure error, got: %v", err)
			}
		})

		t.Run("edits note successfully with mocked editor", func(t *testing.T) {
			_, cleanup := setupNoteTest(t)
			defer cleanup()

			originalEditor := os.Getenv("EDITOR")
			os.Setenv("EDITOR", "test-editor")
			defer os.Setenv("EDITOR", originalEditor)

			ctx := context.Background()

			err := Create(ctx, []string{"Original Title", "Original content"})
			if err != nil {
				t.Fatalf("Failed to create test note: %v", err)
			}

			handler, err := NewNoteHandler()
			if err != nil {
				t.Fatalf("NewNoteHandler failed: %v", err)
			}
			defer handler.Close()

			handler.openInEditorFunc = func(editor, filePath string) error {
				newContent := `# Updated Title

This is updated content.

<!-- Tags: updated, test -->`
				return os.WriteFile(filePath, []byte(newContent), 0644)
			}

			err = handler.editNote(ctx, 1)
			if err != nil {
				t.Errorf("Edit should succeed with mocked editor: %v", err)
			}

			note, err := handler.repos.Notes.Get(ctx, 1)
			if err != nil {
				t.Fatalf("Failed to get updated note: %v", err)
			}

			if note.Title != "Updated Title" {
				t.Errorf("Expected title 'Updated Title', got %q", note.Title)
			}

			if !strings.Contains(note.Content, "This is updated content") {
				t.Errorf("Expected content to contain 'This is updated content', got %q", note.Content)
			}

			expectedTags := []string{"updated", "test"}
			if len(note.Tags) != len(expectedTags) {
				t.Errorf("Expected %d tags, got %d", len(expectedTags), len(note.Tags))
			}
			for i, tag := range expectedTags {
				if i >= len(note.Tags) || note.Tags[i] != tag {
					t.Errorf("Expected tag %q at index %d, got %q", tag, i, note.Tags[i])
				}
			}
		})

		t.Run("handles editor cancellation (no changes)", func(t *testing.T) {
			_, cleanup := setupNoteTest(t)
			defer cleanup()

			originalEditor := os.Getenv("EDITOR")
			os.Setenv("EDITOR", "test-editor")
			defer os.Setenv("EDITOR", originalEditor)

			ctx := context.Background()

			err := Create(ctx, []string{"Test Note", "Test content"})
			if err != nil {
				t.Fatalf("Failed to create test note: %v", err)
			}

			handler, err := NewNoteHandler()
			if err != nil {
				t.Fatalf("NewNoteHandler failed: %v", err)
			}
			defer handler.Close()

			handler.openInEditorFunc = func(editor, filePath string) error {
				return nil
			}

			err = handler.editNote(ctx, 1)
			if err != nil {
				t.Errorf("Edit should handle cancellation gracefully: %v", err)
			}

			note, err := handler.repos.Notes.Get(ctx, 1)
			if err != nil {
				t.Fatalf("Failed to get note: %v", err)
			}

			if note.Title != "Test Note" {
				t.Errorf("Expected title 'Test Note', got %q", note.Title)
			}

			if note.Content != "Test content" {
				t.Errorf("Expected content 'Test content', got %q", note.Content)
			}
		})
	})
}
