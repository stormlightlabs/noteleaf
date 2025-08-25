// Movies & TV: Rotten Tomatoes with colly
//
// Music: Album of the Year with chromedp
//
// Books: OpenLibrary API
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/stormlightlabs/noteleaf/internal/models"
	"golang.org/x/time/rate"
)

const (
	// Open Library API endpoints
	openLibraryBaseURL = "https://openlibrary.org"
	openLibrarySearch  = openLibraryBaseURL + "/search.json"

	// Rate limiting: 180 requests per minute = 3 requests per second
	requestsPerSecond = 3
	burstLimit        = 5

	// User agent
	// TODO: See https://www.digitalocean.com/community/tutorials/using-ldflags-to-set-version-information-for-go-applications
	userAgent string = "Noteleaf/1.0.0 (info@stormlightlabs.org)"
)

// APIService defines the contract for API interactions
type APIService interface {
	Get(ctx context.Context, id string) (*models.Model, error)
	Search(ctx context.Context, query string, page, limit int) ([]*models.Model, error)
	Check(ctx context.Context) error
	Close() error
}

// BookService implements APIService for Open Library
type BookService struct {
	client  *http.Client
	limiter *rate.Limiter
	baseURL string // Allow configurable base URL for testing
}

// NewBookService creates a new book service with rate limiting
func NewBookService() *BookService {
	return &BookService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		limiter: rate.NewLimiter(rate.Limit(requestsPerSecond), burstLimit),
		baseURL: openLibraryBaseURL,
	}
}

// NewBookServiceWithBaseURL creates a book service with custom base URL (for testing)
func NewBookServiceWithBaseURL(baseURL string) *BookService {
	return &BookService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		limiter: rate.NewLimiter(rate.Limit(requestsPerSecond), burstLimit),
		baseURL: baseURL,
	}
}

// OpenLibrarySearchResponse represents the search response from Open Library
type OpenLibrarySearchResponse struct {
	NumFound      int                    `json:"numFound"`
	Start         int                    `json:"start"`
	NumFoundExact bool                   `json:"numFoundExact"`
	Docs          []OpenLibrarySearchDoc `json:"docs"`
}

// OpenLibrarySearchDoc represents a book document in search results
type OpenLibrarySearchDoc struct {
	Key              string   `json:"key"`
	Title            string   `json:"title"`
	AuthorName       []string `json:"author_name"`
	FirstPublishYear int      `json:"first_publish_year"`
	PublishYear      []int    `json:"publish_year"`
	Edition_count    int      `json:"edition_count"`
	ISBN             []string `json:"isbn"`
	PublisherName    []string `json:"publisher"`
	Subject          []string `json:"subject"`
	CoverI           int      `json:"cover_i"`
	HasFulltext      bool     `json:"has_fulltext"`
	PublicScanB      bool     `json:"public_scan_b"`
	ReadinglogCount  int      `json:"readinglog_count"`
	WantToReadCount  int      `json:"want_to_read_count"`
	CurrentlyReading int      `json:"currently_reading_count"`
	AlreadyReadCount int      `json:"already_read_count"`
}

// OpenLibraryWork represents a work details from Open Library
type OpenLibraryWork struct {
	Key              string                 `json:"key"`
	Title            string                 `json:"title"`
	Authors          []OpenLibraryAuthorRef `json:"authors"`
	Description      any                    `json:"description"` // Can be string or object
	Subjects         []string               `json:"subjects"`
	Covers           []int                  `json:"covers"`
	FirstPublishDate string                 `json:"first_publish_date"`
}

// OpenLibraryAuthorRef represents an author reference in a work
type OpenLibraryAuthorRef struct {
	Author OpenLibraryAuthorKey `json:"author"`
	Type   OpenLibraryType      `json:"type"`
}

// OpenLibraryAuthorKey represents an author key
type OpenLibraryAuthorKey struct {
	Key string `json:"key"`
}

// OpenLibraryType represents a type reference
type OpenLibraryType struct {
	Key string `json:"key"`
}

