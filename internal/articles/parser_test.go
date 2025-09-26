package articles

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stormlightlabs/noteleaf/internal/models"
)

// ExampleParser_Convert demonstrates parsing a local HTML file using Wikipedia rules.
func ExampleParser_Convert() {
	parser, err := NewArticleParser(http.DefaultClient)
	if err != nil {
		fmt.Printf("Failed to create parser: %v\n", err)
		return
	}

	htmlPath := "examples/christopher-lloyd.html"
	htmlContent, err := os.ReadFile(htmlPath)
	if err != nil {
		fmt.Printf("Local HTML file not found: %v\n", err)
		return
	}

	markdown, err := parser.Convert(string(htmlContent), ".wikipedia.org", "https://en.wikipedia.org/wiki/Christopher_Lloyd")
	if err != nil {
		fmt.Printf("Failed to convert HTML: %v\n", err)
		return
	}

	parts := strings.Split(markdown, "\n---\n")
	if len(parts) > 0 {
		frontmatter := strings.TrimSpace(parts[0])
		lines := strings.Split(frontmatter, "\n")

		for i, line := range lines {
			if i >= 4 {
				break
			}

			if !strings.Contains(line, "**Saved:**") {
				fmt.Println(line)
			}
		}
	}

	// Output: # Christopher Lloyd
	//
	// **Source:** https://en.wikipedia.org/wiki/Christopher_Lloyd
}

