package utils

import (
	"bytes"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/charmbracelet/log"
)

func getConfigPath() string {
	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData != "" {
			return appData + "\\noteleaf"
		}
		return "C:\\noteleaf"
	default:
		configHome := os.Getenv("XDG_CONFIG_HOME")
		if configHome != "" {
			return configHome + "/noteleaf"
		}
		return os.Getenv("HOME") + "/.config/noteleaf"
	}
}

func simulateWindowsConfigPath() string {
	appData := os.Getenv("APPDATA")
	if appData != "" {
		return appData + "\\noteleaf"
	}
	return "C:\\noteleaf"
}

func getBookStatusDisplay(status string) string {
	switch status {
	case "reading":
		return "currently reading"
	case "finished":
		return "completed"
	case "queued":
		return "to be read"
	default:
		return status
	}
}

func classifyMediaType(link string) string {
	if strings.HasPrefix(link, "/m/") {
		return "movie"
	} else if strings.HasPrefix(link, "/tv/") {
		return "tv"
	}
	return "unknown"
}

func validatePriority(priority string) bool {
	switch strings.ToLower(priority) {
	case "high", "medium", "low":
		return true
	case "h", "m", "l":
		return true
	case "1", "2", "3", "4", "5":
		return true
	default:
		return false
	}
}

func TestLogger(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		t.Run("creates logger with info level", func(t *testing.T) {
			logger := NewLogger("info", "text")
			if logger == nil {
				t.Fatal("Logger should not be nil")
			}

			if logger.GetLevel() != log.InfoLevel {
				t.Errorf("Expected InfoLevel, got %v", logger.GetLevel())
			}
		})

		t.Run("creates logger with debug level", func(t *testing.T) {
			logger := NewLogger("debug", "text")
			if logger.GetLevel() != log.DebugLevel {
				t.Errorf("Expected DebugLevel, got %v", logger.GetLevel())
			}
		})

		t.Run("creates logger with warn level", func(t *testing.T) {
			logger := NewLogger("warn", "text")
			if logger.GetLevel() != log.WarnLevel {
				t.Errorf("Expected WarnLevel, got %v", logger.GetLevel())
			}
		})

		t.Run("creates logger with warning level alias", func(t *testing.T) {
			logger := NewLogger("warning", "text")
			if logger.GetLevel() != log.WarnLevel {
				t.Errorf("Expected WarnLevel, got %v", logger.GetLevel())
			}
		})

		t.Run("creates logger with error level", func(t *testing.T) {
			logger := NewLogger("error", "text")
			if logger.GetLevel() != log.ErrorLevel {
				t.Errorf("Expected ErrorLevel, got %v", logger.GetLevel())
			}
		})

		t.Run("defaults to info level for invalid level", func(t *testing.T) {
			logger := NewLogger("invalid", "text")
			if logger.GetLevel() != log.InfoLevel {
				t.Errorf("Expected InfoLevel for invalid input, got %v", logger.GetLevel())
			}
		})

		t.Run("handles case insensitive levels", func(t *testing.T) {
			logger := NewLogger("DEBUG", "text")
			if logger.GetLevel() != log.DebugLevel {
				t.Errorf("Expected DebugLevel for uppercase input, got %v", logger.GetLevel())
			}
		})

		t.Run("creates logger with json format", func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger("info", "json")
			logger.SetOutput(&buf)

			logger.Info("test message")
			output := buf.String()

			if !strings.Contains(output, "{") || !strings.Contains(output, "}") {
				t.Error("Expected JSON formatted output")
			}
		})

		t.Run("creates logger with text format", func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger("info", "text")
			logger.SetOutput(&buf)

			logger.Info("test message")
			output := buf.String()

			if strings.Contains(output, "{") && strings.Contains(output, "}") {
				t.Error("Expected text formatted output, not JSON")
			}
		})

		t.Run("text format includes timestamp", func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger("info", "text")
			logger.SetOutput(&buf)

			logger.Info("test message")
			output := buf.String()

			if !strings.Contains(output, ":") {
				t.Error("Expected timestamp in text format output")
			}
		})
	})

	t.Run("Get", func(t *testing.T) {
		t.Run("returns global logger when set", func(t *testing.T) {
			originalLogger := Logger
			defer func() { Logger = originalLogger }()

			testLogger := NewLogger("debug", "json")
			Logger = testLogger

			retrieved := GetLogger()
			if retrieved != testLogger {
				t.Error("GetLogger should return the global logger")
			}
		})

		t.Run("creates default logger when global is nil", func(t *testing.T) {
			originalLogger := Logger
			defer func() { Logger = originalLogger }()

			Logger = nil

			retrieved := GetLogger()
			if retrieved == nil {
				t.Fatal("GetLogger should create a default logger")
			}

			if retrieved.GetLevel() != log.InfoLevel {
				t.Error("Default logger should have InfoLevel")
			}

			if Logger != retrieved {
				t.Error("Global logger should be set after GetLogger call")
			}
		})

		t.Run("subsequent calls return same logger", func(t *testing.T) {
			originalLogger := Logger
			defer func() { Logger = originalLogger }()

			Logger = nil

			logger1 := GetLogger()
			logger2 := GetLogger()

			if logger1 != logger2 {
				t.Error("Subsequent GetLogger calls should return the same instance")
			}
		})
	})

	t.Run("Integration", func(t *testing.T) {
		t.Run("logger writes to stderr by default", func(t *testing.T) {
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			logger := NewLogger("info", "text")
			logger.Info("test message")

			w.Close()
			os.Stderr = oldStderr

			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			if !strings.Contains(output, "test message") {
				t.Error("Logger should write to stderr by default")
			}
		})

		t.Run("logger respects level filtering", func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger("error", "text")
			logger.SetOutput(&buf)

			logger.Debug("debug message")
			logger.Info("info message")
			logger.Warn("warn message")
			logger.Error("error message")

			output := buf.String()

			if strings.Contains(output, "debug message") {
				t.Error("Debug message should be filtered out at error level")
			}
			if strings.Contains(output, "info message") {
				t.Error("Info message should be filtered out at error level")
			}
			if strings.Contains(output, "warn message") {
				t.Error("Warn message should be filtered out at error level")
			}
			if !strings.Contains(output, "error message") {
				t.Error("Error message should be included at error level")
			}
		})

		t.Run("global logger persists between function calls", func(t *testing.T) {
			originalLogger := Logger
			defer func() { Logger = originalLogger }()

			Logger = NewLogger("debug", "json")

			retrieved := GetLogger()

			if retrieved.GetLevel() != log.DebugLevel {
				t.Error("Global logger settings should persist")
			}
		})
	})
}

