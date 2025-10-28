---
id: articles
title: Articles
sidebar_position: 3
description: Save and archive web articles
---

## article

Save and archive web articles locally.

Parse articles from supported websites, extract clean content, and save as
both markdown and HTML. Maintains a searchable archive of articles with
metadata including author, title, and publication date.

```bash
noteleaf article
```

### Subcommands

#### add

Parse and save article content from a supported website.

The article will be parsed using domain-specific XPath rules and saved
as both Markdown and HTML files. Article metadata is stored in the database.

**Usage:**

```bash
noteleaf article add <url>
```

#### list

List saved articles with optional filtering.

Use query to filter by title, or use flags for more specific filtering.

**Usage:**

```bash
noteleaf article list [query] [flags]
```

**Options:**

```
      --author string   Filter by author
  -l, --limit int       Limit number of results (0 = no limit)
```

**Aliases:** ls

#### view

Display article metadata and summary.

Shows article title, author, publication date, URL, and a brief content
preview. Use 'read' command to view the full article content.

**Usage:**

```bash
noteleaf article view <id>
```

**Aliases:** show

#### read

Read the full markdown content of an article with beautiful formatting.

This displays the complete article content using syntax highlighting and proper formatting.

**Usage:**

```bash
noteleaf article read <id>
```

#### remove

Delete an article and its files permanently.

Removes the article metadata from the database and deletes associated markdown
and HTML files. This operation cannot be undone.

**Usage:**

```bash
noteleaf article remove <id>
```

**Aliases:** rm, delete

