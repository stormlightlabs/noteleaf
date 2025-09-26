# Testing Documentation

This document outlines the testing patterns and practices used in the `noteleaf` application.

## Overview

The codebase follows Go's standard testing practices without external libraries. Tests use only the standard library package and avoid mock frameworks or assertion libraries. This is to keep dependencies minimal and tests readable using standard Go patterns.

### Organization

Each package contains its own test files alongside the source code. Test files are organized by functionality and mirror the structure of the source code they test.

## Patterns

### Handler Creation Pattern

Tests create real handler instances using temporary databases to ensure test isolation. Factory functions handle both database setup and handler initialization, returning both the handler and a cleanup function.

### Database Isolation

Tests use temporary directories and environment variable manipulation to create isolated database instances. Each test gets its own temporary SQLite database that is automatically cleaned up after the test completes.

The `setupCommandTest` function creates a temporary directory, sets `XDG_CONFIG_HOME` to point to it, and initializes the database schema. This ensures tests don't interfere with each other or with development data.

### Resource Management

Tests properly manage resources using cleanup functions returned by factory methods. The cleanup function handles both handler closure and temporary directory removal. This pattern ensures complete resource cleanup even if tests fail.

### Error Handling

Tests use `t.Fatal` for setup errors that prevent test execution and `t.Error` for test assertion failures. Fatal errors stop test execution while errors allow tests to continue checking other conditions.

### Context Cancellation Testing Pattern

Error case testing frequently uses context cancellation to simulate database and network failures. The pattern creates a context, immediately cancels it, then calls the function under test to verify error handling. This provides a reliable way to test error paths without requiring complex mock setups or external failure injection.

### Command Structure Testing

Command group tests verify cobra command structure including use strings, aliases, short descriptions, and subcommand presence. Tests check that commands are properly configured without executing their logic.

### Interface Compliance Testing

Tests verify interface compliance using compile-time checks with blank identifier assignments. This ensures structs implement expected interfaces without runtime overhead.

## Test Organization Patterns

### Single Root Test

The preferred test organization pattern uses a single root test function with nested subtests using `t.Run`. This provides clear hierarchical organization and allows running specific test sections while maintaining shared setup and context. This pattern offers several advantages: clear test hierarchy with logical grouping, ability to run specific test sections, consistent test structure across the codebase, and shared setup that can be inherited by subtests.

### Integration vs Unit Testing

The codebase emphasizes integration testing over heavy mocking by using real handlers and services to verify actual behavior rather than mocked interactions. The goal is to catch integration issues while maintaining test reliability.

### Static Output

UI components support static output modes for testing. Tests capture output using bytes.Buffer and verify content using string contains checks rather than exact string matching for better test maintainability.

### Standard Output Redirection

For testing functions that write to stdout, tests use a pipe redirection pattern with goroutines to capture output. The pattern saves the original stdout, redirects to a pipe, captures output in a separate goroutine, and restores stdout after the test. This ensures clean output capture without interfering with the testing framework.

## Utilities

### Helpers

Test files include helper functions for creating test data and finding elements in collections. These utilities reduce code duplication and improve test readability.

### Mock Data

Tests create realistic mock data using factory functions (powered by faker) that return properly initialized structs with sensible defaults.

## Testing CLI Commands

Command group tests focus on structure verification rather than execution testing. Tests check command configuration, subcommand presence, and interface compliance. This approach ensures command trees are properly constructed without requiring complex execution mocking.

### CommandGroup Interface Testing

The CommandGroup interface enables testable CLI architecture. Tests verify that command groups implement the interface correctly and return properly configured cobra commands. This pattern separates command structure from command execution.

Interface compliance is tested using compile-time checks within the "Interface Implementations" subtest, ensuring all command structs properly implement the CommandGroup interface without runtime overhead.

## Performance Considerations

Tests avoid expensive operations in setup functions. Handler creation uses real instances but tests focus on structure verification rather than full execution paths. This keeps test suites fast while maintaining coverage of critical functionality.

The single root test pattern allows for efficient resource management where setup costs can be amortized across multiple related test cases.

## Interactive Component Testing

Interactive components that use `fmt.Scanf` for user input require special testing infrastructure to prevent tests from hanging while waiting for stdin.

### Testing Success Scenarios

Interactive handlers should test both success and error paths:

- **Valid user selections** - User chooses valid menu options
- **Cancellation** - User chooses to cancel (option 0)
- **Invalid choices** - User selects out-of-range options
- **Empty results** - Search returns no results
- **Network errors** - Service calls fail

This ensures tests run reliably in automated environments while maintaining coverage of the non-interactive code paths.

## Errors

Error coverage follows a systematic approach to identify and test failure scenarios:

1. **Context Cancellation** - Primary method for testing database and network timeout scenarios
2. **Invalid Input** - Malformed data, empty inputs, boundary conditions
3. **Resource Exhaustion** - Database connection failures, memory limits
4. **Constraint Violations** - Duplicate keys, foreign key failures
5. **State Validation** - Testing functions with invalid system states
6. **Interactive Input** - Invalid user choices, cancellation handling, input simulation errors
