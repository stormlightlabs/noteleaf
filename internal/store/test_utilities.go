package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"
	"time"
)

// Error Testing Utilities
//
// This file provides reusable error testing helpers for systematically testing
// error paths in the store package. The utilities are organized into categories:
//
// 1. ErrorInjector - File system and permission errors
// 2. DBErrorSimulator - Database-specific errors
// 3. ContextHelper - Context cancellation and timeout scenarios
// 4. Platform-specific testing utilities
//
// Usage patterns are documented for each helper with examples from existing tests.

// ErrorInjector provides utilities for simulating file system and permission errors
type ErrorInjector struct {
	t               *testing.T
	tempDirs        []string
	originalPerms   map[string]os.FileMode
	cleanupFuncs    []func()
	envVarsModified map[string]string
}

// NewErrorInjector creates a new error injector for testing
func NewErrorInjector(t *testing.T) *ErrorInjector {
	t.Helper()
	return &ErrorInjector{
		t:               t,
		tempDirs:        make([]string, 0),
		originalPerms:   make(map[string]os.FileMode),
		cleanupFuncs:    make([]func(), 0),
		envVarsModified: make(map[string]string),
	}
}

// CreateUnwritableDir creates a directory with read-only permissions
//
// Use this to test errors when the application cannot write to config or data directories.
// Example from config_test.go:
//
//	injector := NewErrorInjector(t)
//	dir := injector.CreateUnwritableDir()
//	defer injector.Cleanup()
//	// Test code that should fail when dir is read-only
func (ei *ErrorInjector) CreateUnwritableDir() string {
	ei.t.Helper()

	if runtime.GOOS == "windows" {
		ei.t.Skip("Permission test not reliable on Windows")
	}

	tempDir, err := os.MkdirTemp("", "noteleaf-unwritable-*")
	if err != nil {
		ei.t.Fatalf("Failed to create temp directory: %v", err)
	}

	ei.tempDirs = append(ei.tempDirs, tempDir)

	if err := os.Chmod(tempDir, 0555); err != nil {
		ei.t.Fatalf("Failed to make directory read-only: %v", err)
	}

	ei.originalPerms[tempDir] = 0755
	return tempDir
}

// CreateUnreadableFile creates a file with no read permissions
//
// Use this to test errors when the application cannot read config files.
// Example pattern:
//
//	injector := NewErrorInjector(t)
//	filePath := injector.CreateUnreadableFile("config.toml", "content")
//	defer injector.Cleanup()
//	// Test code that should fail when file is unreadable
func (ei *ErrorInjector) CreateUnreadableFile(filename, content string) string {
	ei.t.Helper()

	if runtime.GOOS == "windows" {
		ei.t.Skip("Permission test not reliable on Windows")
	}

	tempDir, err := os.MkdirTemp("", "noteleaf-unreadable-*")
	if err != nil {
		ei.t.Fatalf("Failed to create temp directory: %v", err)
	}

	ei.tempDirs = append(ei.tempDirs, tempDir)

	filePath := filepath.Join(tempDir, filename)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		ei.t.Fatalf("Failed to create file: %v", err)
	}

	if err := os.Chmod(filePath, 0000); err != nil {
		ei.t.Fatalf("Failed to make file unreadable: %v", err)
	}

	ei.originalPerms[filePath] = 0644
	return filePath
}

// MockFuncError injects a custom error into a variable function
//
// Use this to test error handling when system functions fail.
// Example from database_test.go:
//
//	injector := NewErrorInjector(t)
//	injector.MockFuncError("GetConfigDir", func() (string, error) {
//	    return "", os.ErrPermission
//	})
//	defer injector.Cleanup()
func (ei *ErrorInjector) MockFuncError(name string, cleanup func()) {
	ei.t.Helper()
	ei.cleanupFuncs = append(ei.cleanupFuncs, cleanup)
}

// SetEnv sets an environment variable and tracks it for cleanup
//
// Use this to test environment variable handling and overrides.
// Example from config_test.go:
//
//	injector := NewErrorInjector(t)
//	injector.SetEnv("NOTELEAF_CONFIG", "/custom/path")
//	defer injector.Cleanup()
func (ei *ErrorInjector) SetEnv(key, value string) {
	ei.t.Helper()

	if _, exists := ei.envVarsModified[key]; !exists {
		ei.envVarsModified[key] = os.Getenv(key)
	}

	os.Setenv(key, value)
}

// UnsetEnv unsets an environment variable and tracks it for restoration
func (ei *ErrorInjector) UnsetEnv(key string) {
	ei.t.Helper()

	if _, exists := ei.envVarsModified[key]; !exists {
		ei.envVarsModified[key] = os.Getenv(key)
	}

	os.Unsetenv(key)
}

