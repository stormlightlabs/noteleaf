package repo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"stormlightlabs.org/noteleaf/internal/models"
)

// MovieRepository provides database operations for movies
type MovieRepository struct {
	db *sql.DB
}

// NewMovieRepository creates a new movie repository
func NewMovieRepository(db *sql.DB) *MovieRepository {
	return &MovieRepository{db: db}
}

// Create stores a new movie and returns its assigned ID
func (r *MovieRepository) Create(ctx context.Context, movie *models.Movie) (int64, error) {
	now := time.Now()
	movie.Added = now

	query := `
		INSERT INTO movies (title, year, status, rating, notes, added, watched)
		VALUES (?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.ExecContext(ctx, query,
		movie.Title, movie.Year, movie.Status, movie.Rating, movie.Notes, movie.Added, movie.Watched)
	if err != nil {
		return 0, fmt.Errorf("failed to insert movie: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	movie.ID = id
	return id, nil
}

// Get retrieves a movie by ID
func (r *MovieRepository) Get(ctx context.Context, id int64) (*models.Movie, error) {
	query := `
		SELECT id, title, year, status, rating, notes, added, watched
		FROM movies WHERE id = ?`

	movie := &models.Movie{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&movie.ID, &movie.Title, &movie.Year, &movie.Status, &movie.Rating,
		&movie.Notes, &movie.Added, &movie.Watched)
	if err != nil {
		return nil, fmt.Errorf("failed to get movie: %w", err)
	}

	return movie, nil
}

// Update modifies an existing movie
func (r *MovieRepository) Update(ctx context.Context, movie *models.Movie) error {
	query := `
		UPDATE movies SET title = ?, year = ?, status = ?, rating = ?, notes = ?, watched = ?
		WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query,
		movie.Title, movie.Year, movie.Status, movie.Rating, movie.Notes, movie.Watched, movie.ID)
	if err != nil {
		return fmt.Errorf("failed to update movie: %w", err)
	}

	return nil
}

// Delete removes a movie by ID
func (r *MovieRepository) Delete(ctx context.Context, id int64) error {
	query := "DELETE FROM movies WHERE id = ?"
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete movie: %w", err)
	}
	return nil
}

// List retrieves movies with optional filtering and sorting
func (r *MovieRepository) List(ctx context.Context, opts MovieListOptions) ([]*models.Movie, error) {
	query := r.buildListQuery(opts)
	args := r.buildListArgs(opts)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list movies: %w", err)
	}
	defer rows.Close()

	var movies []*models.Movie
	for rows.Next() {
		movie := &models.Movie{}
		if err := r.scanMovieRow(rows, movie); err != nil {
			return nil, err
		}
		movies = append(movies, movie)
	}

	return movies, rows.Err()
}

func (r *MovieRepository) buildListQuery(opts MovieListOptions) string {
	query := "SELECT id, title, year, status, rating, notes, added, watched FROM movies"

	var conditions []string

	if opts.Status != "" {
		conditions = append(conditions, "status = ?")
	}
	if opts.Year > 0 {
		conditions = append(conditions, "year = ?")
	}
	if opts.MinRating > 0 {
		conditions = append(conditions, "rating >= ?")
	}

	if opts.Search != "" {
		searchConditions := []string{
			"title LIKE ?",
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

func (r *MovieRepository) buildListArgs(opts MovieListOptions) []any {
	var args []any

	if opts.Status != "" {
		args = append(args, opts.Status)
	}
	if opts.Year > 0 {
		args = append(args, opts.Year)
	}
	if opts.MinRating > 0 {
		args = append(args, opts.MinRating)
	}

	if opts.Search != "" {
		searchPattern := "%" + opts.Search + "%"
		args = append(args, searchPattern, searchPattern)
	}

	return args
}

func (r *MovieRepository) scanMovieRow(rows *sql.Rows, movie *models.Movie) error {
	return rows.Scan(&movie.ID, &movie.Title, &movie.Year, &movie.Status, &movie.Rating,
		&movie.Notes, &movie.Added, &movie.Watched)
}

// Find retrieves movies matching specific conditions
func (r *MovieRepository) Find(ctx context.Context, conditions MovieListOptions) ([]*models.Movie, error) {
	return r.List(ctx, conditions)
}

// Count returns the number of movies matching conditions
func (r *MovieRepository) Count(ctx context.Context, opts MovieListOptions) (int64, error) {
	query := "SELECT COUNT(*) FROM movies"
	args := []any{}

	var conditions []string

	if opts.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, opts.Status)
	}
	if opts.Year > 0 {
		conditions = append(conditions, "year = ?")
		args = append(args, opts.Year)
	}
	if opts.MinRating > 0 {
		conditions = append(conditions, "rating >= ?")
		args = append(args, opts.MinRating)
	}

	if opts.Search != "" {
		searchConditions := []string{
			"title LIKE ?",
			"notes LIKE ?",
		}
		conditions = append(conditions, fmt.Sprintf("(%s)", strings.Join(searchConditions, " OR ")))
		searchPattern := "%" + opts.Search + "%"
		args = append(args, searchPattern, searchPattern)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count movies: %w", err)
	}

	return count, nil
}

// GetQueued retrieves all movies in the queue
func (r *MovieRepository) GetQueued(ctx context.Context) ([]*models.Movie, error) {
	return r.List(ctx, MovieListOptions{Status: "queued"})
}

// GetWatched retrieves all watched movies
func (r *MovieRepository) GetWatched(ctx context.Context) ([]*models.Movie, error) {
	return r.List(ctx, MovieListOptions{Status: "watched"})
}

// MarkWatched marks a movie as watched
func (r *MovieRepository) MarkWatched(ctx context.Context, id int64) error {
	movie, err := r.Get(ctx, id)
	if err != nil {
		return err
	}

	now := time.Now()
	movie.Status = "watched"
	movie.Watched = &now

	return r.Update(ctx, movie)
}

// MovieListOptions defines options for listing movies
type MovieListOptions struct {
	Status    string
	Year      int
	MinRating float64
	Search    string
	SortBy    string
	SortOrder string
	Limit     int
	Offset    int
}
