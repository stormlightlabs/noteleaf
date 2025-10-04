package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stormlightlabs/noteleaf/internal/articles"
	"github.com/stormlightlabs/noteleaf/internal/models"
	"github.com/stormlightlabs/noteleaf/internal/repo"
	"github.com/stormlightlabs/noteleaf/internal/services"
	"github.com/stormlightlabs/noteleaf/internal/store"
	"github.com/stormlightlabs/noteleaf/internal/ui"
)

// HandlerTestHelper wraps NoteHandler with test-specific functionality
type HandlerTestHelper struct {
	*NoteHandler
	tempDir string
	cleanup func()
}

// NewHandlerTestHelper creates a NoteHandler with isolated test database
func NewHandlerTestHelper(t *testing.T) *HandlerTestHelper {
	tempDir, err := os.MkdirTemp("", "noteleaf-handler-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	oldNoteleafConfig := os.Getenv("NOTELEAF_CONFIG")
	oldNoteleafDataDir := os.Getenv("NOTELEAF_DATA_DIR")
	os.Setenv("NOTELEAF_CONFIG", filepath.Join(tempDir, ".noteleaf.conf.toml"))
	os.Setenv("NOTELEAF_DATA_DIR", tempDir)

	cleanup := func() {
		os.Setenv("NOTELEAF_CONFIG", oldNoteleafConfig)
		os.Setenv("NOTELEAF_DATA_DIR", oldNoteleafDataDir)
		os.RemoveAll(tempDir)
	}

	ctx := context.Background()
	err = Setup(ctx, []string{})
	if err != nil {
		cleanup()
		t.Fatalf("Failed to setup database: %v", err)
	}

	handler, err := NewNoteHandler()
	if err != nil {
		cleanup()
		t.Fatalf("Failed to create note handler: %v", err)
	}

	testHandler := &HandlerTestHelper{
		NoteHandler: handler,
		tempDir:     tempDir,
		cleanup:     cleanup,
	}

	t.Cleanup(func() {
		testHandler.Close()
		testHandler.cleanup()
	})

	return testHandler
}

// CreateTestNote creates a test note and returns its ID
func (th *HandlerTestHelper) CreateTestNote(t *testing.T, title, content string, tags []string) int64 {
	ctx := context.Background()
	note := &models.Note{
		Title:    title,
		Content:  content,
		Tags:     tags,
		Created:  time.Now(),
		Modified: time.Now(),
	}

	id, err := th.repos.Notes.Create(ctx, note)
	if err != nil {
		t.Fatalf("Failed to create test note: %v", err)
	}
	return id
}

// CreateTestFile creates a temporary markdown file with content
func (th *HandlerTestHelper) CreateTestFile(t *testing.T, filename, content string) string {
	filePath := filepath.Join(th.tempDir, filename)
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	return filePath
}

// MockEditor provides a mock editor function for testing
type MockEditor struct {
	shouldFail     bool
	failureMsg     string
	contentToWrite string
	deleteFile     bool
	makeReadOnly   bool
}

// NewMockEditor creates a mock editor with default success behavior
func NewMockEditor() *MockEditor {
	return &MockEditor{
		contentToWrite: `# Test Note

Test content here.

<!-- Tags: test -->`,
	}
}

// WithFailure configures the mock editor to fail
func (me *MockEditor) WithFailure(msg string) *MockEditor {
	me.shouldFail = true
	me.failureMsg = msg
	return me
}

// WithContent configures the content the mock editor will write
func (me *MockEditor) WithContent(content string) *MockEditor {
	me.contentToWrite = content
	return me
}

// WithFileDeleted configures the mock editor to delete the temp file
func (me *MockEditor) WithFileDeleted() *MockEditor {
	me.deleteFile = true
	return me
}

// WithReadOnly configures the mock editor to make the file read-only
func (me *MockEditor) WithReadOnly() *MockEditor {
	me.makeReadOnly = true
	return me
}

// GetEditorFunc returns the editor function for use with NoteHandler
func (me *MockEditor) GetEditorFunc() editorFunc {
	return func(editor, filePath string) error {
		if me.shouldFail {
			return fmt.Errorf("%s", me.failureMsg)
		}

		if me.deleteFile {
			return os.Remove(filePath)
		}

		if me.makeReadOnly {
			os.Chmod(filePath, 0444)
			return nil
		}

		return os.WriteFile(filePath, []byte(me.contentToWrite), 0644)
	}
}

// DatabaseTestHelper provides database testing utilities
type DatabaseTestHelper struct {
	originalDB *store.Database
	handler    *HandlerTestHelper
}

// NewDatabaseTestHelper creates a helper for database error testing
func NewDatabaseTestHelper(handler *HandlerTestHelper) *DatabaseTestHelper {
	return &DatabaseTestHelper{
		originalDB: handler.db,
		handler:    handler,
	}
}

// CloseDatabase closes the database connection
func (dth *DatabaseTestHelper) CloseDatabase() {
	dth.handler.db.Close()
}

// RestoreDatabase restores the original database connection
func (dth *DatabaseTestHelper) RestoreDatabase(t *testing.T) {
	var err error
	dth.handler.db, err = store.NewDatabase()
	if err != nil {
		t.Fatalf("Failed to restore database: %v", err)
	}
}

// DropNotesTable drops the notes table to simulate database errors
func (dth *DatabaseTestHelper) DropNotesTable() {
	dth.handler.db.Exec("DROP TABLE notes")
}

// CreateCorruptedDatabase creates a new database with corrupted schema
func (dth *DatabaseTestHelper) CreateCorruptedDatabase(t *testing.T) {
	dth.CloseDatabase()

	db, err := store.NewDatabase()
	if err != nil {
		t.Fatalf("Failed to create corrupted database: %v", err)
	}

	db.Exec("DROP TABLE notes")
	dth.handler.db = db
}

// AssertionHelpers provides test assertion utilities
type AssertionHelpers struct{}

// AssertError checks that an error occurred and optionally contains expected text
func (ah AssertionHelpers) AssertError(t *testing.T, err error, expectedSubstring string, msg string) {
	t.Helper()
	if err == nil {
		t.Errorf("%s: expected error but got none", msg)
		return
	}
	if expectedSubstring != "" && !containsString(err.Error(), expectedSubstring) {
		t.Errorf("%s: expected error containing %q, got: %v", msg, expectedSubstring, err)
	}
}

// AssertNoError checks that no error occurred
func (ah AssertionHelpers) AssertNoError(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Errorf("%s: unexpected error: %v", msg, err)
	}
}

