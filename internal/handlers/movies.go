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

// MovieHandler handles all movie-related commands
type MovieHandler struct {
	db      *store.Database
	config  *store.Config
	repos   *repo.Repositories
	service *services.MovieService
}

// NewMovieHandler creates a new movie handler
func NewMovieHandler() (*MovieHandler, error) {
	db, err := store.NewDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	config, err := store.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	repos := repo.NewRepositories(db.DB)
	service := services.NewMovieService()

	return &MovieHandler{
		db:      db,
		config:  config,
		repos:   repos,
		service: service,
	}, nil
}

// Close cleans up resources
func (h *MovieHandler) Close() error {
	if err := h.service.Close(); err != nil {
		return fmt.Errorf("failed to close service: %w", err)
	}
	return h.db.Close()
}

// SearchAndAdd searches for movies and allows user to select and add to queue
func (h *MovieHandler) SearchAndAdd(ctx context.Context, query string, interactive bool) error {
	if query == "" {
		return fmt.Errorf("search query cannot be empty")
	}

	fmt.Printf("Searching for movies: %s\n", query)
	fmt.Print("Loading...")

	results, err := h.service.Search(ctx, query, 1, 5)
	if err != nil {
		fmt.Println(" failed!")
		return fmt.Errorf("search failed: %w", err)
	}

	fmt.Println(" done!")
	fmt.Println()

	if len(results) == 0 {
		fmt.Println("No movies found.")
		return nil
	}

	fmt.Printf("Found %d result(s):\n\n", len(results))
	for i, result := range results {
		if movie, ok := (*result).(*models.Movie); ok {
			fmt.Printf("[%d] %s", i+1, movie.Title)
			if movie.Year > 0 {
				fmt.Printf(" (%d)", movie.Year)
			}
			if movie.Rating > 0 {
				fmt.Printf(" ★%.1f", movie.Rating)
			}
			if movie.Notes != "" {
				notes := movie.Notes
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

	selectedMovie, ok := (*results[choice-1]).(*models.Movie)
	if !ok {
		return fmt.Errorf("error processing selected movie")
	}

	if _, err := h.repos.Movies.Create(ctx, selectedMovie); err != nil {
		return fmt.Errorf("failed to add movie: %w", err)
	}

	fmt.Printf("✓ Added movie: %s", selectedMovie.Title)
	if selectedMovie.Year > 0 {
		fmt.Printf(" (%d)", selectedMovie.Year)
	}
	fmt.Println()

	return nil
}

// List movies with status filtering
func (h *MovieHandler) List(ctx context.Context, status string) error {
	var movies []*models.Movie
	var err error

	switch status {
	case "":
		movies, err = h.repos.Movies.List(ctx, repo.MovieListOptions{})
		if err != nil {
			return fmt.Errorf("failed to list movies: %w", err)
		}
	case "queued":
		movies, err = h.repos.Movies.GetQueued(ctx)
		if err != nil {
			return fmt.Errorf("failed to get queued movies: %w", err)
		}
	case "watched":
		movies, err = h.repos.Movies.GetWatched(ctx)
		if err != nil {
			return fmt.Errorf("failed to get watched movies: %w", err)
		}
	default:
		return fmt.Errorf("invalid status: %s (use: queued, watched, or leave empty for all)", status)
	}

	if len(movies) == 0 {
		if status == "" {
			fmt.Println("No movies found")
		} else {
			fmt.Printf("No %s movies found\n", status)
		}
		return nil
	}

	fmt.Printf("Found %d movie(s):\n\n", len(movies))
	for _, movie := range movies {
		h.printMovie(movie)
	}

	return nil
}

// View displays detailed information about a specific movie
func (h *MovieHandler) View(ctx context.Context, movieID int64) error {
	movie, err := h.repos.Movies.Get(ctx, movieID)
	if err != nil {
		return fmt.Errorf("failed to get movie %d: %w", movieID, err)
	}

	fmt.Printf("Movie: %s", movie.Title)
	if movie.Year > 0 {
		fmt.Printf(" (%d)", movie.Year)
	}
	fmt.Printf("\nID: %d\n", movie.ID)
	fmt.Printf("Status: %s\n", movie.Status)

	if movie.Rating > 0 {
		fmt.Printf("Rating: ★%.1f\n", movie.Rating)
	}

	fmt.Printf("Added: %s\n", movie.Added.Format("2006-01-02 15:04:05"))

	if movie.Watched != nil {
		fmt.Printf("Watched: %s\n", movie.Watched.Format("2006-01-02 15:04:05"))
	}

	if movie.Notes != "" {
		fmt.Printf("Notes: %s\n", movie.Notes)
	}

	return nil
}

// UpdateStatus changes the status of a movie
func (h *MovieHandler) UpdateStatus(ctx context.Context, movieID int64, status string) error {
	validStatuses := []string{"queued", "watched", "removed"}
	if !slices.Contains(validStatuses, status) {
		return fmt.Errorf("invalid status: %s (valid: %s)", status, strings.Join(validStatuses, ", "))
	}

	movie, err := h.repos.Movies.Get(ctx, movieID)
	if err != nil {
		return fmt.Errorf("movie %d not found: %w", movieID, err)
	}

	movie.Status = status
	if status == "watched" && movie.Watched == nil {
		now := time.Now()
		movie.Watched = &now
	}

	if err := h.repos.Movies.Update(ctx, movie); err != nil {
		return fmt.Errorf("failed to update movie status: %w", err)
	}

	fmt.Printf("✓ Movie '%s' marked as %s\n", movie.Title, status)
	return nil
}

// MarkWatched marks a movie as watched
func (h *MovieHandler) MarkWatched(ctx context.Context, movieID int64) error {
	return h.UpdateStatus(ctx, movieID, "watched")
}

// Remove removes a movie from the queue
func (h *MovieHandler) Remove(ctx context.Context, movieID int64) error {
	movie, err := h.repos.Movies.Get(ctx, movieID)
	if err != nil {
		return fmt.Errorf("movie %d not found: %w", movieID, err)
	}

	if err := h.repos.Movies.Delete(ctx, movieID); err != nil {
		return fmt.Errorf("failed to remove movie: %w", err)
	}

	fmt.Printf("✓ Removed movie: %s", movie.Title)
	if movie.Year > 0 {
		fmt.Printf(" (%d)", movie.Year)
	}
	fmt.Println()

	return nil
}

func (h *MovieHandler) printMovie(movie *models.Movie) {
	fmt.Printf("[%d] %s", movie.ID, movie.Title)
	if movie.Year > 0 {
		fmt.Printf(" (%d)", movie.Year)
	}
	if movie.Status != "queued" {
		fmt.Printf(" (%s)", movie.Status)
	}
	if movie.Rating > 0 {
		fmt.Printf(" ★%.1f", movie.Rating)
	}
	fmt.Println()
}

// SearchAndAddMovie searches for movies and allows user to select and add to queue
func (h *MovieHandler) SearchAndAddMovie(ctx context.Context, query string, interactive bool) error {
	return h.SearchAndAdd(ctx, query, interactive)
}

// ListMovies lists all movies in the queue with status filtering
func (h *MovieHandler) ListMovies(ctx context.Context, status string) error {
	return h.List(ctx, status)
}

// ViewMovie displays detailed information about a specific movie
func (h *MovieHandler) ViewMovie(ctx context.Context, id string) error {
	movieID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid movie ID: %s", id)
	}
	return h.View(ctx, movieID)
}

// UpdateMovieStatus changes the status of a movie
func (h *MovieHandler) UpdateMovieStatus(ctx context.Context, id, status string) error {
	movieID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid movie ID: %s", id)
	}
	return h.UpdateStatus(ctx, movieID, status)
}

// MarkMovieWatched marks a movie as watched
func (h *MovieHandler) MarkMovieWatched(ctx context.Context, id string) error {
	movieID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid movie ID: %s", id)
	}
	return h.MarkWatched(ctx, movieID)
}

// RemoveMovie removes a movie from the queue
func (h *MovieHandler) RemoveMovie(ctx context.Context, id string) error {
	movieID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid movie ID: %s", id)
	}

	return h.Remove(ctx, movieID)
}
