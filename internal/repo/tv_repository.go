package repo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"stormlightlabs.org/noteleaf/internal/models"
)

// TVRepository provides database operations for TV shows
type TVRepository struct {
	db *sql.DB
}

// NewTVRepository creates a new TV show repository
func NewTVRepository(db *sql.DB) *TVRepository {
	return &TVRepository{db: db}
}

// Create stores a new TV show and returns its assigned ID
func (r *TVRepository) Create(ctx context.Context, tvShow *models.TVShow) (int64, error) {
	now := time.Now()
	tvShow.Added = now

	query := `
		INSERT INTO tv_shows (title, season, episode, status, rating, notes, added, last_watched)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.ExecContext(ctx, query,
		tvShow.Title, tvShow.Season, tvShow.Episode, tvShow.Status, tvShow.Rating,
		tvShow.Notes, tvShow.Added, tvShow.LastWatched)
	if err != nil {
		return 0, fmt.Errorf("failed to insert TV show: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	tvShow.ID = id
	return id, nil
}

// Get retrieves a TV show by ID
func (r *TVRepository) Get(ctx context.Context, id int64) (*models.TVShow, error) {
	query := `
		SELECT id, title, season, episode, status, rating, notes, added, last_watched
		FROM tv_shows WHERE id = ?`

	tvShow := &models.TVShow{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&tvShow.ID, &tvShow.Title, &tvShow.Season, &tvShow.Episode, &tvShow.Status,
		&tvShow.Rating, &tvShow.Notes, &tvShow.Added, &tvShow.LastWatched)
	if err != nil {
		return nil, fmt.Errorf("failed to get TV show: %w", err)
	}

	return tvShow, nil
}

// Update modifies an existing TV show
func (r *TVRepository) Update(ctx context.Context, tvShow *models.TVShow) error {
	query := `
		UPDATE tv_shows SET title = ?, season = ?, episode = ?, status = ?, rating = ?,
		notes = ?, last_watched = ?
		WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query,
		tvShow.Title, tvShow.Season, tvShow.Episode, tvShow.Status, tvShow.Rating,
		tvShow.Notes, tvShow.LastWatched, tvShow.ID)
	if err != nil {
		return fmt.Errorf("failed to update TV show: %w", err)
	}

	return nil
}

// Delete removes a TV show by ID
func (r *TVRepository) Delete(ctx context.Context, id int64) error {
	query := "DELETE FROM tv_shows WHERE id = ?"
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete TV show: %w", err)
	}
	return nil
}

// List retrieves TV shows with optional filtering and sorting
func (r *TVRepository) List(ctx context.Context, opts TVListOptions) ([]*models.TVShow, error) {
	query := r.buildListQuery(opts)
	args := r.buildListArgs(opts)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list TV shows: %w", err)
	}
	defer rows.Close()

	var tvShows []*models.TVShow
	for rows.Next() {
		tvShow := &models.TVShow{}
		if err := r.scanTVShowRow(rows, tvShow); err != nil {
			return nil, err
		}
		tvShows = append(tvShows, tvShow)
	}

	return tvShows, rows.Err()
}

func (r *TVRepository) buildListQuery(opts TVListOptions) string {
	query := "SELECT id, title, season, episode, status, rating, notes, added, last_watched FROM tv_shows"

	var conditions []string

	if opts.Status != "" {
		conditions = append(conditions, "status = ?")
	}
	if opts.Title != "" {
		conditions = append(conditions, "title = ?")
	}
	if opts.Season > 0 {
		conditions = append(conditions, "season = ?")
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
		query += " ORDER BY title, season, episode"
	}

	if opts.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", opts.Limit)
		if opts.Offset > 0 {
			query += fmt.Sprintf(" OFFSET %d", opts.Offset)
		}
	}

	return query
}

func (r *TVRepository) buildListArgs(opts TVListOptions) []any {
	var args []any

	if opts.Status != "" {
		args = append(args, opts.Status)
	}
	if opts.Title != "" {
		args = append(args, opts.Title)
	}
	if opts.Season > 0 {
		args = append(args, opts.Season)
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

func (r *TVRepository) scanTVShowRow(rows *sql.Rows, tvShow *models.TVShow) error {
	return rows.Scan(&tvShow.ID, &tvShow.Title, &tvShow.Season, &tvShow.Episode, &tvShow.Status,
		&tvShow.Rating, &tvShow.Notes, &tvShow.Added, &tvShow.LastWatched)
}

// Find retrieves TV shows matching specific conditions
func (r *TVRepository) Find(ctx context.Context, conditions TVListOptions) ([]*models.TVShow, error) {
	return r.List(ctx, conditions)
}

// Count returns the number of TV shows matching conditions
func (r *TVRepository) Count(ctx context.Context, opts TVListOptions) (int64, error) {
	query := "SELECT COUNT(*) FROM tv_shows"
	args := []any{}

	var conditions []string

	if opts.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, opts.Status)
	}
	if opts.Title != "" {
		conditions = append(conditions, "title = ?")
		args = append(args, opts.Title)
	}
	if opts.Season > 0 {
		conditions = append(conditions, "season = ?")
		args = append(args, opts.Season)
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
		return 0, fmt.Errorf("failed to count TV shows: %w", err)
	}

	return count, nil
}

// GetQueued retrieves all TV shows in the queue
func (r *TVRepository) GetQueued(ctx context.Context) ([]*models.TVShow, error) {
	return r.List(ctx, TVListOptions{Status: "queued"})
}

// GetWatching retrieves all TV shows currently being watched
func (r *TVRepository) GetWatching(ctx context.Context) ([]*models.TVShow, error) {
	return r.List(ctx, TVListOptions{Status: "watching"})
}

// GetWatched retrieves all watched TV shows
func (r *TVRepository) GetWatched(ctx context.Context) ([]*models.TVShow, error) {
	return r.List(ctx, TVListOptions{Status: "watched"})
}

// GetByTitle retrieves all episodes for a specific TV show title
func (r *TVRepository) GetByTitle(ctx context.Context, title string) ([]*models.TVShow, error) {
	return r.List(ctx, TVListOptions{Title: title})
}

// GetBySeason retrieves all episodes for a specific season of a show
func (r *TVRepository) GetBySeason(ctx context.Context, title string, season int) ([]*models.TVShow, error) {
	return r.List(ctx, TVListOptions{Title: title, Season: season})
}

// MarkWatched marks a TV show episode as watched
func (r *TVRepository) MarkWatched(ctx context.Context, id int64) error {
	tvShow, err := r.Get(ctx, id)
	if err != nil {
		return err
	}

	now := time.Now()
	tvShow.Status = "watched"
	tvShow.LastWatched = &now

	return r.Update(ctx, tvShow)
}

// StartWatching marks a TV show as currently being watched
func (r *TVRepository) StartWatching(ctx context.Context, id int64) error {
	tvShow, err := r.Get(ctx, id)
	if err != nil {
		return err
	}

	now := time.Now()
	tvShow.Status = "watching"
	tvShow.LastWatched = &now

	return r.Update(ctx, tvShow)
}

// TVListOptions defines options for listing TV shows
type TVListOptions struct {
	Status    string
	Title     string
	Season    int
	MinRating float64
	Search    string
	SortBy    string
	SortOrder string
	Limit     int
	Offset    int
}