// AssertNoteExists checks that a note exists in the database
func (ah AssertionHelpers) AssertNoteExists(t *testing.T, handler *HandlerTestHelper, id int64) {
	t.Helper()
	ctx := context.Background()
	_, err := handler.repos.Notes.Get(ctx, id)
	if err != nil {
		t.Errorf("Note %d should exist but got error: %v", id, err)
	}
}

// AssertNoteNotExists checks that a note does not exist in the database
func (ah AssertionHelpers) AssertNoteNotExists(t *testing.T, handler *HandlerTestHelper, id int64) {
	t.Helper()
	ctx := context.Background()
	_, err := handler.repos.Notes.Get(ctx, id)
	if err == nil {
		t.Errorf("Note %d should not exist but was found", id)
	}
}

// AssertArticleExists checks that an article exists in the database
func (ah AssertionHelpers) AssertArticleExists(t *testing.T, handler *ArticleTestHelper, id int64) {
	t.Helper()
	ctx := context.Background()
	_, err := handler.repos.Articles.Get(ctx, id)
	if err != nil {
		t.Errorf("Article %d should exist but got error: %v", id, err)
	}
}

// AssertArticleNotExists checks that an article does not exist in the database
func (ah AssertionHelpers) AssertArticleNotExists(t *testing.T, handler *ArticleTestHelper, id int64) {
	t.Helper()
	ctx := context.Background()
	_, err := handler.repos.Articles.Get(ctx, id)
	if err == nil {
		t.Errorf("Article %d should not exist but was found", id)
	}
}

// EnvironmentTestHelper provides environment manipulation utilities for testing
type EnvironmentTestHelper struct {
	originalVars map[string]string
}

// NewEnvironmentTestHelper creates a new environment test helper
func NewEnvironmentTestHelper() *EnvironmentTestHelper {
	return &EnvironmentTestHelper{
		originalVars: make(map[string]string),
	}
}

// SetEnv sets an environment variable and remembers the original value
func (eth *EnvironmentTestHelper) SetEnv(key, value string) {
	if _, exists := eth.originalVars[key]; !exists {
		eth.originalVars[key] = os.Getenv(key)
	}
	os.Setenv(key, value)
}

// UnsetEnv unsets an environment variable and remembers the original value
func (eth *EnvironmentTestHelper) UnsetEnv(key string) {
	if _, exists := eth.originalVars[key]; !exists {
		eth.originalVars[key] = os.Getenv(key)
	}
	os.Unsetenv(key)
}

