# Noteleaf Documentation TODO

This document tracks documentation coverage for the Noteleaf website. The goal is to provide comprehensive documentation for both the productivity system features and the leaflet.pub publishing capabilities.

## Core Concepts

- [ ] Introduction to Noteleaf
  - [ ] What is Noteleaf?
  - [ ] Why terminal-based productivity?
  - [ ] Unified data model philosophy
- [ ] Installation and Setup
  - [ ] System requirements
  - [ ] Building from source
  - [ ] Configuration overview
  - [ ] Database initialization
  - [ ] Seeding sample data
- [ ] Architecture Overview
  - [ ] Application structure
  - [ ] Storage and database
  - [ ] TUI framework (Bubble Tea)
  - [ ] Color palette and design system

## Task Management

- [ ] Task Basics
  - [ ] Creating tasks
  - [ ] Task properties (priority, project, tags, due dates)
  - [ ] Task statuses and lifecycle
  - [ ] Task dependencies
- [ ] Task Operations
  - [ ] Listing and filtering tasks
  - [ ] Updating tasks
  - [ ] Completing and deleting tasks
  - [ ] Recurring tasks
- [ ] Time Tracking
  - [ ] Starting and stopping time tracking
  - [ ] Viewing timesheets
  - [ ] Time reports and analytics
- [ ] Task Organization
  - [ ] Projects
  - [ ] Contexts
  - [ ] Tags
  - [ ] Custom attributes
- [ ] Advanced Task Features
  - [ ] Task queries and filters
  - [ ] Custom views
  - [ ] Batch operations
  - [ ] Task templates

## Note Taking

- [ ] Note Basics
  - [ ] Creating notes
  - [ ] Note metadata (title, tags, dates)
  - [ ] Markdown support
- [ ] Note Operations
  - [ ] Creating from command line vs editor
  - [ ] Creating from files
  - [ ] Editing notes
  - [ ] Reading and viewing notes
  - [ ] Searching notes
- [ ] Note Organization
  - [ ] Tagging system
  - [ ] Linking notes to tasks
  - [ ] Note templates
- [ ] Advanced Note Features
  - [ ] Full-text search
  - [ ] Note exports
  - [ ] Backlinks and references

## Media Tracking

- [ ] Books
  - [ ] Adding books (manual and Open Library integration)
  - [ ] Reading progress tracking
  - [ ] Reading status (to-read, reading, finished)
  - [ ] Book metadata and search
  - [ ] Reading lists
- [ ] Movies
  - [ ] Adding movies
  - [ ] Watch status
  - [ ] Movie metadata
  - [ ] Watchlists
- [ ] TV Shows
  - [ ] Adding TV shows
  - [ ] Watch status
  - [ ] Episode tracking
  - [ ] Show metadata
- [ ] Media Organization
  - [ ] Tags and categories
  - [ ] Custom lists
  - [ ] Ratings and reviews
  - [ ] Linking media to notes

## Articles

- [ ] Article Management
  - [ ] Saving articles from URLs
  - [ ] Article parsing and extraction
  - [ ] Reading articles in terminal
  - [ ] Article metadata (author, date, source)
- [ ] Article Organization
  - [ ] Filtering by author
  - [ ] Tagging articles
  - [ ] Read/unread status
  - [ ] Article archives

## Leaflet.pub Publishing

- [ ] Introduction to Leaflet
  - [ ] What is leaflet.pub?
  - [ ] ATProto and decentralized publishing
  - [ ] How Noteleaf integrates with Leaflet
- [ ] Publications
  - [ ] Creating publications
  - [ ] Publication metadata
  - [ ] Managing publications
- [ ] Documents
  - [ ] Document structure
  - [ ] Creating documents
  - [ ] Document drafts
  - [ ] Publishing workflow
- [ ] Rich Text and Blocks
  - [ ] Text blocks
  - [ ] Headers
  - [ ] Code blocks
  - [ ] Images and media
  - [ ] Blockquotes
  - [ ] Lists
  - [ ] Horizontal rules
- [ ] Rich Text Formatting
  - [ ] Bold, italic, code
  - [ ] Links
  - [ ] Strikethrough, underline, highlight
  - [ ] Facets and styling
