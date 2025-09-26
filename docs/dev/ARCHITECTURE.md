# System Architecture

Noteleaf is a CLI/TUI application for task and content management built on Go with SQLite persistence and terminal-based interfaces.

## Core Architecture

### Application Structure

The application follows a layered architecture with clear separation between presentation, business logic, and data access layers.

**Entry Point** - `cmd/main.go` initializes the application with dependency injection, creates handlers, and configures the Cobra command tree using the CommandGroup pattern.

**CLI Framework** - Built on `spf13/cobra` with `charmbracelet/fang` providing enhanced CLI features including color schemes, versioning, and improved help formatting.

**TUI Components** - Interactive interfaces use `charmbracelet/bubbletea` for state management and `charmbracelet/lipgloss` for styling with a consistent color palette system in `internal/ui/colors.go`.

### Data Layer

**Database** - SQLite with schema migrations in `internal/store/sql/migrations/`. The `internal/store` package manages database connections and configuration.

**Repository Pattern** - Data access abstracts through repository interfaces in `internal/repo/` with validation logic ensuring data integrity at the persistence boundary.

**Models** - Entity definitions in `internal/models/` implement standardized Model interfaces with common fields (ID, Created, Modified).

### Business Logic

**Handlers** - Business logic resides in `internal/handlers/` with one handler per domain (TaskHandler, NoteHandler, ArticleHandler, etc.). Handlers receive repository dependencies through constructor injection.

**Services** - Domain-specific operations in `internal/services/` handle complex business workflows and external integrations.

**Validation** - Schema-based validation at repository level with custom ValidationError types providing detailed field-level error messages.

## Domain Features

### Content Management

**Articles** - Web scraping using `gocolly/colly` with domain-specific extraction rules stored in `internal/articles/rules/`. Articles are parsed to markdown and stored with dual file references (markdown + HTML).

**Tasks** - Todo/task management inspired by TaskWarrior with filtering, status tracking, and interactive TUI views.

**Notes** - Simple note management with markdown support and glamour-based terminal rendering.

**Media Queues** - Separate queues for books, movies, and TV shows with status tracking and metadata management.

### User Interface

**Command Groups** - Commands organized into core functionality (task, note, article, media) and management operations (setup, config, status) using the CommandGroup interface pattern.

**Interactive Views** - Bubbletea-based TUI components for list navigation, item selection, and data entry with consistent styling through the lipgloss color system.

**Terminal Output** - Markdown rendering through `charmbracelet/glamour` for rich text display in terminal environments.

## Dependencies

**CLI/TUI** - Cobra command framework, Bubbletea state management, Lipgloss styling, Fang CLI enhancements.

**Data** - SQLite driver (`mattn/go-sqlite3`), TOML configuration parsing.

**Content Processing** - Colly web scraping, HTML/XML query libraries, Glamour markdown rendering, text processing utilities.

**Utilities** - UUID generation, time handling, logging through `charmbracelet/log`.

## Design Decisions and Tradeoffs

### Technology Choices

**Go over Rust** - Go was selected for its simplicity, excellent CLI ecosystem (Cobra, Charm libraries), and faster development velocity. While Rust + Ratatui would provide better memory safety and potentially superior performance, Go's straightforward concurrency model and mature tooling ecosystem made it the pragmatic choice for rapid prototyping and iteration.

**SQLite over PostgreSQL** - SQLite provides zero-configuration deployment and sufficient performance for single-user CLI applications. The embedded database eliminates setup complexity while supporting full SQL features needed for filtering and querying. PostgreSQL would add deployment overhead without meaningful benefits for this use case.

**Repository Pattern over Active Record** - Repository interfaces enable clean separation between business logic and data access, facilitating testing through dependency injection. This pattern scales better than Active Record for complex domain logic while maintaining clear boundaries between layers.

### Architectural Tradeoffs

**CommandGroup Interface** - Centralizes command registration while enabling modular command organization. The pattern requires additional abstraction but provides consistent dependency injection and testing capabilities across all command groups.

**Handler-based Business Logic** - Business logic in handlers rather than rich domain models keeps the codebase simple and avoids over-engineering. While this approach may not scale to complex business rules, it provides clear separation of concerns for the current feature set.

**Dual Storage for Articles** - Articles store both markdown and HTML versions to balance processing speed with format flexibility. This doubles storage requirements but eliminates runtime conversion overhead and preserves original formatting.

### Component Interactions

**Handler → Repository → Database** - Request flow follows a linear path from CLI commands through business logic to data persistence. This pattern ensures consistent validation and error handling while maintaining clear separation of concerns.

**TUI State Management** - Bubbletea's unidirectional data flow provides predictable state updates for interactive components. The model-view-update pattern ensures consistent UI behavior across different terminal environments.

**Configuration and Migration** - Application startup validates configuration and runs database migrations before initializing handlers. This fail-fast approach prevents runtime errors and ensures consistent database schema across deployments.
