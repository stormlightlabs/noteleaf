// Package handlers provides command handlers for leaflet publication operations.
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
	atproto services.ATProtoClient
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

	if configDir, err := store.GetConfigDir(); err == nil {
		logFile := filepath.Join(configDir, "logs", fmt.Sprintf("publication_%s.log", time.Now().Format("2006-01-02")))
		ui.Infoln("Detailed logs: %s", logFile)
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

	ui.Infoln("Removing existing publications...")
	if err := h.repos.Notes.DeleteAllLeafletNotes(ctx); err != nil {
		return fmt.Errorf("failed to delete existing publications: %w", err)
	}

	var created, failed int

	for _, doc := range docs {
		content, err := documentToMarkdown(doc)
		if err != nil {
			ui.Warningln("Skipping document %s: %v", doc.Document.Title, err)
			failed++
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
			failed++
			continue
		}

		created++
		ui.Infoln("  Created: %s", doc.Document.Title)
	}

	if failed > 0 {
		ui.Successln("Sync complete: %d created, %d failed", created, failed)
	} else {
		ui.Successln("Sync complete: %d created", created)
	}
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

	note, doc, err := h.prepareDocumentForPublish(ctx, noteID, isDraft, false)
	if err != nil {
		return err
	}

	ui.Infoln("Creating document '%s' on leaflet...", note.Title)

	result, err := h.atproto.PostDocument(ctx, *doc, isDraft)
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

	tempNote, err := h.repos.Notes.Get(ctx, noteID)
	if err != nil {
		return fmt.Errorf("failed to get note: %w", err)
	}

	note, doc, err := h.prepareDocumentForPublish(ctx, noteID, tempNote.IsDraft, true)
	if err != nil {
		return err
	}

	if !note.IsDraft && note.PublishedAt == nil && doc.PublishedAt != "" {
		publishedAt, err := time.Parse(time.RFC3339, doc.PublishedAt)
		if err == nil {
			note.PublishedAt = &publishedAt
		}
	}

	ui.Infoln("Updating document '%s' on leaflet...", note.Title)

	result, err := h.atproto.PatchDocument(ctx, *note.LeafletRKey, *doc, note.IsDraft)
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

// Push creates or updates multiple documents on leaflet from local notes
func (h *PublicationHandler) Push(ctx context.Context, noteIDs []int64, isDraft bool) error {
	if !h.atproto.IsAuthenticated() {
		return fmt.Errorf("not authenticated - run 'noteleaf pub auth' first")
	}

	if len(noteIDs) == 0 {
		return fmt.Errorf("no note IDs provided")
	}

	ui.Infoln("Processing %d note(s)...\n", len(noteIDs))

	var created, updated, failed int
	var errors []string

	for _, noteID := range noteIDs {
		note, err := h.repos.Notes.Get(ctx, noteID)
		if err != nil {
			ui.Warningln("  [%d] Failed to get note: %v", noteID, err)
			errors = append(errors, fmt.Sprintf("note %d: %v", noteID, err))
			failed++
			continue
		}

		if note.HasLeafletAssociation() {
			err = h.Patch(ctx, noteID)
			if err != nil {
				ui.Warningln("  [%d] Failed to update '%s': %v", noteID, note.Title, err)
				errors = append(errors, fmt.Sprintf("note %d (%s): %v", noteID, note.Title, err))
				failed++
			} else {
				updated++
			}
		} else {
			err = h.Post(ctx, noteID, isDraft)
			if err != nil {
				ui.Warningln("  [%d] Failed to create '%s': %v", noteID, note.Title, err)
				errors = append(errors, fmt.Sprintf("note %d (%s): %v", noteID, note.Title, err))
				failed++
			} else {
				created++
			}
		}
	}

	ui.Newline()
	ui.Successln("Push complete: %d created, %d updated, %d failed", created, updated, failed)

	if len(errors) > 0 {
		return fmt.Errorf("push completed with %d error(s)", failed)
	}

	return nil
}

// Browse opens an interactive TUI for browsing publications
func (h *PublicationHandler) Browse(ctx context.Context, filter string) error {
	if filter == "" {
		filter = "all"
	}

	opts := ui.DataListOptions{
		Title: "Publications - " + filter,
	}

	list := ui.NewPublicationDataList(h.repos.Notes, opts, filter)
	return list.Browse(ctx)
}

// Read displays a publication's content with formatted markdown rendering.
// The identifier can be:
// - empty string: display the newest publication
// - numeric string: treat as database ID
// - non-numeric string: treat as AT Protocol rkey
func (h *PublicationHandler) Read(ctx context.Context, identifier string) error {
	var note *models.Note
	var err error

	if identifier == "" {
		note, err = h.repos.Notes.GetNewestPublication(ctx)
		if err != nil {
			return fmt.Errorf("failed to get newest publication: %w", err)
		}
	} else {
		var id int64
		_, scanErr := fmt.Sscanf(identifier, "%d", &id)
		if scanErr == nil {
			note, err = h.repos.Notes.Get(ctx, id)
			if err != nil {
				return fmt.Errorf("failed to get publication by ID: %w", err)
			}
			if !note.HasLeafletAssociation() {
				return fmt.Errorf("note %d is not a publication", id)
			}
		} else {
			note, err = h.repos.Notes.GetByLeafletRKey(ctx, identifier)
			if err != nil {
				return fmt.Errorf("failed to get publication by rkey: %w", err)
			}
		}
	}

	view := ui.NewPublicationView(note, ui.PublicationViewOptions{})
	return view.Show(ctx)
}

// prepareDocumentForPublish prepares a note for publication by converting to Leaflet format
func (h *PublicationHandler) prepareDocumentForPublish(ctx context.Context, noteID int64, isDraft bool, forPatch bool) (*models.Note, *public.Document, error) {
	note, err := h.repos.Notes.Get(ctx, noteID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get note: %w", err)
	}

	if !forPatch && note.HasLeafletAssociation() {
		return nil, nil, fmt.Errorf("note already published - use patch to update")
	}

	if forPatch && !note.HasLeafletAssociation() {
		return nil, nil, fmt.Errorf("note not published - use post to create")
	}

	session, err := h.atproto.GetSession()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get session: %w", err)
	}

	converter := public.NewMarkdownConverter()

	noteDir := extractNoteDirectory(note)
	if noteDir != "" {
		imageResolver := &public.LocalImageResolver{
			BlobUploader: func(data []byte, mimeType string) (public.Blob, error) {
				return h.atproto.UploadBlob(ctx, data, mimeType)
			},
		}
		converter = converter.WithImageResolver(imageResolver, noteDir)
	}

	blocks, err := converter.ToLeaflet(note.Content)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert markdown to leaflet format: %w", err)
	}

	doc := &public.Document{
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
		if forPatch && note.PublishedAt != nil {
			doc.PublishedAt = note.PublishedAt.Format(time.RFC3339)
		} else {
			now := time.Now()
			doc.PublishedAt = now.Format(time.RFC3339)
		}
	}

	return note, doc, nil
}

