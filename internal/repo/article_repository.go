package repo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/stormlightlabs/noteleaf/internal/models"
	"github.com/stormlightlabs/noteleaf/internal/services"
)

func ArticleNotFoundError(id int64) error {
	return fmt.Errorf("article with id %d not found", id)
}

// ArticleRepository provides database operations for articles
type ArticleRepository struct {
	db *sql.DB
}

// NewArticleRepository creates a new article repository
func NewArticleRepository(db *sql.DB) *ArticleRepository {
	return &ArticleRepository{db: db}
}

// ArticleListOptions defines filtering options for listing articles
type ArticleListOptions struct {
	URL      string
	Title    string
	Author   string
	DateFrom string
	DateTo   string
	Limit    int
	Offset   int
}

// scanArticle scans a database row into an Article model
func (r *ArticleRepository) scanArticle(s scanner) (*models.Article, error) {
	var article models.Article
	err := s.Scan(&article.ID, &article.URL, &article.Title, &article.Author, &article.Date,
		&article.MarkdownPath, &article.HTMLPath, &article.Created, &article.Modified)
	if err != nil {
		return nil, err
	}
	return &article, nil
}

// queryOne executes a query that returns a single article
func (r *ArticleRepository) queryOne(ctx context.Context, query string, args ...any) (*models.Article, error) {
	row := r.db.QueryRowContext(ctx, query, args...)
	article, err := r.scanArticle(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("article not found")
		}
		return nil, fmt.Errorf("failed to scan article: %w", err)
	}
	return article, nil
}

// queryMany executes a query that returns multiple articles
func (r *ArticleRepository) queryMany(ctx context.Context, query string, args ...any) ([]*models.Article, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query articles: %w", err)
	}
	defer rows.Close()

	var articles []*models.Article
	for rows.Next() {
		article, err := r.scanArticle(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan article: %w", err)
		}
		articles = append(articles, article)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over articles: %w", err)
	}

	return articles, nil
}

// buildListQuery constructs a query and arguments for the List method
func (r *ArticleRepository) buildListQuery(opts *ArticleListOptions) (string, []any) {
	query := queryArticlesList
	var conditions []string
	var args []any

	if opts != nil {
		if opts.URL != "" {
			conditions = append(conditions, "url LIKE ?")
			args = append(args, "%"+opts.URL+"%")
		}
		if opts.Title != "" {
			conditions = append(conditions, "title LIKE ?")
			args = append(args, "%"+opts.Title+"%")
		}
		if opts.Author != "" {
			conditions = append(conditions, "author LIKE ?")
			args = append(args, "%"+opts.Author+"%")
		}
		if opts.DateFrom != "" {
			conditions = append(conditions, "date >= ?")
			args = append(args, opts.DateFrom)
		}
		if opts.DateTo != "" {
			conditions = append(conditions, "date <= ?")
			args = append(args, opts.DateTo)
		}
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY created DESC"

	if opts != nil && opts.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, opts.Limit)
		if opts.Offset > 0 {
			query += " OFFSET ?"
			args = append(args, opts.Offset)
		}
	}

	return query, args
}

