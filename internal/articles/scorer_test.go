package articles

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func parseHTML(htmlStr string) *html.Node {
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		return nil
	}
	return doc
}

func findElement(node *html.Node, tagName string) *html.Node {
	if node == nil {
		return nil
	}

	if node.Type == html.ElementNode && strings.EqualFold(node.Data, tagName) {
		return node
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if result := findElement(child, tagName); result != nil {
			return result
		}
	}

	return nil
}

func findElementWithClass(node *html.Node, className string) *html.Node {
	if node == nil {
		return nil
	}

	if node.Type == html.ElementNode {
		for _, attr := range node.Attr {
			if attr.Key == "class" && strings.Contains(attr.Val, className) {
				return node
			}
		}
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if result := findElementWithClass(child, className); result != nil {
			return result
		}
	}

	return nil
}

func TestScorer(t *testing.T) {
	t.Run("NewScorer", func(t *testing.T) {
		t.Run("creates scorer with default weights", func(t *testing.T) {
			scorer := NewScorer()

			if scorer == nil {
				t.Fatal("Expected scorer to be created, got nil")
			}

			if scorer.minContentLength != 140 {
				t.Errorf("Expected minContentLength 140, got %d", scorer.minContentLength)
			}

			if scorer.minScore != 20.0 {
				t.Errorf("Expected minScore 20.0, got %f", scorer.minScore)
			}
		})
	})

	t.Run("ScoreNode", func(t *testing.T) {
		scorer := NewScorer()

		t.Run("scores article tag highly", func(t *testing.T) {
			htmlStr := `<html><body><article class="main-content">Article content</article></body></html>`
			doc := parseHTML(htmlStr)
			article := findElement(doc, "article")

			score := scorer.ScoreNode(article)

			if score == nil {
				t.Fatal("Expected score, got nil")
			}

			if score.Score <= 0 {
				t.Errorf("Expected positive score for article tag, got %f", score.Score)
			}
		})

		t.Run("penalizes navigation elements", func(t *testing.T) {
			htmlStr := `<html><body><div class="navigation sidebar">Nav</div></body></html>`
			doc := parseHTML(htmlStr)
			nav := findElementWithClass(doc, "navigation")

			score := scorer.ScoreNode(nav)

			if score == nil {
				t.Fatal("Expected score, got nil")
			}

			if score.Score >= 0 {
				t.Errorf("Expected negative score for navigation, got %f", score.Score)
			}
		})

		t.Run("calculates text length", func(t *testing.T) {
			htmlStr := `<html><body><div>This is some test content with multiple words</div></body></html>`
			doc := parseHTML(htmlStr)
			div := findElement(doc, "div")

			score := scorer.ScoreNode(div)

			if score == nil {
				t.Fatal("Expected score, got nil")
			}

			if score.TextLength == 0 {
				t.Error("Expected non-zero text length")
			}
		})

		t.Run("returns nil for text nodes", func(t *testing.T) {
			textNode := &html.Node{Type: html.TextNode, Data: "text"}
			score := scorer.ScoreNode(textNode)

			if score != nil {
				t.Error("Expected nil score for text node")
			}
		})
	})

	t.Run("calculateLinkDensity", func(t *testing.T) {
		scorer := NewScorer()

		t.Run("calculates high link density", func(t *testing.T) {
			htmlStr := `<html><body><div><a href="#">link1</a> <a href="#">link2</a></div></body></html>`
			doc := parseHTML(htmlStr)
			div := findElement(doc, "div")

			density := scorer.calculateLinkDensity(div)

			if density < 0.5 {
				t.Errorf("Expected high link density (>0.5), got %f", density)
			}
		})

		t.Run("calculates low link density", func(t *testing.T) {
			htmlStr := `<html><body><div>Lots of regular text content here with just <a href="#">one link</a> in it</div></body></html>`
			doc := parseHTML(htmlStr)
			div := findElement(doc, "div")

			density := scorer.calculateLinkDensity(div)

			if density > 0.3 {
				t.Errorf("Expected low link density (<0.3), got %f", density)
			}
		})

		t.Run("returns zero for empty content", func(t *testing.T) {
			htmlStr := `<html><body><div></div></body></html>`
			doc := parseHTML(htmlStr)
			div := findElement(doc, "div")

			density := scorer.calculateLinkDensity(div)

			if density != 0.0 {
				t.Errorf("Expected zero density for empty content, got %f", density)
			}
		})
	})

	t.Run("getClassIdScore", func(t *testing.T) {
		scorer := NewScorer()

		t.Run("positive score for content class", func(t *testing.T) {
			node := &html.Node{
				Type: html.ElementNode,
				Data: "div",
				Attr: []html.Attribute{{Key: "class", Val: "article-content"}},
			}

			score := scorer.getClassIdScore(node)

			if score <= 0 {
				t.Errorf("Expected positive score for content class, got %f", score)
			}
		})

		t.Run("negative score for sidebar class", func(t *testing.T) {
			node := &html.Node{
				Type: html.ElementNode,
				Data: "div",
				Attr: []html.Attribute{{Key: "class", Val: "sidebar"}},
			}

			score := scorer.getClassIdScore(node)

			if score >= 0 {
				t.Errorf("Expected negative score for sidebar class, got %f", score)
			}
		})

		t.Run("strong negative score for banner", func(t *testing.T) {
			node := &html.Node{
				Type: html.ElementNode,
				Data: "div",
				Attr: []html.Attribute{{Key: "id", Val: "banner"}},
			}

			score := scorer.getClassIdScore(node)

			if score > -30 {
				t.Errorf("Expected strong negative score for banner, got %f", score)
			}
		})
	})

	t.Run("countParagraphs", func(t *testing.T) {
		scorer := NewScorer()

		t.Run("counts multiple paragraphs", func(t *testing.T) {
			htmlStr := `<html><body><div><p>First</p><p>Second</p><p>Third</p></div></body></html>`
			doc := parseHTML(htmlStr)
			div := findElement(doc, "div")

			count := scorer.countParagraphs(div)

			if count != 3 {
				t.Errorf("Expected 3 paragraphs, got %d", count)
			}
		})

		t.Run("returns zero for no paragraphs", func(t *testing.T) {
			htmlStr := `<html><body><div>Just text</div></body></html>`
			doc := parseHTML(htmlStr)
			div := findElement(doc, "div")

			count := scorer.countParagraphs(div)

			if count != 0 {
				t.Errorf("Expected 0 paragraphs, got %d", count)
			}
		})
	})

	t.Run("FindTopCandidates", func(t *testing.T) {
		scorer := NewScorer()

		t.Run("finds article with substantial content", func(t *testing.T) {
			htmlStr := `<html><body>
				<article class="main-content">
					<p>This is a long paragraph with substantial content that should score well in the readability algorithm.</p>
					<p>This is another paragraph with more content to increase the score.</p>
					<p>And a third paragraph to ensure we have enough text and structure.</p>
				</article>
				<aside class="sidebar">
					<a href="#">Link</a>
				</aside>
			</body></html>`
			doc := parseHTML(htmlStr)

			candidates := scorer.FindTopCandidates(doc, 5)

			if len(candidates) == 0 {
				t.Fatal("Expected to find candidates")
			}

			topScore := candidates[0]
			if topScore.Score <= 0 {
				t.Errorf("Expected positive score for top candidate, got %f", topScore.Score)
			}

			if topScore.ParagraphCount < 3 {
				t.Errorf("Expected top candidate to contain paragraphs, got %d", topScore.ParagraphCount)
			}
		})

		t.Run("filters out low-scoring nodes", func(t *testing.T) {
			htmlStr := `<html><body>
				<div class="ad">Short ad</div>
				<nav class="menu"><a href="#">Link</a></nav>
			</body></html>`
			doc := parseHTML(htmlStr)

			candidates := scorer.FindTopCandidates(doc, 5)

			for _, candidate := range candidates {
				if candidate.Score < scorer.minScore {
					t.Errorf("Expected all candidates to meet minimum score, got %f", candidate.Score)
				}
				if candidate.TextLength < scorer.minContentLength {
					t.Errorf("Expected all candidates to meet minimum length, got %d", candidate.TextLength)
				}
			}
		})

		t.Run("returns empty for nil root", func(t *testing.T) {
			candidates := scorer.FindTopCandidates(nil, 5)

			if candidates != nil {
				t.Error("Expected nil for nil root")
			}
		})
	})

	t.Run("calculateConfidence", func(t *testing.T) {
		scorer := NewScorer()

		t.Run("high confidence for good content", func(t *testing.T) {
			score := &ContentScore{
				Score:          60.0,
				TextLength:     500,
				LinkDensity:    0.1,
				ParagraphCount: 5,
			}

			confidence := scorer.calculateConfidence(score)

			if confidence < 0.5 {
				t.Errorf("Expected high confidence (>0.5) for good content, got %f", confidence)
			}

			if confidence > 1.0 {
				t.Errorf("Expected confidence <= 1.0, got %f", confidence)
			}
		})

		t.Run("low confidence for poor content", func(t *testing.T) {
			score := &ContentScore{
				Score:          10.0,
				TextLength:     50,
				LinkDensity:    0.8,
				ParagraphCount: 0,
			}

			confidence := scorer.calculateConfidence(score)

			if confidence > 0.3 {
				t.Errorf("Expected low confidence (<0.3) for poor content, got %f", confidence)
			}
		})

		t.Run("returns zero for nil score", func(t *testing.T) {
			confidence := scorer.calculateConfidence(nil)

			if confidence != 0.0 {
				t.Errorf("Expected 0.0 for nil score, got %f", confidence)
			}
		})
	})

	t.Run("IsProbablyReadable", func(t *testing.T) {
		scorer := NewScorer()

		t.Run("returns true for readable document", func(t *testing.T) {
			htmlStr := `<html><body>
				<article>
					<p>First paragraph with sufficient text content to be considered readable.</p>
					<p>Second paragraph with more text.</p>
					<p>Third paragraph with additional content.</p>
				</article>
			</body></html>`
			doc := parseHTML(htmlStr)

			readable := scorer.IsProbablyReadable(doc)

			if !readable {
				t.Error("Expected document to be readable")
			}
		})

		t.Run("returns false for short document", func(t *testing.T) {
			htmlStr := `<html><body><div>Short</div></body></html>`
			doc := parseHTML(htmlStr)

			readable := scorer.IsProbablyReadable(doc)

			if readable {
				t.Error("Expected document to not be readable")
			}
		})

		t.Run("returns false for nil document", func(t *testing.T) {
			readable := scorer.IsProbablyReadable(nil)

			if readable {
				t.Error("Expected nil document to not be readable")
			}
		})
	})

	t.Run("ScoreAncestors", func(t *testing.T) {
		scorer := NewScorer()

		t.Run("propagates score to parent nodes", func(t *testing.T) {
			htmlStr := `<html><body><div><article><p>Content</p></article></div></body></html>`
			doc := parseHTML(htmlStr)
			p := findElement(doc, "p")

			scores := make(map[*html.Node]*ContentScore)
			scores[p] = &ContentScore{Node: p, Score: 10.0}

			scorer.ScoreAncestors(scores, p, 100.0)

			article := findElement(doc, "article")
			if scores[article] == nil {
				t.Error("Expected article to receive propagated score")
			}

			if scores[article].Score <= 0 {
				t.Errorf("Expected positive propagated score, got %f", scores[article].Score)
			}
		})
	})
}