// writeDocumentOutput writes document to a file in JSON or plaintext format
func writeDocumentOutput(doc *public.Document, note *models.Note, outputPath string, plaintext bool) error {
	var content []byte
	var err error

	if plaintext {
		status := "published"
		if note != nil && note.IsDraft {
			status = "draft"
		}

		output := "Document Preview\n"
		output += "================\n\n"
		output += fmt.Sprintf("Title: %s\n", doc.Title)
		output += fmt.Sprintf("Status: %s\n", status)
		if note != nil {
			output += fmt.Sprintf("Note ID: %d\n", note.ID)
			if note.LeafletRKey != nil {
				output += fmt.Sprintf("RKey: %s\n", *note.LeafletRKey)
			}
		}
		output += fmt.Sprintf("Pages: %d\n", len(doc.Pages))
		if len(doc.Pages) > 0 {
			output += fmt.Sprintf("Blocks: %d\n", len(doc.Pages[0].Blocks))
		}
		if doc.PublishedAt != "" {
			output += fmt.Sprintf("PublishedAt: %s\n", doc.PublishedAt)
		}
		if doc.Author != "" {
			output += fmt.Sprintf("Author: %s\n", doc.Author)
		}

		content = []byte(output)
	} else {
		content, err = json.MarshalIndent(doc, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal document to JSON: %w", err)
		}
	}

	if err := os.WriteFile(outputPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write output to %s: %w", outputPath, err)
	}

	return nil
}

// PostPreview shows what would be posted without actually posting
func (h *PublicationHandler) PostPreview(ctx context.Context, noteID int64, isDraft bool, outputPath string, plaintext bool) error {
	if !h.atproto.IsAuthenticated() {
		return fmt.Errorf("not authenticated - run 'noteleaf pub auth' first")
	}

	note, doc, err := h.prepareDocumentForPublish(ctx, noteID, isDraft, false)
	if err != nil {
		return err
	}

	status := "published"
	if isDraft {
		status = "draft"
	}

	ui.Infoln("Preview: Would create document on leaflet")
	ui.Infoln("  Title: %s", doc.Title)
	ui.Infoln("  Status: %s", status)
	ui.Infoln("  Pages: %d", len(doc.Pages))
	ui.Infoln("  Blocks: %d", len(doc.Pages[0].Blocks))
	if doc.PublishedAt != "" {
		ui.Infoln("  PublishedAt: %s", doc.PublishedAt)
	}
	ui.Infoln("  Note ID: %d", note.ID)

	if outputPath != "" {
		if err := writeDocumentOutput(doc, note, outputPath, plaintext); err != nil {
			return err
		}
		format := "JSON"
		if plaintext {
			format = "plaintext"
		}
		ui.Successln("Output written to %s (%s format)", outputPath, format)
	}

	ui.Successln("Preview complete - no changes made")

	return nil
}