// Cleanup restores all modified permissions, removes temp directories, and restores environment
func (ei *ErrorInjector) Cleanup() {
	for path, perm := range ei.originalPerms {
		os.Chmod(path, perm)
	}

	for _, dir := range ei.tempDirs {
		os.RemoveAll(dir)
	}

	for _, cleanup := range ei.cleanupFuncs {
		cleanup()
	}

	for key, value := range ei.envVarsModified {
		if value == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, value)
		}
	}
}

// DBErrorSimulator provides utilities for testing database error scenarios
type DBErrorSimulator struct {
	t  *testing.T
	db *sql.DB
}

// NewDBErrorSimulator creates a database error simulator
func NewDBErrorSimulator(t *testing.T, db *sql.DB) *DBErrorSimulator {
	t.Helper()
	return &DBErrorSimulator{t: t, db: db}
}

// CorruptTable drops a table to simulate database corruption
//
// Use this to test error handling when database schema is corrupted.
// Example from handlers/test_utilities.go:
//
//	sim := NewDBErrorSimulator(t, db.DB)
//	sim.CorruptTable("tasks")
//	// Test code that should handle missing table gracefully
func (sim *DBErrorSimulator) CorruptTable(tableName string) {
	sim.t.Helper()
	sim.db.Exec(fmt.Sprintf("DROP TABLE %s", tableName))
}

// SimulateConnectionFailure closes the database connection
//
// Use this to test error handling when database connection is lost.
// Example pattern:
//
//	sim := NewDBErrorSimulator(t, db.DB)
//	sim.SimulateConnectionFailure()
//	// Test code that should handle connection errors
func (sim *DBErrorSimulator) SimulateConnectionFailure() {
	sim.t.Helper()
	sim.db.Close()
}

// ConstraintViolationHelper provides utilities for testing constraint violations
type ConstraintViolationHelper struct {
	t   *testing.T
	db  *sql.DB
	ctx context.Context
}

// NewConstraintViolationHelper creates a constraint violation helper
func NewConstraintViolationHelper(t *testing.T, db *sql.DB) *ConstraintViolationHelper {
	t.Helper()
	return &ConstraintViolationHelper{
		t:   t,
		db:  db,
		ctx: context.Background(),
	}
}

// TestUniqueConstraint tests unique constraint violations
//
// Use this to verify error handling when inserting duplicate unique values.
// Pattern:
//
//	helper := NewConstraintViolationHelper(t, db)
//	helper.TestUniqueConstraint("tasks", "uuid", "duplicate-uuid", func() error {
//	    return repo.Create(ctx, task)
//	})
func (cvh *ConstraintViolationHelper) TestUniqueConstraint(tableName, columnName, duplicateValue string, operation func() error) {
	cvh.t.Helper()

	// First operation should succeed
	err := operation()
	if err != nil {
		cvh.t.Fatalf("First operation should succeed: %v", err)
	}

	// Second operation should fail with constraint violation
	err = operation()
	if err == nil {
		cvh.t.Errorf("Expected unique constraint violation on %s.%s", tableName, columnName)
	}
}

// TestForeignKeyConstraint tests foreign key constraint violations
//
// Use this to verify error handling when referencing non-existent foreign keys.
// Pattern:
//
//	helper := NewConstraintViolationHelper(t, db)
//	helper.TestForeignKeyConstraint("time_entries", "task_id", "nonexistent-task-id", func() error {
//	    return repo.CreateTimeEntry(ctx, entry)
//	})
func (cvh *ConstraintViolationHelper) TestForeignKeyConstraint(tableName, columnName, invalidFK string, operation func() error) {
	cvh.t.Helper()

	err := operation()
	if err == nil {
		cvh.t.Errorf("Expected foreign key constraint violation on %s.%s", tableName, columnName)
	}
}

// ContextHelper provides utilities for testing context cancellation and timeouts
type ContextHelper struct {
	t *testing.T
}

// NewContextHelper creates a context helper
func NewContextHelper(t *testing.T) *ContextHelper {
	t.Helper()
	return &ContextHelper{t: t}
}

// CancelledContext returns a pre-cancelled context
//
// Use this to test error handling when operations are called with cancelled context.
// Example from task_repository_test.go:
//
//	helper := NewContextHelper(t)
//	ctx := helper.CancelledContext()
//	_, err := repo.Create(ctx, task)
//	if err == nil {
//	    t.Error("Expected error with cancelled context")
//	}
func (ch *ContextHelper) CancelledContext() context.Context {
	ch.t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

// TimeoutContext returns a context that times out after the specified duration
//
// Use this to test timeout handling in long-running operations.
// Pattern:
//
//	helper := NewContextHelper(t)
//	ctx := helper.TimeoutContext(1 * time.Millisecond)
//	// Slow operation that should timeout
func (ch *ContextHelper) TimeoutContext(timeout time.Duration) context.Context {
	ch.t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	ch.t.Cleanup(cancel)
	return ctx
}

// DeadlineExceededContext returns a context with an already-expired deadline
//
// Use this to test error handling when context deadline is already exceeded.
func (ch *ContextHelper) DeadlineExceededContext() context.Context {
	ch.t.Helper()
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-1*time.Second))
	ch.t.Cleanup(cancel)
	return ctx
}

