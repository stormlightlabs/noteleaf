package articles

import (
	"math"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// ContentScore represents the score and metadata for a content node.
type ContentScore struct {
	Node            *html.Node
	Score           float64
	TextLength      int
	LinkDensity     float64
	ParagraphCount  int
	AncestorDepth   int
	ConfidenceLevel float64
}

// Scorer implements Readability-style heuristic scoring for content extraction.
type Scorer struct {
	linkDensityWeight   float64
	classWeightPositive float64
	classWeightNegative float64
	paragraphWeight     float64
	ancestorDecayFactor float64
	minContentLength    int
	minScore            float64
	positivePattern     *regexp.Regexp
	negativePattern     *regexp.Regexp
	unlikelyPattern     *regexp.Regexp
}

// NewScorer creates a new Scorer with default Readability.js-inspired weights.
func NewScorer() *Scorer {
	return &Scorer{
		linkDensityWeight:   -1.0,
		classWeightPositive: 25.0,
		classWeightNegative: -25.0,
		paragraphWeight:     1.0,
		ancestorDecayFactor: 0.5,
		minContentLength:    140,
		minScore:            20.0,

		positivePattern: regexp.MustCompile(`(?i)(article|body|content|entry|hentry|h-entry|main|page|pagination|post|text|blog|story)`),
		negativePattern: regexp.MustCompile(`(?i)(combx|comment|com-|contact|foot|footer|footnote|masthead|media|meta|outbrain|promo|related|scroll|share|shoutbox|sidebar|skyscraper|sponsor|shopping|tags|tool|widget|ad-|advertisement|breadcrumb|hidden|nav|menu|header)`),
		unlikelyPattern: regexp.MustCompile(`(?i)(banner|cookie|popup|modal)`),
	}
}

// ScoreNode calculates a content score for the given node based on multiple heuristics.
// This implements the core Readability scoring algorithm.
func (s *Scorer) ScoreNode(node *html.Node) *ContentScore {
	if node == nil || node.Type != html.ElementNode {
		return nil
	}

	score := &ContentScore{
		Node:          node,
		Score:         0.0,
		AncestorDepth: s.calculateDepth(node),
	}

	score.Score = s.getTagScore(node.Data)
	score.Score += s.getClassIdScore(node)

	score.TextLength = s.calculateTextLength(node)
	score.LinkDensity = s.calculateLinkDensity(node)
	score.ParagraphCount = s.countParagraphs(node)

	score.Score += score.LinkDensity * s.linkDensityWeight
	score.Score += float64(score.ParagraphCount) * s.paragraphWeight
	score.Score += s.getTextLengthScore(score.TextLength)

	score.ConfidenceLevel = s.calculateConfidence(score)
	return score
}

// getTagScore returns a base score based on the HTML tag type.
// Some tags are more likely to contain main content than others.
func (s *Scorer) getTagScore(tagName string) float64 {
	switch strings.ToLower(tagName) {
	case "article":
		return 30.0
	case "section":
		return 15.0
	case "div":
		return 5.0
	case "main":
		return 40.0
	case "p":
		return 3.0
	case "pre", "td", "blockquote":
		return 3.0
	case "address", "ol", "ul", "dl", "dd", "dt", "li", "form":
		return -3.0
	case "h1", "h2", "h3", "h4", "h5", "h6", "th":
		return -5.0
	default:
		return 0.0
	}
}

// getClassIdScore analyzes class and ID attributes for positive/negative indicators.
// Returns a positive score for content-like names, negative for navigation/ads.
func (s *Scorer) getClassIdScore(node *html.Node) float64 {
	score := 0.0
	classID := s.getClassAndID(node)

	if classID == "" {
		return 0.0
	}

	if s.unlikelyPattern.MatchString(classID) {
		return -50.0
	}

	if s.negativePattern.MatchString(classID) {
		score += s.classWeightNegative
	}

	if s.positivePattern.MatchString(classID) {
		score += s.classWeightPositive
	}

	return score
}

// getClassAndID concatenates class and ID attributes for pattern matching.
func (s *Scorer) getClassAndID(node *html.Node) string {
	var parts []string

	for _, attr := range node.Attr {
		if attr.Key == "class" || attr.Key == "id" {
			parts = append(parts, attr.Val)
		}
	}

	return strings.Join(parts, " ")
}

// calculateTextLength returns the total text length within the node.
func (s *Scorer) calculateTextLength(node *html.Node) int {
	text := s.getInnerText(node)
	return len(strings.TrimSpace(text))
}

// calculateLinkDensity calculates the ratio of link text to total text.
// Higher link density indicates navigation or related links, not main content.
func (s *Scorer) calculateLinkDensity(node *html.Node) float64 {
	totalText := s.getInnerText(node)
	linkText := s.getLinkText(node)

	totalLen := len(strings.TrimSpace(totalText))
	linkLen := len(strings.TrimSpace(linkText))

	if totalLen == 0 {
		return 0.0
	}

	return float64(linkLen) / float64(totalLen)
}

// getInnerText extracts all text content from a node and its descendants.
func (s *Scorer) getInnerText(node *html.Node) string {
	var buf strings.Builder
	s.extractText(node, &buf)
	return buf.String()
}

// extractText recursively extracts text from a node tree.
func (s *Scorer) extractText(node *html.Node, buf *strings.Builder) {
	if node == nil {
		return
	}

	if node.Type == html.TextNode {
		buf.WriteString(node.Data)
		buf.WriteString(" ")
		return
	}

	if node.Type == html.ElementNode {
		tag := strings.ToLower(node.Data)
		if tag == "script" || tag == "style" || tag == "noscript" {
			return
		}
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		s.extractText(child, buf)
	}
}

// getLinkText extracts text from anchor tags only.
func (s *Scorer) getLinkText(node *html.Node) string {
	var buf strings.Builder
	s.extractLinkText(node, &buf)
	return buf.String()
}

// extractLinkText recursively extracts text from anchor tags.
func (s *Scorer) extractLinkText(node *html.Node, buf *strings.Builder) {
	if node == nil {
		return
	}

	if node.Type == html.ElementNode && strings.ToLower(node.Data) == "a" {
		s.extractText(node, buf)
		return
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		s.extractLinkText(child, buf)
	}
}

// countParagraphs counts paragraph elements within the node.
func (s *Scorer) countParagraphs(node *html.Node) int {
	count := 0
	s.walkParagraphs(node, &count)
	return count
}

// walkParagraphs recursively counts paragraph elements.
func (s *Scorer) walkParagraphs(node *html.Node, count *int) {
	if node == nil {
		return
	}

	if node.Type == html.ElementNode && strings.ToLower(node.Data) == "p" {
		*count++
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		s.walkParagraphs(child, count)
	}
}

// getTextLengthScore provides a bonus for nodes with substantial text content.
func (s *Scorer) getTextLengthScore(textLen int) float64 {
	if textLen < 25 {
		return 0.0
	}
	return math.Log10(float64(textLen)) * 2.0
}

// calculateDepth calculates how deep in the DOM tree this node is.
func (s *Scorer) calculateDepth(node *html.Node) int {
	depth := 0
	for n := node.Parent; n != nil; n = n.Parent {
		depth++
	}
	return depth
}

// ScoreAncestors propagates scores up the DOM tree with decay.
// This implements the Readability algorithm's ancestor scoring.
func (s *Scorer) ScoreAncestors(scores map[*html.Node]*ContentScore, node *html.Node, baseScore float64) {
	if node == nil || baseScore <= 0 {
		return
	}

	currentScore := baseScore
	level := 0

	for parent := node.Parent; parent != nil && level < 5; parent = parent.Parent {
		if parent.Type != html.ElementNode {
			continue
		}

		if _, exists := scores[parent]; !exists {
			scores[parent] = s.ScoreNode(parent)
			if scores[parent] == nil {
				continue
			}
		}

		decayedScore := currentScore * math.Pow(s.ancestorDecayFactor, float64(level+1))
		scores[parent].Score += decayedScore
		level++
	}
}

// FindTopCandidates identifies the N highest-scoring content candidates.
func (s *Scorer) FindTopCandidates(root *html.Node, n int) []*ContentScore {
	if root == nil || n <= 0 {
		return nil
	}

	scores := make(map[*html.Node]*ContentScore)
	s.scoreTree(root, scores)

	var candidates []*ContentScore
	for _, score := range scores {
		if score.Score >= s.minScore && score.TextLength >= s.minContentLength {
			candidates = append(candidates, score)
		}
	}

	for i := 0; i < len(candidates); i++ {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[j].Score > candidates[i].Score {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}

	if len(candidates) > n {
		candidates = candidates[:n]
	}

	return candidates
}

// scoreTree recursively scores all nodes in the tree.
func (s *Scorer) scoreTree(node *html.Node, scores map[*html.Node]*ContentScore) {
	if node == nil {
		return
	}

	if node.Type == html.ElementNode {
		tag := strings.ToLower(node.Data)
		if tag != "script" && tag != "style" && tag != "noscript" {
			score := s.ScoreNode(node)
			if score != nil && score.Score > 0 {
				scores[node] = score
				s.ScoreAncestors(scores, node, score.Score)
			}
		}
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		s.scoreTree(child, scores)
	}
}

// calculateConfidence estimates how confident we are in this content selection (between 0 & 1).
func (s *Scorer) calculateConfidence(score *ContentScore) float64 {
	if score == nil {
		return 0.0
	}

	confidence := 0.0

	if score.Score > s.minScore*2 {
		confidence += 0.3
	} else if score.Score > s.minScore {
		confidence += 0.15
	}

	if score.TextLength > s.minContentLength*3 {
		confidence += 0.3
	} else if score.TextLength > s.minContentLength {
		confidence += 0.15
	}

	if score.LinkDensity < 0.2 {
		confidence += 0.2
	} else if score.LinkDensity < 0.4 {
		confidence += 0.1
	}

	if score.ParagraphCount >= 3 {
		confidence += 0.2
	} else if score.ParagraphCount >= 1 {
		confidence += 0.1
	}

	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// IsProbablyReadable determines if a document is likely to have extractable content.
// This is inspired by Readability.js's isProbablyReaderable function.
func (s *Scorer) IsProbablyReadable(doc *html.Node) bool {
	if doc == nil {
		return false
	}

	paragraphCount := s.countParagraphs(doc)
	textLength := s.calculateTextLength(doc)
	return paragraphCount >= 3 && textLength >= s.minContentLength
}
