package repo

import (
	"context"

	"stormlightlabs.org/noteleaf/internal/models"
)

// Repository defines a general, behavior-focused interface for data access
type Repository interface {
	// Create stores a new model and returns its assigned ID
	Create(ctx context.Context, model models.Model) (int64, error)

	// Get retrieves a model by ID
	Get(ctx context.Context, table string, id int64, dest models.Model) error

	// Update modifies an existing model
	Update(ctx context.Context, model models.Model) error

	// Delete removes a model by ID
	Delete(ctx context.Context, table string, id int64) error

	// List retrieves models with optional filtering and sorting
	List(ctx context.Context, table string, opts ListOptions, dest any) error

	// Find retrieves models matching specific conditions
	Find(ctx context.Context, table string, conditions map[string]any, dest any) error

	// Count returns the number of models matching conditions
	Count(ctx context.Context, table string, conditions map[string]any) (int64, error)

	// Execute runs a custom query with parameters
	Execute(ctx context.Context, query string, args ...any) error

	// Query runs a custom query and returns results
	Query(ctx context.Context, query string, dest any, args ...any) error
}

// ListOptions defines generic options for listing items
type ListOptions struct {
	// field: value pairs for WHERE conditions
	Where  map[string]any
	Limit  int
	Offset int
	// field name to sort by
	SortBy string
	// "asc" or "desc"
	SortOrder string
	// general search term
	Search string
	// fields to search in
	SearchFields []string
}