// PostValidate validates markdown conversion without posting
func (h *PublicationHandler) PostValidate(ctx context.Context, noteID int64, isDraft bool, outputPath string, plaintext bool) error {
	if !h.atproto.IsAuthenticated() {
		return fmt.Errorf("not authenticated - run 'noteleaf pub auth' first")
	}

	note, doc, err := h.prepareDocumentForPublish(ctx, noteID, isDraft, false)
	if err != nil {
		return err
	}

	ui.Infoln("Validating markdown conversion for note %d...", note.ID)
	ui.Successln("Validation successful!")
	ui.Infoln("  Title: %s", doc.Title)
	ui.Infoln("  Blocks converted: %d", len(doc.Pages[0].Blocks))

	if outputPath != "" {
		if err := writeDocumentOutput(doc, note, outputPath, plaintext); err != nil {
			return err
		}
		format := "JSON"
		if plaintext {
			format = "plaintext"
		}
		ui.Successln("Output written to %s (%s format)", outputPath, format)
	}

	return nil
}

// PatchPreview shows what would be patched without actually patching
func (h *PublicationHandler) PatchPreview(ctx context.Context, noteID int64, outputPath string, plaintext bool) error {
	if !h.atproto.IsAuthenticated() {
		return fmt.Errorf("not authenticated - run 'noteleaf pub auth' first")
	}

	tempNote, err := h.repos.Notes.Get(ctx, noteID)
	if err != nil {
		return fmt.Errorf("failed to get note: %w", err)
	}

	note, doc, err := h.prepareDocumentForPublish(ctx, noteID, tempNote.IsDraft, true)
	if err != nil {
		return err
	}

	status := "published"
	if note.IsDraft {
		status = "draft"
	}

	ui.Infoln("Preview: Would update document on leaflet")
	ui.Infoln("  Title: %s", doc.Title)
	ui.Infoln("  Status: %s", status)
	ui.Infoln("  RKey: %s", *note.LeafletRKey)
	ui.Infoln("  Pages: %d", len(doc.Pages))
	ui.Infoln("  Blocks: %d", len(doc.Pages[0].Blocks))
	if doc.PublishedAt != "" {
		ui.Infoln("  PublishedAt: %s", doc.PublishedAt)
	}

	if outputPath != "" {
		if err := writeDocumentOutput(doc, note, outputPath, plaintext); err != nil {
			return err
		}
		format := "JSON"
		if plaintext {
			format = "plaintext"
		}
		ui.Successln("Output written to %s (%s format)", outputPath, format)
	}

	ui.Successln("Preview complete - no changes made")

	return nil
}

// PatchValidate validates markdown conversion without patching
func (h *PublicationHandler) PatchValidate(ctx context.Context, noteID int64, outputPath string, plaintext bool) error {
	if !h.atproto.IsAuthenticated() {
		return fmt.Errorf("not authenticated - run 'noteleaf pub auth' first")
	}

	tempNote, err := h.repos.Notes.Get(ctx, noteID)
	if err != nil {
		return fmt.Errorf("failed to get note: %w", err)
	}

	note, doc, err := h.prepareDocumentForPublish(ctx, noteID, tempNote.IsDraft, true)
	if err != nil {
		return err
	}

	ui.Infoln("Validating markdown conversion for note %d...", note.ID)
	ui.Successln("Validation successful!")
	ui.Infoln("  Title: %s", doc.Title)
	ui.Infoln("  RKey: %s", *note.LeafletRKey)
	ui.Infoln("  Blocks converted: %d", len(doc.Pages[0].Blocks))

	if outputPath != "" {
		if err := writeDocumentOutput(doc, note, outputPath, plaintext); err != nil {
			return err
		}
		format := "JSON"
		if plaintext {
			format = "plaintext"
		}
		ui.Successln("Output written to %s (%s format)", outputPath, format)
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

// extractNoteDirectory extracts the directory path from a note's FilePath
func extractNoteDirectory(note *models.Note) string {
	if note.FilePath == "" {
		return ""
	}

	return filepath.Dir(note.FilePath)
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
