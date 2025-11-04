// Package public provides conversion between markdown and leaflet block formats
//
// Image handling follows a two-pass approach:
//  1. Gather all image URLs from the markdown AST
//  2. Resolve images (fetch bytes, get dimensions, upload to blob storage)
//  3. Convert markdown to blocks using the resolved image metadata
package public

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
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

// ImageInfo contains resolved image metadata
type ImageInfo struct {
	Blob   Blob
	Width  int
	Height int
}

// ImageResolver resolves image URLs to blob data and metadata
type ImageResolver interface {
	// ResolveImage resolves an image URL to blob data and dimensions
	// The url parameter may be a local file path or remote URL
	ResolveImage(url string) (*ImageInfo, error)
}

// LocalImageResolver resolves local file paths to image metadata
type LocalImageResolver struct {
	// BlobUploader is called to upload image bytes and get a blob reference
	// If nil, creates a placeholder blob with a hash-based CID
	//
	// TODO: CLI commands that publish documents must provide this function to upload
	// images to AT Protocol blob storage via com.atproto.repo.uploadBlob
	BlobUploader func(data []byte, mimeType string) (Blob, error)
}

// ResolveImage reads a local image file and extracts metadata
func (r *LocalImageResolver) ResolveImage(path string) (*ImageInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read image: %w", err)
	}

	img, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	mimeType := "image/" + format

	var blob Blob
	if r.BlobUploader != nil {
		blob, err = r.BlobUploader(data, mimeType)
		if err != nil {
			return nil, fmt.Errorf("failed to upload blob: %w", err)
		}
	} else {
		blob = Blob{
			Type:     TypeBlob,
			Ref:      CID{Link: "bafkreiplaceholder"},
			MimeType: mimeType,
			Size:     len(data),
		}
	}

	return &ImageInfo{
		Blob:   blob,
		Width:  img.Width,
		Height: img.Height,
	}, nil
}

// MarkdownConverter implements the [Converter] interface
type MarkdownConverter struct {
	extensions    parser.Extensions
	imageResolver ImageResolver
	basePath      string // Base path for resolving relative image paths
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

// WithImageResolver sets an image resolver for the converter
func (c *MarkdownConverter) WithImageResolver(resolver ImageResolver, basePath string) *MarkdownConverter {
	c.imageResolver = resolver
	c.basePath = basePath
	return c
}

// ToLeaflet converts markdown to leaflet blocks
func (c *MarkdownConverter) ToLeaflet(markdown string) ([]BlockWrap, error) {
	p := parser.NewWithExtensions(c.extensions)
	doc := p.Parse([]byte(markdown))
	imageURLs := c.gatherImages(doc)

	resolvedImages := make(map[string]*ImageInfo)
	if c.imageResolver != nil {
		for _, url := range imageURLs {
			resolvedPath := url
			if !filepath.IsAbs(url) && c.basePath != "" {
				resolvedPath = filepath.Join(c.basePath, url)
			}

			info, err := c.imageResolver.ResolveImage(resolvedPath)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve image %s: %w", url, err)
			}
			resolvedImages[url] = info
		}
	}

	var blocks []BlockWrap

	for _, child := range doc.GetChildren() {
		switch n := child.(type) {
		case *ast.Heading:
			if block := c.convertHeading(n, resolvedImages); block != nil {
				blocks = append(blocks, *block)
			}
		case *ast.Paragraph:
			convertedBlocks := c.convertParagraph(n, resolvedImages)
			blocks = append(blocks, convertedBlocks...)
		case *ast.CodeBlock:
			if block := c.convertCodeBlock(n); block != nil {
				blocks = append(blocks, *block)
			}
		case *ast.BlockQuote:
			if block := c.convertBlockquote(n, resolvedImages); block != nil {
				blocks = append(blocks, *block)
			}
		case *ast.List:
			if block := c.convertList(n, resolvedImages); block != nil {
				blocks = append(blocks, *block)
			}
		case *ast.HorizontalRule:
			blocks = append(blocks, BlockWrap{
				Type: TypeBlock,
				Block: HorizontalRuleBlock{
					Type: TypeHorizontalRuleBlock,
				},
			})
		case *ast.Image:
			if block := c.convertImage(n, resolvedImages); block != nil {
				blocks = append(blocks, *block)
			}
		}
	}

	return blocks, nil
}

// gatherImages walks the AST and collects all image URLs
func (c *MarkdownConverter) gatherImages(node ast.Node) []string {
	var urls []string

	ast.WalkFunc(node, func(n ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.GoToNext
		}

		if img, ok := n.(*ast.Image); ok {
			urls = append(urls, string(img.Destination))
		}

		return ast.GoToNext
	})

	return urls
}

