package public

import (
	"encoding/json"
	"testing"

	"github.com/stormlightlabs/noteleaf/internal/shared"
)

func TestBlockWrap(t *testing.T) {
	t.Run("UnmarshalJSON", func(t *testing.T) {
		t.Run("unmarshals text block", func(t *testing.T) {
			jsonData := `{
				"$type": "pub.leaflet.pages.linearDocument#block",
				"block": {
					"$type": "pub.leaflet.blocks.text",
					"plaintext": "Hello world"
				}
			}`

			var bw BlockWrap
			err := json.Unmarshal([]byte(jsonData), &bw)
			shared.AssertNoError(t, err, "unmarshal should succeed")
			shared.AssertEqual(t, TypeBlock, bw.Type, "type should match")

			block, ok := bw.Block.(TextBlock)
			shared.AssertTrue(t, ok, "block should be TextBlock")
			shared.AssertEqual(t, TypeTextBlock, block.Type, "block type should match")
			shared.AssertEqual(t, "Hello world", block.Plaintext, "plaintext should match")
		})

		t.Run("unmarshals header block", func(t *testing.T) {
			jsonData := `{
				"$type": "pub.leaflet.pages.linearDocument#block",
				"block": {
					"$type": "pub.leaflet.blocks.header",
					"level": 2,
					"plaintext": "Section Title"
				}
			}`

			var bw BlockWrap
			err := json.Unmarshal([]byte(jsonData), &bw)
			shared.AssertNoError(t, err, "unmarshal should succeed")

			block, ok := bw.Block.(HeaderBlock)
			shared.AssertTrue(t, ok, "block should be HeaderBlock")
			shared.AssertEqual(t, TypeHeaderBlock, block.Type, "block type should match")
			shared.AssertEqual(t, 2, block.Level, "level should match")
			shared.AssertEqual(t, "Section Title", block.Plaintext, "plaintext should match")
		})

		t.Run("unmarshals code block", func(t *testing.T) {
			jsonData := `{
				"$type": "pub.leaflet.pages.linearDocument#block",
				"block": {
					"$type": "pub.leaflet.blocks.code",
					"plaintext": "fmt.Println(\"test\")",
					"language": "go"
				}
			}`

			var bw BlockWrap
			err := json.Unmarshal([]byte(jsonData), &bw)
			shared.AssertNoError(t, err, "unmarshal should succeed")

			block, ok := bw.Block.(CodeBlock)
			shared.AssertTrue(t, ok, "block should be CodeBlock")
			shared.AssertEqual(t, TypeCodeBlock, block.Type, "block type should match")
			shared.AssertEqual(t, "go", block.Language, "language should match")
			shared.AssertEqual(t, "fmt.Println(\"test\")", block.Plaintext, "plaintext should match")
		})

		t.Run("unmarshals image block", func(t *testing.T) {
			jsonData := `{
				"$type": "pub.leaflet.pages.linearDocument#block",
				"block": {
					"$type": "pub.leaflet.blocks.image",
					"image": {
						"$type": "blob",
						"ref": {
							"$link": "bafytest123"
						},
						"mimeType": "image/png",
						"size": 1024
					},
					"alt": "Test image",
					"aspectRatio": {
						"$type": "pub.leaflet.blocks.image#aspectRatio",
						"width": 800,
						"height": 600
					}
				}
			}`

			var bw BlockWrap
			err := json.Unmarshal([]byte(jsonData), &bw)
			shared.AssertNoError(t, err, "unmarshal should succeed")

			block, ok := bw.Block.(ImageBlock)
			shared.AssertTrue(t, ok, "block should be ImageBlock")
			shared.AssertEqual(t, TypeImageBlock, block.Type, "block type should match")
			shared.AssertEqual(t, "Test image", block.Alt, "alt text should match")
			shared.AssertEqual(t, 800, block.AspectRatio.Width, "width should match")
			shared.AssertEqual(t, 600, block.AspectRatio.Height, "height should match")
			shared.AssertEqual(t, "image/png", block.Image.MimeType, "mime type should match")
			shared.AssertEqual(t, 1024, block.Image.Size, "size should match")
		})

		t.Run("unmarshals blockquote block", func(t *testing.T) {
			jsonData := `{
				"$type": "pub.leaflet.pages.linearDocument#block",
				"block": {
					"$type": "pub.leaflet.blocks.blockquote",
					"plaintext": "This is a quote"
				}
			}`

			var bw BlockWrap
			err := json.Unmarshal([]byte(jsonData), &bw)
			shared.AssertNoError(t, err, "unmarshal should succeed")

			block, ok := bw.Block.(BlockquoteBlock)
			shared.AssertTrue(t, ok, "block should be BlockquoteBlock")
			shared.AssertEqual(t, TypeBlockquoteBlock, block.Type, "block type should match")
			shared.AssertEqual(t, "This is a quote", block.Plaintext, "plaintext should match")
		})

		t.Run("unmarshals unordered list block", func(t *testing.T) {
			jsonData := `{
				"$type": "pub.leaflet.pages.linearDocument#block",
				"block": {
					"$type": "pub.leaflet.blocks.unorderedList",
					"children": [
						{
							"$type": "pub.leaflet.blocks.unorderedList#listItem",
							"content": {
								"$type": "pub.leaflet.blocks.text",
								"plaintext": "First item"
							}
						}
					]
				}
			}`

			var bw BlockWrap
			err := json.Unmarshal([]byte(jsonData), &bw)
			shared.AssertNoError(t, err, "unmarshal should succeed")

			block, ok := bw.Block.(UnorderedListBlock)
			shared.AssertTrue(t, ok, "block should be UnorderedListBlock")
			shared.AssertEqual(t, TypeUnorderedListBlock, block.Type, "block type should match")
			shared.AssertEqual(t, 1, len(block.Children), "should have 1 child")
		})

		t.Run("unmarshals horizontal rule block", func(t *testing.T) {
			jsonData := `{
				"$type": "pub.leaflet.pages.linearDocument#block",
				"block": {
					"$type": "pub.leaflet.blocks.horizontalRule"
				}
			}`

			var bw BlockWrap
			err := json.Unmarshal([]byte(jsonData), &bw)
			shared.AssertNoError(t, err, "unmarshal should succeed")

			block, ok := bw.Block.(HorizontalRuleBlock)
			shared.AssertTrue(t, ok, "block should be HorizontalRuleBlock")
			shared.AssertEqual(t, TypeHorizontalRuleBlock, block.Type, "block type should match")
		})

		t.Run("unmarshals unknown block type as map", func(t *testing.T) {
			jsonData := `{
				"$type": "pub.leaflet.pages.linearDocument#block",
				"block": {
					"$type": "pub.leaflet.blocks.unknown",
					"customField": "value"
				}
			}`

			var bw BlockWrap
			err := json.Unmarshal([]byte(jsonData), &bw)
			shared.AssertNoError(t, err, "unmarshal should succeed")

			block, ok := bw.Block.(map[string]any)
			shared.AssertTrue(t, ok, "block should be map for unknown type")
			shared.AssertEqual(t, "pub.leaflet.blocks.unknown", block["$type"], "type should be preserved")
		})

		t.Run("handles block with alignment", func(t *testing.T) {
			jsonData := `{
				"$type": "pub.leaflet.pages.linearDocument#block",
				"block": {
					"$type": "pub.leaflet.blocks.text",
					"plaintext": "Centered text"
				},
				"alignment": "#textAlignCenter"
			}`

			var bw BlockWrap
			err := json.Unmarshal([]byte(jsonData), &bw)
			shared.AssertNoError(t, err, "unmarshal should succeed")
			shared.AssertEqual(t, "#textAlignCenter", bw.Alignment, "alignment should match")
		})

		t.Run("returns error for invalid JSON", func(t *testing.T) {
			jsonData := `{invalid json`

			var bw BlockWrap
			err := json.Unmarshal([]byte(jsonData), &bw)
			shared.AssertError(t, err, "invalid JSON should return error")
		})

		t.Run("returns error for malformed block", func(t *testing.T) {
			jsonData := `{
				"$type": "pub.leaflet.pages.linearDocument#block",
				"block": "not an object"
			}`

			var bw BlockWrap
			err := json.Unmarshal([]byte(jsonData), &bw)
			shared.AssertError(t, err, "malformed block should return error")
		})
	})
}

