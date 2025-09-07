package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
)

type Media struct {
	Title string
	Link  string
	// "movie" or "tv"
	Type           string
	CriticScore    string
	CertifiedFresh bool
}

// ParseSearch parses Rotten Tomatoes search results HTML into Media entries.
func ParseSearch(r io.Reader) ([]Media, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}

	var results []Media
	doc.Find("a.score-list-item").Each(func(i int, e *goquery.Selection) {
		title := e.Find("span").Text()
		link, _ := e.Attr("href")
		if link == "" {
			return
		}
		fullLink := "https://www.rottentomatoes.com" + link

		mediaType := "unknown"
		if strings.HasPrefix(link, "/m/") {
			mediaType = "movie"
		} else if strings.HasPrefix(link, "/tv/") {
			mediaType = "tv"
		}

		score := e.Find("rt-text.critics-score").Text()
		if score == "" {
			score = "--"
		}

		certified := false
		if v, ok := e.Find("score-icon-critics").Attr("certified"); ok && v == "true" {
			certified = true
		}

		results = append(results, Media{
			Title:          title,
			Link:           fullLink,
			Type:           mediaType,
			CriticScore:    score,
			CertifiedFresh: certified,
		})
	})
	return results, nil
}

// SearchRottenTomatoes fetches live search results for a query.
func SearchRottenTomatoes(q string) ([]Media, error) {
	searchURL := "https://www.rottentomatoes.com/search?search=" + url.QueryEscape(q)
	html, err := FetchHTML(searchURL)
	if err != nil {
		return nil, err
	}
	return ParseSearch(strings.NewReader(html))
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
	Context       string `json:"@context"`
	Type          string `json:"@type"`
	Name          string `json:"name"`
	URL           string `json:"url"`
	Description   string `json:"description"`
	Image         string `json:"image"`
	SeasonNumber  int    `json:"seasonNumber"`
	DatePublished string `json:"datePublished"`
	PartOfSeries  struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"partOfSeries"`
	AggregateRating AggregateRating `json:"aggregateRating"`
}

func ExtractTVSeriesMetadata(r io.Reader) (*TVSeries, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}
	var series TVSeries
	found := false
	doc.Find("script[type='application/ld+json']").Each(func(i int, s *goquery.Selection) {
		var tmp map[string]interface{}
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

func ExtractMovieMetadata(r io.Reader) (*Movie, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}
	var movie Movie
	found := false
	doc.Find("script[type='application/ld+json']").Each(func(i int, s *goquery.Selection) {
		var tmp map[string]interface{}
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

func ExtractTVSeasonMetadata(r io.Reader) (*TVSeason, error) {
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

func FetchHTML(url string) (string, error) {
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

func FetchTVSeries(url string) (*TVSeries, error) {
	html, err := FetchHTML(url)
	if err != nil {
		return nil, err
	}
	return ExtractTVSeriesMetadata(strings.NewReader(html))
}

func FetchMovie(url string) (*Movie, error) {
	html, err := FetchHTML(url)
	if err != nil {
		return nil, err
	}
	return ExtractMovieMetadata(strings.NewReader(html))
}

func FetchTVSeason(url string) (*TVSeason, error) {
	html, err := FetchHTML(url)
	if err != nil {
		return nil, err
	}
	return ExtractTVSeasonMetadata(strings.NewReader(html))
}
