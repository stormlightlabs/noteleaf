package articles

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestHeuristicExtractor(t *testing.T) {
	t.Run("NewHeuristicExtractor", func(t *testing.T) {
		t.Run("creates extractor with scorer", func(t *testing.T) {
			extractor := NewHeuristicExtractor()

			if extractor == nil {
				t.Fatal("Expected extractor to be created, got nil")
			}

			if extractor.scorer == nil {
				t.Error("Expected extractor to have scorer")
			}
		})
	})

	t.Run("ExtractContent", func(t *testing.T) {
		extractor := NewHeuristicExtractor()

		t.Run("extracts content from article", func(t *testing.T) {
			htmlStr := `<html><body>
				<article class="main-content">
					<p>This is the first paragraph of the article with substantial content.</p>
					<p>This is the second paragraph with more information and details.</p>
					<p>And this is the third paragraph to ensure sufficient content.</p>
				</article>
				<aside class="sidebar"><a href="#">Sidebar link</a></aside>
			</body></html>`
			doc := parseHTML(htmlStr)

			result := extractor.ExtractContent(doc)

			if result == nil {
				t.Fatal("Expected extraction result, got nil")
			}

			if result.Content == "" {
				t.Error("Expected content to be extracted")
			}

			if result.Confidence == 0.0 {
				t.Error("Expected non-zero confidence")
			}

			if !strings.Contains(result.Content, "first paragraph") {
				t.Error("Expected content to contain article text")
			}
		})

		t.Run("returns low confidence for unreadable document", func(t *testing.T) {
			htmlStr := `<html><body><div>Short</div></body></html>`
			doc := parseHTML(htmlStr)

			result := extractor.ExtractContent(doc)

			if result == nil {
				t.Fatal("Expected extraction result, got nil")
			}

			if result.Confidence > 0.3 {
				t.Errorf("Expected low confidence for short document, got %f", result.Confidence)
			}
		})

		t.Run("returns nil for nil document", func(t *testing.T) {
			result := extractor.ExtractContent(nil)

			if result != nil {
				t.Error("Expected nil for nil document")
			}
		})
	})

	t.Run("cleanDocument", func(t *testing.T) {
		extractor := NewHeuristicExtractor()

		t.Run("removes script and style tags", func(t *testing.T) {
			htmlStr := `<html><body>
				<script>alert('test');</script>
				<style>.test { color: red; }</style>
				<p>Content</p>
			</body></html>`
			doc := parseHTML(htmlStr)

			cleaned := extractor.cleanDocument(doc)

			script := findElement(cleaned, "script")
			style := findElement(cleaned, "style")

			if script != nil {
				t.Error("Expected script tag to be removed")
			}

			if style != nil {
				t.Error("Expected style tag to be removed")
			}
		})

		t.Run("removes hidden elements", func(t *testing.T) {
			htmlStr := `<html><body>
				<div style="display:none">Hidden</div>
				<div hidden>Also hidden</div>
				<p>Visible</p>
			</body></html>`
			doc := parseHTML(htmlStr)

			cleaned := extractor.cleanDocument(doc)

			// Count divs - should only have visible ones
			divCount := 0
			var countDivs func(*html.Node)
			countDivs = func(node *html.Node) {
				if node.Type == html.ElementNode && node.Data == "div" {
					divCount++
				}
				for child := node.FirstChild; child != nil; child = child.NextSibling {
					countDivs(child)
				}
			}
			countDivs(cleaned)

			if divCount > 0 {
				t.Errorf("Expected hidden divs to be removed, found %d", divCount)
			}
		})

		t.Run("removes high link density elements", func(t *testing.T) {
			htmlStr := `<html><body>
				<div class="links">
					<a href="#">Link1</a>
					<a href="#">Link2</a>
					<a href="#">Link3</a>
				</div>
				<p>Regular paragraph with actual content that should remain.</p>
			</body></html>`
			doc := parseHTML(htmlStr)

			cleaned := extractor.cleanDocument(doc)

			p := findElement(cleaned, "p")
			if p == nil {
				t.Error("Expected paragraph to remain")
			}
		})
	})

	t.Run("extractTextContent", func(t *testing.T) {
		extractor := NewHeuristicExtractor()

		t.Run("extracts text with basic formatting", func(t *testing.T) {
			htmlStr := `<html><body><div>
				<p>First paragraph</p>
				<p>Second paragraph</p>
			</div></body></html>`
			doc := parseHTML(htmlStr)
			div := findElement(doc, "div")

			text := extractor.extractTextContent(div)

			if !strings.Contains(text, "First paragraph") {
				t.Error("Expected text to contain first paragraph")
			}

			if !strings.Contains(text, "Second paragraph") {
				t.Error("Expected text to contain second paragraph")
			}
		})

		t.Run("formats list items with bullets", func(t *testing.T) {
			htmlStr := `<html><body><ul>
				<li>Item 1</li>
				<li>Item 2</li>
			</ul></body></html>`
			doc := parseHTML(htmlStr)
			ul := findElement(doc, "ul")

			text := extractor.extractTextContent(ul)

			if !strings.Contains(text, "•") {
				t.Error("Expected text to contain bullet points")
			}
		})

		t.Run("returns empty string for nil node", func(t *testing.T) {
			text := extractor.extractTextContent(nil)

			if text != "" {
				t.Error("Expected empty string for nil node")
			}
		})
	})

	t.Run("CompareWithXPath", func(t *testing.T) {
		extractor := NewHeuristicExtractor()

		t.Run("high confidence when XPath and heuristics agree", func(t *testing.T) {
			htmlStr := `<html><body>
				<article>
					<p>This is substantial content that both methods should find.</p>
					<p>Another paragraph with more details and information.</p>
					<p>And a third paragraph for good measure and completeness.</p>
				</article>
			</body></html>`
			doc := parseHTML(htmlStr)
			article := findElement(doc, "article")

			result := extractor.CompareWithXPath(doc, article)

			if result == nil {
				t.Fatal("Expected result, got nil")
			}

			if result.Confidence < 0.8 {
				t.Errorf("Expected high confidence when methods agree, got %f", result.Confidence)
			}

			if !strings.Contains(result.ExtractionMethod, "dual") && !strings.Contains(result.ExtractionMethod, "validated") {
				t.Errorf("Expected dual validation method, got %s", result.ExtractionMethod)
			}
		})

		t.Run("prefers XPath when it extracts more content", func(t *testing.T) {
			htmlStr := `<html><body>
				<div class="content">
					<p>Short content</p>
				</div>
				<div class="more">
					<p>This is additional content that XPath found but heuristics might miss.</p>
					<p>Even more content here to make a significant difference in length.</p>
					<p>And yet another paragraph to ensure XPath extraction is substantially longer.</p>
				</div>
			</body></html>`
			doc := parseHTML(htmlStr)

			// XPath would get more content
			body := findElement(doc, "body")

			result := extractor.CompareWithXPath(doc, body)

			if result == nil {
				t.Fatal("Expected result, got nil")
			}

			// Should prefer one method over the other
			if result.ExtractionMethod == "heuristic" {
				t.Errorf("Expected method preference, got %s", result.ExtractionMethod)
			}
		})

		t.Run("uses heuristics when XPath node is nil", func(t *testing.T) {
			htmlStr := `<html><body>
				<article>
					<p>Content that heuristics should find on its own.</p>
					<p>Additional paragraph for sufficient content length.</p>
					<p>Third paragraph to meet minimum requirements.</p>
				</article>
			</body></html>`
			doc := parseHTML(htmlStr)

			result := extractor.CompareWithXPath(doc, nil)

			if result == nil {
				t.Fatal("Expected result, got nil")
			}

			if result.ExtractionMethod != "heuristic" {
				t.Errorf("Expected heuristic method when XPath is nil, got %s", result.ExtractionMethod)
			}
		})

		t.Run("returns nil for nil document", func(t *testing.T) {
			result := extractor.CompareWithXPath(nil, nil)

			if result != nil {
				t.Error("Expected nil for nil document")
			}
		})
	})

	t.Run("calculateSimilarity", func(t *testing.T) {
		extractor := NewHeuristicExtractor()

		t.Run("returns high similarity for identical text", func(t *testing.T) {
			text := "This is some test content"

			similarity := extractor.calculateSimilarity(text, text)

			if similarity < 0.9 {
				t.Errorf("Expected high similarity for identical text, got %f", similarity)
			}
		})

		t.Run("returns low similarity for different text", func(t *testing.T) {
			text1 := "This is the first piece of content"
			text2 := "Completely different words and phrases"

			similarity := extractor.calculateSimilarity(text1, text2)

			if similarity > 0.3 {
				t.Errorf("Expected low similarity for different text, got %f", similarity)
			}
		})

		t.Run("returns zero for empty strings", func(t *testing.T) {
			similarity := extractor.calculateSimilarity("text", "")

			if similarity != 0.0 {
				t.Errorf("Expected zero similarity for empty string, got %f", similarity)
			}
		})

		t.Run("returns one for both empty", func(t *testing.T) {
			similarity := extractor.calculateSimilarity("", "")

			if similarity != 1.0 {
				t.Errorf("Expected 1.0 similarity for both empty, got %f", similarity)
			}
		})
	})

	t.Run("ExtractWithSemanticHTML", func(t *testing.T) {
		extractor := NewHeuristicExtractor()

		t.Run("extracts from article tag", func(t *testing.T) {
			htmlStr := `<html><body>
				<nav>Navigation</nav>
				<article>
					<p>This is the main article content that should be extracted.</p>
					<p>Second paragraph of the article with more information.</p>
					<p>Third paragraph to provide sufficient content length.</p>
				</article>
				<aside>Sidebar</aside>
			</body></html>`
			doc := parseHTML(htmlStr)

			result := extractor.ExtractWithSemanticHTML(doc)

			if result == nil {
				t.Fatal("Expected result, got nil")
			}

			if result.ExtractionMethod != "semantic-html" {
				t.Errorf("Expected semantic-html method, got %s", result.ExtractionMethod)
			}

			if !strings.Contains(result.Content, "main article content") {
				t.Error("Expected content from article tag")
			}

			if result.Confidence < 0.85 {
				t.Errorf("Expected high confidence for semantic HTML, got %f", result.Confidence)
			}
		})

		t.Run("extracts from main tag", func(t *testing.T) {
			htmlStr := `<html><body>
				<header>Header</header>
				<main>
					<p>This is the main content area with sufficient text.</p>
					<p>Additional content paragraph with more details.</p>
					<p>Third paragraph for completeness and length.</p>
				</main>
				<footer>Footer</footer>
			</body></html>`
			doc := parseHTML(htmlStr)

			result := extractor.ExtractWithSemanticHTML(doc)

			if result == nil {
				t.Fatal("Expected result, got nil")
			}

			if result.ExtractionMethod != "semantic-html" {
				t.Errorf("Expected semantic-html method, got %s", result.ExtractionMethod)
			}

			if !strings.Contains(result.Content, "main content area") {
				t.Error("Expected content from main tag")
			}
		})

		t.Run("falls back to heuristics without semantic tags", func(t *testing.T) {
			htmlStr := `<html><body>
				<div class="content">
					<p>Content in a regular div without semantic HTML tags.</p>
					<p>Second paragraph with additional information.</p>
					<p>Third paragraph for sufficient content.</p>
				</div>
			</body></html>`
			doc := parseHTML(htmlStr)

			result := extractor.ExtractWithSemanticHTML(doc)

			if result == nil {
				t.Fatal("Expected result, got nil")
			}

			if result.ExtractionMethod == "semantic-html" {
				t.Error("Should not use semantic-html method without semantic tags")
			}
		})

		t.Run("returns nil for nil document", func(t *testing.T) {
			result := extractor.ExtractWithSemanticHTML(nil)

			if result != nil {
				t.Error("Expected nil for nil document")
			}
		})
	})

	t.Run("isBlockElement", func(t *testing.T) {
		extractor := NewHeuristicExtractor()

		t.Run("identifies block elements", func(t *testing.T) {
			blockTags := []string{"p", "div", "article", "h1", "section"}

			for _, tag := range blockTags {
				if !extractor.isBlockElement(tag) {
					t.Errorf("Expected %s to be a block element", tag)
				}
			}
		})

		t.Run("identifies non-block elements", func(t *testing.T) {
			inlineTags := []string{"span", "a", "em", "strong", "code"}

			for _, tag := range inlineTags {
				if extractor.isBlockElement(tag) {
					t.Errorf("Expected %s to not be a block element", tag)
				}
			}
		})
	})
}