func TestListItem(t *testing.T) {
	t.Run("UnmarshalJSON", func(t *testing.T) {
		t.Run("unmarshals text block content", func(t *testing.T) {
			jsonData := `{
				"$type": "pub.leaflet.blocks.unorderedList#listItem",
				"content": {
					"$type": "pub.leaflet.blocks.text",
					"plaintext": "List item text"
				}
			}`

			var li ListItem
			err := json.Unmarshal([]byte(jsonData), &li)
			shared.AssertNoError(t, err, "unmarshal should succeed")
			shared.AssertEqual(t, TypeListItem, li.Type, "type should match")

			content, ok := li.Content.(TextBlock)
			shared.AssertTrue(t, ok, "content should be TextBlock")
			shared.AssertEqual(t, "List item text", content.Plaintext, "plaintext should match")
		})

		t.Run("unmarshals header block content", func(t *testing.T) {
			jsonData := `{
				"$type": "pub.leaflet.blocks.unorderedList#listItem",
				"content": {
					"$type": "pub.leaflet.blocks.header",
					"level": 3,
					"plaintext": "List header"
				}
			}`

			var li ListItem
			err := json.Unmarshal([]byte(jsonData), &li)
			shared.AssertNoError(t, err, "unmarshal should succeed")

			content, ok := li.Content.(HeaderBlock)
			shared.AssertTrue(t, ok, "content should be HeaderBlock")
			shared.AssertEqual(t, 3, content.Level, "level should match")
		})

		t.Run("unmarshals image block content", func(t *testing.T) {
			jsonData := `{
				"$type": "pub.leaflet.blocks.unorderedList#listItem",
				"content": {
					"$type": "pub.leaflet.blocks.image",
					"image": {
						"$type": "blob",
						"ref": {"$link": "cid123"},
						"mimeType": "image/jpeg",
						"size": 2048
					},
					"alt": "List image",
					"aspectRatio": {
						"$type": "pub.leaflet.blocks.image#aspectRatio",
						"width": 400,
						"height": 300
					}
				}
			}`

			var li ListItem
			err := json.Unmarshal([]byte(jsonData), &li)
			shared.AssertNoError(t, err, "unmarshal should succeed")

			content, ok := li.Content.(ImageBlock)
			shared.AssertTrue(t, ok, "content should be ImageBlock")
			shared.AssertEqual(t, "List image", content.Alt, "alt should match")
		})

		t.Run("unmarshals nested list items", func(t *testing.T) {
			jsonData := `{
				"$type": "pub.leaflet.blocks.unorderedList#listItem",
				"content": {
					"$type": "pub.leaflet.blocks.text",
					"plaintext": "Parent item"
				},
				"children": [
					{
						"$type": "pub.leaflet.blocks.unorderedList#listItem",
						"content": {
							"$type": "pub.leaflet.blocks.text",
							"plaintext": "Child item"
						}
					}
				]
			}`

			var li ListItem
			err := json.Unmarshal([]byte(jsonData), &li)
			shared.AssertNoError(t, err, "unmarshal should succeed")
			shared.AssertEqual(t, 1, len(li.Children), "should have 1 child")

			childContent, ok := li.Children[0].Content.(TextBlock)
			shared.AssertTrue(t, ok, "child content should be TextBlock")
			shared.AssertEqual(t, "Child item", childContent.Plaintext, "child plaintext should match")
		})

		t.Run("unmarshals unknown content type as map", func(t *testing.T) {
			jsonData := `{
				"$type": "pub.leaflet.blocks.unorderedList#listItem",
				"content": {
					"$type": "pub.leaflet.blocks.custom",
					"data": "value"
				}
			}`

			var li ListItem
			err := json.Unmarshal([]byte(jsonData), &li)
			shared.AssertNoError(t, err, "unmarshal should succeed")

			content, ok := li.Content.(map[string]any)
			shared.AssertTrue(t, ok, "unknown content should be map")
			shared.AssertEqual(t, "pub.leaflet.blocks.custom", content["$type"], "type should be preserved")
		})

		t.Run("returns error for invalid JSON", func(t *testing.T) {
			jsonData := `{invalid`

			var li ListItem
			err := json.Unmarshal([]byte(jsonData), &li)
			shared.AssertError(t, err, "invalid JSON should return error")
		})
	})
}

