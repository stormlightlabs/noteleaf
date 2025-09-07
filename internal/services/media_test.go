package services

import (
	"bytes"
	_ "embed"
	"strings"
	"testing"
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

func TestMediaService(t *testing.T) {
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
