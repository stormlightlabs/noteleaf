---
title: Architecture Overview
sidebar_label: Architecture
description: Application structure, storage, and UI layers.
sidebar_position: 3
---

# Architecture Overview

## Architecture Overview

### Application Structure

Noteleaf follows a clean architecture pattern with clear separation of concerns:

```
cmd/                    - CLI commands and user interface
internal/
  handlers/            - Business logic and orchestration
  repo/                - Database access layer
  ui/                  - Terminal UI components (Bubble Tea)
  models/              - Domain models
  public/              - Leaflet.pub integration
```

Each layer has defined responsibilities with minimal coupling between them.

### Storage and Database

**SQLite Database**: All structured data (tasks, metadata, relationships) lives in a single SQLite file at `~/.local/share/noteleaf/noteleaf.db` (Linux) or `~/Library/Application Support/noteleaf/noteleaf.db` (macOS).

**Markdown Files**: Note content is stored as individual markdown files on disk. The database tracks metadata while keeping your notes in a portable, human-readable format.

**Database Schema**: Tables for tasks, notes, articles, books, movies, TV shows, publications, and linking tables for tags and relationships. Migrations handle schema evolution.

### TUI Framework (Bubble Tea)

The interactive interface is built with [Bubble Tea](https://github.com/charmbracelet/bubbletea), a Go framework for terminal user interfaces based on The Elm Architecture:

- **Model**: Application state (current view, selected item, filters)
- **Update**: State transitions based on user input
- **View**: Render the current state to the terminal

This architecture makes the UI predictable, testable, and composable. Each screen is an independent component that can be developed and tested in isolation.
