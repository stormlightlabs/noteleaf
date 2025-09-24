package services

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"strings"
	"testing"

	"github.com/stormlightlabs/noteleaf/internal/models"
)

// From: https://www.rottentomatoes.com/m/the_fantastic_four_first_steps
//
//go:embed samples/movie.html
var MovieSample []byte

// From: https://www.rottentomatoes.com/search?search=peacemaker
//
//go:embed samples/search.html
var SearchSample []byte

// From: https://www.rottentomatoes.com/tv/peacemaker_2022
//
//go:embed samples/series_overview.html
var SeriesSample []byte

// From: https://www.rottentomatoes.com/tv/peacemaker_2022/s02
//
//go:embed samples/series_season.html
var SeasonSample []byte

// From: https://www.rottentomatoes.com/search?search=Fantastic%20Four
//
//go:embed samples/movie_search.html
var MovieSearchSample []byte

// MockConfig holds configuration for mocking media services
type MockConfig struct {
	SearchResults  []Media
	SearchError    error
	MovieResult    *Movie
	MovieError     error
	TVSeriesResult *TVSeries
	TVSeriesError  error
	TVSeasonResult *TVSeason
	TVSeasonError  error
	HTMLResult     string
	HTMLError      error
}

// MockSetup contains the original function variables for restoration
type MockSetup struct {
	originalSearchRottenTomatoes func(string) ([]Media, error)
	originalFetchMovie           func(string) (*Movie, error)
	originalFetchTVSeries        func(string) (*TVSeries, error)
	originalFetchTVSeason        func(string) (*TVSeason, error)
	originalFetchHTML            func(string) (string, error)
}

// SetupMediaMocks configures mock functions for media services testing
func SetupMediaMocks(t *testing.T, config MockConfig) func() {
	t.Helper()

	setup := &MockSetup{
		originalSearchRottenTomatoes: SearchRottenTomatoes,
		originalFetchMovie:           FetchMovie,
		originalFetchTVSeries:        FetchTVSeries,
		originalFetchTVSeason:        FetchTVSeason,
		originalFetchHTML:            FetchHTML,
	}

	SearchRottenTomatoes = func(q string) ([]Media, error) {
		if config.SearchError != nil {
			return nil, config.SearchError
		}
		return config.SearchResults, nil
	}

	FetchMovie = func(url string) (*Movie, error) {
		if config.MovieError != nil {
			return nil, config.MovieError
		}
		return config.MovieResult, nil
	}

	FetchTVSeries = func(url string) (*TVSeries, error) {
		if config.TVSeriesError != nil {
			return nil, config.TVSeriesError
		}
		return config.TVSeriesResult, nil
	}

	FetchTVSeason = func(url string) (*TVSeason, error) {
		if config.TVSeasonError != nil {
			return nil, config.TVSeasonError
		}
		return config.TVSeasonResult, nil
	}

	FetchHTML = func(url string) (string, error) {
		if config.HTMLError != nil {
			return "", config.HTMLError
		}
		return config.HTMLResult, nil
	}

	return func() {
		SearchRottenTomatoes = setup.originalSearchRottenTomatoes
		FetchMovie = setup.originalFetchMovie
		FetchTVSeries = setup.originalFetchTVSeries
		FetchTVSeason = setup.originalFetchTVSeason
		FetchHTML = setup.originalFetchHTML
	}
}

// Sample data access helpers - these use the embedded samples
func GetSampleMovieSearchResults() ([]Media, error) {
	return ParseSearch(bytes.NewReader(MovieSearchSample))
}

func GetSampleSearchResults() ([]Media, error) {
	return ParseSearch(bytes.NewReader(SearchSample))
}

func GetSampleMovie() (*Movie, error) {
	return ExtractMovieMetadata(bytes.NewReader(MovieSample))
}

func GetSampleTVSeries() (*TVSeries, error) {
	return ExtractTVSeriesMetadata(bytes.NewReader(SeriesSample))
}

func GetSampleTVSeason() (*TVSeason, error) {
	return ExtractTVSeasonMetadata(bytes.NewReader(SeasonSample))
}

// SetupSuccessfulMovieMocks configures mocks for successful movie operations
func SetupSuccessfulMovieMocks(t *testing.T) func() {
	t.Helper()

	movieResults, err := GetSampleMovieSearchResults()
	if err != nil {
		t.Fatalf("failed to get sample movie results: %v", err)
	}

	movie, err := GetSampleMovie()
	if err != nil {
		t.Fatalf("failed to get sample movie: %v", err)
	}

	return SetupMediaMocks(t, MockConfig{
		SearchResults: movieResults,
		MovieResult:   movie,
		HTMLResult:    "ok",
	})
}

// SetupSuccessfulTVMocks configures mocks for successful TV operations
func SetupSuccessfulTVMocks(t *testing.T) func() {
	t.Helper()

	searchResults, err := GetSampleSearchResults()
	if err != nil {
		t.Fatalf("failed to get sample search results: %v", err)
	}

	series, err := GetSampleTVSeries()
	if err != nil {
		t.Fatalf("failed to get sample TV series: %v", err)
	}

	return SetupMediaMocks(t, MockConfig{
		SearchResults:  searchResults,
		TVSeriesResult: series,
		HTMLResult:     "ok",
	})
}

// SetupFailureMocks configures mocks that return errors
func SetupFailureMocks(t *testing.T, errorMsg string) func() {
	t.Helper()

	err := errors.New(errorMsg)
	return SetupMediaMocks(t, MockConfig{
		SearchError:   err,
		MovieError:    err,
		TVSeriesError: err,
		TVSeasonError: err,
		HTMLError:     err,
	})
}

// AssertMovieInResults checks if a movie with the given title exists in results
func AssertMovieInResults(t *testing.T, results []*models.Model, expectedTitle string) {
	t.Helper()

	for _, result := range results {
		if movie, ok := (*result).(*models.Movie); ok {
			if strings.Contains(movie.Title, expectedTitle) {
				return
			}
		}
	}
	t.Errorf("expected to find movie containing '%s' in results", expectedTitle)
}

// AssertTVShowInResults checks if a TV show with the given title exists in results
func AssertTVShowInResults(t *testing.T, results []*models.Model, expectedTitle string) {
	t.Helper()

	for _, result := range results {
		if show, ok := (*result).(*models.TVShow); ok {
			if strings.Contains(show.Title, expectedTitle) {
				return // Found it
			}
		}
	}
	t.Errorf("expected to find TV show containing '%s' in results", expectedTitle)
}

// AssertErrorContains checks that an error contains the expected message
func AssertErrorContains(t *testing.T, err error, expectedMsg string) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected error containing '%s', got nil", expectedMsg)
	}
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("expected error to contain '%s', got '%v'", expectedMsg, err)
	}
}

// CreateMovieService returns a new movie service for testing
func CreateMovieService() *MovieService {
	return NewMovieService()
}

// CreateTVService returns a new TV service for testing
func CreateTVService() *TVService {
	return NewTVService()
}

// TestMovieSearch runs a standard movie search test
func TestMovieSearch(t *testing.T, service *MovieService, query string, expectedTitleFragment string) {
	t.Helper()

	results, err := service.Search(context.Background(), query, 1, 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected search results, got none")
	}

	AssertMovieInResults(t, results, expectedTitleFragment)
}

// TestTVSearch runs a standard TV search test
func TestTVSearch(t *testing.T, service *TVService, query string, expectedTitleFragment string) {
	t.Helper()

	results, err := service.Search(context.Background(), query, 1, 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected search results, got none")
	}

	AssertTVShowInResults(t, results, expectedTitleFragment)
}
