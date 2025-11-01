// Package handlers provides command handlers for leaflet publication operations.
//
// Pull command:
//  1. Authenticates with AT Protocol
//  2. Fetches all pub.leaflet.document records
//  3. Creates new notes for documents not seen before
//  4. Updates existing notes (matched by leaflet_rkey)
//  5. Shows summary of pulled documents
//
// List command:
//  1. Query notes where leaflet_rkey IS NOT NULL
//  2. Apply filter (published vs draft) - "all", "published", "draft", or empty (default: all)
//  3. Static output (TUI viewing marked as TODO)
//
// TODO: Add TUI viewing for document details
package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/stormlightlabs/noteleaf/internal/models"
	"github.com/stormlightlabs/noteleaf/internal/public"
	"github.com/stormlightlabs/noteleaf/internal/repo"
	"github.com/stormlightlabs/noteleaf/internal/services"
	"github.com/stormlightlabs/noteleaf/internal/store"
	"github.com/stormlightlabs/noteleaf/internal/ui"
)

// PublicationHandler handles leaflet publication commands
type PublicationHandler struct {
	db      *store.Database
	config  *store.Config
	repos   *repo.Repositories
	atproto *services.ATProtoService
}

// NewPublicationHandler creates a new publication handler
func NewPublicationHandler() (*PublicationHandler, error) {
	db, err := store.NewDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	config, err := store.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	repos := repo.NewRepositories(db.DB)
	atproto := services.NewATProtoService()

	if config.ATProtoDID != "" && config.ATProtoAccessJWT != "" && config.ATProtoRefreshJWT != "" {
		session, err := sessionFromConfig(config)
		if err == nil {
			_ = atproto.RestoreSession(session)
		}
	}

	return &PublicationHandler{
		db:      db,
		config:  config,
		repos:   repos,
		atproto: atproto,
	}, nil
}

// Close cleans up resources
func (h *PublicationHandler) Close() error {
	if h.atproto != nil {
		if err := h.atproto.Close(); err != nil {
			return err
		}
	}
	if h.db != nil {
		return h.db.Close()
	}
	return nil
}

// Auth handles authentication with BlueSky/leaflet
func (h *PublicationHandler) Auth(ctx context.Context, handle, password string) error {
	if handle == "" {
		return fmt.Errorf("handle is required")
	}

	if password == "" {
		return fmt.Errorf("password is required")
	}

	ui.Infoln("Authenticating as %s...", handle)

	if err := h.atproto.Authenticate(ctx, handle, password); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	session, err := h.atproto.GetSession()
	if err != nil {
		return fmt.Errorf("failed to get session after authentication: %w", err)
	}

	h.config.ATProtoDID = session.DID
	h.config.ATProtoHandle = session.Handle
	h.config.ATProtoAccessJWT = session.AccessJWT
	h.config.ATProtoRefreshJWT = session.RefreshJWT
	h.config.ATProtoPDSURL = session.PDSURL
	h.config.ATProtoExpiresAt = session.ExpiresAt.Format("2006-01-02T15:04:05Z07:00")

	if err := store.SaveConfig(h.config); err != nil {
		return fmt.Errorf("authentication successful but failed to save credentials: %w", err)
	}

	ui.Successln("Authentication successful!")
	ui.Successln("Credentials saved")
	return nil
}

// documentToMarkdown converts a leaflet Document to markdown content
func documentToMarkdown(doc services.DocumentWithMeta) (string, error) {
	converter := public.NewMarkdownConverter()
	var allBlocks []public.BlockWrap

	for _, page := range doc.Document.Pages {
		allBlocks = append(allBlocks, page.Blocks...)
	}

	content, err := converter.FromLeaflet(allBlocks)
	if err != nil {
		return "", fmt.Errorf("failed to convert document to markdown: %w", err)
	}

	return content, nil
}

