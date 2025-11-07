package services

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stormlightlabs/noteleaf/internal/models"
	"github.com/stormlightlabs/noteleaf/internal/public"
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

// MockATProtoService is a mock implementation of ATProtoService for testing
type MockATProtoService struct {
	AuthenticateFunc          func(ctx context.Context, handle, password string) error
	GetSessionFunc            func() (*Session, error)
	IsAuthenticatedVal        bool
	RestoreSessionFunc        func(session *Session) error
	PullDocumentsFunc         func(ctx context.Context) ([]DocumentWithMeta, error)
	PostDocumentFunc          func(ctx context.Context, doc public.Document, isDraft bool) (*DocumentWithMeta, error)
	PatchDocumentFunc         func(ctx context.Context, rkey string, doc public.Document, isDraft bool) (*DocumentWithMeta, error)
	DeleteDocumentFunc        func(ctx context.Context, rkey string, isDraft bool) error
	UploadBlobFunc            func(ctx context.Context, data []byte, mimeType string) (public.Blob, error)
	GetDefaultPublicationFunc func(ctx context.Context) (string, error)
	CloseFunc                 func() error
	Session                   *Session // Exported for test access
}

// NewMockATProtoService creates a new mock AT Proto service
func NewMockATProtoService() *MockATProtoService {
	return &MockATProtoService{IsAuthenticatedVal: false}
}

// Authenticate mocks authentication
func (m *MockATProtoService) Authenticate(ctx context.Context, handle, password string) error {
	if m.AuthenticateFunc != nil {
		return m.AuthenticateFunc(ctx, handle, password)
	}

	// Default successful authentication
	m.Session = &Session{
		DID:           "did:plc:test123",
		Handle:        handle,
		AccessJWT:     "mock_access_token",
		RefreshJWT:    "mock_refresh_token",
		PDSURL:        "https://bsky.social",
		ExpiresAt:     time.Now().Add(2 * time.Hour),
		Authenticated: true,
	}
	m.IsAuthenticatedVal = true
	return nil
}

// GetSession returns the current session
func (m *MockATProtoService) GetSession() (*Session, error) {
	if m.GetSessionFunc != nil {
		return m.GetSessionFunc()
	}

	if m.Session == nil || !m.Session.Authenticated {
		return nil, errors.New("not authenticated - run 'noteleaf pub auth' first")
	}
	return m.Session, nil
}

// IsAuthenticated returns authentication status
func (m *MockATProtoService) IsAuthenticated() bool {
	return m.IsAuthenticatedVal
}

// RestoreSession restores a session
func (m *MockATProtoService) RestoreSession(session *Session) error {
	if m.RestoreSessionFunc != nil {
		return m.RestoreSessionFunc(session)
	}

	m.Session = session
	m.IsAuthenticatedVal = true
	return nil
}

// PullDocuments mocks pulling documents
func (m *MockATProtoService) PullDocuments(ctx context.Context) ([]DocumentWithMeta, error) {
	if m.PullDocumentsFunc != nil {
		return m.PullDocumentsFunc(ctx)
	}
	return []DocumentWithMeta{}, nil
}

// PostDocument mocks posting a document
func (m *MockATProtoService) PostDocument(ctx context.Context, doc public.Document, isDraft bool) (*DocumentWithMeta, error) {
	if m.PostDocumentFunc != nil {
		return m.PostDocumentFunc(ctx, doc, isDraft)
	}

	// Default successful post
	return &DocumentWithMeta{
		Document: doc,
		Meta: public.DocumentMeta{
			RKey:      "mock_rkey_123",
			CID:       "mock_cid_456",
			URI:       "at://did:plc:test123/pub.leaflet.document/mock_rkey_123",
			IsDraft:   isDraft,
			FetchedAt: time.Now(),
		},
	}, nil
}