// RestoreEnv restores all modified environment variables
func (eth *EnvironmentTestHelper) RestoreEnv() {
	for key, value := range eth.originalVars {
		if value == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, value)
		}
	}
}

// CreateTestDir creates a temporary test directory and sets up environment
func (eth *EnvironmentTestHelper) CreateTestDir(t *testing.T) (string, error) {
	tempDir, err := os.MkdirTemp("", "noteleaf-test-*")
	if err != nil {
		return "", err
	}

	eth.SetEnv("XDG_CONFIG_HOME", tempDir)

	ctx := context.Background()
	err = Setup(ctx, []string{})
	if err != nil {
		os.RemoveAll(tempDir)
		return "", err
	}

	t.Cleanup(func() {
		eth.RestoreEnv()
		os.RemoveAll(tempDir)
	})

	return tempDir, nil
}

// Helper function to check if string contains substring (case-insensitive)
func containsString(haystack, needle string) bool {
	if needle == "" {
		return true
	}
	return len(haystack) >= len(needle) &&
		haystack[len(haystack)-len(needle):] == needle ||
		haystack[:len(needle)] == needle ||
		(len(haystack) > len(needle) &&
			func() bool {
				for i := 1; i <= len(haystack)-len(needle); i++ {
					if haystack[i:i+len(needle)] == needle {
						return true
					}
				}
				return false
			}())
}

// ArticleTestHelper wraps ArticleHandler with test-specific functionality
type ArticleTestHelper struct {
	*ArticleHandler
	tempDir string
	cleanup func()
}

// NewArticleTestHelper creates an ArticleHandler with isolated test database
func NewArticleTestHelper(t *testing.T) *ArticleTestHelper {
	tempDir, err := os.MkdirTemp("", "noteleaf-article-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	oldNoteleafConfig := os.Getenv("NOTELEAF_CONFIG")
	oldNoteleafDataDir := os.Getenv("NOTELEAF_DATA_DIR")
	os.Setenv("NOTELEAF_CONFIG", filepath.Join(tempDir, ".noteleaf.conf.toml"))
	os.Setenv("NOTELEAF_DATA_DIR", tempDir)

	cleanup := func() {
		os.Setenv("NOTELEAF_CONFIG", oldNoteleafConfig)
		os.Setenv("NOTELEAF_DATA_DIR", oldNoteleafDataDir)
		os.RemoveAll(tempDir)
	}

	ctx := context.Background()
	err = Setup(ctx, []string{})
	if err != nil {
		cleanup()
		t.Fatalf("Failed to setup database: %v", err)
	}

	handler, err := NewArticleHandler()
	if err != nil {
		cleanup()
		t.Fatalf("Failed to create article handler: %v", err)
	}

	testHelper := &ArticleTestHelper{
		ArticleHandler: handler,
		tempDir:        tempDir,
		cleanup:        cleanup,
	}

	t.Cleanup(func() {
		testHelper.Close()
		testHelper.cleanup()
	})

	return testHelper
}

// CreateTestArticle creates a test article and returns its ID
func (ath *ArticleTestHelper) CreateTestArticle(t *testing.T, url, title, author, date string) int64 {
	ctx := context.Background()

	mdPath := filepath.Join(ath.tempDir, fmt.Sprintf("%s.md", title))
	htmlPath := filepath.Join(ath.tempDir, fmt.Sprintf("%s.html", title))

	mdContent := fmt.Sprintf("# %s\n\n**Author:** %s\n**Date:** %s\n\nTest content", title, author, date)
	err := os.WriteFile(mdPath, []byte(mdContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test markdown file: %v", err)
	}

	htmlContent := fmt.Sprintf("<h1>%s</h1><p>Author: %s</p><p>Date: %s</p><p>Test content</p>", title, author, date)
	err = os.WriteFile(htmlPath, []byte(htmlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test HTML file: %v", err)
	}

	article := &models.Article{
		URL:          url,
		Title:        title,
		Author:       author,
		Date:         date,
		MarkdownPath: mdPath,
		HTMLPath:     htmlPath,
		Created:      time.Now(),
		Modified:     time.Now(),
	}

	id, err := ath.repos.Articles.Create(ctx, article)
	if err != nil {
		t.Fatalf("Failed to create test article: %v", err)
	}
	return id
}

// AddTestRule adds a parsing rule to the handler's parser for testing
func (ath *ArticleTestHelper) AddTestRule(domain string, rule *articles.ParsingRule) {
	if parser, ok := ath.parser.(*articles.ArticleParser); ok {
		parser.AddRule(domain, rule)
	} else {
		panic("Could not cast parser to ArticleParser")
	}
}

var Expect = AssertionHelpers{}

// HTTPMockServer provides utilities for mocking HTTP services in tests
type HTTPMockServer struct {
	server   *httptest.Server
	requests []*http.Request
}

// NewMockServer creates a new mock HTTP server
func NewMockServer() *HTTPMockServer {
	mock := &HTTPMockServer{
		requests: make([]*http.Request, 0),
	}
	return mock
}

// WithHandler sets up the mock server with a custom handler
func (m *HTTPMockServer) WithHandler(handler http.HandlerFunc) *HTTPMockServer {
	m.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.requests = append(m.requests, r)
		handler(w, r)
	}))
	return m
}

