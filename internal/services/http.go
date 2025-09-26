package services

import (
	"net/url"
	"strings"

	"github.com/gocolly/colly/v2"
)

type Searchable interface {
	Search(query string) ([]Media, error)
}

type Fetchable interface {
	MakeRequest(url string) (string, error)
	MovieRequest(url string) (*Movie, error)
	TVRequest(url string) (*TVSeries, error)
}

// DefaultFetcher provides the default implementation using colly
type DefaultFetcher struct{}

func (f *DefaultFetcher) MakeRequest(url string) (string, error) {
	return FetchHTML(url)
}

func (f *DefaultFetcher) MovieRequest(url string) (*Movie, error) {
	return FetchMovie(url)
}

func (f *DefaultFetcher) TVRequest(url string) (*TVSeries, error) {
	return FetchTVSeries(url)
}

func (f *DefaultFetcher) Search(query string) ([]Media, error) {
	return SearchRottenTomatoes(query)
}

// SearchRottenTomatoes fetches live search results for a query.
var SearchRottenTomatoes = func(q string) ([]Media, error) {
	searchURL := "https://www.rottentomatoes.com/search?search=" + url.QueryEscape(q)
	html, err := FetchHTML(searchURL)
	if err != nil {
		return nil, err
	}
	return ParseSearch(strings.NewReader(html))
}

var FetchHTML = func(url string) (string, error) {
	var html string
	c := colly.NewCollector(
		colly.AllowedDomains("www.rottentomatoes.com", "rottentomatoes.com"),
	)
	c.OnResponse(func(r *colly.Response) { html = string(r.Body) })
	if err := c.Visit(url); err != nil {
		return "", err
	}
	return html, nil
}

var FetchTVSeries = func(url string) (*TVSeries, error) {
	html, err := FetchHTML(url)
	if err != nil {
		return nil, err
	}
	return ExtractTVSeriesMetadata(strings.NewReader(html))
}

var FetchMovie = func(url string) (*Movie, error) {
	html, err := FetchHTML(url)
	if err != nil {
		return nil, err
	}
	return ExtractMovieMetadata(strings.NewReader(html))
}

var FetchTVSeason = func(url string) (*TVSeason, error) {
	html, err := FetchHTML(url)
	if err != nil {
		return nil, err
	}
	return ExtractTVSeasonMetadata(strings.NewReader(html))
}
