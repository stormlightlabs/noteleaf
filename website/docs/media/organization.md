---
title: Media Organization
sidebar_label: Organization
description: Keep queues manageable with filters, reviews, and note links.
sidebar_position: 5
---

# Media Organization

Media entries share the same database as tasks and notes, so you can cross-reference everything. This page outlines the practical workflows for keeping large queues in check.

## Tags and Categories

Dedicated media tags have not shipped yet. Until they do:

- Use the free-form `Notes` column (populated automatically from Open Library or Rotten Tomatoes) to stash keywords such as “Hugo shortlist” or “Documentary”.
- When you need stricter structure, create a note that tracks an ad-hoc category and reference media IDs inside it:

```markdown
## Cozy backlog
- Book #11 – comfort reread
- Movie #25 – rainy-day pick
```

Full-text search (`noteleaf note list` → `/` and search) will surface the note instantly, and the numeric IDs jump you right back into the media commands.

## Custom Lists

You already get status-based filters out of the box:

```sh
noteleaf media book list --reading
noteleaf media movie list --watched
noteleaf media tv list --all | rg "FX"   # filter with ripgrep
```

For more bespoke dashboards:

1. Use `noteleaf status` to grab the SQLite path.
2. Query it with tools like `sqlite-utils` or `datasette` to build spreadsheets or dashboards.
3. Export subsets via `sqlite3 noteleaf.db "SELECT * FROM books WHERE status='reading'" > reading.csv`.

That approach keeps the CLI fast while still letting you slice the data any way you need.

## Ratings and Reviews

The database schema already includes a `rating` column for every media type. Rotten Tomatoes/Open Library populate it with critic hints for now; personal star ratings will become editable in a future release.

Until then, keep reviews as regular notes:

```sh
noteleaf note create "Thoughts on Book #7"
```

Inside the note, link back to the record (`Book #7`, `Movie #18`, etc.) so searches tie everything together. Because notes live on disk you can also version-control your reviews.

## Linking Media to Notes

There is no special “link” command yet, but the following pattern works well:

1. Create a dedicated note per book/movie/show (or per collection).
2. Add a heading with the media ID and paste the generated markdown path from `noteleaf article view` or the queue list.
3. Optionally embed checklists or quotes gathered while reading/watching.

Example snippet:

```markdown
### Book #7 — Project Hail Mary
- Status: reading (45%)
- Tasks: todo #128 covers the experiment described in chapter 12
- Next action: finish Part II before Friday
```

Because tasks, notes, and media share the same SQLite file, future automation can join across them without migrations. When official linking lands it will reuse these IDs, so the prep work you do now keeps paying off.
