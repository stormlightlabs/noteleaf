---
title: Books
sidebar_label: Books
description: Build and maintain your reading list with Open Library metadata.
sidebar_position: 2
---

# Books

The book workflow revolves around Open Library search results. Each command lives under `noteleaf media book`.

## Add Books

Search Open Library and pick a result:

```sh
noteleaf media book add "Project Hail Mary"
```

Flags:

- `-i, --interactive`: open the TUI browser (currently shows your local list—useful for triage).
- Plain mode prints the top five matches inline and prompts for a numeric selection.

Behind the scenes Noteleaf records the title, authors, edition details, and any subjects returned by the API. New entries start in the `queued` status.

## Manage the Reading List

List and filter:

```sh
# Everything
noteleaf media book list --all

# Only active reads
noteleaf media book list --reading

# Completed books
noteleaf media book list --finished
```

Each line shows the ID, title, author, status, progress percentage, and any captured metadata (publishers, edition counts, etc.).

Remove items you no longer care about:

```sh
noteleaf media book remove 42
```

## Track Progress

You can explicitly set the status:

```sh
noteleaf media book reading 7
noteleaf media book finished 7
noteleaf media book update 7 queued
```

But the fastest way is to update the percentage:

```sh
noteleaf media book progress 7 45   # Moves status to reading and records start time
noteleaf media book progress 7 100  # Marks finished and records completion time
```

Logic applied automatically:

- `0%` → resets to `queued` and clears the “started” timestamp.
- `1‑99%` → flips to `reading` (start time captured).
- `100%` → marks `finished`, sets end time, and locks progress at 100%.

## Reading Lists and Search

Common workflows:

- **Focus view**: `noteleaf media book list --reading | fzf` to pick the next session book.
- **Backlog grooming**: `noteleaf media book list --queued` to prune items before they go stale.
- **Author sprint**: pipe the list to `rg` to filter by author (`noteleaf media book list --all | rg "Le Guin"`).

The TUI (`noteleaf media book add -i` or `noteleaf media book list` with the `--interactive` switch) supports `/` to search titles/authors/notes live and `v` for a detailed preview with timestamps.

## Metadata and Notes

Each record stores:

- Title & authors (comma separated when multiple).
- Edition count, publishers, subject tags, or cover IDs exposed as inline notes.
- Added/started/finished timestamps.
- Optional page count (if Open Library exposes it).

Use those IDs anywhere else (tasks or notes). Example note snippet:

```markdown
## Reading Log
- 2024-02-01 → Started book #7 ("Project Hail Mary")
- 2024-02-05 → Captured ideas in note #128 linked back to the book.
```

Because media lives in the same database as tasks and notes, full-text search will surface those references instantly.
