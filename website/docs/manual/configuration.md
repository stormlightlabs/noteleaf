---
id: configuration
title: Configuration
sidebar_position: 7
description: Manage application configuration
---

## config

Manage noteleaf configuration

```bash
noteleaf config
```

### Subcommands

#### get

Display configuration values.

If no key is provided, displays all configuration values.
Otherwise, displays the value for the specified key.

**Usage:**

```bash
noteleaf config get [key]
```

#### set

Update a configuration value.

Available keys:
  database_path      - Custom database file path
  data_dir           - Custom data directory
  date_format        - Date format string (default: 2006-01-02)
  color_scheme       - Color scheme (default: default)
  default_view       - Default view mode (default: list)
  default_priority   - Default task priority
  editor             - Preferred text editor
  articles_dir       - Articles storage directory
  notes_dir          - Notes storage directory
  auto_archive       - Auto-archive completed items (true/false)
  sync_enabled       - Enable synchronization (true/false)
  sync_endpoint      - Synchronization endpoint URL
  sync_token         - Synchronization token
  export_format      - Default export format (default: json)
  movie_api_key      - API key for movie database
  book_api_key       - API key for book database

**Usage:**

```bash
noteleaf config set <key> <value>
```

#### path

Display the path to the configuration file being used.

**Usage:**

```bash
noteleaf config path
```

#### reset

Reset all configuration values to their defaults.

**Usage:**

```bash
noteleaf config reset
```

