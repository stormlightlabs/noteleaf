# ROADMAP

## Task Management Commands (TaskWarrior-inspired)

### Implemented Commands

- [x] `todo add [description]` - Create new task with metadata (priority, project, context, due, tags)
- [x] `todo list` - Display tasks with filtering (status, priority, project, context) and interactive/static modes
- [x] `todo view [task-id]` - View task details with format options (detailed, brief, json)
- [x] `todo update [task-id]` - Edit task properties via flags
- [x] `todo edit [task-id]` - Interactive task editor with status picker and priority toggle
- [x] `todo done [task-id]` - Mark task as completed
- [x] `todo delete [task-id]` - Remove task permanently

---

- [x] `todo projects` - List all project names (interactive/static modes)
- [x] `todo tags` - List all tag names (interactive/static modes)
- [x] `todo contexts` - List all contexts/locations (interactive/static modes)

---

- [x] `todo start [task-id]` - Start time tracking for a task
- [x] `todo stop [task-id]` - Stop time tracking for a task
- [x] `todo timesheet` - Show time tracking summaries (with date range and task filters)

### Commands To Be Implemented

- [ ] Due dates & scheduling - Including recurring tasks
- [ ] Task dependencies - Task A blocks task B relationships
- [ ] `annotate` - Add notes/comments to existing tasks
- [ ] Recurring tasks
- [ ] Smart due date suggestions
- [ ] Completion notifications
- [ ] `calendar` - Display tasks in calendar view

## Media Queue Management Commands

### Implemented Commands

Book Management

- [x] `media book add [search query...]` - Search and add book to reading list (with interactive mode)
- [x] `media book list` - Show reading queue with progress and status filtering
- [x] `media book reading <id>` - Mark book as currently reading
- [x] `media book finished <id>` - Mark book as completed
- [x] `media book remove <id>` - Remove from reading list
- [x] `media book progress <id> <percentage>` - Update reading progress (0-100%)
- [x] `media book update <id> <status>` - Update book status (queued|reading|finished|removed)

Movie Management

- [x] `media movie add [title]` - Add movie to watch queue (with interactive mode)
- [x] `media movie list` - Show movie queue with status filtering
- [x] `media movie watched <id>` - Mark movie as watched
- [x] `media movie remove <id>` - Remove from queue

TV Show Management

- [x] `media tv add [title]` - Add TV show/season to queue (with interactive mode)
- [x] `media tv list` - Show TV queue with status filtering
- [x] `media tv watching <id>` - Mark TV show as currently watching
- [x] `media tv watched <id>` - Mark episodes/seasons as watched
- [x] `media tv remove <id>` - Remove from TV queue

### Commands To Be Implemented

---

- [ ] Articles, papers, blogs support (implement article parser)
- [ ] Source tracking (recommendation sources)
- [ ] Ratings and personal notes
- [ ] Genre/topic tagging
- [ ] Episode/season progress tracking for TV
- [ ] Platform tracking (Netflix, Amazon, etc.)
- [ ] Watch status: queued, watching, completed, dropped

## Management Commands

### Implemented Commands

Application Management

- [x] `status` - Show application status and configuration
- [x] `setup` - Initialize and manage application setup
- [x] `setup seed` - Populate database with test data (with --force flag)
- [x] `reset` - Reset the application (removes all data)
- [x] `config [key] [value]` - Manage configuration settings (stubbed)

### Commands To Be Implemented

Organization Features

- [ ] Custom queries and saved searches
- [ ] Context-aware suggestions
- [ ] Overdue/urgent highlighting
- [ ] Recently added/modified items
- [ ] Seasonal/mood-based filtering
- [ ] Full-text search across titles, notes, tags

Analytics

- [ ] Reading/watching velocity tracking
- [ ] Completion rates by content type
- [ ] Time investment analysis
- [ ] Personal productivity metrics
- [ ] Content source analysis

Integrations

- [ ] `import` - Import from various formats (CSV, JSON, todo.txt)
- [ ] `export` - Export to various formats
- [ ] Goodreads import for books
- [ ] IMDB/Letterboxd import for movies
- [ ] Todo.txt format compatibility
- [ ] TaskWarrior import/export
- [ ] URL parsing for automatic metadata

`todo.txt` Compatibility

- [ ] `archive` - Move completed tasks to done.txt
- [ ] `[con]texts` - List all contexts (@context)
- [ ] `[proj]ects` - List all projects (+project)
- [ ] `[pri]ority` - Set task priority (A-Z)
- [ ] `[depri]oritize` - Remove priority from task
- [ ] `[re]place` - Replace task text entirely
- [ ] `[pre]pend/[app]end` - Add text to beginning/end of task

Automation

- [ ] Auto-categorization of new items
- [ ] Smart due date suggestions
- [ ] Recurring content (weekly podcast check-ins)
- [ ] Completion notifications

Storage

- [ ] `sync` - Synchronize with remote storage
- [ ] `sync setup` - Setup remote storage
- [ ] Local SQLite database with optional cloud sync
- [ ] Multiple profile support
- [ ] `backup` - Create local backup
- [ ] Backup/restore functionality

Configuration

- [x] Enhanced `config` command implementation (basic stubbed version)
- [ ] `undo` - Reverse last operation
- [ ] Themes and personalization
- [ ] Customizable output formats

## Notes Management Commands

### Implemented Commands

Core Notes Operations

