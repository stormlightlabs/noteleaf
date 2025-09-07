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

func TestMovieService(t *testing.T) {
	t.Run("Search", func(t *testing.T) {
		originalSearch := SearchRottenTomatoes
		defer func() { SearchRottenTomatoes = originalSearch }()

		SearchRottenTomatoes = func(q string) ([]Media, error) {
			if q == "error" {
				return nil, errors.New("search error")
			}
			if q == "Fantastic Four" {
				return ParseSearch(bytes.NewReader(MovieSearchSample))
			}
			return nil, errors.New("unexpected query")
		}

		service := NewMovieService()

		t.Run("successful search", func(t *testing.T) {
			results, err := service.Search(context.Background(), "Fantastic Four", 1, 10)
			if err != nil {
				t.Fatalf("Search failed: %v", err)
			}
			if len(results) == 0 {
				t.Fatal("expected search results, got none")
			}

			var movieFound bool
			for _, r := range results {
				m, ok := (*r).(*models.Movie)
				if !ok {
					continue
				}
				if strings.Contains(m.Title, "Fantastic Four") {
					movieFound = true
					break
				}
			}
			if !movieFound {
				t.Error("expected to find a movie in search results")
			}
		})

		t.Run("search returns error", func(t *testing.T) {
			_, err := service.Search(context.Background(), "error", 1, 10)
			if err == nil {
				t.Fatal("expected error from search, got nil")
			}
			if !strings.Contains(err.Error(), "search error") {
				t.Errorf("expected error to contain 'search error', got %v", err)
			}
		})
	})

	t.Run("Get", func(t *testing.T) {
		originalFetch := FetchMovie
		defer func() { FetchMovie = originalFetch }()

		FetchMovie = func(url string) (*Movie, error) {
			if url == "error" {
				return nil, errors.New("fetch error")
			}
			return ExtractMovieMetadata(bytes.NewReader(MovieSample))
		}

		service := NewMovieService()

		t.Run("successful get", func(t *testing.T) {
			result, err := service.Get(context.Background(), "some-url")
			if err != nil {
				t.Fatalf("Get failed: %v", err)
			}
			movie, ok := (*result).(*models.Movie)
			if !ok {
				t.Fatalf("expected a movie model, got %T", *result)
			}
			if movie.Title != "The Fantastic Four: First Steps" {
				t.Errorf("expected title 'The Fantastic Four: First Steps', got '%s'", movie.Title)
			}
		})

		t.Run("get returns error", func(t *testing.T) {
			_, err := service.Get(context.Background(), "error")
			if err == nil {
				t.Fatal("expected error from get, got nil")
			}
			if !strings.Contains(err.Error(), "fetch error") {
				t.Errorf("expected error to contain 'fetch error', got %v", err)
			}
		})
	})

	t.Run("Check", func(t *testing.T) {
		originalFetchHTML := FetchHTML
		defer func() { FetchHTML = originalFetchHTML }()

		service := NewMovieService()

		t.Run("successful check", func(t *testing.T) {
			FetchHTML = func(url string) (string, error) {
				return "ok", nil
			}
			err := service.Check(context.Background())
			if err != nil {
				t.Fatalf("Check failed: %v", err)
			}
		})

		t.Run("check returns error", func(t *testing.T) {
			FetchHTML = func(url string) (string, error) {
				return "", errors.New("html fetch error")
			}
			err := service.Check(context.Background())
			if err == nil {
				t.Fatal("expected error from check, got nil")
			}
			if !strings.Contains(err.Error(), "html fetch error") {
				t.Errorf("expected error to contain 'html fetch error', got %v", err)
			}
		})
	})

	t.Run("Parse Search results", func(t *testing.T) {
		results, err := ParseSearch(bytes.NewReader(SearchSample))
		if err != nil {
			t.Fatalf("ParseSearch failed: %v", err)
		}
		if len(results) == 0 {
			t.Fatal("expected non-empty search results")
		}
	})

	t.Run("Parse Search error", func(t *testing.T) {
		results, err := ParseSearch(strings.NewReader("\x00bad html"))
		if err != nil {
			t.Fatalf("unexpected error for malformed HTML: %v", err)
		}
		if len(results) != 0 {
			t.Errorf("expected 0 results for malformed HTML, got %d", len(results))
		}

		html := `<a class="score-list-item"><span>Test</span></a>`
		results, err = ParseSearch(strings.NewReader(html))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) != 0 {
			t.Errorf("expected 0 results, got %d", len(results))
		}
	})

	t.Run("Extract Metadata", func(t *testing.T) {
		movie, err := ExtractMovieMetadata(bytes.NewReader(MovieSample))
		if err != nil {
			t.Fatalf("ExtractMovieMetadata failed: %v", err)
		}
		if movie.Type != "Movie" {
			t.Errorf("expected Type=Movie, got %s", movie.Type)
		}
		if movie.Name == "" {
			t.Error("expected non-empty Name")
		}
	})

	t.Run("Extract Metadata Errors", func(t *testing.T) {
		if _, err := ExtractMovieMetadata(strings.NewReader("not html")); err == nil {
			t.Error("expected error for invalid HTML")
		}

		html := `<script type="application/ld+json">{"@type":"Other"}</script>`
		if _, err := ExtractMovieMetadata(strings.NewReader(html)); err == nil || !strings.Contains(err.Error(), "no Movie JSON-LD") {
			t.Errorf("expected 'no Movie JSON-LD', got %v", err)
		}

		html = `<script type="application/ld+json">{oops}</script>`
		if _, err := ExtractMovieMetadata(strings.NewReader(html)); err == nil {
			t.Error("expected error for invalid JSON")
		}
	})

	t.Run("Extract TV Series Metadata", func(t *testing.T) {
		series, err := ExtractTVSeriesMetadata(bytes.NewReader(SeriesSample))
		if err != nil {
			t.Fatalf("ExtractTVSeriesMetadata failed: %v", err)
		}
		if series.Type != "TVSeries" {
			t.Errorf("expected Type=TVSeries, got %s", series.Type)
		}
		if series.NumberOfSeasons <= 0 {
			t.Error("expected NumberOfSeasons > 0")
		}
	})

	t.Run("Extract TV Series Metadata Errors", func(t *testing.T) {
		if _, err := ExtractTVSeriesMetadata(strings.NewReader("not html")); err == nil {
			t.Error("expected error for invalid HTML")
		}

		html := `<script type="application/ld+json">{"@type":"Other"}</script>`
		if _, err := ExtractTVSeriesMetadata(strings.NewReader(html)); err == nil || !strings.Contains(err.Error(), "no TVSeries JSON-LD") {
			t.Errorf("expected 'no TVSeries JSON-LD', got %v", err)
		}
	})

	t.Run("Extract TV Series Season metadata", func(t *testing.T) {
		season, err := ExtractTVSeasonMetadata(bytes.NewReader(SeasonSample))
		if err != nil {
			t.Fatalf("ExtractTVSeasonMetadata failed: %v", err)
		}
		if season.Type != "TVSeason" {
			t.Errorf("expected Type=TVSeason, got %s", season.Type)
		}
		if season.SeasonNumber <= 0 {
			t.Error("expected SeasonNumber > 0")
		}
		if season.PartOfSeries.Name != "Peacemaker" {
			t.Errorf("expected PartOfSeries.Name=Peacemaker, got %s", season.PartOfSeries.Name)
		}
	})

	t.Run("Extract TV Series Season errors", func(t *testing.T) {
		if _, err := ExtractTVSeasonMetadata(strings.NewReader("not html")); err == nil {
			t.Error("expected error for invalid HTML")
		}

		html := `<script type="application/ld+json">{"@type":"Other"}</script>`
		if _, err := ExtractTVSeasonMetadata(strings.NewReader(html)); err == nil || !strings.Contains(err.Error(), "no TVSeason JSON-LD") {
			t.Errorf("expected 'no TVSeason JSON-LD', got %v", err)
		}
	})

	t.Run("Fetch HTML errors", func(t *testing.T) {
		if _, err := FetchHTML("://bad-url"); err == nil {
			t.Error("expected error for invalid URL")
		}
	})

}

