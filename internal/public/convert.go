// Package public provides conversion between markdown and leaflet block formats
//
// TODO: Handle overlapping facets
// TODO: Implement image handling - requires blob resolution
package public

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
)

// Converter defines the interface for converting between a document and leaflet formats
type Converter interface {
	// ToLeaflet converts content to leaflet blocks
	ToLeaflet(content string) ([]BlockWrap, error)
	// FromLeaflet converts leaflet blocks back to the original format
	FromLeaflet(blocks []BlockWrap) (string, error)
}

// MarkdownConverter implements the [Converter] interface
type MarkdownConverter struct {
	extensions parser.Extensions
}

type formatContext struct {
	features []FacetFeature
	start    int
}

// NewMarkdownConverter creates a new markdown converter
func NewMarkdownConverter() *MarkdownConverter {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	return &MarkdownConverter{
		extensions: extensions,
	}
}

// ToLeaflet converts markdown to leaflet blocks
func (c *MarkdownConverter) ToLeaflet(markdown string) ([]BlockWrap, error) {
	p := parser.NewWithExtensions(c.extensions)
	doc := p.Parse([]byte(markdown))

	var blocks []BlockWrap

	for _, child := range doc.GetChildren() {
		switch n := child.(type) {
		case *ast.Heading:
			if block := c.convertHeading(n); block != nil {
				blocks = append(blocks, *block)
			}
		case *ast.Paragraph:
			if block := c.convertParagraph(n); block != nil {
				blocks = append(blocks, *block)
			}
		case *ast.CodeBlock:
			if block := c.convertCodeBlock(n); block != nil {
				blocks = append(blocks, *block)
			}
		case *ast.BlockQuote:
			if block := c.convertBlockquote(n); block != nil {
				blocks = append(blocks, *block)
			}
		case *ast.List:
			if block := c.convertList(n); block != nil {
				blocks = append(blocks, *block)
			}
		case *ast.HorizontalRule:
			blocks = append(blocks, BlockWrap{
				Type: TypeBlock,
				Block: HorizontalRuleBlock{
					Type: TypeHorizontalRuleBlock,
				},
			})
		}
	}

	return blocks, nil
}

// convertHeading converts an AST heading to a leaflet HeaderBlock
func (c *MarkdownConverter) convertHeading(node *ast.Heading) *BlockWrap {
	text, facets := c.extractTextAndFacets(node)
	return &BlockWrap{
		Type: TypeBlock,
		Block: HeaderBlock{
			Type:      TypeHeaderBlock,
			Level:     node.Level,
			Plaintext: text,
			Facets:    facets,
		},
	}
}

// convertParagraph converts an AST paragraph to a leaflet TextBlock
func (c *MarkdownConverter) convertParagraph(node *ast.Paragraph) *BlockWrap {
	text, facets := c.extractTextAndFacets(node)
	if strings.TrimSpace(text) == "" {
		return nil
	}

	return &BlockWrap{
		Type: TypeBlock,
		Block: TextBlock{
			Type:      TypeTextBlock,
			Plaintext: text,
			Facets:    facets,
		},
	}
}

// convertCodeBlock converts an AST code block to a leaflet CodeBlock
func (c *MarkdownConverter) convertCodeBlock(node *ast.CodeBlock) *BlockWrap {
	return &BlockWrap{
		Type: TypeBlock,
		Block: CodeBlock{
			Type:                    TypeCodeBlock,
			Plaintext:               string(node.Literal),
			Language:                string(node.Info),
			SyntaxHighlightingTheme: "catppuccin-mocha",
		},
	}
}

// convertBlockquote converts an AST blockquote to a leaflet BlockquoteBlock
func (c *MarkdownConverter) convertBlockquote(node *ast.BlockQuote) *BlockWrap {
	text, facets := c.extractTextAndFacets(node)
	return &BlockWrap{
		Type: TypeBlock,
		Block: BlockquoteBlock{
			Type:      TypeBlockquoteBlock,
			Plaintext: text,
			Facets:    facets,
		},
	}
}

// convertList converts an AST list to a leaflet UnorderedListBlock
func (c *MarkdownConverter) convertList(node *ast.List) *BlockWrap {
	var items []ListItem

	for _, child := range node.Children {
		if listItem, ok := child.(*ast.ListItem); ok {
			item := c.convertListItem(listItem)
			if item != nil {
				items = append(items, *item)
			}
		}
	}

	return &BlockWrap{
		Type: TypeBlock,
		Block: UnorderedListBlock{
			Type:     TypeUnorderedListBlock,
			Children: items,
		},
	}
}

// convertListItem converts an AST list item to a leaflet ListItem
func (c *MarkdownConverter) convertListItem(node *ast.ListItem) *ListItem {
	text, facets := c.extractTextAndFacets(node)
	return &ListItem{
		Type: TypeListItem,
		Content: TextBlock{
			Type:      TypeTextBlock,
			Plaintext: text,
			Facets:    facets,
		},
	}
}

