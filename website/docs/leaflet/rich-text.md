---
title: Leaflet Rich Text and Blocks
sidebar_label: Rich Text
description: How markdown maps to leaflet blocks and formatting.
sidebar_position: 5
---

# Leaflet Rich Text and Blocks

## Document Structure

Leaflet documents consist of blocks—discrete content units:

**Text Blocks**: Paragraphs of formatted text
**Header Blocks**: Section titles (level 1-6)
**Code Blocks**: Syntax-highlighted code with language annotation
**Quote Blocks**: Blockquotes for citations
**List Blocks**: Ordered or unordered lists
**Rule Blocks**: Horizontal rules for visual separation

## Text Formatting

Text within blocks can have inline formatting called facets:

**Bold**: `**bold text**` → Bold facet
**Italic**: `*italic text*` → Italic facet
**Code**: `` `inline code` `` → Code facet
**Links**: `[text](url)` → Link facet with URL
**Strikethrough**: `~~struck~~` → Strikethrough facet

Multiple formats can be combined:

```markdown
**bold and *italic* text with [a link](https://example.com)**
```

## Code Blocks

Code blocks preserve language information for syntax highlighting:

````markdown
```python
def hello():
    print("Hello, leaflet!")
```
````

Converts to a code block with language="python".

Supported languages: Any language identifier is preserved, but rendering depends on leaflet.pub's syntax highlighter support.

## Blockquotes

Markdown blockquotes become quote blocks:

```markdown
> This is a quote from another source.
> It can span multiple lines.
```

Nested blockquotes are flattened (leaflet doesn't support nesting).

## Lists

Both ordered and unordered lists are supported:

```markdown
- Unordered item 1
- Unordered item 2
  - Nested item

1. Ordered item 1
2. Ordered item 2
   1. Nested ordered item
```

Nesting is preserved up to leaflet's limits.

## Horizontal Rules

Markdown horizontal rules become rule blocks:

```markdown
---
```

Use for section breaks.

## Images and Media

**Current status**: Image support is not yet implemented in the Noteleaf-to-leaflet converter.

**Future plans**: Images will be uploaded to blob storage and embedded in documents with image blocks.

**Workaround**: For now, images in markdown are either skipped or converted to links.
