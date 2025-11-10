---
id: quickstart
title: Quickstart Guide
sidebar_label: Quickstart
sidebar_position: 1
description: Install Noteleaf and learn the essentials in minutes.
---

# Quickstart Guide

This guide will walk you through installing Noteleaf and getting productive with tasks, notes, and media tracking in under 15 minutes.

## Installation

### Requirements

- Go 1.24 or higher
- Git (for cloning the repository)

### Build and Install

Clone the repository and build the binary:

```sh
git clone https://github.com/stormlightlabs/noteleaf
cd noteleaf
go build -o ./tmp/noteleaf ./cmd
```

Optionally, install to your GOPATH:

```sh
go install
```

## Initialize Noteleaf

Set up the database and configuration:

```sh
noteleaf setup
```

This creates:

- Database at `~/.local/share/noteleaf/noteleaf.db` (Linux) or `~/Library/Application Support/noteleaf/noteleaf.db` (macOS)
- Configuration file at `~/.config/noteleaf/config.toml` (Linux) or `~/Library/Application Support/noteleaf/config.toml` (macOS)

### Optional: Add Sample Data

Explore with pre-populated examples:

```sh
noteleaf setup seed
```

## Task Management

### Create Your First Task

```sh
noteleaf task add "Write project proposal"
```

### Add a Task with Priority and Project

```sh
noteleaf task add "Review pull requests" --priority high --project work
```

### List Tasks

Interactive mode with arrow key navigation:

```sh
noteleaf task list
```

Static output for scripting:

```sh
noteleaf task list --static
```

### Mark a Task as Done

```sh
noteleaf task done 1
```

### Track Time

Start tracking:

```sh
noteleaf task start 1
```

Stop tracking:

```sh
noteleaf task stop 1
```

View timesheet:

```sh
noteleaf task timesheet
```

## Note Taking

### Create a Note

Quick note from command line:

```sh
noteleaf note create "Meeting Notes" "Discussed Q4 roadmap and priorities"
```

Create with your editor:

```sh
noteleaf note create --interactive
```

Create from a file:

```sh
noteleaf note create --file notes.md
```

### List and Read Notes

List all notes (interactive):

```sh
noteleaf note list
```

Read a specific note:

```sh
noteleaf note read 1
```

### Edit a Note

Opens in your `$EDITOR`:

```sh
noteleaf note edit 1
```

## Media Tracking

### Books

Search and add from Open Library:

```sh
noteleaf media book add "Project Hail Mary"
```

List your reading queue:

```sh
noteleaf media book list
```

Update reading progress:

```sh
noteleaf media book progress 1 45
```

Mark as finished:

```sh
noteleaf media book finished 1
```

### Movies

Add a movie:

```sh
noteleaf media movie add "The Matrix"
```

Mark as watched:

```sh
noteleaf media movie watched 1
```

### TV Shows

Add a show:

```sh
noteleaf media tv add "Breaking Bad"
```

Update status:

```sh
noteleaf media tv watching 1
```

## Articles

### Save an Article

Parse and save from URL:

```sh
noteleaf article add https://example.com/interesting-post
```

### List Articles

```sh
noteleaf article list
```

Filter by author:

```sh
noteleaf article list --author "Jane Smith"
```

### Read an Article

View in terminal:

```sh
noteleaf article view 1
```

## Configuration

### View Current Configuration

```sh
noteleaf config show
```

### Set Configuration Values

```sh
noteleaf config set editor vim
noteleaf config set default_priority medium
```

### Check Status

View application status and paths:

```sh
noteleaf status
```

## Getting Help

View help for any command:

```sh
noteleaf --help
noteleaf task --help
noteleaf task add --help
```

## Next Steps

Now that you have the basics down:

- Explore advanced task filtering and queries
- Create custom projects and contexts for organizing tasks
- Link notes to tasks and media items
- Set up recurring tasks and dependencies
- Configure the application to match your workflow

For detailed documentation on each command, see the CLI reference in the manual section.
