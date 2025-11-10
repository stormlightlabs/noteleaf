---
title: Media Overview
sidebar_label: Overview
description: Manage reading lists and watch queues from the CLI.
sidebar_position: 1
---

# Media Tracking Overview

Noteleaf keeps book, movie, and TV data next to your tasks and notes so you do not need a separate “watch list” app.
All media commands hang off a single entry point:

```sh
noteleaf media <book|movie|tv> <subcommand>
```

- **Books** pull metadata from the Open Library API.
- **Movies/TV** scrape Rotten Tomatoes search results to capture critic scores and canonical titles.
- Everything is stored in the local SQLite database located in your data directory (`~/.local/share/noteleaf` on Linux, `~/Library/Application Support/noteleaf` on macOS, `%LOCALAPPDATA%\noteleaf` on Windows).

## Lifecycle Statuses

| Type   | Statuses                                   | Notes                                                                                             |
| ------ | ------------------------------------------ | ------------------------------------------------------------------------------------------------- |
| Books  | `queued`, `reading`, `finished`, `removed` | Progress updates automatically bump status (0% → `queued`, 1-99% → `reading`, 100% → `finished`). |
| Movies | `queued`, `watched`, `removed`             | Marking as watched stores the completion timestamp.                                               |
| TV     | `queued`, `watching`, `watched`, `removed` | Watching/watched commands also record the last watched time.                                      |

Statuses control list filtering and show up beside each item in the TUI.

## Metadata That Gets Saved

- **Books**: title, authors, Open Library notes (editions, publishers, subjects), started/finished timestamps, progress percentage.
- **Movies**: release year when available, Rotten Tomatoes critic score details inside the notes field, watched timestamp.
- **TV**: show title plus critic score details, optional season/episode numbers, last watched timestamp.

You can safely edit the generated markdown/notes in your favorite editor—the records keep pointing to the updated files.

## Interactive vs Static Workflows

All `list` commands default to a simple textual table. For books you can pass `-i/--interactive` to open the Bubble Tea list browser (TV and movie interactive selectors are planned). Inside the list view:

- `j/k` or arrow keys move between entries.
- `/` starts search across titles, authors, and metadata.
- `v` opens a focused preview.
- `?` shows all shortcuts.

If you prefer scripts, combine the static lists with tools like `rg` or `jq`.

## Storage Layout

Media records live in the SQLite database (`noteleaf.db`). Binary assets are not downloaded; the metadata stores canonical URLs so you can jump back to the source at any time.
Use `noteleaf status` to see the exact paths for your database, data directory, and configuration file.
