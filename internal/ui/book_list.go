package ui

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/stormlightlabs/noteleaf/internal/models"
	"github.com/stormlightlabs/noteleaf/internal/repo"
	"github.com/stormlightlabs/noteleaf/internal/services"
)

// BookListOptions configures the book list UI behavior
type BookListOptions struct {
	// Output destination (stdout for interactive, buffer for testing)
	Output io.Writer
	// Input source (stdin for interactive, strings reader for testing)
	Input io.Reader
	// Enable static mode (no interactive components)
	Static bool
}

// BookList handles book search and selection UI
type BookList struct {
	service services.APIService
	repo    *repo.BookRepository
	opts    BookListOptions
}

// NewBookList creates a new book list UI component
func NewBookList(service services.APIService, repo *repo.BookRepository, opts BookListOptions) *BookList {
	if opts.Output == nil {
		opts.Output = os.Stdout
	}
	if opts.Input == nil {
		opts.Input = os.Stdin
	}
	return &BookList{
		service: service,
		repo:    repo,
		opts:    opts,
	}
}

type searchModel struct {
	query        string
	results      []*models.Book
	selected     int
	searching    bool
	err          error
	service      services.APIService
	repo         *repo.BookRepository
	opts         BookListOptions
	currentPage  int
	totalResults int
	confirmed    bool
	addedBook    *models.Book
}

type searchMsg []*models.Book
type errorMsg error
type bookAddedMsg *models.Book

func (m searchModel) Init() tea.Cmd {
	return m.searchBooks(m.query)
}

func (m searchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.results)-1 {
				m.selected++
			}
		case "enter":
			if len(m.results) > 0 && m.selected < len(m.results) {
				return m, m.addBook(m.results[m.selected])
			}
		case "n":
			if !m.searching && len(m.results) > 0 && m.currentPage*10 < m.totalResults {
				m.currentPage++
				return m, m.searchBooks(m.query)
			}
		case "p":
			if !m.searching && m.currentPage > 1 {
				m.currentPage--
				return m, m.searchBooks(m.query)
			}
		}
	case searchMsg:
		m.results = []*models.Book(msg)
		m.searching = false
		m.selected = 0
	case errorMsg:
		m.err = error(msg)
		m.searching = false
	case bookAddedMsg:
		m.addedBook = (*models.Book)(msg)
		m.confirmed = true
		return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
			return tea.Quit()
		})
	}
	return m, nil
}

func (m searchModel) View() string {
	var s strings.Builder

	style := lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)

	s.WriteString(titleStyle.Render(fmt.Sprintf("Search Results for: %s", m.query)))
	s.WriteString("\n\n")

	if m.searching {
		s.WriteString("Searching...")
		return s.String()
	}

	if m.err != nil {
		s.WriteString(fmt.Sprintf("Error: %s", m.err))
		return s.String()
	}

	if len(m.results) == 0 {
		s.WriteString("No books found")
		return s.String()
	}

	if m.confirmed && m.addedBook != nil {
		s.WriteString(fmt.Sprintf("✓ Added book: %s", m.addedBook.Title))
		if m.addedBook.Author != "" {
			s.WriteString(fmt.Sprintf(" by %s", m.addedBook.Author))
		}
		return s.String()
	}

	for i, book := range m.results {
		prefix := "  "
		if i == m.selected {
			prefix = "> "
		}

		line := fmt.Sprintf("%s%s", prefix, book.Title)
		if book.Author != "" {
			line += fmt.Sprintf(" by %s", book.Author)
		}

		if i == m.selected {
			s.WriteString(selectedStyle.Render(line))
		} else {
			s.WriteString(style.Render(line))
		}
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString("Use ↑/↓ to navigate, Enter to select, q to quit")
	if m.currentPage*10 < m.totalResults {
		s.WriteString(", n for next page")
	}
	if m.currentPage > 1 {
		s.WriteString(", p for previous page")
	}

	return s.String()
}

func (m searchModel) searchBooks(query string) tea.Cmd {
	return func() tea.Msg {
		results, err := m.service.Search(context.Background(), query, m.currentPage, 10)
		if err != nil {
			return errorMsg(err)
		}

		books := make([]*models.Book, 0, len(results))
		for _, result := range results {
			if book, ok := (*result).(*models.Book); ok {
				books = append(books, book)
			}
		}

		return searchMsg(books)
	}
}

func (m searchModel) addBook(book *models.Book) tea.Cmd {
	return func() tea.Msg {
		if _, err := m.repo.Create(context.Background(), book); err != nil {
			return errorMsg(fmt.Errorf("failed to add book: %w", err))
		}
		return bookAddedMsg(book)
	}
}

// SearchAndSelect searches for books with the given query and allows selection
func (bl *BookList) SearchAndSelect(ctx context.Context, query string) error {
	if bl.opts.Static {
		return bl.staticSelect(ctx, query)
	}

	model := searchModel{
		query:       query,
		searching:   true,
		service:     bl.service,
		repo:        bl.repo,
		opts:        bl.opts,
		currentPage: 1,
	}

	program := tea.NewProgram(model, tea.WithInput(bl.opts.Input), tea.WithOutput(bl.opts.Output))

	_, err := program.Run()
	return err
}

func (bl *BookList) staticSelect(ctx context.Context, query string) error {
	results, err := bl.service.Search(ctx, query, 1, 10)
	if err != nil {
		fmt.Fprintf(bl.opts.Output, "Error: %s\n", err)
		return err
	}

	fmt.Fprintf(bl.opts.Output, "Search Results for: %s\n\n", query)

	if len(results) == 0 {
		fmt.Fprintf(bl.opts.Output, "No books found\n")
		return nil
	}

	for i, result := range results {
		if book, ok := (*result).(*models.Book); ok {
			fmt.Fprintf(bl.opts.Output, "[%d] %s", i+1, book.Title)
			if book.Author != "" {
				fmt.Fprintf(bl.opts.Output, " by %s", book.Author)
			}
			fmt.Fprintf(bl.opts.Output, "\n")
		}
	}

	if len(results) > 0 {
		if book, ok := (*results[0]).(*models.Book); ok {
			if bl.repo != nil {
				if _, err := bl.repo.Create(ctx, book); err != nil {
					fmt.Fprintf(bl.opts.Output, "Error adding book: %s\n", err)
					return err
				}
			}
			fmt.Fprintf(bl.opts.Output, "✓ Added book: %s", book.Title)
			if book.Author != "" {
				fmt.Fprintf(bl.opts.Output, " by %s", book.Author)
			}
			fmt.Fprintf(bl.opts.Output, "\n")
		}
	}

	return nil
}

// InteractiveSearch provides an interactive search interface
func (bl *BookList) InteractiveSearch(ctx context.Context) error {
	if bl.opts.Static {
		return bl.staticSearch(ctx)
	}

	var query string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Search for books").
				Placeholder("Enter book title or author").
				Value(&query),
		),
	)

	if err := form.WithTheme(huh.ThemeCharm()).Run(); err != nil {
		return err
	}

	if strings.TrimSpace(query) == "" {
		return fmt.Errorf("search query cannot be empty")
	}

	return bl.SearchAndSelect(ctx, query)
}

func (bl *BookList) staticSearch(ctx context.Context) error {
	fmt.Fprintf(bl.opts.Output, "Search for books: test query\n")
	return bl.staticSelect(ctx, "test query")
}