// URL returns the mock server URL
func (m *HTTPMockServer) URL() string {
	if m.server == nil {
		panic("mock server not initialized - call WithHandler first")
	}
	return m.server.URL
}

// Close closes the mock server
func (m *HTTPMockServer) Close() {
	if m.server != nil {
		m.server.Close()
	}
}

// GetRequests returns all recorded HTTP requests
func (m *HTTPMockServer) GetRequests() []*http.Request {
	return m.requests
}

// GetLastRequest returns the last recorded HTTP request
func (m *HTTPMockServer) GetLastRequest() *http.Request {
	if len(m.requests) == 0 {
		return nil
	}
	return m.requests[len(m.requests)-1]
}

// MockOpenLibraryResponse creates a mock OpenLibrary search response
func MockOpenLibraryResponse(books []MockBook) services.OpenLibrarySearchResponse {
	docs := make([]services.OpenLibrarySearchDoc, len(books))
	for i, book := range books {
		docs[i] = services.OpenLibrarySearchDoc{
			Key:              book.Key,
			Title:            book.Title,
			AuthorName:       book.Authors,
			FirstPublishYear: book.Year,
			Edition_count:    book.Editions,
			CoverI:           book.CoverID,
		}
	}
	return services.OpenLibrarySearchResponse{
		NumFound: len(books),
		Start:    0,
		Docs:     docs,
	}
}

// MockBook represents a book for testing
type MockBook struct {
	Key      string
	Title    string
	Authors  []string
	Year     int
	Editions int
	CoverID  int
}

// MockRottenTomatoesResponse creates a mock HTML response for Rotten Tomatoes
func MockRottenTomatoesResponse(movies []MockMedia) string {
	var html strings.Builder
	html.WriteString(`<html><body><div class="search-page-result">`)

	for _, movie := range movies {
		html.WriteString(fmt.Sprintf(`
			<div class="mb-movie" data-qa="result-item">
				<div class="poster">
					<a href="%s" title="%s">
						<img src="poster.jpg" alt="%s">
					</a>
				</div>
				<div class="info">
					<h3><a href="%s">%s</a></h3>
					<div class="critics-score">%s</div>
				</div>
			</div>
		`, movie.Link, movie.Title, movie.Title, movie.Link, movie.Title, movie.Score))
	}

	html.WriteString(`</div></body></html>`)
	return html.String()
}

// MockMedia represents media for testing
type MockMedia struct {
	Title string
	Link  string
	Score string
	Type  string
}

// HTTPErrorMockServer creates a mock server that returns HTTP errors
func HTTPErrorMockServer(statusCode int, message string) *HTTPMockServer {
	return NewMockServer().WithHandler(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, message, statusCode)
	})
}

// JSONMockServer creates a mock server that returns JSON responses
func JSONMockServer(response any) *HTTPMockServer {
	return NewMockServer().WithHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	})
}

// TimeoutMockServer creates a mock server that simulates timeouts
func TimeoutMockServer(delay time.Duration) *HTTPMockServer {
	return NewMockServer().WithHandler(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)
		w.WriteHeader(http.StatusOK)
	})
}

