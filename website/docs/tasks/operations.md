---
title: Task Operations
sidebar_label: Operations
description: List, view, and update tasks from the CLI and TUI.
sidebar_position: 3
---

# Task Operations

## Listing and Filtering

**Interactive list** (default):

```sh
noteleaf task list
```

Navigate with arrow keys, press Enter to view details, `q` to quit.

**Static list** (for scripting):

```sh
noteleaf task list --static
```

**Filter by status**:

```sh
noteleaf task list --status pending
noteleaf task list --status completed
```

**Filter by project**:

```sh
noteleaf task list --project work
```

**Filter by priority**:

```sh
noteleaf task list --priority high
```

**Filter by context**:

```sh
noteleaf task list --context @office
```

**Show all tasks** (including completed):

```sh
noteleaf task list --all
```

## Viewing Task Details

View complete task information:

```sh
noteleaf task view 1
```

JSON output for scripts:

```sh
noteleaf task view 1 --json
```

Brief format without metadata:

```sh
noteleaf task view 1 --format brief
```

## Updating Tasks

Update single attribute:

```sh
noteleaf task update 1 --priority urgent
```

Update multiple attributes:

```sh
noteleaf task update 1 \
  --priority urgent \
  --due tomorrow \
  --add-tag critical
```

Change description:

```sh
noteleaf task update 1 --description "New task description"
```

Add and remove tags:

```sh
noteleaf task update 1 --add-tag urgent --remove-tag later
```

## Interactive Editing

Open interactive editor for complex changes:

```sh
noteleaf task edit 1
```

This provides a TUI with visual pickers for status and priority, making updates faster than command flags.
