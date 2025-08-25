package ui

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stormlightlabs/noteleaf/internal/models"
)

// MockBookService implements services.APIService for testing
type MockBookService struct {
	searchResults []*models.Model
	searchError   error
	getResult     *models.Model
	getError      error
}

func (m *MockBookService) Search(ctx context.Context, query string, page, limit int) ([]*models.Model, error) {
	if m.searchError != nil {
		return nil, m.searchError
	}
	return m.searchResults, nil
}

func (m *MockBookService) Get(ctx context.Context, id string) (*models.Model, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	return m.getResult, nil
}

func (m *MockBookService) Check(ctx context.Context) error { return nil }
func (m *MockBookService) Close() error                    { return nil }

func TestBookList(t *testing.T) {
	t.Run("Options", func(t *testing.T) {
		t.Run("default options", func(t *testing.T) {
			opts := BookListOptions{}
			if opts.Static {
				t.Error("StaticMode should default to false")
			}
		})

		t.Run("static mode enabled", func(t *testing.T) {
			var buf bytes.Buffer
			opts := BookListOptions{
				Output: &buf,
				Static: true,
			}

			if !opts.Static {
				t.Error("StaticMode should be enabled")
			}
			if opts.Output != &buf {
				t.Error("Output should be set to buffer")
			}
		})
	})

	t.Run("Search & Select errors", func(t *testing.T) {
		t.Run("service search error", func(t *testing.T) {
			service := &MockBookService{
				searchError: errors.New("API error"),
			}

			var buf bytes.Buffer

			bl := &BookList{
				service: service,
				repo:    nil,
				opts: BookListOptions{
					Output: &buf,
					Static: true,
				},
			}

			err := bl.staticSelect(context.Background(), "test query")
			if err == nil {
				t.Fatal("Expected error, got nil")
			}

			output := buf.String()
			if !strings.Contains(output, "Error: API error") {
				t.Error("Error message not displayed")
			}
		})

		t.Run("no results found", func(t *testing.T) {
			service := &MockBookService{
				searchResults: []*models.Model{},
			}

			var buf bytes.Buffer

			bl := &BookList{
				service: service,
				repo:    nil,
				opts: BookListOptions{
					Output: &buf,
					Static: true,
				},
			}

			err := bl.staticSelect(context.Background(), "nonexistent")
			if err != nil {
				t.Fatalf("searchAndSelectStatic failed: %v", err)
			}

			output := buf.String()
			if !strings.Contains(output, "No books found") {
				t.Error("No results message not displayed")
			}
		})

		t.Run("successful search display", func(t *testing.T) {
			book1 := &models.Book{Title: "Test Book 1", Author: "Test Author 1"}
			book2 := &models.Book{Title: "Test Book 2", Author: "Test Author 2"}

			var model1 models.Model = book1
			var model2 models.Model = book2

			service := &MockBookService{
				searchResults: []*models.Model{&model1, &model2},
			}

			var buf bytes.Buffer

			bl := &BookList{
				service: service,
				// Skip repo operations for this test
				// repo:    nil,
				opts: BookListOptions{
					Output: &buf,
					Static: true,
				},
			}

			results, err := bl.service.Search(context.Background(), "test query", 1, 10)
			if err != nil {
				t.Fatalf("Search failed: %v", err)
			}

			_, err = bl.opts.Output.Write([]byte("Search Results for: test query\n\n"))
			if err != nil {
				t.Fatal(err)
			}

			for i, result := range results {
				if book, ok := (*result).(*models.Book); ok {
					line := []byte{}
					line = append(line, fmt.Sprintf("[%d] %s", i+1, book.Title)...)
					if book.Author != "" {
						line = append(line, fmt.Sprintf(" by %s", book.Author)...)
					}
					line = append(line, '\n')
					_, err = bl.opts.Output.Write(line)
					if err != nil {
						t.Fatal(err)
					}
				}
			}

			output := buf.String()

			if !strings.Contains(output, "Search Results for: test query") {
				t.Error("Search results title not found")
			}
			if !strings.Contains(output, "Test Book 1 by Test Author 1") {
				t.Error("First book not displayed")
			}
			if !strings.Contains(output, "Test Book 2 by Test Author 2") {
				t.Error("Second book not displayed")
			}
		})
	})

	t.Run("Interactive search", func(t *testing.T) {
		t.Run("static mode interactive search", func(t *testing.T) {
			book1 := &models.Book{Title: "Interactive Book", Author: "Interactive Author"}
			var model1 models.Model = book1

			service := &MockBookService{
				searchResults: []*models.Model{&model1},
			}

			var buf bytes.Buffer

			bl := &BookList{
				service: service,
				repo:    nil,
				opts: BookListOptions{
					Output: &buf,
					Static: true,
				},
			}

			err := bl.staticSearch(context.Background())
			if err != nil {
				t.Fatalf("InteractiveSearch failed: %v", err)
			}

			output := buf.String()

			if !strings.Contains(output, "Search for books: test query") {
				t.Error("Search prompt not displayed")
			}
		})
	})

	t.Run("View model", func(t *testing.T) {
		service := &MockBookService{}

		t.Run("searching state", func(t *testing.T) {
			model := searchModel{
				query:     "test",
				searching: true,
				service:   service,
				repo:      nil,
			}

			view := model.View()
			if !strings.Contains(view, "Searching...") {
				t.Error("Searching message not displayed")
			}
		})

		t.Run("error state", func(t *testing.T) {
			model := searchModel{
				query:   "test",
				err:     errors.New("test error"),
				service: service,
				repo:    nil,
			}

			view := model.View()
			if !strings.Contains(view, "Error: test error") {
				t.Error("Error message not displayed")
			}
		})

		t.Run("no results", func(t *testing.T) {
			model := searchModel{
				query:   "test",
				results: []*models.Book{},
				service: service,
				repo:    nil,
			}

			view := model.View()
			if !strings.Contains(view, "No books found") {
				t.Error("No results message not displayed")
			}
		})

		t.Run("with results", func(t *testing.T) {
			model := searchModel{
				query: "test",
				results: []*models.Book{
					{Title: "Book 1", Author: "Author 1"},
					{Title: "Book 2", Author: "Author 2"},
				},
				selected: 0,
				service:  service,
				repo:     nil,
			}

			view := model.View()
			if !strings.Contains(view, "Search Results for: test") {
				t.Error("Search results title not displayed")
			}
			if !strings.Contains(view, "Book 1 by Author 1") {
				t.Error("First book not displayed")
			}
			if !strings.Contains(view, "Book 2 by Author 2") {
				t.Error("Second book not displayed")
			}
			if !strings.Contains(view, "Use ↑/↓ to navigate") {
				t.Error("Navigation instructions not displayed")
			}
		})

		t.Run("confirmed state", func(t *testing.T) {
			book := &models.Book{Title: "Added Book", Author: "Added Author"}
			model := searchModel{
				query:     "test",
				confirmed: true,
				addedBook: book,
				results:   []*models.Book{book},
				service:   service,
				repo:      nil,
			}

			view := model.View()
			expected := "✓ Added book: Added Book by Added Author"
			if !strings.Contains(view, expected) {
				t.Errorf("Confirmation message not displayed correctly.\nExpected: %q\nActual: %q", expected, view)
			}
		})
	})
}
