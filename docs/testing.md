# Testing Documentation

This document outlines the testing patterns and practices used in the noteleaf application.

## Testing Principles

The codebase follows Go's standard testing practices without external libraries. Tests use only the standard library `testing` package and avoid mock frameworks or assertion libraries. This keeps dependencies minimal and tests readable using standard Go patterns.

## Test File Organization

Test files follow the standard Go convention of `*_test.go` naming. Each package contains its own test files alongside the source code. Test files are organized by functionality and mirror the structure of the source code they test.

## Testing Patterns

### Handler Creation Pattern

Tests create real handler instances using temporary databases to ensure test isolation. Factory functions handle both database setup and handler initialization, returning both the handler and a cleanup function.

### Database Isolation

Tests use temporary directories and environment variable manipulation to create isolated database instances. Each test gets its own temporary SQLite database that is automatically cleaned up after the test completes.

The `setupCommandTest` function creates a temporary directory, sets `XDG_CONFIG_HOME` to point to it, and initializes the database schema. This ensures tests don't interfere with each other or with development data.

### Resource Management

Tests properly manage resources using cleanup functions returned by factory methods. The cleanup function handles both handler closure and temporary directory removal. This pattern ensures complete resource cleanup even if tests fail.

### Error Handling

Tests use `t.Fatal` for setup errors that prevent test execution and `t.Error` for test assertion failures. Fatal errors stop test execution while errors allow tests to continue checking other conditions.

### Command Structure Testing

Command group tests verify cobra command structure including use strings, aliases, short descriptions, and subcommand presence. Tests check that commands are properly configured without executing their logic.

### Interface Compliance Testing

Tests verify interface compliance using compile-time checks with blank identifier assignments. This ensures structs implement expected interfaces without runtime overhead.

```go
var _ CommandGroup = NewTaskCommands(handler)
```

## Test Organization Patterns

### Single Root Test Pattern

The preferred test organization pattern uses a single root test function with nested subtests using `t.Run`. This provides clear hierarchical organization and allows running specific test sections while maintaining shared setup and context.

```go
func TestCommandGroup(t *testing.T) {
    t.Run("Interface Implementations", func(t *testing.T) {
        // Test interface compliance
    })

    t.Run("Create", func(t *testing.T) {
        t.Run("TaskCommand", func(t *testing.T) {
            // Test task command creation
        })
        t.Run("MovieCommand", func(t *testing.T) {
            // Test movie command creation
        })
    })
}
```

This pattern offers several advantages: clear test hierarchy with logical grouping, ability to run specific test sections with `go test -run TestCommandGroup/Create/TaskCommand`, consistent test structure across the codebase, and shared setup that can be inherited by subtests.

### Integration vs Unit Testing

The codebase emphasizes integration testing over heavy mocking. Tests use real handlers and services to verify actual behavior rather than mocked interactions. This approach catches integration issues while maintaining test reliability.

### Static Output Testing

UI components support static output modes for testing. Tests capture output using bytes.Buffer and verify content using string contains checks rather than exact string matching for better test maintainability.

## Test Utilities

### Helper Functions

Test files include helper functions for creating test data and finding elements in collections. These utilities reduce code duplication and improve test readability.

### Mock Data Creation

Tests create realistic mock data using factory functions that return properly initialized structs with sensible defaults. This approach provides consistent test data across different test cases.

## Testing CLI Commands

Command group tests focus on structure verification rather than execution testing. Tests check command configuration, subcommand presence, and interface compliance. This approach ensures command trees are properly constructed without requiring complex execution mocking.

### CommandGroup Interface Testing

The CommandGroup interface enables testable CLI architecture. Tests verify that command groups implement the interface correctly and return properly configured cobra commands. This pattern separates command structure from command execution.

Interface compliance is tested using compile-time checks within the "Interface Implementations" subtest, ensuring all command structs properly implement the CommandGroup interface without runtime overhead.

## Performance Considerations

Tests avoid expensive operations in setup functions. Handler creation uses real instances but tests focus on structure verification rather than full execution paths. This keeps test suites fast while maintaining coverage of critical functionality.

The single root test pattern allows for efficient resource management where setup costs can be amortized across multiple related test cases.

## Best Practices Summary

Use factory functions for test handler creation with proper cleanup patterns. Organize tests using single root test functions with nested subtests for clear hierarchy. Manage resources with cleanup functions returned by factory methods. Prefer integration testing over mocking for real-world behavior validation. Verify interface compliance at compile time within dedicated subtests. Focus command tests on structure verification rather than execution testing. Leverage the single root test pattern for logical grouping and selective test execution. Use realistic test data with factory functions for consistent test scenarios.
