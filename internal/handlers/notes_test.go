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
	})

	t.Run("Read/View", func(t *testing.T) {
		ctx := context.Background()

		t.Run("views note successfully", func(t *testing.T) {
			err := handler.View(ctx, 1)
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
			err := handler.List(ctx, true, true, nil)
			if err != nil {
				t.Errorf("List with archived filter should succeed: %v", err)
			}
		})

		t.Run("lists with tag filter", func(t *testing.T) {
			err := handler.List(ctx, true, false, []string{"work", "personal"})
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
			err := handler.Delete(ctx, 999)
			if err == nil {
				t.Error("Delete should fail with non-existent note ID")
			}
			if !strings.Contains(err.Error(), "failed to get note") && !strings.Contains(err.Error(), "failed to find note") {
				t.Errorf("Expected note not found error, got: %v", err)
			}
		})

		t.Run("deletes note successfully", func(t *testing.T) {
			err := handler.Create(ctx, "Note to Delete", "This will be deleted", "", false)
			if err != nil {
				t.Fatalf("Failed to create test note: %v", err)
			}

			// Delete the note (should be a high ID number since we've created many notes)
			err = handler.Delete(ctx, 1)
			if err != nil {
				t.Errorf("Delete should succeed: %v", err)
			}

			err = handler.View(ctx, 1)
			if err == nil {
				t.Error("Note should be gone after deletion")
			}
		})

		t.Run("deletes note with file path", func(t *testing.T) {
			filePath := createTestMarkdownFile(t, tempDir, "delete-test.md", "# Test Note\n\nTest content")

			err := handler.Create(ctx, "", "", filePath, false)
			if err != nil {
				t.Fatalf("Failed to create test note from file: %v", err)
			}

			err = handler.Create(ctx, "File Note to Delete", "", "", false)
			if err != nil {
				t.Fatalf("Failed to create file note: %v", err)
			}

			err = handler.Delete(ctx, 2)
			if err != nil {
				t.Errorf("Delete should succeed: %v", err)
			}

			err = handler.View(ctx, 2)
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
}
