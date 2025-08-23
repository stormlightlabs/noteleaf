// Movies & TV: Rotten Tomatoes with colly
//
// Music: Album of the Year with chromedp
//
// Books: OpenLibrary API
package services

import (
	"context"

	"github.com/stormlightlabs/noteleaf/internal/models"
)

// APIService defines the contract for API interactions
type APIService interface {
	Get(ctx context.Context, id string) (*models.Model, error)
	Search(ctx context.Context, page, limit int) ([]*models.Model, error)
	Check(ctx context.Context) error
	Close() error
}
