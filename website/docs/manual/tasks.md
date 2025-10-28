---
id: task-management
title: Task Management
sidebar_position: 1
description: Manage tasks with TaskWarrior-inspired features
---

## todo

Manage tasks with TaskWarrior-inspired features.

Track todos with priorities, projects, contexts, and tags. Supports hierarchical
tasks with parent/child relationships, task dependencies, recurring tasks, and
time tracking. Tasks can be filtered by status, priority, project, or context.

```bash
noteleaf todo
```

### Subcommands

#### add

Create a new task with description and optional attributes.

Tasks can be created with priority levels (low, medium, high, urgent), assigned
to projects and contexts, tagged for organization, and configured with due dates
and recurrence rules. Dependencies can be established to ensure tasks are
completed in order.

Examples:
  noteleaf todo add "Write documentation" --priority high --project docs
  noteleaf todo add "Weekly review" --recur "FREQ=WEEKLY" --due 2024-01-15

**Usage:**

```bash
noteleaf todo add [description] [flags]
```

**Options:**

```
  -c, --context string      Set task context
      --depends-on string   Set task dependencies (comma-separated UUIDs)
  -d, --due string          Set due date (YYYY-MM-DD)
      --parent string       Set parent task UUID
  -p, --priority string     Set task priority
      --project string      Set task project
      --recur string        Set recurrence rule (e.g., FREQ=DAILY)
  -t, --tags strings        Add tags to task
      --until string        Set recurrence end date (YYYY-MM-DD)
```

**Aliases:** create, new

#### list

List tasks with optional filtering and display modes.

By default, shows tasks in an interactive TaskWarrior-like interface.
Use --static to show a simple text list instead.
Use --all to show all tasks, otherwise only pending tasks are shown.

**Usage:**

```bash
noteleaf todo list [flags]
```

**Options:**

```
  -a, --all               Show all tasks (default: pending only)
      --context string    Filter by context
  -i, --interactive       Force interactive mode (default)
      --priority string   Filter by priority
      --project string    Filter by project
      --static            Use static text output instead of interactive
      --status string     Filter by status
```

**Aliases:** ls

#### view

Display detailed information for a specific task.

Shows all task attributes including description, status, priority, project,
context, tags, due date, creation time, and modification history. Use --json
for machine-readable output or --no-metadata to show only the description.

**Usage:**

```bash
noteleaf todo view [task-id] [flags]
```

**Options:**

```
      --format string   Output format (detailed, brief) (default "detailed")
      --json            Output as JSON
      --no-metadata     Hide creation/modification timestamps
```

#### update

Modify attributes of an existing task.

Update any task property including description, status, priority, project,
context, due date, recurrence rule, or parent task. Add or remove tags and
dependencies. Multiple attributes can be updated in a single command.

Examples:
  noteleaf todo update 123 --priority urgent --due tomorrow
  noteleaf todo update 456 --add-tag urgent --project website

**Usage:**

```bash
noteleaf todo update [task-id] [flags]
```

**Options:**

```
      --add-depends string      Add task dependencies (comma-separated UUIDs)
      --add-tag strings         Add tags to task
  -c, --context string          Set task context
      --description string      Update task description
  -d, --due string              Set due date (YYYY-MM-DD)
      --parent string           Set parent task UUID
  -p, --priority string         Set task priority
      --project string          Set task project
      --recur string            Set recurrence rule (e.g., FREQ=DAILY)
      --remove-depends string   Remove task dependencies (comma-separated UUIDs)
      --remove-tag strings      Remove tags from task
      --status string           Update task status
  -t, --tags strings            Add tags to task
      --until string            Set recurrence end date (YYYY-MM-DD)
```

#### edit

Open interactive editor for task modification.

Provides a user-friendly interface with status picker and priority toggle.
Easier than using multiple command-line flags for complex updates.

**Usage:**

```bash
noteleaf todo edit [task-id]
```

**Aliases:** e

#### delete

Permanently remove a task from the database.

This operation cannot be undone. Consider updating the task status to
'deleted' instead if you want to preserve the record for historical purposes.

**Usage:**

```bash
noteleaf todo delete [task-id]
```

#### projects

Display all projects with task counts.

Shows each project used in your tasks along with the number of tasks in each
project. Use --todo-txt to format output with +project syntax for compatibility
with todo.txt tools.

**Usage:**

```bash
noteleaf todo projects [flags]
```