// buildCountQuery constructs a count query and arguments
func (r *ArticleRepository) buildCountQuery(opts *ArticleListOptions) (string, []any) {
	query := queryArticlesCount
	var conditions []string
	var args []any

	if opts != nil {
		if opts.URL != "" {
			conditions = append(conditions, "url LIKE ?")
			args = append(args, "%"+opts.URL+"%")
		}
		if opts.Title != "" {
			conditions = append(conditions, "title LIKE ?")
			args = append(args, "%"+opts.Title+"%")
		}
		if opts.Author != "" {
			conditions = append(conditions, "author LIKE ?")
			args = append(args, "%"+opts.Author+"%")
		}
		if opts.DateFrom != "" {
			conditions = append(conditions, "date >= ?")
			args = append(args, opts.DateFrom)
		}
		if opts.DateTo != "" {
			conditions = append(conditions, "date <= ?")
			args = append(args, opts.DateTo)
		}
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	return query, args
}

// Create stores a new article and returns its assigned ID
func (r *ArticleRepository) Create(ctx context.Context, article *models.Article) (int64, error) {
	if err := r.Validate(article); err != nil {
		return 0, err
	}

	now := time.Now()
	article.Created = now
	article.Modified = now

	result, err := r.db.ExecContext(ctx, queryArticleInsert,
		article.URL, article.Title, article.Author, article.Date,
		article.MarkdownPath, article.HTMLPath, article.Created, article.Modified)
	if err != nil {
		return 0, fmt.Errorf("failed to insert article: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	article.ID = id
	return id, nil
}

// Get retrieves an article by its ID
func (r *ArticleRepository) Get(ctx context.Context, id int64) (*models.Article, error) {
	article, err := r.queryOne(ctx, queryArticleByID, id)
	if err != nil {
		return nil, ArticleNotFoundError(id)
	}
	return article, nil
}

// GetByURL retrieves an article by its URL
func (r *ArticleRepository) GetByURL(ctx context.Context, url string) (*models.Article, error) {
	article, err := r.queryOne(ctx, queryArticleByURL, url)
	if err != nil {
		return nil, fmt.Errorf("article with url %s not found", url)
	}
	return article, nil
}

// Update modifies an existing article
func (r *ArticleRepository) Update(ctx context.Context, article *models.Article) error {
	if err := r.Validate(article); err != nil {
		return err
	}

	article.Modified = time.Now()

	result, err := r.db.ExecContext(ctx, queryArticleUpdate,
		article.Title, article.Author, article.Date, article.MarkdownPath,
		article.HTMLPath, article.Modified, article.ID)
	if err != nil {
		return fmt.Errorf("failed to update article: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ArticleNotFoundError(article.ID)
	}

	return nil
}

// Delete removes an article from the database
func (r *ArticleRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, queryArticleDelete, id)
	if err != nil {
		return fmt.Errorf("failed to delete article: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ArticleNotFoundError(id)
	}

	return nil
}

// List retrieves articles with optional filtering
func (r *ArticleRepository) List(ctx context.Context, opts *ArticleListOptions) ([]*models.Article, error) {
	query, args := r.buildListQuery(opts)
	return r.queryMany(ctx, query, args...)
}

// Count returns the total number of articles matching the given options
func (r *ArticleRepository) Count(ctx context.Context, opts *ArticleListOptions) (int64, error) {
	query, args := r.buildCountQuery(opts)

	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count articles: %w", err)
	}

	return count, nil
}

// Validate validates a model using the validation service
func (r *ArticleRepository) Validate(model models.Model) error {
	article, ok := (model).(*models.Article)
	if !ok {
		return services.ValidationError{
			Field:   "model",
			Message: "expected Article model",
		}
	}

	validator := services.NewValidator()

	validator.Check(services.RequiredString("URL", article.URL))
	validator.Check(services.RequiredString("Title", article.Title))
	validator.Check(services.RequiredString("MarkdownPath", article.MarkdownPath))
	validator.Check(services.RequiredString("HTMLPath", article.HTMLPath))

	validator.Check(services.ValidURL("URL", article.URL))

	validator.Check(services.ValidFilePath("MarkdownPath", article.MarkdownPath))
	validator.Check(services.ValidFilePath("HTMLPath", article.HTMLPath))

	if article.Date != "" {
		validator.Check(services.ValidDate("Date", article.Date))
	}

	validator.Check(services.StringLength("Title", article.Title, 1, 500))
	validator.Check(services.StringLength("Author", article.Author, 0, 200))

	if article.ID > 0 {
		validator.Check(services.PositiveID("ID", article.ID))
	}

	if !article.Created.IsZero() && !article.Modified.IsZero() {
		if article.Created.After(article.Modified) {
			validator.Check(services.ValidationError{
				Field:   "Created",
				Message: "cannot be after Modified timestamp",
			})
		}
	}

	return validator.Errors()
}
