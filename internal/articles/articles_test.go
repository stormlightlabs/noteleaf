package articles

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// ExampleParser_Convert demonstrates parsing a local HTML file using Wikipedia rules.
func ExampleParser_Convert() {
	parser, err := NewArticleParser()
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
			parser, err := NewArticleParser()
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
			parser, err := NewArticleParser()
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

	t.Run("parseRuleFile", func(t *testing.T) {
		parser := &ArticleParser{rules: make(map[string]*ParsingRule)}

		t.Run("parses valid rule file", func(t *testing.T) {
			content := `title: //h1
author: //span[@class='author']
date: //time
body: //article
strip: //nav
strip: //footer
test_url: https://example.com/article`

			rule, err := parser.parseRuleFile("example.com", content)
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

			rule, err := parser.parseRuleFile("test.com", content)
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
		parser, err := NewArticleParser()
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
}