// Search searches for books using the Open Library API
func (bs *BookService) Search(ctx context.Context, query string, page, limit int) ([]*models.Model, error) {
	if err := bs.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %w", err)
	}

	// Build search URL
	params := url.Values{}
	params.Add("q", query)
	params.Add("offset", strconv.Itoa((page-1)*limit))
	params.Add("limit", strconv.Itoa(limit))
	params.Add("fields", "key,title,author_name,first_publish_year,edition_count,isbn,publisher,subject,cover_i,has_fulltext")

	searchURL := bs.baseURL + "/search.json?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := bs.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var searchResp OpenLibrarySearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to models
	var books []*models.Model
	for _, doc := range searchResp.Docs {
		book := bs.searchDocToBook(doc)
		var model models.Model = book
		books = append(books, &model)
	}

	return books, nil
}

// Get retrieves a specific book by Open Library work key
func (bs *BookService) Get(ctx context.Context, id string) (*models.Model, error) {
	if err := bs.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %w", err)
	}

	// Ensure id starts with /works/
	workKey := id
	if !strings.HasPrefix(workKey, "/works/") {
		workKey = "/works/" + id
	}

	workURL := bs.baseURL + workKey + ".json"

	req, err := http.NewRequestWithContext(ctx, "GET", workURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := bs.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("book not found: %s", id)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var work OpenLibraryWork
	if err := json.NewDecoder(resp.Body).Decode(&work); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	book := bs.workToBook(work)
	var model models.Model = book
	return &model, nil
}

// Check verifies the API connection
func (bs *BookService) Check(ctx context.Context) error {
	if err := bs.limiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit wait failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", bs.baseURL+"/search.json?q=test&limit=1", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)

	resp, err := bs.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Open Library: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("open Library API returned status %d", resp.StatusCode)
	}

	return nil
}

// Close cleans up the service resources
//
//	HTTP client doesn't need explicit cleanup
func (bs *BookService) Close() error {
	return nil
}

// Helper functions

func (bs *BookService) searchDocToBook(doc OpenLibrarySearchDoc) *models.Book {
	book := &models.Book{
		Title:  doc.Title,
		Status: "queued",
		Added:  time.Now(),
	}

	if len(doc.AuthorName) > 0 {
		book.Author = strings.Join(doc.AuthorName, ", ")
	}

	// Set publication year as pages (approximation)
	if doc.FirstPublishYear > 0 {
		// We don't have page count, so we'll leave it as 0
		// Could potentially estimate based on edition count or other factors
	}

	var notes []string
	if doc.Edition_count > 0 {
		notes = append(notes, fmt.Sprintf("%d editions", doc.Edition_count))
	}
	if len(doc.PublisherName) > 0 {
		notes = append(notes, "Publishers: "+strings.Join(doc.PublisherName, ", "))
	}
	if doc.CoverI > 0 {
		notes = append(notes, fmt.Sprintf("Cover ID: %d", doc.CoverI))
	}

	if len(notes) > 0 {
		book.Notes = strings.Join(notes, " | ")
	}

	return book
}

func (bs *BookService) workToBook(work OpenLibraryWork) *models.Book {
	book := &models.Book{
		Title:  work.Title,
		Status: "queued",
		Added:  time.Now(),
	}

	// Extract author names (would need additional API calls to get full names)
	if len(work.Authors) > 0 {
		// For now, just use the keys
		var authorKeys []string
		for _, author := range work.Authors {
			key := strings.TrimPrefix(author.Author.Key, "/authors/")
			authorKeys = append(authorKeys, key)
		}
		book.Author = strings.Join(authorKeys, ", ")
	}

	if work.Description != nil {
		switch desc := work.Description.(type) {
		case string:
			book.Notes = desc
		case map[string]any:
			if value, ok := desc["value"].(string); ok {
				book.Notes = value
			}
		}
	}

	if book.Notes == "" && len(work.Subjects) > 0 {
		subjects := work.Subjects
		if len(subjects) > 5 {
			subjects = subjects[:5]
		}
		book.Notes = "Subjects: " + strings.Join(subjects, ", ")
	}

	return book
}
