package handlers

import (
	"context"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/stormlightlabs/noteleaf/internal/models"
	"github.com/stormlightlabs/noteleaf/internal/repo"
	"github.com/stormlightlabs/noteleaf/internal/services"
	"github.com/stormlightlabs/noteleaf/internal/store"
	"github.com/stormlightlabs/noteleaf/internal/ui"
)

// BookHandler handles all book-related commands
//
// Implements MediaHandler interface for polymorphic media handling
type BookHandler struct {
	db      *store.Database
	config  *store.Config
	repos   *repo.Repositories
	service *services.BookService
	reader  io.Reader
}

// Ensure BookHandler implements MediaHandler interface
var _ MediaHandler = (*BookHandler)(nil)

// NewBookHandler creates a new book handler
func NewBookHandler() (*BookHandler, error) {
	db, err := store.NewDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	config, err := store.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	repos := repo.NewRepositories(db.DB)
	service := services.NewBookService(services.OpenLibraryBaseURL)

	return &BookHandler{
		db:      db,
		config:  config,
		repos:   repos,
		service: service,
	}, nil
}

// Close cleans up resources
func (h *BookHandler) Close() error {
	if err := h.service.Close(); err != nil {
		return fmt.Errorf("failed to close service: %w", err)
	}
	return h.db.Close()
}

// SetInputReader sets the input reader
func (h *BookHandler) SetInputReader(reader io.Reader) {
	h.reader = reader
}

func (h *BookHandler) printBook(book *models.Book) {
	fmt.Printf("[%d] %s", book.ID, book.Title)

	if book.Author != "" {
		fmt.Printf(" by %s", book.Author)
	}

	if book.Status != "queued" {
		fmt.Printf(" (%s)", book.Status)
	}

	if book.Progress > 0 {
		fmt.Printf(" [%d%%]", book.Progress)
	}

	if book.Rating > 0 {
		fmt.Printf(" ★%.1f", book.Rating)
	}

	fmt.Println()

	if book.Notes != "" {
		notes := book.Notes
		if len(notes) > 80 {
			notes = notes[:77] + "..."
		}
		fmt.Printf("    %s\n", notes)
	}

	fmt.Println()
}

// SearchAndAdd searches for books and allows user to select and add to queue
func (h *BookHandler) SearchAndAdd(ctx context.Context, query string, interactive bool) error {
	if query == "" {
		return fmt.Errorf("search query cannot be empty")
	}

	if interactive {
		bookList := ui.NewBookListFromList(h.repos.Books, os.Stdout, os.Stdin, false, "")
		return bookList.Browse(ctx)
	}

	fmt.Printf("Searching for books: %s\n", query)
	fmt.Print("Loading...")

	results, err := h.service.Search(ctx, query, 1, 5)
	if err != nil {
		fmt.Println(" failed!")
		return fmt.Errorf("search failed: %w", err)
	}

	fmt.Println(" done!")
	fmt.Println()

	if len(results) == 0 {
		fmt.Println("No books found.")
		return nil
	}

	fmt.Printf("Found %d result(s):\n\n", len(results))
	for i, result := range results {
		if book, ok := (*result).(*models.Book); ok {
			fmt.Printf("[%d] %s", i+1, book.Title)
			if book.Author != "" {
				fmt.Printf(" by %s", book.Author)
			}
			if book.Notes != "" {
				notes := book.Notes
				if len(notes) > 80 {
					notes = notes[:77] + "..."
				}
				fmt.Printf("\n    %s", notes)
			}
			fmt.Println()
		}
	}

	fmt.Print("\nEnter number to add (1-", len(results), "), or 0 to cancel: ")

	var choice int
	if h.reader != nil {
		if _, err := fmt.Fscanf(h.reader, "%d", &choice); err != nil {
			return fmt.Errorf("invalid input")
		}
	} else {
		if _, err := fmt.Scanf("%d", &choice); err != nil {
			return fmt.Errorf("invalid input")
		}
	}

	if choice == 0 {
		fmt.Println("Cancelled.")
		return nil
	}

	if choice < 1 || choice > len(results) {
		return fmt.Errorf("invalid choice: %d", choice)
	}

	selectedBook, ok := (*results[choice-1]).(*models.Book)
	if !ok {
		return fmt.Errorf("error processing selected book")
	}

	if _, err := h.repos.Books.Create(ctx, selectedBook); err != nil {
		return fmt.Errorf("failed to add book: %w", err)
	}

	fmt.Printf("✓ Added book: %s", selectedBook.Title)
	if selectedBook.Author != "" {
		fmt.Printf(" by %s", selectedBook.Author)
	}
	fmt.Println()

	return nil
}

