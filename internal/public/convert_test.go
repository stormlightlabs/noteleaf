package public

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stormlightlabs/noteleaf/internal/shared"
)

func TestMarkdownConverter(t *testing.T) {
	converter := NewMarkdownConverter()

	t.Run("Conversion", func(t *testing.T) {
		t.Run("converts heading to HeaderBlock", func(t *testing.T) {
			markdown := "# Hello World"
			blocks, err := converter.ToLeaflet(markdown)
			shared.AssertNoError(t, err, "ToLeaflet should succeed")
			shared.AssertEqual(t, 1, len(blocks), "should have 1 block")

			header, ok := blocks[0].Block.(HeaderBlock)
			shared.AssertTrue(t, ok, "block should be HeaderBlock")
			shared.AssertEqual(t, TypeHeaderBlock, header.Type, "type should match")
			shared.AssertEqual(t, 1, header.Level, "level should be 1")
			shared.AssertEqual(t, "Hello World", header.Plaintext, "text should match")
		})

		t.Run("converts multiple heading levels", func(t *testing.T) {
			markdown := "## Level 2\n\n### Level 3\n\n###### Level 6"
			blocks, err := converter.ToLeaflet(markdown)
			shared.AssertNoError(t, err, "ToLeaflet should succeed")
			shared.AssertEqual(t, 3, len(blocks), "should have 3 blocks")

			h2 := blocks[0].Block.(HeaderBlock)
			shared.AssertEqual(t, 2, h2.Level, "first heading level")

			h3 := blocks[1].Block.(HeaderBlock)
			shared.AssertEqual(t, 3, h3.Level, "second heading level")

			h6 := blocks[2].Block.(HeaderBlock)
			shared.AssertEqual(t, 6, h6.Level, "third heading level")
		})

		t.Run("converts paragraph to TextBlock", func(t *testing.T) {
			markdown := "This is a simple paragraph."
			blocks, err := converter.ToLeaflet(markdown)
			shared.AssertNoError(t, err, "ToLeaflet should succeed")
			shared.AssertEqual(t, 1, len(blocks), "should have 1 block")

			text, ok := blocks[0].Block.(TextBlock)
			shared.AssertTrue(t, ok, "block should be TextBlock")
			shared.AssertEqual(t, TypeTextBlock, text.Type, "type should match")
			shared.AssertEqual(t, "This is a simple paragraph.", text.Plaintext, "text should match")
		})

		t.Run("converts code block to CodeBlock", func(t *testing.T) {
			markdown := "```go\nfunc main() {\n}\n```"
			blocks, err := converter.ToLeaflet(markdown)
			shared.AssertNoError(t, err, "ToLeaflet should succeed")
			shared.AssertEqual(t, 1, len(blocks), "should have 1 block")

			code, ok := blocks[0].Block.(CodeBlock)
			shared.AssertTrue(t, ok, "block should be CodeBlock")
			shared.AssertEqual(t, TypeCodeBlock, code.Type, "type should match")
			shared.AssertEqual(t, "go", code.Language, "language should match")
			shared.AssertTrue(t, strings.Contains(code.Plaintext, "func main"), "code content should match")
		})

		t.Run("converts blockquote to BlockquoteBlock", func(t *testing.T) {
			markdown := "> This is a quote"
			blocks, err := converter.ToLeaflet(markdown)
			shared.AssertNoError(t, err, "ToLeaflet should succeed")
			shared.AssertEqual(t, 1, len(blocks), "should have 1 block")

			quote, ok := blocks[0].Block.(BlockquoteBlock)
			shared.AssertTrue(t, ok, "block should be BlockquoteBlock")
			shared.AssertEqual(t, TypeBlockquoteBlock, quote.Type, "type should match")
			shared.AssertTrue(t, strings.Contains(quote.Plaintext, "This is a quote"), "quote text should match")
		})

		t.Run("converts list to UnorderedListBlock", func(t *testing.T) {
			markdown := "- Item 1\n- Item 2\n- Item 3"
			blocks, err := converter.ToLeaflet(markdown)
			shared.AssertNoError(t, err, "ToLeaflet should succeed")
			shared.AssertEqual(t, 1, len(blocks), "should have 1 block")

			list, ok := blocks[0].Block.(UnorderedListBlock)
			shared.AssertTrue(t, ok, "block should be UnorderedListBlock")
			shared.AssertEqual(t, TypeUnorderedListBlock, list.Type, "type should match")
			shared.AssertEqual(t, 3, len(list.Children), "should have 3 items")

			item1 := list.Children[0].Content.(TextBlock)
			shared.AssertTrue(t, strings.Contains(item1.Plaintext, "Item 1"), "first item text")
		})

		t.Run("converts horizontal rule to HorizontalRuleBlock", func(t *testing.T) {
			markdown := "---"
			blocks, err := converter.ToLeaflet(markdown)
			shared.AssertNoError(t, err, "ToLeaflet should succeed")
			shared.AssertEqual(t, 1, len(blocks), "should have 1 block")

			hr, ok := blocks[0].Block.(HorizontalRuleBlock)
			shared.AssertTrue(t, ok, "block should be HorizontalRuleBlock")
			shared.AssertEqual(t, TypeHorizontalRuleBlock, hr.Type, "type should match")
		})

		t.Run("converts mixed blocks", func(t *testing.T) {
			markdown := `# Title

This is a paragraph.

## Subtitle

- List item 1
- List item 2

---

` + "```go\ncode\n```"

			blocks, err := converter.ToLeaflet(markdown)

			shared.AssertNoError(t, err, "ToLeaflet should succeed")
			shared.AssertTrue(t, len(blocks) >= 5, "should have multiple blocks")
		})
	})

	t.Run("Facets", func(t *testing.T) {
		t.Run("extracts bold facet", func(t *testing.T) {
			markdown := "This is **bold** text"
			blocks, err := converter.ToLeaflet(markdown)

			shared.AssertNoError(t, err, "ToLeaflet should succeed")
			text := blocks[0].Block.(TextBlock)

			shared.AssertTrue(t, len(text.Facets) > 0, "should have facets")
			shared.AssertTrue(t, strings.Contains(text.Plaintext, "bold"), "text should contain 'bold'")
		})

		t.Run("extracts italic facet", func(t *testing.T) {
			markdown := "This is *italic* text"
			blocks, err := converter.ToLeaflet(markdown)

			shared.AssertNoError(t, err, "ToLeaflet should succeed")
			text := blocks[0].Block.(TextBlock)

			shared.AssertTrue(t, len(text.Facets) > 0, "should have facets")
		})

		t.Run("extracts inline code facet", func(t *testing.T) {
			markdown := "This is `code` text"
			blocks, err := converter.ToLeaflet(markdown)

			shared.AssertNoError(t, err, "ToLeaflet should succeed")
			text := blocks[0].Block.(TextBlock)

			shared.AssertTrue(t, len(text.Facets) > 0, "should have facets")
			shared.AssertTrue(t, strings.Contains(text.Plaintext, "code"), "text should contain 'code'")
		})

		t.Run("extracts link facet", func(t *testing.T) {
			markdown := "This is a [link](https://example.com)"
			blocks, err := converter.ToLeaflet(markdown)

			shared.AssertNoError(t, err, "ToLeaflet should succeed")
			text := blocks[0].Block.(TextBlock)

			shared.AssertTrue(t, len(text.Facets) > 0, "should have facets")
			shared.AssertTrue(t, strings.Contains(text.Plaintext, "link"), "text should contain 'link'")

			foundLink := false
			for _, facet := range text.Facets {
				for _, feature := range facet.Features {
					if link, ok := feature.(FacetLink); ok {
						shared.AssertEqual(t, "https://example.com", link.URI, "link URI should match")
						foundLink = true
					}
				}
			}
			shared.AssertTrue(t, foundLink, "should have found link facet")
		})

		t.Run("extracts strikethrough facet", func(t *testing.T) {
			markdown := "This is ~~deleted~~ text"
			blocks, err := converter.ToLeaflet(markdown)

			shared.AssertNoError(t, err, "ToLeaflet should succeed")
			text := blocks[0].Block.(TextBlock)

			shared.AssertTrue(t, len(text.Facets) > 0, "should have facets")
		})

		t.Run("extracts multiple facets", func(t *testing.T) {
			markdown := "This has **bold** and *italic* and `code`"
			blocks, err := converter.ToLeaflet(markdown)

			shared.AssertNoError(t, err, "ToLeaflet should succeed")
			text := blocks[0].Block.(TextBlock)

			shared.AssertTrue(t, len(text.Facets) >= 3, "should have at least 3 facets")
		})

		t.Run("handles overlapping bold and italic", func(t *testing.T) {
			markdown := "***bold and italic***"
			blocks, err := converter.ToLeaflet(markdown)

			shared.AssertNoError(t, err, "ToLeaflet should succeed")
			text := blocks[0].Block.(TextBlock)

			shared.AssertEqual(t, "bold and italic", text.Plaintext, "text should be correct")
			shared.AssertTrue(t, len(text.Facets) > 0, "should have facets")

			facet := text.Facets[0]
			shared.AssertEqual(t, 2, len(facet.Features), "should have 2 features")

			hasBold := false
			hasItalic := false
			for _, feature := range facet.Features {
				switch feature.(type) {
				case FacetBold:
					hasBold = true
				case FacetItalic:
					hasItalic = true
				}
			}
			shared.AssertTrue(t, hasBold, "should have bold feature")
			shared.AssertTrue(t, hasItalic, "should have italic feature")
		})

		t.Run("handles nested bold in italic", func(t *testing.T) {
			markdown := "*italic **and bold** still italic*"
			blocks, err := converter.ToLeaflet(markdown)

			shared.AssertNoError(t, err, "ToLeaflet should succeed")
			text := blocks[0].Block.(TextBlock)

			shared.AssertEqual(t, "italic and bold still italic", text.Plaintext, "text should be correct")
			shared.AssertTrue(t, len(text.Facets) >= 2, "should have multiple facets")

			foundOverlap := false
			for _, facet := range text.Facets {
				if strings.Contains(text.Plaintext[facet.Index.ByteStart:facet.Index.ByteEnd], "and bold") {
					shared.AssertTrue(t, len(facet.Features) >= 2, "overlapping section should have multiple features")
					foundOverlap = true
				}
			}
			shared.AssertTrue(t, foundOverlap, "should find overlapping facet")
		})

		t.Run("handles link with formatting", func(t *testing.T) {
			markdown := "[**bold link**](https://example.com)"
			blocks, err := converter.ToLeaflet(markdown)

			shared.AssertNoError(t, err, "ToLeaflet should succeed")
			text := blocks[0].Block.(TextBlock)

			shared.AssertEqual(t, "bold link", text.Plaintext, "text should be correct")
			shared.AssertTrue(t, len(text.Facets) > 0, "should have facets")

			hasLink := false
			hasBold := false
			for _, facet := range text.Facets {
				for _, feature := range facet.Features {
					switch f := feature.(type) {
					case FacetLink:
						hasLink = true
						shared.AssertEqual(t, "https://example.com", f.URI, "link URI should match")
					case FacetBold:
						hasBold = true
					}
				}
			}
			shared.AssertTrue(t, hasLink, "should have link feature")
			shared.AssertTrue(t, hasBold, "should have bold feature")
		})

		t.Run("handles strikethrough with bold", func(t *testing.T) {
			markdown := "~~**deleted bold**~~"
			blocks, err := converter.ToLeaflet(markdown)

			shared.AssertNoError(t, err, "ToLeaflet should succeed")
			text := blocks[0].Block.(TextBlock)

			shared.AssertEqual(t, "deleted bold", text.Plaintext, "text should be correct")
			shared.AssertTrue(t, len(text.Facets) > 0, "should have facets")

			hasStrike := false
			hasBold := false
			for _, facet := range text.Facets {
				for _, feature := range facet.Features {
					switch feature.(type) {
					case FacetStrikethrough:
						hasStrike = true
					case FacetBold:
						hasBold = true
					}
				}
			}
			shared.AssertTrue(t, hasStrike, "should have strikethrough feature")
			shared.AssertTrue(t, hasBold, "should have bold feature")
		})

		t.Run("handles complex nested formatting", func(t *testing.T) {
			markdown := "*italic **bold and italic** italic*"
			blocks, err := converter.ToLeaflet(markdown)

			shared.AssertNoError(t, err, "ToLeaflet should succeed")
			text := blocks[0].Block.(TextBlock)

			foundBoldItalic := false
			for _, facet := range text.Facets {
				content := text.Plaintext[facet.Index.ByteStart:facet.Index.ByteEnd]
				if strings.Contains(content, "bold and italic") {
					shared.AssertTrue(t, len(facet.Features) >= 2, "nested section should have multiple features")
					foundBoldItalic = true
				}
			}
			shared.AssertTrue(t, foundBoldItalic, "should find nested bold and italic section")
		})
	})

	t.Run("Round-trip Conversion", func(t *testing.T) {
		t.Run("heading round-trip", func(t *testing.T) {
			original := "## Hello World"
			blocks, err := converter.ToLeaflet(original)
			shared.AssertNoError(t, err, "ToLeaflet should succeed")

			markdown, err := converter.FromLeaflet(blocks)
			shared.AssertNoError(t, err, "FromLeaflet should succeed")
			shared.AssertTrue(t, strings.Contains(markdown, "Hello World"), "should contain original text")
			shared.AssertTrue(t, strings.HasPrefix(markdown, "##"), "should have heading markers")
		})

		t.Run("text round-trip", func(t *testing.T) {
			original := "Simple paragraph"
			blocks, err := converter.ToLeaflet(original)
			shared.AssertNoError(t, err, "ToLeaflet should succeed")

			markdown, err := converter.FromLeaflet(blocks)
			shared.AssertNoError(t, err, "FromLeaflet should succeed")
			shared.AssertTrue(t, strings.Contains(markdown, "Simple paragraph"), "should contain original text")
		})

		t.Run("code block round-trip", func(t *testing.T) {
			original := "```go\nfunc test() {}\n```"
			blocks, err := converter.ToLeaflet(original)
			shared.AssertNoError(t, err, "ToLeaflet should succeed")

			markdown, err := converter.FromLeaflet(blocks)
			shared.AssertNoError(t, err, "FromLeaflet should succeed")
			shared.AssertTrue(t, strings.Contains(markdown, "```"), "should have code fences")
			shared.AssertTrue(t, strings.Contains(markdown, "func test"), "should contain code")
		})

		t.Run("list round-trip", func(t *testing.T) {
			original := "- Item 1\n- Item 2"
			blocks, err := converter.ToLeaflet(original)
			shared.AssertNoError(t, err, "ToLeaflet should succeed")

			markdown, err := converter.FromLeaflet(blocks)
			shared.AssertNoError(t, err, "FromLeaflet should succeed")

			shared.AssertTrue(t, strings.Contains(markdown, "Item 1"), "should contain first item")
			shared.AssertTrue(t, strings.Contains(markdown, "Item 2"), "should contain second item")
			shared.AssertTrue(t, strings.Contains(markdown, "-"), "should have list markers")
		})
	})

	t.Run("Edge Cases", func(t *testing.T) {
		t.Run("handles empty markdown", func(t *testing.T) {
			blocks, err := converter.ToLeaflet("")
			shared.AssertNoError(t, err, "should handle empty string")
			shared.AssertEqual(t, 0, len(blocks), "should have no blocks")
		})

		t.Run("skips empty paragraphs", func(t *testing.T) {
			markdown := "\n\n\n"
			blocks, err := converter.ToLeaflet(markdown)
			shared.AssertNoError(t, err, "should succeed")
			shared.AssertEqual(t, 0, len(blocks), "should skip empty paragraphs")
		})

		t.Run("handles special characters", func(t *testing.T) {
			markdown := "Text with *special* characters"
			blocks, err := converter.ToLeaflet(markdown)
			shared.AssertNoError(t, err, "should handle special characters")
			shared.AssertEqual(t, 1, len(blocks), "should have 1 block")

			text := blocks[0].Block.(TextBlock)
			shared.AssertTrue(t, strings.Contains(text.Plaintext, "special"), "should preserve text")
			shared.AssertTrue(t, strings.Contains(text.Plaintext, "characters"), "should preserve text")
		})

		t.Run("handles multiple paragraphs", func(t *testing.T) {
			markdown := "First paragraph\n\nSecond paragraph"
			blocks, err := converter.ToLeaflet(markdown)
			shared.AssertNoError(t, err, "should succeed")
			shared.AssertEqual(t, 2, len(blocks), "should have 2 blocks")
		})
	})

	t.Run("Image Handling", func(t *testing.T) {
		tmpDir := t.TempDir()
		createTestImage := func(t *testing.T, name string, width, height int) string {
			path := filepath.Join(tmpDir, name)

			img := image.NewRGBA(image.Rect(0, 0, width, height))
			f, err := os.Create(path)
			shared.AssertNoError(t, err, "should create image file")
			defer f.Close()

			err = png.Encode(f, img)
			shared.AssertNoError(t, err, "should encode image")

			return path
		}

		t.Run("converts image without resolver (placeholder)", func(t *testing.T) {
			markdown := "![alt text](image.png)"
			converter := NewMarkdownConverter()
			blocks, err := converter.ToLeaflet(markdown)

			shared.AssertNoError(t, err, "ToLeaflet should succeed")
			shared.AssertTrue(t, len(blocks) >= 1, "should have at least 1 block")

			var imgBlock ImageBlock
			found := false
			for _, block := range blocks {
				if img, ok := block.Block.(ImageBlock); ok {
					imgBlock = img
					found = true
					break
				}
			}

			shared.AssertTrue(t, found, "should find image block")
			shared.AssertEqual(t, TypeImageBlock, imgBlock.Type, "type should match")
			shared.AssertEqual(t, "alt text", imgBlock.Alt, "alt text should match")
			shared.AssertEqual(t, "bafkreiplaceholder", imgBlock.Image.Ref.Link, "should have placeholder CID")
		})

		t.Run("resolves local image with dimensions", func(t *testing.T) {
			_ = createTestImage(t, "test.png", 800, 600)
			markdown := "![test image](test.png)"

			resolver := &LocalImageResolver{}
			converter := NewMarkdownConverter().WithImageResolver(resolver, tmpDir)

			blocks, err := converter.ToLeaflet(markdown)
			shared.AssertNoError(t, err, "ToLeaflet should succeed")
			shared.AssertTrue(t, len(blocks) >= 1, "should have at least 1 block")

			var imgBlock ImageBlock
			found := false
			for _, block := range blocks {
				if img, ok := block.Block.(ImageBlock); ok {
					imgBlock = img
					found = true
					break
				}
			}

			shared.AssertTrue(t, found, "should find image block")
			shared.AssertEqual(t, "test image", imgBlock.Alt, "alt text should match")
			shared.AssertEqual(t, 800, imgBlock.AspectRatio.Width, "width should match")
			shared.AssertEqual(t, 600, imgBlock.AspectRatio.Height, "height should match")
			shared.AssertEqual(t, "image/png", imgBlock.Image.MimeType, "mime type should match")
		})

		t.Run("handles inline images in paragraph", func(t *testing.T) {
			_ = createTestImage(t, "inline.png", 100, 100)
			markdown := "Some text before ![inline](inline.png) and text after"

			resolver := &LocalImageResolver{}
			converter := NewMarkdownConverter().WithImageResolver(resolver, tmpDir)

			blocks, err := converter.ToLeaflet(markdown)
			shared.AssertNoError(t, err, "ToLeaflet should succeed")
			shared.AssertTrue(t, len(blocks) >= 2, "should have multiple blocks for inline images")

			textBlock1, ok := blocks[0].Block.(TextBlock)
			shared.AssertTrue(t, ok, "first block should be text")
			shared.AssertTrue(t, strings.Contains(textBlock1.Plaintext, "Some text before"), "should contain text before image")

			imgBlock, ok := blocks[1].Block.(ImageBlock)
			shared.AssertTrue(t, ok, "second block should be image")
			shared.AssertEqual(t, "inline", imgBlock.Alt, "alt text should match")
		})

		t.Run("handles multiple images", func(t *testing.T) {
			_ = createTestImage(t, "img1.png", 200, 150)
			_ = createTestImage(t, "img2.png", 300, 200)
			markdown := "![first](img1.png)\n\n![second](img2.png)"

			resolver := &LocalImageResolver{}
			converter := NewMarkdownConverter().WithImageResolver(resolver, tmpDir)

			blocks, err := converter.ToLeaflet(markdown)
			shared.AssertNoError(t, err, "ToLeaflet should succeed")

			var imageBlocks []ImageBlock
			for _, block := range blocks {
				if img, ok := block.Block.(ImageBlock); ok {
					imageBlocks = append(imageBlocks, img)
				}
			}

			shared.AssertEqual(t, 2, len(imageBlocks), "should have 2 image blocks")
			shared.AssertEqual(t, "first", imageBlocks[0].Alt, "first alt text")
			shared.AssertEqual(t, 200, imageBlocks[0].AspectRatio.Width, "first width")
			shared.AssertEqual(t, "second", imageBlocks[1].Alt, "second alt text")
			shared.AssertEqual(t, 300, imageBlocks[1].AspectRatio.Width, "second width")
		})

		t.Run("uses custom blob uploader", func(t *testing.T) {
			_ = createTestImage(t, "upload.png", 100, 100)
			markdown := "![uploaded](upload.png)"

			uploadCalled := false
			resolver := &LocalImageResolver{
				BlobUploader: func(data []byte, mimeType string) (Blob, error) {
					uploadCalled = true
					shared.AssertEqual(t, "image/png", mimeType, "mime type should be png")
					shared.AssertTrue(t, len(data) > 0, "should have data")

					return Blob{
						Type:     TypeBlob,
						Ref:      CID{Link: "bafkreicustomcid"},
						MimeType: mimeType,
						Size:     len(data),
					}, nil
				},
			}

			converter := NewMarkdownConverter().WithImageResolver(resolver, tmpDir)
			blocks, err := converter.ToLeaflet(markdown)

			shared.AssertNoError(t, err, "ToLeaflet should succeed")
			shared.AssertTrue(t, uploadCalled, "upload should be called")

			imgBlock, ok := blocks[0].Block.(ImageBlock)
			shared.AssertTrue(t, ok, "block should be image")
			shared.AssertEqual(t, "bafkreicustomcid", imgBlock.Image.Ref.Link, "should use custom CID")
		})

		t.Run("handles missing image gracefully", func(t *testing.T) {
			markdown := "![missing](nonexistent.png)"
			resolver := &LocalImageResolver{}
			converter := NewMarkdownConverter().WithImageResolver(resolver, tmpDir)

			_, err := converter.ToLeaflet(markdown)
			shared.AssertError(t, err, "should error on missing image")
			shared.AssertTrue(t, strings.Contains(err.Error(), "failed to resolve image"), "error should mention resolution failure")
		})

		t.Run("gathers images from complex document", func(t *testing.T) {
			_ = createTestImage(t, "header.png", 100, 100)
			_ = createTestImage(t, "body.png", 200, 200)
			_ = createTestImage(t, "list.png", 50, 50)

			markdown := `# Header

![header image](header.png)

Some text with ![inline](body.png) image.

- List item
- Another item with ![list img](list.png)
`

			resolver := &LocalImageResolver{}
			converter := NewMarkdownConverter().WithImageResolver(resolver, tmpDir)

			blocks, err := converter.ToLeaflet(markdown)
			shared.AssertNoError(t, err, "ToLeaflet should succeed")

			imageCount := 0
			for _, block := range blocks {
				if _, ok := block.Block.(ImageBlock); ok {
					imageCount++
				}
			}
			shared.AssertTrue(t, imageCount >= 2, "should find multiple images")
		})

		t.Run("preserves image dimensions accurately", func(t *testing.T) {
			testCases := []struct {
				name   string
				width  int
				height int
			}{
				{"square.png", 100, 100},
				{"landscape.png", 1920, 1080},
				{"portrait.png", 1080, 1920},
				{"wide.png", 2560, 1440},
			}

			for _, tc := range testCases {
				createTestImage(t, tc.name, tc.width, tc.height)
				markdown := "![test](" + tc.name + ")"

				resolver := &LocalImageResolver{}
				converter := NewMarkdownConverter().WithImageResolver(resolver, tmpDir)

				blocks, err := converter.ToLeaflet(markdown)
				shared.AssertNoError(t, err, "should convert "+tc.name)

				imgBlock := blocks[0].Block.(ImageBlock)
				shared.AssertEqual(t, tc.width, imgBlock.AspectRatio.Width, tc.name+" width")
				shared.AssertEqual(t, tc.height, imgBlock.AspectRatio.Height, tc.name+" height")
			}
		})
	})
}
