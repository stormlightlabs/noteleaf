package handlers

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/stormlightlabs/noteleaf/internal/models"
	"github.com/stormlightlabs/noteleaf/internal/repo"
	"github.com/stormlightlabs/noteleaf/internal/services"
	"github.com/stormlightlabs/noteleaf/internal/store"
)

// TVHandler handles all TV show-related commands
type TVHandler struct {
	db      *store.Database
	config  *store.Config
	repos   *repo.Repositories
	service *services.TVService
}

// NewTVHandler creates a new TV handler
func NewTVHandler() (*TVHandler, error) {
	db, err := store.NewDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	config, err := store.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	repos := repo.NewRepositories(db.DB)
	service := services.NewTVService()

	return &TVHandler{
		db:      db,
		config:  config,
		repos:   repos,
		service: service,
	}, nil
}

// Close cleans up resources
func (h *TVHandler) Close() error {
	if err := h.service.Close(); err != nil {
		return fmt.Errorf("failed to close service: %w", err)
	}
	return h.db.Close()
}

// SearchAndAdd searches for TV shows and allows user to select and add to queue
func (h *TVHandler) SearchAndAdd(ctx context.Context, query string, interactive bool) error {
	if query == "" {
		return fmt.Errorf("search query cannot be empty")
	}

	fmt.Printf("Searching for TV shows: %s\n", query)
	fmt.Print("Loading...")

	results, err := h.service.Search(ctx, query, 1, 5)
	if err != nil {
		fmt.Println(" failed!")
		return fmt.Errorf("search failed: %w", err)
	}

	fmt.Println(" done!")
	fmt.Println()

	if len(results) == 0 {
		fmt.Println("No TV shows found.")
		return nil
	}

	fmt.Printf("Found %d result(s):\n\n", len(results))
	for i, result := range results {
		if show, ok := (*result).(*models.TVShow); ok {
			fmt.Printf("[%d] %s", i+1, show.Title)
			if show.Season > 0 {
				fmt.Printf(" (Season %d)", show.Season)
			}
			if show.Rating > 0 {
				fmt.Printf(" ★%.1f", show.Rating)
			}
			if show.Notes != "" {
				notes := show.Notes
				if len(notes) > 80 {
					notes = notes[:77] + "..."
				}
				fmt.Printf("\n    %s", notes)
			}
			fmt.Println()
		}
	}

	fmt.Print("\nEnter number to add (1-", len(results), "), or 0 to cancel: ")

	var choice int
	if _, err := fmt.Scanf("%d", &choice); err != nil {
		return fmt.Errorf("invalid input")
	}

	if choice == 0 {
		fmt.Println("Cancelled.")
		return nil
	}

	if choice < 1 || choice > len(results) {
		return fmt.Errorf("invalid choice: %d", choice)
	}

	selectedShow, ok := (*results[choice-1]).(*models.TVShow)
	if !ok {
		return fmt.Errorf("error processing selected TV show")
	}

	if _, err := h.repos.TV.Create(ctx, selectedShow); err != nil {
		return fmt.Errorf("failed to add TV show: %w", err)
	}

	fmt.Printf("✓ Added TV show: %s", selectedShow.Title)
	if selectedShow.Season > 0 {
		fmt.Printf(" (Season %d)", selectedShow.Season)
	}
	fmt.Println()

	return nil
}

// List TV shows with status filtering
func (h *TVHandler) List(ctx context.Context, status string) error {
	var shows []*models.TVShow
	var err error

	switch status {
	case "":
		shows, err = h.repos.TV.List(ctx, repo.TVListOptions{})
		if err != nil {
			return fmt.Errorf("failed to list TV shows: %w", err)
		}
	case "queued":
		shows, err = h.repos.TV.GetQueued(ctx)
		if err != nil {
			return fmt.Errorf("failed to get queued TV shows: %w", err)
		}
	case "watching":
		shows, err = h.repos.TV.GetWatching(ctx)
		if err != nil {
			return fmt.Errorf("failed to get watching TV shows: %w", err)
		}
	case "watched":
		shows, err = h.repos.TV.GetWatched(ctx)
		if err != nil {
			return fmt.Errorf("failed to get watched TV shows: %w", err)
		}
	default:
		return fmt.Errorf("invalid status: %s (use: queued, watching, watched, or leave empty for all)", status)
	}

	if len(shows) == 0 {
		if status == "" {
			fmt.Println("No TV shows found")
		} else {
			fmt.Printf("No %s TV shows found\n", status)
		}
		return nil
	}

	fmt.Printf("Found %d TV show(s):\n\n", len(shows))
	for _, show := range shows {
		h.print(show)
	}

	return nil
}

