---
title: Advanced Note Features
sidebar_label: Advanced
description: Search, exports, backlinks, and automation tips.
sidebar_position: 5
---

# Advanced Note Features

## Full-Text Search

While not exposed as a dedicated command, you can search note content using the database:

**Search with grep** (searches file content):

```sh
grep -r "search term" ~/.local/share/noteleaf/notes/
```

**Search titles and metadata**:

```sh
noteleaf note list --static | grep "keyword"
```

Future versions may include built-in full-text search with relevance ranking.

## Note Exports

Export notes to different formats using standard markdown tools:

**Convert to HTML with pandoc**:

```sh
noteleaf note view 1 --format=raw | pandoc -o output.html
```

**Convert to PDF**:

```sh
noteleaf note view 1 --format=raw | pandoc -o output.pdf
```

**Batch export all notes**:

```sh
for note in ~/.local/share/noteleaf/notes/*.md; do
  pandoc "$note" -o "${note%.md}.html"
done
```

## Backlinks and References

Manually create backlinks between notes using markdown links:

```markdown
See also: [[Research on Authentication]] for background
Related: [[API Design Principles]]
```

While Noteleaf doesn't automatically parse or display backlinks yet, this syntax prepares notes for future backlink support and works with tools like Obsidian if you point it at the notes directory.