func TestFacet(t *testing.T) {
	t.Run("UnmarshalJSON", func(t *testing.T) {
		t.Run("unmarshals bold facet", func(t *testing.T) {
			jsonData := `{
				"$type": "pub.leaflet.richtext.facet",
				"index": {
					"$type": "pub.leaflet.richtext.facet#byteSlice",
					"byteStart": 0,
					"byteEnd": 5
				},
				"features": [
					{
						"$type": "pub.leaflet.richtext.facet#bold"
					}
				]
			}`

			var f Facet
			err := json.Unmarshal([]byte(jsonData), &f)
			shared.AssertNoError(t, err, "unmarshal should succeed")
			shared.AssertEqual(t, TypeFacet, f.Type, "type should match")
			shared.AssertEqual(t, 0, f.Index.ByteStart, "byte start should match")
			shared.AssertEqual(t, 5, f.Index.ByteEnd, "byte end should match")
			shared.AssertEqual(t, 1, len(f.Features), "should have 1 feature")

			bold, ok := f.Features[0].(FacetBold)
			shared.AssertTrue(t, ok, "feature should be FacetBold")
			shared.AssertEqual(t, TypeFacetBold, bold.GetFacetType(), "facet type should match")
		})

		t.Run("unmarshals italic facet", func(t *testing.T) {
			jsonData := `{
				"$type": "pub.leaflet.richtext.facet",
				"index": {
					"$type": "pub.leaflet.richtext.facet#byteSlice",
					"byteStart": 5,
					"byteEnd": 10
				},
				"features": [
					{
						"$type": "pub.leaflet.richtext.facet#italic"
					}
				]
			}`

			var f Facet
			err := json.Unmarshal([]byte(jsonData), &f)
			shared.AssertNoError(t, err, "unmarshal should succeed")

			italic, ok := f.Features[0].(FacetItalic)
			shared.AssertTrue(t, ok, "feature should be FacetItalic")
			shared.AssertEqual(t, TypeFacetItalic, italic.GetFacetType(), "facet type should match")
		})

		t.Run("unmarshals code facet", func(t *testing.T) {
			jsonData := `{
				"$type": "pub.leaflet.richtext.facet",
				"index": {
					"$type": "pub.leaflet.richtext.facet#byteSlice",
					"byteStart": 0,
					"byteEnd": 10
				},
				"features": [
					{
						"$type": "pub.leaflet.richtext.facet#code"
					}
				]
			}`

			var f Facet
			err := json.Unmarshal([]byte(jsonData), &f)
			shared.AssertNoError(t, err, "unmarshal should succeed")

			code, ok := f.Features[0].(FacetCode)
			shared.AssertTrue(t, ok, "feature should be FacetCode")
			shared.AssertEqual(t, TypeFacetCode, code.GetFacetType(), "facet type should match")
		})

		t.Run("unmarshals link facet", func(t *testing.T) {
			jsonData := `{
				"$type": "pub.leaflet.richtext.facet",
				"index": {
					"$type": "pub.leaflet.richtext.facet#byteSlice",
					"byteStart": 0,
					"byteEnd": 15
				},
				"features": [
					{
						"$type": "pub.leaflet.richtext.facet#link",
						"uri": "https://example.com"
					}
				]
			}`

			var f Facet
			err := json.Unmarshal([]byte(jsonData), &f)
			shared.AssertNoError(t, err, "unmarshal should succeed")

			link, ok := f.Features[0].(FacetLink)
			shared.AssertTrue(t, ok, "feature should be FacetLink")
			shared.AssertEqual(t, TypeFacetLink, link.GetFacetType(), "facet type should match")
			shared.AssertEqual(t, "https://example.com", link.URI, "URI should match")
		})

		t.Run("unmarshals strikethrough facet", func(t *testing.T) {
			jsonData := `{
				"$type": "pub.leaflet.richtext.facet",
				"index": {
					"$type": "pub.leaflet.richtext.facet#byteSlice",
					"byteStart": 0,
					"byteEnd": 8
				},
				"features": [
					{
						"$type": "pub.leaflet.richtext.facet#strikethrough"
					}
				]
			}`

			var f Facet
			err := json.Unmarshal([]byte(jsonData), &f)
			shared.AssertNoError(t, err, "unmarshal should succeed")

			strike, ok := f.Features[0].(FacetStrikethrough)
			shared.AssertTrue(t, ok, "feature should be FacetStrikethrough")
			shared.AssertEqual(t, TypeFacetStrike, strike.GetFacetType(), "facet type should match")
		})

		t.Run("unmarshals underline facet", func(t *testing.T) {
			jsonData := `{
				"$type": "pub.leaflet.richtext.facet",
				"index": {
					"$type": "pub.leaflet.richtext.facet#byteSlice",
					"byteStart": 0,
					"byteEnd": 12
				},
				"features": [
					{
						"$type": "pub.leaflet.richtext.facet#underline"
					}
				]
			}`

			var f Facet
			err := json.Unmarshal([]byte(jsonData), &f)
			shared.AssertNoError(t, err, "unmarshal should succeed")

			underline, ok := f.Features[0].(FacetUnderline)
			shared.AssertTrue(t, ok, "feature should be FacetUnderline")
			shared.AssertEqual(t, TypeFacetUnderline, underline.GetFacetType(), "facet type should match")
		})

		t.Run("unmarshals highlight facet", func(t *testing.T) {
			jsonData := `{
				"$type": "pub.leaflet.richtext.facet",
				"index": {
					"$type": "pub.leaflet.richtext.facet#byteSlice",
					"byteStart": 5,
					"byteEnd": 15
				},
				"features": [
					{
						"$type": "pub.leaflet.richtext.facet#highlight"
					}
				]
			}`

			var f Facet
			err := json.Unmarshal([]byte(jsonData), &f)
			shared.AssertNoError(t, err, "unmarshal should succeed")

			highlight, ok := f.Features[0].(FacetHighlight)
			shared.AssertTrue(t, ok, "feature should be FacetHighlight")
			shared.AssertEqual(t, TypeFacetHighlight, highlight.GetFacetType(), "facet type should match")
		})

		t.Run("unmarshals multiple features", func(t *testing.T) {
			jsonData := `{
				"$type": "pub.leaflet.richtext.facet",
				"index": {
					"$type": "pub.leaflet.richtext.facet#byteSlice",
					"byteStart": 0,
					"byteEnd": 10
				},
				"features": [
					{
						"$type": "pub.leaflet.richtext.facet#bold"
					},
					{
						"$type": "pub.leaflet.richtext.facet#italic"
					},
					{
						"$type": "pub.leaflet.richtext.facet#link",
						"uri": "https://test.com"
					}
				]
			}`

			var f Facet
			err := json.Unmarshal([]byte(jsonData), &f)
			shared.AssertNoError(t, err, "unmarshal should succeed")
			shared.AssertEqual(t, 3, len(f.Features), "should have 3 features")

			_, isBold := f.Features[0].(FacetBold)
			shared.AssertTrue(t, isBold, "first feature should be bold")

			_, isItalic := f.Features[1].(FacetItalic)
			shared.AssertTrue(t, isItalic, "second feature should be italic")

			link, isLink := f.Features[2].(FacetLink)
			shared.AssertTrue(t, isLink, "third feature should be link")
			shared.AssertEqual(t, "https://test.com", link.URI, "link URI should match")
		})

		t.Run("skips unknown feature types", func(t *testing.T) {
			jsonData := `{
				"$type": "pub.leaflet.richtext.facet",
				"index": {
					"$type": "pub.leaflet.richtext.facet#byteSlice",
					"byteStart": 0,
					"byteEnd": 10
				},
				"features": [
					{
						"$type": "pub.leaflet.richtext.facet#bold"
					},
					{
						"$type": "pub.leaflet.richtext.facet#unknown"
					},
					{
						"$type": "pub.leaflet.richtext.facet#italic"
					}
				]
			}`

			var f Facet
			err := json.Unmarshal([]byte(jsonData), &f)
			shared.AssertNoError(t, err, "unmarshal should succeed")
			shared.AssertEqual(t, 2, len(f.Features), "unknown features should be skipped")
		})

		t.Run("returns error for invalid JSON", func(t *testing.T) {
			jsonData := `{invalid`

			var f Facet
			err := json.Unmarshal([]byte(jsonData), &f)
			shared.AssertError(t, err, "invalid JSON should return error")
		})

		t.Run("returns error for malformed feature", func(t *testing.T) {
			jsonData := `{
				"$type": "pub.leaflet.richtext.facet",
				"index": {
					"$type": "pub.leaflet.richtext.facet#byteSlice",
					"byteStart": 0,
					"byteEnd": 10
				},
				"features": [
					"not an object"
				]
			}`

			var f Facet
			err := json.Unmarshal([]byte(jsonData), &f)
			shared.AssertError(t, err, "malformed feature should return error")
		})
	})
}

