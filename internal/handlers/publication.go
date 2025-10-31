// TODO: Store credentials securely in [PublicationHandler.Auth]
// Options:
//  1. Use system keyring (go-keyring)
//  2. Store encrypted in config file
//  3. Store in environment variables
//
// TODO: Implement document processing
// For each document:
//  1. Check if note with this leaflet_rkey exists
//  2. If exists: Update note content, title, metadata
//  3. If new: Create new note with leaflet metadata
//  4. Convert document blocks to markdown
//  5. Save to database
//
// TODO: Implement list functionality
//  1. Query notes where leaflet_rkey IS NOT NULL
//  2. Apply filter (published vs draft) - "all", "published", "draft", or empty (default: all)
//  3. Use prior art from package ui and other handlers to render
//
// TODO: Implmenent pull command
//  1. Authenticates with AT Protocol
//  2. Fetches all pub.leaflet.document records
//  3. Creates new notes for documents not seen before
//  4. Updates existing notes (matched by leaflet_rkey)
//  5. Shows summary of pulled documents
package handlers

import (
	"context"
	"fmt"

	"github.com/stormlightlabs/noteleaf/internal/repo"
	"github.com/stormlightlabs/noteleaf/internal/services"
	"github.com/stormlightlabs/noteleaf/internal/store"
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

	fmt.Printf("Authenticating as %s...\n", handle)

	if err := h.atproto.Authenticate(ctx, handle, password); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	fmt.Println("✓ Authentication successful!")
	fmt.Println("TODO: Implement persistent credential storage")
	return nil
}

// Pull fetches all documents from leaflet and creates/updates local notes
func (h *PublicationHandler) Pull(ctx context.Context) error {
	fmt.Println("TODO: Implement document conversion and note creation")
	return nil
}

// List displays notes with leaflet publication metadata, showing all notes that
// have been pulled from or pushed to leaflet
func (h *PublicationHandler) List(ctx context.Context, filter string) error {
	fmt.Println("TODO: Implement leaflet document listing")
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