// List lists all books with status filtering
func (h *BookHandler) List(ctx context.Context, status string) error {
	var books []*models.Book
	var err error

	if status == "" {
		books, err = h.repos.Books.List(ctx, repo.BookListOptions{})
		if err != nil {
			return fmt.Errorf("failed to list books: %w", err)
		}
	} else {
		switch status {
		case "queued":
			books, err = h.repos.Books.GetQueued(ctx)
		case "reading":
			books, err = h.repos.Books.GetReading(ctx)
		case "finished":
			books, err = h.repos.Books.GetFinished(ctx)
		}
		if err != nil {
			return fmt.Errorf("failed to get %s books: %w", status, err)
		}
	}

	if len(books) == 0 {
		if status == "" {
			fmt.Println("No books found")
		} else {
			fmt.Printf("No %s books found\n", status)
		}
		return nil
	}

	fmt.Printf("Found %d book(s):\n\n", len(books))
	for _, book := range books {
		h.printBook(book)
	}

	return nil
}

// UpdateStatus changes the status of a [models.Book]
func (h *BookHandler) UpdateStatus(ctx context.Context, id, status string) error {
	bookID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid book ID: %s", id)
	}

	validStatuses := []string{"queued", "reading", "finished", "removed"}
	if !slices.Contains(validStatuses, status) {
		return fmt.Errorf("invalid status: %s (valid: %v)", status, validStatuses)
	}

	book, err := h.repos.Books.Get(ctx, bookID)
	if err != nil {
		return fmt.Errorf("failed to get book: %w", err)
	}

	book.Status = status
	if status == "reading" && book.Started == nil {
		now := time.Now()
		book.Started = &now
	}
	if status == "finished" && book.Finished == nil {
		now := time.Now()
		book.Finished = &now
		book.Progress = 100
	}

	if err := h.repos.Books.Update(ctx, book); err != nil {
		return fmt.Errorf("failed to update book: %w", err)
	}

	fmt.Printf("Book status updated: %s -> %s\n", book.Title, status)
	return nil
}

// UpdateProgress updates a [models.Book]'s reading progress percentage
func (h *BookHandler) UpdateProgress(ctx context.Context, id string, progress int) error {
	bookID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid book ID: %s", id)
	}

	if progress < 0 || progress > 100 {
		return fmt.Errorf("progress must be between 0 and 100, got %d", progress)
	}

	book, err := h.repos.Books.Get(ctx, bookID)
	if err != nil {
		return fmt.Errorf("failed to get book: %w", err)
	}

	book.Progress = progress

	if progress == 0 && book.Status == "reading" {
		book.Status = "queued"
		book.Started = nil
	} else if progress > 0 && book.Status == "queued" {
		book.Status = "reading"
		if book.Started == nil {
			now := time.Now()
			book.Started = &now
		}
	} else if progress == 100 {
		book.Status = "finished"
		if book.Finished == nil {
			now := time.Now()
			book.Finished = &now
		}
	}

	if err := h.repos.Books.Update(ctx, book); err != nil {
		return fmt.Errorf("failed to update book progress: %w", err)
	}

	fmt.Printf("Book progress updated: %s -> %d%%", book.Title, progress)
	if book.Status != "queued" {
		fmt.Printf(" (%s)", book.Status)
	}
	fmt.Println()
	return nil
}

// Remove removes a book from the queue
func (h *BookHandler) Remove(ctx context.Context, id string) error {
	bookID, err := ParseID(id, "book")
	if err != nil {
		return err
	}

	book, err := h.repos.Books.Get(ctx, bookID)
	if err != nil {
		return fmt.Errorf("book %d not found: %w", bookID, err)
	}

	if err := h.repos.Books.Delete(ctx, bookID); err != nil {
		return fmt.Errorf("failed to remove book: %w", err)
	}

	fmt.Printf("✓ Removed book: %s", book.Title)
	if book.Author != "" {
		fmt.Printf(" by %s", book.Author)
	}
	fmt.Println()

	return nil
}
