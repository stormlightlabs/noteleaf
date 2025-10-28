---
id: tv-shows
title: TV Shows
sidebar_position: 6
description: Manage TV show watching
---

## tv

Track TV shows and episodes.

Search TMDB for TV shows and add them to your queue. Track which shows you're
currently watching, mark episodes as watched, and maintain a complete history
of your viewing activity.

```bash
noteleaf media tv
```

### Subcommands

#### add

Search for TV shows and add them to your watch queue.

By default, shows search results in a simple list format where you can select by number.
Use the -i flag for an interactive interface with navigation keys.

**Usage:**

```bash
noteleaf media tv add [search query...] [flags]
```

**Options:**

```
  -i, --interactive   Use interactive interface for TV show selection
```

#### list

Display TV shows in your queue with optional status filters.

Shows show titles, air dates, and current status. Filter by --all, --queued,
--watching for shows in progress, or --watched for completed series. Default
shows queued shows only.

**Usage:**

```bash
noteleaf media tv list [--all|--queued|--watching|--watched]
```

#### watching

Mark a TV show as currently watching. Use this when you start watching a series.

**Usage:**

```bash
noteleaf media tv watching [id]
```

#### watched

Mark TV show episodes or entire series as watched.

Updates episode tracking and completion status. Can mark individual episodes
or complete seasons/series depending on ID format.

**Usage:**

```bash
noteleaf media tv watched [id]
```

**Aliases:** seen

#### remove

Remove a TV show from your watch queue. Use this for shows you no longer want to track.

**Usage:**

```bash
noteleaf media tv remove [id]
```

**Aliases:** rm

