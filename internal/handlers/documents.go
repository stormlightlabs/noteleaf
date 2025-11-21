package handlers

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/stormlightlabs/noteleaf/internal/repo"
	"github.com/stormlightlabs/noteleaf/internal/store"
)

// DocumentHandler provides operations for document search
type DocumentHandler struct {
	db     *sql.DB
	repos  *repo.Repositories
	engine *repo.SearchEngine
}

// NewDocumentHandler creates a new document handler
func NewDocumentHandler() (*DocumentHandler, error) {
	db, err := store.NewDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	repos := repo.NewRepositories(db.DB)
	engine := repo.NewSearchEngine(repos.Documents)

	return &DocumentHandler{
		db:     db.DB,
		repos:  repos,
		engine: engine,
	}, nil
}

// RebuildIndex rebuilds the search index from notes
func (h *DocumentHandler) RebuildIndex(ctx context.Context) error {
	if err := h.repos.Documents.DeleteAll(ctx); err != nil {
		return fmt.Errorf("failed to clear documents: %w", err)
	}

	if err := h.repos.Documents.RebuildFromNotes(ctx, h.repos.Notes); err != nil {
		return fmt.Errorf("failed to rebuild from notes: %w", err)
	}

	if err := h.engine.Rebuild(ctx); err != nil {
		return fmt.Errorf("failed to rebuild search index: %w", err)
	}

	fmt.Println("Search index rebuilt successfully")
	return nil
}

// Search performs a TF-IDF search and displays results
func (h *DocumentHandler) Search(ctx context.Context, query string, limit int) error {
	if err := h.engine.Rebuild(ctx); err != nil {
		return fmt.Errorf("failed to rebuild index: %w", err)
	}

	results, docs, err := h.engine.SearchWithScores(ctx, query, limit)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if len(results) == 0 {
		fmt.Println("No results found")
		return nil
	}

	fmt.Printf("Found %d results:\n\n", len(results))
	for i, doc := range docs {
		score := results[i].Score
		fmt.Printf("%d. [Score: %.2f] %s\n", i+1, score, doc.Title)
		fmt.Printf("   %s\n\n", truncate(doc.Body, 100))
	}

	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
