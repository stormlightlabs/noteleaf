package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/stormlightlabs/noteleaf/internal/articles"
	"github.com/stormlightlabs/noteleaf/internal/models"
	"github.com/stormlightlabs/noteleaf/internal/repo"
	"github.com/stormlightlabs/noteleaf/internal/store"
	"github.com/stormlightlabs/noteleaf/internal/ui"
	"github.com/stormlightlabs/noteleaf/internal/utils"
)

// ArticleHandler handles all article-related commands
type ArticleHandler struct {
	db     *store.Database
	config *store.Config
	repos  *repo.Repositories
	parser articles.Parser
}

// NewArticleHandler creates a new article handler
func NewArticleHandler() (*ArticleHandler, error) {
	db, err := store.NewDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	config, err := store.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	repos := repo.NewRepositories(db.DB)

	parser, err := articles.NewArticleParser(http.DefaultClient)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize article parser: %w", err)
	}

	return &ArticleHandler{
		db:     db,
		config: config,
		repos:  repos,
		parser: parser,
	}, nil
}

// Close cleans up resources
func (h *ArticleHandler) Close() error {
	if h.db != nil {
		return h.db.Close()
	}
	return nil
}

// Add handles adding an article from a URL
func (h *ArticleHandler) Add(ctx context.Context, url string) error {
	logger := utils.GetLogger()

	existing, err := h.repos.Articles.GetByURL(ctx, url)
	if err == nil {
		fmt.Printf("Article already exists: %s (ID: %d)\n", ui.TitleColorStyle.Render(existing.Title), existing.ID)
		return nil
	}

	logger.Info("Parsing article", "url", url)
	fmt.Printf("Parsing article from: %s\n", url)

	dir, err := h.getStorageDirectory()
	if err != nil {
		return fmt.Errorf("failed to get article storage dir %w", err)
	}

	content, err := h.parser.ParseURL(url)
	if err != nil {
		return fmt.Errorf("failed to parse article: %w", err)
	}

	mdPath, htmlPath, err := h.parser.SaveArticle(content, dir)
	if err != nil {
		return fmt.Errorf("failed to save article: %w", err)
	}

	article := &models.Article{
		URL:          url,
		Title:        content.Title,
		Author:       content.Author,
		Date:         content.Date,
		MarkdownPath: mdPath,
		HTMLPath:     htmlPath,
		Created:      time.Now(),
		Modified:     time.Now(),
	}

	id, err := h.repos.Articles.Create(ctx, article)
	if err != nil {
		os.Remove(article.MarkdownPath)
		os.Remove(article.HTMLPath)
		return fmt.Errorf("failed to save article to database: %w", err)
	}

	fmt.Printf("Article saved successfully!\n")
	fmt.Printf("ID: %d\n", id)
	fmt.Printf("Title: %s\n", ui.TitleColorStyle.Render(article.Title))
	if article.Author != "" {
		fmt.Printf("Author: %s\n", ui.HeaderColorStyle.Render(article.Author))
	}
	if article.Date != "" {
		fmt.Printf("Date: %s\n", article.Date)
	}
	fmt.Printf("Markdown: %s\n", article.MarkdownPath)
	fmt.Printf("HTML: %s\n", article.HTMLPath)

	logger.Info("Article saved", "id", id, "title", article.Title)

	return nil
}

// List handles listing articles with optional filtering
func (h *ArticleHandler) List(ctx context.Context, query string, author string, limit int) error {
	opts := &repo.ArticleListOptions{
		Title:  query,
		Author: author,
		Limit:  limit,
	}

	articles, err := h.repos.Articles.List(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to list articles: %w", err)
	}

	if len(articles) == 0 {
		fmt.Println("No articles found.")
		return nil
	}

	fmt.Printf("Found %d article(s):\n\n", len(articles))

	for _, article := range articles {
		fmt.Printf("ID: %d\n", article.ID)
		fmt.Printf("Title: %s\n", ui.TitleColorStyle.Render(article.Title))
		if article.Author != "" {
			fmt.Printf("Author: %s\n", ui.HeaderColorStyle.Render(article.Author))
		}
		if article.Date != "" {
			fmt.Printf("Date: %s\n", article.Date)
		}
		fmt.Printf("URL: %s\n", article.URL)
		fmt.Printf("Added: %s\n", article.Created.Format("2006-01-02 15:04:05"))
		fmt.Println("---")
	}

	return nil
}