func TestArticleParser(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		t.Run("successfully creates parser", func(t *testing.T) {
			parser, err := NewArticleParser(http.DefaultClient)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}
			if parser == nil {
				t.Fatal("Expected parser to be created, got nil")
			}
			if len(parser.rules) == 0 {
				t.Error("Expected rules to be loaded")
			}
		})

		t.Run("loads expected domains", func(t *testing.T) {
			parser, err := NewArticleParser(http.DefaultClient)
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			domains := parser.GetSupportedDomains()
			expectedDomains := []string{".wikipedia.org", "arxiv.org", "baseballprospectus.com"}

			if len(domains) != len(expectedDomains) {
				t.Errorf("Expected %d domains, got %d", len(expectedDomains), len(domains))
			}

			domainMap := make(map[string]bool)
			for _, domain := range domains {
				domainMap[domain] = true
			}

			for _, expected := range expectedDomains {
				if !domainMap[expected] {
					t.Errorf("Expected domain %s not found in supported domains", expected)
				}
			}
		})
	})

	t.Run("parseRules", func(t *testing.T) {
		parser := &ArticleParser{rules: make(map[string]*ParsingRule)}

		t.Run("parses valid rule file", func(t *testing.T) {
			content := `title: //h1
author: //span[@class='author']
date: //time
body: //article
strip: //nav
strip: //footer
test_url: https://example.com/article`

			rule, err := parser.parseRules("example.com", content)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if rule.Domain != "example.com" {
				t.Errorf("Expected domain 'example.com', got %s", rule.Domain)
			}
			if rule.Title != "//h1" {
				t.Errorf("Expected title '//h1', got %s", rule.Title)
			}
			if rule.Author != "//span[@class='author']" {
				t.Errorf("Expected author '//span[@class='author']', got %s", rule.Author)
			}
			if len(rule.Strip) != 2 {
				t.Errorf("Expected 2 strip rules, got %d", len(rule.Strip))
			}
			if len(rule.TestURLs) != 1 {
				t.Errorf("Expected 1 test URL, got %d", len(rule.TestURLs))
			}
		})

		t.Run("handles empty lines and comments", func(t *testing.T) {
			content := `# This is a comment
title: //h1

# Another comment
body: //article
`

			rule, err := parser.parseRules("test.com", content)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if rule.Title != "//h1" {
				t.Errorf("Expected title '//h1', got %s", rule.Title)
			}
			if rule.Body != "//article" {
				t.Errorf("Expected body '//article', got %s", rule.Body)
			}
		})
	})

	t.Run("slugify", func(t *testing.T) {
		parser := &ArticleParser{}

		testCases := []struct {
			input    string
			expected string
		}{
			{"Simple Title", "simple-title"},
			{"Title with Numbers 123", "title-with-numbers-123"},
			{"Title-with-Hyphens", "title-with-hyphens"},
			{"Title with Spaces and    Multiple   Spaces", "title-with-spaces-and-multiple-spaces"},
			{"Title!@#$%^&*()with Special Characters", "title-with-special-characters"},
			{"", ""},
			{strings.Repeat("a", 150), strings.Repeat("a", 100)},
		}

		for _, tc := range testCases {
			t.Run(fmt.Sprintf("slugify '%s'", tc.input), func(t *testing.T) {
				result := parser.slugify(tc.input)
				if result != tc.expected {
					t.Errorf("Expected '%s', got '%s'", tc.expected, result)
				}
			})
		}
	})

	t.Run("Convert", func(t *testing.T) {
		parser, err := NewArticleParser(http.DefaultClient)
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}

		t.Run("fails with unsupported domain", func(t *testing.T) {
			htmlContent := "<html><head><title>Test</title></head><body><p>Content</p></body></html>"
			_, err := parser.Convert(htmlContent, "unsupported.com", "https://unsupported.com/article")

			if err == nil {
				t.Error("Expected error for unsupported domain")
			}
			if !strings.Contains(err.Error(), "no parsing rule found") {
				t.Errorf("Expected 'no parsing rule found' error, got %v", err)
			}
		})

		t.Run("fails with invalid HTML", func(t *testing.T) {
			invalidHTML := "<html><head><title>Test</head></body>"
			_, err := parser.Convert(invalidHTML, ".wikipedia.org", "https://en.wikipedia.org/wiki/Test")

			if err == nil {
				t.Error("Expected error for invalid HTML")
			}
		})

		t.Run("fails when no title extracted", func(t *testing.T) {
			htmlContent := "<html><head><title>Test</title></head><body><p>Content</p></body></html>"
			_, err := parser.Convert(htmlContent, ".wikipedia.org", "https://en.wikipedia.org/wiki/Test")

			if err == nil {
				t.Error("Expected error when no title can be extracted")
			}
			if !strings.Contains(err.Error(), "could not extract title") {
				t.Errorf("Expected 'could not extract title' error, got %v", err)
			}
		})

		t.Run("successfully converts valid Wikipedia HTML", func(t *testing.T) {
			htmlContent := `<html>
			<head><title>Test Article</title></head>
			<body>
				<h1 id="firstHeading">Test Article Title</h1>
				<div id="bodyContent">
					<p>This is the main content of the article.</p>
					<div class="noprint">This should be stripped</div>
					<p>More content here.</p>
				</div>
			</body>
		</html>`

			markdown, err := parser.Convert(htmlContent, ".wikipedia.org", "https://en.wikipedia.org/wiki/Test")
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if !strings.Contains(markdown, "# Test Article Title") {
				t.Error("Expected markdown to contain title")
			}
			if !strings.Contains(markdown, "**Source:** https://en.wikipedia.org/wiki/Test") {
				t.Error("Expected markdown to contain source URL")
			}
			if !strings.Contains(markdown, "This is the main content") {
				t.Error("Expected markdown to contain article content")
			}
			if strings.Contains(markdown, "This should be stripped") {
				t.Error("Expected stripped content to be removed from markdown")
			}
		})
	})

	t.Run("ParseURL", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case strings.Contains(r.URL.Path, "404"):
				w.WriteHeader(http.StatusNotFound)
			case strings.Contains(r.URL.Path, "unsupported"):
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("<html><head><title>Test</title></head><body><p>Content</p></body></html>"))
			default:
				// Return Wikipedia-like structure for localhost rule
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`<html>
					<head><title>Test Article</title></head>
					<body>
						<h1 id="firstHeading">Test Wikipedia Article</h1>
						<div id="bodyContent">
							<p>This is the article content.</p>
							<div class="noprint">This gets stripped</div>
						</div>
					</body>
				</html>`))
			}
		}))
		defer server.Close()

		parser, err := NewArticleParser(server.Client())
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}

		localhostRule := &ParsingRule{
			Domain: "127.0.0.1",
			Title:  "//h1[@id='firstHeading']",
			Body:   "//div[@id='bodyContent']",
			Strip:  []string{"//div[@class='noprint']"},
		}
		parser.AddRule("127.0.0.1", localhostRule)

		t.Run("fails with invalid URL", func(t *testing.T) {
			_, err := parser.ParseURL("not-a-url")
			if err == nil {
				t.Error("Expected error for invalid URL")
			}
			if !strings.Contains(err.Error(), "unsupported protocol scheme") {
				t.Errorf("Expected 'unsupported protocol scheme' error, got %v", err)
			}
		})

		t.Run("fails with unsupported domain", func(t *testing.T) {
			_, err := parser.ParseURL(server.URL + "/unsupported.com")
			if err == nil {
				t.Error("Expected error for unsupported domain")
			}
		})

		t.Run("fails with HTTP error", func(t *testing.T) {
			_, err := parser.ParseURL(server.URL + "/404/en.wikipedia.org/wiki/test")
			if err == nil {
				t.Error("Expected error for HTTP 404")
			}
		})

	})

	t.Run("SaveArticle", func(t *testing.T) {
		parser := &ArticleParser{}
		tempDir := t.TempDir()

		content := &ParsedContent{
			Title:   "Test Article",
			Author:  "Test Author",
			Date:    "2023-01-01",
			Content: "This is test content.",
			URL:     "https://example.com/test",
		}

		t.Run("successfully saves article", func(t *testing.T) {
			mdPath, htmlPath, err := parser.SaveArticle(content, tempDir)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if _, err := os.Stat(mdPath); os.IsNotExist(err) {
				t.Error("Expected markdown file to exist")
			}
			if _, err := os.Stat(htmlPath); os.IsNotExist(err) {
				t.Error("Expected HTML file to exist")
			}

			mdContent, err := os.ReadFile(mdPath)
			if err != nil {
				t.Fatalf("Failed to read markdown file: %v", err)
			}
			if !strings.Contains(string(mdContent), "# Test Article") {
				t.Error("Expected markdown to contain title")
			}
			if !strings.Contains(string(mdContent), "**Author:** Test Author") {
				t.Error("Expected markdown to contain author")
			}

			htmlContentBytes, err := os.ReadFile(htmlPath)
			if err != nil {
				t.Fatalf("Failed to read HTML file: %v", err)
			}
			if !strings.Contains(string(htmlContentBytes), "<title>Test Article</title>") {
				t.Error("Expected HTML to contain title")
			}
		})

		t.Run("handles duplicate filenames", func(t *testing.T) {
			mdPath1, htmlPath1, err := parser.SaveArticle(content, tempDir)
			if err != nil {
				t.Fatalf("Expected no error for first save, got %v", err)
			}

			mdPath2, htmlPath2, err := parser.SaveArticle(content, tempDir)
			if err != nil {
				t.Fatalf("Expected no error for second save, got %v", err)
			}

			if mdPath1 == mdPath2 {
				t.Error("Expected different markdown paths for duplicate saves")
			}
			if htmlPath1 == htmlPath2 {
				t.Error("Expected different HTML paths for duplicate saves")
			}

			if _, err := os.Stat(mdPath1); os.IsNotExist(err) {
				t.Error("Expected first markdown file to exist")
			}
			if _, err := os.Stat(mdPath2); os.IsNotExist(err) {
				t.Error("Expected second markdown file to exist")
			}
		})

		t.Run("fails with invalid directory", func(t *testing.T) {
			invalidDir := "/nonexistent/directory"
			_, _, err := parser.SaveArticle(content, invalidDir)
			if err == nil {
				t.Error("Expected error for invalid directory")
			}
		})
	})

	t.Run("createHTML", func(t *testing.T) {
		parser := &ArticleParser{}
		content := &ParsedContent{
			Title:   "Test HTML Article",
			Author:  "HTML Author",
			Date:    "2023-12-25",
			Content: "This is **bold** content with *emphasis*.",
			URL:     "https://example.com/html-test",
		}

		t.Run("creates valid HTML", func(t *testing.T) {
			markdown := parser.createMarkdown(content)
			html := parser.createHTML(content, markdown)

			if !strings.Contains(html, "<!DOCTYPE html>") {
				t.Error("Expected HTML to contain DOCTYPE")
			}
			if !strings.Contains(html, "<title>Test HTML Article</title>") {
				t.Error("Expected HTML to contain title")
			}
			if !strings.Contains(html, "<h1") || !strings.Contains(html, "Test HTML Article") {
				t.Error("Expected HTML to contain h1 heading with title")
			}
			if !strings.Contains(html, "<strong>bold</strong>") {
				t.Error("Expected HTML to contain bold formatting")
			}
			if !strings.Contains(html, "<em>emphasis</em>") {
				t.Error("Expected HTML to contain emphasis formatting")
			}
		})
	})

	t.Run("createMarkdown", func(t *testing.T) {
		parser := &ArticleParser{}

		t.Run("creates markdown with all fields", func(t *testing.T) {
			content := &ParsedContent{
				Title:   "Full Content Article",
				Author:  "Complete Author",
				Date:    "2023-01-15",
				Content: "Complete article content here.",
				URL:     "https://example.com/full",
			}

			markdown := parser.createMarkdown(content)

			if !strings.Contains(markdown, "# Full Content Article") {
				t.Error("Expected markdown to contain title")
			}
			if !strings.Contains(markdown, "**Author:** Complete Author") {
				t.Error("Expected markdown to contain author")
			}
			if !strings.Contains(markdown, "**Date:** 2023-01-15") {
				t.Error("Expected markdown to contain date")
			}
			if !strings.Contains(markdown, "**Source:** https://example.com/full") {
				t.Error("Expected markdown to contain source URL")
			}
			if !strings.Contains(markdown, "**Saved:**") {
				t.Error("Expected markdown to contain saved timestamp")
			}
			if !strings.Contains(markdown, "---") {
				t.Error("Expected markdown to contain separator")
			}
			if !strings.Contains(markdown, "Complete article content here.") {
				t.Error("Expected markdown to contain article content")
			}
		})

		t.Run("creates markdown with minimal fields", func(t *testing.T) {
			content := &ParsedContent{
				Title:   "Minimal Article",
				Content: "Just content.",
				URL:     "https://example.com/minimal",
			}

			markdown := parser.createMarkdown(content)

			if !strings.Contains(markdown, "# Minimal Article") {
				t.Error("Expected markdown to contain title")
			}
			if strings.Contains(markdown, "**Author:**") {
				t.Error("Expected no author field for empty author")
			}
			if strings.Contains(markdown, "**Date:**") {
				t.Error("Expected no date field for empty date")
			}
			if !strings.Contains(markdown, "**Source:** https://example.com/minimal") {
				t.Error("Expected markdown to contain source URL")
			}
		})
	})
}

