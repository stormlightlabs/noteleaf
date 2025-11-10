---
title: Article Management
sidebar_label: Management
description: Save URLs, inspect metadata, and read articles without leaving the CLI.
sidebar_position: 2
---

# Article Management

## Save Articles from URLs

```sh
noteleaf article add https://example.com/long-form-piece
```

What happens:

1. The CLI checks the database to ensure the URL was not imported already.  
2. It fetches the page with a reader-friendly User-Agent (`curl/8.4.0`) and English `Accept-Language` headers to avoid blocked responses.  
3. Parsed content is written to Markdown and HTML files under `articles_dir`.  
4. A database record is inserted with all metadata and file paths.

If parsing fails (unsupported domain, network issue, etc.) nothing is written to disk, so partial entries never appear in your archive.

## Parsing and Extraction

The parser uses a two-layer strategy:

1. **Domain-specific rules** check the XPath selectors defined in `internal/articles/rules`. These rules strip unwanted elements (cookie banners, nav bars), capture the main body, and record author/date fields accurately.  
2. **Heuristic fallback** scores every DOM node, penalizes high link-density sections, and picks the most “article-like” block. It also pulls metadata from JSON-LD `Article` objects when available.

During saving, the Markdown file gets a generated header:

```markdown
# Article Title

**Author:** Jane Smith

**Date:** 2024-01-02

**Source:** https://example.com/long-form-piece

**Saved:** 2024-02-05 10:45:12
```

Everything after the `---` separator is the cleaned article content.

## Reading in the Terminal

There are two ways to inspect what you saved:

- `noteleaf article view <id>` shows metadata, verifies whether the files still exist, and prints the first ~20 lines as a preview.  
- `noteleaf article read <id>` renders the full Markdown using [Charm’s Glamour](https://github.com/charmbracelet/glamour), giving you syntax highlighting, proper headings, and wrapped paragraphs directly in the terminal.

If you prefer your editor, open the Markdown path printed by `view`. Both Markdown and HTML copies belong to you, so feel free to annotate or reformat them.

## Article Metadata Reference

Use `noteleaf article list` to see titles and authors:

```sh
noteleaf article list                 # newest first
noteleaf article list "sqlite"        # full-text filter on titles
noteleaf article list --author "Kim"  # author filter
noteleaf article list -l 5            # top 5 results
```

Each entry includes the created timestamp. The `view` command provides the raw paths so you can script around them, for example:

```sh
md=$(noteleaf article view 12 | rg 'Markdown:' | awk '{print $3}')
$EDITOR "$md"
```

All metadata lives in the SQLite `articles` table, making it easy to run your own reports with `sqlite3` if needed.
