---
title: Testing
sidebar_label: Testing
sidebar_position: 2
description: Running tests and understanding test patterns.
---

# Testing

Noteleaf maintains comprehensive test coverage using Go's built-in testing framework with consistent patterns across the codebase.

## Running Tests

### All Tests

```sh
task test
# or
go test ./...
```

### Coverage Report

Generate HTML coverage report:

```sh
task coverage
```

Output: `coverage.html` (opens in browser)

### Terminal Coverage

View coverage in terminal:

```sh
task cov
```

Shows function-level coverage percentages.

### Package-Specific Tests

Test specific package:

```sh
go test ./internal/repo
go test ./internal/handlers
go test ./cmd
```

### Verbose Output

```sh
go test -v ./...
```

## Test Organization

Tests follow a hierarchical 3-level structure:

```go
func TestRepositoryName(t *testing.T) {
    // Setup once
    db := CreateTestDB(t)
    repos := SetupTestData(t, db)

    t.Run("Feature", func(t *testing.T) {
        t.Run("scenario description", func(t *testing.T) {
            // Test logic
        })
    })
}
```

Levels:

1. Package (top function)
2. Feature (first t.Run)
3. Scenario (nested t.Run)

## Test Patterns

### Repository Tests

Repository tests use scaffolding from `internal/repo/test_utilities.go`:

```go
func TestTaskRepository(t *testing.T) {
    db := CreateTestDB(t)
    repos := SetupTestData(t, db)
    ctx := context.Background()

    t.Run("Create", func(t *testing.T) {
        t.Run("creates task successfully", func(t *testing.T) {
            task := NewTaskBuilder().
                WithDescription("Test task").
                Build()

            created, err := repos.Tasks.Create(ctx, task)
            AssertNoError(t, err, "create should succeed")
            AssertEqual(t, "Test task", created.Description, "description should match")
        })
    })
}
```

### Handler Tests

Handler tests use `internal/handlers/handler_test_suite.go`:

```go
func TestHandlerName(t *testing.T) {
    suite := NewHandlerTestSuite(t)
    defer suite.cleanup()
    handler := CreateHandler(t, NewHandlerFunc)

    t.Run("Feature", func(t *testing.T) {
        t.Run("scenario", func(t *testing.T) {
            AssertNoError(t, handler.Method(), "operation should succeed")
        })
    })
}
```

## Test Utilities

### Assertion Helpers

Located in `internal/repo/test_utilities.go` and `internal/handlers/test_utilities.go`:

```go
// Error checking
AssertNoError(t, err, "operation should succeed")
AssertError(t, err, "operation should fail")

// Value comparison
AssertEqual(t, expected, actual, "values should match")
AssertTrue(t, condition, "should be true")
AssertFalse(t, condition, "should be false")

// Nil checking
AssertNil(t, value, "should be nil")
AssertNotNil(t, value, "should not be nil")

// String operations
AssertContains(t, str, substr, "should contain substring")
```

### Test Data Builders

Create test data with builders:

```go
task := NewTaskBuilder().
    WithDescription("Test task").
    WithStatus("pending").
    WithPriority("high").
    WithProject("test-project").
    Build()

book := NewBookBuilder().
    WithTitle("Test Book").
    WithAuthor("Test Author").
    Build()

note := NewNoteBuilder().
    WithTitle("Test Note").
    WithContent("Test content").
    Build()
```

### Test Database

In-memory SQLite for isolated tests:

```go
db := CreateTestDB(t)  // Automatic cleanup via t.Cleanup()
```

### Sample Data

Pre-populated test data:

```go
repos := SetupTestData(t, db)
// Creates tasks, notes, books, movies, TV shows
```

## Test Naming

Use direct descriptions without "should":

```go
t.Run("creates task successfully", func(t *testing.T) { })      // Good
t.Run("should create task", func(t *testing.T) { })             // Bad
t.Run("returns error for invalid input", func(t *testing.T) { }) // Good
```

## Test Independence

Each test must be independent:

- Use `CreateTestDB(t)` for isolated database
- Don't rely on test execution order
- Clean up resources with `t.Cleanup()`
- Avoid package-level state

## Coverage Targets

Maintain high coverage for:

- Repository layer (data access)
- Handler layer (business logic)
- Services (external integrations)
- Models (data validation)

Current coverage visible via:

```sh
task cov
```

## Continuous Integration

Tests run automatically on:

- Pull requests
- Main branch commits
- Release builds

CI configuration validates:

- All tests pass
- No race conditions
- Coverage thresholds met

## Debugging Tests

### Run Single Test

```sh
go test -run TestTaskRepository ./internal/repo
go test -run TestTaskRepository/Create ./internal/repo
```

### Race Detector

```sh
go test -race ./...
```

### Verbose with Stack Traces

```sh
go test -v -race ./internal/repo 2>&1 | grep -A 10 "FAIL"
```

## Best Practices

1. Write tests for all public APIs
2. Use builders for complex test data
3. Apply semantic assertion helpers
4. Keep tests focused and readable
5. Test both success and error paths
6. Avoid brittle time-based tests
7. Mock external dependencies
8. Use table-driven tests for variations
