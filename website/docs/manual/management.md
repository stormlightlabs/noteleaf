---
id: management
title: Management
sidebar_position: 8
description: Application management commands
---

## status

Display comprehensive application status information.

Shows database location, configuration file path, data directories, and current
settings. Use this command to verify your noteleaf installation and diagnose
configuration issues.

```bash
noteleaf status
```

## setup

Initialize noteleaf for first use.

Creates the database, configuration file, and required data directories. Run
this command after installing noteleaf or when setting up a new environment.
Safe to run multiple times as it will skip existing resources.

```bash
noteleaf setup
```

### Subcommands

#### seed

Add sample tasks, books, and notes to the database for testing and demonstration purposes

**Usage:**

```bash
noteleaf setup seed [flags]
```

**Options:**

```
  -f, --force   Clear existing data and re-seed
```

## reset

Remove all application data and return to initial state.

This command deletes the database, all media files, notes, and articles. The
configuration file is preserved. Use with caution as this operation cannot be
undone. You will be prompted for confirmation before deletion proceeds.

```bash
noteleaf reset
```

