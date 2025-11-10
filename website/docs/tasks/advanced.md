---
title: Advanced Task Features
sidebar_label: Advanced
description: Recurring tasks, dependencies, hierarchies, and custom fields.
sidebar_position: 6
---

# Advanced Task Features

## Recurrence

Create tasks that repeat on a schedule using iCalendar recurrence rules.

**Daily task**:

```sh
noteleaf task add "Daily standup" --recur "FREQ=DAILY"
```

**Weekly task on specific days**:

```sh
noteleaf task add "Team meeting" --recur "FREQ=WEEKLY;BYDAY=MO,WE"
```

**Monthly task**:

```sh
noteleaf task add "Invoice review" --recur "FREQ=MONTHLY;BYMONTHDAY=1"
```

**With end date**:

```sh
noteleaf task add "Q1 review" \
  --recur "FREQ=WEEKLY" \
  --until 2025-03-31
```

**Manage recurrence**:

Set recurrence on existing task:

```sh
noteleaf task recur set 1 --rule "FREQ=DAILY"
```

View recurrence info:

```sh
noteleaf task recur show 1
```

Remove recurrence:

```sh
noteleaf task recur clear 1
```

When you complete a recurring task, Noteleaf automatically generates the next instance based on the recurrence rule.

## Dependencies

Create relationships where tasks must be completed in order.

**Add dependency** (task 1 depends on task 2):

```sh
noteleaf task depend add 1 <uuid-of-task-2>
```

**List dependencies** (what must be done first):

```sh
noteleaf task depend list 1
```

**List blocked tasks** (what's waiting on this task):

```sh
noteleaf task depend blocked-by 1
```

**Remove dependency**:

```sh
noteleaf task depend remove 1 <uuid-of-task-2>
```

Dependencies use task UUIDs (shown in `task view`) rather than IDs for stability across database changes.

## Hierarchical Tasks

Create parent-child relationships for breaking down large tasks.

**Create child task**:

```sh
noteleaf task add "Write API documentation" --parent <parent-uuid>
```

Parent tasks can have multiple children, creating a tree structure for complex projects.

## Custom Attributes

While not exposed through specific flags, the database schema supports extending tasks with custom attributes for advanced use cases or scripting.