// Pull fetches all documents from leaflet and creates/updates local notes
func (h *PublicationHandler) Pull(ctx context.Context) error {
	if !h.atproto.IsAuthenticated() {
		return fmt.Errorf("not authenticated - run 'noteleaf pub auth' first")
	}

	ui.Infoln("Fetching documents from leaflet...")

	docs, err := h.atproto.PullDocuments(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch documents: %w", err)
	}

	if len(docs) == 0 {
		ui.Infoln("No documents found in leaflet.")
		return nil
	}

	ui.Infoln("Found %d document(s). Syncing...\n", len(docs))

	var created, updated int

	for _, doc := range docs {
		existing, err := h.repos.Notes.GetByLeafletRKey(ctx, doc.Meta.RKey)
		if err == nil && existing != nil {
			content, err := documentToMarkdown(doc)
			if err != nil {
				ui.Warningln("⚠ Skipping document %s: %v", doc.Document.Title, err)
				continue
			}

			existing.Title = doc.Document.Title
			existing.Content = content
			existing.LeafletCID = &doc.Meta.CID
			existing.IsDraft = doc.Meta.IsDraft

			if doc.Document.PublishedAt != "" {
				publishedAt, err := time.Parse(time.RFC3339, doc.Document.PublishedAt)
				if err == nil {
					existing.PublishedAt = &publishedAt
				}
			}

			if err := h.repos.Notes.Update(ctx, existing); err != nil {
				ui.Warningln("⚠ Failed to update note for document %s: %v", doc.Document.Title, err)
				continue
			}

			updated++
			ui.Infoln("  Updated: %s", doc.Document.Title)
		} else {
			content, err := documentToMarkdown(doc)
			if err != nil {
				ui.Warningln("⚠ Skipping document %s: %v", doc.Document.Title, err)
				continue
			}

			note := &models.Note{
				Title:       doc.Document.Title,
				Content:     content,
				LeafletRKey: &doc.Meta.RKey,
				LeafletCID:  &doc.Meta.CID,
				IsDraft:     doc.Meta.IsDraft,
			}

			if doc.Document.PublishedAt != "" {
				publishedAt, err := time.Parse(time.RFC3339, doc.Document.PublishedAt)
				if err == nil {
					note.PublishedAt = &publishedAt
				}
			}

			_, err = h.repos.Notes.Create(ctx, note)
			if err != nil {
				ui.Warningln("⚠ Failed to create note for document %s: %v", doc.Document.Title, err)
				continue
			}

			created++
			ui.Infoln("  Created: %s", doc.Document.Title)
		}
	}

	ui.Successln("Sync complete: %d created, %d updated", created, updated)
	return nil
}

// printPublication prints a single publication note in static format
func printPublication(note *models.Note) {
	status := "published"
	if note.IsDraft {
		status = "draft"
	}

	ui.Infoln("[%d] %s (%s)", note.ID, note.Title, status)

	if note.LeafletRKey != nil {
		ui.Infoln("    rkey: %s", *note.LeafletRKey)
	}

	if note.PublishedAt != nil {
		ui.Infoln("    published: %s", note.PublishedAt.Format("2006-01-02 15:04:05"))
	}

	ui.Infoln("    modified: %s", note.Modified.Format("2006-01-02 15:04:05"))
	ui.Newline()
}

// List displays notes with leaflet publication metadata, showing all notes that have been pulled from or pushed to leaflet
func (h *PublicationHandler) List(ctx context.Context, filter string) error {
	if filter == "" {
		filter = "all"
	}

	var notes []*models.Note
	var err error

	switch filter {
	case "all":
		notes, err = h.repos.Notes.GetLeafletNotes(ctx)
		if err != nil {
			return fmt.Errorf("failed to fetch leaflet notes: %w", err)
		}
	case "published":
		notes, err = h.repos.Notes.ListPublished(ctx)
		if err != nil {
			return fmt.Errorf("failed to fetch published notes: %w", err)
		}
	case "draft":
		notes, err = h.repos.Notes.ListDrafts(ctx)
		if err != nil {
			return fmt.Errorf("failed to fetch draft notes: %w", err)
		}
	default:
		return fmt.Errorf("invalid filter: %s (must be 'all', 'published', or 'draft')", filter)
	}

	if len(notes) == 0 {
		ui.Infoln("No %s documents found.", filter)
		return nil
	}

	ui.Infoln("Found %d %s document(s):", len(notes), filter)
	ui.Newline()

	for _, note := range notes {
		printPublication(note)
	}

	return nil
}

// GetAuthStatus returns the current authentication status
func (h *PublicationHandler) GetAuthStatus() string {
	if h.atproto.IsAuthenticated() {
		session, _ := h.atproto.GetSession()
		if session != nil {
			return fmt.Sprintf("Authenticated as %s", session.Handle)
		}
		return "Authenticated (session details unavailable)"
	}
	return "Not authenticated"
}

// sessionFromConfig converts config AT Protocol fields to a Session
func sessionFromConfig(config *store.Config) (*services.Session, error) {
	if config.ATProtoDID == "" || config.ATProtoAccessJWT == "" || config.ATProtoRefreshJWT == "" {
		return nil, fmt.Errorf("incomplete session data in config")
	}

	var expiresAt time.Time
	if config.ATProtoExpiresAt != "" {
		parsed, err := time.Parse("2006-01-02T15:04:05Z07:00", config.ATProtoExpiresAt)
		if err != nil {
			expiresAt = time.Now().Add(-1 * time.Hour)
		} else {
			expiresAt = parsed
		}
	} else {
		expiresAt = time.Now().Add(-1 * time.Hour)
	}

	return &services.Session{
		DID:           config.ATProtoDID,
		Handle:        config.ATProtoHandle,
		AccessJWT:     config.ATProtoAccessJWT,
		RefreshJWT:    config.ATProtoRefreshJWT,
		PDSURL:        config.ATProtoPDSURL,
		ExpiresAt:     expiresAt,
		Authenticated: true,
	}, nil
}
