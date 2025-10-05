package handlers

import (
	"context"
	"io"
)

// Core Handler Behavior Interfaces
//
// These interfaces define composable behaviors that handlers can implement
// to provide consistent functionality across different domain handlers.

// Closeable manages resource cleanup for handlers
//
// Implemented by all handlers that manage database connections, services, or other resources
type Closeable interface {
	Close() error
}

// Viewable retrieves and displays a single item by ID (int64)
//
// Implemented by handlers that support viewing individual items with integer IDs
type Viewable interface {
	View(ctx context.Context, id int64) error
}

// ViewableByString retrieves and displays a single item by ID (string)
//
// Implemented by handlers that support viewing items with string IDs
type ViewableByString interface {
	View(ctx context.Context, id string) error
}

// Note: Listable already defined in media_handler.go
// We reuse that interface: List(ctx context.Context, status string) error

// Removable handles item removal (already defined in media_handler.go)
// We extend it with type-specific variants for consistency

// RemovableByInt64 handles removal by integer ID
type RemovableByInt64 interface {
	Remove(ctx context.Context, id int64) error
}

// DeletableByInt64 handles deletion by integer ID
type DeletableByInt64 interface {
	Delete(ctx context.Context, id int64) error
}

// InputReader provides testable input handling
//
// Handlers implementing this interface can have their input source replaced for testing
type InputReader interface {
	SetInputReader(reader io.Reader)
}

// InteractiveSupport indicates handler support for interactive/TUI modes
//
// Handlers can check if they're running in test mode and adjust behavior accordingly
type InteractiveSupport interface {
	SupportsInteractive() bool
}

// Renderable provides content rendering capabilities
//
// Implemented by handlers that need to render markdown or other formatted content
type Renderable interface {
	Render(content string) (string, error)
}

// MediaHandler is already defined in media_handler.go
// It composes: Searchable, Listable, StatusUpdatable, Removable, InputReader, Closeable

// Compile-time interface checks - verifying shared behaviors across handlers
//
// Note: Handlers have domain-specific List() signatures, so we cannot enforce
// a common Listable interface. Each handler's List method is tailored to its domain:
// - ArticleHandler: List(ctx, query, author, limit)
// - NoteHandler: List(ctx, static, showArchived, tags)
// - TaskHandler: List(ctx, static, showAll, status, priority, project, context)
// - Media handlers: List(ctx, status) via MediaHandler.Listable
var (
	// All handlers implement Closeable for resource cleanup
	_ Closeable = (*ArticleHandler)(nil)
	_ Closeable = (*NoteHandler)(nil)
	_ Closeable = (*TaskHandler)(nil)
	_ Closeable = (*BookHandler)(nil)
	_ Closeable = (*MovieHandler)(nil)
	_ Closeable = (*TVHandler)(nil)

	// Handlers with View by ID (int64)
	_ Viewable = (*ArticleHandler)(nil)
	_ Viewable = (*MovieHandler)(nil)

	// Handlers with Remove by ID (int64)
	_ RemovableByInt64 = (*ArticleHandler)(nil)

	// Handlers with Delete by ID (int64)
	_ DeletableByInt64 = (*NoteHandler)(nil)

	// Media handlers implement MediaHandler (defined in media_handler.go)
	_ MediaHandler = (*BookHandler)(nil)
	_ MediaHandler = (*MovieHandler)(nil)
	_ MediaHandler = (*TVHandler)(nil)

	// Media handlers support input reading for testing
	_ InputReader = (*BookHandler)(nil)
	_ InputReader = (*MovieHandler)(nil)
	_ InputReader = (*TVHandler)(nil)
)
