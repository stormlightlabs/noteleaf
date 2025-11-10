---
title: Article Overview
sidebar_label: Overview
description: How the article parser saves content for offline reading.
sidebar_position: 1
---

# Article Overview

The `noteleaf article` command turns any supported URL into two files on disk:

- A clean Markdown document (great for terminal reading).
- A rendered HTML copy (handy for rich export or sharing).

Both files live inside your configured `articles_dir` (defaults to `<data_dir>/articles`). The SQLite database stores the metadata and file paths so you can query, list, and delete articles without worrying about directories.

## How Parsing Works

1. **Domain rules first**: Each supported site has a small XPath rule file (`internal/articles/rules/*.txt`).
2. **Heuristic fallback**: When no rule exists, the parser falls back to the readability-style heuristic extractor that scores DOM nodes, removes nav bars, and preserves headings/links.
3. **Metadata extraction**: The parser also looks for OpenGraph/JSON-LD tags to capture author names and publish dates.

You can see the currently loaded rule set by running:

```sh
noteleaf article --help
```

The help output prints the supported domains and the storage directory that is currently in use.

## Saved Metadata

Every article record contains:

- URL and canonical title
- Author (if present in metadata)
- Publication date (stored as plain text, e.g., `2024-01-02`)
- Markdown file path
- HTML file path
- Created/modified timestamps

These fields make it easy to build reading logs, cite sources in notes, or reference articles from tasks.

## Commands at a Glance

| Command                           | Purpose |
|----------------------------------|---------|
| `noteleaf article add <url>`     | Parse, save, and index a URL |
| `noteleaf article list [query]`  | Show saved items; filter with `--author` or `--limit` |
| `noteleaf article view <id>`     | Inspect metadata + a short preview |
| `noteleaf article read <id>`     | Render the Markdown nicely in your terminal |
| `noteleaf article remove <id>`   | Delete the DB entry and the files |

The CLI automatically prevents duplicate imports by checking the URL before parsing.