- [x] `note create [title] [content...]` - Create new markdown note with optional interactive editor
- [x] `note list` - Interactive TUI browser for navigating and viewing notes (with archive and tag filtering)
- [x] `note read <note-id>` - Display formatted note content with syntax highlighting
- [x] `note edit <note-id>` - Edit note in configured editor
- [x] `note remove <note-id>` - Permanently remove note file and metadata

Additional Options

- [x] `--interactive|-i` flag for create command (opens editor)
- [x] `--file|-f` flag for create command (create from markdown file)
- [x] `--archived|-a` flag for list command
- [x] `--tags` filtering for list command

### Commands To Be Implemented

- [ ] `note search [query]` - Search notes by content, title, or tags
- [ ] `note tag <note-id> [tags...]` - Add/remove tags from notes
- [ ] `note recent` - Show recently created/modified notes
- [ ] `note templates` - Create notes from predefined templates
- [ ] `note archive <note-id>` - Archive old notes
- [ ] `note export` - Export notes to various formats
- [ ] Full-text search integration
- [ ] Linking between notes and tasks/content

## User Experience

- [x] Interactive TUI modes for task lists, projects, tags, contexts, and notes
- [x] Static output modes as alternatives to interactive TUI
- [x] Color-coded priority and status indicators
- [x] Comprehensive help system via cobra CLI framework

---

- [ ] Quick-add commands for rapid entry
- [ ] Enhanced progress tracking UI
- [ ] Calendar view for tasks

### Technical Infrastructure

- [ ] CI/CD pipeline -> pre-build binaries
- [ ] Complete README/documentation
- [ ] Installation instructions
- [ ] Usage examples

## Tech Debt

### Signatures

We've got inconsistent argument parsing and sanitization leading to calls to strconv.Atoi in tests & handler funcs.
This is only done correctly in the note command -> handler sequence

- [ ] TaskCommand
- [ ] MovieCommand
- [ ] TVCommand
- [ ] BookCommand

### Movie Commands - Missing Tests

- [x] movie watched [id] - marks movie as watched

### TV Commands - Missing Tests

- [x] tv watching [id] - marks TV show as watching
- [x] tv watched [id] - marks TV show as watched

### Book Commands - Missing Tests

- [x] book add [search query...] - search and add book
- [x] book reading `<id>` - marks book as reading
- [x] book finished `<id>` - marks book as finished
- [x] book progress `<id>` `<percentage>` - updates reading progress

## Ideas

### Task Management Enhancements

- Sub-tasks and hierarchical tasks - Break complex tasks into child tasks for better organization
- Linking - Establish relationships between related tasks without strict dependencies
- Batching - Group related tasks for bulk operations (completion, priority changes, etc.)
- Retrospectives - Analysis of completed tasks to improve future estimates and planning
- Automation rules - Create rules that automatically modify tasks based on conditions
- Habit formation - Track recurring micro-tasks that build into larger goals
- Context switching - Automatically adjust system settings, apps, or environment based on current task
- Forecasting - Predict future tasks based on patterns, calendar events, or seasonal trends
- "Energy" matching - Recommend tasks based on current energy levels or time of day
- Priority rebalancing - Automatically suggest priority adjustments based on deadlines and importance
- Dependency visualization: Visual flow charts showing how tasks connect and block each other

### Media Management Enhancements

- Podcast management: Add podcast tracking with episode progress and subscription management
- YouTube/video content management: Track video content queues and viewing progress
- Multi-format media support: Include audiobooks, comics, and other content formats
- Media consumption goals: Set reading/watching goals (e.g., "2 books per month")
- Media cross-referencing: Connect related content across different media types
    - Series
- Media note integration: Link notes to specific books, movies, or shows for reviews
    - Review system - Write and store personal reviews of consumed content
- Media budget tracking: Track spending on media content (subscriptions, purchases)
- Media consumption patterns: Analyze personal consumption patterns and preferences over time
- Media seasonal tracking: Track seasonal media preferences and suggest accordingly
- Media completion streaks: Gamification elements for consistent media consumption
- Media progress synchronization: Sync progress across different devices or platforms

### Notes Management Enhancements

- Linking and graph view: Create bidirectional links between related notes
- Templates system: Predefined templates for different note types (meeting notes, book summaries)
- Versioning and history: Track changes to notes over time with ability to revert
- Export with formatting: Rich export options with preserved formatting and links
- Import capabilities: Import notes from other systems (Notion, Evernote, etc.)
- Content extraction: Extract key points or action items from longer notes
- Encryption: End-to-end encryption for sensitive notes
- Content validation: Check for broken links, missing references, or inconsistent information

### System Integration & Automation

- Calendar integration: Sync tasks with calendar systems (Google Calendar, Outlook)
- Email integration: Create tasks or notes from emails automatically
- Browser extension: Quick capture of web content as tasks or notes
- IDE/plugin integration: Direct integration with code editors and development environments
- File system integration: Monitor files for content that should become tasks or notes

### Advanced UI/UX Features

- Customizable themes: Multiple visual themes and color schemes
- Terminal interface enhancements: Rich TUI with advanced navigation and visualization
- Web-based interface: Alternative web UI for browser-based access
- Advanced filtering and sorting: Complex query systems for data manipulation
- Visual task mapping: Gantt charts, Kanban boards, and other visual representations
- Quick entry mode: Rapid capture interface for minimal friction
- Keyboard customization: Fully customizable keyboard shortcuts

### Security and Privacy

- End-to-end encryption: Full encryption of sensitive data
- Local-first architecture: Guarantee that all data remains local