func TestDocument(t *testing.T) {
	t.Run("marshals and unmarshals correctly", func(t *testing.T) {
		doc := Document{
			Type:        TypeDocument,
			Author:      "did:plc:test123",
			Title:       "Test Document",
			Description: "A test document",
			PublishedAt: "2024-01-01T00:00:00Z",
			Publication: "at://did:plc:test123/pub.leaflet.publication/rkey",
			Pages: []LinearDocument{
				{
					Type: TypeLinearDocument,
					ID:   "page1",
					Blocks: []BlockWrap{
						{
							Type: TypeBlock,
							Block: TextBlock{
								Type:      TypeTextBlock,
								Plaintext: "Test content",
							},
						},
					},
				},
			},
		}

		data, err := json.Marshal(doc)
		shared.AssertNoError(t, err, "marshal should succeed")

		var decoded Document
		err = json.Unmarshal(data, &decoded)
		shared.AssertNoError(t, err, "unmarshal should succeed")
		shared.AssertEqual(t, doc.Title, decoded.Title, "title should match")
		shared.AssertEqual(t, doc.Author, decoded.Author, "author should match")
		shared.AssertEqual(t, 1, len(decoded.Pages), "should have 1 page")
	})
}

func TestLinearDocument(t *testing.T) {
	t.Run("marshals with multiple block types", func(t *testing.T) {
		ld := LinearDocument{
			Type: TypeLinearDocument,
			ID:   "page1",
			Blocks: []BlockWrap{
				{
					Type: TypeBlock,
					Block: TextBlock{
						Type:      TypeTextBlock,
						Plaintext: "Text",
					},
				},
				{
					Type: TypeBlock,
					Block: HeaderBlock{
						Type:      TypeHeaderBlock,
						Level:     1,
						Plaintext: "Header",
					},
				},
			},
		}

		data, err := json.Marshal(ld)
		shared.AssertNoError(t, err, "marshal should succeed")

		var decoded LinearDocument
		err = json.Unmarshal(data, &decoded)
		shared.AssertNoError(t, err, "unmarshal should succeed")
		shared.AssertEqual(t, 2, len(decoded.Blocks), "should have 2 blocks")
	})
}

