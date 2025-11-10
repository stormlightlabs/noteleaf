# ROADMAP

Noteleaf is a command-line and TUI tool for managing tasks, notes, media, and articles. This roadmap outlines milestones: current capabilities, planned baseline features (v1), and future directions.

## Core Usability

The foundation across all domains is implemented. Tasks support CRUD operations, projects, tags, contexts, and time tracking.
Notes have create, list, read, edit, and remove commands with interactive and static modes. Media queues exist for books, movies, and TV with progress and status management. SQLite persistence is in place with setup, seed, and reset commands. TUIs and colorized output are available.

## RC

### CORE

- [x] Ensure **all documented subcommands** exist and work:
    - Tasks: add, list, view, update, edit, delete, projects, tags, contexts, done, start, stop, timesheet
    - Notes: create, list, read, edit, remove
    - Books: add, list, reading, finished, remove, progress, update
    - Movies: add, list, watched, remove
    - TV: add, list, watching, watched, remove
    - Articles: add, list, view, read, remove
- [x] Confirm all **aliases** work (`todo`, `ls`, `rm`, etc.).
- [x] Verify **flags** and argument parsing match man page (priority, project, context, due, tags, etc.).
- [x] Implement or finish stubs (e.g. `config management` noted in code).

### Task Management Domain

- [x] Verify tasks can be created with all attributes (priority, project, context, due date, tags).
- [x] Confirm task listing supports interactive and static modes.
- [x] Implement status filtering (`pending`, `completed`, etc.).
- [x] Validate time tracking (start/stop) writes entries and timesheet summarizes correctly.
- [x] Ensure update supports add/remove tags and all fields.
- [x] Test interactive editor (`task edit`).

### Notes Domain

- [x] Implement note creation from:
    - Inline text
    - File (`--file`)
    - Interactive input (`--interactive`)
- [x] Verify note list interactive TUI works, static list fallback works.
- [x] Confirm filtering by tags and `--archived`.
- [x] Ensure notes can be opened, edited in `$EDITOR`, and deleted.

#### Publication

