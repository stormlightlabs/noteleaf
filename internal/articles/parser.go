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
)

//go:embed rules/*.txt
var rulesFS embed.FS

// ParsedContent represents the extracted content from a web page
type ParsedContent struct {
	Title   string
	Author  string
	Date    string
	Content string
	URL     string
}

// ParsingRule represents XPath rules for extracting content from a specific domain
type ParsingRule struct {
	Domain   string
	Title    string
	Author   string
	Date     string
	Body     string
	Strip    []string // XPath selectors for elements to remove
	TestURLs []string
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
	rules  map[string]*ParsingRule
	client *http.Client
}

// NewArticleParser creates a new ArticleParser with the specified HTTP client and loaded rules
func NewArticleParser(client *http.Client) (*ArticleParser, error) {
	parser := &ArticleParser{
		rules:  make(map[string]*ParsingRule),
		client: client,
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
		case "test_url":
			rule.TestURLs = append(rule.TestURLs, value)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading rule file: %w", err)
	}

	return rule, nil
}

// ParseURL extracts article content from a given URL
func (p *ArticleParser) ParseURL(s string) (*ParsedContent, error) {
	parsedURL, err := url.Parse(s)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	domain := parsedURL.Hostname()

	resp, err := p.client.Get(s)
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

// ParseHTML extracts article content from HTML string using domain-specific rules
func (p *ArticleParser) Parse(htmlContent, domain, sourceURL string) (*ParsedContent, error) {
	var rule *ParsingRule
	for ruleDomain, r := range p.rules {
		if strings.Contains(domain, ruleDomain) {
			rule = r
			break
		}
	}

	if rule == nil {
		return nil, fmt.Errorf("no parsing rule found for domain: %s", domain)
	}

	doc, err := htmlquery.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	content := &ParsedContent{URL: sourceURL}

	if rule.Title != "" {
		if titleNode := htmlquery.FindOne(doc, rule.Title); titleNode != nil {
			content.Title = strings.TrimSpace(htmlquery.InnerText(titleNode))
		}
	}

	if rule.Author != "" {
		if authorNode := htmlquery.FindOne(doc, rule.Author); authorNode != nil {
			content.Author = strings.TrimSpace(htmlquery.InnerText(authorNode))
		}
	}

	if rule.Date != "" {
		if dateNode := htmlquery.FindOne(doc, rule.Date); dateNode != nil {
			content.Date = strings.TrimSpace(htmlquery.InnerText(dateNode))
		}
	}

	if rule.Body != "" {
		if bodyNode := htmlquery.FindOne(doc, rule.Body); bodyNode != nil {
			for _, stripXPath := range rule.Strip {
				stripNodes := htmlquery.Find(bodyNode, stripXPath)
				for _, node := range stripNodes {
					node.Parent.RemoveChild(node)
				}
			}

			content.Content = strings.TrimSpace(htmlquery.InnerText(bodyNode))
		}
	}

	if content.Title == "" {
		return nil, fmt.Errorf("could not extract title from HTML")
	}

	return content, nil
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
