package handlers

import (
	"context"
	"fmt"
	"os"
	"slices"
	"time"

	"github.com/stormlightlabs/noteleaf/internal/models"
	"github.com/stormlightlabs/noteleaf/internal/repo"
	"github.com/stormlightlabs/noteleaf/internal/services"
	"github.com/stormlightlabs/noteleaf/internal/store"
	"github.com/stormlightlabs/noteleaf/internal/ui"
)

// BookHandler handles all book-related commands
type BookHandler struct {
	db      *store.Database
	config  *store.Config
	repos   *repo.Repositories
	service *services.BookService
}

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
	service := services.NewBookService()

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

// SearchAndAdd searches for books and allows user to select and add to queue
func SearchAndAdd(ctx context.Context, args []string) error {
	handler, err := NewBookHandler()
	if err != nil {
		return fmt.Errorf("failed to initialize book handler: %w", err)
	}
	defer handler.Close()

	return handler.searchAndAdd(ctx, args)
}

// SearchAndAddWithOptions searches for books with interactive option
func SearchAndAddWithOptions(ctx context.Context, args []string, interactive bool) error {
	handler, err := NewBookHandler()
	if err != nil {
		return fmt.Errorf("failed to initialize book handler: %w", err)
	}
	defer handler.Close()

	return handler.searchAndAddWithOptions(ctx, args, interactive)
}

func (h *BookHandler) searchAndAdd(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: book add <search query> [-i for interactive mode]")
	}

	interactive := false
	searchArgs := args
	if len(args) > 0 && args[len(args)-1] == "-i" {
		interactive = true
		searchArgs = args[:len(args)-1]
	}

	if len(searchArgs) == 0 {
		return fmt.Errorf("search query cannot be empty")
	}

	query := searchArgs[0]
	if len(searchArgs) > 1 {
		for _, arg := range searchArgs[1:] {
			query += " " + arg
		}
	}

	return h.searchAndAddWithOptions(ctx, searchArgs, interactive)
}

func (h *BookHandler) searchAndAddWithOptions(ctx context.Context, args []string, interactive bool) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: book add <search query>")
	}

	query := args[0]
	if len(args) > 1 {
		for _, arg := range args[1:] {
			query += " " + arg
		}
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
	if _, err := fmt.Scanf("%d", &choice); err != nil {
		return fmt.Errorf("invalid input")
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

// ListBooks lists all books in the queue
func ListBooks(ctx context.Context, args []string) error {
	handler, err := NewBookHandler()
	if err != nil {
		return fmt.Errorf("failed to initialize book handler: %w", err)
	}
	defer handler.Close()

	return handler.listBooks(ctx, args)
}

func (h *BookHandler) listBooks(ctx context.Context, args []string) error {
	status := "queued"
	if len(args) > 0 {
		switch args[0] {
		case "all", "--all", "-a":
			status = ""
		case "reading", "--reading", "-r":
			status = "reading"
		case "finished", "--finished", "-f":
			status = "finished"
		case "queued", "--queued", "-q":
			status = "queued"
		}
	}

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

func UpdateBookStatus(ctx context.Context, args []string) error {
	handler, err := NewBookHandler()
	if err != nil {
		return fmt.Errorf("failed to initialize book handler: %w", err)
	}
	defer handler.Close()

	return handler.updateBookStatus(ctx, args)
}

func (h *BookHandler) updateBookStatus(ctx context.Context, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: book update <id> <status>")
	}

	var bookID int64
	if _, err := fmt.Sscanf(args[0], "%d", &bookID); err != nil {
		return fmt.Errorf("invalid book ID: %s", args[0])
	}

	status := args[1]
	validStatuses := []string{"queued", "reading", "finished", "removed"}
	valid := slices.Contains(validStatuses, status)
	if !valid {
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

// UpdateBookProgress updates a book's reading progress percentage
func UpdateBookProgress(ctx context.Context, args []string) error {
	handler, err := NewBookHandler()
	if err != nil {
		return fmt.Errorf("failed to initialize book handler: %w", err)
	}
	defer handler.Close()

	return handler.updateBookProgress(ctx, args)
}

func (h *BookHandler) updateBookProgress(ctx context.Context, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: book progress <id> <percentage>")
	}

	var bookID int64
	if _, err := fmt.Sscanf(args[0], "%d", &bookID); err != nil {
		return fmt.Errorf("invalid book ID: %s", args[0])
	}

	var progress int
	if _, err := fmt.Sscanf(args[1], "%d", &progress); err != nil {
		return fmt.Errorf("invalid progress percentage: %s", args[1])
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
