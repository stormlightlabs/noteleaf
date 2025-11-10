---
title: Task Basics
sidebar_label: Basics
description: Create tasks and understand their core attributes.
sidebar_position: 2
---

# Task Basics

## Creation

Create a simple task:

```sh
noteleaf task add "Write documentation"
```

Create a task with attributes:

```sh
noteleaf task add "Review pull requests" \
  --priority high \
  --project work \
  --tags urgent,code-review \
  --due 2025-01-15
```

## Properties

**Description**: What needs to be done. Can be updated later with `task update`.

**Status**: Task lifecycle state:

- `pending`: Not yet started (default for new tasks)
- `active`: Currently being worked on
- `completed`: Finished successfully
- `deleted`: Removed but preserved for history
- `waiting`: Blocked or postponed

**Priority**: Importance level affects sorting and display:

- `low`: Nice to have, defer if busy
- `medium`: Standard priority (default)
- `high`: Important, should be done soon
- `urgent`: Critical, top of the list

**Project**: Group related tasks together. Examples: `work`, `home`, `side-project`. Projects create organizational boundaries and enable filtering.

**Context**: Location or mode where task can be done. Examples: `@home`, `@office`, `@phone`, `@computer`. Contexts help filter tasks based on current situation.

**Tags**: Flexible categorization orthogonal to projects. Examples: `urgent`, `quick-win`, `research`, `bug`. Multiple tags per task.

**Due Date**: When the task should be completed. Format: `YYYY-MM-DD` or relative (`tomorrow`, `next week`).

### Lifecycle

Tasks move through statuses as work progresses:

```
pending -> active -> completed
           |
           v
        waiting
           |
           v
        deleted
```

**Mark task as active**:

```sh
noteleaf task update 1 --status active
```

**Complete a task**:

```sh
noteleaf task done 1
```

**Delete a task**:

```sh
noteleaf task delete 1
```
