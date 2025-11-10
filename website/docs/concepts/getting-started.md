---
title: Getting Started
sidebar_label: Getting Started
description: Installation steps, configuration overview, and ways to find help.
sidebar_position: 5
---

# Getting Started

## Installation and Setup

### System Requirements

- Go 1.24 or higher
- SQLite 3.35 or higher (usually bundled)
- Terminal with 256-color support
- Unix-like OS (Linux, macOS, WSL)

### Building from Source

Clone the repository and build:

```sh
git clone https://github.com/stormlightlabs/noteleaf
cd noteleaf
go build -o ./tmp/noteleaf ./cmd
```

Install to your GOPATH:

```sh
go install
```

### Database Initialization

Run setup to create the database and configuration file:

```sh
noteleaf setup
```

This creates:

- Database at platform-specific application data directory
- Configuration file at platform-specific config directory
- Default settings for editor, priorities, and display options

### Seeding Sample Data

For exploration and testing, populate the database with example data:

```sh
noteleaf setup seed
```

This creates sample tasks, notes, books, and other items to help you understand the system's capabilities.

## Configuration Overview

Configuration lives in `config.toml` at the platform-specific config directory.

**Editor Settings**:

```toml
[editor]
command = "nvim"
args = []
```

**Task Defaults**:

```toml
[task]
default_priority = "medium"
default_status = "pending"
```

**Display Options**:

```toml
[display]
date_format = "2006-01-02"
time_format = "15:04"
```

View current configuration:

```sh
noteleaf config show
```

Update settings:

```sh
noteleaf config set editor vim
```

See the [Configuration](../Configuration.md) guide for complete options.

## Getting Help

Every command includes help text:

```sh
noteleaf --help
noteleaf task --help
noteleaf task add --help
```

For detailed command reference, run `noteleaf --help` and drill into the subcommand-specific help pages.
