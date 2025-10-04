package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/stormlightlabs/noteleaf/internal/models"
	"golang.org/x/time/rate"
)

type MediaKind string

const (
	TVKind    MediaKind = "tv"
	MovieKind MediaKind = "movie"
)

type Media struct {
	Title          string
	Link           string
	Type           MediaKind
	CriticScore    string
	CertifiedFresh bool
}

type Person struct {
	Name   string `json:"name"`
	SameAs string `json:"sameAs"`
	Image  string `json:"image"`
}

type AggregateRating struct {
	RatingValue string `json:"ratingValue"`
	RatingCount int    `json:"ratingCount"`
	ReviewCount int    `json:"reviewCount"`
}

type Season struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type PartOfSeries struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type TVSeries struct {
	Context         string          `json:"@context"`
	Type            string          `json:"@type"`
	Name            string          `json:"name"`
	URL             string          `json:"url"`
	Description     string          `json:"description"`
	Image           string          `json:"image"`
	Genre           []string        `json:"genre"`
	ContentRating   string          `json:"contentRating"`
	DateCreated     string          `json:"dateCreated"`
	NumberOfSeasons int             `json:"numberOfSeasons"`
	Actors          []Person        `json:"actor"`
	Producers       []Person        `json:"producer"`
	AggregateRating AggregateRating `json:"aggregateRating"`
	Seasons         []Season        `json:"containsSeason"`
}

type Movie struct {
	Context         string          `json:"@context"`
	Type            string          `json:"@type"`
	Name            string          `json:"name"`
	URL             string          `json:"url"`
	Description     string          `json:"description"`
	Image           string          `json:"image"`
	Genre           []string        `json:"genre"`
	ContentRating   string          `json:"contentRating"`
	DateCreated     string          `json:"dateCreated"`
	Actors          []Person        `json:"actor"`
	Directors       []Person        `json:"director"`
	Producers       []Person        `json:"producer"`
	AggregateRating AggregateRating `json:"aggregateRating"`
}

type TVSeason struct {
	Context         string          `json:"@context"`
	Type            string          `json:"@type"`
	Name            string          `json:"name"`
	URL             string          `json:"url"`
	Description     string          `json:"description"`
	Image           string          `json:"image"`
	SeasonNumber    int             `json:"seasonNumber"`
	DatePublished   string          `json:"datePublished"`
	PartOfSeries    PartOfSeries    `json:"partOfSeries"`
	AggregateRating AggregateRating `json:"aggregateRating"`
}

type MovieService struct {
	client   *http.Client
	limiter  *rate.Limiter
	fetcher  Fetchable
	searcher Searchable
	baseURL  string
}

type TVService struct {
	client   *http.Client
	limiter  *rate.Limiter
	fetcher  Fetchable
	searcher Searchable
	baseURL  string
}

// ParseSearch parses Rotten Tomatoes search results HTML into Media entries.
var ParseSearch = func(r io.Reader) ([]Media, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}

	var results []Media
	doc.Find("search-page-result").Each(func(i int, resultBlock *goquery.Selection) {
		mediaType, _ := resultBlock.Attr("type")

		resultBlock.Find("search-page-media-row").Each(func(j int, s *goquery.Selection) {
			link, _ := s.Find("a[slot='thumbnail']").Attr("href")
			if link == "" {
				link, _ = s.Find("a[slot='title']").Attr("href")
				if link == "" {
					return
				}
			}

			title := s.Find("a[slot='title']").Text()

			var itemKind MediaKind
			switch mediaType {
			case "movie":
				itemKind = MovieKind
			case "tvSeries":
				itemKind = TVKind
			default:
				if strings.HasPrefix(link, "/m/") {
					itemKind = MovieKind
				} else if strings.HasPrefix(link, "/tv/") {
					itemKind = TVKind
				}
			}

			score, _ := s.Attr("tomatometerscore")
			if score == "" {
				score = "--"
			}

			certified := false
			if v, ok := s.Attr("tomatometeriscertified"); ok && v == "true" {
				certified = true
			}

			results = append(results, Media{
				Title:          strings.TrimSpace(title),
				Link:           link,
				Type:           itemKind,
				CriticScore:    score,
				CertifiedFresh: certified,
			})
		})
	})

	return results, nil
}