// View handles viewing an article by ID
func (h *ArticleHandler) View(ctx context.Context, id int64) error {

	article, err := h.repos.Articles.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get article: %w", err)
	}

	fmt.Printf("Title: %s\n", ui.TitleColorStyle.Render(article.Title))
	if article.Author != "" {
		fmt.Printf("Author: %s\n", ui.HeaderColorStyle.Render(article.Author))
	}
	if article.Date != "" {
		fmt.Printf("Date: %s\n", article.Date)
	}
	fmt.Printf("URL: %s\n", article.URL)
	fmt.Printf("Added: %s\n", article.Created.Format("2006-01-02 15:04:05"))
	fmt.Printf("Modified: %s\n", article.Modified.Format("2006-01-02 15:04:05"))
	fmt.Println()

	fmt.Printf("Markdown file: %s", article.MarkdownPath)
	if _, err := os.Stat(article.MarkdownPath); os.IsNotExist(err) {
		fmt.Printf(" (file not found)")
	}
	fmt.Println()

	fmt.Printf("HTML file: %s", article.HTMLPath)
	if _, err := os.Stat(article.HTMLPath); os.IsNotExist(err) {
		fmt.Printf(" (file not found)")
	}
	fmt.Println()

	if _, err := os.Stat(article.MarkdownPath); err == nil {
		fmt.Printf("\n%s\n", ui.HeaderColorStyle.Render("--- Content Preview ---"))
		content, err := os.ReadFile(article.MarkdownPath)
		if err == nil {
			lines := strings.Split(string(content), "\n")
			previewLines := min(len(lines), 20)

			for i := range previewLines {
				fmt.Println(lines[i])
			}

			if len(lines) > previewLines {
				fmt.Printf("\n... (%d more lines)\n", len(lines)-previewLines)
				fmt.Printf("Read full content: %s\n", article.MarkdownPath)
			}
		}
	}

	return nil
}

// Remove handles removing an article by ID
func (h *ArticleHandler) Remove(ctx context.Context, id int64) error {
	article, err := h.repos.Articles.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get article: %w", err)
	}

	err = h.repos.Articles.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to remove article from database: %w", err)
	}

	if _, err := os.Stat(article.MarkdownPath); err == nil {
		if rmErr := os.Remove(article.MarkdownPath); rmErr != nil {
			fmt.Printf("Warning: failed to remove markdown file: %v\n", rmErr)
		}
	}

	if _, err := os.Stat(article.HTMLPath); err == nil {
		if rmErr := os.Remove(article.HTMLPath); rmErr != nil {
			fmt.Printf("Warning: failed to remove HTML file: %v\n", rmErr)
		}
	}

	fmt.Printf("Article removed: %s (ID: %d)\n", ui.TitleColorStyle.Render(article.Title), id)

	return nil
}

// Help shows supported domains (to complement default cobra/fang help)
func (h *ArticleHandler) Help() error {
	domains := h.parser.GetSupportedDomains()

	fmt.Println()

	if len(domains) > 0 {
		fmt.Printf("%s\n", ui.HeaderColorStyle.Render(fmt.Sprintf("Supported sites (%d):", len(domains))))
		for _, domain := range domains {
			fmt.Printf("  - %s\n", domain)
		}
	} else {
		fmt.Println("No parsing rules loaded.")
	}

	fmt.Println()
	dir, err := h.getStorageDirectory()
	if err != nil {
		return fmt.Errorf("failed to get storage directory: %w", err)
	}
	fmt.Printf("%s %s\n", ui.HeaderColorStyle.Render("Storage directory:"), dir)

	return nil
}

// Read displays an article's content with formatted markdown rendering
func (h *ArticleHandler) Read(ctx context.Context, id int64) error {
	article, err := h.repos.Articles.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get article: %w", err)
	}

	if _, err := os.Stat(article.MarkdownPath); os.IsNotExist(err) {
		return fmt.Errorf("markdown file not found: %s", article.MarkdownPath)
	}

	content, err := os.ReadFile(article.MarkdownPath)
	if err != nil {
		return fmt.Errorf("failed to read markdown file: %w", err)
	}

	if rendered, err := renderMarkdown(string(content)); err != nil {
		return err
	} else {
		fmt.Print(rendered)
		return nil
	}

}

// TODO: Try to get from config first (could be added later)
// For now, use default ~/Documents/Leaf/
func (h *ArticleHandler) getStorageDirectory() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, "Documents", "Leaf"), nil
}
