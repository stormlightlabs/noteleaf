package ui

import (
	"context"
	"io"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/stormlightlabs/noteleaf/internal/models"
	"github.com/stormlightlabs/noteleaf/internal/repo"
	"github.com/stormlightlabs/noteleaf/internal/utils"
)

// NoteRecord adapts models.Note to work with DataList (since notes work better as a list than table)
type NoteRecord struct {
	*models.Note
}

func (n *NoteRecord) GetField(name string) any {
	switch name {
	case "id":
		return n.ID
	case "title":
		return n.Title
	case "content":
		return n.Content
	case "tags":
		return n.Tags
	case "archived":
		return n.Archived
	case "created":
		return n.Created
	case "modified":
		return n.Modified
	case "file_path":
		return n.FilePath
	default:
		return ""
	}
}

func (n *NoteRecord) GetTitle() string {
	return n.Title
}

func (n *NoteRecord) GetDescription() string {
	var parts []string

	if len(n.Tags) > 0 {
		parts = append(parts, strings.Join(n.Tags, ", "))
	}

	parts = append(parts, "Modified: "+n.Modified.Format("2006-01-02 15:04"))

	return strings.Join(parts, " • ")
}

func (n *NoteRecord) GetFilterValue() string {
	// Make notes searchable by title, content, and tags
	searchable := []string{n.Title, n.Content}
	searchable = append(searchable, n.Tags...)
	return strings.Join(searchable, " ")
}

// NoteDataSource adapts NoteRepository to work with DataList
type NoteDataSource struct {
	repo         utils.TestNoteRepository
	showArchived bool
	tags         []string
}

func (n *NoteDataSource) Load(ctx context.Context, opts ListOptions) ([]ListItem, error) {
	repoOpts := repo.NoteListOptions{
		Tags: n.tags,
	}

	if !n.showArchived {
		archived := false
		repoOpts.Archived = &archived
	}

	if opts.Search != "" {
		repoOpts.Content = opts.Search
	}

	if opts.Limit > 0 {
		repoOpts.Limit = opts.Limit
	}

	notes, err := n.repo.List(ctx, repoOpts)
	if err != nil {
		return nil, err
	}

	items := make([]ListItem, len(notes))
	for i, note := range notes {
		items[i] = &NoteRecord{Note: note}
	}

	return items, nil
}

func (n *NoteDataSource) Count(ctx context.Context, opts ListOptions) (int, error) {
	items, err := n.Load(ctx, opts)
	if err != nil {
		return 0, err
	}
	return len(items), nil
}

func (n *NoteDataSource) Search(ctx context.Context, query string, opts ListOptions) ([]ListItem, error) {
	opts.Search = query
	return n.Load(ctx, opts)
}

// NewNoteDataList creates a new DataList for browsing notes
func NewNoteDataList(repo utils.TestNoteRepository, opts DataListOptions, showArchived bool, tags []string) *DataList {
	if opts.Title == "" {
		opts.Title = "Notes"
	}

	opts.ShowSearch = true
	opts.Searchable = true

	if opts.ViewHandler == nil {
		opts.ViewHandler = func(item ListItem) string {
			if noteRecord, ok := item.(*NoteRecord); ok {
				return formatNoteForView(noteRecord.Note)
			}
			return "Unable to display note"
		}
	}

	source := &NoteDataSource{
		repo:         repo,
		showArchived: showArchived,
		tags:         tags,
	}

	return NewDataList(source, opts)
}

// NewNoteListFromList creates a NoteList-compatible interface using DataList
func NewNoteListFromList(repo utils.TestNoteRepository, output io.Writer, input io.Reader, static bool, showArchived bool, tags []string) *DataList {
	opts := DataListOptions{
		Output: output,
		Input:  input,
		Static: static,
		Title:  "Notes",
	}
	return NewNoteDataList(repo, opts, showArchived, tags)
}

// formatNoteForView formats a note for display (similar to original implementation)
func formatNoteForView(note *models.Note) string {
	var content strings.Builder

	content.WriteString("# " + note.Title + "\n\n")

	if len(note.Tags) > 0 {
		content.WriteString("**Tags:** ")
		for i, tag := range note.Tags {
			if i > 0 {
				content.WriteString(", ")
			}
			content.WriteString("`" + tag + "`")
		}
		content.WriteString("\n\n")
	}

	content.WriteString("**Created:** " + note.Created.Format("2006-01-02 15:04") + "\n")
	content.WriteString("**Modified:** " + note.Modified.Format("2006-01-02 15:04") + "\n\n")
	content.WriteString("---\n\n")

	noteContent := strings.TrimSpace(note.Content)
	if !strings.HasPrefix(noteContent, "# ") {
		content.WriteString(noteContent)
	} else {
		lines := strings.Split(noteContent, "\n")
		if len(lines) > 1 {
			content.WriteString(strings.Join(lines[1:], "\n"))
		}
	}

	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		return content.String()
	}

	rendered, err := renderer.Render(content.String())
	if err != nil {
		return content.String()
	}

	return rendered
}
