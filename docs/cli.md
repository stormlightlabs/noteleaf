# CLI Docs

## CommandGroup Interface Pattern

This section outlines the CommandGroup interface pattern for implementing CLI commands in the noteleaf application.

### Core Concepts

Each major command group implements the CommandGroup interface with a `Create() *cobra.Command` method. Command groups receive handlers as constructor dependencies, enabling dependency injection for testing. Handler initialization occurs centrally in main.go with `log.Fatalf` error handling to fail fast during application startup.

### CommandGroup Interface

interface `CommandGroup` provides a consistent contract for all command groups. Each implementation encapsulates related commands and the shared handler dependency. The Create method returns a fully configured cobra command tree.

#### Implementations

TaskCommands handles todo and task-related operations using TaskHandler. MovieCommand manages movie queue operations via MovieHandler.
TVCommand handles TV show queue operations through TVHandler. NoteCommand manages note operations using NoteHandler.

### Handler Lifecycle

Handlers are created once in `main.go` during application startup. Initialization errors prevent application launch rather than causing runtime failures.
Handlers persist for the application lifetime without requiring cleanup. Commands access handlers through struct fields rather than creating new instances.

### Testing Benefits

`CommandGroup` structs accept handlers as constructor parameters, enabling easy dependency injection of mock handlers for testing.
Command logic can be tested independently of handler implementations. The interface allows mocking entire command groups for integration testing.

### Registry Pattern

`main.go` uses a registry pattern to organize command groups by category. Core commands include task, note, and media functionality.
Management commands handle configuration, setup, and maintenance operations. The pattern provides clean separation and easy extension for new command groups.
