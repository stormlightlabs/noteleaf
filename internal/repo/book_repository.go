package repo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/stormlightlabs/noteleaf/internal/models"
)

// BookRepository provides database operations for books
//
// Uses BaseMediaRepository for common CRUD operations.
// TODO: Implement Repository interface (Validate method) similar to ArticleRepository
type BookRepository struct {
	*BaseMediaRepository[*models.Book]
	db *sql.DB
}

// NewBookRepository creates a new book repository
func NewBookRepository(db *sql.DB) *BookRepository {
	config := MediaConfig[*models.Book]{
		TableName:     "books",
		New:           func() *models.Book { return &models.Book{} },
		InsertColumns: "title, author, status, progress, pages, rating, notes, added, started, finished",
		UpdateColumns: "title = ?, author = ?, status = ?, progress = ?, pages = ?, rating = ?, notes = ?, started = ?, finished = ?",
		InsertValues: func(book *models.Book) []any {
			return []any{book.Title, book.Author, book.Status, book.Progress, book.Pages, book.Rating, book.Notes, book.Added, book.Started, book.Finished}
		},
		UpdateValues: func(book *models.Book) []any {
			return []any{book.Title, book.Author, book.Status, book.Progress, book.Pages, book.Rating, book.Notes, book.Started, book.Finished, book.ID}
		},
		Scan: func(rows *sql.Rows, book *models.Book) error {
			return scanBookRow(rows, book)
		},
		ScanSingle: func(row *sql.Row, book *models.Book) error {
			return scanBookRowSingle(row, book)
		},
	}

	return &BookRepository{
		BaseMediaRepository: NewBaseMediaRepository(db, config),
		db:                  db,
	}
}

// Create stores a new book and returns its assigned ID
func (r *BookRepository) Create(ctx context.Context, book *models.Book) (int64, error) {
	now := time.Now()
	book.Added = now

	id, err := r.BaseMediaRepository.Create(ctx, book)
	if err != nil {
		return 0, err
	}

	book.ID = id
	return id, nil
}

// List retrieves books with optional filtering and sorting
func (r *BookRepository) List(ctx context.Context, opts BookListOptions) ([]*models.Book, error) {
	query := r.buildListQuery(opts)
	args := r.buildListArgs(opts)

	items, err := r.BaseMediaRepository.ListQuery(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r *BookRepository) buildListQuery(opts BookListOptions) string {
	query := "SELECT id, title, author, status, progress, pages, rating, notes, added, started, finished FROM books"

	var conditions []string

	if opts.Status != "" {
		conditions = append(conditions, "status = ?")
	}
	if opts.Author != "" {
		conditions = append(conditions, "author = ?")
	}
	if opts.MinProgress > 0 {
		conditions = append(conditions, "progress >= ?")
	}
	if opts.MinRating > 0 {
		conditions = append(conditions, "rating >= ?")
	}

	if opts.Search != "" {
		searchConditions := []string{
			"title LIKE ?",
			"author LIKE ?",
			"notes LIKE ?",
		}
		conditions = append(conditions, fmt.Sprintf("(%s)", strings.Join(searchConditions, " OR ")))
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	if opts.SortBy != "" {
		order := "ASC"
		if strings.ToUpper(opts.SortOrder) == "DESC" {
			order = "DESC"
		}
		query += fmt.Sprintf(" ORDER BY %s %s", opts.SortBy, order)
	} else {
		query += " ORDER BY added DESC"
	}

	if opts.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", opts.Limit)
		if opts.Offset > 0 {
			query += fmt.Sprintf(" OFFSET %d", opts.Offset)
		}
	}

	return query
}

func (r *BookRepository) buildListArgs(opts BookListOptions) []any {
	var args []any

	if opts.Status != "" {
		args = append(args, opts.Status)
	}
	if opts.Author != "" {
		args = append(args, opts.Author)
	}
	if opts.MinProgress > 0 {
		args = append(args, opts.MinProgress)
	}
	if opts.MinRating > 0 {
		args = append(args, opts.MinRating)
	}

	if opts.Search != "" {
		searchPattern := "%" + opts.Search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern)
	}

	return args
}

