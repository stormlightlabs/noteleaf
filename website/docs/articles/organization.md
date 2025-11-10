---
title: Article Organization
sidebar_label: Organization
description: Filter your archive, add lightweight tags, and keep backups tidy.
sidebar_position: 3
---

# Article Organization

## Filter by Author or Title

`noteleaf article list` accepts both a free-form query (matches the title) and dedicated flags:

```sh
# Anything with "SQLite" in the title
noteleaf article list SQLite

# Limit to a single author
noteleaf article list --author "Ada Palmer"

# Cap the output for quick reviews
noteleaf article list --author "Ada Palmer" --limit 3
```

Because the database stores created timestamps, results come back with the newest article first, making it easy to run weekly reviews.

## Tagging Articles

There is no first-class tagging UI yet, but Markdown files are yours to edit. Common patterns:

```markdown
---
tags: [distributed-systems, reference]
project: moonshot
---
```

Drop that block right after the generated metadata and tools like `rg` or `ripgrep --json` can surface tagged snippets instantly. You can also maintain a separate note that lists article IDs per topic if you prefer not to edit the captured files.

## Read vs Unread

Opening an article in the terminal does not flip a status flag. Use one of these lightweight conventions instead:

- Prefix the Markdown filename with `read-` once you are done.
- Keep a running checklist note (e.g., “Articles Inbox”) that references IDs and mark them off as you read them.
- Create a task linked to the article ID (`todo add "Summarize article #14"`), then close the task when you finish.

All three approaches work today and will map cleanly to future built-in read/unread tracking.

## Archiving and Backups

The archive lives under `articles_dir`. By default that is `<data_dir>/articles`, where `<data_dir>` depends on your OS:

| Platform | Default |
|----------|---------|
| Linux    | `~/.local/share/noteleaf/articles` |
| macOS    | `~/Library/Application Support/noteleaf/articles` |
| Windows  | `%LOCALAPPDATA%\noteleaf\articles` |

You can override the location via the `articles_dir` setting in `~/.config/noteleaf/.noteleaf.conf.toml` (or by pointing `NOTELEAF_DATA_DIR` to a different root before launching the CLI).

Because every import produces Markdown + HTML, the directory is perfect for version control:

```sh
cd ~/.local/share/noteleaf/articles
git init
git add .
git commit -m "Initial snapshot of article archive"
```

Pair that with your cloud backup tool of choice and you have a durable, fully-offline knowledge base that still integrates seamlessly with Noteleaf’s search commands.
