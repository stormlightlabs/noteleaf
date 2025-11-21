package repo

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/stormlightlabs/noteleaf/internal/documents"
)

func DocumentNotFoundError(id int64) error {
	return fmt.Errorf("document with id %d not found", id)
}

// DocumentRepository provides database operations for documents
type DocumentRepository struct {
	db *sql.DB
}

// NewDocumentRepository creates a new document repository
func NewDocumentRepository(db *sql.DB) *DocumentRepository {
	return &DocumentRepository{db: db}
}

// scanDocument scans a database row into a Document model
func (r *DocumentRepository) scanDocument(s scanner) (*documents.Document, error) {
	var doc documents.Document
	err := s.Scan(&doc.ID, &doc.Title, &doc.Body, &doc.CreatedAt, &doc.DocKind)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

// queryOne executes a query that returns a single document
func (r *DocumentRepository) queryOne(ctx context.Context, query string, args ...any) (*documents.Document, error) {
	row := r.db.QueryRowContext(ctx, query, args...)
	doc, err := r.scanDocument(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("document not found")
		}
		return nil, fmt.Errorf("failed to scan document: %w", err)
	}
	return doc, nil
}

// queryMany executes a query that returns multiple documents
func (r *DocumentRepository) queryMany(ctx context.Context, query string, args ...any) ([]documents.Document, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query documents: %w", err)
	}
	defer rows.Close()

	var docs []documents.Document
	for rows.Next() {
		doc, err := r.scanDocument(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan document: %w", err)
		}
		docs = append(docs, *doc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over documents: %w", err)
	}

	return docs, nil
}

// Create stores a new document and returns its assigned ID
func (r *DocumentRepository) Create(ctx context.Context, doc *documents.Document) (int64, error) {
	result, err := r.db.ExecContext(ctx, queryDocumentInsert,
		doc.Title, doc.Body, doc.CreatedAt, doc.DocKind)
	if err != nil {
		return 0, fmt.Errorf("failed to insert document: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	doc.ID = id
	return id, nil
}

// Get retrieves a document by its ID
func (r *DocumentRepository) Get(ctx context.Context, id int64) (*documents.Document, error) {
	doc, err := r.queryOne(ctx, queryDocumentByID, id)
	if err != nil {
		return nil, DocumentNotFoundError(id)
	}
	return doc, nil
}

// Delete removes a document by its ID
func (r *DocumentRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, queryDocumentDelete, id)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return DocumentNotFoundError(id)
	}

	return nil
}

// List retrieves all documents
func (r *DocumentRepository) List(ctx context.Context) ([]documents.Document, error) {
	return r.queryMany(ctx, queryDocumentsList)
}

// ListByKind retrieves documents of a specific kind
func (r *DocumentRepository) ListByKind(ctx context.Context, kind documents.DocKind) ([]documents.Document, error) {
	return r.queryMany(ctx, queryDocumentsByKind, int64(kind))
}

// DeleteAll removes all documents from the database
func (r *DocumentRepository) DeleteAll(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, queryDocumentsDeleteAll)
	if err != nil {
		return fmt.Errorf("failed to delete all documents: %w", err)
	}
	return nil
}

// RebuildFromNotes rebuilds the documents table from notes
func (r *DocumentRepository) RebuildFromNotes(ctx context.Context, noteRepo *NoteRepository) error {
	notes, err := noteRepo.List(ctx, NoteListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list notes: %w", err)
	}

	for _, note := range notes {
		doc := &documents.Document{
			Title:     note.Title,
			Body:      note.Content,
			CreatedAt: note.Created,
			DocKind:   int64(documents.NoteDoc),
		}

		if _, err := r.Create(ctx, doc); err != nil {
			return fmt.Errorf("failed to create document from note %d: %w", note.ID, err)
		}
	}

	return nil
}

// BuildIndex creates a TF-IDF index from all documents in the database
func (r *DocumentRepository) BuildIndex(ctx context.Context) (*documents.Index, error) {
	docs, err := r.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}

	return documents.BuildIndex(docs), nil
}

// SearchEngine wraps a DocumentRepository with search capabilities
type SearchEngine struct {
	repo  *DocumentRepository
	index *documents.Index
}

// NewSearchEngine creates a new search engine with the given repository
func NewSearchEngine(repo *DocumentRepository) *SearchEngine {
	return &SearchEngine{
		repo:  repo,
		index: nil,
	}
}

// Rebuild rebuilds the search index from the database
func (se *SearchEngine) Rebuild(ctx context.Context) error {
	idx, err := se.repo.BuildIndex(ctx)
	if err != nil {
		return fmt.Errorf("failed to build index: %w", err)
	}

	se.index = idx
	return nil
}

// Search performs a TF-IDF search and returns matching documents
func (se *SearchEngine) Search(ctx context.Context, query string, limit int) ([]documents.Document, error) {
	if se.index == nil {
		return nil, fmt.Errorf("search index not initialized")
	}

	results, err := se.index.Search(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}

	docs := make([]documents.Document, 0, len(results))
	for _, result := range results {
		doc, err := se.repo.Get(ctx, result.DocID)
		if err != nil {
			return nil, fmt.Errorf("failed to get document %d: %w", result.DocID, err)
		}
		docs = append(docs, *doc)
	}

	return docs, nil
}

// SearchWithScores performs a TF-IDF search and returns results with scores
func (se *SearchEngine) SearchWithScores(ctx context.Context, query string, limit int) ([]documents.Result, []documents.Document, error) {
	if se.index == nil {
		return nil, nil, fmt.Errorf("search index not initialized")
	}

	results, err := se.index.Search(query, limit)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to search: %w", err)
	}

	docs := make([]documents.Document, 0, len(results))
	for _, result := range results {
		doc, err := se.repo.Get(ctx, result.DocID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get document %d: %w", result.DocID, err)
		}
		docs = append(docs, *doc)
	}

	return results, docs, nil
}
