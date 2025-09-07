package ui

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/stormlightlabs/noteleaf/internal/models"
	"github.com/stormlightlabs/noteleaf/internal/repo"
	"github.com/stormlightlabs/noteleaf/internal/utils"
)

// BookRecord adapts models.Book to work with DataList
type BookRecord struct {
	*models.Book
}

func (b *BookRecord) GetField(name string) any {
	switch name {
	case "id":
		return b.ID
	case "title":
		return b.Title
	case "author":
		return b.Author
	case "status":
		return b.Status
	case "progress":
		return b.Progress
	case "pages":
		return b.Pages
	case "rating":
		return b.Rating
	case "notes":
		return b.Notes
	case "added":
		return b.Added
	case "started":
		return b.Started
	case "finished":
		return b.Finished
	default:
		return ""
	}
}

func (b *BookRecord) GetTitle() string {
	return b.Title
}

func (b *BookRecord) GetDescription() string {
	var parts []string

	if b.Author != "" {
		parts = append(parts, "by "+b.Author)
	}

	if b.Status != "" {
		parts = append(parts, utils.Titlecase(b.Status))
	}

	if b.Pages > 0 {
		parts = append(parts, fmt.Sprintf("%d pages", b.Pages))
	}

	if b.Progress > 0 && b.Progress < 100 {
		parts = append(parts, fmt.Sprintf("%d%%", b.Progress))
	}

	return strings.Join(parts, " • ")
}

func (b *BookRecord) GetFilterValue() string {
	// Make books searchable by title, author, and notes
	searchable := []string{b.Title, b.Author, b.Notes}
	return strings.Join(searchable, " ")
}

// BookDataSource adapts BookRepository to work with DataList
type BookDataSource struct {
	repo   utils.TestBookRepository
	status string
}

func (b *BookDataSource) Load(ctx context.Context, opts ListOptions) ([]ListItem, error) {
	repoOpts := repo.BookListOptions{}

	if b.status != "" {
		repoOpts.Status = b.status
	}

	if opts.Search != "" {
		// Simple search in title/author (could be enhanced)
		repoOpts.Search = opts.Search
	}

	if opts.Limit > 0 {
		repoOpts.Limit = opts.Limit
	}

	books, err := b.repo.List(ctx, repoOpts)
	if err != nil {
		return nil, err
	}

	items := make([]ListItem, len(books))
	for i, book := range books {
		items[i] = &BookRecord{Book: book}
	}

	return items, nil
}

func (b *BookDataSource) Count(ctx context.Context, opts ListOptions) (int, error) {
	// For simplicity, load all and count (could be optimized)
	items, err := b.Load(ctx, opts)
	if err != nil {
		return 0, err
	}
	return len(items), nil
}

func (b *BookDataSource) Search(ctx context.Context, query string, opts ListOptions) ([]ListItem, error) {
	// Set search in options and use regular Load
	opts.Search = query
	return b.Load(ctx, opts)
}

// formatBookForView formats a book for detailed viewing
func formatBookForView(book *models.Book) string {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("# %s\n\n", book.Title))

	if book.Author != "" {
		content.WriteString(fmt.Sprintf("**Author:** %s\n", book.Author))
	}

	if book.Status != "" {
		content.WriteString(fmt.Sprintf("**Status:** %s\n", utils.Titlecase(book.Status)))
	}

	if book.Progress > 0 {
		content.WriteString(fmt.Sprintf("**Progress:** %d%%\n", book.Progress))
	}

	if book.Pages > 0 {
		content.WriteString(fmt.Sprintf("**Pages:** %d\n", book.Pages))
	}

	if book.Rating > 0 {
		content.WriteString(fmt.Sprintf("**Rating:** %.1f/5\n", book.Rating))
	}

	content.WriteString(fmt.Sprintf("**Added:** %s\n", book.Added.Format("2006-01-02 15:04")))

	if book.Started != nil {
		content.WriteString(fmt.Sprintf("**Started:** %s\n", book.Started.Format("2006-01-02 15:04")))
	}

	if book.Finished != nil {
		content.WriteString(fmt.Sprintf("**Finished:** %s\n", book.Finished.Format("2006-01-02 15:04")))
	}

	if book.Notes != "" {
		content.WriteString(fmt.Sprintf("\n**Notes:**\n%s\n", book.Notes))
	}

	return content.String()
}

// NewBookDataList creates a new DataList for browsing books
func NewBookDataList(repo utils.TestBookRepository, opts DataListOptions, status string) *DataList {
	if opts.Title == "" {
		opts.Title = "Books"
	}

	// Enable search functionality
	opts.ShowSearch = true
	opts.Searchable = true

	// Set up view handler for book details
	if opts.ViewHandler == nil {
		opts.ViewHandler = func(item ListItem) string {
			if bookRecord, ok := item.(*BookRecord); ok {
				return formatBookForView(bookRecord.Book)
			}
			return "Unable to display book"
		}
	}

	source := &BookDataSource{
		repo:   repo,
		status: status,
	}

	return NewDataList(source, opts)
}

// NewBookListFromList creates a BookList-compatible interface using DataList
func NewBookListFromList(repo utils.TestBookRepository, output io.Writer, input io.Reader, static bool, status string) *DataList {
	opts := DataListOptions{
		Output: output,
		Input:  input,
		Static: static,
		Title:  "Books",
	}
	return NewBookDataList(repo, opts, status)
}
