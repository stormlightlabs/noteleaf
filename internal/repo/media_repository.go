package repo

import (
	"context"

	"github.com/stormlightlabs/noteleaf/internal/models"
)

// MediaRepository defines CRUD operations for media types (Books, Movies, TV)
//
// This interface captures the shared behavior across media repositories
type MediaRepository[T models.Model] interface {
	// Create stores a new media item and returns its assigned ID
	Create(ctx context.Context, item *T) (int64, error)

	// Get retrieves a media item by ID
	Get(ctx context.Context, id int64) (*T, error)

	// Update modifies an existing media item
	Update(ctx context.Context, item *T) error

	// Delete removes a media item by ID
	Delete(ctx context.Context, id int64) error

	// List retrieves media items with optional filtering and sorting
	List(ctx context.Context, opts any) ([]*T, error)

	// Count returns the number of media items matching conditions
	Count(ctx context.Context, opts any) (int64, error)
}

// StatusFilterable extends MediaRepository with status-based filtering
//
// Media types (Books, Movies, TV) support status-based queries like "queued", "reading", "watching", "watched", "finished"
type StatusFilterable[T models.Model] interface {
	MediaRepository[T]

	// GetByStatus retrieves all items with the given status
	GetByStatus(ctx context.Context, status string) ([]*T, error)
}