func TestCreateArticleFromURL(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("fails with invalid URL", func(t *testing.T) {
		_, err := CreateArticleFromURL("not-a-url", tempDir)
		if err == nil {
			t.Error("Expected error for invalid URL")
		}
		if !strings.Contains(err.Error(), "invalid URL") && !strings.Contains(err.Error(), "failed to parse URL") {
			t.Errorf("Expected URL parsing error, got %v", err)
		}
	})

	t.Run("fails with empty URL", func(t *testing.T) {
		_, err := CreateArticleFromURL("", tempDir)
		if err == nil {
			t.Error("Expected error for empty URL")
		}
	})

	t.Run("fails with unsupported domain", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("<html><head><title>Test</title></head><body><p>Content</p></body></html>"))
		}))
		defer server.Close()

		_, err := CreateArticleFromURL(server.URL, tempDir)
		if err == nil {
			t.Error("Expected error for unsupported domain")
		}
		if !strings.Contains(err.Error(), "no parsing rule found") {
			t.Errorf("Expected 'no parsing rule found' error, got %v", err)
		}
	})

	t.Run("fails with HTTP error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		// Use a direct Wikipedia URL that would be processed by the real function
		_, err := CreateArticleFromURL("https://en.wikipedia.org/wiki/NonExistentPage12345", tempDir)
		if err == nil {
			t.Error("Expected error for HTTP 404")
		}
		if !strings.Contains(err.Error(), "HTTP error") && !strings.Contains(err.Error(), "404") {
			t.Errorf("Expected HTTP error, got %v", err)
		}
	})

	t.Run("fails with network error", func(t *testing.T) {
		// Use a non-existent server to trigger network error
		_, err := CreateArticleFromURL("http://localhost:99999/test", tempDir)
		if err == nil {
			t.Error("Expected error for network failure")
		}
		if !strings.Contains(err.Error(), "failed to fetch URL") && !strings.Contains(err.Error(), "connection refused") {
			t.Errorf("Expected network error, got %v", err)
		}
	})

	t.Run("fails with invalid directory", func(t *testing.T) {
		// Skip this test as it would require network access to test with real URLs
		t.Skip("Skipping invalid directory test - requires network access")
	})

	t.Run("fails with malformed HTML", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("<html><head><title>Test</head></body>")) // Malformed HTML
		}))
		defer server.Close()

		// Create a custom parser with localhost rule for testing
		parser, err := NewArticleParser(server.Client())
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}

		localhostRule := &ParsingRule{
			Domain: "127.0.0.1",
			Title:  "//h1[@id='firstHeading']",
			Body:   "//div[@id='bodyContent']",
			Strip:  []string{"//div[@class='noprint']"},
		}
		parser.AddRule("127.0.0.1", localhostRule)

		_, err = parser.ParseURL(server.URL)
		if err == nil {
			t.Error("Expected error for malformed HTML")
		}
		// Malformed HTML may either fail to parse or fail to extract title
		if !strings.Contains(err.Error(), "failed to parse HTML") && !strings.Contains(err.Error(), "could not extract title") {
			t.Errorf("Expected HTML parsing or title extraction error, got %v", err)
		}
	})

	t.Run("fails when no title can be extracted", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<html>
				<head><title>Test</title></head>
				<body>
					<div id="bodyContent">
						<p>Content without proper title</p>
					</div>
				</body>
			</html>`)) // No h1 with id="firstHeading"
		}))
		defer server.Close()

		// Create a custom parser with localhost rule for testing
		parser, err := NewArticleParser(server.Client())
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}

		localhostRule := &ParsingRule{
			Domain: "127.0.0.1",
			Title:  "//h1[@id='firstHeading']",
			Body:   "//div[@id='bodyContent']",
			Strip:  []string{"//div[@class='noprint']"},
		}
		parser.AddRule("127.0.0.1", localhostRule)

		_, err = parser.ParseURL(server.URL)
		if err == nil {
			t.Error("Expected error when no title can be extracted")
		}
		if !strings.Contains(err.Error(), "could not extract title") {
			t.Errorf("Expected 'could not extract title' error, got %v", err)
		}
	})

	t.Run("successfully creates article structure from parsed content", func(t *testing.T) {
		wikipediaHTML := `<html>
			<head><title>Integration Test Article</title></head>
			<body>
				<h1 id="firstHeading">Integration Test Article</h1>
				<div id="bodyContent">
					<p>This is integration test content.</p>
					<div class="noprint">This should be stripped</div>
					<p>More content here.</p>
				</div>
			</body>
		</html>`

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(wikipediaHTML))
		}))
		defer server.Close()

		// Create a custom parser with localhost rule for testing
		parser, err := NewArticleParser(server.Client())
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}

		localhostRule := &ParsingRule{
			Domain: "127.0.0.1",
			Title:  "//h1[@id='firstHeading']",
			Body:   "//div[@id='bodyContent']",
			Strip:  []string{"//div[@class='noprint']"},
		}
		parser.AddRule("127.0.0.1", localhostRule)

		content, err := parser.ParseURL(server.URL)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		mdPath, htmlPath, err := parser.SaveArticle(content, tempDir)
		if err != nil {
			t.Fatalf("Failed to save article: %v", err)
		}

		// Test that it creates a proper models.Article structure (simulating CreateArticleFromURL)
		article := &models.Article{
			URL:          server.URL,
			Title:        content.Title,
			MarkdownPath: mdPath,
			HTMLPath:     htmlPath,
			Created:      time.Now(),
			Modified:     time.Now(),
		}

		if article.Title != "Integration Test Article" {
			t.Errorf("Expected title 'Integration Test Article', got %s", article.Title)
		}
		if article.URL != server.URL {
			t.Errorf("Expected URL %s, got %s", server.URL, article.URL)
		}
		if article.MarkdownPath == "" {
			t.Error("Expected non-empty markdown path")
		}
		if article.HTMLPath == "" {
			t.Error("Expected non-empty HTML path")
		}
		if article.Created.IsZero() {
			t.Error("Expected Created timestamp to be set")
		}
		if article.Modified.IsZero() {
			t.Error("Expected Modified timestamp to be set")
		}

		// Check files exist
		if _, err := os.Stat(article.MarkdownPath); os.IsNotExist(err) {
			t.Error("Expected markdown file to exist")
		}
		if _, err := os.Stat(article.HTMLPath); os.IsNotExist(err) {
			t.Error("Expected HTML file to exist")
		}

		// Verify file contents
		mdContent, err := os.ReadFile(article.MarkdownPath)
		if err != nil {
			t.Fatalf("Failed to read markdown file: %v", err)
		}
		if !strings.Contains(string(mdContent), "# Integration Test Article") {
			t.Error("Expected markdown to contain title")
		}
		if !strings.Contains(string(mdContent), "This is integration test content") {
			t.Error("Expected markdown to contain article content")
		}
		if strings.Contains(string(mdContent), "This should be stripped") {
			t.Error("Expected stripped content to be removed from markdown")
		}

		htmlContent, err := os.ReadFile(article.HTMLPath)
		if err != nil {
			t.Fatalf("Failed to read HTML file: %v", err)
		}
		if !strings.Contains(string(htmlContent), "<title>Integration Test Article</title>") {
			t.Error("Expected HTML to contain title")
		}
		if !strings.Contains(string(htmlContent), "<!DOCTYPE html>") {
			t.Error("Expected HTML to contain DOCTYPE")
		}
	})

	t.Run("successfully handles article with metadata", func(t *testing.T) {
		contentHTML := `<html>
			<head>
				<title>Test Paper</title>
				<meta name="citation_author" content="Dr. Test Author">
				<meta name="citation_date" content="2024-01-01">
			</head>
			<body>
				<h1 class="title">Test Research Paper</h1>
				<blockquote class="abstract">
					<p>This is the abstract of the research paper.</p>
					<p>It contains important research findings.</p>
				</blockquote>
			</body>
		</html>`

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(contentHTML))
		}))
		defer server.Close()

		// Create a custom parser with arXiv-like rule for testing
		parser, err := NewArticleParser(server.Client())
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}

		localhostRule := &ParsingRule{
			Domain: "127.0.0.1",
			Title:  "//h1[contains(concat(' ',normalize-space(@class),' '),' title ')]",
			Body:   "//blockquote[contains(concat(' ',normalize-space(@class),' '),' abstract ')]",
			Date:   "//meta[@name='citation_date']/@content",
			Author: "//meta[@name='citation_author']/@content",
		}
		parser.AddRule("127.0.0.1", localhostRule)

		content, err := parser.ParseURL(server.URL)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if content.Title != "Test Research Paper" {
			t.Errorf("Expected title 'Test Research Paper', got %s", content.Title)
		}
		if content.Author != "Dr. Test Author" {
			t.Errorf("Expected author 'Dr. Test Author', got %s", content.Author)
		}
		if content.Date != "2024-01-01" {
			t.Errorf("Expected date '2024-01-01', got %s", content.Date)
		}

		mdPath, _, err := parser.SaveArticle(content, tempDir)
		if err != nil {
			t.Fatalf("Failed to save article: %v", err)
		}

		// Verify markdown contains all metadata
		mdContent, err := os.ReadFile(mdPath)
		if err != nil {
			t.Fatalf("Failed to read markdown file: %v", err)
		}
		if !strings.Contains(string(mdContent), "**Author:** Dr. Test Author") {
			t.Error("Expected markdown to contain author")
		}
		if !strings.Contains(string(mdContent), "**Date:** 2024-01-01") {
			t.Error("Expected markdown to contain date")
		}

		article := &models.Article{
			Author: content.Author,
			Date:   content.Date,
		}

		if article.Author != "Dr. Test Author" {
			t.Errorf("Expected article author 'Dr. Test Author', got %s", article.Author)
		}
		if article.Date != "2024-01-01" {
			t.Errorf("Expected article date '2024-01-01', got %s", article.Date)
		}
	})
}
