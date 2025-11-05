package ui

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/stormlightlabs/noteleaf/internal/models"
	"github.com/stormlightlabs/noteleaf/internal/utils"
)

// PublicationRecord adapts models.Note with leaflet metadata to work with DataList
type PublicationRecord struct {
	*models.Note
}

func (p *PublicationRecord) GetField(name string) any {
	switch name {
	case "id":
		return p.ID
	case "title":
		return p.Title
	case "status":
		if p.IsDraft {
			return "draft"
		}
		return "published"
	case "published_at":
		return p.PublishedAt
	case "modified":
		return p.Modified
	case "leaflet_rkey":
		return p.LeafletRKey
	case "leaflet_cid":
		return p.LeafletCID
	default:
		return ""
	}
}

func (p *PublicationRecord) GetTitle() string {
	status := "draft"
	if !p.IsDraft {
		status = "published"
	}
	return fmt.Sprintf("[%d] %s (%s)", p.ID, p.Title, status)
}

func (p *PublicationRecord) GetDescription() string {
	var parts []string

	if p.PublishedAt != nil {
		parts = append(parts, "Published: "+p.PublishedAt.Format("2006-01-02 15:04"))
	}

	parts = append(parts, "Modified: "+p.Modified.Format("2006-01-02 15:04"))

	if p.LeafletRKey != nil {
		parts = append(parts, "rkey: "+*p.LeafletRKey)
	}

	return strings.Join(parts, " • ")
}

func (p *PublicationRecord) GetFilterValue() string {
	searchable := []string{p.Title, p.Content}
	if p.LeafletRKey != nil {
		searchable = append(searchable, *p.LeafletRKey)
	}
	return strings.Join(searchable, " ")
}

// PublicationDataSource loads notes with leaflet metadata
type PublicationDataSource struct {
	repo   utils.TestNoteRepository
	filter string // "all", "published", or "draft"
}

func (p *PublicationDataSource) Load(ctx context.Context, opts ListOptions) ([]ListItem, error) {
	var notes []*models.Note
	var err error

	switch p.filter {
	case "published":
		notes, err = p.repo.ListPublished(ctx)
	case "draft":
		notes, err = p.repo.ListDrafts(ctx)
	default:
		notes, err = p.repo.GetLeafletNotes(ctx)
	}

	if err != nil {
		return nil, err
	}

	if opts.Search != "" {
		var filtered []*models.Note
		searchLower := strings.ToLower(opts.Search)
		for _, note := range notes {
			if strings.Contains(strings.ToLower(note.Title), searchLower) ||
				strings.Contains(strings.ToLower(note.Content), searchLower) ||
				(note.LeafletRKey != nil && strings.Contains(strings.ToLower(*note.LeafletRKey), searchLower)) {
				filtered = append(filtered, note)
			}
		}
		notes = filtered
	}

	if opts.Limit > 0 && opts.Limit < len(notes) {
		notes = notes[:opts.Limit]
	}

	items := make([]ListItem, len(notes))
	for i, note := range notes {
		items[i] = &PublicationRecord{Note: note}
	}

	return items, nil
}

func (p *PublicationDataSource) Count(ctx context.Context, opts ListOptions) (int, error) {
	items, err := p.Load(ctx, opts)
	if err != nil {
		return 0, err
	}
	return len(items), nil
}

func (p *PublicationDataSource) Search(ctx context.Context, query string, opts ListOptions) ([]ListItem, error) {
	opts.Search = query
	return p.Load(ctx, opts)
}

// NewPublicationDataList creates a new DataList for browsing published/draft documents
func NewPublicationDataList(repo utils.TestNoteRepository, opts DataListOptions, filter string) *DataList {
	if opts.Title == "" {
		opts.Title = "Publications"
	}

	opts.ShowSearch = true
	opts.Searchable = true

	if opts.ViewHandler == nil {
		opts.ViewHandler = func(item ListItem) string {
			if pubRecord, ok := item.(*PublicationRecord); ok {
				return formatPublicationForView(pubRecord.Note)
			}
			return "Unable to display publication"
		}
	}

	source := &PublicationDataSource{
		repo:   repo,
		filter: filter,
	}

	return NewDataList(source, opts)
}

// NewPublicationListFromList creates a publication list using DataList
func NewPublicationListFromList(repo utils.TestNoteRepository, output io.Writer, input io.Reader, static bool, filter string) *DataList {
	opts := DataListOptions{
		Output: output,
		Input:  input,
		Static: static,
		Title:  "Publications",
	}
	return NewPublicationDataList(repo, opts, filter)
}

// formatPublicationForView formats a publication for display with glamour
func formatPublicationForView(note *models.Note) string {
	var content strings.Builder

	content.WriteString("# " + note.Title + "\n\n")

	status := "published"
	if note.IsDraft {
		status = "draft"
	}
	content.WriteString("**Status:** " + status + "\n")

	if note.PublishedAt != nil {
		content.WriteString("**Published:** " + note.PublishedAt.Format("2006-01-02 15:04") + "\n")
	}

	content.WriteString("**Modified:** " + note.Modified.Format("2006-01-02 15:04") + "\n")

	if note.LeafletRKey != nil {
		content.WriteString("**RKey:** `" + *note.LeafletRKey + "`\n")
	}

	if note.LeafletCID != nil {
		content.WriteString("**CID:** `" + *note.LeafletCID + "`\n")
	}

	content.WriteString("\n---\n\n")

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
