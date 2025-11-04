package services

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stormlightlabs/noteleaf/internal/models"
	"github.com/stormlightlabs/noteleaf/internal/shared"
)

func TestMediaServices(t *testing.T) {
	t.Run("MovieService", func(t *testing.T) {
		t.Run("Search", func(t *testing.T) {
			t.Run("successful search", func(t *testing.T) {
				cleanup := SetupSuccessfulMovieMocks(t)
				defer cleanup()

				service := CreateMovieService()
				TestMovieSearch(t, service, "Fantastic Four", "Fantastic Four")
			})

			t.Run("search returns error", func(t *testing.T) {
				cleanup := SetupFailureMocks(t, "search error")
				defer cleanup()

				service := CreateMovieService()
				_, err := service.Search(context.Background(), "error", 1, 10)
				shared.AssertErrorContains(t, err, "search error", "")
			})
		})

		t.Run("Get", func(t *testing.T) {
			t.Run("successful get", func(t *testing.T) {
				cleanup := SetupSuccessfulMovieMocks(t)
				defer cleanup()

				service := CreateMovieService()
				result, err := service.Get(context.Background(), "some-url")
				if err != nil {
					t.Fatalf("Get failed: %v", err)
				}
				movie, ok := (*result).(*models.Movie)
				if !ok {
					t.Fatalf("expected a movie model, got %T", *result)
				}
				if movie.Title == "" {
					t.Error("expected non-empty movie title")
				}
			})

			t.Run("get returns error", func(t *testing.T) {
				cleanup := SetupFailureMocks(t, "fetch error")
				defer cleanup()

				service := CreateMovieService()
				_, err := service.Get(context.Background(), "error")
				shared.AssertErrorContains(t, err, "fetch error", "")
			})
		})

		t.Run("Check", func(t *testing.T) {
			t.Run("successful check", func(t *testing.T) {
				cleanup := SetupSuccessfulMovieMocks(t)
				defer cleanup()

				service := CreateMovieService()
				err := service.Check(context.Background())
				if err != nil {
					t.Fatalf("Check failed: %v", err)
				}
			})

			t.Run("check returns error", func(t *testing.T) {
				cleanup := SetupFailureMocks(t, "html fetch error")
				defer cleanup()

				service := CreateMovieService()
				err := service.Check(context.Background())
				shared.AssertErrorContains(t, err, "html fetch error", "")
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
			if season.PartOfSeries.Name == "" {
				t.Error("expected non-empty PartOfSeries.Name")
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

	})

	t.Run("TVService", func(t *testing.T) {
		t.Run("Search", func(t *testing.T) {
			t.Run("successful search", func(t *testing.T) {
				cleanup := SetupSuccessfulTVMocks(t)
				defer cleanup()

				service := CreateTVService()
				TestTVSearch(t, service, "peacemaker", "Peacemaker")
			})

			t.Run("search returns error", func(t *testing.T) {
				cleanup := SetupFailureMocks(t, "search error")
				defer cleanup()

				service := CreateTVService()
				_, err := service.Search(context.Background(), "error", 1, 10)
				shared.AssertErrorContains(t, err, "search error", "")
			})
		})

		t.Run("Get", func(t *testing.T) {
			t.Run("successful get", func(t *testing.T) {
				cleanup := SetupSuccessfulTVMocks(t)
				defer cleanup()

				service := CreateTVService()
				result, err := service.Get(context.Background(), "some-url")
				if err != nil {
					t.Fatalf("Get failed: %v", err)
				}
				show, ok := (*result).(*models.TVShow)
				if !ok {
					t.Fatalf("expected a tv show model, got %T", *result)
				}
				if show.Title == "" {
					t.Error("expected non-empty TV show title")
				}
			})

			t.Run("get returns error", func(t *testing.T) {
				cleanup := SetupFailureMocks(t, "fetch error")
				defer cleanup()

				service := CreateTVService()
				_, err := service.Get(context.Background(), "error")
				shared.AssertErrorContains(t, err, "fetch error", "")
			})
		})

		t.Run("Check", func(t *testing.T) {
			t.Run("successful check", func(t *testing.T) {
				cleanup := SetupSuccessfulTVMocks(t)
				defer cleanup()

				service := CreateTVService()
				err := service.Check(context.Background())
				if err != nil {
					t.Fatalf("Check failed: %v", err)
				}
			})

			t.Run("check returns error", func(t *testing.T) {
				cleanup := SetupFailureMocks(t, "html fetch error")
				defer cleanup()

				service := CreateTVService()
				err := service.Check(context.Background())
				shared.AssertErrorContains(t, err, "html fetch error", "")
			})
		})
	})
}