func TestFacetFeatures(t *testing.T) {
	t.Run("GetFacetType", func(t *testing.T) {
		t.Run("returns correct type for bold", func(t *testing.T) {
			bold := FacetBold{Type: TypeFacetBold}
			shared.AssertEqual(t, TypeFacetBold, bold.GetFacetType(), "type should match")
		})

		t.Run("returns correct type for italic", func(t *testing.T) {
			italic := FacetItalic{Type: TypeFacetItalic}
			shared.AssertEqual(t, TypeFacetItalic, italic.GetFacetType(), "type should match")
		})

		t.Run("returns correct type for code", func(t *testing.T) {
			code := FacetCode{Type: TypeFacetCode}
			shared.AssertEqual(t, TypeFacetCode, code.GetFacetType(), "type should match")
		})

		t.Run("returns correct type for link", func(t *testing.T) {
			link := FacetLink{Type: TypeFacetLink, URI: "https://example.com"}
			shared.AssertEqual(t, TypeFacetLink, link.GetFacetType(), "type should match")
		})

		t.Run("returns correct type for strikethrough", func(t *testing.T) {
			strike := FacetStrikethrough{Type: TypeFacetStrike}
			shared.AssertEqual(t, TypeFacetStrike, strike.GetFacetType(), "type should match")
		})

		t.Run("returns correct type for underline", func(t *testing.T) {
			underline := FacetUnderline{Type: TypeFacetUnderline}
			shared.AssertEqual(t, TypeFacetUnderline, underline.GetFacetType(), "type should match")
		})

		t.Run("returns correct type for highlight", func(t *testing.T) {
			highlight := FacetHighlight{Type: TypeFacetHighlight}
			shared.AssertEqual(t, TypeFacetHighlight, highlight.GetFacetType(), "type should match")
		})
	})
}

