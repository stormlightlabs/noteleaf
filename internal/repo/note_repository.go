package repo

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/stormlightlabs/noteleaf/internal/models"
)

func NoteNotFoundError(id int64) error {
	return fmt.Errorf("note with id %d not found", id)
}

// NoteRepository provides database operations for notes
type NoteRepository struct {
	db *sql.DB
}

// NewNoteRepository creates a new note repository
func NewNoteRepository(db *sql.DB) *NoteRepository {
	return &NoteRepository{db: db}
}

// NoteListOptions defines filtering options for listing notes
type NoteListOptions struct {
	Tags     []string
	Archived *bool
	Title    string
	Content  string
	Limit    int
	Offset   int
}

func (r *NoteRepository) scanNote(s scanner) (*models.Note, error) {
	var note models.Note
	var tags string
	err := s.Scan(&note.ID, &note.Title, &note.Content, &tags, &note.Archived,
		&note.Created, &note.Modified, &note.FilePath, &note.LeafletRKey,
		&note.LeafletCID, &note.PublishedAt, &note.IsDraft)
	if err != nil {
		return nil, err
	}

	if err := note.UnmarshalTags(tags); err != nil {
		return nil, UnmarshalTagsError(err)
	}

	return &note, nil
}

func (r *NoteRepository) queryOne(ctx context.Context, query string, args ...any) (*models.Note, error) {
	row := r.db.QueryRowContext(ctx, query, args...)
	note, err := r.scanNote(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("note not found")
		}
		return nil, fmt.Errorf("failed to scan note: %w", err)
	}
	return note, nil
}

func (r *NoteRepository) queryMany(ctx context.Context, query string, args ...any) ([]*models.Note, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query notes: %w", err)
	}
	defer rows.Close()

	var notes []*models.Note
	for rows.Next() {
		note, err := r.scanNote(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan note: %w", err)
		}
		notes = append(notes, note)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over notes: %w", err)
	}

	return notes, nil
}

