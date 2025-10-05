package repo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/stormlightlabs/noteleaf/internal/models"
)

// MovieRepository provides database operations for movies
type MovieRepository struct {
	*BaseMediaRepository[*models.Movie]
	db *sql.DB
}

// NewMovieRepository creates a new movie repository
func NewMovieRepository(db *sql.DB) *MovieRepository {
	config := MediaConfig[*models.Movie]{
		TableName:     "movies",
		New:           func() *models.Movie { return &models.Movie{} },
		InsertColumns: "title, year, status, rating, notes, added, watched",
		UpdateColumns: "title = ?, year = ?, status = ?, rating = ?, notes = ?, watched = ?",
		InsertValues: func(movie *models.Movie) []any {
			return []any{movie.Title, movie.Year, movie.Status, movie.Rating, movie.Notes, movie.Added, movie.Watched}
		},
		UpdateValues: func(movie *models.Movie) []any {
			return []any{movie.Title, movie.Year, movie.Status, movie.Rating, movie.Notes, movie.Watched, movie.ID}
		},
		Scan: func(rows *sql.Rows, movie *models.Movie) error {
			return scanMovieRow(rows, movie)
		},
		ScanSingle: func(row *sql.Row, movie *models.Movie) error {
			return scanMovieRowSingle(row, movie)
		},
	}

	return &MovieRepository{
		BaseMediaRepository: NewBaseMediaRepository(db, config),
		db:                  db,
	}
}

// Create stores a new movie and returns its assigned ID
func (r *MovieRepository) Create(ctx context.Context, movie *models.Movie) (int64, error) {
	now := time.Now()
	movie.Added = now

	id, err := r.BaseMediaRepository.Create(ctx, movie)
	if err != nil {
		return 0, err
	}

	movie.ID = id
	return id, nil
}

// List retrieves movies with optional filtering and sorting
func (r *MovieRepository) List(ctx context.Context, opts MovieListOptions) ([]*models.Movie, error) {
	query := r.buildListQuery(opts)
	args := r.buildListArgs(opts)

	items, err := r.BaseMediaRepository.ListQuery(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return items, nil
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

// scanMovieRow scans a database row into a [models.Movie]
func scanMovieRow(rows *sql.Rows, movie *models.Movie) error {
	return rows.Scan(&movie.ID, &movie.Title, &movie.Year, &movie.Status, &movie.Rating,
		&movie.Notes, &movie.Added, &movie.Watched)
}

// scanMovieRowSingle scans a single database row into a [models.Movie]
func scanMovieRowSingle(row *sql.Row, movie *models.Movie) error {
	return row.Scan(&movie.ID, &movie.Title, &movie.Year, &movie.Status, &movie.Rating,
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
	return r.BaseMediaRepository.CountQuery(ctx, query, args...)
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
