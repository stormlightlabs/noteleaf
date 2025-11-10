---
title: Note Basics
sidebar_label: Basics
description: Creating notes, metadata, and storage model.
sidebar_position: 2
---

# Note Basics

## Creation

**Quick note from command line**:

```sh
noteleaf note create "Meeting Notes" "Discussed Q4 roadmap and hiring plans"
```

**Interactive creation** (opens editor):

```sh
noteleaf note create --interactive
```

**From existing file**:

```sh
noteleaf note create --file ~/Documents/draft.md
```

**Create and immediately edit**:

```sh
noteleaf note create "Research Notes" --editor
```

## Structure

Notes consist of:

**Title**: Short descriptor shown in lists and searches. Can be updated later.

**Content**: Full markdown text. Supports all standard markdown features including code blocks, lists, tables, and links.

**Tags**: Categorization labels for organizing and filtering notes. Multiple tags per note.

**Dates**: Creation and modification timestamps tracked automatically.

**File Path**: Location of the markdown file on disk, managed by Noteleaf.

## Storage

**File Location**: Notes are stored as individual `.md` files in your notes directory (typically `~/.local/share/noteleaf/notes` or `~/Library/Application Support/noteleaf/notes`).

**Naming**: Files are named with a UUID to ensure uniqueness. The title is stored in the database, not the filename.

**Portability**: Since notes are plain markdown, you can read them with any text editor or markdown viewer.
The database provides additional functionality like tagging and search, but the files remain standalone.