**Options:**

```
      --static     Use static text output instead of interactive
      --todo-txt   Format output with +project prefix for todo.txt compatibility
```

**Aliases:** proj

#### tags

Display all tags used across tasks.

Shows each tag with the number of tasks using it. Tags provide flexible
categorization orthogonal to projects and contexts.

**Usage:**

```bash
noteleaf todo tags [flags]
```

**Options:**

```
      --static   Use static text output instead of interactive
```

**Aliases:** t

#### contexts

Display all contexts with task counts.

Contexts represent locations or environments where tasks can be completed (e.g.,
@home, @office, @errands). Use --todo-txt to format output with @context syntax
for compatibility with todo.txt tools.

**Usage:**

```bash
noteleaf todo contexts [flags]
```

**Options:**

```
      --static     Use static text output instead of interactive
      --todo-txt   Format output with @context prefix for todo.txt compatibility
```

**Aliases:** con, loc, ctx, locations

#### done

Mark a task as completed with current timestamp.

Sets the task status to 'completed' and records the completion time. For
recurring tasks, generates the next instance based on the recurrence rule.

**Usage:**

```bash
noteleaf todo done [task-id]
```

**Aliases:** complete

#### start

Begin tracking time spent on a task.

Records the start time for a work session. Only one task can be actively
tracked at a time. Use --note to add a description of what you're working on.

**Usage:**

```bash
noteleaf todo start [task-id] [flags]
```

**Options:**

```
  -n, --note string   Add a note to the time entry
```

#### stop

End time tracking for the active task.

Records the end time and calculates duration for the current work session.
Duration is added to the task's total time tracked.

**Usage:**

```bash
noteleaf todo stop [task-id]
```

#### timesheet

Show time tracking summary for tasks.

By default shows time entries for the last 7 days.
Use --task to show timesheet for a specific task.
Use --days to change the date range.

**Usage:**

```bash
noteleaf todo timesheet [flags]
```

**Options:**

```
  -d, --days int      Number of days to show in timesheet (default 7)
  -t, --task string   Show timesheet for specific task ID
```

#### recur

Configure recurring task patterns.

Create tasks that repeat on a schedule using iCalendar recurrence rules (RRULE).
Supports daily, weekly, monthly, and yearly patterns with optional end dates.

**Usage:**

```bash
noteleaf todo recur
```

**Aliases:** repeat

##### set

Apply a recurrence rule to create repeating task instances.

Uses iCalendar RRULE syntax (e.g., "FREQ=DAILY" for daily tasks, "FREQ=WEEKLY;BYDAY=MO,WE,FR"
for specific weekdays). When a recurring task is completed, the next instance is
automatically generated.

Examples:
  noteleaf todo recur set 123 --rule "FREQ=DAILY"
  noteleaf todo recur set 456 --rule "FREQ=WEEKLY;BYDAY=MO" --until 2024-12-31

**Usage:**

```bash
noteleaf todo recur set [task-id] [flags]
```

**Options:**

```
      --rule string    Recurrence rule (e.g., FREQ=DAILY)
      --until string   Recurrence end date (YYYY-MM-DD)
```

##### clear

Remove recurrence from a task.

Converts a recurring task to a one-time task. Existing future instances are not
affected.

**Usage:**

```bash
noteleaf todo recur clear [task-id]
```

##### show

Display recurrence rule and schedule information.

Shows the RRULE pattern, next occurrence date, and recurrence end date if
configured.

**Usage:**

```bash
noteleaf todo recur show [task-id]
```

#### depend

Create and manage task dependencies.

Establish relationships where one task must be completed before another can
begin. Useful for multi-step workflows and project management.

**Usage:**

```bash
noteleaf todo depend
```

**Aliases:** dep, deps

##### add

Make a task dependent on another task's completion.

The first task cannot be started until the second task is completed. Use task
UUIDs to specify dependencies.

**Usage:**

```bash
noteleaf todo depend add [task-id] [depends-on-uuid]
```

##### remove

Delete a dependency relationship between two tasks.

**Usage:**

```bash
noteleaf todo depend remove [task-id] [depends-on-uuid]
```

**Aliases:** rm

##### list

Show all tasks that must be completed before this task can be started.

**Usage:**

```bash
noteleaf todo depend list [task-id]
```

**Aliases:** ls

##### blocked-by

Display all tasks that depend on this task's completion.