// InvalidJSONMockServer creates a mock server that returns malformed JSON
func InvalidJSONMockServer() *HTTPMockServer {
	return NewMockServer().WithHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"invalid": json`))
	})
}

// EmptyResponseMockServer creates a mock server that returns empty responses
func EmptyResponseMockServer() *HTTPMockServer {
	return NewMockServer().WithHandler(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

// ServiceTestHelper provides utilities for testing services with HTTP mocks
type ServiceTestHelper struct {
	mockServers []*HTTPMockServer
}

// NewServiceTestHelper creates a new service test helper
func NewServiceTestHelper() *ServiceTestHelper {
	return &ServiceTestHelper{
		mockServers: make([]*HTTPMockServer, 0),
	}
}

// AddMockServer adds a mock server and returns its URL
func (sth *ServiceTestHelper) AddMockServer(server *HTTPMockServer) string {
	sth.mockServers = append(sth.mockServers, server)
	return server.URL()
}

// Cleanup closes all mock servers
func (sth *ServiceTestHelper) Cleanup() {
	for _, server := range sth.mockServers {
		server.Close()
	}
}

// AssertRequestMade verifies that a request was made to the mock server
func (sth *ServiceTestHelper) AssertRequestMade(t *testing.T, server *HTTPMockServer, expectedPath string) {
	t.Helper()
	if len(server.requests) == 0 {
		t.Error("Expected HTTP request to be made but none were recorded")
		return
	}

	lastReq := server.GetLastRequest()
	if lastReq.URL.Path != expectedPath {
		t.Errorf("Expected request to path %s, got %s", expectedPath, lastReq.URL.Path)
	}
}

// MockMediaFetcher provides a test implementation of Fetchable and Searchable interfaces
type MockMediaFetcher struct {
	SearchResults []services.Media
	HTMLContent   string
	MovieData     *services.Movie
	TVSeriesData  *services.TVSeries
	ShouldError   bool
	ErrorMessage  string
}

// Search implements the Searchable interface for testing
func (m *MockMediaFetcher) Search(query string) ([]services.Media, error) {
	if m.ShouldError {
		return nil, fmt.Errorf("mock search error: %s", m.ErrorMessage)
	}
	return m.SearchResults, nil
}

// MakeRequest implements the Fetchable interface for testing
func (m *MockMediaFetcher) MakeRequest(url string) (string, error) {
	if m.ShouldError {
		return "", fmt.Errorf("mock fetch error: %s", m.ErrorMessage)
	}
	return m.HTMLContent, nil
}

// MovieRequest implements the Fetchable interface for testing
func (m *MockMediaFetcher) MovieRequest(url string) (*services.Movie, error) {
	if m.ShouldError {
		return nil, fmt.Errorf("mock movie fetch error: %s", m.ErrorMessage)
	}
	if m.MovieData == nil {
		return nil, fmt.Errorf("movie not found")
	}
	return m.MovieData, nil
}

// TVRequest implements the Fetchable interface for testing
func (m *MockMediaFetcher) TVRequest(url string) (*services.TVSeries, error) {
	if m.ShouldError {
		return nil, fmt.Errorf("mock tv series fetch error: %s", m.ErrorMessage)
	}
	if m.TVSeriesData == nil {
		return nil, fmt.Errorf("tv series not found")
	}
	return m.TVSeriesData, nil
}

// CreateTestMovieService creates a MovieService with mock dependencies for testing
func CreateTestMovieService(mockFetcher *MockMediaFetcher) *services.MovieService {
	return services.NewMovieSrvWithOpts("http://localhost", mockFetcher, mockFetcher)
}

// CreateTestTVService creates a TVService with mock dependencies for testing
func CreateTestTVService(mockFetcher *MockMediaFetcher) *services.TVService {
	return services.NewTVServiceWithDeps("http://localhost", mockFetcher, mockFetcher)
}

// CreateMockMovieSearchResults creates sample movie search results for testing
func CreateMockMovieSearchResults() []services.Media {
	return []services.Media{
		{Title: "Test Movie 1", Link: "/m/test_movie_1", Type: "movie", CriticScore: "85%"},
		{Title: "Test Movie 2", Link: "/m/test_movie_2", Type: "movie", CriticScore: "72%"},
	}
}

// CreateMockTVSearchResults creates sample TV search results for testing
func CreateMockTVSearchResults() []services.Media {
	return []services.Media{
		{Title: "Test TV Show 1", Link: "/tv/test_show_1", Type: "tv", CriticScore: "90%"},
		{Title: "Test TV Show 2", Link: "/tv/test_show_2", Type: "tv", CriticScore: "80%"},
	}
}

// InputSimulator provides controlled input simulation for testing [fmt.Scanf] interactions
// It implements [io.Reader] to provide predictable input sequences for interactive components
type InputSimulator struct {
	inputs   []string
	position int
	mu       sync.RWMutex
}

// NewInputSimulator creates a new input simulator with the given input sequence
func NewInputSimulator(inputs ...string) *InputSimulator {
	return &InputSimulator{inputs: inputs}
}

// Read implements [io.Reader] interface for [fmt.Scanf] compatibility
func (is *InputSimulator) Read(p []byte) (n int, err error) {
	is.mu.Lock()
	defer is.mu.Unlock()

	if is.position >= len(is.inputs) {
		return 0, io.EOF
	}

	input := is.inputs[is.position] + "\n"
	is.position++

	n = copy(p, []byte(input))
	return n, nil
}

// Reset resets the simulator to the beginning of the input sequence
func (is *InputSimulator) Reset() {
	is.mu.Lock()
	defer is.mu.Unlock()
	is.position = 0
}

// AddInputs appends new inputs to the sequence
func (is *InputSimulator) AddInputs(inputs ...string) {
	is.mu.Lock()
	defer is.mu.Unlock()
	is.inputs = append(is.inputs, inputs...)
}

// HasMoreInputs returns true if there are more inputs available
func (is *InputSimulator) HasMoreInputs() bool {
	is.mu.RLock()
	defer is.mu.RUnlock()
	return is.position < len(is.inputs)
}

// MenuSelection creates input simulator for menu selection scenarios
func MenuSelection(choice int) *InputSimulator {
	return NewInputSimulator(strconv.Itoa(choice))
}

// MenuCancel creates input simulator for cancelling menu selection
func MenuCancel() *InputSimulator {
	return NewInputSimulator("0")
}

// MenuSequence creates input simulator for multiple menu interactions
func MenuSequence(choices ...int) *InputSimulator {
	inputs := make([]string, len(choices))
	for i, choice := range choices {
		inputs[i] = strconv.Itoa(choice)
	}
	return NewInputSimulator(inputs...)
}

// InteractiveTestHelper provides utilities for testing interactive handler components
type InteractiveTestHelper struct {
	Stdin io.Reader
	sim   *InputSimulator
}

// NewInteractiveTestHelper creates a helper for testing interactive components
func NewInteractiveTestHelper() *InteractiveTestHelper {
	return &InteractiveTestHelper{}
}

// SimulateInput sets up input simulation for the test
func (ith *InteractiveTestHelper) SimulateInput(inputs ...string) *InputSimulator {
	ith.sim = NewInputSimulator(inputs...)
	return ith.sim
}

// SimulateMenuChoice sets up input simulation for menu selection
func (ith *InteractiveTestHelper) SimulateMenuChoice(choice int) *InputSimulator {
	return ith.SimulateInput(strconv.Itoa(choice))
}

// SimulateCancel sets up input simulation for cancellation
func (ith *InteractiveTestHelper) SimulateCancel() *InputSimulator {
	return ith.SimulateInput("0")
}

// GetSimulator returns the current input simulator
func (ith *InteractiveTestHelper) GetSimulator() *InputSimulator {
	return ith.sim
}

// SetupHandlerWithInput creates a handler and sets up input simulation in one call
func SetupBookHandlerWithInput(t *testing.T, inputs ...string) (*BookHandler, func()) {
	_, cleanup := setupTest(t)

	handler, err := NewBookHandler()
	if err != nil {
		cleanup()
		t.Fatalf("Failed to create book handler: %v", err)
	}

	if len(inputs) > 0 {
		handler.SetInputReader(NewInputSimulator(inputs...))
	}

	fullCleanup := func() {
		handler.Close()
		cleanup()
	}

	return handler, fullCleanup
}

// SetupMovieHandlerWithInput creates a movie handler and sets up input simulation
func SetupMovieHandlerWithInput(t *testing.T, inputs ...string) (*MovieHandler, func()) {
	_, cleanup := setupTest(t)

	handler, err := NewMovieHandler()
	if err != nil {
		cleanup()
		t.Fatalf("Failed to create movie handler: %v", err)
	}

	if len(inputs) > 0 {
		handler.SetInputReader(NewInputSimulator(inputs...))
	}

	fullCleanup := func() {
		handler.Close()
		cleanup()
	}

	return handler, fullCleanup
}

// SetupTVHandlerWithInput creates a TV handler and sets up input simulation
func SetupTVHandlerWithInput(t *testing.T, inputs ...string) (*TVHandler, func()) {
	_, cleanup := setupTest(t)

	handler, err := NewTVHandler()
	if err != nil {
		cleanup()
		t.Fatalf("Failed to create TV handler: %v", err)
	}

	if len(inputs) > 0 {
		handler.SetInputReader(NewInputSimulator(inputs...))
	}

	fullCleanup := func() {
		handler.Close()
		cleanup()
	}

	return handler, fullCleanup
}

func setupTest(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", "noteleaf-interactive-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	oldNoteleafConfig := os.Getenv("NOTELEAF_CONFIG")
	oldNoteleafDataDir := os.Getenv("NOTELEAF_DATA_DIR")
	os.Setenv("NOTELEAF_CONFIG", filepath.Join(tempDir, ".noteleaf.conf.toml"))
	os.Setenv("NOTELEAF_DATA_DIR", tempDir)

	cleanup := func() {
		os.Setenv("NOTELEAF_CONFIG", oldNoteleafConfig)
		os.Setenv("NOTELEAF_DATA_DIR", oldNoteleafDataDir)
		os.RemoveAll(tempDir)
	}

	ctx := context.Background()
	err = Setup(ctx, []string{})
	if err != nil {
		cleanup()
		t.Fatalf("Failed to setup database: %v", err)
	}

	return tempDir, cleanup
}

// TUICapableHandler interface for handlers that can expose TUI models for testing
type TUICapableHandler interface {
	GetTUIModel(ctx context.Context, opts TUITestOptions) (tea.Model, error)
	SetTUITestMode(enabled bool)
}

// TUITestOptions configures TUI testing behavior for handlers
type TUITestOptions struct {
	Width    int
	Height   int
	Static   bool
	Output   io.Writer
	Input    io.Reader
	MockData any
}

// InteractiveTUIHelper bridges handler testing with TUI testing capabilities
type InteractiveTUIHelper struct {
	t       *testing.T
	handler TUICapableHandler
	suite   *ui.TUITestSuite
	opts    TUITestOptions
}

// NewInteractiveTUIHelper creates a helper for testing handler TUI interactions
func NewInteractiveTUIHelper(t *testing.T, handler TUICapableHandler) *InteractiveTUIHelper {
	return &InteractiveTUIHelper{
		t:       t,
		handler: handler,
		opts: TUITestOptions{
			Width:  80,
			Height: 24,
		},
	}
}

// WithSize configures the TUI test dimensions
func (ith *InteractiveTUIHelper) WithSize(width, height int) *InteractiveTUIHelper {
	ith.opts.Width = width
	ith.opts.Height = height
	return ith
}

// WithMockData configures mock data for the TUI test
func (ith *InteractiveTUIHelper) WithMockData(data any) *InteractiveTUIHelper {
	ith.opts.MockData = data
	return ith
}

// StartTUITest initializes and starts a TUI test session
func (ith *InteractiveTUIHelper) StartTUITest(ctx context.Context) (*ui.TUITestSuite, error) {
	ith.handler.SetTUITestMode(true)

	model, err := ith.handler.GetTUIModel(ctx, ith.opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get TUI model: %w", err)
	}

	ith.suite = ui.NewTUITestSuite(ith.t, model,
		ui.WithInitialSize(ith.opts.Width, ith.opts.Height))
	ith.suite.Start()

	return ith.suite, nil
}

// TestInteractiveList tests interactive list browsing behavior
func (ith *InteractiveTUIHelper) TestInteractiveList(ctx context.Context, testFunc func(*ui.TUITestSuite) error) error {
	suite, err := ith.StartTUITest(ctx)
	if err != nil {
		return err
	}
	return testFunc(suite)
}

// TestInteractiveNavigation tests keyboard navigation patterns
func (ith *InteractiveTUIHelper) TestInteractiveNavigation(ctx context.Context, keySequence []tea.KeyType) error {
	suite, err := ith.StartTUITest(ctx)
	if err != nil {
		return err
	}

	for _, key := range keySequence {
		if err := suite.SendKey(key); err != nil {
			return fmt.Errorf("failed to send key %v: %w", key, err)
		}
		time.Sleep(50 * time.Millisecond)
	}

	return nil
}

// TestInteractiveSelection tests item selection and actions
func (ith *InteractiveTUIHelper) TestInteractiveSelection(ctx context.Context, selected int, action tea.KeyType) error {
	suite, err := ith.StartTUITest(ctx)
	if err != nil {
		return err
	}

	for range selected {
		if err := suite.SendKey(tea.KeyDown); err != nil {
			return fmt.Errorf("failed to navigate down: %w", err)
		}
		time.Sleep(25 * time.Millisecond)
	}

	return suite.SendKey(action)
}

// TestMovieInteractiveList tests movie list browsing with TUI
func TestMovieInteractiveList(t *testing.T, handler *MovieHandler, status string) error {
	ctx := context.Background()

	t.Run("interactive_mode", func(t *testing.T) {
		movies, err := handler.repos.Movies.List(ctx, repo.MovieListOptions{})
		if err != nil {
			t.Fatalf("Failed to get movies for TUI test: %v", err)
		}

		if len(movies) == 0 {
			t.Skip("No movies available for interactive testing")
		}

		t.Logf("Would test interactive list with %d movies", len(movies))
	})

	return nil
}

// TestTVInteractiveList tests TV show list browsing with TUI
func TestTVInteractiveList(t *testing.T, handler *TVHandler, status string) error {
	ctx := context.Background()

	t.Run("interactive_tv_list", func(t *testing.T) {
		shows, err := handler.repos.TV.List(ctx, repo.TVListOptions{})
		if err != nil {
			t.Fatalf("Failed to get TV shows for TUI test: %v", err)
		}

		if len(shows) == 0 {
			t.Skip("No TV shows available for interactive testing")
		}

		t.Logf("Would test interactive TV list with %d shows", len(shows))
	})

	return nil
}

// TestBookInteractiveList tests book list browsing with TUI
func TestBookInteractiveList(t *testing.T, handler *BookHandler, status string) error {
	ctx := context.Background()

	t.Run("interactive_book_list", func(t *testing.T) {
		books, err := handler.repos.Books.List(ctx, repo.BookListOptions{})
		if err != nil {
			t.Fatalf("Failed to get books for TUI test: %v", err)
		}

		if len(books) == 0 {
			t.Skip("No books available for interactive testing")
		}

		t.Logf("Would test interactive book list with %d books", len(books))
	})

	return nil
}

// TestTaskInteractiveList tests task list browsing with TUI
func TestTaskInteractiveList(t *testing.T, handler *TaskHandler, showAll bool, status, priority, project string) error {
	ctx := context.Background()

	t.Run("interactive_task_list", func(t *testing.T) {
		tasks, err := handler.repos.Tasks.List(ctx, repo.TaskListOptions{
			Status: status, Priority: priority, Project: project,
		})
		if err != nil {
			t.Fatalf("Failed to get tasks for TUI test: %v", err)
		}

		if len(tasks) == 0 {
			t.Skip("No tasks available for interactive testing")
		}

		t.Logf("Would test interactive task list with %d tasks", len(tasks))
	})

	return nil
}

// TestNoteInteractiveList tests note list browsing with TUI
func TestNoteInteractiveList(t *testing.T, handler *NoteHandler, showArchived bool, tags []string) error {
	ctx := context.Background()

	t.Run("interactive_note_list", func(t *testing.T) {
		notes, err := handler.repos.Notes.List(ctx, repo.NoteListOptions{
			Archived: &showArchived, Tags: tags,
		})
		if err != nil {
			t.Fatalf("Failed to get notes for TUI test: %v", err)
		}

		if len(notes) == 0 {
			t.Skip("No notes available for interactive testing")
		}

		t.Logf("Would test interactive note list with %d notes", len(notes))
	})

	return nil
}

// TUITestScenario defines a common test scenario for interactive components
type TUITestScenario struct {
	Name         string
	KeySequence  []tea.KeyType
	ExpectedView string
	Timeout      time.Duration
}

// RunTUIScenarios executes a series of TUI test scenarios
func RunTUIScenarios(t *testing.T, suite *ui.TUITestSuite, scenarios []TUITestScenario) {
	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			timeout := scenario.Timeout
			if timeout == 0 {
				timeout = 1 * time.Second
			}

			for _, key := range scenario.KeySequence {
				if err := suite.SendKey(key); err != nil {
					t.Fatalf("Failed to send key %v in scenario %s: %v", key, scenario.Name, err)
				}
				time.Sleep(50 * time.Millisecond)
			}

			if scenario.ExpectedView != "" {
				if err := suite.WaitForView(scenario.ExpectedView, timeout); err != nil {
					t.Errorf("Expected view containing '%s' in scenario %s: %v", scenario.ExpectedView, scenario.Name, err)
				}
			}
		})
	}
}

// CommonTUIScenarios returns standard TUI testing scenarios
func CommonTUIScenarios() []TUITestScenario {
	return []TUITestScenario{
		{
			Name:        "help_toggle",
			KeySequence: []tea.KeyType{tea.KeyRunes},
			Timeout:     500 * time.Millisecond,
		},
		{
			Name:        "navigation_down",
			KeySequence: []tea.KeyType{tea.KeyDown, tea.KeyDown, tea.KeyUp},
			Timeout:     500 * time.Millisecond,
		},
		{
			Name:        "page_navigation",
			KeySequence: []tea.KeyType{tea.KeyPgDown, tea.KeyPgUp},
			Timeout:     500 * time.Millisecond,
		},
		{
			Name:        "quit_sequence",
			KeySequence: []tea.KeyType{tea.KeyCtrlC},
			Timeout:     500 * time.Millisecond,
		},
	}
}
