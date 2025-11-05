# Task Examples

Examples of common task management workflows using Noteleaf.

## Basic Task Management

### Create a Simple Task

```sh
noteleaf task add "Buy groceries"
```

### Create Task with Priority

```sh
noteleaf task add "Fix critical bug" --priority urgent
noteleaf task add "Update documentation" --priority low
```

### Create Task with Project

```sh
noteleaf task add "Design new homepage" --project website
noteleaf task add "Refactor auth service" --project backend
```

### Create Task with Due Date

```sh
noteleaf task add "Submit report" --due 2024-12-31
noteleaf task add "Review PRs" --due tomorrow
```

### Create Task with Tags

```sh
noteleaf task add "Write blog post" --tags writing,blog
noteleaf task add "Server maintenance" --tags ops,backend,infra
```

### Create Task with Context

```sh
noteleaf task add "Call client" --context phone
noteleaf task add "Deploy to production" --context office
```

### Create Task with All Attributes

```sh
noteleaf task add "Launch marketing campaign" \
  --project marketing \
  --priority high \
  --due 2024-06-15 \
  --tags campaign,social \
  --context office
```

## Viewing Tasks

### List All Tasks

Interactive mode:

```sh
noteleaf task list
```

Static output:

```sh
noteleaf task list --static
```

### Filter by Status

```sh
noteleaf task list --status pending
noteleaf task list --status completed
```

### Filter by Priority

```sh
noteleaf task list --priority high
noteleaf task list --priority urgent
```

### Filter by Project

```sh
noteleaf task list --project website
noteleaf task list --project backend
```

### Filter by Tags

```sh
noteleaf task list --tags urgent,bug
```

### View Task Details

```sh
noteleaf task view 1
```

## Updating Tasks

### Mark Task as Done

```sh
noteleaf task done 1
```

### Update Task Priority

```sh
noteleaf task update 1 --priority high
```

### Update Task Project

```sh
noteleaf task update 1 --project website
```

### Add Tags to Task

```sh
noteleaf task update 1 --add-tags backend,api
```

### Remove Tags from Task

```sh
noteleaf task update 1 --remove-tags urgent
```

### Edit Task Interactively

Opens task in your editor:

```sh
noteleaf task edit 1
```

## Time Tracking

### Start Time Tracking

```sh
noteleaf task start 1
```

### Stop Time Tracking

```sh
noteleaf task stop 1
```

### View Timesheet

All entries:

```sh
noteleaf task timesheet
```

Filtered by date:

```sh
noteleaf task timesheet --from 2024-01-01 --to 2024-01-31
```

Filtered by project:

```sh
noteleaf task timesheet --project website
```

## Project Management

### List All Projects

```sh
noteleaf task projects
```

### View Tasks in Project

```sh
noteleaf task list --project website
```

## Tag Management

### List All Tags

```sh
noteleaf task tags
```

### View Tasks with Tag

```sh
noteleaf task list --tags urgent
```

## Context Management

### List All Contexts

```sh
noteleaf task contexts
```

### View Tasks in Context

```sh
noteleaf task list --context office
```

## Advanced Workflows

### Daily Planning

View today's tasks:

```sh
noteleaf task list --due today
```

View overdue tasks:

```sh
noteleaf task list --due overdue
```

### Weekly Review

View completed tasks this week:

```sh
noteleaf task list --status completed --from monday
```

View pending high-priority tasks:

```sh
noteleaf task list --status pending --priority high
```

### Project Focus

List all tasks for a project, sorted by priority:

```sh
noteleaf task list --project website --sort priority
```

### Bulk Operations

Mark multiple tasks as done:

```sh
noteleaf task done 1 2 3 4
```

Delete multiple tasks:

```sh
noteleaf task delete 5 6 7
```

## Task Deletion

### Delete a Task

```sh
noteleaf task delete 1
```

### Delete with Confirmation

```sh
noteleaf task delete 1 --confirm
```
