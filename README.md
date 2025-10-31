# Noteleaf

[![codecov](https://codecov.io/gh/stormlightlabs/noteleaf/branch/main/graph/badge.svg)](https://codecov.io/gh/stormlightlabs/noteleaf)
[![Go Report Card](https://goreportcard.com/badge/github.com/stormlightlabs/noteleaf)](https://goreportcard.com/report/github.com/stormlightlabs/noteleaf)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/stormlightlabs/noteleaf)](go.mod)

Noteleaf is a unified personal productivity CLI that combines task management, note-taking, and media tracking in one place.
It provides TaskWarrior-inspired task management with additional support for notes, articles, books, movies, and TV shows - all built with Golang & Charm.sh libs. Inspired by TaskWarrior & todo.txt CLI applications.

## Why?

- **Fragmented productivity tools**: Instead of juggling multiple apps for tasks, notes, reading lists, and media queues, Noteleaf provides a single CLI interface
- **Terminal-native workflow**: For developers and power users who prefer staying in the terminal, Noteleaf offers rich TUIs without leaving your command line
    - **Lightweight and fast**: No desktop apps or web interfaces - just a fast, focused CLI tool
- **Unified data model**: Tasks, notes, and media items can reference each other, creating a connected knowledge and productivity system

## Getting started

### Prerequisites

Go v1.24+

### Installation

```sh
git clone https://github.com/stormlightlabs/noteleaf
cd noteleaf
go build -o ./tmp/noteleaf ./cmd
go install
```

### Basic usage

```sh
# Initialize the application
noteleaf setup

# Add sample data for exploration
noteleaf setup seed

# Create your first task
noteleaf task add "Learn Noteleaf CLI"

# View tasks
noteleaf task list

# Create a note
noteleaf note add "My first note"

# Add a book to your reading list
noteleaf media book add "The Name of the Wind"

# Generate docs
noteleaf docgen --format docusaurus --out ./website/docs/manual
```

## Status

**Status**: Work in Progress (MVP completed)

### Completed

- Task management with projects and tags
- Note-taking system
- Article parsing from URLs
- Media tracking (books, movies, TV shows)

### Planned

- Time tracking integration
- Advanced search and filtering
- Export/import functionality
- Plugin system
