# ROADMAP

## Core Task Management (TaskWarrior-inspired)

- [x] `list` - Display tasks with filtering and sorting options
- [ ] `projects` - List all project names
- [ ] `tags` - List all tag names

- [x] `create|new` - Add new task with description and optional metadata

- [x] `view` - View task by ID
- [x] `done` - Mark task as completed
- [x] `update` - Edit task properties (description, priority, project, tags)
- [ ] `start/stop` - Track active time on tasks
- [ ] `annotate` - Add notes/comments to existing tasks

- [x] `delete` - Remove task permanently

- [ ] `calendar` - Display tasks in calendar view
- [ ] `timesheet` - Show time tracking summaries

## Todo.txt Compatibility

- [ ] `archive` - Move completed tasks to done.txt
- [ ] `[con]texts` - List all contexts (@context)
- [ ] `[proj]ects` - List all projects (+project)
- [ ] `[pri]ority` - Set task priority (A-Z)
- [ ] `[depri]oritize` - Remove priority from task
- [ ] `[re]place` - Replace task text entirely
- [ ] `prepend/append` - Add text to beginning/end of task

## Media Queue Management

- [ ] `movie add` - Add movie to watch queue
- [ ] `movie list` - Show movie queue with ratings/metadata
- [ ] `movie watched|seen` - Mark movie as watched
- [ ] `movie remove|rm` - Remove from queue

- [ ] `tv add` - Add TV show/season to queue
- [ ] `tv list` - Show TV queue with episode tracking
- [ ] `tv watched|seen` - Mark episodes/seasons as watched
- [ ] `tv remove|rm` - Remove from TV queue

## Reading List Management

- [x] `book add` - Add book to reading list
- [x] `book list` - Show reading queue with progress
- [x] `book reading` - Mark book as currently reading
- [x] `book finished|read` - Mark book as completed
- [x] `book remove|rm` - Remove from reading list
- [x] `book progress` - Update reading progress percentage

## Data Management

- [ ] `sync` - Synchronize with remote storage
- [ ] `sync setup` - Setup remote storage

- [ ] `backup` - Create local backup

- [ ] `import` - Import from various formats (CSV, JSON, todo.txt)
- [ ] `export` - Export to various formats

- [ ] `config` - Manage configuration settings

- [ ] `undo` - Reverse last operation

## Notes

- [x] `create|new` - Creates a new markdown note and optionally opens in configured editor
    - Creates a note from existing markdown file content
- [x] `list` - Opens interactive TUI browser for navigating and viewing notes
- [x] `read|view` - Displays formatted note content with syntax highlighting
- [x] `edit|update` - Opens configured editor OR Replaces note content with new markdown file
- [x] `remove|rm|delete|del` - Permanently removes the note file and metadata

- [ ] `search` - Search notes by content, title, or tags
- [ ] `tag` - Add/remove tags from notes
- [ ] `recent` - Show recently created/modified notes
- [ ] `templates` - Create notes from predefined templates
- [ ] `archive` - Archive old notes
- [ ] `export` - Export notes to various formats
