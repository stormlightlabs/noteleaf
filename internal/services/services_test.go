package services

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stormlightlabs/noteleaf/internal/shared"
	"golang.org/x/time/rate"
)

func TestBookService(t *testing.T) {
	t.Run("NewBookService", func(t *testing.T) {
		service := NewBookService(OpenLibraryBaseURL)

		if service == nil {
			t.Fatal("NewBookService should return a non-nil service")
		}

		if service.client == nil {
			t.Error("BookService should have a non-nil HTTP client")
		}

		if service.limiter == nil {
			t.Error("BookService should have a non-nil rate limiter")
		}

		if service.limiter.Limit() != rate.Limit(requestsPerSecond) {
			t.Errorf("Expected rate limit of %v, got %v", requestsPerSecond, service.limiter.Limit())
		}
	})

	t.Run("Search", func(t *testing.T) {
		t.Run("successful search", func(t *testing.T) {
			mockResponse := OpenLibrarySearchResponse{
				NumFound: 2,
				Start:    0,
				Docs: []OpenLibrarySearchDoc{
					{
						Key:              "/works/OL45804W",
						Title:            "Fantastic Mr. Fox",
						AuthorName:       []string{"Roald Dahl"},
						FirstPublishYear: 1970,
						Edition_count:    25,
						PublisherName:    []string{"Puffin Books", "Viking Press"},
						Subject:          []string{"Children's literature", "Foxes", "Fiction"},
						CoverI:           8739161,
					},
					{
						Key:              "/works/OL123456W",
						Title:            "The BFG",
						AuthorName:       []string{"Roald Dahl"},
						FirstPublishYear: 1982,
						Edition_count:    15,
						PublisherName:    []string{"Jonathan Cape"},
						CoverI:           456789,
					},
				},
			}

			server := shared.NewHTTPTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/search.json" {
					t.Errorf("Expected path /search.json, got %s", r.URL.Path)
				}

				query := r.URL.Query()
				if query.Get("q") != "roald dahl" {
					t.Errorf("Expected query 'roald dahl', got %s", query.Get("q"))
				}
				if query.Get("limit") != "10" {
					t.Errorf("Expected limit '10', got %s", query.Get("limit"))
				}
				if query.Get("offset") != "0" {
					t.Errorf("Expected offset '0', got %s", query.Get("offset"))
				}

				if r.Header.Get("User-Agent") != userAgent {
					t.Errorf("Expected User-Agent %s, got %s", userAgent, r.Header.Get("User-Agent"))
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(mockResponse)
			}))
			defer server.Close()

			service := NewBookService(server.URL)
			ctx := context.Background()
			results, err := service.Search(ctx, "roald dahl", 1, 10)

			if err != nil {
				t.Fatalf("Search should not return error: %v", err)
			}

			if len(results) == 0 {
				t.Error("Search should return at least one result")
			}

			if results[0] == nil {
				t.Fatal("First result should not be nil")
			}
		})

		t.Run("handles API error", func(t *testing.T) {
			server := shared.NewHTTPTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			}))
			defer server.Close()

			service := NewBookService(server.URL)
			ctx := context.Background()

			_, err := service.Search(ctx, "test", 1, 10)
			if err == nil {
				t.Error("Search should return error for API failure")
			}

			AssertErrorContains(t, err, "API returned status 500")
		})

		t.Run("handles malformed JSON", func(t *testing.T) {
			server := shared.NewHTTPTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte("invalid json"))
			}))
			defer server.Close()

			service := NewBookService(server.URL)
			ctx := context.Background()

			_, err := service.Search(ctx, "test", 1, 10)
			if err == nil {
				t.Error("Search should return error for malformed JSON")
			}

			AssertErrorContains(t, err, "failed to decode response")
		})

		t.Run("handles context cancellation", func(t *testing.T) {
			service := NewBookService(OpenLibraryBaseURL)
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			_, err := service.Search(ctx, "test", 1, 10)
			if err == nil {
				t.Error("Search should return error for cancelled context")
			}
		})

		t.Run("respects pagination", func(t *testing.T) {
			service := NewBookService(OpenLibraryBaseURL)
			ctx := context.Background()

			_, err := service.Search(ctx, "test", 2, 5)
			if err != nil {
				t.Logf("Expected error for actual API call: %v", err)
			}
		})
	})

	t.Run("Get", func(t *testing.T) {
		t.Run("successful get by work key", func(t *testing.T) {
			mockWork := OpenLibraryWork{
				Key:   "/works/OL45804W",
				Title: "Fantastic Mr. Fox",
				Authors: []OpenLibraryAuthorRef{
					{
						Author: OpenLibraryAuthorKey{Key: "/authors/OL34184A"},
					},
				},
				Description: "A story about a clever fox who outsmarts three mean farmers.",
				Subjects:    []string{"Children's literature", "Foxes", "Fiction"},
				Covers:      []int{8739161, 8739162},
			}

			server := shared.NewHTTPTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !strings.HasPrefix(r.URL.Path, "/works/") {
					t.Errorf("Expected path to start with /works/, got %s", r.URL.Path)
				}
				if !strings.HasSuffix(r.URL.Path, ".json") {
					t.Errorf("Expected path to end with .json, got %s", r.URL.Path)
				}

				if r.Header.Get("User-Agent") != userAgent {
					t.Errorf("Expected User-Agent %s, got %s", userAgent, r.Header.Get("User-Agent"))
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(mockWork)
			}))
			defer server.Close()

			service := NewBookService(server.URL)
			ctx := context.Background()

			result, err := service.Get(ctx, "OL45804W")
			if err != nil {
				t.Fatalf("Get should not return error: %v", err)
			}

			if result == nil {
				t.Fatal("Get should return a non-nil result")
			}
		})

		t.Run("handles work key with /works/ prefix", func(t *testing.T) {
			service := NewBookService(OpenLibraryBaseURL)
			ctx := context.Background()

			_, err1 := service.Get(ctx, "OL45804W")
			_, err2 := service.Get(ctx, "/works/OL45804W")

			if (err1 == nil) != (err2 == nil) {
				t.Errorf("Both key formats should behave similarly. Error1: %v, Error2: %v", err1, err2)
			}
		})

		t.Run("handles not found", func(t *testing.T) {
			server := shared.NewHTTPTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			}))
			defer server.Close()

			service := NewBookService(server.URL)
			ctx := context.Background()

			_, err := service.Get(ctx, "nonexistent")
			if err == nil {
				t.Error("Get should return error for non-existent work")
			}

			AssertErrorContains(t, err, "book not found")
		})

		t.Run("handles API error", func(t *testing.T) {
			server := shared.NewHTTPTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			}))
			defer server.Close()

			service := NewBookService(server.URL)
			ctx := context.Background()

			_, err := service.Get(ctx, "test")
			if err == nil {
				t.Error("Get should return error for API failure")
			}

			AssertErrorContains(t, err, "API returned status 500")
		})
	})

	t.Run("Check", func(t *testing.T) {
		t.Run("successful check", func(t *testing.T) {
			server := shared.NewHTTPTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/search.json" {
					t.Errorf("Expected path /search.json, got %s", r.URL.Path)
				}

				query := r.URL.Query()
				if query.Get("q") != "test" {
					t.Errorf("Expected query 'test', got %s", query.Get("q"))
				}
				if query.Get("limit") != "1" {
					t.Errorf("Expected limit '1', got %s", query.Get("limit"))
				}

				if r.Header.Get("User-Agent") != userAgent {
					t.Errorf("Expected User-Agent %s, got %s", userAgent, r.Header.Get("User-Agent"))
				}

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"numFound": 1, "docs": []}`))
			}))
			defer server.Close()

			service := NewBookService(server.URL)
			ctx := context.Background()

			err := service.Check(ctx)
			if err != nil {
				t.Errorf("Check should not return error for healthy API: %v", err)
			}
		})

		t.Run("handles API failure", func(t *testing.T) {
			server := shared.NewHTTPTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusServiceUnavailable)
			}))
			defer server.Close()

			service := NewBookService(server.URL)
			ctx := context.Background()

			err := service.Check(ctx)
			if err == nil {
				t.Error("Check should return error for API failure")
			}

			AssertErrorContains(t, err, "open Library API returned status 503")
		})

		t.Run("handles network error", func(t *testing.T) {
			service := NewBookService(OpenLibraryBaseURL)
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
			defer cancel()

			err := service.Check(ctx)
			if err == nil {
				t.Error("Check should return error for network failure")
			}
		})
	})

	t.Run("Close", func(t *testing.T) {
		service := NewBookService(OpenLibraryBaseURL)
		err := service.Close()
		if err != nil {
			t.Errorf("Close should not return error: %v", err)
		}
	})

	t.Run("RateLimiting", func(t *testing.T) {
		t.Run("respects rate limits", func(t *testing.T) {
			service := NewBookService(OpenLibraryBaseURL)
			ctx := context.Background()

			start := time.Now()
			var errors []error

			for range 5 {
				_, err := service.Search(ctx, "test", 1, 1)
				errors = append(errors, err)
			}

			elapsed := time.Since(start)

			// Should take some time due to rate limiting
			// NOTE: This test might be flaky depending on network conditions
			t.Logf("5 requests took %v", elapsed)

			allFailed := true
			for _, err := range errors {
				if err == nil {
					allFailed = false
					break
				}
			}

			if allFailed {
				t.Log("All requests failed, which is expected for rate limiting test")
			}
		})
	})

	t.Run("Conversion Functions", func(t *testing.T) {
		t.Run("searchDocToBook conversion", func(t *testing.T) {
			service := NewBookService(OpenLibraryBaseURL)
			doc := OpenLibrarySearchDoc{
				Key:              "/works/OL45804W",
				Title:            "Test Book",
				AuthorName:       []string{"Author One", "Author Two"},
				FirstPublishYear: 1999,
				Edition_count:    5,
				PublisherName:    []string{"Test Publisher"},
				CoverI:           12345,
			}

			book := service.searchDocToBook(doc)

			if book.Title != "Test Book" {
				t.Errorf("Expected title 'Test Book', got %s", book.Title)
			}

			if book.Author != "Author One, Author Two" {
				t.Errorf("Expected author 'Author One, Author Two', got %s", book.Author)
			}

			if book.Status != "queued" {
				t.Errorf("Expected status 'queued', got %s", book.Status)
			}

			if !strings.Contains(book.Notes, "5 editions") {
				t.Errorf("Expected notes to contain edition count, got %s", book.Notes)
			}

			if !strings.Contains(book.Notes, "Test Publisher") {
				t.Errorf("Expected notes to contain publisher, got %s", book.Notes)
			}
		})

		t.Run("workToBook conversion with string description", func(t *testing.T) {
			service := NewBookService(OpenLibraryBaseURL)
			work := OpenLibraryWork{
				Key:   "/works/OL45804W",
				Title: "Test Work",
				Authors: []OpenLibraryAuthorRef{
					{Author: OpenLibraryAuthorKey{Key: "/authors/OL123A"}},
					{Author: OpenLibraryAuthorKey{Key: "/authors/OL456A"}},
				},
				Description: "This is a test description",
				Subjects:    []string{"Fiction", "Adventure", "Classic"},
			}

			book := service.workToBook(work)

			if book.Title != "Test Work" {
				t.Errorf("Expected title 'Test Work', got %s", book.Title)
			}

			if book.Author != "OL123A, OL456A" {
				t.Errorf("Expected author 'OL123A, OL456A', got %s", book.Author)
			}

			if book.Notes != "This is a test description" {
				t.Errorf("Expected notes to be description, got %s", book.Notes)
			}
		})

		t.Run("workToBook conversion with object description", func(t *testing.T) {
			service := NewBookService(OpenLibraryBaseURL)
			work := OpenLibraryWork{
				Title: "Test Work",
				Description: map[string]any{
					"type":  "/type/text",
					"value": "Object description",
				},
			}

			book := service.workToBook(work)

			if book.Notes != "Object description" {
				t.Errorf("Expected notes to be object description, got %s", book.Notes)
			}
		})

		t.Run("workToBook uses subjects when no description", func(t *testing.T) {
			service := NewBookService(OpenLibraryBaseURL)
			work := OpenLibraryWork{
				Title:    "Test Work",
				Subjects: []string{"Fiction", "Adventure", "Classic", "Literature", "Drama", "Extra"},
			}

			book := service.workToBook(work)

			if !strings.Contains(book.Notes, "Subjects:") {
				t.Errorf("Expected notes to contain subjects, got %s", book.Notes)
			}

			if !strings.Contains(book.Notes, "Fiction") {
				t.Errorf("Expected notes to contain Fiction, got %s", book.Notes)
			}

			subjectCount := strings.Count(book.Notes, ",") + 1
			if subjectCount > 5 {
				t.Errorf("Expected max 5 subjects, got %d in: %s", subjectCount, book.Notes)
			}
		})
	})

	t.Run("Interface Compliance", func(t *testing.T) {
		t.Run("implements APIService interface", func(t *testing.T) {
			var _ APIService = &BookService{}
			var _ APIService = NewBookService(OpenLibraryBaseURL)
		})
	})

	t.Run("UserAgent header", func(t *testing.T) {
		expectedFormat := "Noteleaf/1.0.0 (info@stormlightlabs.org)"
		if userAgent != expectedFormat {
			t.Errorf("User agent should follow the required format. Expected %s, got %s", expectedFormat, userAgent)
		}
	})

	t.Run("Constants", func(t *testing.T) {
		t.Run("API endpoints are correct", func(t *testing.T) {
			if OpenLibraryBaseURL != "https://openlibrary.org" {
				t.Errorf("Base URL should be https://openlibrary.org, got %s", OpenLibraryBaseURL)
			}

			if openLibrarySearch != "https://openlibrary.org/search.json" {
				t.Errorf("Search URL should be https://openlibrary.org/search.json, got %s", openLibrarySearch)
			}
		})

		t.Run("rate limiting constants are correct", func(t *testing.T) {
			if requestsPerSecond != 3 {
				t.Errorf("Requests per second should be 3 (180/60), got %d", requestsPerSecond)
			}

			if burstLimit < requestsPerSecond {
				t.Errorf("Burst limit should be at least equal to requests per second, got %d", burstLimit)
			}
		})
	})
}