func TestBlob(t *testing.T) {
	t.Run("marshals and unmarshals correctly", func(t *testing.T) {
		blob := Blob{
			Type: TypeBlob,
			Ref: CID{
				Link: "bafytest123",
			},
			MimeType: "image/png",
			Size:     4096,
		}

		data, err := json.Marshal(blob)
		shared.AssertNoError(t, err, "marshal should succeed")

		var decoded Blob
		err = json.Unmarshal(data, &decoded)
		shared.AssertNoError(t, err, "unmarshal should succeed")
		shared.AssertEqual(t, blob.MimeType, decoded.MimeType, "mime type should match")
		shared.AssertEqual(t, blob.Size, decoded.Size, "size should match")
		shared.AssertEqual(t, blob.Ref.Link, decoded.Ref.Link, "CID link should match")
	})
}

func TestPublication(t *testing.T) {
	t.Run("marshals and unmarshals correctly", func(t *testing.T) {
		pub := Publication{
			Type:        TypePublication,
			Name:        "Test Publication",
			Description: "A test publication",
		}

		data, err := json.Marshal(pub)
		shared.AssertNoError(t, err, "marshal should succeed")

		var decoded Publication
		err = json.Unmarshal(data, &decoded)
		shared.AssertNoError(t, err, "unmarshal should succeed")
		shared.AssertEqual(t, pub.Name, decoded.Name, "name should match")
		shared.AssertEqual(t, pub.Description, decoded.Description, "description should match")
	})
}

