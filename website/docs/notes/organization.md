---
title: Note Organization
sidebar_label: Organization
description: Tagging, linking, and template workflows for notes.
sidebar_position: 4
---

# Note Organization

## Tagging

Tags provide flexible categorization without hierarchical constraints.

**Add tags during creation**:

```sh
noteleaf note create "API Design" --tags architecture,reference
```

**Add tags to existing note**:

```sh
noteleaf note update 1 --add-tag reference
```

**Remove tags**:

```sh
noteleaf note update 1 --remove-tag draft
```

**List all tags**:

```sh
noteleaf note tags
```

Shows each tag with the count of notes using it.

**Tag naming conventions**: Use lowercase, hyphens for compound tags. Examples: `research`, `meeting-notes`, `how-to`, `reference`, `technical`, `personal`.

## Linking

While not a first-class feature in the current UI, notes can reference tasks by ID or description:

```markdown
# Implementation Plan

Related task: #42 (Deploy authentication service)

## Next Steps
- Complete testing (task #43)
- Write documentation (task #44)
```

Future versions may support automatic linking between notes and tasks in the database.

## Templates

Create reusable note structures using shell functions or scripts:

**In ~/.bashrc or ~/.zshrc**:

```sh
meeting_note() {
  local title="Meeting: $1"
  local date=$(date +%Y-%m-%d)
  local content="# $title

**Date**: $date

## Attendees
-

## Agenda
-

## Discussion
-

## Action Items
- [ ]

## Next Meeting
- "

  echo "$content" | noteleaf note create "$title" --tags meeting --editor
}

daily_note() {
  local date=$(date +%Y-%m-%d)
  local title="Daily: $date"
  local content="# $title

## Completed Today
-

## In Progress
-

## Tomorrow's Focus
-

## Notes
- "

  echo "$content" | noteleaf note create "$title" --tags daily --editor
}
```

Usage:

```sh
meeting_note "Q4 Planning"
daily_note
```