**Usage:**

```bash
noteleaf todo depend blocked-by [task-id]
```

## todo

Manage tasks with TaskWarrior-inspired features.

Track todos with priorities, projects, contexts, and tags. Supports hierarchical
tasks with parent/child relationships, task dependencies, recurring tasks, and
time tracking. Tasks can be filtered by status, priority, project, or context.

```bash
noteleaf todo
```

### Subcommands

#### add

Create a new task with description and optional attributes.

Tasks can be created with priority levels (low, medium, high, urgent), assigned
to projects and contexts, tagged for organization, and configured with due dates
and recurrence rules. Dependencies can be established to ensure tasks are
completed in order.

Examples:
  noteleaf todo add "Write documentation" --priority high --project docs
  noteleaf todo add "Weekly review" --recur "FREQ=WEEKLY" --due 2024-01-15

**Usage:**

```bash
noteleaf todo add [description] [flags]
```

**Options:**

```
  -c, --context string      Set task context
      --depends-on string   Set task dependencies (comma-separated UUIDs)
  -d, --due string          Set due date (YYYY-MM-DD)
      --parent string       Set parent task UUID
  -p, --priority string     Set task priority
      --project string      Set task project
      --recur string        Set recurrence rule (e.g., FREQ=DAILY)
  -t, --tags strings        Add tags to task
      --until string        Set recurrence end date (YYYY-MM-DD)
```

**Aliases:** create, new

#### list

List tasks with optional filtering and display modes.

By default, shows tasks in an interactive TaskWarrior-like interface.
Use --static to show a simple text list instead.
Use --all to show all tasks, otherwise only pending tasks are shown.

**Usage:**

```bash
noteleaf todo list [flags]
```

**Options:**

```
  -a, --all               Show all tasks (default: pending only)
      --context string    Filter by context
  -i, --interactive       Force interactive mode (default)
      --priority string   Filter by priority
      --project string    Filter by project
      --static            Use static text output instead of interactive
      --status string     Filter by status
```

**Aliases:** ls

#### view

Display detailed information for a specific task.

Shows all task attributes including description, status, priority, project,
context, tags, due date, creation time, and modification history. Use --json
for machine-readable output or --no-metadata to show only the description.

**Usage:**

```bash
noteleaf todo view [task-id] [flags]
```

**Options:**

```
      --format string   Output format (detailed, brief) (default "detailed")
      --json            Output as JSON
      --no-metadata     Hide creation/modification timestamps
```

#### update

Modify attributes of an existing task.

Update any task property including description, status, priority, project,
context, due date, recurrence rule, or parent task. Add or remove tags and
dependencies. Multiple attributes can be updated in a single command.

Examples:
  noteleaf todo update 123 --priority urgent --due tomorrow
  noteleaf todo update 456 --add-tag urgent --project website

**Usage:**

```bash
noteleaf todo update [task-id] [flags]
```

**Options:**

```
      --add-depends string      Add task dependencies (comma-separated UUIDs)
      --add-tag strings         Add tags to task
  -c, --context string          Set task context
      --description string      Update task description
  -d, --due string              Set due date (YYYY-MM-DD)
      --parent string           Set parent task UUID
  -p, --priority string         Set task priority
      --project string          Set task project
      --recur string            Set recurrence rule (e.g., FREQ=DAILY)
      --remove-depends string   Remove task dependencies (comma-separated UUIDs)
      --remove-tag strings      Remove tags from task
      --status string           Update task status
  -t, --tags strings            Add tags to task
      --until string            Set recurrence end date (YYYY-MM-DD)
```

#### edit

Open interactive editor for task modification.

Provides a user-friendly interface with status picker and priority toggle.
Easier than using multiple command-line flags for complex updates.

**Usage:**

```bash
noteleaf todo edit [task-id]
```

**Aliases:** e

#### delete

Permanently remove a task from the database.

This operation cannot be undone. Consider updating the task status to
'deleted' instead if you want to preserve the record for historical purposes.

**Usage:**

```bash
noteleaf todo delete [task-id]
```

#### projects

Display all projects with task counts.

Shows each project used in your tasks along with the number of tasks in each
project. Use --todo-txt to format output with +project syntax for compatibility
with todo.txt tools.

**Usage:**

```bash
noteleaf todo projects [flags]
```

**Options:**

