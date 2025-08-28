# ROADMAP

## Core Task Management (TaskWarrior-inspired)

### Basic Operations

- [x] `create|new` - Add new task with description and optional metadata
- [x] `list` - Display tasks with filtering and sorting options
- [x] `view` - View task by ID
- [x] `update` - Edit task properties (description, priority, project, tags)
- [x] `done` - Mark task as completed
- [x] `delete` - Remove task permanently

### Organization & Metadata

- [x] Project & context organization
    - [x] `projects` - List all project names
- [x] Tag management system
    - [x] `tags` - List all tag names
- [ ] Status tracking - todo, in-progress, blocked, done, abandoned
- [ ] Priority system - High/medium/low or numeric scales
- [ ] Due dates & scheduling - Including recurring tasks
- [ ] Task dependencies - Task A blocks task B relationships

### Time Management

- [ ] Time tracking functionality
    - [ ] `start/stop` - Track active time on tasks
    - [ ] `timesheet` - Show time tracking summaries
    - [ ] `calendar` - Display tasks in calendar view

### Advanced Task Features

- [ ] `annotate` - Add notes/comments to existing tasks
- [ ] Recurring tasks
- [ ] Smart due date suggestions
- [ ] Completion notifications

## Content Queue Management

### Reading Management

- [x] `book add` - Add book to reading list
- [x] `book list` - Show reading queue with progress
- [x] `book reading` - Mark book as currently reading
- [x] `book finished|read` - Mark book as completed
- [x] `book remove|rm` - Remove from reading list
- [x] `book progress` - Update reading progress percentage

#### Enhanced Reading Features

- [ ] Articles, papers, blogs support (implement article parser)
- [ ] Reading status: want-to-read, currently-reading, completed, abandoned
- [ ] Source tracking (recommendation sources)
- [ ] Ratings and personal notes
- [ ] Genre/topic tagging
- [ ] Progress tracking (pages/chapters read, completion %)

### Watching Management

- [ ] `movie add` - Add movie to watch queue
- [ ] `movie list` - Show movie queue with ratings/metadata
- [ ] `movie watched|seen` - Mark movie as watched
- [ ] `movie remove|rm` - Remove from queue

- [ ] `tv add` - Add TV show/season to queue
- [ ] `tv list` - Show TV queue with episode tracking
- [ ] `tv watched|seen` - Mark episodes/seasons as watched
- [ ] `tv remove|rm` - Remove from TV queue

#### Enhanced Watching Features

- [ ] Episode/season progress tracking
- [ ] Watch status: queued, watching, completed, dropped
- [ ] Platform tracking (Netflix, Amazon, etc.)
- [ ] Ratings and reviews
- [ ] Genre/mood tagging

## Organization & Discovery Features

### Smart Views & Filtering

- [ ] Custom queries and saved searches
- [ ] Context-aware suggestions
- [ ] Overdue/urgent highlighting
- [ ] Recently added/modified items
- [ ] Seasonal/mood-based filtering
- [ ] Full-text search across titles, notes, tags

### Analytics & Insights

- [ ] Reading/watching velocity tracking
- [ ] Completion rates by content type
- [ ] Time investment analysis
- [ ] Personal productivity metrics
- [ ] Content source analysis

## Advanced Workflow Features

### Integration & Import

- [ ] `import` - Import from various formats (CSV, JSON, todo.txt)
- [ ] `export` - Export to various formats
- [ ] Goodreads import for books
- [ ] IMDB/Letterboxd import for movies
- [ ] Todo.txt format compatibility
- [ ] TaskWarrior import/export
- [ ] URL parsing for automatic metadata

### Todo.txt Compatibility

- [ ] `archive` - Move completed tasks to done.txt
- [ ] `[con]texts` - List all contexts (@context)
- [ ] `[proj]ects` - List all projects (+project)
- [ ] `[pri]ority` - Set task priority (A-Z)
- [ ] `[depri]oritize` - Remove priority from task
- [ ] `[re]place` - Replace task text entirely
- [ ] `prepend/append` - Add text to beginning/end of task

### Automation

- [ ] Auto-categorization of new items
- [ ] Smart due date suggestions
- [ ] Recurring content (weekly podcast check-ins)
- [ ] Completion notifications

## Data Management

### Storage & Sync

- [ ] `sync` - Synchronize with remote storage
- [ ] `sync setup` - Setup remote storage
- [ ] Local SQLite database with optional cloud sync
- [ ] Multiple profile support
- [ ] `backup` - Create local backup
- [ ] Backup/restore functionality

### Configuration

- [ ] `config` - Manage configuration settings
- [ ] `undo` - Reverse last operation
- [ ] Themes and personalization
- [ ] Customizable output formats

## Notes Management

### Basic Operations

- [x] `create|new` - Creates a new markdown note and optionally opens in configured editor
- [x] `list` - Opens interactive TUI browser for navigating and viewing notes
- [x] `read|view` - Displays formatted note content with syntax highlighting
- [x] `edit|update` - Opens configured editor OR replaces note content with new markdown file
- [x] `remove|rm|delete|del` - Permanently removes the note file and metadata

### Advanced Notes Features

- [ ] `search` - Search notes by content, title, or tags
- [ ] `tag` - Add/remove tags from notes
- [ ] `recent` - Show recently created/modified notes
- [ ] `templates` - Create notes from predefined templates
- [ ] `archive` - Archive old notes
- [ ] `export` - Export notes to various formats
- [ ] Full-text search integration
- [ ] Linking between notes and tasks/content

## User Experience

### Interface

- [ ] Interactive TUI mode for browsing (likely using Bubbletea)
- [ ] Quick-add commands for rapid entry
- [ ] Progress tracking UI
- [ ] Comprehensive help system

### Technical Infrastructure

- [ ] CI/CD pipeline -> pre-build binaries
- [ ] Complete README/documentation
- [ ] Installation instructions
- [ ] Usage examples
