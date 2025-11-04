package articles

import (
	"bufio"
	"embed"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/stormlightlabs/noteleaf/internal/models"
	exhtml "golang.org/x/net/html"
)

//go:embed rules/*.txt
var rulesFS embed.FS

// ParsedContent represents the extracted content from a web page
type ParsedContent struct {
	Title            string
	Author           string
	Date             string
	Content          string
	URL              string
	Confidence       float64 // 0-1 scale, confidence in extraction quality
	ExtractionMethod string  // "xpath", "heuristic", "dual-validated", etc.
}

// ParsingRule represents XPath rules for extracting content from a specific domain
type ParsingRule struct {
	Domain            string
	Title             string
	Author            string
	Date              string
	Body              string
	Strip             []string // XPath selectors for elements to remove
	StripIDsOrClasses []string
	TestURLs          []string
	Headers           map[string]string
	Prune             bool
	Tidy              bool
}

// Parser interface defines methods for parsing articles from URLs
type Parser interface {
	// ParseURL extracts article content from a given URL
	ParseURL(url string) (*ParsedContent, error)
	// Convert HTML content directly to markdown using domain-specific rules
	Convert(htmlContent, domain, sourceURL string) (string, error)
	// GetSupportedDomains returns a list of domains that have parsing rules
	GetSupportedDomains() []string
	// SaveArticle saves the parsed content to filesystem and returns file paths
	SaveArticle(content *ParsedContent, storageDir string) (markdownPath, htmlPath string, err error)
}

// ArticleParser implements the Parser interface
type ArticleParser struct {
	rules             map[string]*ParsingRule
	client            *http.Client
	heuristicExtract  *HeuristicExtractor
	metadataExtractor *MetadataExtractor
}

// NewArticleParser creates a new ArticleParser with the specified HTTP client and loaded rules
func NewArticleParser(client *http.Client) (*ArticleParser, error) {
	parser := &ArticleParser{
		rules:             make(map[string]*ParsingRule),
		client:            client,
		heuristicExtract:  NewHeuristicExtractor(),
		metadataExtractor: NewMetadataExtractor(),
	}

	if err := parser.loadRules(); err != nil {
		return nil, fmt.Errorf("failed to load parsing rules: %w", err)
	}

	return parser, nil
}

// AddRule adds or replaces a parsing rule for a specific domain
func (p *ArticleParser) AddRule(domain string, rule *ParsingRule) {
	p.rules[domain] = rule
}

// SetHTTPClient overrides the HTTP client used for fetching article content.
func (p *ArticleParser) SetHTTPClient(client *http.Client) {
	p.client = client
}

func (p *ArticleParser) loadRules() error {
	entries, err := rulesFS.ReadDir("rules")
	if err != nil {
		return fmt.Errorf("failed to read rules directory: %w", err)
	}

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".txt") {
			continue
		}

		domain := strings.TrimSuffix(entry.Name(), ".txt")

		content, err := rulesFS.ReadFile(filepath.Join("rules", entry.Name()))
		if err != nil {
			return fmt.Errorf("failed to read rule file %s: %w", entry.Name(), err)
		}

		rule, err := p.parseRules(domain, string(content))
		if err != nil {
			return fmt.Errorf("failed to parse rule file %s: %w", entry.Name(), err)
		}

		p.rules[domain] = rule
	}

	return nil
}

