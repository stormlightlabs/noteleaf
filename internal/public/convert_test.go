package public

import (
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
}