func TestTVService(t *testing.T) {
	t.Run("Search", func(t *testing.T) {
		originalSearch := SearchRottenTomatoes
		defer func() { SearchRottenTomatoes = originalSearch }()

		SearchRottenTomatoes = func(q string) ([]Media, error) {
			if q == "error" {
				return nil, errors.New("search error")
			}
			return ParseSearch(bytes.NewReader(SearchSample))
		}

		service := NewTVService()

		t.Run("successful search", func(t *testing.T) {
			results, err := service.Search(context.Background(), "peacemaker", 1, 10)
			if err != nil {
				t.Fatalf("Search failed: %v", err)
			}
			if len(results) == 0 {
				t.Fatal("expected search results, got none")
			}

			var tvFound bool
			for _, r := range results {
				s, ok := (*r).(*models.TVShow)
				if !ok {
					continue
				}
				if strings.Contains(s.Title, "Peacemaker") {
					tvFound = true
					break
				}
			}
			if !tvFound {
				t.Error("expected to find a tv show in search results")
			}
		})

		t.Run("search returns error", func(t *testing.T) {
			_, err := service.Search(context.Background(), "error", 1, 10)
			if err == nil {
				t.Fatal("expected error from search, got nil")
			}
		})
	})

	t.Run("Get", func(t *testing.T) {
		originalFetch := FetchTVSeries
		defer func() { FetchTVSeries = originalFetch }()

		FetchTVSeries = func(url string) (*TVSeries, error) {
			if url == "error" {
				return nil, errors.New("fetch error")
			}
			return ExtractTVSeriesMetadata(bytes.NewReader(SeriesSample))
		}

		service := NewTVService()

		t.Run("successful get", func(t *testing.T) {
			result, err := service.Get(context.Background(), "some-url")
			if err != nil {
				t.Fatalf("Get failed: %v", err)
			}
			show, ok := (*result).(*models.TVShow)
			if !ok {
				t.Fatalf("expected a tv show model, got %T", *result)
			}
			if !strings.Contains(show.Title, "Peacemaker") {
				t.Errorf("expected title to contain 'Peacemaker', got '%s'", show.Title)
			}
		})

		t.Run("get returns error", func(t *testing.T) {
			_, err := service.Get(context.Background(), "error")
			if err == nil {
				t.Fatal("expected error from get, got nil")
			}
		})
	})

	t.Run("Check", func(t *testing.T) {
		originalFetchHTML := FetchHTML
		defer func() { FetchHTML = originalFetchHTML }()

		service := NewTVService()

		t.Run("successful check", func(t *testing.T) {
			FetchHTML = func(url string) (string, error) {
				return "ok", nil
			}
			err := service.Check(context.Background())
			if err != nil {
				t.Fatalf("Check failed: %v", err)
			}
		})

		t.Run("check returns error", func(t *testing.T) {
			FetchHTML = func(url string) (string, error) {
				return "", errors.New("html fetch error")
			}
			err := service.Check(context.Background())
			if err == nil {
				t.Fatal("expected error from check, got nil")
			}
		})
	})
}