- [ ] Publishing Workflow
  - [ ] Converting notes to leaflet documents
  - [ ] Draft management
  - [ ] Publishing to leaflet.pub
  - [ ] Updating published documents
  - [ ] Authentication and identity

## Configuration

- [ ] Configuration File
  - [ ] Config file location and structure
  - [ ] Configuration options reference
  - [ ] Editor integration
  - [ ] Default values
- [ ] Environment Variables
  - [ ] Supported environment variables
  - [ ] Overriding defaults
- [ ] Customization
  - [ ] Custom themes (if supported)
  - [ ] Keyboard shortcuts
  - [ ] Output formats

## TUI (Terminal UI)

- [ ] Interactive Mode
  - [ ] Navigation
  - [ ] Keyboard shortcuts
  - [ ] Selection and actions
  - [ ] Help screens
- [ ] Static Mode
  - [ ] Command-line output
  - [ ] Scripting with Noteleaf
  - [ ] Output formatting
  - [ ] JSON output

## Integration and Workflows

- [ ] External Integrations
  - [ ] Open Library API
  - [ ] Leaflet.pub API
  - [ ] ATProto authentication
- [ ] Workflows and Examples
  - [ ] Daily task review workflow
  - [ ] Note-taking for research
  - [ ] Reading list management
  - [ ] Publishing a blog post to leaflet.pub
  - [ ] Linking tasks, notes, and media
- [ ] Import/Export
  - [ ] Exporting data
  - [ ] Backup and restore
  - [ ] Migration from other tools

## CLI Reference

- [ ] Command Structure
  - [ ] Global flags
  - [ ] Command hierarchy
  - [ ] Help system
- [ ] Commands by Category
  - [ ] `task` commands
  - [ ] `note` commands
  - [ ] `media` commands (book, movie, tv)
  - [ ] `article` commands
  - [ ] `pub` commands (leaflet publishing)
  - [ ] `config` commands
  - [ ] `setup` commands
  - [ ] `status` commands
- [ ] Development Tools
  - [ ] `tools` subcommand
  - [ ] Documentation generation
  - [ ] Lexicon fetching
  - [ ] Database utilities

## Development

- [ ] Building Noteleaf
  - [ ] Development vs production builds
  - [ ] Build tags
  - [ ] Task automation (Taskfile)
- [ ] Testing
  - [ ] Running tests
  - [ ] Coverage reports
  - [ ] Test patterns and scaffolding
- [ ] Contributing
  - [ ] Code organization
  - [ ] Adding new commands
  - [ ] UI components
  - [ ] Testing requirements
- [ ] Architecture Deep Dive
  - [ ] Repository pattern
  - [ ] Handler architecture
  - [ ] Service layer
  - [ ] Data models
  - [ ] UI component system

## Troubleshooting

- [ ] Common Issues
  - [ ] Database errors
  - [ ] Configuration problems
  - [ ] Integration failures
- [ ] Debugging
  - [ ] Verbose output
  - [ ] Log files
  - [ ] Development mode
- [ ] FAQ
  - [ ] General questions
  - [ ] Platform-specific issues
  - [ ] Performance optimization

## Examples and Tutorials

- [ ] Getting Started Tutorial
  - [ ] First 15 minutes
  - [ ] Essential workflows
- [ ] Task Management Tutorials
  - [ ] GTD workflow
  - [ ] Time blocking
  - [ ] Project planning
- [ ] Note-Taking Tutorials
  - [ ] Zettelkasten method
  - [ ] Research notes
  - [ ] Meeting notes
- [ ] Publishing Tutorials
  - [ ] Writing a blog post
  - [ ] Creating a publication
  - [ ] Managing drafts
- [ ] Advanced Tutorials
  - [ ] Scripting with Noteleaf
  - [ ] Custom automation
  - [ ] Data analysis

## Appendices

- [ ] Glossary
- [ ] Keyboard Shortcuts Reference
- [ ] Configuration Options Reference
- [ ] API Reference (leaflet schema)
- [ ] Color Palette Reference
- [ ] Migration Guides
  - [ ] From TaskWarrior
  - [ ] From todo.txt
  - [ ] From other note-taking apps
