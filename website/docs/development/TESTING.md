# Testing Documentation

This document outlines the testing patterns and practices used in the `noteleaf` application.

## Overview

The codebase follows Go's standard testing practices with specialized testing utilities for complex scenarios.
Tests use the standard library along with carefully selected dependencies like faker for data generation and BubbleTea for TUI testing.
This approach keeps dependencies minimal while providing robust testing infrastructure for interactive components and complex integrations.

### Organization

Each package contains its own test files alongside the source code. Test files are organized by functionality and mirror the structure of the source code they test.
The codebase includes four main test utility files that provide specialized testing infrastructure:

- `internal/services/test_utilities.go` - HTTP mocking and media service testing
- `internal/repo/test_utilities.go` - Database testing and data generation
- `internal/ui/test_utilities.go` - TUI testing framework and interactive component testing
- `internal/handlers/test_utilities.go` - Handler testing with database isolation and input simulation

## Patterns

### Handler Creation

Tests create real handler instances using temporary databases to ensure test isolation.
Factory functions handle both database setup and handler initialization, returning both the handler and a cleanup function.

### Database Isolation

Tests use temporary directories and environment variable manipulation to create isolated database instances.
Each test gets its own temporary SQLite database that is automatically cleaned up after the test completes.

The `setupCommandTest` function creates a temporary directory, sets `XDG_CONFIG_HOME` to point to it, and initializes the database schema.
This ensures tests don't interfere with each other or with development data.

### Resource Management

Tests properly manage resources using cleanup functions returned by factory methods.
The cleanup function handles both handler closure and temporary directory removal.
This pattern ensures complete resource cleanup even if tests fail.

### Error Handling

Tests use `t.Fatal` for setup errors that prevent test execution and `t.Error` for test assertion failures.
Fatal errors stop test execution while errors allow tests to continue checking other conditions.

### Context Cancellation

Error case testing frequently uses context cancellation to simulate database and network failures.
The pattern creates a context, immediately cancels it, then calls the function under test to verify error handling.
This provides a reliable way to test error paths without requiring complex mock setups or external failure injection.

### Command Structure

Command group tests verify cobra command structure including use strings, aliases, short descriptions, and subcommand presence.
Tests check that commands are properly configured without executing their logic.

### Interface Compliance

Tests verify interface compliance using compile-time checks with blank identifier assignments.
This ensures structs implement expected interfaces without runtime overhead.

## Test Infrastructure

### Test Utility Frameworks

The codebase provides comprehensive testing utilities organized by layer and functionality.
Each test utility file contains specialized helpers, mocks, and test infrastructure for its respective domain.

#### Database Testing Utilities

`internal/repo/test_utilities.go` provides comprehensive database testing infrastructure:

- **In-Memory Database Creation**: `CreateTestDB` creates isolated SQLite databases with full schema
- **Sample Data Factories**: Functions like `CreateSampleTask`, `CreateSampleBook` generate realistic test data
- **Faker Integration**: Uses jaswdr/faker for generating realistic fake data with `CreateFakeArticle`
- **Test Setup Helpers**: `SetupTestData` creates a full set of sample data across all models
- **Custom Assertions**: Generic assertion helpers like `AssertEqual`, `AssertContains`, `AssertNoError`

#### HTTP Service Testing

`internal/services/test_utilities.go` provides HTTP mocking and media service testing:

- **Mock Configuration**: `MockConfig` structure for configuring service behavior
- **Function Replacement**: `SetupMediaMocks` replaces service functions with controllable mocks
- **Sample Data Access**: Helper functions that use embedded HTML samples for realistic testing
- **Specialized Scenarios**: Pre-configured mock setups for success and failure scenarios
- **Assertion Helpers**: Domain-specific assertions for movies, TV shows, and error conditions

#### TUI Testing Framework

`internal/ui/test_utilities.go` provides a comprehensive BubbleTea testing framework:

- **TUITestSuite**: Complete testing infrastructure for interactive TUI components
- **Controlled I/O**: `ControlledOutput` and `ControlledInput` for deterministic testing
- **Message Simulation**: Key press simulation, message queuing, and timing control
- **State Verification**: Model state checking and view content assertions
- **Timeout Handling**: Configurable timeouts for async operations
- **Mock Repository**: Test doubles for repository interfaces

#### Handler Testing Infrastructure

`internal/handlers/test_utilities.go` provides end-to-end handler testing:

