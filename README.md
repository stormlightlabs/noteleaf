# Noteleaf

[![codecov](https://codecov.io/gh/stormlightlabs/noteleaf/branch/main/graph/badge.svg)](https://codecov.io/gh/stormlightlabs/noteleaf)
[![Go Report Card](https://goreportcard.com/badge/github.com/stormlightlabs/noteleaf)](https://goreportcard.com/report/github.com/stormlightlabs/noteleaf)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/stormlightlabs/noteleaf)](go.mod)

Noteleaf is a unified personal productivity CLI that combines task management, note-taking, and media tracking in one place.
It provides TaskWarrior-inspired task management with additional support for notes, articles, books, movies, and TV shows, all built with Golang & Charm.sh libs.
Inspired by TaskWarrior & todo.txt CLI applications.

## Why?

- **Fragmented Ecosystem**: Instead of juggling multiple apps for tasks, notes, reading lists, and media queues, Noteleaf provides a single CLI interface
- **Terminal-native**: For developers and power users who prefer staying in the terminal, Noteleaf offers rich TUIs without leaving your command line
    - **Lightweight**: No desktop apps or web interfaces, just a fast, focused CLI tool
- **Unified data model**: Tasks, notes, and media items can reference each other, creating a connected knowledge and productivity system

## Getting Started

### Quick Install

```sh
git clone https://github.com/stormlightlabs/noteleaf
cd noteleaf
go build -o ./tmp/noteleaf ./cmd
go install
```

### First Steps

For a comprehensive walkthrough including task management, time tracking, notes, and media tracking, see the [Quickstart Guide](website/docs/Quickstart.md).
