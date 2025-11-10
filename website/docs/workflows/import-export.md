---
title: Import and Export
sidebar_label: Import & Export
sidebar_position: 1
description: Data portability, backups, and migration.
---

# Import and Export

Noteleaf stores data in open formats for portability: SQLite for structured data and Markdown for notes.

## Data Storage

### SQLite Database

Location varies by platform:

**macOS:**

```
~/Library/Application Support/noteleaf/noteleaf.db
```

**Linux:**

```
~/.local/share/noteleaf/noteleaf.db
```

**Windows:**

```
%LOCALAPPDATA%\noteleaf\noteleaf.db
```

### Markdown Files

Notes are stored as individual markdown files:

**Default location:**

```
<data_dir>/notes/
```

Configure via `notes_dir` in `.noteleaf.conf.toml`.

### Articles

Saved articles are stored as markdown:

**Default location:**

```
<data_dir>/articles/
```

Configure via `articles_dir` in `.noteleaf.conf.toml`.

## JSON Export

### Task Export

Export tasks to JSON format:

```sh
noteleaf todo view 123 --json
noteleaf todo list --static --json
```

Output includes all task attributes:

- Description
- Status, priority
- Project, context, tags
- Due dates, recurrence
- Dependencies, parent tasks
- Timestamps

### Export Format Configuration

Set default export format:

```sh
noteleaf config set export_format "json"
```

Options:

- `json` (default)
- `csv` (planned)
- `markdown` (planned)

## Backup Strategy

### Full Backup

Back up the entire data directory:

```sh
# macOS
cp -r ~/Library/Application\ Support/noteleaf ~/Backups/noteleaf-$(date +%Y%m%d)

# Linux
cp -r ~/.local/share/noteleaf ~/backups/noteleaf-$(date +%Y%m%d)
```

Includes:

- SQLite database
- Notes directory
- Articles directory
- Configuration file

### Database Only

```sh
# macOS
cp ~/Library/Application\ Support/noteleaf/noteleaf.db ~/Backups/

# Linux
cp ~/.local/share/noteleaf/noteleaf.db ~/backups/
```

### Notes Only

```sh
# Copy notes directory
cp -r <data_dir>/notes ~/Backups/notes-$(date +%Y%m%d)
```

Notes are plain markdown files, easily versioned with Git:

```sh
cd <data_dir>/notes
git init
git add .
git commit -m "Initial notes backup"
```

## Restore from Backup

### Full Restore

```sh
# Stop noteleaf
# Replace data directory
cp -r ~/Backups/noteleaf-20240315 ~/Library/Application\ Support/noteleaf
```

### Database Restore

```sh
cp ~/Backups/noteleaf.db ~/Library/Application\ Support/noteleaf/
```

### Notes Restore

```sh
cp -r ~/Backups/notes-20240315 <data_dir>/notes
```

## Direct Database Access

SQLite database is accessible with standard tools:

```sh
# Open database
sqlite3 ~/Library/Application\ Support/noteleaf/noteleaf.db

# List tables
.tables

# Query tasks
SELECT id, description, status FROM tasks WHERE status = 'pending';

# Export to CSV
.mode csv
.output tasks.csv
SELECT * FROM tasks;
.quit
```

## Portable Installation

Use environment variables for portable setup:

```sh
export NOTELEAF_DATA_DIR=/path/to/usb/noteleaf-data
export NOTELEAF_CONFIG=/path/to/usb/noteleaf.conf.toml
noteleaf todo list
```

Useful for:

- USB drive installations
- Synced folders (Dropbox, iCloud)
- Multiple workspaces
- Testing environments

## Migration Strategies

### From TaskWarrior

Manual migration via SQLite:

1. Export TaskWarrior data to JSON
2. Parse JSON and insert into noteleaf database
3. Map TaskWarrior attributes to Noteleaf schema

Custom migration script required (future documentation).

### From todo.txt

Convert todo.txt to Noteleaf tasks:

1. Parse todo.txt format
2. Map projects, contexts, priorities
3. Bulk insert via SQLite

Custom migration script required (future documentation).

### From Other Note Apps

Notes are markdown files:

1. Export notes from source app
2. Convert to plain markdown
3. Copy to `<data_dir>/notes/`
4. Noteleaf will index them on next scan

## Sync and Cloud Storage

### Cloud Sync

Store data directory in synced folder:

```sh
# Use Dropbox
export NOTELEAF_DATA_DIR=~/Dropbox/noteleaf-data

# Use iCloud
export NOTELEAF_DATA_DIR=~/Library/Mobile\ Documents/com~apple~CloudDocs/noteleaf
```

**Warning:** SQLite databases don't handle concurrent writes well. Only run one Noteleaf instance at a time per database.

### Version Control

Notes directory can be versioned:

```sh
cd <data_dir>/notes
git init
git add .
git commit -m "Initial commit"
git remote add origin <repository-url>
git push -u origin main
```

Automatic git commits planned for future release.

## Data Formats

### SQLite Schema

View schema:

```sh
sqlite3 noteleaf.db .schema
```

Tables include:

- `tasks` - Task management
- `notes` - Note metadata
- `articles` - Article metadata
- `books`, `movies`, `tv_shows` - Media tracking
- `publications` - Leaflet.pub publications
- Linking tables for tags, dependencies

### Markdown Format

Notes use standard markdown with YAML frontmatter:

```markdown
---
title: Note Title
created: 2024-03-15T10:30:00Z
modified: 2024-03-15T11:00:00Z
tags: [tag1, tag2]
---

# Note Content

Regular markdown content...
```