// Create stores a new note and returns its assigned ID
func (r *NoteRepository) Create(ctx context.Context, note *models.Note) (int64, error) {
	now := time.Now()
	note.Created = now
	note.Modified = now

	tags, err := note.MarshalTags()
	if err != nil {
		return 0, fmt.Errorf("failed to marshal tags: %w", err)
	}

	result, err := r.db.ExecContext(ctx, queryNoteInsert,
		note.Title, note.Content, tags, note.Archived, note.Created, note.Modified, note.FilePath,
		note.LeafletRKey, note.LeafletCID, note.PublishedAt, note.IsDraft)
	if err != nil {
		return 0, fmt.Errorf("failed to insert note: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	note.ID = id
	return id, nil
}

// Get retrieves a note by its ID
func (r *NoteRepository) Get(ctx context.Context, id int64) (*models.Note, error) {
	note, err := r.queryOne(ctx, queryNoteByID, id)
	if err != nil {
		return nil, NoteNotFoundError(id)
	}
	return note, nil
}

// Update modifies an existing note
func (r *NoteRepository) Update(ctx context.Context, note *models.Note) error {
	note.Modified = time.Now()

	tags, err := note.MarshalTags()
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	result, err := r.db.ExecContext(ctx, queryNoteUpdate,
		note.Title, note.Content, tags, note.Archived, note.Modified, note.FilePath,
		note.LeafletRKey, note.LeafletCID, note.PublishedAt, note.IsDraft, note.ID)
	if err != nil {
		return fmt.Errorf("failed to update note: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return NoteNotFoundError(note.ID)
	}

	return nil
}

// Delete removes a note by its ID
func (r *NoteRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, queryNoteDelete, id)
	if err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return NoteNotFoundError(id)
	}

	return nil
}

func (r *NoteRepository) buildListQuery(options NoteListOptions) (string, []any) {
	query := queryNotesList
	args := []any{}
	conditions := []string{}

	if options.Archived != nil {
		conditions = append(conditions, "archived = ?")
		args = append(args, *options.Archived)
	}

	if options.Title != "" {
		conditions = append(conditions, "title LIKE ?")
		args = append(args, "%"+options.Title+"%")
	}

	if options.Content != "" {
		conditions = append(conditions, "content LIKE ?")
		args = append(args, "%"+options.Content+"%")
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY modified DESC"

	if options.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", options.Limit)
		if options.Offset > 0 {
			query += fmt.Sprintf(" OFFSET %d", options.Offset)
		}
	}

	return query, args
}

// List retrieves notes with optional filtering
func (r *NoteRepository) List(ctx context.Context, options NoteListOptions) ([]*models.Note, error) {
	query, args := r.buildListQuery(options)
	return r.queryMany(ctx, query, args...)
}

// GetByTitle searches for notes by title pattern
func (r *NoteRepository) GetByTitle(ctx context.Context, title string) ([]*models.Note, error) {
	return r.List(ctx, NoteListOptions{Title: title})
}

// GetArchived retrieves all archived notes
func (r *NoteRepository) GetArchived(ctx context.Context) ([]*models.Note, error) {
	archived := true
	return r.List(ctx, NoteListOptions{Archived: &archived})
}

// GetActive retrieves all non-archived notes
func (r *NoteRepository) GetActive(ctx context.Context) ([]*models.Note, error) {
	archived := false
	return r.List(ctx, NoteListOptions{Archived: &archived})
}

// Archive marks a note as archived
func (r *NoteRepository) Archive(ctx context.Context, id int64) error {
	note, err := r.Get(ctx, id)
	if err != nil {
		return err
	}

	note.Archived = true
	return r.Update(ctx, note)
}

// Unarchive marks a note as not archived
func (r *NoteRepository) Unarchive(ctx context.Context, id int64) error {
	note, err := r.Get(ctx, id)
	if err != nil {
		return err
	}

	note.Archived = false
	return r.Update(ctx, note)
}

// SearchContent searches for notes containing the specified text in content
func (r *NoteRepository) SearchContent(ctx context.Context, searchText string) ([]*models.Note, error) {
	return r.List(ctx, NoteListOptions{Content: searchText})
}

// GetRecent retrieves the most recently modified notes
func (r *NoteRepository) GetRecent(ctx context.Context, limit int) ([]*models.Note, error) {
	return r.List(ctx, NoteListOptions{Limit: limit})
}

// AddTag adds a tag to a note
func (r *NoteRepository) AddTag(ctx context.Context, id int64, tag string) error {
	note, err := r.Get(ctx, id)
	if err != nil {
		return err
	}

	if slices.Contains(note.Tags, tag) {
		return nil
	}

	note.Tags = append(note.Tags, tag)
	return r.Update(ctx, note)
}

// RemoveTag removes a tag from a note
func (r *NoteRepository) RemoveTag(ctx context.Context, id int64, tag string) error {
	note, err := r.Get(ctx, id)
	if err != nil {
		return err
	}
	for i, existingTag := range note.Tags {
		if existingTag == tag {
			note.Tags = append(note.Tags[:i], note.Tags[i+1:]...)
			break
		}
	}
	return r.Update(ctx, note)
}

func (r *NoteRepository) buildTagsQuery(tags []string) (string, []any) {
	conditions := make([]string, len(tags))
	args := make([]any, len(tags))
	for i, tag := range tags {
		conditions[i] = "tags LIKE ?"
		args[i] = "%\"" + tag + "\"%"
	}
	return fmt.Sprintf(`SELECT %s FROM notes WHERE %s ORDER BY modified DESC`, noteColumns, strings.Join(conditions, " OR ")), args
}

// GetByTags retrieves notes that have any of the specified tags
func (r *NoteRepository) GetByTags(ctx context.Context, tags []string) ([]*models.Note, error) {
	if len(tags) == 0 {
		return []*models.Note{}, nil
	}
	query, args := r.buildTagsQuery(tags)
	return r.queryMany(ctx, query, args...)
}

// GetByLeafletRKey returns a note by its leaflet record key
func (r *NoteRepository) GetByLeafletRKey(ctx context.Context, rkey string) (*models.Note, error) {
	query := "SELECT " + noteColumns + " FROM notes WHERE leaflet_rkey = ?"
	return r.queryOne(ctx, query, rkey)
}

// ListPublished returns all published leaflet notes (not drafts)
func (r *NoteRepository) ListPublished(ctx context.Context) ([]*models.Note, error) {
	query := "SELECT " + noteColumns + " FROM notes WHERE leaflet_rkey IS NOT NULL AND is_draft = false ORDER BY published_at DESC"
	return r.queryMany(ctx, query)
}

// ListDrafts returns all draft leaflet notes
func (r *NoteRepository) ListDrafts(ctx context.Context) ([]*models.Note, error) {
	query := "SELECT " + noteColumns + " FROM notes WHERE leaflet_rkey IS NOT NULL AND is_draft = true ORDER BY modified DESC"
	return r.queryMany(ctx, query)
}

// GetLeafletNotes returns all notes with leaflet association (both published and drafts)
func (r *NoteRepository) GetLeafletNotes(ctx context.Context) ([]*models.Note, error) {
	query := "SELECT " + noteColumns + " FROM notes WHERE leaflet_rkey IS NOT NULL ORDER BY modified DESC"
	return r.queryMany(ctx, query)
}

// GetNewestPublication returns the most recently published leaflet note
func (r *NoteRepository) GetNewestPublication(ctx context.Context) (*models.Note, error) {
	query := "SELECT " + noteColumns + " FROM notes WHERE leaflet_rkey IS NOT NULL ORDER BY published_at DESC LIMIT 1"
	return r.queryOne(ctx, query)
}

// DeleteAllLeafletNotes removes all notes with leaflet associations
func (r *NoteRepository) DeleteAllLeafletNotes(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM notes WHERE leaflet_rkey IS NOT NULL")
	if err != nil {
		return fmt.Errorf("failed to delete leaflet notes: %w", err)
	}
	return nil
}
