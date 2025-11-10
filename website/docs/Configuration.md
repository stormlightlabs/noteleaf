---
id: configuration
title: Configuration
sidebar_label: Configuration
sidebar_position: 6
description: Reference for configuration locations, defaults, and options.
---

# Configuration

Noteleaf stores its configuration in a TOML file. The configuration file location depends on your operating system and can be overridden with environment variables.

## Configuration File Location

### Default Paths

**Linux:**

```sh
~/.config/noteleaf/.noteleaf.conf.toml
```

**macOS:**

```sh
~/Library/Application Support/noteleaf/.noteleaf.conf.toml
```

**Windows:**

```sh
%APPDATA%\noteleaf\.noteleaf.conf.toml
```

### Environment Variable Override

Set `NOTELEAF_CONFIG` to use a custom configuration file location:

```sh
export NOTELEAF_CONFIG=/path/to/custom/config.toml
```

### File Structure

Configuration is stored as [TOML](https://toml.io). Each key maps 1:1 with the fields shown by `noteleaf config show`. A minimal file looks like:

```toml
date_format = "2006-01-02"
color_scheme = "default"
```

There is no required ordering—the parser loads the file, applies defaults for missing keys, and writes a normalized version whenever you run `noteleaf config set ...`. This means you can edit the file manually or stick entirely to CLI helpers.

## Configuration Options

### General Settings

#### date_format

Format for displaying dates throughout the application.

**Type:** String
**Default:** `"2006-01-02"` (ISO 8601)
**Example:**

```toml
date_format = "2006-01-02"
```

Common formats:

- `"2006-01-02"` - ISO format: 2024-03-15
- `"01/02/2006"` - US format: 03/15/2024
- `"02-Jan-2006"` - Short month: 15-Mar-2024

#### color_scheme

Color scheme for terminal output.

**Type:** String
**Default:** `"default"`
**Example:**

```toml
color_scheme = "default"
```

#### default_view

Default view mode for interactive lists.

**Type:** String
**Default:** `"list"`
**Example:**

```toml
default_view = "list"
```

#### default_priority

Default priority for new tasks when not specified.

**Type:** String
**Default:** None
**Options:** `"low"`, `"medium"`, `"high"`, `"urgent"`
**Example:**

```toml
default_priority = "medium"
```

#### editor

Text editor for editing notes and tasks. Falls back to `$EDITOR` environment variable if not set.

**Type:** String
**Default:** None (uses `$EDITOR`)
**Example:**

```toml
editor = "vim"
```

### Data Storage

#### database_path

Custom path to SQLite database file. Leave empty to use default location.

**Type:** String
**Default:** Platform-specific data directory
**Example:**

```toml
database_path = "/custom/path/noteleaf.db"
```

#### data_dir

Directory for storing application data (articles, notes, attachments).

**Type:** String
**Default:** Platform-specific data directory
**Example:**

```toml
data_dir = "/custom/data/directory"
```

You can also use the `NOTELEAF_DATA_DIR` environment variable:

```sh
export NOTELEAF_DATA_DIR=/custom/data/directory
```

#### articles_dir

Directory for storing saved articles.

**Type:** String
**Default:** `<data_dir>/articles`
**Example:**

```toml
articles_dir = "/path/to/articles"
```

#### notes_dir

Directory for storing notes.

**Type:** String
**Default:** `<data_dir>/notes`
**Example:**

```toml
notes_dir = "/path/to/notes"
```

### Archive and Export

#### auto_archive

Automatically archive completed tasks after a specified period.

**Type:** Boolean
**Default:** `false`
**Example:**

```toml
auto_archive = true
```

#### export_format

Default format for exporting data.

**Type:** String
**Default:** `"json"`
**Options:** `"json"`, `"csv"`, `"markdown"`
**Example:**

```toml
export_format = "json"
```

### Synchronization

Synchronization features are planned for future releases.

#### sync_enabled

Enable synchronization with remote server.

**Type:** Boolean
**Default:** `false`
**Example:**

```toml
sync_enabled = false
```

#### sync_endpoint

URL of the synchronization server.

**Type:** String
**Default:** None
**Example:**

```toml
sync_endpoint = "https://sync.example.com/api"
```

#### sync_token

Authentication token for synchronization.

**Type:** String
**Default:** None
**Example:**

```toml
sync_token = "your-secret-token"
```

### API Keys

#### movie_api_key

API key for movie database services (future feature).

**Type:** String
**Default:** None
**Example:**

```toml
movie_api_key = "your-api-key"
```

#### book_api_key

API key for book database services. Currently uses Open Library which doesn't require an API key.

**Type:** String
**Default:** None
**Example:**

```toml
book_api_key = "your-api-key"
```

### AT Protocol / Bluesky Integration

Configuration for publishing content to Bluesky/AT Protocol.

#### atproto_did

Your Decentralized Identifier (DID) on the AT Protocol network.

**Type:** String
**Default:** None
**Example:**

```toml
atproto_did = "did:plc:abcd1234efgh5678"
```

#### atproto_handle

Your Bluesky/AT Protocol handle.

**Type:** String
**Default:** None
**Example:**

```toml
atproto_handle = "username.bsky.social"
```

#### atproto_pds_url

Personal Data Server URL.

**Type:** String
**Default:** None
**Example:**

```toml
atproto_pds_url = "https://bsky.social"
```

#### atproto_access_jwt

Access token for authentication (managed automatically).

**Type:** String
**Default:** None
**Example:**

```toml
atproto_access_jwt = "eyJhbGc..."
```

#### atproto_refresh_jwt

Refresh token for authentication (managed automatically).

**Type:** String
**Default:** None

#### atproto_expires_at

Token expiration timestamp (managed automatically).

**Type:** String (ISO8601)
**Default:** None

## Editor Integration

The `editor` key wires Noteleaf into your preferred text editor. Resolution order:

1. `editor` inside `.noteleaf.conf.toml`
2. `$EDITOR` environment variable
3. System default (usually `vi` on Unix)

Where it is used:

- `noteleaf note edit <id>` always opens the configured editor.
- `noteleaf note create -e` or `--editor` lets you capture inline text and immediately refine it in the editor.
- Interactive creation (`noteleaf note create -i`) respects the same setting when you choose to open the note.

Example configuration:

```toml
editor = "nvim"
```

If you frequently switch editors, leave the config empty and export `$EDITOR` before launching Noteleaf:

```sh
EDITOR="zed" noteleaf note edit 5
```

## Managing Configuration

### View Current Configuration

```sh
noteleaf config show
```

### Set Configuration Value

```sh
noteleaf config set <key> <value>
```

Examples:

```sh
noteleaf config set editor "nvim"
noteleaf config set default_priority "high"
noteleaf config set date_format "01/02/2006"
```

### Get Configuration Value

```sh
noteleaf config get <key>
```

Example:

```sh
noteleaf config get editor
```

## Example Configuration

Complete example configuration with common settings:

```toml
# General settings
date_format = "2006-01-02"
color_scheme = "default"
default_view = "list"
default_priority = "medium"
editor = "vim"

# Data storage
# database_path = ""  # Use default location
# data_dir = ""       # Use default location

# Archive and export
auto_archive = false
export_format = "json"

# Synchronization (future feature)
sync_enabled = false
# sync_endpoint = ""
# sync_token = ""

# API keys (optional)
# movie_api_key = ""
# book_api_key = ""

# AT Protocol / Bluesky (optional)
# atproto_did = ""
# atproto_handle = ""
# atproto_pds_url = "https://bsky.social"
```

## Environment Variables

Environment overrides are resolved before configuration values. Set them when you need temporary behavior (CI jobs, alternate workspaces, etc.).

| Variable | Purpose | Notes |
|----------|---------|-------|
| `NOTELEAF_CONFIG` | Absolute path to the TOML file | Overrides platform defaults. Parent directories are created automatically. |
| `NOTELEAF_DATA_DIR` | Root directory for the SQLite DB, notes, articles, and attachments | Useful for portable installs (USB drive, synced folder). |
| `EDITOR` | Fallback editor when the `editor` config key is empty | Checked by all note-related commands. |

Usage example:

```sh
export NOTELEAF_CONFIG=~/.config/noteleaf/work.conf.toml
export NOTELEAF_DATA_DIR=~/Sync/workspace-data
export EDITOR=helix
```

Because `NOTELEAF_DATA_DIR` cascades to the article and note directories, a single export is all you need to relocate the entire knowledge base.

## Customization

### Themes and Colors

The `color_scheme` option controls how Fang (the underlying Cobra replacement) styles command help and certain UI components. Right now the only valid value is `"default"`, which maps to Noteleaf’s Iceberg-inspired palette. Future releases will add explicit `light`/`dark` options; until then customization requires overriding your terminal theme or building Noteleaf from source with changes in `internal/ui/palette.go`.

```toml
color_scheme = "default"  # leave blank to adopt upcoming auto-mode
```

### Keyboard Shortcuts

All interactive views share the same key map:

| Keys | Action |
|------|--------|
| `↑ / k`, `↓ / j` | Move selection |
| `enter` | Open the selected row |
| `v` | View details in a side panel (where supported) |
| `/` | Search/filter (live) |
| `r` | Reload data |
| `?` | Show full help, including custom actions for the current view |
| `q`, `ctrl+c` | Quit the view |
| `esc`, `backspace` | Exit search/help/detail panels |
| `1-9` | Jump directly to the corresponding row index |

Press `?` inside any list/table to confirm the bindings—this uses Bubble Tea’s built-in help component so it always reflects the current screen.

### Output Formats

- `export_format` sets the default for future export commands (currently `json`).
- Task commands support JSON today: `noteleaf todo view 12 --json` or `noteleaf todo list --static --json`.
- The `--format` flag on `noteleaf todo view` switches between `detailed` and `brief` layouts, which is handy when scripting.

Examples:

```sh
noteleaf todo view 12 --format brief --json | jq '.status'
noteleaf todo list --static --json > tasks.json
noteleaf config set export_format "csv"   # prepare for upcoming exporters
```

Even when there is no dedicated exporter yet, the SQLite database lives in the open, so you can always run your own `SELECT ...` queries or use `sqlite-utils` to produce CSV/JSON.
