package services

import (
	"errors"
	"io"
	"testing"
)

func TestHTTPFuncs(t *testing.T) {
	origFetch := FetchHTML
	origMovie := ExtractMovieMetadata
	origTV := ExtractTVSeriesMetadata
	origSeason := ExtractTVSeasonMetadata
	origSearch := ParseSearch

	defer func() {
		FetchHTML = origFetch
		ExtractMovieMetadata = origMovie
		ExtractTVSeriesMetadata = origTV
		ExtractTVSeasonMetadata = origSeason
		ParseSearch = origSearch
	}()

	tests := []struct {
		name      string
		setup     func()
		call      func() error
		expectErr bool
	}{
		{
			name: "FetchMovie success",
			setup: func() {
				FetchHTML = func(url string) (string, error) {
					return "<html>movie</html>", nil
				}
				ExtractMovieMetadata = func(r io.Reader) (*Movie, error) {
					return &Movie{Name: "Fake Movie"}, nil
				}
			},
			call: func() error {
				m, err := FetchMovie("http://fake")
				if err != nil {
					return err
				}
				if m.Name != "Fake Movie" {
					return errors.New("unexpected movie title")
				}
				return nil
			},
		},
		{
			name: "FetchMovie error from FetchHTML",
			setup: func() {
				FetchHTML = func(url string) (string, error) {
					return "", errors.New("boom")
				}
			},
			call: func() error {
				_, err := FetchMovie("http://fake")
				return err
			},
			expectErr: true,
		},
		{
			name: "FetchTVSeries success",
			setup: func() {
				FetchHTML = func(url string) (string, error) {
					return "<html>tv</html>", nil
				}
				ExtractTVSeriesMetadata = func(r io.Reader) (*TVSeries, error) {
					return &TVSeries{Name: "Fake Series"}, nil
				}
			},
			call: func() error {
				tv, err := FetchTVSeries("http://fake")
				if err != nil {
					return err
				}
				if tv.Name != "Fake Series" {
					return errors.New("unexpected series name")
				}
				return nil
			},
		},
		{
			name: "FetchTVSeries error from FetchHTML",
			setup: func() {
				FetchHTML = func(url string) (string, error) {
					return "", errors.New("boom")
				}
			},
			call: func() error {
				_, err := FetchTVSeries("http://fake")
				return err
			},
			expectErr: true,
		},
		{
			name: "FetchTVSeason success",
			setup: func() {
				FetchHTML = func(url string) (string, error) {
					return "<html>season</html>", nil
				}
				ExtractTVSeasonMetadata = func(r io.Reader) (*TVSeason, error) {
					return &TVSeason{SeasonNumber: 1}, nil
				}
			},
			call: func() error {
				season, err := FetchTVSeason("http://fake")
				if err != nil {
					return err
				}
				if season.SeasonNumber != 1 {
					return errors.New("unexpected season number")
				}
				return nil
			},
		},
		{
			name: "FetchTVSeason error from FetchHTML",
			setup: func() {
				FetchHTML = func(url string) (string, error) {
					return "", errors.New("boom")
				}
			},
			call: func() error {
				_, err := FetchTVSeason("http://fake")
				return err
			},
			expectErr: true,
		},
		{
			name: "SearchRottenTomatoes success",
			setup: func() {
				FetchHTML = func(url string) (string, error) {
					return "<html>search</html>", nil
				}
				ParseSearch = func(r io.Reader) ([]Media, error) {
					return []Media{{Title: "Fake Result"}}, nil
				}
			},
			call: func() error {
				results, err := SearchRottenTomatoes("query")
				if err != nil {
					return err
				}
				if len(results) != 1 || results[0].Title != "Fake Result" {
					return errors.New("unexpected search results")
				}
				return nil
			},
		},
		{
			name: "SearchRottenTomatoes error from FetchHTML",
			setup: func() {
				FetchHTML = func(url string) (string, error) {
					return "", errors.New("boom")
				}
			},
			call: func() error {
				_, err := SearchRottenTomatoes("query")
				return err
			},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			FetchHTML = origFetch
			ExtractMovieMetadata = origMovie
			ExtractTVSeriesMetadata = origTV
			ExtractTVSeasonMetadata = origSeason
			ParseSearch = origSearch

			tc.setup()
			err := tc.call()
			if tc.expectErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.expectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
