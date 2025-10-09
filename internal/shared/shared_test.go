package shared

import (
	"errors"
	"testing"
)

func TestErrors(t *testing.T) {
	t.Run("ConfigError", func(t *testing.T) {
		t.Run("creates joined error with message", func(t *testing.T) {
			baseErr := errors.New("invalid format")
			err := ConfigError("database connection failed", baseErr)

			AssertError(t, err, "ConfigError should create an error")
			AssertContains(t, err.Error(), "configuration error", "error should contain config error marker")
			AssertContains(t, err.Error(), "database connection failed", "error should contain custom message")
			AssertContains(t, err.Error(), "invalid format", "error should contain base error")
		})

		t.Run("preserves both error chains", func(t *testing.T) {
			baseErr := errors.New("connection timeout")
			err := ConfigError("failed to connect", baseErr)

			AssertTrue(t, errors.Is(err, ErrConfig), "should identify as config error")
			AssertTrue(t, errors.Is(err, baseErr), "should preserve original error in chain")
		})

		t.Run("wraps multiple errors with Join", func(t *testing.T) {
			baseErr := errors.New("parse error")
			err := ConfigError("invalid config file", baseErr)

			AssertTrue(t, errors.Is(err, ErrConfig), "joined error should contain ErrConfig")
			AssertTrue(t, errors.Is(err, baseErr), "joined error should contain base error")
		})
	})

	t.Run("IsConfigError", func(t *testing.T) {
		t.Run("identifies config errors", func(t *testing.T) {
			baseErr := errors.New("test error")
			err := ConfigError("test message", baseErr)

			AssertTrue(t, IsConfigError(err), "should identify config error")
		})

		t.Run("returns false for regular errors", func(t *testing.T) {
			err := errors.New("regular error")

			AssertFalse(t, IsConfigError(err), "should not identify regular error as config error")
		})

		t.Run("returns false for nil error", func(t *testing.T) {
			AssertFalse(t, IsConfigError(nil), "should return false for nil error")
		})

		t.Run("returns false for wrapped non-config errors", func(t *testing.T) {
			baseErr := errors.New("base error")
			wrappedErr := errors.New("wrapped: " + baseErr.Error())

			AssertFalse(t, IsConfigError(wrappedErr), "should not identify wrapped non-config error")
		})

		t.Run("identifies wrapped config errors", func(t *testing.T) {
			baseErr := errors.New("original error")
			configErr := ConfigError("config issue", baseErr)
			wrappedAgain := errors.Join(errors.New("outer error"), configErr)

			AssertTrue(t, IsConfigError(wrappedAgain), "should identify config error in join chain")
		})
	})
}
