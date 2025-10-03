# ROADMAP

Noteleaf is a command-line and TUI tool for managing tasks, notes, media, and articles. This roadmap outlines milestones: current capabilities, planned baseline features (v1), and future directions.

## Core Usability

The foundation across all domains is implemented. Tasks support CRUD operations, projects, tags, contexts, and time tracking. Notes have create, list, read, edit, and remove commands with interactive and static modes. Media queues exist for books, movies, and TV with progress and status management. SQLite persistence is in place with setup, seed, and reset commands. TUIs and colorized output are available.

## RC

### CORE

- [ ] Ensure **all documented subcommands** exist and work:
    - Tasks: add, list, view, update, edit, delete, projects, tags, contexts, done, start, stop, timesheet
    - Notes: create, list, read, edit, remove
    - Books: add, list, reading, finished, remove, progress, update
    - Movies: add, list, watched, remove
    - TV: add, list, watching, watched, remove
    - Articles: add, list, view, read, remove
- [ ] Confirm all **aliases** work (`todo`, `ls`, `rm`, etc.).
- [ ] Verify **flags** and argument parsing match man page (priority, project, context, due, tags, etc.).
- [ ] Implement or finish stubs (e.g. `config management` noted in code).

### Task Management Domain

- [ ] Verify tasks can be created with all attributes (priority, project, context, due date, tags).
- [ ] Confirm task listing supports interactive and static modes.
- [ ] Implement status filtering (`pending`, `completed`, etc.).
- [ ] Validate time tracking (start/stop) writes entries and timesheet summarizes correctly.
- [ ] Ensure update supports add/remove tags and all fields.
- [ ] Test interactive editor (`task edit`).

### Notes Domain

- [ ] Implement note creation from:
    - Inline text
    - File (`--file`)
    - Interactive input (`--interactive`)
- [ ] Verify note list interactive TUI works, static list fallback works.
- [ ] Confirm filtering by tags and `--archived`.
- [ ] Ensure notes can be opened, edited in `$EDITOR`, and deleted.

### Media Domains

#### Books

- [ ] Implement search + add (possibly external API).
- [ ] Verify list supports statuses (`queued`, `reading`, `finished`).
- [ ] Progress updates (`book progress`) work with percentages.
- [ ] Status update (`book update`) accepts valid values.

#### Movies

- [ ] Implement search + add.
- [ ] Verify `list` with status filtering (`all`, `queued`, `watched`).
- [ ] Confirm `watched`/`remove` commands update correctly.

#### TV

- [ ] Implement search + add.
- [ ] Verify `list` with multiple statuses (`queued`, `watching`, `watched`).
- [ ] Ensure `watching`, `watched`, `remove` commands behave correctly.

#### Articles

- [ ] Implement article parser (XPath/domain-specific rules).
- [ ] Save articles in Markdown + HTML.
- [ ] Verify metadata is stored in DB.
- [ ] Confirm list supports query, author filter, limit.
- [ ] Test article view/read/remove.

### Configuration & Data

- [ ] Implement **config management** (flagged TODO in code).
- [ ] Define config file format (TOML, YAML, JSON).
- [ ] Set default config/data paths:
    - Linux: `~/.config/noteleaf`, `~/.local/share/noteleaf`
    - macOS: `~/Library/Application Support/noteleaf`
    - Windows: `%APPDATA%\noteleaf`
- [ ] Implement overrides with environment variables (`NOTELEAF_CONFIG`, `NOTELEAF_DATA_DIR`).
- [ ] Ensure consistent DB schema migrations and versioning.

### Documentation

- [ ] Finalize **man page** (plaintext + roff).
- [ ] Write quickstart guide in `README.md`.
- [ ] Add examples for each command.
- [ ] Document config file with defaults and examples.
- [ ] Provide developer docs for contributing.

### QA

- [ ] Verify **unit tests** for all handlers (TaskHandler, NoteHandler, Media Handlers).
- [ ] Write **integration tests** covering CLI flows.
- [ ] Ensure error handling works for:
    - Invalid IDs
    - Invalid flags
    - Schema corruption (already tested in repo)
- [ ] Test cross-platform behavior (Linux/macOS/Windows).

### Packaging

- [ ] Provide prebuilt binaries (via GoReleaser).
- [ ] Add installation instructions (Homebrew, AUR, Scoop, etc.).
- [ ] Version bump to `v1.0.0-rc1`.
- [ ] Publish man page with release.
- [ ] Verify `noteleaf --version` returns correct string.