func TestTitlecase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single word lowercase",
			input:    "hello",
			expected: "Hello",
		},
		{
			name:     "single word uppercase",
			input:    "HELLO",
			expected: "HELLO",
		},
		{
			name:     "multiple words",
			input:    "hello world",
			expected: "Hello World",
		},
		{
			name:     "mixed case",
			input:    "hELLo WoRLD",
			expected: "HELLo WoRLD",
		},
		{
			name:     "with punctuation",
			input:    "hello, world!",
			expected: "Hello, World!",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "with numbers",
			input:    "hello 123 world",
			expected: "Hello 123 World",
		},
		{
			name:     "with special characters",
			input:    "hello-world_test",
			expected: "Hello-World_test",
		},
		{
			name:     "already title case",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "single character",
			input:    "a",
			expected: "A",
		},
		{
			name:     "apostrophes",
			input:    "it's a beautiful day",
			expected: "It's A Beautiful Day",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Titlecase(tt.input)
			if result != tt.expected {
				t.Errorf("Titlecase(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestPlatformSpecificPaths(t *testing.T) {
	t.Run("Windows Path Handling", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Run("APPDATA Environment Variable", func(t *testing.T) {
				appData := os.Getenv("APPDATA")
				if appData == "" {
					t.Skip("APPDATA environment variable not set")
				}

				path := getConfigPath()
				if !strings.Contains(path, appData) {
					t.Errorf("Expected config path to contain APPDATA path %s, got %s", appData, path)
				}
			})
		} else {
			t.Run("Simulated Windows Path Handling", func(t *testing.T) {
				originalAppData := os.Getenv("APPDATA")
				defer os.Setenv("APPDATA", originalAppData)

				os.Setenv("APPDATA", "C:\\Users\\Test\\AppData\\Roaming")

				testPath := simulateWindowsConfigPath()
				expected := "C:\\Users\\Test\\AppData\\Roaming"
				if !strings.Contains(testPath, expected) {
					t.Errorf("Expected Windows config path to contain %s, got %s", expected, testPath)
				}
			})
		}
	})

	t.Run("Unix-like Path Handling", func(t *testing.T) {
		if runtime.GOOS != "windows" {
			t.Run("XDG Config Home", func(t *testing.T) {
				originalConfigHome := os.Getenv("XDG_CONFIG_HOME")
				defer os.Setenv("XDG_CONFIG_HOME", originalConfigHome)

				testConfigHome := "/tmp/test-config"
				os.Setenv("XDG_CONFIG_HOME", testConfigHome)

				path := getConfigPath()
				if !strings.Contains(path, testConfigHome) {
					t.Errorf("Expected config path to contain XDG_CONFIG_HOME %s, got %s", testConfigHome, path)
				}
			})
		}
	})
}

func TestStatusFieldMatching(t *testing.T) {
	t.Run("Book Status Field Access", func(t *testing.T) {
		tests := []struct {
			status   string
			expected string
		}{
			{"reading", "currently reading"},
			{"finished", "completed"},
			{"queued", "to be read"},
		}

		for _, tt := range tests {
			t.Run(tt.status, func(t *testing.T) {
				result := getBookStatusDisplay(tt.status)
				if !strings.Contains(result, tt.expected) {
					t.Errorf("Expected status display for %s to contain %s, got %s", tt.status, tt.expected, result)
				}
			})
		}
	})
}

func TestMediaTypeMatching(t *testing.T) {
	t.Run("Media Type Classification", func(t *testing.T) {
		tests := []struct {
			link         string
			expectedType string
		}{
			{"/m/some-movie", "movie"},
			{"/tv/some-show", "tv"},
			{"/other/link", "unknown"},
		}

		for _, tt := range tests {
			t.Run(tt.link, func(t *testing.T) {
				result := classifyMediaType(tt.link)
				if result != tt.expectedType {
					t.Errorf("Expected media type %s for link %s, got %s", tt.expectedType, tt.link, result)
				}
			})
		}
	})
}

func TestTaskPriorityValidation(t *testing.T) {
	t.Run("Priority String Validation", func(t *testing.T) {
		tests := []struct {
			priority string
			valid    bool
		}{
			{"high", true}, {"medium", true}, {"low", true},
			{"H", true}, {"M", true}, {"L", true},
			{"1", true}, {"5", true},
			{"invalid", false},
		}

		for _, tt := range tests {
			t.Run(tt.priority, func(t *testing.T) {
				isValid := validatePriority(tt.priority)
				if isValid != tt.valid {
					t.Errorf("Expected priority %s to be valid=%t, got %t", tt.priority, tt.valid, isValid)
				}
			})
		}
	})
}
