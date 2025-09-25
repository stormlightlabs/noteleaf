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

### Testing

`CommandGroup` structs accept handlers as constructor parameters, enabling easy dependency injection of mock handlers for testing.
Command logic can be tested independently of handler implementations. The interface allows mocking entire command groups for integration testing.

### Registry

`main.go` uses a registry pattern to organize command groups by category. Core commands include task, note, and media functionality.
Management commands handle configuration, setup, and maintenance operations. The pattern provides clean separation and easy extension for new command groups.

## UI and Styling System

The application uses a structured color palette system located in `internal/ui/colors.go` for consistent terminal output styling.

### Color Architecture

The color system implements a `Key` type with 74 predefined colors from the Charm ecosystem, including warm tones (Cumin, Tang, Paprika), cool tones (Sapphire, Oceania, Zinc), and neutral grays (Pepper through Butter). Each color provides hex values via the `Hex()` method and implements Go's `color.Color` interface through `RGBA()`.

### Predefined Styles

Three core lipgloss styles handle common UI elements:

- `TitleColorStyle` uses color 212 with bold formatting for command titles
- `SelectedColorStyle` provides white-on-212 highlighting for selected items
- `HeaderColorStyle` applies color 240 with bold formatting for section headers

### Color Categories

Colors are organized into primary, secondary, and tertiary categories are accessed through `IsPrimary()`, `IsSecondary()`, and `IsTertiary()` methods on the `Key` type.

### Lipgloss Integration

The styling system integrates with the Charmbracelet lipgloss library for terminal UI rendering.
Colors from the `Key` type convert to lipgloss color values through their `Hex()` method. The predefined `TitleColorStyle`, `SelectedColorStyle`, and `HeaderColorStyle` variables provide lipgloss styles that can be applied to strings with `.Render()`.