```
      --static     Use static text output instead of interactive
      --todo-txt   Format output with +project prefix for todo.txt compatibility
```

**Aliases:** proj

#### tags

Display all tags used across tasks.

Shows each tag with the number of tasks using it. Tags provide flexible
categorization orthogonal to projects and contexts.

**Usage:**

```bash
noteleaf todo tags [flags]
```

**Options:**

```
      --static   Use static text output instead of interactive
```

**Aliases:** t

#### contexts

Display all contexts with task counts.

Contexts represent locations or environments where tasks can be completed (e.g.,
@home, @office, @errands). Use --todo-txt to format output with @context syntax
for compatibility with todo.txt tools.

**Usage:**

```bash
noteleaf todo contexts [flags]
```

**Options:**

```
      --static     Use static text output instead of interactive
      --todo-txt   Format output with @context prefix for todo.txt compatibility
```

**Aliases:** con, loc, ctx, locations

#### done

Mark a task as completed with current timestamp.

Sets the task status to 'completed' and records the completion time. For
recurring tasks, generates the next instance based on the recurrence rule.

**Usage:**

```bash
noteleaf todo done [task-id]
```

**Aliases:** complete

#### start

Begin tracking time spent on a task.

Records the start time for a work session. Only one task can be actively
tracked at a time. Use --note to add a description of what you're working on.

**Usage:**

```bash
noteleaf todo start [task-id] [flags]
```

**Options:**

```
  -n, --note string   Add a note to the time entry
```

#### stop

End time tracking for the active task.

Records the end time and calculates duration for the current work session.
Duration is added to the task's total time tracked.

**Usage:**

```bash
noteleaf todo stop [task-id]
```

#### timesheet

Show time tracking summary for tasks.

By default shows time entries for the last 7 days.
Use --task to show timesheet for a specific task.
Use --days to change the date range.

**Usage:**

```bash
noteleaf todo timesheet [flags]
```

**Options:**

```
  -d, --days int      Number of days to show in timesheet (default 7)
  -t, --task string   Show timesheet for specific task ID
```

#### recur

Configure recurring task patterns.

Create tasks that repeat on a schedule using iCalendar recurrence rules (RRULE).
Supports daily, weekly, monthly, and yearly patterns with optional end dates.

**Usage:**

```bash
noteleaf todo recur
```

**Aliases:** repeat

##### set

Apply a recurrence rule to create repeating task instances.

Uses iCalendar RRULE syntax (e.g., "FREQ=DAILY" for daily tasks, "FREQ=WEEKLY;BYDAY=MO,WE,FR"
for specific weekdays). When a recurring task is completed, the next instance is
automatically generated.

Examples:
  noteleaf todo recur set 123 --rule "FREQ=DAILY"
  noteleaf todo recur set 456 --rule "FREQ=WEEKLY;BYDAY=MO" --until 2024-12-31

**Usage:**

```bash
noteleaf todo recur set [task-id] [flags]
```

**Options:**

```
      --rule string    Recurrence rule (e.g., FREQ=DAILY)
      --until string   Recurrence end date (YYYY-MM-DD)
```

##### clear

Remove recurrence from a task.

Converts a recurring task to a one-time task. Existing future instances are not
affected.

**Usage:**

```bash
noteleaf todo recur clear [task-id]
```

##### show

Display recurrence rule and schedule information.

Shows the RRULE pattern, next occurrence date, and recurrence end date if
configured.

**Usage:**

```bash
noteleaf todo recur show [task-id]
```

#### depend

Create and manage task dependencies.

Establish relationships where one task must be completed before another can
begin. Useful for multi-step workflows and project management.

**Usage:**

```bash
noteleaf todo depend
```

**Aliases:** dep, deps

##### add

Make a task dependent on another task's completion.

The first task cannot be started until the second task is completed. Use task
UUIDs to specify dependencies.

**Usage:**

```bash
noteleaf todo depend add [task-id] [depends-on-uuid]
```

##### remove

Delete a dependency relationship between two tasks.

**Usage:**

```bash
noteleaf todo depend remove [task-id] [depends-on-uuid]
```

**Aliases:** rm

##### list

Show all tasks that must be completed before this task can be started.

**Usage:**

```bash
noteleaf todo depend list [task-id]
```

**Aliases:** ls

##### blocked-by

Display all tasks that depend on this task's completion.

**Usage:**

```bash
noteleaf todo depend blocked-by [task-id]
```

