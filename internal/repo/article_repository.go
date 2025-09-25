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

// Create stores a new article and returns its assigned ID
func (r *ArticleRepository) Create(ctx context.Context, article *models.Article) (int64, error) {
	if err := r.Validate(article); err != nil {
		return 0, err
	}

	now := time.Now()
	article.Created = now
	article.Modified = now

	query := `
		INSERT INTO articles (url, title, author, date, markdown_path, html_path, created, modified)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.ExecContext(ctx, query,
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
	query := `
		SELECT id, url, title, author, date, markdown_path, html_path, created, modified
		FROM articles WHERE id = ?`

	row := r.db.QueryRowContext(ctx, query, id)

	var article models.Article
	err := row.Scan(&article.ID, &article.URL, &article.Title, &article.Author, &article.Date,
		&article.MarkdownPath, &article.HTMLPath, &article.Created, &article.Modified)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("article with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to scan article: %w", err)
	}

	return &article, nil
}

// GetByURL retrieves an article by its URL
func (r *ArticleRepository) GetByURL(ctx context.Context, url string) (*models.Article, error) {
	query := `
		SELECT id, url, title, author, date, markdown_path, html_path, created, modified
		FROM articles WHERE url = ?`

	row := r.db.QueryRowContext(ctx, query, url)

	var article models.Article
	err := row.Scan(&article.ID, &article.URL, &article.Title, &article.Author, &article.Date,
		&article.MarkdownPath, &article.HTMLPath, &article.Created, &article.Modified)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("article with url %s not found", url)
		}
		return nil, fmt.Errorf("failed to scan article: %w", err)
	}

	return &article, nil
}

// Update modifies an existing article
func (r *ArticleRepository) Update(ctx context.Context, article *models.Article) error {
	if err := r.Validate(article); err != nil {
		return err
	}

	article.Modified = time.Now()

	query := `
		UPDATE articles
		SET title = ?, author = ?, date = ?, markdown_path = ?, html_path = ?, modified = ?
		WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query,
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
		return fmt.Errorf("article with id %d not found", article.ID)
	}

	return nil
}

// Delete removes an article from the database
func (r *ArticleRepository) Delete(ctx context.Context, id int64) error {
	query := "DELETE FROM articles WHERE id = ?"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete article: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("article with id %d not found", id)
	}

	return nil
}

// List retrieves articles with optional filtering
func (r *ArticleRepository) List(ctx context.Context, opts *ArticleListOptions) ([]*models.Article, error) {
	query := `
		SELECT id, url, title, author, date, markdown_path, html_path, created, modified
		FROM articles`

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

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query articles: %w", err)
	}
	defer rows.Close()

	var articles []*models.Article
	for rows.Next() {
		var article models.Article
		err := rows.Scan(&article.ID, &article.URL, &article.Title, &article.Author, &article.Date,
			&article.MarkdownPath, &article.HTMLPath, &article.Created, &article.Modified)
		if err != nil {
			return nil, fmt.Errorf("failed to scan article: %w", err)
		}
		articles = append(articles, &article)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over articles: %w", err)
	}

	return articles, nil
}

// Count returns the total number of articles matching the given options
func (r *ArticleRepository) Count(ctx context.Context, opts *ArticleListOptions) (int64, error) {
	query := "SELECT COUNT(*) FROM articles"

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
