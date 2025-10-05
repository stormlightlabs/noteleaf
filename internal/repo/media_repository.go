package repo

import (
	"context"

	"github.com/stormlightlabs/noteleaf/internal/models"
)

// MediaRepository defines CRUD operations for media types (Books, Movies, TV)
type MediaRepository[T models.Model] interface {
	Create(ctx context.Context, item *T) (int64, error) // Create stores a new media item and returns its assigned ID
	Get(ctx context.Context, id int64) (*T, error)      // Get retrieves a media item by ID
	Update(ctx context.Context, item *T) error          // Update modifies an existing media item
	Delete(ctx context.Context, id int64) error         // Delete removes a media item by ID
	List(ctx context.Context, opts any) ([]*T, error)   // List retrieves media items with optional filtering and sorting
	Count(ctx context.Context, opts any) (int64, error) // Count returns the number of media items matching conditions
}

// StatusFilterable extends MediaRepository with status-based filtering for queries like "queued", "reading", "watching", "watched", "finished"
type StatusFilterable[T models.Model] interface {
	MediaRepository[T]
	// GetByStatus retrieves all items with the given status
	GetByStatus(ctx context.Context, status string) ([]*T, error)
}
