package ui

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stormlightlabs/noteleaf/internal/models"
	"github.com/stormlightlabs/noteleaf/internal/repo"
)

type mockBookRepository struct {
	books []*models.Book
	err   error
}

func (m *mockBookRepository) List(ctx context.Context, options repo.BookListOptions) ([]*models.Book, error) {
	if m.err != nil {
		return nil, m.err
	}

	var filtered []*models.Book
	for _, book := range m.books {
		if options.Status != "" && book.Status != options.Status {
			continue
		}

		if options.Search != "" && !strings.Contains(strings.ToLower(book.Title), strings.ToLower(options.Search)) {
			continue
		}

		filtered = append(filtered, book)

		if options.Limit > 0 && len(filtered) >= options.Limit {
			break
		}
	}

	return filtered, nil
}

func TestBookAdapter(t *testing.T) {
	now := time.Now()
	book := &models.Book{
		ID:       1,
		Title:    "Test Book",
		Author:   "Test Author",
		Status:   "reading",
		Progress: 45,
		Pages:    250,
		Rating:   4.5,
		Notes:    "Great book so far",
		Added:    now,
		Started:  &now,
	}

	t.Run("BookRecord", func(t *testing.T) {
		record := &BookRecord{Book: book}

		t.Run("GetField", func(t *testing.T) {
			tests := []struct {
				field    string
				expected any
				name     string
			}{
				{"id", int64(1), "should return book ID"},
				{"title", "Test Book", "should return book title"},
				{"author", "Test Author", "should return book author"},
				{"status", "reading", "should return book status"},
				{"progress", 45, "should return book progress"},
				{"pages", 250, "should return page count"},
				{"rating", 4.5, "should return rating"},
				{"notes", "Great book so far", "should return notes"},
				{"added", now, "should return added time"},
				{"started", &now, "should return started time"},
				{"unknown", "", "should return empty string for unknown field"},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					result := record.GetField(tt.field)

					switch expected := tt.expected.(type) {
					case time.Time:
						if resultTime, ok := result.(time.Time); !ok || !resultTime.Equal(expected) {
							t.Errorf("GetField(%q) = %v, want %v", tt.field, result, tt.expected)
						}
					case *time.Time:
						resultPtr, ok := result.(*time.Time)
						if !ok || (expected == nil && resultPtr != nil) || (expected != nil && (resultPtr == nil || !resultPtr.Equal(*expected))) {
							t.Errorf("GetField(%q) = %v, want %v", tt.field, result, tt.expected)
						}
					default:
						if result != tt.expected {
							t.Errorf("GetField(%q) = %v, want %v", tt.field, result, tt.expected)
						}
					}
				})
			}
		})

		t.Run("ListItem methods", func(t *testing.T) {
			if record.GetTitle() != "Test Book" {
				t.Errorf("GetTitle() = %q, want 'Test Book'", record.GetTitle())
			}

			description := record.GetDescription()
			if !strings.Contains(description, "Test Author") {
				t.Errorf("GetDescription() should contain author, got: %s", description)
			}
			if !strings.Contains(description, "Reading") {
				t.Errorf("GetDescription() should contain status, got: %s", description)
			}
			if !strings.Contains(description, "250 pages") {
				t.Errorf("GetDescription() should contain page count, got: %s", description)
			}
			if !strings.Contains(description, "45%") {
				t.Errorf("GetDescription() should contain progress, got: %s", description)
			}

			filterValue := record.GetFilterValue()
			if !strings.Contains(filterValue, "Test Book") || !strings.Contains(filterValue, "Test Author") {
				t.Errorf("GetFilterValue() should contain title and author, got: %s", filterValue)
			}
			if !strings.Contains(filterValue, "Great book so far") {
				t.Errorf("GetFilterValue() should contain notes, got: %s", filterValue)
			}
		})

		t.Run("Model interface", func(t *testing.T) {
			if record.GetID() != 1 {
				t.Errorf("GetID() = %d, want 1", record.GetID())
			}

			if record.GetTableName() != "books" {
				t.Errorf("GetTableName() = %q, want 'books'", record.GetTableName())
			}
		})
	})

	t.Run("BookDataSource", func(t *testing.T) {
		books := []*models.Book{
			{
				ID:     1,
				Title:  "Go Programming",
				Author: "John Doe",
				Status: "reading",
			},
			{
				ID:     2,
				Title:  "Python Guide",
				Author: "Jane Smith",
				Status: "queued",
			},
			{
				ID:     3,
				Title:  "Advanced Go",
				Author: "Bob Johnson",
				Status: "finished",
			},
		}

		t.Run("Load", func(t *testing.T) {
			repo := &mockBookRepository{books: books}
			source := &BookDataSource{repo: repo}

			items, err := source.Load(context.Background(), ListOptions{})
			if err != nil {
				t.Fatalf("Load() failed: %v", err)
			}

			if len(items) != 3 {
				t.Errorf("Load() returned %d items, want 3", len(items))
			}

			if items[0].GetTitle() != "Go Programming" {
				t.Errorf("First item title = %q, want 'Go Programming'", items[0].GetTitle())
			}
		})

		t.Run("Load with status filter", func(t *testing.T) {
			repo := &mockBookRepository{books: books}
			source := &BookDataSource{repo: repo, status: "reading"}

			items, err := source.Load(context.Background(), ListOptions{})
			if err != nil {
				t.Fatalf("Load() failed: %v", err)
			}

			if len(items) != 1 {
				t.Errorf("Load() with status filter returned %d items, want 1", len(items))
			}
			if items[0].GetTitle() != "Go Programming" {
				t.Errorf("Filtered item title = %q, want 'Go Programming'", items[0].GetTitle())
			}
		})

		t.Run("Search", func(t *testing.T) {
			repo := &mockBookRepository{books: books}
			source := &BookDataSource{repo: repo}

			items, err := source.Search(context.Background(), "Python", ListOptions{})
			if err != nil {
				t.Fatalf("Search() failed: %v", err)
			}

			if len(items) != 1 {
				t.Errorf("Search() returned %d items, want 1", len(items))
			}
			if items[0].GetTitle() != "Python Guide" {
				t.Errorf("Search result title = %q, want 'Python Guide'", items[0].GetTitle())
			}
		})

		t.Run("Load error", func(t *testing.T) {
			testErr := fmt.Errorf("test error")
			repo := &mockBookRepository{err: testErr}
			source := &BookDataSource{repo: repo}

			if _, err := source.Load(context.Background(), ListOptions{}); err != testErr {
				t.Errorf("Load() error = %v, want %v", err, testErr)
			}
		})

		t.Run("Count", func(t *testing.T) {
			repo := &mockBookRepository{books: books}
			source := &BookDataSource{repo: repo}

			count, err := source.Count(context.Background(), ListOptions{})
			if err != nil {
				t.Fatalf("Count() failed: %v", err)
			}

			if count != 3 {
				t.Errorf("Count() = %d, want 3", count)
			}
		})

		t.Run("Count error", func(t *testing.T) {
			testErr := fmt.Errorf("test error")
			repo := &mockBookRepository{err: testErr}
			source := &BookDataSource{repo: repo}

			_, err := source.Count(context.Background(), ListOptions{})
			if err != testErr {
				t.Errorf("Count() error = %v, want %v", err, testErr)
			}
		})
	})

	t.Run("NewBookDataList", func(t *testing.T) {
		repo := &mockBookRepository{
			books: []*models.Book{
				{
					ID:     1,
					Title:  "Test Book",
					Author: "Test Author",
					Status: "reading",
				},
			},
		}

		opts := DataListOptions{
			Output: &bytes.Buffer{},
			Input:  strings.NewReader("q\n"),
			Static: true,
		}

		list := NewBookDataList(repo, opts, "")
		if list == nil {
			t.Fatal("NewBookDataList() returned nil")
		}

		err := list.Browse(context.Background())
		if err != nil {
			t.Errorf("Browse() failed: %v", err)
		}
	})

	t.Run("NewBookListFromList", func(t *testing.T) {
		repo := &mockBookRepository{
			books: []*models.Book{
				{
					ID:     1,
					Title:  "Test Book",
					Author: "Test Author",
					Status: "reading",
				},
			},
		}

		output := &bytes.Buffer{}
		input := strings.NewReader("q\n")

		list := NewBookListFromList(repo, output, input, true, "")
		if list == nil {
			t.Fatal("NewBookListFromList() returned nil")
		}

		err := list.Browse(context.Background())
		if err != nil {
			t.Errorf("Browse() failed: %v", err)
		}

		outputStr := output.String()
		if !strings.Contains(outputStr, "Books") {
			t.Error("Output should contain 'Books' title")
		}
		if !strings.Contains(outputStr, "Test Book") {
			t.Error("Output should contain book title")
		}
	})

	t.Run("Format Book for View", func(t *testing.T) {
		now := time.Now()
		started := now.Add(-time.Hour)

		book := &models.Book{
			ID:       1,
			Title:    "Test Book Title",
			Author:   "Test Author",
			Status:   "reading",
			Progress: 75,
			Pages:    250,
			Rating:   4.5,
			Notes:    "This is a great book with detailed explanations.",
			Added:    now,
			Started:  &started,
		}

		result := formatBookForView(book)

		if !strings.Contains(result, "Test Book Title") {
			t.Error("Formatted view should contain book title")
		}
		if !strings.Contains(result, "Test Author") {
			t.Error("Formatted view should contain author")
		}
		if !strings.Contains(result, "Reading") {
			t.Error("Formatted view should contain status")
		}
		if !strings.Contains(result, "75%") {
			t.Error("Formatted view should contain progress")
		}
		if !strings.Contains(result, "250") {
			t.Error("Formatted view should contain page count")
		}
		if !strings.Contains(result, "4.5/5") {
			t.Error("Formatted view should contain rating")
		}
		if !strings.Contains(result, "Added:") {
			t.Error("Formatted view should contain added date")
		}
		if !strings.Contains(result, "Started:") {
			t.Error("Formatted view should contain started date")
		}
		if !strings.Contains(result, "great book with detailed") {
			t.Error("Formatted view should contain notes")
		}
	})
}
