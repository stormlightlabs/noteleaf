package ui

import (
	"bytes"
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stormlightlabs/noteleaf/internal/models"
	"github.com/stormlightlabs/noteleaf/internal/repo"
	"github.com/stormlightlabs/noteleaf/internal/shared"
)

type mockNoteRepository struct {
	notes []*models.Note
	err   error
}

func (m *mockNoteRepository) List(ctx context.Context, options repo.NoteListOptions) ([]*models.Note, error) {
	if m.err != nil {
		return nil, m.err
	}

	var filtered []*models.Note
	for _, note := range m.notes {
		if options.Archived != nil && note.Archived != *options.Archived {
			continue
		}

		if len(options.Tags) > 0 {
			hasTag := false
			for _, filterTag := range options.Tags {
				if slices.Contains(note.Tags, filterTag) {
					hasTag = true
					break
				}
			}
			if !hasTag {
				continue
			}
		}

		if options.Content != "" && !strings.Contains(note.Content, options.Content) {
			continue
		}

		filtered = append(filtered, note)

		if options.Limit > 0 && len(filtered) >= options.Limit {
			break
		}
	}

	return filtered, nil
}

func (m *mockNoteRepository) ListPublished(ctx context.Context) ([]*models.Note, error) {
	if m.err != nil {
		return nil, m.err
	}
	var published []*models.Note
	for _, note := range m.notes {
		if note.LeafletRKey != nil && !note.IsDraft {
			published = append(published, note)
		}
	}
	return published, nil
}

func (m *mockNoteRepository) ListDrafts(ctx context.Context) ([]*models.Note, error) {
	if m.err != nil {
		return nil, m.err
	}
	var drafts []*models.Note
	for _, note := range m.notes {
		if note.LeafletRKey != nil && note.IsDraft {
			drafts = append(drafts, note)
		}
	}
	return drafts, nil
}

func (m *mockNoteRepository) GetLeafletNotes(ctx context.Context) ([]*models.Note, error) {
	if m.err != nil {
		return nil, m.err
	}
	var leafletNotes []*models.Note
	for _, note := range m.notes {
		if note.LeafletRKey != nil {
			leafletNotes = append(leafletNotes, note)
		}
	}
	return leafletNotes, nil
}