## v1 Features

Planned functionality for a complete baseline release.

### Tasks

- [ ] Model
    - [ ] Dependencies
    - [ ] Recurrence (`recur`, `until`, templates)
    - [ ] Wait/scheduled dates
    - [ ] Urgency scoring
- [ ] Operations
    - [ ] `annotate`
    - [ ] Bulk edit and undo/history
    - [ ] `$EDITOR` integration
- [ ] Reports and Views
    - [ ] Next actions
    - [ ] Completed/waiting/blocked reports
    - [ ] Calendar view
    - [ ] Sorting and urgency-based views
- [ ] Queries and Filters
    - [ ] Rich query language
    - [ ] Saved filters and aliases
- [ ] Interoperability
    - [ ] JSON import/export
    - [ ] todo.txt compatibility

### Notes

- [ ] Commands
    - [ ] `note search`
    - [ ] `note tag`
    - [ ] `note recent`
    - [ ] `note templates`
    - [ ] `note archive`
    - [ ] `note export`
- [ ] Features
    - [ ] Full-text search
    - [ ] Linking between notes, tasks, and media

### Media

- [ ] Articles/papers/blogs
    - [ ] Parser with domain-specific rules
    - [ ] Commands: `add`, `list`, `view`, `remove`
    - [ ] Metadata validation and storage
- [ ] Books
    - [ ] Source tracking and ratings
    - [ ] Genre/topic tagging
- [ ] Movies/TV
    - [ ] Ratings and notes
    - [ ] Genre/topic tagging
    - [ ] Episode/season progress for TV
    - [ ] Platform/source tracking

## Beyond v1

Features that demonstrate Go proficiency and broaden Noteleaf’s scope.

### Tasks

- [ ] Parallel report generation and background services
- [ ] Hook system for task lifecycle events
- [ ] Plugin mechanism
- [ ] Generics-based filter engine
- [ ] Functional options for configuration
- [ ] Error handling with wrapping and sentinel checks

### Notes

- [ ] Templates system for note types
- [ ] Versioning and history
- [ ] Export with formatting
- [ ] Import from other systems

### Media

- [ ] External imports (Goodreads, IMDB, Letterboxd)
- [ ] Cross-referencing across media types
- [ ] Analytics: velocity, completion rates

### Articles

- [ ] Enhanced parsing coverage
- [ ] Export to multiple formats
- [ ] Linking with tasks and notes

### User Experience

- [ ] Shell completions
- [ ] Manpages and docs generator
- [ ] Theming and customizable output
- [ ] Calendar integration

### Tasks

- [ ] Sub-tasks and hierarchical tasks
- [ ] Visual dependency mapping
- [ ] Forecasting and smart suggestions
- [ ] Habit and streak tracking
- [ ] Context-aware recommendations

### Notes

- [ ] Graph view of linked notes
- [ ] Content extraction and summarization
- [ ] Encryption and privacy controls

### Media

- [ ] Podcast and YouTube management
- [ ] Multi-format (audiobooks, comics)
- [ ] Media consumption goals and streaks
- [ ] Media budget tracking
- [ ] Seasonal and energy-based filtering

### Articles

- [ ] Content validation
- [ ] Encryption support
- [ ] Advanced classification

## Technical Infrastructure

### Completed

SQLite persistence, CI with GitHub Actions and Codecov, TUIs with Charm stack, initial help system.

### Planned

- Prebuilt binaries for releases
- Installation and usage documentation
- Contribution guide and developer docs
- Consistent argument parsing
- Backup/restore
- Multiple profiles
- Optional synchronization

## v1 Feature Matrix

| Domain   | Feature               | Status    |
|----------|-----------------------|-----------|
| Tasks    | CRUD                  | Complete  |
| Tasks    | Projects/tags         | Complete  |
| Tasks    | Time tracking         | Complete  |
| Tasks    | Dependencies          | Planned   |
| Tasks    | Recurrence            | Planned   |
| Tasks    | Wait/scheduled        | Planned   |
| Tasks    | Urgency scoring       | Planned   |
| Notes    | CRUD                  | Complete  |
| Notes    | Search/tagging        | Planned   |
| Media    | Books/movies/TV       | Complete  |
| Media    | Articles              | Planned   |
| Media    | Source/ratings        | Planned   |
| Articles | Parser + storage      | Planned   |
| System   | SQLite persistence    | Complete  |
| System   | Synchronization       | Future    |
| System   | Import/export formats | Future    |
