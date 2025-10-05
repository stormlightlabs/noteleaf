package repo

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/stormlightlabs/noteleaf/internal/models"
)

// MediaConfig defines configuration for a media repository
type MediaConfig[T models.Model] struct {
	TableName     string                             // TableName is the database table name (e.g., "books", "movies", "tv_shows")
	New           func() T                           // New creates a new zero-value instance of T
	Scan          func(rows *sql.Rows, item T) error // Scan reads a database row into a model instance
	ScanSingle    func(row *sql.Row, item T) error   // ScanSingle reads a single row from QueryRow into a model instance
	InsertColumns string                             // InsertColumns returns the column names for INSERT statements
	UpdateColumns string                             // UpdateColumns returns the SET clause for UPDATE statements (without WHERE)
	InsertValues  func(item T) []any                 // InsertValues extracts values from a model for INSERT
	UpdateValues  func(item T) []any                 // UpdateValues extracts values from a model for UPDATE (item values + ID)
}

// BaseMediaRepository provides shared CRUD operations for media types
type BaseMediaRepository[T models.Model] struct {
	db     *sql.DB
	config MediaConfig[T]
}

// NewBaseMediaRepository creates a new base media repository
func NewBaseMediaRepository[T models.Model](db *sql.DB, config MediaConfig[T]) *BaseMediaRepository[T] {
	return &BaseMediaRepository[T]{
		db:     db,
		config: config,
	}
}

// Create stores a new media item and returns its assigned ID
func (r *BaseMediaRepository[T]) Create(ctx context.Context, item T) (int64, error) {
	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		r.config.TableName,
		r.config.InsertColumns,
		buildPlaceholders(r.config.InsertValues(item)),
	)

	result, err := r.db.ExecContext(ctx, query, r.config.InsertValues(item)...)
	if err != nil {
		return 0, fmt.Errorf("failed to insert %s: %w", r.config.TableName, err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

// Get retrieves a media item by ID
func (r *BaseMediaRepository[T]) Get(ctx context.Context, id int64) (T, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = ?", r.config.TableName)
	row := r.db.QueryRowContext(ctx, query, id)

	item := r.config.New()
	if err := r.config.ScanSingle(row, item); err != nil {
		var zero T
		if err == sql.ErrNoRows {
			return zero, fmt.Errorf("%s with id %d not found", r.config.TableName, id)
		}
		return zero, fmt.Errorf("failed to get %s: %w", r.config.TableName, err)
	}

	return item, nil
}

// Update modifies an existing media item
func (r *BaseMediaRepository[T]) Update(ctx context.Context, item T) error {
	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE id = ?",
		r.config.TableName,
		r.config.UpdateColumns,
	)

	_, err := r.db.ExecContext(ctx, query, r.config.UpdateValues(item)...)
	if err != nil {
		return fmt.Errorf("failed to update %s: %w", r.config.TableName, err)
	}

	return nil
}

// Delete removes a media item by ID
func (r *BaseMediaRepository[T]) Delete(ctx context.Context, id int64) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", r.config.TableName)
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete %s: %w", r.config.TableName, err)
	}
	return nil
}

// ListQuery executes a custom query and scans results
func (r *BaseMediaRepository[T]) ListQuery(ctx context.Context, query string, args ...any) ([]T, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list %s: %w", r.config.TableName, err)
	}
	defer rows.Close()

	var items []T
	for rows.Next() {
		item := r.config.New()
		if err := r.config.Scan(rows, item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

// CountQuery executes a custom COUNT query
func (r *BaseMediaRepository[T]) CountQuery(ctx context.Context, query string, args ...any) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count %s: %w", r.config.TableName, err)
	}
	return count, nil
}

func buildPlaceholders(values []any) string {
	if len(values) == 0 {
		return ""
	}

	placeholders := "?"
	for i := 1; i < len(values); i++ {
		placeholders += ",?"
	}
	return placeholders
}
