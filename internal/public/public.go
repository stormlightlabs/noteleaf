// Package public defines leaflet publication schema types
//
// These types correspond to the pub.leaflet.* lexicons used by leaflet.pub
//
// The types here match the lexicon definitions from:
//
//	https://github.com/hyperlink-academy/leaflet/tree/main/lexicons/pub/leaflet/
package public

import "time"

const (
	TypeDocument       = "pub.leaflet.document"
	TypePublication    = "pub.leaflet.publication"
	TypeLinearDocument = "pub.leaflet.pages.linearDocument"
	TypeBlock          = "pub.leaflet.pages.linearDocument#block"

	TypeTextBlock           = "pub.leaflet.blocks.text"
	TypeHeaderBlock         = "pub.leaflet.blocks.header"
	TypeCodeBlock           = "pub.leaflet.blocks.code"
	TypeImageBlock          = "pub.leaflet.blocks.image"
	TypeBlockquoteBlock     = "pub.leaflet.blocks.blockquote"
	TypeUnorderedListBlock  = "pub.leaflet.blocks.unorderedList"
	TypeHorizontalRuleBlock = "pub.leaflet.blocks.horizontalRule"

	TypeFacet          = "pub.leaflet.richtext.facet"
	TypeByteSlice      = "pub.leaflet.richtext.facet#byteSlice"
	TypeFacetBold      = "pub.leaflet.richtext.facet#bold"
	TypeFacetItalic    = "pub.leaflet.richtext.facet#italic"
	TypeFacetCode      = "pub.leaflet.richtext.facet#code"
	TypeFacetLink      = "pub.leaflet.richtext.facet#link"
	TypeFacetStrike    = "pub.leaflet.richtext.facet#strikethrough"
	TypeFacetUnderline = "pub.leaflet.richtext.facet#underline"
	TypeFacetHighlight = "pub.leaflet.richtext.facet#highlight"

	TypeListItem    = "pub.leaflet.blocks.unorderedList#listItem"
	TypeAspectRatio = "pub.leaflet.blocks.image#aspectRatio"
	TypeBlob        = "blob"
)

// Document represents a leaflet document (pub.leaflet.document)
type Document struct {
	Type        string           `json:"$type"`
	Author      string           `json:"author"`      // DID (Decentralized Identifier)
	Title       string           `json:"title"`       // Max 128 graphemes
	Description string           `json:"description"` // Max 300 graphemes
	PublishedAt string           `json:"publishedAt"` // ISO8601 datetime
	Publication string           `json:"publication"` // URI: at://did/pub.leaflet.publication/rkey
	Pages       []LinearDocument `json:"pages"`
}

// LinearDocument represents a page in a leaflet document (pub.leaflet.pages.linearDocument)
type LinearDocument struct {
	Type   string      `json:"$type"`
	ID     string      `json:"id,omitempty"`
	Blocks []BlockWrap `json:"blocks"`
}

// BlockWrap wraps a block with optional metadata (alignment, etc.)
type BlockWrap struct {
	Type      string `json:"$type"`
	Block     any    `json:"block"`               // One of: TextBlock, HeaderBlock, etc.
	Alignment string `json:"alignment,omitempty"` // #textAlignLeft, etc.
}

// TextBlock represents a text content block (pub.leaflet.blocks.text)
type TextBlock struct {
	Type      string  `json:"$type"`
	Plaintext string  `json:"plaintext"`
	Facets    []Facet `json:"facets,omitempty"`
}

// HeaderBlock represents a heading content block (pub.leaflet.blocks.header)
type HeaderBlock struct {
	Type      string  `json:"$type"`
	Level     int     `json:"level,omitempty"` // h1 - h6
	Plaintext string  `json:"plaintext"`
	Facets    []Facet `json:"facets,omitempty"`
}

// CodeBlock represents a code content block (pub.leaflet.blocks.code)
type CodeBlock struct {
	Type                    string `json:"$type"`
	Plaintext               string `json:"plaintext"`
	Language                string `json:"language,omitempty"`
	SyntaxHighlightingTheme string `json:"syntaxHighlightingTheme,omitempty"`
}