// scanBookRow scans a database row into a book model
func scanBookRow(rows *sql.Rows, book *models.Book) error {
	var pages sql.NullInt64

	if err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.Status, &book.Progress, &pages,
		&book.Rating, &book.Notes, &book.Added, &book.Started, &book.Finished); err != nil {
		return err
	}

	if pages.Valid {
		book.Pages = int(pages.Int64)
	}

	return nil
}

// scanBookRowSingle scans a single database row into a book model
func scanBookRowSingle(row *sql.Row, book *models.Book) error {
	var pages sql.NullInt64

	if err := row.Scan(&book.ID, &book.Title, &book.Author, &book.Status, &book.Progress, &pages,
		&book.Rating, &book.Notes, &book.Added, &book.Started, &book.Finished); err != nil {
		return err
	}

	if pages.Valid {
		book.Pages = int(pages.Int64)
	}

	return nil
}

// Find retrieves books matching specific conditions
func (r *BookRepository) Find(ctx context.Context, conditions BookListOptions) ([]*models.Book, error) {
	return r.List(ctx, conditions)
}

// Count returns the number of books matching conditions
func (r *BookRepository) Count(ctx context.Context, opts BookListOptions) (int64, error) {
	query := "SELECT COUNT(*) FROM books"
	args := []any{}

	var conditions []string

	if opts.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, opts.Status)
	}
	if opts.Author != "" {
		conditions = append(conditions, "author = ?")
		args = append(args, opts.Author)
	}
	if opts.MinProgress > 0 {
		conditions = append(conditions, "progress >= ?")
		args = append(args, opts.MinProgress)
	}
	if opts.MinRating > 0 {
		conditions = append(conditions, "rating >= ?")
		args = append(args, opts.MinRating)
	}

	if opts.Search != "" {
		searchConditions := []string{
			"title LIKE ?",
			"author LIKE ?",
			"notes LIKE ?",
		}
		conditions = append(conditions, fmt.Sprintf("(%s)", strings.Join(searchConditions, " OR ")))
		searchPattern := "%" + opts.Search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	return r.BaseMediaRepository.CountQuery(ctx, query, args...)
}

// GetQueued retrieves all books in the queue
func (r *BookRepository) GetQueued(ctx context.Context) ([]*models.Book, error) {
	return r.List(ctx, BookListOptions{Status: "queued"})
}

// GetReading retrieves all books currently being read
func (r *BookRepository) GetReading(ctx context.Context) ([]*models.Book, error) {
	return r.List(ctx, BookListOptions{Status: "reading"})
}

// GetFinished retrieves all finished books
func (r *BookRepository) GetFinished(ctx context.Context) ([]*models.Book, error) {
	return r.List(ctx, BookListOptions{Status: "finished"})
}

// GetByAuthor retrieves all books by a specific author
func (r *BookRepository) GetByAuthor(ctx context.Context, author string) ([]*models.Book, error) {
	return r.List(ctx, BookListOptions{Author: author})
}

// StartReading marks a book as started
func (r *BookRepository) StartReading(ctx context.Context, id int64) error {
	book, err := r.Get(ctx, id)
	if err != nil {
		return err
	}

	now := time.Now()
	book.Status = "reading"
	book.Started = &now

	return r.Update(ctx, book)
}

// FinishReading marks a book as finished
func (r *BookRepository) FinishReading(ctx context.Context, id int64) error {
	book, err := r.Get(ctx, id)
	if err != nil {
		return err
	}

	now := time.Now()
	book.Status = "finished"
	book.Progress = 100
	book.Finished = &now

	return r.Update(ctx, book)
}

// UpdateProgress updates the reading progress of a book
func (r *BookRepository) UpdateProgress(ctx context.Context, id int64, progress int) error {
	book, err := r.Get(ctx, id)
	if err != nil {
		return err
	}

	book.Progress = progress

	if progress >= 100 {
		book.Status = "finished"
		now := time.Now()
		book.Finished = &now
	} else if progress > 0 && book.Status == "queued" {
		book.Status = "reading"
		if book.Started == nil {
			now := time.Now()
			book.Started = &now
		}
	}

	return r.Update(ctx, book)
}

// BookListOptions defines options for listing books
type BookListOptions struct {
	Status      string
	Author      string
	MinProgress int
	MinRating   float64
	Search      string
	SortBy      string
	SortOrder   string
	Limit       int
	Offset      int
}
