---
title: Time Tracking
sidebar_label: Time Tracking
description: Track work, review sessions, and generate timesheets.
sidebar_position: 4
---

# Time Tracking

Track hours spent on tasks for billing, reporting, or personal analytics.

## Starting and Stopping

**Start tracking**:

```sh
noteleaf task start 1
```

With a note about what you're doing:

```sh
noteleaf task start 1 --note "Implementing authentication"
```

**Stop tracking**:

```sh
noteleaf task stop 1
```

Only one task can be actively tracked at a time.

## Viewing Timesheets

**Last 7 days** (default):

```sh
noteleaf task timesheet
```

**Specific time range**:

```sh
noteleaf task timesheet --days 30
```

**For specific task**:

```sh
noteleaf task timesheet --task 1
```

Timesheet shows:

- Date and time range for each session
- Duration
- Notes attached to the session
- Total time per task
- Total time across all tasks