// View displays detailed information about a specific TV show
func (h *TVHandler) View(ctx context.Context, id string) error {
	showID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid TV show ID: %s", id)
	}

	show, err := h.repos.TV.Get(ctx, showID)
	if err != nil {
		return fmt.Errorf("failed to get TV show %d: %w", showID, err)
	}

	fmt.Printf("TV Show: %s", show.Title)
	if show.Season > 0 {
		fmt.Printf(" (Season %d", show.Season)
		if show.Episode > 0 {
			fmt.Printf(", Episode %d", show.Episode)
		}
		fmt.Print(")")
	}
	fmt.Printf("\nID: %d\n", show.ID)
	fmt.Printf("Status: %s\n", show.Status)

	if show.Rating > 0 {
		fmt.Printf("Rating: ★%.1f\n", show.Rating)
	}

	fmt.Printf("Added: %s\n", show.Added.Format("2006-01-02 15:04:05"))

	if show.LastWatched != nil {
		fmt.Printf("Last Watched: %s\n", show.LastWatched.Format("2006-01-02 15:04:05"))
	}

	if show.Notes != "" {
		fmt.Printf("Notes: %s\n", show.Notes)
	}

	return nil
}

// UpdateStatus changes the status of a TV show
func (h *TVHandler) UpdateStatus(ctx context.Context, showID int64, status string) error {
	validStatuses := []string{"queued", "watching", "watched", "removed"}
	if !slices.Contains(validStatuses, status) {
		return fmt.Errorf("invalid status: %s (valid: %s)", status, strings.Join(validStatuses, ", "))
	}

	show, err := h.repos.TV.Get(ctx, showID)
	if err != nil {
		return fmt.Errorf("TV show %d not found: %w", showID, err)
	}

	show.Status = status
	if (status == "watching" || status == "watched") && show.LastWatched == nil {
		now := time.Now()
		show.LastWatched = &now
	}

	if err := h.repos.TV.Update(ctx, show); err != nil {
		return fmt.Errorf("failed to update TV show status: %w", err)
	}

	fmt.Printf("✓ TV show '%s' marked as %s\n", show.Title, status)
	return nil
}

// MarkWatching marks a TV show as currently watching
func (h *TVHandler) MarkWatching(ctx context.Context, showID int64) error {
	return h.UpdateStatus(ctx, showID, "watching")
}

// MarkWatched marks a TV show as watched
func (h *TVHandler) MarkWatched(ctx context.Context, id string) error {
	showID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid TV show ID: %s", id)
	}

	return h.UpdateStatus(ctx, showID, "watched")
}

// Remove removes a TV show from the queue
func (h *TVHandler) Remove(ctx context.Context, id string) error {
	showID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid TV show ID: %s", id)
	}

	show, err := h.repos.TV.Get(ctx, showID)
	if err != nil {
		return fmt.Errorf("TV show %d not found: %w", showID, err)
	}

	if err := h.repos.TV.Delete(ctx, showID); err != nil {
		return fmt.Errorf("failed to remove TV show: %w", err)
	}

	fmt.Printf("✓ Removed TV show: %s", show.Title)
	if show.Season > 0 {
		fmt.Printf(" (Season %d)", show.Season)
	}
	fmt.Println()

	return nil
}

func (h *TVHandler) print(show *models.TVShow) {
	fmt.Printf("[%d] %s", show.ID, show.Title)
	if show.Season > 0 {
		fmt.Printf(" (Season %d", show.Season)
		if show.Episode > 0 {
			fmt.Printf(", Ep %d", show.Episode)
		}
		fmt.Print(")")
	}
	if show.Status != "queued" {
		fmt.Printf(" (%s)", show.Status)
	}
	if show.Rating > 0 {
		fmt.Printf(" ★%.1f", show.Rating)
	}
	fmt.Println()
}

// UpdateTVShowStatus changes the status of a TV show
func (h *TVHandler) UpdateTVShowStatus(ctx context.Context, id, status string) error {
	showID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid TV show ID: %s", id)
	}
	return h.UpdateStatus(ctx, showID, status)
}

// MarkTVShowWatching marks a TV show as currently watching
func (h *TVHandler) MarkTVShowWatching(ctx context.Context, id string) error {
	showID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid TV show ID: %s", id)
	}
	return h.MarkWatching(ctx, showID)
}
