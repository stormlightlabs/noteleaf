---
title: Task Queries and Filtering
sidebar_label: Queries
description: Compose filters for precise task lists and reports.
sidebar_position: 7
---

# Task Queries and Filtering

Combine filters for precise task lists:

**High priority work tasks due this week**:

```sh
noteleaf task list \
  --project work \
  --priority high \
  --status pending
```

**All completed tasks from specific project**:

```sh
noteleaf task list \
  --project side-project \
  --status completed
```

**Quick wins** (tasks tagged as quick):

```sh
noteleaf task list --tags quick-win
```
