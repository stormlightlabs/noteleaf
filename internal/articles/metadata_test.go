package articles

import (
	"strings"
	"testing"
)

func TestMetadataExtractor(t *testing.T) {
	t.Run("NewMetadataExtractor", func(t *testing.T) {
		t.Run("creates extractor", func(t *testing.T) {
			extractor := NewMetadataExtractor()

			if extractor == nil {
				t.Fatal("Expected extractor to be created, got nil")
			}
		})
	})

	t.Run("ExtractTitle", func(t *testing.T) {
		extractor := NewMetadataExtractor()

		t.Run("extracts from OpenGraph", func(t *testing.T) {
			htmlStr := `<html><head>
				<meta property="og:title" content="Article Title from OpenGraph">
			</head><body></body></html>`
			doc := parseHTML(htmlStr)

			title := extractor.ExtractTitle(doc)

			if title != "Article Title from OpenGraph" {
				t.Errorf("Expected OpenGraph title, got %q", title)
			}
		})

		t.Run("extracts from title tag", func(t *testing.T) {
			htmlStr := `<html><head>
				<title>Page Title from Title Tag</title>
			</head><body></body></html>`
			doc := parseHTML(htmlStr)

			title := extractor.ExtractTitle(doc)

			if title != "Page Title from Title Tag" {
				t.Errorf("Expected title tag content, got %q", title)
			}
		})

		t.Run("extracts from h1", func(t *testing.T) {
			htmlStr := `<html><body>
				<h1>Heading Title</h1>
			</body></html>`
			doc := parseHTML(htmlStr)

			title := extractor.ExtractTitle(doc)

			if title != "Heading Title" {
				t.Errorf("Expected h1 content, got %q", title)
			}
		})

		t.Run("returns empty for nil document", func(t *testing.T) {
			title := extractor.ExtractTitle(nil)

			if title != "" {
				t.Errorf("Expected empty string for nil document, got %q", title)
			}
		})

		t.Run("prioritizes OpenGraph over title tag", func(t *testing.T) {
			htmlStr := `<html><head>
				<meta property="og:title" content="OpenGraph Title">
				<title>HTML Title</title>
			</head><body></body></html>`
			doc := parseHTML(htmlStr)

			title := extractor.ExtractTitle(doc)

			if title != "OpenGraph Title" {
				t.Errorf("Expected OpenGraph title to have priority, got %q", title)
			}
		})
	})

	t.Run("ExtractAuthor", func(t *testing.T) {
		extractor := NewMetadataExtractor()

		t.Run("extracts from OpenGraph", func(t *testing.T) {
			htmlStr := `<html><head>
				<meta property="og:author" content="John Doe">
			</head><body></body></html>`
			doc := parseHTML(htmlStr)

			author := extractor.ExtractAuthor(doc)

			if author != "John Doe" {
				t.Errorf("Expected OpenGraph author, got %q", author)
			}
		})

		t.Run("extracts from meta tag", func(t *testing.T) {
			htmlStr := `<html><head>
				<meta name="author" content="Jane Smith">
			</head><body></body></html>`
			doc := parseHTML(htmlStr)

			author := extractor.ExtractAuthor(doc)

			if author != "Jane Smith" {
				t.Errorf("Expected meta tag author, got %q", author)
			}
		})

		t.Run("extracts from rel=author link", func(t *testing.T) {
			htmlStr := `<html><body>
				<a rel="author" href="/author/bob">Bob Johnson</a>
			</body></html>`
			doc := parseHTML(htmlStr)

			author := extractor.ExtractAuthor(doc)

			if author != "Bob Johnson" {
				t.Errorf("Expected rel=author link text, got %q", author)
			}
		})

		t.Run("extracts from byline class", func(t *testing.T) {
			htmlStr := `<html><body>
				<span class="author-name">Alice Brown</span>
			</body></html>`
			doc := parseHTML(htmlStr)

			author := extractor.ExtractAuthor(doc)

			if author != "Alice Brown" {
				t.Errorf("Expected byline class text, got %q", author)
			}
		})

		t.Run("returns empty for nil document", func(t *testing.T) {
			author := extractor.ExtractAuthor(nil)

			if author != "" {
				t.Errorf("Expected empty string for nil document, got %q", author)
			}
		})
	})

	t.Run("ExtractPublishedDate", func(t *testing.T) {
		extractor := NewMetadataExtractor()

		t.Run("extracts from OpenGraph", func(t *testing.T) {
			htmlStr := `<html><head>
				<meta property="og:published_time" content="2025-01-15T10:00:00Z">
			</head><body></body></html>`
			doc := parseHTML(htmlStr)

			date := extractor.ExtractPublishedDate(doc)

			if date != "2025-01-15T10:00:00Z" {
				t.Errorf("Expected OpenGraph date, got %q", date)
			}
		})

		t.Run("extracts from article:published_time", func(t *testing.T) {
			htmlStr := `<html><head>
				<meta property="article:published_time" content="2025-02-20">
			</head><body></body></html>`
			doc := parseHTML(htmlStr)

			date := extractor.ExtractPublishedDate(doc)

			if date != "2025-02-20" {
				t.Errorf("Expected article:published_time, got %q", date)
			}
		})

		t.Run("extracts from time element", func(t *testing.T) {
			htmlStr := `<html><body>
				<time datetime="2025-03-25T14:30:00">March 25, 2025</time>
			</body></html>`
			doc := parseHTML(htmlStr)

			date := extractor.ExtractPublishedDate(doc)

			if date != "2025-03-25T14:30:00" {
				t.Errorf("Expected time element datetime, got %q", date)
			}
		})

		t.Run("returns empty for nil document", func(t *testing.T) {
			date := extractor.ExtractPublishedDate(nil)

			if date != "" {
				t.Errorf("Expected empty string for nil document, got %q", date)
			}
		})
	})

	t.Run("ExtractSiteName", func(t *testing.T) {
		extractor := NewMetadataExtractor()

		t.Run("extracts from OpenGraph", func(t *testing.T) {
			htmlStr := `<html><head>
				<meta property="og:site_name" content="Example News">
			</head><body></body></html>`
			doc := parseHTML(htmlStr)

			siteName := extractor.ExtractSiteName(doc)

			if siteName != "Example News" {
				t.Errorf("Expected OpenGraph site_name, got %q", siteName)
			}
		})

		t.Run("extracts from application-name", func(t *testing.T) {
			htmlStr := `<html><head>
				<meta name="application-name" content="Tech Blog">
			</head><body></body></html>`
			doc := parseHTML(htmlStr)

			siteName := extractor.ExtractSiteName(doc)

			if siteName != "Tech Blog" {
				t.Errorf("Expected application-name, got %q", siteName)
			}
		})

		t.Run("returns empty for nil document", func(t *testing.T) {
			siteName := extractor.ExtractSiteName(nil)

			if siteName != "" {
				t.Errorf("Expected empty string for nil document, got %q", siteName)
			}
		})
	})

	t.Run("ExtractLanguage", func(t *testing.T) {
		extractor := NewMetadataExtractor()

		t.Run("extracts from html lang attribute", func(t *testing.T) {
			htmlStr := `<html lang="en-US"><body></body></html>`
			doc := parseHTML(htmlStr)

			lang := extractor.ExtractLanguage(doc)

			if lang != "en-US" {
				t.Errorf("Expected html lang attribute, got %q", lang)
			}
		})

		t.Run("extracts from OpenGraph locale", func(t *testing.T) {
			htmlStr := `<html><head>
				<meta property="og:locale" content="fr-FR">
			</head><body></body></html>`
			doc := parseHTML(htmlStr)

			lang := extractor.ExtractLanguage(doc)

			if lang != "fr-FR" {
				t.Errorf("Expected OpenGraph locale, got %q", lang)
			}
		})

		t.Run("returns empty for nil document", func(t *testing.T) {
			lang := extractor.ExtractLanguage(nil)

			if lang != "" {
				t.Errorf("Expected empty string for nil document, got %q", lang)
			}
		})
	})

	t.Run("getSchemaOrgField", func(t *testing.T) {
		extractor := NewMetadataExtractor()

		t.Run("extracts from JSON-LD Article", func(t *testing.T) {
			htmlStr := `<html><head>
				<script type="application/ld+json">
				{
					"@context": "https://schema.org",
					"@type": "Article",
					"headline": "Test Article",
					"author": "Test Author",
					"datePublished": "2025-01-15"
				}
				</script>
			</head><body></body></html>`
			doc := parseHTML(htmlStr)

			headline := extractor.getSchemaOrgField(doc, "headline")
			author := extractor.getSchemaOrgField(doc, "author")
			date := extractor.getSchemaOrgField(doc, "datePublished")

			if headline != "Test Article" {
				t.Errorf("Expected headline from JSON-LD, got %q", headline)
			}

			if author != "Test Author" {
				t.Errorf("Expected author from JSON-LD, got %q", author)
			}

			if date != "2025-01-15" {
				t.Errorf("Expected datePublished from JSON-LD, got %q", date)
			}
		})

		t.Run("extracts from NewsArticle type", func(t *testing.T) {
			htmlStr := `<html><head>
				<script type="application/ld+json">
				{
					"@context": "https://schema.org",
					"@type": "NewsArticle",
					"headline": "Breaking News"
				}
				</script>
			</head><body></body></html>`
			doc := parseHTML(htmlStr)

			headline := extractor.getSchemaOrgField(doc, "headline")

			if headline != "Breaking News" {
				t.Errorf("Expected headline from NewsArticle, got %q", headline)
			}
		})

		t.Run("handles nested author object", func(t *testing.T) {
			htmlStr := `<html><head>
				<script type="application/ld+json">
				{
					"@context": "https://schema.org",
					"@type": "Article",
					"author": {
						"@type": "Person",
						"name": "Nested Author"
					}
				}
				</script>
			</head><body></body></html>`
			doc := parseHTML(htmlStr)

			author := extractor.getSchemaOrgField(doc, "author")

			if author != "Nested Author" {
				t.Errorf("Expected nested author name, got %q", author)
			}
		})

		t.Run("returns empty for invalid JSON", func(t *testing.T) {
			htmlStr := `<html><head>
				<script type="application/ld+json">
				{ invalid json }
				</script>
			</head><body></body></html>`
			doc := parseHTML(htmlStr)

			result := extractor.getSchemaOrgField(doc, "headline")

			if result != "" {
				t.Errorf("Expected empty for invalid JSON, got %q", result)
			}
		})

		t.Run("returns empty for non-Article types", func(t *testing.T) {
			htmlStr := `<html><head>
				<script type="application/ld+json">
				{
					"@context": "https://schema.org",
					"@type": "WebPage",
					"headline": "Not an article"
				}
				</script>
			</head><body></body></html>`
			doc := parseHTML(htmlStr)

			result := extractor.getSchemaOrgField(doc, "headline")

			if result != "" {
				t.Errorf("Expected empty for WebPage type, got %q", result)
			}
		})
	})

	t.Run("ExtractMetadata", func(t *testing.T) {
		extractor := NewMetadataExtractor()

		t.Run("extracts all metadata fields", func(t *testing.T) {
			htmlStr := `<html lang="en"><head>
				<title>Full Article Title</title>
				<meta property="og:author" content="Full Name">
				<meta property="article:published_time" content="2025-01-20">
				<meta property="og:site_name" content="News Site">
			</head><body></body></html>`
			doc := parseHTML(htmlStr)

			result := extractor.ExtractMetadata(doc)

			if result == nil {
				t.Fatal("Expected result, got nil")
			}

			if !strings.Contains(result.Title, "Full Article Title") {
				t.Errorf("Expected title to be extracted, got %q", result.Title)
			}

			if result.Author != "Full Name" {
				t.Errorf("Expected author to be extracted, got %q", result.Author)
			}

			if result.PublishedDate != "2025-01-20" {
				t.Errorf("Expected date to be extracted, got %q", result.PublishedDate)
			}

			if result.SiteName != "News Site" {
				t.Errorf("Expected site name to be extracted, got %q", result.SiteName)
			}

			if result.Language != "en" {
				t.Errorf("Expected language to be extracted, got %q", result.Language)
			}
		})

		t.Run("returns empty result for nil document", func(t *testing.T) {
			result := extractor.ExtractMetadata(nil)

			if result == nil {
				t.Error("Expected empty result, got nil")
			}
		})
	})
}
