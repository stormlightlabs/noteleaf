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

// Create stores a new note and returns its assigned ID
func (r *NoteRepository) Create(ctx context.Context, note *models.Note) (int64, error) {
	now := time.Now()
	note.Created = now
	note.Modified = now

	tags, err := note.MarshalTags()
	if err != nil {
		return 0, fmt.Errorf("failed to marshal tags: %w", err)
	}

	query := `
		INSERT INTO notes (title, content, tags, archived, created, modified, file_path)
		VALUES (?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.ExecContext(ctx, query,
		note.Title, note.Content, tags, note.Archived, note.Created, note.Modified, note.FilePath)
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
	query := `
		SELECT id, title, content, tags, archived, created, modified, file_path
		FROM notes WHERE id = ?`

	row := r.db.QueryRowContext(ctx, query, id)

	var note models.Note
	var tags string
	err := row.Scan(&note.ID, &note.Title, &note.Content, &tags, &note.Archived,
		&note.Created, &note.Modified, &note.FilePath)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("note with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to scan note: %w", err)
	}

	if err := note.UnmarshalTags(tags); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
	}

	return &note, nil
}

// Update modifies an existing note
func (r *NoteRepository) Update(ctx context.Context, note *models.Note) error {
	note.Modified = time.Now()

	tags, err := note.MarshalTags()
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	query := `
		UPDATE notes
		SET title = ?, content = ?, tags = ?, archived = ?, modified = ?, file_path = ?
		WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query,
		note.Title, note.Content, tags, note.Archived, note.Modified, note.FilePath, note.ID)
	if err != nil {
		return fmt.Errorf("failed to update note: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("note with id %d not found", note.ID)
	}

	return nil
}

// Delete removes a note by its ID
func (r *NoteRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM notes WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("note with id %d not found", id)
	}

	return nil
}

// List retrieves notes with optional filtering
func (r *NoteRepository) List(ctx context.Context, options NoteListOptions) ([]*models.Note, error) {
	query := "SELECT id, title, content, tags, archived, created, modified, file_path FROM notes"
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

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query notes: %w", err)
	}
	defer rows.Close()

	var notes []*models.Note
	for rows.Next() {
		var note models.Note
		var tags string
		err := rows.Scan(&note.ID, &note.Title, &note.Content, &tags, &note.Archived,
			&note.Created, &note.Modified, &note.FilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to scan note: %w", err)
		}

		if err := note.UnmarshalTags(tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}

		notes = append(notes, &note)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over notes: %w", err)
	}

	return notes, nil
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

// GetByTags retrieves notes that have any of the specified tags
func (r *NoteRepository) GetByTags(ctx context.Context, tags []string) ([]*models.Note, error) {
	if len(tags) == 0 {
		return []*models.Note{}, nil
	}

	placeholders := make([]string, len(tags))
	args := make([]any, len(tags))
	for i, tag := range tags {
		placeholders[i] = "?"
		args[i] = "%\"" + tag + "\"%"
	}

	query := fmt.Sprintf(`
		SELECT id, title, content, tags, archived, created, modified, file_path
		FROM notes
		WHERE %s
		ORDER BY modified DESC`,
		strings.Join(func() []string {
			conditions := make([]string, len(tags))
			for i := range tags {
				conditions[i] = "tags LIKE ?"
			}
			return conditions
		}(), " OR "))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query notes by tags: %w", err)
	}
	defer rows.Close()

	var notes []*models.Note
	for rows.Next() {
		var note models.Note
		var tagsJSON string
		err := rows.Scan(&note.ID, &note.Title, &note.Content, &tagsJSON, &note.Archived,
			&note.Created, &note.Modified, &note.FilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to scan note: %w", err)
		}

		if err := note.UnmarshalTags(tagsJSON); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}

		notes = append(notes, &note)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over notes: %w", err)
	}

	return notes, nil
}