func TestNoteAdapter(t *testing.T) {
	t.Run("NoteRecord", func(t *testing.T) {
		note := &models.Note{
			ID:       1,
			Title:    "Test Note",
			Content:  "This is test content",
			Tags:     []string{"work", "important"},
			Archived: false,
			Created:  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			Modified: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
			FilePath: "/path/to/note.md",
		}
		record := &NoteRecord{Note: note}

		t.Run("GetField", func(t *testing.T) {
			tests := []struct {
				field    string
				expected any
				name     string
			}{
				{"id", int64(1), "should return note ID"},
				{"title", "Test Note", "should return note title"},
				{"content", "This is test content", "should return note content"},
				{"tags", []string{"work", "important"}, "should return note tags"},
				{"archived", false, "should return archived status"},
				{"unknown", "", "should return empty string for unknown field"},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					result := record.GetField(tt.field)
					if tags, ok := tt.expected.([]string); ok {
						resultTags, ok := result.([]string)
						if !ok || len(resultTags) != len(tags) {
							t.Errorf("GetField(%q) = %v, want %v", tt.field, result, tt.expected)
							return
						}
						for i, tag := range tags {
							if resultTags[i] != tag {
								t.Errorf("GetField(%q) = %v, want %v", tt.field, result, tt.expected)
								return
							}
						}
					} else if result != tt.expected {
						t.Errorf("GetField(%q) = %v, want %v", tt.field, result, tt.expected)
					}
				})
			}
		})

		t.Run("ListItem methods", func(t *testing.T) {
			if record.GetTitle() != "Test Note" {
				t.Errorf("GetTitle() = %q, want 'Test Note'", record.GetTitle())
			}

			description := record.GetDescription()
			if !strings.Contains(description, "work, important") {
				t.Errorf("GetDescription() should contain tags, got: %s", description)
			}
			if !strings.Contains(description, "2023-01-02 12:00") {
				t.Errorf("GetDescription() should contain modified time, got: %s", description)
			}

			filterValue := record.GetFilterValue()
			if !strings.Contains(filterValue, "Test Note") || !strings.Contains(filterValue, "work") {
				t.Errorf("GetFilterValue() should contain title and tags, got: %s", filterValue)
			}
		})

		t.Run("Model interface", func(t *testing.T) {
			if record.GetID() != 1 {
				t.Errorf("GetID() = %d, want 1", record.GetID())
			}

			if record.GetTableName() != "notes" {
				t.Errorf("GetTableName() = %q, want 'notes'", record.GetTableName())
			}
		})
	})

	t.Run("NoteDataSource", func(t *testing.T) {
		notes := []*models.Note{
			{
				ID:       1,
				Title:    "Work Note",
				Content:  "Work content",
				Tags:     []string{"work"},
				Archived: false,
				Created:  time.Now(),
				Modified: time.Now(),
			},
			{
				ID:       2,
				Title:    "Personal Note",
				Content:  "Personal content",
				Tags:     []string{"personal"},
				Archived: false,
				Created:  time.Now(),
				Modified: time.Now(),
			},
			{
				ID:       3,
				Title:    "Archived Note",
				Content:  "Archived content",
				Tags:     []string{"old"},
				Archived: true,
				Created:  time.Now(),
				Modified: time.Now(),
			},
		}

		t.Run("Load", func(t *testing.T) {
			repo := &mockNoteRepository{notes: notes}
			source := &NoteDataSource{repo: repo, showArchived: true}

			items, err := source.Load(context.Background(), ListOptions{})
			if err != nil {
				t.Fatalf("Load() failed: %v", err)
			}

			if len(items) != 3 {
				t.Errorf("Load() returned %d items, want 3", len(items))
			}

			if items[0].GetTitle() != "Work Note" {
				t.Errorf("First item title = %q, want 'Work Note'", items[0].GetTitle())
			}
		})

		t.Run("Load with archived filter", func(t *testing.T) {
			repo := &mockNoteRepository{notes: notes}
			source := &NoteDataSource{repo: repo, showArchived: false}

			items, err := source.Load(context.Background(), ListOptions{})
			if err != nil {
				t.Fatalf("Load() failed: %v", err)
			}

			if len(items) != 2 {
				t.Errorf("Load() with archived=false returned %d items, want 2", len(items))
			}
		})

		t.Run("Load with tag filter", func(t *testing.T) {
			repo := &mockNoteRepository{notes: notes}
			source := &NoteDataSource{repo: repo, showArchived: true, tags: []string{"work"}}

			items, err := source.Load(context.Background(), ListOptions{})
			if err != nil {
				t.Fatalf("Load() failed: %v", err)
			}

			if len(items) != 1 {
				t.Errorf("Load() with tags filter returned %d items, want 1", len(items))
			}
			if items[0].GetTitle() != "Work Note" {
				t.Errorf("Filtered item title = %q, want 'Work Note'", items[0].GetTitle())
			}
		})

		t.Run("Search", func(t *testing.T) {
			repo := &mockNoteRepository{notes: notes}
			source := &NoteDataSource{repo: repo, showArchived: true}

			items, err := source.Search(context.Background(), "Work", ListOptions{})
			if err != nil {
				t.Fatalf("Search() failed: %v", err)
			}

			if len(items) != 1 {
				t.Errorf("Search() returned %d items, want 1", len(items))
			}
			if items[0].GetTitle() != "Work Note" {
				t.Errorf("Search result title = %q, want 'Work Note'", items[0].GetTitle())
			}
		})

		t.Run("Load error", func(t *testing.T) {
			testErr := fmt.Errorf("test error")
			repo := &mockNoteRepository{err: testErr}
			source := &NoteDataSource{repo: repo}

			_, err := source.Load(context.Background(), ListOptions{})
			if err != testErr {
				t.Errorf("Load() error = %v, want %v", err, testErr)
			}
		})

		t.Run("Count", func(t *testing.T) {
			repo := &mockNoteRepository{notes: notes}
			source := &NoteDataSource{repo: repo, showArchived: true}

			count, err := source.Count(context.Background(), ListOptions{})
			if err != nil {
				t.Fatalf("Count() failed: %v", err)
			}

			if count != 3 {
				t.Errorf("Count() = %d, want 3", count)
			}
		})
	})

	t.Run("NewNoteDataList", func(t *testing.T) {
		repo := &mockNoteRepository{
			notes: []*models.Note{
				{
					ID:       1,
					Title:    "Test Note",
					Content:  "Test content",
					Tags:     []string{"test"},
					Archived: false,
					Created:  time.Now(),
					Modified: time.Now(),
				},
			},
		}

		opts := DataListOptions{
			Output: &bytes.Buffer{},
			Input:  strings.NewReader("q\n"),
			Static: true,
		}

		list := NewNoteDataList(repo, opts, false, nil)
		if list == nil {
			t.Fatal("NewNoteDataList() returned nil")
		}

		err := list.Browse(context.Background())
		if err != nil {
			t.Errorf("Browse() failed: %v", err)
		}
	})

	t.Run("NewNoteListFromList", func(t *testing.T) {
		repo := &mockNoteRepository{
			notes: []*models.Note{
				{
					ID:       1,
					Title:    "Test Note",
					Content:  "Test content",
					Tags:     []string{"test"},
					Archived: false,
					Created:  time.Now(),
					Modified: time.Now(),
				},
			},
		}

		output := &bytes.Buffer{}
		input := strings.NewReader("q\n")

		list := NewNoteListFromList(repo, output, input, true, false, nil)
		if list == nil {
			t.Fatal("NewNoteListFromList() returned nil")
		}

		err := list.Browse(context.Background())
		if err != nil {
			t.Errorf("Browse() failed: %v", err)
		}

		outputStr := output.String()
		shared.AssertContains(t, outputStr, "Notes", "Output should contain 'Notes' title")
		shared.AssertContains(t, outputStr, "Test Note", "Output should contain note title")
	})

	t.Run("Format Note for View", func(t *testing.T) {
		note := &models.Note{
			ID:       1,
			Title:    "Test Note",
			Content:  "# Test Note\n\nThis is the content.",
			Tags:     []string{"test", "example"},
			Archived: false,
			Created:  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			Modified: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
		}

		result := formatNoteForView(note)

		shared.AssertContains(t, result, "Test Note", "Formatted view should contain note title")
		shared.AssertContains(t, result, "test", "Formatted view should contain tags")
		shared.AssertContains(t, result, "2023-01-01", "Formatted view should contain created date")
		shared.AssertContains(t, result, "2023-01-02", "Formatted view should contain modified date")
	})
}
