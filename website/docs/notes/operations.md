---
title: Note Operations
sidebar_label: Operations
description: List, read, edit, and delete notes from the CLI.
sidebar_position: 3
---

# Note Operations

### Listing Notes

**Interactive TUI** (default):
```sh
noteleaf note list
```

Navigate with arrow keys, press Enter to read, `e` to edit, `q` to quit.

**Static list**:
```sh
noteleaf note list --static
```

**Filter by tags**:
```sh
noteleaf note list --tags research,technical
```

**Show archived notes**:
```sh
noteleaf note list --archived
```

### Reading Notes

View note content with formatted rendering:

```sh
noteleaf note read 1
```

Aliases: `noteleaf note view 1`

The viewer renders markdown with syntax highlighting for code blocks, proper formatting for headers and lists, and displays metadata (title, tags, dates).

### Editing Notes

Open note in your configured editor:

```sh
noteleaf note edit 1
```

Noteleaf uses the editor specified in your configuration or the `$EDITOR` environment variable. Common choices: `vim`, `nvim`, `nano`, `code`, `emacs`.

**Configure editor**:
```sh
noteleaf config set editor nvim
```

Changes are saved automatically when you close the editor. The modification timestamp updates to track when notes were last changed.

### Deleting Notes

Remove a note permanently:

```sh
noteleaf note remove 1
```

Aliases: `rm`, `delete`, `del`

This deletes both the markdown file and database metadata. You'll be prompted for confirmation. This operation cannot be undone, so consider archiving instead if you might need the note later.