- **Environment Isolation**: `HandlerTestHelper` creates isolated test environments
- **Input Simulation**: `InputSimulator` for testing interactive components that use `fmt.Scanf`
- **HTTP Mocking**: Comprehensive HTTP server mocking for external API testing
- **Database Helpers**: Database corruption and error scenario testing
- **Editor Mocking**: `MockEditor` for testing file editing workflows
- **Assertion Helpers**: Handler-specific assertions and verification functions

### Advanced Testing Patterns

#### Input Simulation for Interactive Components

Interactive handlers that use `fmt.Scanf` require special testing infrastructure with an `io.Reader` implementation.

The `InputSimulator` provides controlled input sequences that prevent tests from hanging while maintaining coverage of interactive code paths.

#### TUI Testing with BubbleTea Framework

The TUI testing framework addresses the fundamental challenge of testing interactive terminal applications in a deterministic, concurrent environment.
BubbleTea's message-passing architecture creates unique testing requirements that standard Go testing patterns cannot adequately address.

The framework implements a controlled execution environment that replaces BubbleTea's typical program loop with a deterministic testing harness.
Rather than running an actual terminal program, the "testing suite" directly manages model state transitions by simulating the Update/View cycle.
This approach eliminates the non-deterministic behavior inherent in real terminal interactions while preserving the exact message flow patterns that production code experiences.

State verification relies on function composition patterns where test conditions are expressed as closures that capture specific model states.
The `WaitFor` mechanism uses polling with configurable timeouts, addressing the async nature of BubbleTea model updates without creating race conditions.
This pattern bridges imperative test assertions with BubbleTea's declarative update model.
This is inspired by front-end/TS/JS testing patterns.

The framework's I/O abstraction layer replaces terminal input/output with controlled buffers that implement standard Go interfaces.
This design maintains interface compatibility while providing complete control over timing and content.
The controlled I/O system captures all output for later verification and injects precise input sequences, enabling complex interaction testing without external dependencies.

Concurrency management uses channels and context propagation to coordinate between the testing framework and the model under test.
The suite manages goroutine lifecycle and ensures proper cleanup, preventing test interference and resource leaks.
This architecture supports testing of models that perform background operations or handle async events.

#### HTTP Service Mocking

Service testing uses HTTP mocking with request capture. A `MockServer` is instantiated, and its URL is used in test scoped services.

#### Database Schema Testing

Database tests use comprehensive schema setup with (automatic) cleanup

#### Environment Manipulation

Environment testing utilities provide controlled environment manipulation. Environment variables are restored after instantiation.

## Test Organization Patterns

### Single Root Test

The preferred test organization pattern uses a single root test function with nested subtests using `t.Run`.
This provides clear hierarchical organization and allows running specific test sections while maintaining shared setup and context.
This pattern offers several advantages: clear test hierarchy with logical grouping, ability to run specific test sections, consistent test structure across the codebase, and shared setup that can be inherited by subtests.

### Integration vs Unit Testing

The codebase emphasizes integration testing over heavy mocking by using real handlers and services to verify actual behavior rather than mocked interactions.
The goal is to catch integration issues while maintaining test reliability.

### Static Output

UI components support static output modes for testing. Tests capture output using bytes.Buffer and verify content using string contains checks rather than exact string matching for better test maintainability.

### Standard Output Redirection

For testing functions that write to stdout, tests use a pipe redirection pattern with goroutines to capture output.
The pattern saves the original stdout, redirects to a pipe, captures output in a separate goroutine, and restores stdout after the test.
This ensures clean output capture without interfering with the testing framework.

## Utilities

### Test Data Generation

The codebase uses sophisticated data generation strategies:

- **Factory Functions**: Each package provides factory functions for creating valid test data
- **Faker Integration**: Uses `jaswdr/faker` for generating realistic fake data with proper randomization
- **Sample Data Creators**: Functions like `CreateSampleTask`, `CreateSampleBook` provide consistent test data
- **Embedded Resources**: Services use embedded HTML samples from real API responses for realistic testing

### Assertion Helpers

Custom assertion functions provide clear error messages and reduce test code duplication:

- **Generic Assertions**: `AssertEqual`, `AssertNoError`, `AssertContains` for common checks
- **Domain-Specific Assertions**: `AssertMovieInResults`, `AssertNoteExists` for specialized verification
- **TUI Assertions**: `AssertViewContains`, `AssertModelState` for BubbleTea model testing
- **HTTP Assertions**: `AssertRequestMade` for verifying HTTP interactions

