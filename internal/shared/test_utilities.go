// shared test utilities & helpers
package shared

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func NewHTTPTestServer(t testing.TB, handler http.Handler) *httptest.Server {
	t.Helper()
	server, err := startHTTPTestServer(handler)
	if err != nil {
		t.Fatalf("failed to start HTTP test server: %v", err)
	}
	return server
}

func startHTTPTestServer(handler http.Handler) (*httptest.Server, error) {
	ln, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	server := &httptest.Server{
		Listener: ln,
		Config:   &http.Server{Handler: handler},
	}
	server.Start()
	return server, nil
}

func CreateTempDir(p string, t *testing.T) (string, func()) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", p)
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	return tempDir, func() { os.RemoveAll(tempDir) }
}

func AssertNoError(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", msg, err)
	}
}

func AssertError(t *testing.T, err error, msg string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: expected error but got none", msg)
	}
}

// AssertErrorContains checks that an error occurred and optionally contains expected text
func AssertErrorContains(t *testing.T, err error, expected, msg string) {
	t.Helper()
	if err == nil {
		t.Errorf("%s: expected error but got none", msg)
		return
	}
	if expected != "" && !ContainsString(err.Error(), expected) {
		t.Errorf("%s: expected error containing %q, got: %v", msg, expected, err)
	}
}

func AssertTrue(t *testing.T, condition bool, msg string) {
	t.Helper()
	if !condition {
		t.Fatalf("%s: expected true", msg)
	}
}

func AssertFalse(t *testing.T, condition bool, msg string) {
	t.Helper()
	if condition {
		t.Fatalf("%s: expected false", msg)
	}
}

func AssertContains(t *testing.T, str, substr, msg string) {
	t.Helper()
	if !strings.Contains(str, substr) {
		t.Fatalf("%s: expected string '%s' to contain '%s'", msg, str, substr)
	}
}

func AssertEqual[T comparable](t *testing.T, expected, actual T, msg string) {
	t.Helper()
	if expected != actual {
		t.Fatalf("%s: expected %v, got %v", msg, expected, actual)
	}
}

func AssertNotEqual[T comparable](t *testing.T, not, actual T, msg string) {
	t.Helper()
	if not == actual {
		t.Fatalf("%s: expected value to not equal %v", msg, not)
	}
}

func AssertNil(t *testing.T, value any, msg string) {
	t.Helper()
	if value != nil {
		t.Fatalf("%s: expected nil, got %v", msg, value)
	}
}

func AssertNotNil(t *testing.T, value any, msg string) {
	t.Helper()
	if value == nil {
		t.Fatalf("%s: expected non-nil value", msg)
	}
}

func AssertGreaterThan[T interface{ int | int64 | float64 }](t *testing.T, actual, threshold T, msg string) {
	t.Helper()
	if actual <= threshold {
		t.Fatalf("%s: expected %v > %v", msg, actual, threshold)
	}
}

func AssertLessThan[T interface{ int | int64 | float64 }](t *testing.T, actual, threshold T, msg string) {
	t.Helper()
	if actual >= threshold {
		t.Fatalf("%s: expected %v < %v", msg, actual, threshold)
	}
}

// Helper function to check if string contains substring (case-insensitive)
func ContainsString(haystack, needle string) bool {
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
	server, err := startHTTPTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.requests = append(m.requests, r)
		handler(w, r)
	}))
	if err != nil {
		panic(err)
	}

	m.server = server
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

func (m HTTPMockServer) Requests() []*http.Request {
	return m.requests
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

// AssertRequestMade verifies that a request was made to the mock server
func AssertRequestMade(t *testing.T, server *HTTPMockServer, expected string) {
	t.Helper()
	if len(server.requests) == 0 {
		t.Error("Expected HTTP request to be made but none were recorded")
		return
	}

	lastReq := server.GetLastRequest()
	if lastReq.URL.Path != expected {
		t.Errorf("Expected request to path %s, got %s", expected, lastReq.URL.Path)
	}
}