// PatchDocument mocks patching a document
func (m *MockATProtoService) PatchDocument(ctx context.Context, rkey string, doc public.Document, isDraft bool) (*DocumentWithMeta, error) {
	if m.PatchDocumentFunc != nil {
		return m.PatchDocumentFunc(ctx, rkey, doc, isDraft)
	}

	return &DocumentWithMeta{
		Document: doc,
		Meta: public.DocumentMeta{
			RKey:      rkey,
			CID:       "mock_cid_updated_789",
			URI:       "at://did:plc:test123/pub.leaflet.document/" + rkey,
			IsDraft:   isDraft,
			FetchedAt: time.Now(),
		},
	}, nil
}

// DeleteDocument mocks deleting a document
func (m *MockATProtoService) DeleteDocument(ctx context.Context, rkey string, isDraft bool) error {
	if m.DeleteDocumentFunc != nil {
		return m.DeleteDocumentFunc(ctx, rkey, isDraft)
	}
	return nil
}

// UploadBlob mocks blob upload
func (m *MockATProtoService) UploadBlob(ctx context.Context, data []byte, mimeType string) (public.Blob, error) {
	if m.UploadBlobFunc != nil {
		return m.UploadBlobFunc(ctx, data, mimeType)
	}

	return public.Blob{
		Type:     public.TypeBlob,
		Ref:      public.CID{Link: "mock_blob_cid"},
		MimeType: mimeType,
		Size:     len(data),
	}, nil
}

// GetDefaultPublication mocks getting the default publication
func (m *MockATProtoService) GetDefaultPublication(ctx context.Context) (string, error) {
	if m.GetDefaultPublicationFunc != nil {
		return m.GetDefaultPublicationFunc(ctx)
	}

	// Default returns a mock publication URI
	if !m.IsAuthenticatedVal {
		return "", errors.New("not authenticated")
	}
	return "at://did:plc:test123/pub.leaflet.publication/mock_pub_rkey", nil
}

// Close mocks cleanup
func (m *MockATProtoService) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	m.Session = nil
	m.IsAuthenticatedVal = false
	return nil
}

// SetupSuccessfulAuthMocks configures mock for successful authentication
func SetupSuccessfulAuthMocks() *MockATProtoService {
	mock := NewMockATProtoService()
	mock.AuthenticateFunc = func(ctx context.Context, handle, password string) error {
		mock.Session = &Session{
			DID:           "did:plc:test123",
			Handle:        handle,
			AccessJWT:     "mock_access_token",
			RefreshJWT:    "mock_refresh_token",
			PDSURL:        "https://bsky.social",
			ExpiresAt:     time.Now().Add(2 * time.Hour),
			Authenticated: true,
		}
		mock.IsAuthenticatedVal = true
		return nil
	}
	return mock
}

// SetupSuccessfulPullMocks configures mock for successful document pull
func SetupSuccessfulPullMocks() *MockATProtoService {
	mock := NewMockATProtoService()
	mock.IsAuthenticatedVal = true
	mock.Session = &Session{
		DID:           "did:plc:test123",
		Handle:        "test.bsky.social",
		AccessJWT:     "mock_access",
		RefreshJWT:    "mock_refresh",
		PDSURL:        "https://bsky.social",
		ExpiresAt:     time.Now().Add(2 * time.Hour),
		Authenticated: true,
	}

	mock.PullDocumentsFunc = func(ctx context.Context) ([]DocumentWithMeta, error) {
		return []DocumentWithMeta{
			{
				Document: public.Document{
					Type:  public.TypeDocument,
					Title: "Test Document",
					Pages: []public.LinearDocument{
						{
							Type: public.TypeLinearDocument,
							Blocks: []public.BlockWrap{
								{
									Type: "pub.leaflet.pages.linearDocument#block",
									Block: public.TextBlock{
										Type:      "pub.leaflet.pages.linearDocument#textBlock",
										Plaintext: "Test content",
									},
								},
							},
						},
					},
					PublishedAt: time.Now().Format(time.RFC3339),
				},
				Meta: public.DocumentMeta{
					RKey:      "test_rkey",
					CID:       "test_cid",
					URI:       "at://did:plc:test123/pub.leaflet.document/test_rkey",
					IsDraft:   false,
					FetchedAt: time.Now(),
				},
			},
		}, nil
	}

	return mock
}
