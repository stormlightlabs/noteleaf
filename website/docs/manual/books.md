---
id: books
title: Books
sidebar_position: 4
description: Manage reading list and track progress
---

## book

Track books and reading progress.

Search Google Books API to add books to your reading list. Track which books
you're reading, update progress percentages, and maintain a history of finished
books.

```bash
noteleaf media book
```

### Subcommands

#### add

Search for books and add them to your reading list.

By default, shows search results in a simple list format where you can select by number.
Use the -i flag for an interactive interface with navigation keys.

**Usage:**

```bash
noteleaf media book add [search query...] [flags]
```

**Options:**

```
  -i, --interactive   Use interactive interface for book selection
```

#### list

Display books in your reading list with progress indicators.

Shows book titles, authors, and reading progress percentages. Filter by --all,
--reading for books in progress, --finished for completed books, or --queued
for books not yet started. Default shows queued books only.

**Usage:**

```bash
noteleaf media book list [--all|--reading|--finished|--queued]
```

#### reading

Mark a book as currently reading. Use this when you start a book from your queue.

**Usage:**

```bash
noteleaf media book reading <id>
```

#### finished

Mark a book as finished with current timestamp. Sets reading progress to 100%.

**Usage:**

```bash
noteleaf media book finished <id>
```

**Aliases:** read

#### remove

Remove a book from your reading list. Use this for books you no longer want to track.

**Usage:**

```bash
noteleaf media book remove <id>
```

**Aliases:** rm

#### progress

Set reading progress for a book.

Specify a percentage value between 0 and 100 to indicate how far you've
progressed through the book. Automatically updates status to 'reading' if not
already set.

**Usage:**

```bash
noteleaf media book progress <id> <percentage>
```

#### update

Change a book's status directly.

Valid statuses are: queued (not started), reading (in progress), finished
(completed), or removed (no longer tracking).

**Usage:**

```bash
noteleaf media book update <id> <status>
```