- [x] Implement authentication with BlueSky/leaflet (AT Protocol).
    - [ ] Add [OAuth2](#publications--authentication)
- [x] Verify `pub pull` fetches and syncs documents from leaflet.
- [x] Confirm `pub list` with status filtering (`all`, `published`, `draft`).
- [x] Test `pub post` creates new documents with draft/preview/validate modes.
- [x] Ensure `pub patch` updates existing documents correctly.
- [x] Validate `pub push` handles batch operations (create/update).
- [x] Verify markdown conversion to leaflet block format (headings, code, images, facets).

### Media Domains

#### Books

- [x] Implement search + add (possibly external API).
- [x] Verify list supports statuses (`queued`, `reading`, `finished`).
- [x] Progress updates (`book progress`) work with percentages.
- [x] Status update (`book update`) accepts valid values.

#### Movies

- [x] Implement search + add.
- [x] Verify `list` with status filtering (`all`, `queued`, `watched`).
- [x] Confirm `watched`/`remove` commands update correctly.

#### TV

- [x] Implement search + add.
- [x] Verify `list` with multiple statuses (`queued`, `watching`, `watched`).
- [x] Ensure `watching`, `watched`, `remove` commands behave correctly.

#### Articles

- [x] Implement article parser (XPath/domain-specific rules).
- [x] Save articles in Markdown + HTML.
- [x] Verify metadata is stored in DB.
- [x] Confirm list supports query, author filter, limit.
- [x] Test article view/read/remove.

### Configuration & Data

- [x] Implement **config management**
- [x] Define config file format (TOML, YAML, JSON).
- [x] Set default config/data paths:
    - Linux: `~/.config/noteleaf`, `~/.local/share/noteleaf`
    - macOS: `~/Library/Application Support/noteleaf`
    - Windows: `%APPDATA%\noteleaf`
- [x] Implement overrides with environment variables (`NOTELEAF_CONFIG`, `NOTELEAF_DATA_DIR`).
- [x] Ensure consistent DB schema migrations and versioning.

### Documentation

- [x] Finalize **man page** - use `tools/docgen.go` as a dev only command for `website/docs/manual`
- [x] Strictly follow <https://diataxis.fr/>
    - [x] Write quickstart guide in `README.md` & add `website/docs/Quickstart.md`
    - [x] Add examples for each command (`website/docs/examples`)
    - [x] Document config file with defaults and examples in `website/docs/Configuration.md`
    - [x] Provide developer docs for contributing in `docs/dev`
        - [x] Move to `website/docs/development`

### QA

- [x] Verify **unit tests** for all handlers (TaskHandler, NoteHandler, Media Handlers).
- [x] Write **integration tests** covering CLI flows.
- [x] Ensure error handling works for:
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
- [ ] Books (Open Library integration enhancements)
    - [ ] Author detail fetching (full names, bio)
    - [ ] Edition-specific metadata
    - [ ] Cover image download and caching
    - [ ] Reading progress tracking
    - [ ] Personal reading lists sync
- [ ] Movies/TV (external API integration)
    - [ ] Movie databases (TMDb, OMDb)
    - [ ] Rotten Tomatoes integration
- [ ] Music
    - [ ] Music services (MusicBrainz, Album of the Year)

### Articles

- [ ] Enhanced parsing coverage
- [ ] Export to multiple formats
- [ ] Linking with tasks and notes

### Publications & Authentication

- [ ] OAuth2 authentication for AT Protocol
    - [ ] Client metadata server for publishing application details
    - [ ] DPoP (Demonstrating Proof of Possession) implementation
        - [ ] ES256 JWT generation with unique JTI nonces
        - [ ] Server-issued nonce management with 5-minute rotation
        - [ ] Separate nonce tracking for authorization and resource servers
    - [ ] PAR (Pushed Authorization Requests) flow
        - [ ] PKCE code challenge generation
        - [ ] State token management
        - [ ] Request URI handling
    - [ ] Identity resolution and verification
        - [ ] Bidirectional handle verification
        - [ ] DID resolution from handles
        - [ ] Authorization server discovery via .well-known endpoints
    - [ ] Token lifecycle management
        - [ ] Access token refresh (5-15 min lifetime recommended)
        - [ ] Refresh token rotation (180 day max for confidential clients)
        - [ ] Concurrent request handling to prevent duplicate refreshes
        - [ ] Secure token storage (encrypted at rest)
    - [ ] Local callback server for OAuth redirects
        - [ ] Ephemeral HTTP server on localhost
        - [ ] Browser launch integration
        - [ ] Timeout handling for abandoned flows
    - [ ] Support both OAuth & App Passwords but recommend OAuth
- [ ] Leaflet.pub enhancements
    - [ ] Multiple Publications: Manage separate publications for different topics
    - [ ] Image Upload: Automatically upload images to blob storage and embed in documents
    - [ ] Status Management: Publish drafts and unpublish documents from CLI
    - [ ] Metadata Editing: Update document titles, summaries, and tags
    - [ ] Backlink Support: Parse and resolve cross-references between documents
    - [ ] Offline Mode: Queue posts and patches for later upload

### User Experience

- [ ] Shell completions
- [ ] Manpages and docs generator
- [ ] Theming and customizable output
- [ ] Calendar integration
- [ ] Task synchronization services
- [ ] Git repository linking
- [ ] Note export to other platforms

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

### Local API Server

A local HTTP server daemon that exposes Noteleaf data for web UIs and extensions.
Runs on the user's machine and provides programmatic access to tasks, notes, and media.

#### Architecture

- [ ] Daemon mode via `noteleaf server start/stop/status`
- [ ] Binds to localhost by default (configurable port)
- [ ] HTTP/REST API using existing repository layer
- [ ] Shares same SQLite database as CLI
- [ ] Middleware: logging, CORS (for localhost web UIs), compression
- [ ] Health and status endpoints

#### Daemon Management

- [ ] Commands: `start`, `stop`, `restart`, `status`
- [ ] PID file tracking for process management
- [ ] Systemd service file for Linux
- [ ] launchd plist for macOS
- [ ] Graceful shutdown with active request draining
- [ ] Auto-restart on crash option
- [ ] Configurable bind address and port
- [ ] Log file rotation

#### API Endpoints

RESTful design matching CLI command structure:

- [ ] `GET /api/v1/tasks` - List tasks with filters
- [ ] `POST /api/v1/tasks` - Create task
- [ ] `GET /api/v1/tasks/:id` - Get task details
- [ ] `PUT /api/v1/tasks/:id` - Update task
- [ ] `DELETE /api/v1/tasks/:id` - Delete task
- [ ] `POST /api/v1/tasks/:id/start` - Start time tracking
- [ ] `POST /api/v1/tasks/:id/stop` - Stop time tracking
- [ ] `GET /api/v1/tasks/:id/time-entries` - Get time entries
- [ ] Similar CRUD endpoints for notes, books, movies, TV shows, articles
- [ ] `GET /api/v1/projects` - List all projects
- [ ] `GET /api/v1/tags` - List all tags
- [ ] `GET /api/v1/contexts` - List all contexts
- [ ] `GET /api/v1/stats` - Dashboard statistics

#### Real-time Updates

- [ ] WebSocket endpoint for live data updates
- [ ] Server-Sent Events (SSE) as fallback
- [ ] Event types: task created/updated/deleted, note modified, etc.
- [ ] Subscribe to specific domains or IDs
- [ ] Change notification for web UI reactivity

#### Authentication & Security

- [ ] Optional API token authentication (disabled by default for localhost)
- [ ] Token stored in config file
- [ ] Token rotation command
- [ ] CORS configuration for allowed origins
- [ ] Localhost-only binding by default (security through network isolation)
- [ ] Optional TLS for local network access

#### Extension System

- [ ] Webhook endpoints for extension registration
- [ ] Event hooks for task/note lifecycle:
    - [ ] Before/after create, update, delete
    - [ ] Task completion, start, stop
    - [ ] Note archive/unarchive
- [ ] Webhook delivery with retries
- [ ] Extension manifest for discovery
- [ ] JavaScript plugin API (embedded V8/goja runtime)
- [ ] Plugin sandbox for security

#### Web UI

- [ ] Reference web UI implementation
- [ ] Static file serving from embedded assets
- [ ] Single-page application architecture
- [ ] Responsive design (desktop, tablet, mobile)
- [ ] Features:
    - [ ] Task board view (Kanban)
    - [ ] Calendar view for tasks
    - [ ] Note editor with Markdown preview
    - [ ] Media queue management
    - [ ] Search and filtering
    - [ ] Keyboard shortcuts matching CLI

#### Configuration

- [ ] Server config section in noteleaf.toml:
    - [ ] bind_address (default: 127.0.0.1)
    - [ ] port (default: 8080)
    - [ ] enable_auth (default: false)
    - [ ] api_token (optional)
    - [ ] enable_websocket (default: true)
    - [ ] log_level (default: info)
- [ ] Environment variable overrides
- [ ] CLI flag overrides for daemon commands

#### Monitoring & Diagnostics

- [ ] `GET /health` - Health check endpoint
- [ ] `GET /metrics` - Prometheus-compatible metrics
- [ ] Request logging (access log)
- [ ] Error logging with stack traces
- [ ] Performance metrics (request duration, DB query time)
- [ ] Active connections and goroutine count
- [ ] Memory and CPU usage stats

#### Client Libraries

- [ ] Go client library for extensions
- [ ] JavaScript/TypeScript client for web UIs
- [ ] OpenAPI/Swagger specification
- [ ] Auto-generated API documentation

## Technical Infrastructure

### Completed

SQLite persistence, CI with GitHub Actions and Codecov, TUIs with Charm stack, initial help system.

### Planned

- Prebuilt binaries for releases
- Installation and usage documentation
- Contribution guide and developer docs
- Consistent argument parsing

#### Post v1

- Backup/restore
    - [ ] Automated backups
    - [ ] Backup scheduling and rotation
- Multiple profiles
- Optional synchronization
    - [ ] Sync service
- Import/Export
    - [ ] CSV export for tasks
    - [ ] Markdown export for tasks
    - [ ] Bulk export commands
    - [ ] Migration utilities (TaskWarrior, todo.txt, etc.)
    - [ ] Git integration for notes/data versioning

## v1 Feature Matrix

| Domain       | Feature                    | Status    |
|--------------|----------------------------|-----------|
| Tasks        | CRUD                       | Complete  |
| Tasks        | Projects/tags              | Complete  |
| Tasks        | Time tracking              | Complete  |
| Tasks        | Dependencies               | Complete  |
| Tasks        | Recurrence                 | Complete  |
| Tasks        | Wait/scheduled             | Planned   |
| Tasks        | Urgency scoring            | Planned   |
| Notes        | CRUD                       | Complete  |
| Notes        | Search/tagging             | Planned   |
| Publications | AT Protocol sync           | Complete  |
| Publications | Post/patch/push            | Complete  |
| Publications | Markdown conversion        | Complete  |
| Publications | OAuth2                     | Future    |
| Media        | Books/movies/TV            | Complete  |
| Media        | Articles                   | Complete  |
| Media        | Source/ratings             | Planned   |
| Articles     | Parser + storage           | Complete  |
| System       | SQLite persistence         | Complete  |
| System       | Configuration management   | Complete  |
| System       | Synchronization            | Future    |
| System       | Import/export formats      | Future    |
