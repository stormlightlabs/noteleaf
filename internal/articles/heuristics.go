package articles

import (
	"strings"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

// ExtractionResult contains the results of heuristic content extraction.
type ExtractionResult struct {
	Content          string
	Title            string
	Author           string
	PublishedDate    string
	SiteName         string
	Language         string
	Confidence       float64
	ExtractionMethod string // "heuristic" or "xpath" or "dual"
}

// HeuristicExtractor implements Readability-style content extraction.
type HeuristicExtractor struct {
	scorer *Scorer
}

// NewHeuristicExtractor creates a new extractor with default scoring.
func NewHeuristicExtractor() *HeuristicExtractor {
	return &HeuristicExtractor{
		scorer: NewScorer(),
	}
}

// ExtractContent performs heuristic-based content extraction from an HTML document.
func (e *HeuristicExtractor) ExtractContent(doc *html.Node) *ExtractionResult {
	if doc == nil {
		return nil
	}

	if !e.scorer.IsProbablyReadable(doc) {
		return &ExtractionResult{
			Confidence:       0.1,
			ExtractionMethod: "heuristic",
		}
	}

	cleaned := e.cleanDocument(doc)
	candidates := e.scorer.FindTopCandidates(cleaned, 5)
	if len(candidates) == 0 {
		return &ExtractionResult{
			Confidence:       0.2,
			ExtractionMethod: "heuristic",
		}
	}

	topCandidate := candidates[0]
	content := e.extractTextContent(topCandidate.Node)
	result := &ExtractionResult{
		Content:          content,
		Confidence:       topCandidate.ConfidenceLevel,
		ExtractionMethod: "heuristic",
	}

	return result
}

// cleanDocument removes unwanted elements and prepares the document for extraction.
func (e *HeuristicExtractor) cleanDocument(doc *html.Node) *html.Node {

	cloned := e.cloneNode(doc)

	e.removeElements(cloned, "script", "style", "noscript", "iframe", "embed", "object")
	e.removeHiddenElements(cloned)
	e.removeUnlikelyCandidates(cloned)
	e.removeHighLinkDensityElements(cloned)

	return cloned
}

// cloneNode creates a deep copy of an HTML node tree.
func (e *HeuristicExtractor) cloneNode(node *html.Node) *html.Node {
	if node == nil {
		return nil
	}

	clone := &html.Node{
		Type:      node.Type,
		Data:      node.Data,
		DataAtom:  node.DataAtom,
		Namespace: node.Namespace,
		Attr:      make([]html.Attribute, len(node.Attr)),
	}

	copy(clone.Attr, node.Attr)

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		clonedChild := e.cloneNode(child)
		if clonedChild != nil {
			clone.AppendChild(clonedChild)
		}
	}

	return clone
}

// removeElements removes all elements with the specified tag names.
func (e *HeuristicExtractor) removeElements(root *html.Node, tagNames ...string) {
	if root == nil {
		return
	}

	tagMap := make(map[string]bool)
	for _, tag := range tagNames {
		tagMap[strings.ToLower(tag)] = true
	}

	var toRemove []*html.Node

	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode {
			if tagMap[strings.ToLower(node.Data)] {
				toRemove = append(toRemove, node)
				return
			}
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}

	walk(root)

	for _, node := range toRemove {
		if node.Parent != nil {
			node.Parent.RemoveChild(node)
		}
	}
}

// removeHiddenElements removes elements that are hidden via CSS or attributes.
func (e *HeuristicExtractor) removeHiddenElements(root *html.Node) {
	if root == nil {
		return
	}

	var toRemove []*html.Node

	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode {
			for _, attr := range node.Attr {
				if attr.Key == "hidden" {
					toRemove = append(toRemove, node)
					return
				}

				if attr.Key == "style" {
					style := strings.ToLower(attr.Val)
					if strings.Contains(style, "display:none") || strings.Contains(style, "display: none") ||
						strings.Contains(style, "visibility:hidden") || strings.Contains(style, "visibility: hidden") {
						toRemove = append(toRemove, node)
						return
					}
				}

				if attr.Key == "aria-hidden" && attr.Val == "true" {
					toRemove = append(toRemove, node)
					return
				}
			}
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}

	walk(root)

	for _, node := range toRemove {
		if node.Parent != nil {
			node.Parent.RemoveChild(node)
		}
	}
}

// removeUnlikelyCandidates removes elements that are unlikely to be main content.
func (e *HeuristicExtractor) removeUnlikelyCandidates(root *html.Node) {
	if root == nil {
		return
	}

	var toRemove []*html.Node

	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode {
			score := e.scorer.getClassIdScore(node)

			if score < -40 {
				toRemove = append(toRemove, node)
				return
			}
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}

	walk(root)

	for _, node := range toRemove {
		if node.Parent != nil {
			node.Parent.RemoveChild(node)
		}
	}
}

// removeHighLinkDensityElements removes elements with excessive link density.
func (e *HeuristicExtractor) removeHighLinkDensityElements(root *html.Node) {
	if root == nil {
		return
	}

	const linkDensityThreshold = 0.75

	var toRemove []*html.Node

	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode {
			if strings.ToLower(node.Data) == "a" {
				for child := node.FirstChild; child != nil; child = child.NextSibling {
					walk(child)
				}
				return
			}

			density := e.scorer.calculateLinkDensity(node)
			textLen := e.scorer.calculateTextLength(node)

			if density > linkDensityThreshold && textLen < 500 {
				toRemove = append(toRemove, node)
				return
			}
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}

	walk(root)

	for _, node := range toRemove {
		if node.Parent != nil {
			node.Parent.RemoveChild(node)
		}
	}
}