func TestComplexDocument(t *testing.T) {
	t.Run("unmarshals complex nested document", func(t *testing.T) {
		jsonData := `{
			"$type": "pub.leaflet.document",
			"author": "did:plc:abc123",
			"title": "Complex Document",
			"description": "Testing complex structures",
			"publishedAt": "2024-01-15T10:30:00Z",
			"publication": "at://did:plc:abc123/pub.leaflet.publication/xyz",
			"pages": [
				{
					"$type": "pub.leaflet.pages.linearDocument",
					"id": "page1",
					"blocks": [
						{
							"$type": "pub.leaflet.pages.linearDocument#block",
							"block": {
								"$type": "pub.leaflet.blocks.header",
								"level": 1,
								"plaintext": "Introduction",
								"facets": [
									{
										"$type": "pub.leaflet.richtext.facet",
										"index": {
											"$type": "pub.leaflet.richtext.facet#byteSlice",
											"byteStart": 0,
											"byteEnd": 12
										},
										"features": [
											{
												"$type": "pub.leaflet.richtext.facet#bold"
											}
										]
									}
								]
							}
						},
						{
							"$type": "pub.leaflet.pages.linearDocument#block",
							"block": {
								"$type": "pub.leaflet.blocks.text",
								"plaintext": "This is a link to example",
								"facets": [
									{
										"$type": "pub.leaflet.richtext.facet",
										"index": {
											"$type": "pub.leaflet.richtext.facet#byteSlice",
											"byteStart": 10,
											"byteEnd": 14
										},
										"features": [
											{
												"$type": "pub.leaflet.richtext.facet#link",
												"uri": "https://example.com"
											}
										]
									}
								]
							}
						},
						{
							"$type": "pub.leaflet.pages.linearDocument#block",
							"block": {
								"$type": "pub.leaflet.blocks.unorderedList",
								"children": [
									{
										"$type": "pub.leaflet.blocks.unorderedList#listItem",
										"content": {
											"$type": "pub.leaflet.blocks.text",
											"plaintext": "First item"
										},
										"children": [
											{
												"$type": "pub.leaflet.blocks.unorderedList#listItem",
												"content": {
													"$type": "pub.leaflet.blocks.text",
													"plaintext": "Nested item"
												}
											}
										]
									}
								]
							}
						},
						{
							"$type": "pub.leaflet.pages.linearDocument#block",
							"block": {
								"$type": "pub.leaflet.blocks.horizontalRule"
							}
						}
					]
				}
			]
		}`

		var doc Document
		err := json.Unmarshal([]byte(jsonData), &doc)
		shared.AssertNoError(t, err, "unmarshal should succeed")
		shared.AssertEqual(t, TypeDocument, doc.Type, "type should match")
		shared.AssertEqual(t, "Complex Document", doc.Title, "title should match")
		shared.AssertEqual(t, 1, len(doc.Pages), "should have 1 page")
		shared.AssertEqual(t, 4, len(doc.Pages[0].Blocks), "should have 4 blocks")

		headerBlock, ok := doc.Pages[0].Blocks[0].Block.(HeaderBlock)
		shared.AssertTrue(t, ok, "first block should be HeaderBlock")
		shared.AssertEqual(t, 1, headerBlock.Level, "header level should be 1")
		shared.AssertEqual(t, 1, len(headerBlock.Facets), "header should have 1 facet")

		textBlock, ok := doc.Pages[0].Blocks[1].Block.(TextBlock)
		shared.AssertTrue(t, ok, "second block should be TextBlock")
		shared.AssertEqual(t, 1, len(textBlock.Facets), "text should have 1 facet")
		link, ok := textBlock.Facets[0].Features[0].(FacetLink)
		shared.AssertTrue(t, ok, "facet feature should be link")
		shared.AssertEqual(t, "https://example.com", link.URI, "link URI should match")

		listBlock, ok := doc.Pages[0].Blocks[2].Block.(UnorderedListBlock)
		shared.AssertTrue(t, ok, "third block should be UnorderedListBlock")
		shared.AssertEqual(t, 1, len(listBlock.Children), "list should have 1 child")
		shared.AssertEqual(t, 1, len(listBlock.Children[0].Children), "first item should have 1 nested child")

		hrBlock, ok := doc.Pages[0].Blocks[3].Block.(HorizontalRuleBlock)
		shared.AssertTrue(t, ok, "fourth block should be HorizontalRuleBlock")
		shared.AssertEqual(t, TypeHorizontalRuleBlock, hrBlock.Type, "HR type should match")
	})
}
