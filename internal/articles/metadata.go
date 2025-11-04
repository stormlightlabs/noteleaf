package articles

import (
	"encoding/json"
	"strings"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

// MetadataExtractor implements multi-strategy metadata extraction from HTML documents.
// It attempts to extract article metadata using OpenGraph, Schema.org, meta tags,
// and semantic HTML5 elements, with fallback chains for each field.
type MetadataExtractor struct{}

// NewMetadataExtractor creates a new metadata extractor.
func NewMetadataExtractor() *MetadataExtractor {
	return &MetadataExtractor{}
}

// ExtractMetadata extracts all available metadata from an HTML document.
// Returns an ExtractionResult with populated metadata fields.
func (m *MetadataExtractor) ExtractMetadata(doc *html.Node) *ExtractionResult {
	if doc == nil {
		return &ExtractionResult{}
	}

	result := &ExtractionResult{}

	result.Title = m.ExtractTitle(doc)
	result.Author = m.ExtractAuthor(doc)
	result.PublishedDate = m.ExtractPublishedDate(doc)
	result.SiteName = m.ExtractSiteName(doc)
	result.Language = m.ExtractLanguage(doc)

	return result
}

// ExtractTitle extracts the article title using multiple strategies.
// Tries in order: OpenGraph, Schema.org, meta tags, h1, title tag.
func (m *MetadataExtractor) ExtractTitle(doc *html.Node) string {
	if doc == nil {
		return ""
	}

	if title := m.getMetaContent(doc, "property", "og:title"); title != "" {
		return title
	}

	if title := m.getSchemaOrgField(doc, "headline"); title != "" {
		return title
	}

	if title := m.getSchemaOrgField(doc, "name"); title != "" {
		return title
	}

	if title := m.getMetaContent(doc, "name", "twitter:title"); title != "" {
		return title
	}

	if title := m.getMetaContent(doc, "property", "article:title"); title != "" {
		return title
	}

	if h1 := htmlquery.FindOne(doc, "//h1"); h1 != nil {
		if title := htmlquery.InnerText(h1); title != "" {
			return strings.TrimSpace(title)
		}
	}

	if titleNode := htmlquery.FindOne(doc, "//title"); titleNode != nil {
		if title := htmlquery.InnerText(titleNode); title != "" {
			return strings.TrimSpace(title)
		}
	}

	return ""
}

// ExtractAuthor extracts the article author using multiple strategies.
// Tries in order: OpenGraph, Schema.org, meta tags, rel=author, byline elements.
func (m *MetadataExtractor) ExtractAuthor(doc *html.Node) string {
	if doc == nil {
		return ""
	}

	if author := m.getMetaContent(doc, "property", "og:author"); author != "" {
		return author
	}

	if author := m.getSchemaOrgField(doc, "author"); author != "" {
		return author
	}

	if author := m.getMetaContent(doc, "property", "article:author"); author != "" {
		return author
	}

	if author := m.getMetaContent(doc, "name", "twitter:creator"); author != "" {
		return author
	}

	if author := m.getMetaContent(doc, "name", "author"); author != "" {
		return author
	}

	if authorLink := htmlquery.FindOne(doc, "//a[@rel='author']"); authorLink != nil {
		if author := htmlquery.InnerText(authorLink); author != "" {
			return strings.TrimSpace(author)
		}
	}

	bylineSelectors := []string{
		"//span[contains(@class, 'author')]",
		"//div[contains(@class, 'author')]",
		"//p[contains(@class, 'byline')]",
		"//span[contains(@class, 'byline')]",
	}

	for _, selector := range bylineSelectors {
		if node := htmlquery.FindOne(doc, selector); node != nil {
			if author := htmlquery.InnerText(node); author != "" {
				return strings.TrimSpace(author)
			}
		}
	}

	return ""
}

// ExtractPublishedDate extracts the publication date using multiple strategies.
// Tries in order: OpenGraph, Schema.org, article:published_time, time elements.
func (m *MetadataExtractor) ExtractPublishedDate(doc *html.Node) string {
	if doc == nil {
		return ""
	}

	if date := m.getMetaContent(doc, "property", "og:published_time"); date != "" {
		return date
	}

	if date := m.getSchemaOrgField(doc, "datePublished"); date != "" {
		return date
	}

	if date := m.getSchemaOrgField(doc, "publishDate"); date != "" {
		return date
	}

	if date := m.getMetaContent(doc, "property", "article:published_time"); date != "" {
		return date
	}

	if date := m.getMetaContent(doc, "name", "publication_date"); date != "" {
		return date
	}

	if date := m.getMetaContent(doc, "name", "date"); date != "" {
		return date
	}

	if timeNode := htmlquery.FindOne(doc, "//time[@datetime]"); timeNode != nil {
		for _, attr := range timeNode.Attr {
			if attr.Key == "datetime" {
				return attr.Val
			}
		}
	}

	return ""
}

// ExtractSiteName extracts the site name using multiple strategies.
// Tries in order: OpenGraph, Schema.org, meta tags.
func (m *MetadataExtractor) ExtractSiteName(doc *html.Node) string {
	if doc == nil {
		return ""
	}

	if siteName := m.getMetaContent(doc, "property", "og:site_name"); siteName != "" {
		return siteName
	}

	if publisher := m.getSchemaOrgField(doc, "publisher"); publisher != "" {
		return publisher
	}

	if siteName := m.getMetaContent(doc, "name", "application-name"); siteName != "" {
		return siteName
	}

	return ""
}

// ExtractLanguage extracts the document language.
// Tries in order: html lang attribute, OpenGraph, meta tags.
func (m *MetadataExtractor) ExtractLanguage(doc *html.Node) string {
	if doc == nil {
		return ""
	}

	if htmlNode := htmlquery.FindOne(doc, "//html"); htmlNode != nil {
		for _, attr := range htmlNode.Attr {
			if attr.Key == "lang" {
				return attr.Val
			}
		}
	}

	if locale := m.getMetaContent(doc, "property", "og:locale"); locale != "" {
		return locale
	}

	if lang := m.getMetaContent(doc, "http-equiv", "content-language"); lang != "" {
		return lang
	}

	return ""
}

// getMetaContent retrieves the content attribute from a meta tag.
// Searches for meta tags with the specified attribute name and value.
func (m *MetadataExtractor) getMetaContent(doc *html.Node, attrName, attrValue string) string {
	if doc == nil {
		return ""
	}

	xpath := "//meta[@" + attrName + "='" + attrValue + "']"
	metaNode := htmlquery.FindOne(doc, xpath)

	if metaNode == nil {
		return ""
	}

	for _, attr := range metaNode.Attr {
		if attr.Key == "content" {
			return strings.TrimSpace(attr.Val)
		}
	}

	return ""
}

// getSchemaOrgField extracts a field from Schema.org JSON-LD structured data.
func (m *MetadataExtractor) getSchemaOrgField(doc *html.Node, fieldName string) string {
	if doc == nil {
		return ""
	}

	scripts := htmlquery.Find(doc, "//script[@type='application/ld+json']")

	for _, script := range scripts {
		if script.FirstChild == nil || script.FirstChild.Type != html.TextNode {
			continue
		}

		var data map[string]any
		if err := json.Unmarshal([]byte(script.FirstChild.Data), &data); err != nil {
			continue
		}

		context, hasContext := data["@context"]
		typeVal, hasType := data["@type"]

		if !hasContext || !hasType {
			continue
		}

		contextStr, ok := context.(string)
		if !ok || !strings.Contains(contextStr, "schema.org") {
			continue
		}

		typeStr, ok := typeVal.(string)
		if !ok || (!strings.Contains(typeStr, "Article") && !strings.Contains(typeStr, "NewsArticle")) {
			continue
		}

		if value, exists := data[fieldName]; exists {
			return m.extractStringValue(value)
		}
	}

	return ""
}

// extractStringValue extracts a string from various JSON value types.
func (m *MetadataExtractor) extractStringValue(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case map[string]any:
		if name, exists := v["name"]; exists {
			if nameStr, ok := name.(string); ok {
				return nameStr
			}
		}
	case []any:
		if len(v) > 0 {
			return m.extractStringValue(v[0])
		}
	}
	return ""
}