// extractTextContent extracts cleaned text from a node.
func (e *HeuristicExtractor) extractTextContent(node *html.Node) string {
	if node == nil {
		return ""
	}

	var buf strings.Builder
	e.extractTextRecursive(node, &buf)

	text := buf.String()
	text = normalizeWhitespace(text)
	text = strings.TrimSpace(text)

	return text
}

// extractTextRecursive recursively extracts text with basic formatting.
func (e *HeuristicExtractor) extractTextRecursive(node *html.Node, buf *strings.Builder) {
	if node == nil {
		return
	}

	if node.Type == html.TextNode {
		buf.WriteString(node.Data)
		return
	}

	if node.Type == html.ElementNode {
		tag := strings.ToLower(node.Data)

		if e.isBlockElement(tag) && buf.Len() > 0 {
			buf.WriteString("\n\n")
		}

		if tag == "li" {
			buf.WriteString("\n• ")
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			e.extractTextRecursive(child, buf)
		}

		if e.isBlockElement(tag) {
			buf.WriteString("\n")
		}
	}
}

// isBlockElement returns true for block-level HTML elements.
func (e *HeuristicExtractor) isBlockElement(tagName string) bool {
	blockElements := map[string]bool{
		"p":          true,
		"div":        true,
		"article":    true,
		"section":    true,
		"h1":         true,
		"h2":         true,
		"h3":         true,
		"h4":         true,
		"h5":         true,
		"h6":         true,
		"blockquote": true,
		"pre":        true,
		"ul":         true,
		"ol":         true,
		"table":      true,
		"tr":         true,
		"td":         true,
		"th":         true,
	}

	return blockElements[tagName]
}

// CompareWithXPath compares heuristic extraction with XPath-based extraction.
func (e *HeuristicExtractor) CompareWithXPath(doc *html.Node, xpathNode *html.Node) *ExtractionResult {
	if doc == nil {
		return nil
	}

	heuristicResult := e.ExtractContent(doc)
	if heuristicResult == nil {
		heuristicResult = &ExtractionResult{
			ExtractionMethod: "heuristic",
			Confidence:       0.0,
		}
	}

	if xpathNode == nil {
		return heuristicResult
	}

	xpathContent := e.extractTextContent(xpathNode)
	xpathLen := len(xpathContent)
	heuristicLen := len(heuristicResult.Content)

	similarity := e.calculateSimilarity(xpathContent, heuristicResult.Content)

	if similarity > 0.8 {
		heuristicResult.Confidence = 0.95
		heuristicResult.ExtractionMethod = "dual-validated"
		return heuristicResult
	} else if float64(xpathLen) > float64(heuristicLen)*1.5 {
		return &ExtractionResult{
			Content:          xpathContent,
			Confidence:       0.85,
			ExtractionMethod: "xpath-preferred",
		}
	} else if float64(heuristicLen) > float64(xpathLen)*1.5 {
		heuristicResult.Confidence = 0.80
		heuristicResult.ExtractionMethod = "heuristic-preferred"
		return heuristicResult
	} else {
		heuristicResult.Confidence = 0.70
		heuristicResult.ExtractionMethod = "heuristic-fallback"
		return heuristicResult
	}
}

// calculateSimilarity estimates content similarity (simple ratio of common words).
func (e *HeuristicExtractor) calculateSimilarity(text1, text2 string) float64 {
	if len(text1) == 0 || len(text2) == 0 {
		if len(text1) == 0 && len(text2) == 0 {
			return 1.0
		}
		return 0.0
	}

	words1 := strings.Fields(strings.ToLower(text1))
	words2 := strings.Fields(strings.ToLower(text2))

	if len(words1) == 0 || len(words2) == 0 {
		return 0.0
	}

	freq1 := make(map[string]int)
	freq2 := make(map[string]int)

	for _, word := range words1 {
		freq1[word]++
	}

	for _, word := range words2 {
		freq2[word]++
	}

	common := 0
	for word := range freq1 {
		if freq2[word] > 0 {
			common++
		}
	}

	union := len(freq1) + len(freq2) - common
	if union == 0 {
		return 0.0
	}

	return float64(common) / float64(union)
}

// ExtractWithSemanticHTML attempts extraction using semantic HTML5 elements first.
// Falls back to heuristic scoring if semantic elements aren't found.
func (e *HeuristicExtractor) ExtractWithSemanticHTML(doc *html.Node) *ExtractionResult {
	if doc == nil {
		return nil
	}

	articleNode := htmlquery.FindOne(doc, "//article")
	if articleNode != nil {
		content := e.extractTextContent(articleNode)
		if len(content) > e.scorer.minContentLength {
			return &ExtractionResult{
				Content:          content,
				Confidence:       0.90,
				ExtractionMethod: "semantic-html",
			}
		}
	}

	mainNode := htmlquery.FindOne(doc, "//main")
	if mainNode != nil {
		content := e.extractTextContent(mainNode)
		if len(content) > e.scorer.minContentLength {
			return &ExtractionResult{
				Content:          content,
				Confidence:       0.88,
				ExtractionMethod: "semantic-html",
			}
		}
	}

	return e.ExtractContent(doc)
}
