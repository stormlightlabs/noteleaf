---
title: Task Organization
sidebar_label: Organization
description: Use projects, contexts, and tags to structure work.
sidebar_position: 5
---

# Task Organization

## Projects

Projects group related tasks. Useful for separating work contexts or major initiatives.

**List all projects**:

```sh
noteleaf task projects
```

Shows each project with task count.

**Filter tasks by project**:

```sh
noteleaf task list --project work
```

**Project naming**: Use lowercase, hyphens for spaces. Examples: `work`, `side-project`, `home-improvement`.

## Contexts

Contexts represent where or how a task can be done. Helps with GTD-style workflow.

**List all contexts**:

```sh
noteleaf task contexts
```

**Filter by context**:

```sh
noteleaf task list --context @home
```

**Context naming**: Prefix with `@` following GTD convention. Examples: `@home`, `@office`, `@phone`, `@errands`.

## Tags

Tags provide flexible categorization. Unlike projects and contexts, tasks can have multiple tags.

**List all tags**:

```sh
noteleaf task tags
```

**Filter by tags** (tasks must have all specified tags):

```sh
noteleaf task list --tags urgent,bug
```

**Tag naming**: Use lowercase, hyphens for compound tags. Examples: `urgent`, `quick-win`, `code-review`, `waiting-on-feedback`.
