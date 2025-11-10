---
title: Terminal UI
sidebar_label: Terminal UI
description: Navigate Noteleaf’s Bubble Tea interfaces and their script-friendly counterparts.
sidebar_position: 6
---

# Terminal UI

Most list-style commands (tasks, notes, books) have two personalities: an interactive Bubble Tea view for exploration and a static text output for piping into other tools. This page explains how both modes behave.

## Interactive Mode

### Navigation

- Launch the TUI with the default command (`noteleaf todo list`, `noteleaf note list`, `noteleaf media book add -i`, etc.).  
- Use `j`/`k` or the arrow keys to move the selection. Page Up/Down jump faster, while `g`/`G` (or Home/End) snap to the top or bottom depending on the view.  
- Search is always available—press `/` and start typing to filter titles, tags, projects, or notes in real time.

### Keyboard shortcuts

All interactive components reuse the same key map defined in `internal/ui/data_list.go` and `internal/ui/data_table.go`:

| Keys | Action |
|------|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `enter`   | Select the highlighted row |
| `v`       | Open the detail preview (when supported) |
| `/`       | Start search |
| `r`       | Refresh data from the database |
| `1-9`     | Jump directly to a row index |
| `q`, `ctrl+c` | Quit the view |

The shortcuts appear in the on-screen help so you never have to memorize them all.

### Selection and actions

- Press `enter` to activate the primary action (open a note, view a task, confirm a media selection, etc.).  
- Some screens expose extra actions on letter keys (e.g., `a` to archive, `e` to edit). They are listed alongside the contextual help (`?`).  
- Interactive prompts such as `noteleaf media movie add` use the same selection model, so keyboard muscle memory carries over.

### Help screens

Hit `?` at any time to open the inline help overlay. It mirrors the bindings configured for the active component and also hints at hidden actions. Press `esc`, `backspace`, or `?` again to exit.

## Static Mode

### Command-line output

Add `--static` (or remove `-i`) to force plain text output. Examples:

```sh
noteleaf todo list --static
noteleaf note list --static --tag meeting
noteleaf media book list --all --static
```

Static mode prints tables with headings so they are easy to read or parse. Commands that default to prompts (like `noteleaf media movie add`) fall back to a numbered list when you omit `-i`.

### Scripting with Noteleaf

Static output is predictable, making it straightforward to combine with familiar utilities:

```sh
noteleaf todo list --static --project docs | rg "pending"
noteleaf note list --static | fzf
```

Because each row includes the record ID, you can feed the result back into follow-up commands (`noteleaf note view 42`, `noteleaf todo done 128`, etc.).

### Output formatting

The task viewer supports the `--format` flag for quick summaries:

```sh
noteleaf todo view 12 --format brief
noteleaf todo view 12 --format detailed  # default
```

Brief mode hides timestamps and auxiliary metadata, which keeps CI logs or chat snippets short. Future commands will inherit the same pattern.

### JSON output

Use `--json` wherever it exists (currently on task views/lists) for structured output:

```sh
noteleaf todo view 12 --json | jq '.status'
noteleaf todo list --static --json | jq '.[] | select(.status=="pending")'
```

JSON mode ignores terminal colors and uses machine-friendly field names so you can script exports without touching the SQLite file directly.
