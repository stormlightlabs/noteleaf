---
id: movies
title: Movies
sidebar_position: 5
description: Track movies in watch queue
---

## movie

Track movies you want to watch.

Search TMDB for movies and add them to your queue. Mark movies as watched when
completed. Maintains a history of your movie watching activity.

```bash
noteleaf media movie
```

### Subcommands

#### add

Search for movies and add them to your watch queue.

By default, shows search results in a simple list format where you can select by number.
Use the -i flag for an interactive interface with navigation keys.

**Usage:**

```bash
noteleaf media movie add [search query...] [flags]
```

**Options:**

```
  -i, --interactive   Use interactive interface for movie selection
```

#### list

Display movies in your queue with optional status filters.

Shows movie titles, release years, and current status. Filter by --all to show
everything, --watched for completed movies, or --queued for unwatched items.
Default shows queued movies only.

**Usage:**

```bash
noteleaf media movie list [--all|--watched|--queued]
```

#### watched

Mark a movie as watched with current timestamp. Moves the movie from queued to watched status.

**Usage:**

```bash
noteleaf media movie watched [id]
```

**Aliases:** seen

#### remove

Remove a movie from your watch queue. Use this for movies you no longer want to track.

**Usage:**

```bash
noteleaf media movie remove [id]
```

**Aliases:** rm