func (p *ArticleParser) parseRules(domain, content string) (*ParsingRule, error) {
	rule := &ParsingRule{Domain: domain, Strip: []string{}}
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "title":
			rule.Title = value
		case "author":
			rule.Author = value
		case "date":
			rule.Date = value
		case "body":
			rule.Body = value
		case "strip":
			rule.Strip = append(rule.Strip, value)
		case "strip_id_or_class":
			rule.StripIDsOrClasses = append(rule.StripIDsOrClasses, value)
		case "prune":
			rule.Prune = parseBool(value)
		case "tidy":
			rule.Tidy = parseBool(value)
		case "test_url":
			rule.TestURLs = append(rule.TestURLs, value)
		default:
			if strings.HasPrefix(key, "http_header(") && strings.HasSuffix(key, ")") {
				headerName := strings.TrimSuffix(strings.TrimPrefix(key, "http_header("), ")")
				if headerName != "" {
					if rule.Headers == nil {
						rule.Headers = make(map[string]string)
					}
					rule.Headers[http.CanonicalHeaderKey(headerName)] = value
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading rule file: %w", err)
	}

	return rule, nil
}

func parseBool(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func (p *ArticleParser) findRule(domain string) *ParsingRule {
	for ruleDomain, rule := range p.rules {
		if domain == ruleDomain || strings.HasSuffix(domain, ruleDomain) {
			return rule
		}
	}
	return nil
}

// ParseURL extracts article content from a given URL
func (p *ArticleParser) ParseURL(s string) (*ParsedContent, error) {
	parsedURL, err := url.Parse(s)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	domain := parsedURL.Hostname()
	rule := p.findRule(domain)
	req, err := http.NewRequest(http.MethodGet, s, nil)

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if rule != nil {
		for header, value := range rule.Headers {
			if value == "" {
				continue
			}
			if req.Header.Get(header) == "" {
				req.Header.Set(header, value)
			}
		}
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	htmlBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return p.Parse(string(htmlBytes), domain, s)
}

// ParseHTML extracts article content from HTML string using domain-specific rules with heuristic fallback.
// Implements dual validation: compares XPath results with heuristic extraction when rules exist.
func (p *ArticleParser) Parse(htmlContent, domain, sourceURL string) (*ParsedContent, error) {
	doc, err := htmlquery.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	rule := p.findRule(domain)

	if rule == nil {
		return p.parseWithHeuristics(doc, sourceURL)
	}

	content := &ParsedContent{
		URL:              sourceURL,
		ExtractionMethod: "xpath",
		Confidence:       0.85,
	}

	if rule.Title != "" {
		if titleNode := htmlquery.FindOne(doc, rule.Title); titleNode != nil {
			content.Title = strings.TrimSpace(htmlquery.InnerText(titleNode))
		}
	}
	if content.Title == "" {
		content.Title = p.metadataExtractor.ExtractTitle(doc)
	}

	if rule.Author != "" {
		if authorNode := htmlquery.FindOne(doc, rule.Author); authorNode != nil {
			content.Author = strings.TrimSpace(htmlquery.InnerText(authorNode))
		}
	}
	if content.Author == "" {
		content.Author = p.metadataExtractor.ExtractAuthor(doc)
	}

	if rule.Date != "" {
		if dateNode := htmlquery.FindOne(doc, rule.Date); dateNode != nil {
			content.Date = strings.TrimSpace(htmlquery.InnerText(dateNode))
		}
	}
	if content.Date == "" {
		content.Date = p.metadataExtractor.ExtractPublishedDate(doc)
	}

	if rule.Body != "" {
		bodyNode := htmlquery.FindOne(doc, rule.Body)
		if bodyNode == nil {
			return p.parseWithHeuristics(doc, sourceURL)
		}

		for _, stripXPath := range rule.Strip {
			removeNodesByXPath(bodyNode, stripXPath)
		}

		for _, identifier := range rule.StripIDsOrClasses {
			removeNodesByIdentifier(bodyNode, identifier)
		}

		removeDefaultNonContentNodes(bodyNode)

		xpathContent := normalizeWhitespace(htmlquery.InnerText(bodyNode))

		heuristicResult := p.heuristicExtract.CompareWithXPath(doc, bodyNode)
		if heuristicResult != nil {
			content.Content = heuristicResult.Content
			if content.Content == "" {
				content.Content = xpathContent
			}
			content.Confidence = heuristicResult.Confidence
			content.ExtractionMethod = heuristicResult.ExtractionMethod
		} else {
			content.Content = xpathContent
		}
	}

	if content.Title == "" {
		return nil, fmt.Errorf("could not extract title from HTML")
	}

	return content, nil
}

// parseWithHeuristics performs heuristic-only extraction when no XPath rule exists.
func (p *ArticleParser) parseWithHeuristics(doc *exhtml.Node, sourceURL string) (*ParsedContent, error) {
	result := p.heuristicExtract.ExtractWithSemanticHTML(doc)
	if result == nil {
		result = &ExtractionResult{
			ExtractionMethod: "heuristic-failed",
			Confidence:       0.0,
		}
	}

	metadata := p.metadataExtractor.ExtractMetadata(doc)
	if metadata != nil {
		if result.Title == "" {
			result.Title = metadata.Title
		}
		if result.Author == "" {
			result.Author = metadata.Author
		}
		if result.PublishedDate == "" {
			result.PublishedDate = metadata.PublishedDate
		}
	}

	content := &ParsedContent{
		Title:            result.Title,
		Author:           result.Author,
		Date:             result.PublishedDate,
		Content:          result.Content,
		URL:              sourceURL,
		Confidence:       result.Confidence,
		ExtractionMethod: result.ExtractionMethod,
	}

	if content.Title == "" {
		return nil, fmt.Errorf("could not extract title from HTML using heuristics")
	}

	if content.Confidence < 0.3 {
		return nil, fmt.Errorf("heuristic extraction confidence too low (%.2f)", content.Confidence)
	}

	return content, nil
}

func removeNodesByXPath(root *exhtml.Node, xpath string) {
	if root == nil {
		return
	}

	xpath = strings.TrimSpace(xpath)
	if xpath == "" {
		return
	}

	nodes := htmlquery.Find(root, xpath)
	for _, node := range nodes {
		if node != nil && node.Parent != nil {
			node.Parent.RemoveChild(node)
		}
	}
}

func removeNodesByIdentifier(root *exhtml.Node, identifier string) {
	identifier = strings.TrimSpace(identifier)
	if root == nil || identifier == "" {
		return
	}

	idLiteral := buildXPathLiteral(identifier)
	removeNodesByXPath(root, fmt.Sprintf(".//*[@id=%s]", idLiteral))

	classLiteral := buildXPathLiteral(" " + identifier + " ")
	removeNodesByXPath(root, fmt.Sprintf(".//*[contains(concat(' ', normalize-space(@class), ' '), %s)]", classLiteral))
}

func removeDefaultNonContentNodes(root *exhtml.Node) {
	for _, xp := range []string{
		".//script",
		".//style",
		".//noscript",
	} {
		removeNodesByXPath(root, xp)
	}
}

func normalizeWhitespace(value string) string {
	value = strings.ReplaceAll(value, "\u00a0", " ")
	return strings.TrimSpace(value)
}

func buildXPathLiteral(value string) string {
	if !strings.Contains(value, "'") {
		return "'" + value + "'"
	}

	if !strings.Contains(value, "\"") {
		return `"` + value + `"`
	}

	segments := strings.Split(value, "'")
	var builder strings.Builder
	builder.WriteString("concat(")

	for i, segment := range segments {
		if i > 0 {
			builder.WriteString(", \"'\", ")
		}
		if segment == "" {
			builder.WriteString("''")
			continue
		}
		builder.WriteString("'")
		builder.WriteString(segment)
		builder.WriteString("'")
	}

	builder.WriteString(")")
	return builder.String()
}

// Convert HTML content directly to markdown using domain-specific rules
func (p *ArticleParser) Convert(htmlContent, domain, sourceURL string) (string, error) {
	content, err := p.Parse(htmlContent, domain, sourceURL)
	if err != nil {
		return "", err
	}

	return p.createMarkdown(content), nil
}

// GetSupportedDomains returns a list of domains that have parsing rules
func (p *ArticleParser) GetSupportedDomains() []string {
	var domains []string
	for domain := range p.rules {
		domains = append(domains, domain)
	}
	return domains
}

// SaveArticle saves the parsed content to filesystem and returns file paths
func (p *ArticleParser) SaveArticle(content *ParsedContent, dir string) (markdownPath, htmlPath string, err error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create storage directory: %w", err)
	}

	slug := p.slugify(content.Title)
	if slug == "" {
		slug = "article"
	}

	baseMarkdownPath := filepath.Join(dir, slug+".md")
	baseHTMLPath := filepath.Join(dir, slug+".html")
	markdownPath = baseMarkdownPath
	htmlPath = baseHTMLPath

	counter := 1
	for {
		if _, err := os.Stat(markdownPath); os.IsNotExist(err) {
			if _, err := os.Stat(htmlPath); os.IsNotExist(err) {
				break
			}
		}
		markdownPath = filepath.Join(dir, fmt.Sprintf("%s_%d.md", slug, counter))
		htmlPath = filepath.Join(dir, fmt.Sprintf("%s_%d.html", slug, counter))
		counter++
	}

	markdownContent := p.createMarkdown(content)

	if err := os.WriteFile(markdownPath, []byte(markdownContent), 0644); err != nil {
		return "", "", fmt.Errorf("failed to write markdown file: %w", err)
	}

	htmlContent := p.createHTML(content, markdownContent)

	if err := os.WriteFile(htmlPath, []byte(htmlContent), 0644); err != nil {
		os.Remove(markdownPath)
		return "", "", fmt.Errorf("failed to write HTML file: %w", err)
	}

	return markdownPath, htmlPath, nil
}

func (p *ArticleParser) slugify(title string) string {
	slug := strings.ToLower(title)

	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug = reg.ReplaceAllString(slug, "-")

	slug = strings.Trim(slug, "-")

	if len(slug) > 100 {
		slug = slug[:100]
		slug = strings.Trim(slug, "-")
	}

	return slug
}

func (p *ArticleParser) createMarkdown(content *ParsedContent) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("# %s\n\n", content.Title))

	if content.Author != "" {
		builder.WriteString(fmt.Sprintf("**Author:** %s\n\n", content.Author))
	}

	if content.Date != "" {
		builder.WriteString(fmt.Sprintf("**Date:** %s\n\n", content.Date))
	}

	builder.WriteString(fmt.Sprintf("**Source:** %s\n\n", content.URL))
	builder.WriteString(fmt.Sprintf("**Saved:** %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	builder.WriteString("---\n\n")
	builder.WriteString(content.Content)

	return builder.String()
}

func (p *ArticleParser) createHTML(content *ParsedContent, markdownContent string) string {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	mdParser := parser.NewWithExtensions(extensions)
	doc := mdParser.Parse([]byte(markdownContent))

	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	htmlBody := markdown.Render(doc, renderer)

	var builder strings.Builder
	builder.WriteString("<!DOCTYPE html>\n")
	builder.WriteString("<html>\n<head>\n")
	builder.WriteString(fmt.Sprintf("  <title>%s</title>\n", content.Title))
	builder.WriteString("  <meta charset=\"UTF-8\">\n")
	builder.WriteString("  <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n")
	builder.WriteString("  <style>\n")
	builder.WriteString("    body { font-family: Arial, sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; }\n")
	builder.WriteString("    pre { background-color: #f4f4f4; padding: 10px; border-radius: 4px; overflow-x: auto; }\n")
	builder.WriteString("    blockquote { border-left: 4px solid #ccc; padding-left: 16px; margin-left: 0; }\n")
	builder.WriteString("  </style>\n")
	builder.WriteString("</head>\n<body>\n")
	builder.Write(htmlBody)
	builder.WriteString("\n</body>\n</html>")

	return builder.String()
}

// CreateArticleFromURL is a convenience function that parses a URL and creates an instance of [models.Article]
func CreateArticleFromURL(url, dir string) (*models.Article, error) {
	parser, err := NewArticleParser(http.DefaultClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create parser: %w", err)
	}

	content, err := parser.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	mdPath, htmlPath, err := parser.SaveArticle(content, dir)
	if err != nil {
		return nil, fmt.Errorf("failed to save article: %w", err)
	}

	return &models.Article{
		URL:          url,
		Title:        content.Title,
		Author:       content.Author,
		Date:         content.Date,
		MarkdownPath: mdPath,
		HTMLPath:     htmlPath,
		Created:      time.Now(),
		Modified:     time.Now(),
	}, nil
}
