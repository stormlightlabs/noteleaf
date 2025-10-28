---
id: notes
title: Notes
sidebar_position: 2
description: Create and organize markdown notes
---

## note

Create and organize markdown notes with tags.

Write notes in markdown format, organize them with tags, browse them in an
interactive TUI, and edit them in your preferred editor. Notes are stored as
files on disk with metadata tracked in the database.

```bash
noteleaf note
```

### Subcommands

#### create

Create a new markdown note.

Provide a title and optional content inline, or use --interactive to open an
editor. Use --file to import content from an existing markdown file. Notes
support tags for organization and full-text search.

Examples:
  noteleaf note create "Meeting notes" "Discussed project timeline"
  noteleaf note create -i
  noteleaf note create --file ~/documents/draft.md

**Usage:**

```bash
noteleaf note create [title] [content...] [flags]
```

**Options:**

```
  -e, --editor        Prompt to open note in editor after creation
  -f, --file string   Create note from markdown file
  -i, --interactive   Open interactive editor
```

**Aliases:** new

#### list

Opens interactive TUI browser for navigating and viewing notes

**Usage:**

```bash
noteleaf note list [--archived] [--static] [--tags=tag1,tag2] [flags]
```

**Options:**

```
  -a, --archived      Show archived notes
  -s, --static        Show static list instead of interactive TUI
      --tags string   Filter by tags (comma-separated)
```

**Aliases:** ls

#### read

Display note content with formatted markdown rendering.

Shows the note with syntax highlighting, proper formatting, and metadata.
Useful for quick viewing without opening an editor.

**Usage:**

```bash
noteleaf note read [note-id]
```

**Aliases:** view

#### edit

Open note in your configured text editor.

Uses the editor specified in your noteleaf configuration or the EDITOR
environment variable. Changes are automatically saved when you close the
editor.

**Usage:**

```bash
noteleaf note edit [note-id]
```

#### remove

Delete a note permanently.

Removes both the markdown file and database metadata. This operation cannot be
undone. You will be prompted for confirmation before deletion.

**Usage:**

```bash
noteleaf note remove [note-id]
```

**Aliases:** rm, delete, del