// PlatformSpecific provides utilities for platform-specific testing
type PlatformSpecific struct {
	t *testing.T
}

// NewPlatformSpecific creates platform-specific test helper
func NewPlatformSpecific(t *testing.T) *PlatformSpecific {
	t.Helper()
	return &PlatformSpecific{t: t}
}

// SkipOnWindows skips the test if running on Windows
//
// Use this for tests that rely on Unix-specific features like chmod.
func (ps *PlatformSpecific) SkipOnWindows(reason string) {
	ps.t.Helper()
	if runtime.GOOS == "windows" {
		ps.t.Skip(reason)
	}
}

// SkipOnMac skips the test if running on macOS
//
// Use this for tests that have platform-specific behavior on macOS.
func (ps *PlatformSpecific) SkipOnMac(reason string) {
	ps.t.Helper()
	if runtime.GOOS == "darwin" {
		ps.t.Skip(reason)
	}
}

// RunOnlyOn runs the test only on specified platforms
//
// Use this to create platform-specific test cases.
func (ps *PlatformSpecific) RunOnlyOn(platforms []string) {
	ps.t.Helper()
	if slices.Contains(platforms, runtime.GOOS) {
		return
	}
	ps.t.Skipf("Test only runs on %v, skipping on %s", platforms, runtime.GOOS)
}

// IsolatedEnvironment creates an isolated test environment with custom dirs
type IsolatedEnvironment struct {
	t                    *testing.T
	TempDir              string
	ConfigDir            string
	DataDir              string
	originalGetConfigDir func() (string, error)
	originalGetDataDir   func() (string, error)
}

// NewIsolatedEnvironment creates an isolated test environment
//
// Use this to test in complete isolation from the actual system config/data directories.
// Example from database_test.go:
//
//	env := NewIsolatedEnvironment(t)
//	defer env.Cleanup()
//	// All operations will use env.TempDir
func NewIsolatedEnvironment(t *testing.T) *IsolatedEnvironment {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "noteleaf-isolated-*")
	if err != nil {
		t.Fatalf("Failed to create isolated temp directory: %v", err)
	}

	env := &IsolatedEnvironment{
		t:                    t,
		TempDir:              tempDir,
		ConfigDir:            tempDir,
		DataDir:              tempDir,
		originalGetConfigDir: GetConfigDir,
		originalGetDataDir:   GetDataDir,
	}

	GetConfigDir = func() (string, error) {
		return env.ConfigDir, nil
	}

	GetDataDir = func() (string, error) {
		return env.DataDir, nil
	}

	t.Cleanup(func() {
		env.Cleanup()
	})

	return env
}

// SetConfigDirError configures GetConfigDir to return an error
func (env *IsolatedEnvironment) SetConfigDirError(err error) {
	env.t.Helper()
	GetConfigDir = func() (string, error) {
		return "", err
	}
}

// SetDataDirError configures GetDataDir to return an error
func (env *IsolatedEnvironment) SetDataDirError(err error) {
	env.t.Helper()
	GetDataDir = func() (string, error) {
		return "", err
	}
}

// Cleanup restores original functions and removes temp directory
func (env *IsolatedEnvironment) Cleanup() {
	GetConfigDir = env.originalGetConfigDir
	GetDataDir = env.originalGetDataDir
	os.RemoveAll(env.TempDir)
}

// AssertError checks that an error occurred
//
// Use this for simple error existence checks.
func AssertError(t *testing.T, err error, msg string) {
	t.Helper()
	if err == nil {
		t.Errorf("%s: expected error but got none", msg)
	}
}

// AssertNoError checks that no error occurred
//
// Use this to verify successful operations.
func AssertNoError(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Errorf("%s: unexpected error: %v", msg, err)
	}
}

// AssertErrorContains checks that error contains expected substring
//
// Use this to verify specific error messages.
func AssertErrorContains(t *testing.T, err error, expectedSubstring, msg string) {
	t.Helper()
	if err == nil {
		t.Errorf("%s: expected error containing %q but got none", msg, expectedSubstring)
		return
	}
	if expectedSubstring != "" && !contains(err.Error(), expectedSubstring) {
		t.Errorf("%s: expected error containing %q, got: %v", msg, expectedSubstring, err)
	}
}

// AssertErrorIs checks that error matches target using [errors.Is]
//
// Use this to verify error types and wrapped errors.
func AssertErrorIs(t *testing.T, err, target error, msg string) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Errorf("%s: expected error to be %v, got: %v", msg, target, err)
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) &&
		(haystack[len(haystack)-len(needle):] == needle ||
			haystack[:len(needle)] == needle ||
			(len(haystack) > len(needle) && func() bool {
				for i := 1; i <= len(haystack)-len(needle); i++ {
					if haystack[i:i+len(needle)] == needle {
						return true
					}
				}
				return false
			}()))
}