var ExtractTVSeriesMetadata = func(r io.Reader) (*TVSeries, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}
	var series TVSeries
	found := false
	doc.Find("script[type='application/ld+json']").Each(func(i int, s *goquery.Selection) {
		var tmp map[string]any
		if err := json.Unmarshal([]byte(s.Text()), &tmp); err == nil {
			if t, ok := tmp["@type"].(string); ok && t == "TVSeries" {
				if err := json.Unmarshal([]byte(s.Text()), &series); err == nil {
					found = true
				}
			}
		}
	})
	if !found {
		return nil, fmt.Errorf("no TVSeries JSON-LD found")
	}
	return &series, nil
}

var ExtractMovieMetadata = func(r io.Reader) (*Movie, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}
	var movie Movie
	found := false
	doc.Find("script[type='application/ld+json']").Each(func(i int, s *goquery.Selection) {
		var tmp map[string]any
		if err := json.Unmarshal([]byte(s.Text()), &tmp); err == nil {
			if t, ok := tmp["@type"].(string); ok && t == "Movie" {
				if err := json.Unmarshal([]byte(s.Text()), &movie); err == nil {
					found = true
				}
			}
		}
	})
	if !found {
		return nil, fmt.Errorf("no Movie JSON-LD found")
	}
	return &movie, nil
}

var ExtractTVSeasonMetadata = func(r io.Reader) (*TVSeason, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}
	var season TVSeason
	found := false
	doc.Find("script[type='application/ld+json']").Each(func(i int, s *goquery.Selection) {
		var tmp map[string]any
		if err := json.Unmarshal([]byte(s.Text()), &tmp); err == nil {
			if t, ok := tmp["@type"].(string); ok && t == "TVSeason" {
				if err := json.Unmarshal([]byte(s.Text()), &season); err == nil {
					found = true
				}
			}
		}
	})
	if !found {
		return nil, fmt.Errorf("no TVSeason JSON-LD found")
	}

	if season.SeasonNumber == 0 {
		if season.URL != "" {
			parts := strings.SplitSeq(season.URL, "/")
			for part := range parts {
				if strings.HasPrefix(part, "s") && len(part) > 1 {
					if num, err := strconv.Atoi(part[1:]); err == nil {
						season.SeasonNumber = num
						break
					}
				}
			}
		}

		if season.SeasonNumber == 0 && season.Name != "" {
			parts := strings.Fields(season.Name)
			for i, part := range parts {
				if strings.ToLower(part) == "season" && i+1 < len(parts) {
					if num, err := strconv.Atoi(parts[i+1]); err == nil {
						season.SeasonNumber = num
						break
					}
				}
			}
		}
	}

	return &season, nil
}

// NewMovieService creates a new movie service with rate limiting
func NewMovieService() *MovieService {
	return NewMovieSrvWithOpts("https://www.rottentomatoes.com", &DefaultFetcher{}, &DefaultFetcher{})
}

// NewMovieSrvWithOpts creates a new movie service with custom dependencies (for testing)
func NewMovieSrvWithOpts(baseURL string, fetcher Fetchable, searcher Searchable) *MovieService {
	return &MovieService{
		client:   &http.Client{Timeout: 30 * time.Second},
		limiter:  rate.NewLimiter(rate.Limit(requestsPerSecond), burstLimit),
		baseURL:  baseURL,
		fetcher:  fetcher,
		searcher: searcher,
	}
}

// Search searches for movies on Rotten Tomatoes
func (s *MovieService) Search(ctx context.Context, query string, page, limit int) ([]*models.Model, error) {
	if err := s.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %w", err)
	}

	results, err := s.searcher.Search(query)
	if err != nil {
		return nil, fmt.Errorf("failed to search rotten tomatoes: %w", err)
	}

	var movies []*models.Model
	for _, media := range results {
		if media.Type == "movie" {
			movie := &models.Movie{
				Title:  media.Title,
				Status: "queued",
				Added:  time.Now(),
				Notes:  fmt.Sprintf("Critic Score: %s, Certified: %v, URL: %s", media.CriticScore, media.CertifiedFresh, media.Link),
			}
			var m models.Model = movie
			movies = append(movies, &m)
		}
	}

	start := (page - 1) * limit
	end := start + limit
	if start > len(movies) {
		return []*models.Model{}, nil
	}
	if end > len(movies) {
		end = len(movies)
	}

	return movies[start:end], nil
}

