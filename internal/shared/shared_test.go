package shared

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

func TestCompoundWriter(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		t.Run("creates writer with primary and secondary", func(t *testing.T) {
			var primary bytes.Buffer
			var secondary bytes.Buffer

			cw := New(&primary, &secondary)

			if cw == nil {
				t.Fatal("Expected CompoundWriter to be created")
			}
			if cw.primary == nil {
				t.Error("Expected primary writer to be set")
			}
			if cw.secondary == nil {
				t.Error("Expected secondary writer to be set")
			}
		})
	})

	t.Run("Write", func(t *testing.T) {
		t.Run("writes to both primary and secondary", func(t *testing.T) {
			var primary bytes.Buffer
			var secondary bytes.Buffer

			cw := New(&primary, &secondary)

			testData := []byte("test message")
			n, err := cw.Write(testData)

			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			if n != len(testData) {
				t.Errorf("Expected to write %d bytes, got %d", len(testData), n)
			}

			if primary.String() != "test message" {
				t.Errorf("Expected primary to contain 'test message', got '%s'", primary.String())
			}
			if secondary.String() != "test message" {
				t.Errorf("Expected secondary to contain 'test message', got '%s'", secondary.String())
			}
		})

		t.Run("writes multiple times to both sinks", func(t *testing.T) {
			var primary bytes.Buffer
			var secondary bytes.Buffer

			cw := New(&primary, &secondary)

			messages := []string{"first", "second", "third"}
			for _, msg := range messages {
				_, err := cw.Write([]byte(msg))
				if err != nil {
					t.Errorf("Expected no error writing '%s', got %v", msg, err)
				}
			}

			expected := "firstsecondthird"
			if primary.String() != expected {
				t.Errorf("Expected primary to contain '%s', got '%s'", expected, primary.String())
			}
			if secondary.String() != expected {
				t.Errorf("Expected secondary to contain '%s', got '%s'", expected, secondary.String())
			}
		})

		t.Run("returns error from primary writer", func(t *testing.T) {
			var secondary bytes.Buffer
			expectedErr := errors.New("primary write failed")
			primary := &errorWriter{err: expectedErr}

			cw := New(primary, &secondary)

			_, err := cw.Write([]byte("test"))

			if err == nil {
				t.Error("Expected error from primary writer")
			}
			if !errors.Is(err, expectedErr) {
				t.Errorf("Expected error '%v', got '%v'", expectedErr, err)
			}
		})

		t.Run("returns error from secondary writer", func(t *testing.T) {
			var primary bytes.Buffer
			expectedErr := errors.New("secondary write failed")
			secondary := &errorWriter{err: expectedErr}

			cw := New(&primary, secondary)

			_, err := cw.Write([]byte("test"))

			if err == nil {
				t.Error("Expected error from secondary writer")
			}
			if !errors.Is(err, expectedErr) {
				t.Errorf("Expected error '%v', got '%v'", expectedErr, err)
			}
		})

		t.Run("writes to primary even if secondary fails", func(t *testing.T) {
			var primary bytes.Buffer
			expectedErr := errors.New("secondary write failed")
			secondary := &errorWriter{err: expectedErr}

			cw := New(&primary, secondary)

			testData := []byte("test message")
			_, _ = cw.Write(testData)

			if primary.String() != "test message" {
				t.Errorf("Expected primary to contain 'test message' even with secondary error, got '%s'", primary.String())
			}
		})

		t.Run("handles empty write", func(t *testing.T) {
			var primary bytes.Buffer
			var secondary bytes.Buffer

			cw := New(&primary, &secondary)

			n, err := cw.Write([]byte{})

			if err != nil {
				t.Errorf("Expected no error on empty write, got %v", err)
			}
			if n != 0 {
				t.Errorf("Expected to write 0 bytes, got %d", n)
			}
		})

		t.Run("handles large write", func(t *testing.T) {
			var primary bytes.Buffer
			var secondary bytes.Buffer

			cw := New(&primary, &secondary)

			largeData := make([]byte, 1024*1024) // 1MB
			for i := range largeData {
				largeData[i] = byte(i % 256)
			}

			n, err := cw.Write(largeData)

			if err != nil {
				t.Errorf("Expected no error on large write, got %v", err)
			}
			if n != len(largeData) {
				t.Errorf("Expected to write %d bytes, got %d", len(largeData), n)
			}

			if !bytes.Equal(primary.Bytes(), largeData) {
				t.Error("Primary writer didn't receive correct data")
			}
			if !bytes.Equal(secondary.Bytes(), largeData) {
				t.Error("Secondary writer didn't receive correct data")
			}
		})
	})

	t.Run("WithStdErr", func(t *testing.T) {
		t.Run("creates writer with stderr as primary", func(t *testing.T) {
			var buf bytes.Buffer
			closer := &nopCloser{Writer: &buf}

			cw := LogWithStdErr(closer)

			if cw == nil {
				t.Fatal("Expected CompoundWriter to be created")
			}

			testData := []byte("test")
			_, err := cw.Write(testData)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if buf.String() != "test" {
				t.Errorf("Expected secondary to contain 'test', got '%s'", buf.String())
			}
		})
	})

	t.Run("WithStdOut", func(t *testing.T) {
		t.Run("creates writer with stdout as primary", func(t *testing.T) {
			var buf bytes.Buffer
			closer := &nopCloser{Writer: &buf}

			cw := LogWithStdOut(closer)

			if cw == nil {
				t.Fatal("Expected CompoundWriter to be created")
			}

			testData := []byte("test")
			_, err := cw.Write(testData)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if buf.String() != "test" {
				t.Errorf("Expected secondary to contain 'test', got '%s'", buf.String())
			}
		})
	})
}

func TestConfigError(t *testing.T) {
	t.Run("wraps error with message", func(t *testing.T) {
		originalErr := errors.New("original error")
		configErr := ConfigError("test message", originalErr)

		if configErr == nil {
			t.Fatal("Expected error to be returned")
		}

		errMsg := configErr.Error()
		if errMsg != "configuration error\ntest message: original error" {
			t.Errorf("Expected specific error format, got '%s'", errMsg)
		}
	})

	t.Run("is detectable with IsConfigError", func(t *testing.T) {
		originalErr := errors.New("original error")
		configErr := ConfigError("test message", originalErr)

		if !IsConfigError(configErr) {
			t.Error("Expected IsConfigError to return true")
		}
	})

	t.Run("wraps original error", func(t *testing.T) {
		originalErr := errors.New("original error")
		configErr := ConfigError("test message", originalErr)

		if !errors.Is(configErr, originalErr) {
			t.Error("Expected config error to wrap original error")
		}
	})
}

func TestIsConfigError(t *testing.T) {
	t.Run("returns true for config error", func(t *testing.T) {
		configErr := ConfigError("test", errors.New("inner"))

		if !IsConfigError(configErr) {
			t.Error("Expected IsConfigError to return true for config error")
		}
	})

	t.Run("returns false for non-config error", func(t *testing.T) {
		regularErr := errors.New("regular error")

		if IsConfigError(regularErr) {
			t.Error("Expected IsConfigError to return false for regular error")
		}
	})

	t.Run("returns false for nil error", func(t *testing.T) {
		if IsConfigError(nil) {
			t.Error("Expected IsConfigError to return false for nil error")
		}
	})
}

type errorWriter struct {
	err error
}

func (w *errorWriter) Write(p []byte) (int, error) {
	return 0, w.err
}

type nopCloser struct {
	io.Writer
}

func (nc *nopCloser) Close() error {
	return nil
}
