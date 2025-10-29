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
)

const (
	articleUserAgent    = "curl/8.4.0"
	articleAcceptHeader = "*/*"
	articleLangHeader   = "en-US,en;q=0.8"
)

type headerRoundTripper struct {
	rt http.RoundTripper
}

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
	parser, err := articles.NewArticleParser(newArticleHTTPClient())
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

func newArticleHTTPClient() *http.Client {
	baseTransport := http.DefaultTransport

	if transport, ok := http.DefaultTransport.(*http.Transport); ok {
		baseTransport = transport.Clone()
	}

	return &http.Client{
		Timeout:   30 * time.Second,
		Transport: &headerRoundTripper{rt: baseTransport},
	}
}

func (h *headerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if h.rt == nil {
		h.rt = http.DefaultTransport
	}

	clone := req.Clone(req.Context())
	clone.Header = req.Header.Clone()

	if clone.Header.Get("User-Agent") == "" {
		clone.Header.Set("User-Agent", articleUserAgent)
	}

	if clone.Header.Get("Accept") == "" {
		clone.Header.Set("Accept", articleAcceptHeader)
	}

	if clone.Header.Get("Accept-Language") == "" {
		clone.Header.Set("Accept-Language", articleLangHeader)
	}

	if clone.Header.Get("Connection") == "" {
		clone.Header.Set("Connection", "keep-alive")
	}

	return h.rt.RoundTrip(clone)
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
	existing, err := h.repos.Articles.GetByURL(ctx, url)
	if err == nil {
		ui.Warningln("Article already exists: %s (ID: %d)", ui.TitleColorStyle.Render(existing.Title), existing.ID)
		return nil
	}

	ui.Infoln("Parsing article from: %s", url)

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

	ui.Infoln("Article saved successfully!")
	ui.Infoln("ID: %d", id)
	ui.Infoln("Title: %s", ui.TitleColorStyle.Render(article.Title))
	if article.Author != "" {
		ui.Infoln("Author: %s", ui.HeaderColorStyle.Render(article.Author))
	}
	if article.Date != "" {
		ui.Infoln("Date: %s", article.Date)
	}
	ui.Infoln("Markdown: %s", article.MarkdownPath)
	ui.Infoln("HTML: %s", article.HTMLPath)

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
		ui.Warningln("No articles found.")
		return nil
	}

	ui.Infoln("Found %d article(s):\n", len(articles))
	for _, article := range articles {
		ui.Infoln("ID: %d", article.ID)
		ui.Infoln("Title: %s", ui.TitleColorStyle.Render(article.Title))
		if article.Author != "" {
			ui.Infoln("Author: %s", ui.HeaderColorStyle.Render(article.Author))
		}
		if article.Date != "" {
			ui.Infoln("Date: %s", article.Date)
		}
		ui.Infoln("URL: %s", article.URL)
		ui.Infoln("Added: %s", article.Created.Format("2006-01-02 15:04:05"))
		ui.Plainln("---")
	}
	return nil
}

// View handles viewing an article by ID
func (h *ArticleHandler) View(ctx context.Context, id int64) error {
	article, err := h.repos.Articles.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get article: %w", err)
	}

	ui.Infoln("Title: %s", ui.TitleColorStyle.Render(article.Title))
	if article.Author != "" {
		ui.Infoln("Author: %s", ui.HeaderColorStyle.Render(article.Author))
	}
	if article.Date != "" {
		ui.Infoln("Date: %s", article.Date)
	}
	ui.Infoln("URL: %s", article.URL)
	ui.Infoln("Added: %s", article.Created.Format("2006-01-02 15:04:05"))
	ui.Infoln("Modified: %s", article.Modified.Format("2006-01-02 15:04:05"))
	ui.Newline()

	ui.Info("Markdown file: %s", article.MarkdownPath)
	if _, err := os.Stat(article.MarkdownPath); os.IsNotExist(err) {
		ui.Warning(" (file not found)")
	}

	ui.Newline()

	ui.Info("HTML file: %s", article.HTMLPath)
	if _, err := os.Stat(article.HTMLPath); os.IsNotExist(err) {
		ui.Warning(" (file not found)")
	}
	ui.Newline()

	if _, err := os.Stat(article.MarkdownPath); err == nil {
		ui.Headerln("--- Content Preview ---")
		content, err := os.ReadFile(article.MarkdownPath)
		if err == nil {
			lines := strings.Split(string(content), "\n")
			previewLines := min(len(lines), 20)

			for i := range previewLines {
				ui.Plainln("%v", lines[i])
			}

			if len(lines) > previewLines {
				ui.Plainln("\n... (%d more lines)", len(lines)-previewLines)
				ui.Plainln("Read full content: %s", article.MarkdownPath)
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
			ui.Warningln("Warning: failed to remove markdown file: %v", rmErr)
		}
	}

	if _, err := os.Stat(article.HTMLPath); err == nil {
		if rmErr := os.Remove(article.HTMLPath); rmErr != nil {
			ui.Warningln("Warning: failed to remove HTML file: %v", rmErr)
		}
	}

	ui.Titleln("Article removed: %s (ID: %d)", article.Title, id)
	return nil
}

// Help shows supported domains (to complement default cobra/fang help)
func (h *ArticleHandler) Help() error {
	domains := h.parser.GetSupportedDomains()

	ui.Newline()

	if len(domains) > 0 {
		ui.Headerln("Supported sites (%d):", len(domains))
		for _, domain := range domains {
			ui.Plainln("  - %s", domain)
		}
	} else {
		ui.Plainln("No parsing rules loaded.")
	}

	ui.Newline()
	dir, err := h.getStorageDirectory()
	if err != nil {
		return fmt.Errorf("failed to get storage directory: %w", err)
	}
	ui.Headerln("%s %s", ui.HeaderColorStyle.Render("Storage directory:"), dir)

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

func (h *ArticleHandler) getStorageDirectory() (string, error) {
	if h.config.ArticlesDir != "" {
		return h.config.ArticlesDir, nil
	}

	dataDir, err := store.GetDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, "articles"), nil
}