// Get retrieves a specific movie by its Rotten Tomatoes URL
func (s *MovieService) Get(ctx context.Context, id string) (*models.Model, error) {
	if err := s.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %w", err)
	}

	data, err := s.fetcher.MovieRequest(id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch movie: %w", err)
	}

	movie := &models.Movie{
		Title:  data.Name,
		Status: "queued",
		Added:  time.Now(),
		Notes:  data.Description,
	}

	if data.DateCreated != "" {
		if year, err := strconv.Atoi(strings.Split(data.DateCreated, "-")[0]); err == nil {
			movie.Year = year
		}
	}

	var model models.Model = movie
	return &model, nil
}

// Check verifies the API connection to Rotten Tomatoes
func (s *MovieService) Check(ctx context.Context) error {
	if err := s.limiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit wait failed: %w", err)
	}

	_, err := s.fetcher.MakeRequest(s.baseURL)
	return err
}

// Close cleans up the service resources
func (s *MovieService) Close() error {
	return nil
}

// NewTVService creates a new TV service with rate limiting
func NewTVService() *TVService {
	return NewTVServiceWithDeps("https://www.rottentomatoes.com", &DefaultFetcher{}, &DefaultFetcher{})
}

// NewTVServiceWithDeps creates a new TV service with custom dependencies (for testing)
func NewTVServiceWithDeps(baseURL string, fetcher Fetchable, searcher Searchable) *TVService {
	return &TVService{
		client:   &http.Client{Timeout: 30 * time.Second},
		limiter:  rate.NewLimiter(rate.Limit(requestsPerSecond), burstLimit),
		baseURL:  baseURL,
		fetcher:  fetcher,
		searcher: searcher,
	}
}

// Search searches for TV shows on Rotten Tomatoes
func (s *TVService) Search(ctx context.Context, query string, page, limit int) ([]*models.Model, error) {
	if err := s.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %w", err)
	}

	results, err := s.searcher.Search(query)
	if err != nil {
		return nil, fmt.Errorf("failed to search rotten tomatoes: %w", err)
	}

	var shows []*models.Model
	for _, media := range results {
		if media.Type == "tv" {
			show := &models.TVShow{
				Title:  media.Title,
				Status: "queued",
				Added:  time.Now(),
				Notes:  fmt.Sprintf("Critic Score: %s, Certified: %v, URL: %s", media.CriticScore, media.CertifiedFresh, media.Link),
			}
			var m models.Model = show
			shows = append(shows, &m)
		}
	}

	start := (page - 1) * limit
	end := start + limit
	if start > len(shows) {
		return []*models.Model{}, nil
	}
	if end > len(shows) {
		end = len(shows)
	}

	return shows[start:end], nil
}

// Get retrieves a specific TV show by its Rotten Tomatoes URL
func (s *TVService) Get(ctx context.Context, id string) (*models.Model, error) {
	if err := s.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %w", err)
	}

	seriesData, err := s.fetcher.TVRequest(id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tv series: %w", err)
	}

	show := &models.TVShow{
		Title:  seriesData.Name,
		Status: "queued",
		Added:  time.Now(),
		Notes:  seriesData.Description,
	}

	if seriesData.NumberOfSeasons > 0 {
		show.Notes = fmt.Sprintf("%s\nSeasons: %d", show.Notes, seriesData.NumberOfSeasons)
	}

	var model models.Model = show
	return &model, nil
}

// Check verifies the API connection to Rotten Tomatoes
func (s *TVService) Check(ctx context.Context) error {
	if err := s.limiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit wait failed: %w", err)
	}

	_, err := s.fetcher.MakeRequest(s.baseURL)
	return err
}

// Close cleans up the service resources
func (s *TVService) Close() error {
	return nil
}