// convertHeading converts an AST heading to a leaflet HeaderBlock
func (c *MarkdownConverter) convertHeading(node *ast.Heading, resolvedImages map[string]*ImageInfo) *BlockWrap {
	text, facets, _ := c.extractTextAndFacets(node, resolvedImages)
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

// convertParagraph converts an AST paragraph to leaflet blocks
func (c *MarkdownConverter) convertParagraph(node *ast.Paragraph, resolvedImages map[string]*ImageInfo) []BlockWrap {
	text, facets, imageBlocks := c.extractTextAndFacets(node, resolvedImages)

	if len(imageBlocks) > 0 {
		return imageBlocks
	}

	if strings.TrimSpace(text) == "" {
		return nil
	}

	return []BlockWrap{{
		Type: TypeBlock,
		Block: TextBlock{
			Type:      TypeTextBlock,
			Plaintext: text,
			Facets:    facets,
		},
	}}
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
func (c *MarkdownConverter) convertBlockquote(node *ast.BlockQuote, resolvedImages map[string]*ImageInfo) *BlockWrap {
	text, facets, _ := c.extractTextAndFacets(node, resolvedImages)
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
func (c *MarkdownConverter) convertList(node *ast.List, resolvedImages map[string]*ImageInfo) *BlockWrap {
	var items []ListItem

	for _, child := range node.Children {
		if listItem, ok := child.(*ast.ListItem); ok {
			item := c.convertListItem(listItem, resolvedImages)
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
func (c *MarkdownConverter) convertListItem(node *ast.ListItem, resolvedImages map[string]*ImageInfo) *ListItem {
	text, facets, _ := c.extractTextAndFacets(node, resolvedImages)
	return &ListItem{
		Type: TypeListItem,
		Content: TextBlock{
			Type:      TypeTextBlock,
			Plaintext: text,
			Facets:    facets,
		},
	}
}

// convertImage converts an AST image to a leaflet ImageBlock
func (c *MarkdownConverter) convertImage(node *ast.Image, resolvedImages map[string]*ImageInfo) *BlockWrap {
	alt := string(node.Title)
	if alt == "" {
		for _, child := range node.Children {
			if text, ok := child.(*ast.Text); ok {
				alt = string(text.Literal)
				break
			}
		}
	}

	info, hasInfo := resolvedImages[string(node.Destination)]

	var blob Blob
	var aspectRatio AspectRatio

	if hasInfo {
		blob = info.Blob
		aspectRatio = AspectRatio{
			Type:   TypeAspectRatio,
			Width:  info.Width,
			Height: info.Height,
		}
	} else {
		blob = Blob{
			Type:     TypeBlob,
			Ref:      CID{Link: "bafkreiplaceholder"},
			MimeType: "image/jpeg",
			Size:     0,
		}
		aspectRatio = AspectRatio{
			Type:   TypeAspectRatio,
			Width:  1,
			Height: 1,
		}
	}

	return &BlockWrap{
		Type: TypeBlock,
		Block: ImageBlock{
			Type:        TypeImageBlock,
			Image:       blob,
			Alt:         alt,
			AspectRatio: aspectRatio,
		},
	}
}

// extractTextAndFacets extracts plaintext, facets, and image blocks from an AST node
func (c *MarkdownConverter) extractTextAndFacets(node ast.Node, resolvedImages map[string]*ImageInfo) (string, []Facet, []BlockWrap) {
	var buf bytes.Buffer
	var facets []Facet
	var blocks []BlockWrap
	offset := 0

	var stack []formatContext

	ast.WalkFunc(node, func(n ast.Node, entering bool) ast.WalkStatus {
		switch v := n.(type) {
		case *ast.Text:
			if entering {
				content := string(v.Literal)
				buf.WriteString(content)

				if len(stack) > 0 {
					var allFeatures []FacetFeature
					for _, ctx := range stack {
						allFeatures = append(allFeatures, ctx.features...)
					}
					facet := Facet{
						Type: TypeFacet,
						Index: ByteSlice{
							Type:      TypeByteSlice,
							ByteStart: offset,
							ByteEnd:   offset + len(content),
						},
						Features: allFeatures,
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
		case *ast.Image:
			if entering {
				if buf.Len() > 0 {
					blocks = append(blocks, BlockWrap{
						Type: TypeBlock,
						Block: TextBlock{
							Type:      TypeTextBlock,
							Plaintext: buf.String(),
							Facets:    facets,
						},
					})
					buf.Reset()
					facets = nil
					offset = 0
				}

				if imgBlock := c.convertImage(v, resolvedImages); imgBlock != nil {
					blocks = append(blocks, *imgBlock)
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

	// If we created blocks, add any remaining text
	if len(blocks) > 0 && buf.Len() > 0 {
		blocks = append(blocks, BlockWrap{
			Type: TypeBlock,
			Block: TextBlock{
				Type:      TypeTextBlock,
				Plaintext: buf.String(),
				Facets:    facets,
			},
		})
	}

	return buf.String(), facets, blocks
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
