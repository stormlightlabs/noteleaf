---
title: CLI Reference
sidebar_label: CLI Reference
description: Overview of Noteleaf’s command hierarchy, flags, and developer utilities.
sidebar_position: 7
---

# CLI Reference

This reference is a map of the top-level commands exposed by `noteleaf`. For flag-by-flag detail run `noteleaf <command> --help`—the human-friendly Fang help screens are always the source of truth.

## Command Structure

### Global flags

| Flag                      | Description                                                       |
| ------------------------- | ----------------------------------------------------------------- |
| `--help`, `-h`            | Show help for any command or subcommand                           |
| `--version`               | Print the Noteleaf build string (includes git SHA when available) |
| `--color <auto\|on\|off>` | Optional Fang flag to control ANSI colors                         |

Environment variables such as `NOTELEAF_CONFIG`, `NOTELEAF_DATA_DIR`, and `EDITOR` affect how commands behave but are not flags.

### Command hierarchy

- Root command: `noteleaf`
- Task commands live under the `todo` alias (e.g., `noteleaf todo add`).
- Media commands are grouped and require a subtype: `noteleaf media book`, `noteleaf media movie`, `noteleaf media tv`.
- Publishing flows live under `noteleaf pub`.
- Management helpers (`config`, `setup`, `status`, `reset`) sit at the top level.

### Help system

Every command inherits Fang’s colorized help plus Noteleaf-specific additions:

- `noteleaf article --help` prints the supported parser domains and storage directory by calling into the handler.
- Interactive commands show the keyboard shortcuts inside their help output.
- You can always drill down: `noteleaf todo add --help`, `noteleaf media book list --help`, etc.

## Commands by Category

### `todo` / `task`

Add, list, view, update, complete, and annotate tasks. Supports priorities, contexts, tags, dependencies, recurrence, and JSON output for scripting. Related metadata commands (`projects`, `tags`, `contexts`) summarize usage counts.

### `note`

Create Markdown notes (inline, from files, or via the interactive editor), list them with the TUI, search, view, edit in `$EDITOR`, archive/unarchive, and delete. Notes share IDs with leaflet publishing so they can be synced later.

### `media`

Umbrella group for personal queues:

- `noteleaf media book` — Search Open Library, add books, update status (`queued`/`reading`/`finished`), edit progress percentages, and remove titles.
- `noteleaf media movie` — Search Rotten Tomatoes, queue movies, mark them watched, or remove them.
- `noteleaf media tv` — Same as movies but with watching/watched states and optional season/episode tracking.

Each subtype has its own `list`, status-changing verbs, and removal commands. Use `-i/--interactive` on `add` to open the TUI selector (books today, other media soon).

### `article`

Parse and save web articles with `add <url>`, inspect them via `list`, `view`, or `read`, and delete them with `remove`. All commands operate on the local Markdown/HTML archive referenced in the handler output.

### `pub`

Leaflet.pub commands for AT Protocol publishing:

- `pull` / `push` to sync notes with the remote publication.
- `status`, `list`, and `diff` to inspect what is linked.
- Support for working drafts, batch pushes, and file-based imports (`--file`) when publishing is combined with local markdown.

### `config`

Inspect and mutate `~/.noteleaf.conf.toml`:

- `noteleaf config show` (or `get <key>`) prints values.
- `noteleaf config set <key> <value>` writes back to disk.
- `noteleaf config path` reveals the file location.
- `noteleaf config reset` rewinds to defaults.

### `setup`

`noteleaf setup` initializes the database, config file, and data directories if they do not exist. `noteleaf setup seed` can load sample data (pass `--force` to wipe existing rows first).

### `status`

`noteleaf status` prints absolute paths for the config file, data directory, database, and media folders along with environment overrides—handy for debugging or verifying a portable install.

## Development Tools

`noteleaf tools ...` is available in development builds (`task build:dev`, `go run ./cmd`). It bundles maintenance utilities:

### Documentation generation

```
noteleaf tools docgen --format docusaurus --out website/docs/manual
noteleaf tools docgen --format man --out docs/manual
```

Generates reference docs straight from the command definitions, keeping terminal help and published docs in sync.

### Lexicon fetching

```
noteleaf tools fetch lexicons
noteleaf tools fetch lexicons --sha <commit>
```

Pulls the latest `leaflet.pub` lexicons from GitHub so the AT Protocol client stays current. You can point it at a specific commit for reproducible builds.

### Database utilities

```
noteleaf tools fetch gh-repo --repo owner/repo --path schemas --output tmp/schemas
```

Provides generic fetchers plus helpers used by CI and local testing to refresh schema files, warm caches, or introspect the SQLite database.

These tools intentionally live behind the dev build tag so production binaries stay lean. Use them when contributing documentation or publishing features.