### Mock Infrastructure

Each layer provides specialized mocking capabilities:

- **Service Mocking**: Function replacement with configurable behavior and embedded test data
- **HTTP Mocking**: `HTTPMockServer` with request capture and response customization
- **Input Mocking**: `InputSimulator` for deterministic interactive component testing
- **Editor Mocking**: `MockEditor` for file editing workflow testing
- **Repository Mocking**: `MockTaskRepository` for TUI component testing

### Environment and Resource Management

Testing utilities provide comprehensive resource management:

- **Environment Isolation**: `EnvironmentTestHelper` for controlled environment variable manipulation
- **Database Isolation**: Temporary SQLite databases with automatic cleanup
- **File System Isolation**: Temporary directories with automatic cleanup
- **Process Isolation**: Handler helpers that create completely isolated test environments

## Testing CLI Commands

Command group tests focus on structure verification rather than execution testing.
Tests check command configuration, subcommand presence, and interface compliance. This approach ensures command trees are properly constructed without requiring complex execution mocking.

### CommandGroup Interface Testing

The CommandGroup interface enables testable CLI architecture. Tests verify that command groups implement the interface correctly and return properly configured cobra commands.
This pattern separates command structure from command execution.

Interface compliance is tested using compile-time checks within the "Interface Implementations" subtest, ensuring all command structs properly implement the CommandGroup interface without runtime overhead.

## Performance Considerations

Tests avoid expensive operations in setup functions. Handler creation uses real instances but tests focus on structure verification rather than full execution paths.
This keeps test suites fast while maintaining coverage of critical functionality.

The single root test pattern allows for efficient resource management where setup costs can be amortized across multiple related test cases.

## Interactive Component Testing

The codebase provides comprehensive testing infrastructure for interactive components, including both terminal UI applications and command-line interfaces that require user input.

### Input Simulation Framework

Interactive handlers that use `fmt.Scanf` require specialized testing infrastructure:

- **InputSimulator**: Provides controlled input sequences that implement `io.Reader`
- **Menu Selection Helpers**: `MenuSelection`, `MenuCancel`, `MenuSequence` for common interaction patterns
- **Handler Integration**: Handlers can accept `io.Reader` for input, enabling deterministic testing
- **Cleanup Management**: Automatic cleanup prevents resource leaks in test environments

### TUI Testing with BubbleTea

The TUI testing framework provides complete testing infrastructure for interactive terminal interfaces:

- **TUITestSuite**: Comprehensive testing framework for BubbleTea models
- **Message Simulation**: Key press simulation, window resize events, and custom message handling
- **State Verification**: Model state checking with custom condition functions
- **View Assertions**: Content verification and output capture
- **Timing Control**: Configurable timeouts and delay handling for async operations
- **Mock Integration**: Repository mocking for isolated component testing

### Interactive Test Scenarios

Interactive handlers should test comprehensive scenarios:

- **Valid user selections** - User chooses valid menu options and inputs
- **Cancellation flows** - User chooses to cancel operations (option 0 or escape keys)
- **Invalid choices** - User selects out-of-range options or provides invalid input
- **Navigation patterns** - Keyboard navigation, scrolling, and multi-step interactions
- **Error handling** - Network errors, service failures, and data validation errors
- **Empty states** - Search returns no results, empty lists, and missing data
- **Edge cases** - Boundary conditions, malformed input, and resource constraints

### TUI Component Testing Patterns

BubbleTea components use specialized testing patterns:

- **Key Sequence Testing**: Simulate complex user interactions with timing
- **State Transition Testing**: Verify model state changes through user actions
- **View Content Testing**: Assert specific content appears in rendered output
- **Async Operation Testing**: Handle loading states and network operations
- **Responsive Design Testing**: Test different terminal sizes and window resize handling

This comprehensive testing approach ensures interactive components work reliably in automated environments while maintaining full coverage of user interaction paths.

## Errors

Error coverage follows a systematic approach to identify and test failure scenarios:

1. **Context Cancellation** - Primary method for testing database and network timeout scenarios
2. **Invalid Input** - Malformed data, empty inputs, boundary conditions
3. **Resource Exhaustion** - Database connection failures, memory limits
4. **Constraint Violations** - Duplicate keys, foreign key failures
5. **State Validation** - Testing functions with invalid system states
6. **Interactive Input** - Invalid user choices, cancellation handling, input simulation errors
