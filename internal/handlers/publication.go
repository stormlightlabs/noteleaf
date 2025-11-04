// Package handlers provides command handlers for leaflet publication operations.
//
// TODO: Post (create 1)
// TODO: Patch (update 1)
// TODO: Push (create or update - more than 1)
// TODO: Add TUI viewing for document details
// TODO: Repost - "Reblog" - post to BlueSky
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
				ui.Warningln("Skipping document %s: %v", doc.Document.Title, err)
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
				ui.Warningln("Failed to update note for document %s: %v", doc.Document.Title, err)
				continue
			}

			updated++
			ui.Infoln("  Updated: %s", doc.Document.Title)
		} else {
			content, err := documentToMarkdown(doc)
			if err != nil {
				ui.Warningln("Skipping document %s: %v", doc.Document.Title, err)
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
				ui.Warningln("Failed to create note for document %s: %v", doc.Document.Title, err)
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

// Post creates a new document on leaflet from a local note
func (h *PublicationHandler) Post(ctx context.Context, noteID int64, isDraft bool) error {
	if !h.atproto.IsAuthenticated() {
		return fmt.Errorf("not authenticated - run 'noteleaf pub auth' first")
	}

	note, err := h.repos.Notes.Get(ctx, noteID)
	if err != nil {
		return fmt.Errorf("failed to get note: %w", err)
	}

	if note.HasLeafletAssociation() {
		return fmt.Errorf("note already published - use patch to update")
	}

	session, err := h.atproto.GetSession()
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// TODO: Implement image handling for markdown conversion
	// 1. Extract note's directory from filepath/database
	// 2. Create LocalImageResolver with BlobUploader that calls h.atproto.UploadBlob()
	// 3. Use converter.WithImageResolver(resolver, noteDir) before ToLeaflet()
	// This will upload images to AT Protocol and get real CIDs/dimensions
	converter := public.NewMarkdownConverter()
	blocks, err := converter.ToLeaflet(note.Content)
	if err != nil {
		return fmt.Errorf("failed to convert markdown to leaflet format: %w", err)
	}

	doc := public.Document{
		Author:      session.DID,
		Title:       note.Title,
		Description: "",
		Pages: []public.LinearDocument{
			{
				Type:   public.TypeLinearDocument,
				Blocks: blocks,
			},
		},
	}

	if !isDraft {
		now := time.Now()
		doc.PublishedAt = now.Format(time.RFC3339)
	}

	ui.Infoln("Creating document '%s' on leaflet...", note.Title)

	result, err := h.atproto.PostDocument(ctx, doc, isDraft)
	if err != nil {
		return fmt.Errorf("failed to post document: %w", err)
	}

	note.LeafletRKey = &result.Meta.RKey
	note.LeafletCID = &result.Meta.CID
	note.IsDraft = isDraft

	if !isDraft && doc.PublishedAt != "" {
		publishedAt, err := time.Parse(time.RFC3339, doc.PublishedAt)
		if err == nil {
			note.PublishedAt = &publishedAt
		}
	}

	if err := h.repos.Notes.Update(ctx, note); err != nil {
		return fmt.Errorf("document created but failed to update local note: %w", err)
	}

	if isDraft {
		ui.Successln("Draft created successfully!")
	} else {
		ui.Successln("Document published successfully!")
	}
	ui.Infoln("  RKey: %s", result.Meta.RKey)
	ui.Infoln("  CID: %s", result.Meta.CID)

	return nil
}

// Patch updates an existing document on leaflet from a local note
func (h *PublicationHandler) Patch(ctx context.Context, noteID int64) error {
	if !h.atproto.IsAuthenticated() {
		return fmt.Errorf("not authenticated - run 'noteleaf pub auth' first")
	}

	note, err := h.repos.Notes.Get(ctx, noteID)
	if err != nil {
		return fmt.Errorf("failed to get note: %w", err)
	}

	if !note.HasLeafletAssociation() {
		return fmt.Errorf("note not published - use post to create")
	}

	session, err := h.atproto.GetSession()
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// TODO: Implement image handling for markdown conversion (same as Post method)
	// 1. Extract note's directory from filepath/database
	// 2. Create LocalImageResolver with BlobUploader that calls h.atproto.UploadBlob()
	// 3. Use converter.WithImageResolver(resolver, noteDir) before ToLeaflet()
	// This will upload images to AT Protocol and get real CIDs/dimensions
	converter := public.NewMarkdownConverter()
	blocks, err := converter.ToLeaflet(note.Content)
	if err != nil {
		return fmt.Errorf("failed to convert markdown to leaflet format: %w", err)
	}

	doc := public.Document{
		Author:      session.DID,
		Title:       note.Title,
		Description: "",
		Pages: []public.LinearDocument{
			{
				Type:   public.TypeLinearDocument,
				Blocks: blocks,
			},
		},
	}

	if !note.IsDraft && note.PublishedAt != nil {
		doc.PublishedAt = note.PublishedAt.Format(time.RFC3339)
	} else if !note.IsDraft {
		now := time.Now()
		doc.PublishedAt = now.Format(time.RFC3339)
		note.PublishedAt = &now
	}

	ui.Infoln("Updating document '%s' on leaflet...", note.Title)

	result, err := h.atproto.PatchDocument(ctx, *note.LeafletRKey, doc, note.IsDraft)
	if err != nil {
		return fmt.Errorf("failed to patch document: %w", err)
	}

	note.LeafletCID = &result.Meta.CID

	if err := h.repos.Notes.Update(ctx, note); err != nil {
		return fmt.Errorf("document updated but failed to update local note: %w", err)
	}

	ui.Successln("Document updated successfully!")
	ui.Infoln("  RKey: %s", result.Meta.RKey)
	ui.Infoln("  CID: %s", result.Meta.CID)

	return nil
}

// Delete removes a document from leaflet
func (h *PublicationHandler) Delete(ctx context.Context, noteID int64) error {
	if !h.atproto.IsAuthenticated() {
		return fmt.Errorf("not authenticated - run 'noteleaf pub auth' first")
	}

	note, err := h.repos.Notes.Get(ctx, noteID)
	if err != nil {
		return fmt.Errorf("failed to get note: %w", err)
	}

	if !note.HasLeafletAssociation() {
		return fmt.Errorf("note not published on leaflet")
	}

	ui.Infoln("Deleting document '%s' from leaflet...", note.Title)

	err = h.atproto.DeleteDocument(ctx, *note.LeafletRKey, note.IsDraft)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	note.LeafletRKey = nil
	note.LeafletCID = nil
	note.PublishedAt = nil
	note.IsDraft = false

	if err := h.repos.Notes.Update(ctx, note); err != nil {
		return fmt.Errorf("document deleted but failed to update local note: %w", err)
	}

	ui.Successln("Document deleted successfully!")

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