// extractTextAndFacets extracts plaintext and facets from an AST node
func (c *MarkdownConverter) extractTextAndFacets(node ast.Node) (string, []Facet) {
	var buf bytes.Buffer
	var facets []Facet
	offset := 0

	var stack []formatContext

	ast.WalkFunc(node, func(n ast.Node, entering bool) ast.WalkStatus {
		switch v := n.(type) {
		case *ast.Text:
			if entering {
				content := string(v.Literal)
				buf.WriteString(content)

				if len(stack) > 0 {
					ctx := stack[len(stack)-1]
					facet := Facet{
						Type: TypeFacet,
						Index: ByteSlice{
							Type:      TypeByteSlice,
							ByteStart: offset,
							ByteEnd:   offset + len(content),
						},
						Features: ctx.features,
					}
					facets = append(facets, facet)
				}

				offset += len(content)
			}
		case *ast.Strong:
			if entering {
				stack = append(stack, formatContext{
					features: []FacetFeature{FacetBold{Type: TypeFacetBold}},
					start:    offset,
				})
			} else {
				if len(stack) > 0 {
					stack = stack[:len(stack)-1]
				}
			}
		case *ast.Emph:
			if entering {
				stack = append(stack, formatContext{
					features: []FacetFeature{FacetItalic{Type: TypeFacetItalic}},
					start:    offset,
				})
			} else {
				if len(stack) > 0 {
					stack = stack[:len(stack)-1]
				}
			}
		case *ast.Del:
			if entering {
				stack = append(stack, formatContext{
					features: []FacetFeature{FacetStrikethrough{Type: TypeFacetStrike}},
					start:    offset,
				})
			} else {
				if len(stack) > 0 {
					stack = stack[:len(stack)-1]
				}
			}
		case *ast.Code:
			if entering {
				content := string(v.Literal)
				buf.WriteString(content)

				facet := Facet{
					Type: TypeFacet,
					Index: ByteSlice{
						Type:      TypeByteSlice,
						ByteStart: offset,
						ByteEnd:   offset + len(content),
					},
					Features: []FacetFeature{FacetCode{Type: TypeFacetCode}},
				}
				facets = append(facets, facet)

				offset += len(content)
			}
		case *ast.Link:
			if entering {
				stack = append(stack, formatContext{
					features: []FacetFeature{FacetLink{
						Type: TypeFacetLink,
						URI:  string(v.Destination),
					}},
					start: offset,
				})
			} else {
				if len(stack) > 0 {
					stack = stack[:len(stack)-1]
				}
			}
		case *ast.Softbreak, *ast.Hardbreak:
			if entering {
				buf.WriteString(" ")
				offset++
			}
		}
		return ast.GoToNext
	})
	return buf.String(), facets
}

// FromLeaflet converts leaflet blocks back to markdown
func (c *MarkdownConverter) FromLeaflet(blocks []BlockWrap) (string, error) {
	var buf bytes.Buffer
	for i, wrap := range blocks {
		if i > 0 {
			buf.WriteString("\n\n")
		}

		switch block := wrap.Block.(type) {
		case TextBlock:
			buf.WriteString(c.facetsToMarkdown(block.Plaintext, block.Facets))
		case HeaderBlock:
			buf.WriteString(strings.Repeat("#", block.Level))
			buf.WriteString(" ")
			buf.WriteString(c.facetsToMarkdown(block.Plaintext, block.Facets))
		case CodeBlock:
			buf.WriteString("```")
			if block.Language != "" {
				buf.WriteString(block.Language)
			}
			buf.WriteString("\n")
			buf.WriteString(block.Plaintext)
			if !strings.HasSuffix(block.Plaintext, "\n") {
				buf.WriteString("\n")
			}
			buf.WriteString("```")
		case BlockquoteBlock:
			buf.WriteString("> ")
			buf.WriteString(c.facetsToMarkdown(block.Plaintext, block.Facets))
		case UnorderedListBlock:
			c.listToMarkdown(&buf, block.Children, 0)
		case HorizontalRuleBlock:
			buf.WriteString("---")
		case ImageBlock:
			buf.WriteString("![")
			buf.WriteString(block.Alt)
			buf.WriteString("](image-placeholder)")
		default:
			return "", fmt.Errorf("unsupported block type: %T", block)
		}
	}

	return buf.String(), nil
}

// facetsToMarkdown applies facets to plaintext and generates markdown
func (c *MarkdownConverter) facetsToMarkdown(text string, facets []Facet) string {
	if len(facets) == 0 {
		return text
	}

	var buf bytes.Buffer
	lastEnd := 0

	for _, facet := range facets {
		if facet.Index.ByteStart > lastEnd {
			buf.WriteString(text[lastEnd:facet.Index.ByteStart])
		}

		facetText := text[facet.Index.ByteStart:facet.Index.ByteEnd]

		for _, feature := range facet.Features {
			switch f := feature.(type) {
			case FacetBold:
				facetText = "**" + facetText + "**"
			case FacetItalic:
				facetText = "*" + facetText + "*"
			case FacetCode:
				facetText = "`" + facetText + "`"
			case FacetStrikethrough:
				facetText = "~~" + facetText + "~~"
			case FacetLink:
				facetText = "[" + facetText + "](" + f.URI + ")"
			}
		}

		buf.WriteString(facetText)
		lastEnd = facet.Index.ByteEnd
	}

	if lastEnd < len(text) {
		buf.WriteString(text[lastEnd:])
	}

	return buf.String()
}

// listToMarkdown converts a list to markdown with proper indentation
func (c *MarkdownConverter) listToMarkdown(buf *bytes.Buffer, items []ListItem, depth int) {
	indent := strings.Repeat("  ", depth)

	for _, item := range items {
		buf.WriteString(indent)
		buf.WriteString("- ")

		switch content := item.Content.(type) {
		case TextBlock:
			buf.WriteString(c.facetsToMarkdown(content.Plaintext, content.Facets))
		case HeaderBlock:
			buf.WriteString(c.facetsToMarkdown(content.Plaintext, content.Facets))
		}

		buf.WriteString("\n")

		if len(item.Children) > 0 {
			c.listToMarkdown(buf, item.Children, depth+1)
		}
	}
}