// ImageBlock represents an image content block (pub.leaflet.blocks.image)
type ImageBlock struct {
	Type        string      `json:"$type"`
	Image       Blob        `json:"image"`
	Alt         string      `json:"alt,omitempty"`
	AspectRatio AspectRatio `json:"aspectRatio"`
}

// AspectRatio represents image dimensions (pub.leaflet.blocks.image#aspectRatio)
type AspectRatio struct {
	Type   string `json:"$type"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// BlockquoteBlock represents a blockquote content block (pub.leaflet.blocks.blockquote)
type BlockquoteBlock struct {
	Type      string  `json:"$type"`
	Plaintext string  `json:"plaintext"`
	Facets    []Facet `json:"facets,omitempty"`
}

// UnorderedListBlock represents an unordered list (pub.leaflet.blocks.unorderedList)
type UnorderedListBlock struct {
	Type     string     `json:"$type"`
	Children []ListItem `json:"children"`
}

// ListItem represents a single list item (pub.leaflet.blocks.unorderedList#listItem)
type ListItem struct {
	Type     string     `json:"$type"`
	Content  any        `json:"content"`            // [TextBlock], [HeaderBlock], [ImageBlock]
	Children []ListItem `json:"children,omitempty"` // Nested list items
}

// HorizontalRuleBlock represents a horizontal rule/thematic break (pub.leaflet.blocks.horizontalRule)
type HorizontalRuleBlock struct {
	Type string `json:"$type"`
}

// Facet represents text annotation (pub.leaflet.richtext.facet)
type Facet struct {
	Type     string         `json:"$type"`
	Index    ByteSlice      `json:"index"`
	Features []FacetFeature `json:"features"`
}

// ByteSlice specifies a substring range using UTF-8 byte offsets (pub.leaflet.richtext.facet#byteSlice)
type ByteSlice struct {
	Type      string `json:"$type"`
	ByteStart int    `json:"byteStart"`
	ByteEnd   int    `json:"byteEnd"`
}

// FacetFeature is a marker interface for facet features
type FacetFeature interface {
	GetFacetType() string
}

// FacetBold represents bold text styling
type FacetBold struct {
	Type string `json:"$type"`
}

func (f FacetBold) GetFacetType() string { return TypeFacetBold }

// FacetItalic represents italic text styling
type FacetItalic struct {
	Type string `json:"$type"`
}

func (f FacetItalic) GetFacetType() string { return TypeFacetItalic }

// FacetCode represents inline code styling
type FacetCode struct {
	Type string `json:"$type"`
}

func (f FacetCode) GetFacetType() string { return TypeFacetCode }

// FacetLink represents a hyperlink
type FacetLink struct {
	Type string `json:"$type"`
	URI  string `json:"uri"`
}

func (f FacetLink) GetFacetType() string { return TypeFacetLink }

// FacetStrikethrough represents strikethrough text styling
type FacetStrikethrough struct {
	Type string `json:"$type"`
}

func (f FacetStrikethrough) GetFacetType() string { return TypeFacetStrike }

// FacetUnderline represents underline text styling
type FacetUnderline struct {
	Type string `json:"$type"`
}

func (f FacetUnderline) GetFacetType() string { return TypeFacetUnderline }

// FacetHighlight represents highlighted text
type FacetHighlight struct {
	Type string `json:"$type"`
}

func (f FacetHighlight) GetFacetType() string { return TypeFacetHighlight }

// Blob represents binary content (images, files)
type Blob struct {
	Type     string `json:"$type"`
	Ref      CID    `json:"ref"`
	MimeType string `json:"mimeType"`
	Size     int    `json:"size"`
}

// CID represents a Content Identifier (IPFS CID)
type CID struct {
	Link string `json:"$link"`
}

// Publication represents a leaflet publication (pub.leaflet.publication)
type Publication struct {
	Type        string    `json:"$type"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

// DocumentMeta holds metadata about a fetched document
type DocumentMeta struct {
	RKey      string    // Record key (TID)
	CID       string    // Content identifier
	URI       string    // Full AT URI
	IsDraft   bool      // Draft vs published
	FetchedAt time.Time // When we fetched it
}
