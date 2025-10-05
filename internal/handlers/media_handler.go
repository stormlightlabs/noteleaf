package handlers

import (
	"context"
	"io"
)

// MediaHandler defines common operations for media handlers and captures the shared behavior across media handlers for polymorphic handling of different media types.
type MediaHandler interface {
	SearchAndAdd(ctx context.Context, query string, interactive bool) error // SearchAndAdd searches for media and allows user to select and add to queue
	List(ctx context.Context, status string) error                          // List lists all media items with optional status filtering
	UpdateStatus(ctx context.Context, id, status string) error              // UpdateStatus changes the status of a media item
	Remove(ctx context.Context, id string) error                            // Remove removes a media item from the queue
	SetInputReader(reader io.Reader)                                        // SetInputReader sets the input reader for interactive prompts
	Close() error                                                           // Close cleans up resources
}

// Searchable defines search behavior for media handlers
type Searchable interface {
	SearchAndAdd(ctx context.Context, query string, interactive bool) error
}

// Listable defines list behavior for media handlers
type Listable interface {
	List(ctx context.Context, status string) error
}

// StatusUpdatable defines status update behavior for media handlers
type StatusUpdatable interface {
	UpdateStatus(ctx context.Context, id, status string) error
}

// Removable defines remove behavior for media handlers
type Removable interface {
	Remove(ctx context.Context, id string) error
}
