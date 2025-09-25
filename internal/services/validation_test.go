package services

import (
	"errors"
	"strings"
	"testing"
)

type validationTC struct {
	name  string
	value string
	err   bool
}

func TestValidation(t *testing.T) {
	t.Run("ValidationError", func(t *testing.T) {
		err := ValidationError{Field: "testField", Message: "test message"}
		expected := "validation error for field 'testField': test message"
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("ValidationErrors", func(t *testing.T) {
		t.Run("empty errors", func(t *testing.T) {
			var errs ValidationErrors
			expected := "no validation errors"
			if errs.Error() != expected {
				t.Errorf("Expected %q, got %q", expected, errs.Error())
			}
		})

		t.Run("single error", func(t *testing.T) {
			errs := ValidationErrors{{Field: "field1", Message: "message1"}}
			expected := "validation error for field 'field1': message1"
			if errs.Error() != expected {
				t.Errorf("Expected %q, got %q", expected, errs.Error())
			}
		})

		t.Run("multiple errors", func(t *testing.T) {
			errs := ValidationErrors{{Field: "field1", Message: "message1"}, {Field: "field2", Message: "message2"}}
			result := errs.Error()
			if !strings.Contains(result, "multiple validation errors") {
				t.Error("Expected 'multiple validation errors' in result")
			}
			if !strings.Contains(result, "field1") || !strings.Contains(result, "field2") {
				t.Error("Expected both field names in result")
			}
		})
	})

	t.Run("RequiredString", func(t *testing.T) {
		tests := []validationTC{
			{"empty string", "", true},
			{"whitespace only", "   ", true},
			{"valid string", "test", false},
			{"string with spaces", "test value", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := RequiredString("testField", tt.value)
				if (err != nil) != tt.err {
					t.Errorf("Expected error: %v, got error: %v", tt.err, err != nil)
				}
				if err != nil {
					if !strings.Contains(err.Error(), "testField") {
						t.Error("Expected field name in error message")
					}
				}
			})
		}
	})

	t.Run("ValidURL", func(t *testing.T) {
		tests := []validationTC{
			{"empty string", "", false},
			{"valid http URL", "http://example.com", false},
			{"valid https URL", "https://example.com", false},
			{"invalid URL", "not-a-url", true},
			{"ftp scheme", "ftp://example.com", true},
			{"URL with path", "https://example.com/path", false},
			{"URL with query", "https://example.com?param=value", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := ValidURL("testField", tt.value)
				if (err != nil) != tt.err {
					t.Errorf("Expected error: %v, got error: %v", tt.err, err != nil)
				}
			})
		}
	})

	t.Run("ValidEmail", func(t *testing.T) {
		tests := []validationTC{
			{"empty string", "", false},
			{"valid email", "test@example.com", false},
			{"valid email with subdomain", "test@mail.example.com", false},
			{"invalid email no @", "testexample.com", true},
			{"invalid email no domain", "test@", true},
			{"invalid email no local part", "@example.com", true},
			{"invalid email spaces", "test @example.com", true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := ValidEmail("testField", tt.value)
				if (err != nil) != tt.err {
					t.Errorf("Expected error: %v, got error: %v", tt.err, err != nil)
				}
			})
		}
	})

	t.Run("StringLength", func(t *testing.T) {
		tests := []struct {
			name      string
			value     string
			min       int
			max       int
			shouldErr bool
		}{
			{"within range", "test", 2, 10, false},
			{"too short", "a", 2, 10, true},
			{"too long", "verylongstring", 2, 10, true},
			{"exact min", "ab", 2, 10, false},
			{"exact max", "1234567890", 2, 10, false},
			{"no min constraint", "a", 0, 10, false},
			{"no max constraint", "verylongstring", 2, 0, false},
			{"whitespace trimmed", "  test  ", 3, 10, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := StringLength("testField", tt.value, tt.min, tt.max)
				if (err != nil) != tt.shouldErr {
					t.Errorf("Expected error: %v, got error: %v", tt.shouldErr, err != nil)
				}
			})
		}
	})

	t.Run("ValidDate", func(t *testing.T) {
		tests := []validationTC{
			{"empty string", "", false},
			{"YYYY-MM-DD format", "2024-01-01", false},
			{"ISO format with time", "2024-01-01T15:04:05Z", false},
			{"ISO format with timezone", "2024-01-01T15:04:05-07:00", false},
			{"datetime format", "2024-01-01 15:04:05", false},
			{"invalid date", "not-a-date", true},
			{"invalid format", "01/01/2024", true},
			{"incomplete date", "2024-01", true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := ValidDate("testField", tt.value)
				if (err != nil) != tt.err {
					t.Errorf("Expected error: %v, got error: %v", tt.err, err != nil)
				}
			})
		}
	})

	t.Run("PositiveID", func(t *testing.T) {
		tests := []struct {
			name      string
			value     int64
			shouldErr bool
		}{
			{"positive ID", 1, false},
			{"zero ID", 0, true},
			{"negative ID", -1, true},
			{"large positive ID", 999999, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := PositiveID("testField", tt.value)
				if (err != nil) != tt.shouldErr {
					t.Errorf("Expected error: %v, got error: %v", tt.shouldErr, err != nil)
				}
			})
		}
	})

	t.Run("ValidEnum", func(t *testing.T) {
		allowed := []string{"option1", "option2", "option3"}

		tests := []validationTC{
			{"empty string", "", false},
			{"valid option1", "option1", false},
			{"valid option2", "option2", false},
			{"valid option3", "option3", false},
			{"invalid option", "option4", true},
			{"case sensitive", "Option1", true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := ValidEnum("testField", tt.value, allowed)
				if (err != nil) != tt.err {
					t.Errorf("Expected error: %v, got error: %v", tt.err, err != nil)
				}
			})
		}
	})

	t.Run("ValidFilePath", func(t *testing.T) {
		tests := []validationTC{
			{"empty string", "", false},
			{"valid path", "/path/to/file.txt", false},
			{"relative path", "path/to/file.txt", false},
			{"path traversal", "../../../etc/passwd", true},
			{"path with .. in middle", "/path/../to/file.txt", true},
			{"invalid characters", "/path/to/file<>.txt", true},
			{"pipe character", "/path/to/file|.txt", true},
			{"question mark", "/path/to/file?.txt", true},
			{"asterisk", "/path/to/file*.txt", true},
			{"colon", "/path/to/file:.txt", true},
			{"quotes", "/path/to/\"file\".txt", true},
			{"windows path", "C:\\path\\to\\file.txt", true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := ValidFilePath("testField", tt.value)
				if (err != nil) != tt.err {
					t.Errorf("Expected error: %v, got error: %v", tt.err, err != nil)
				}
			})
		}
	})

	t.Run("Validator", func(t *testing.T) {
		t.Run("empty validator", func(t *testing.T) {
			v := NewValidator()
			if !v.IsValid() {
				t.Error("Expected new validator to be valid")
			}
			if v.Errors() != nil {
				t.Error("Expected new validator to have no errors")
			}
		})

		t.Run("single validation error", func(t *testing.T) {
			v := NewValidator()
			v.Check(RequiredString("testField", ""))

			if v.IsValid() {
				t.Error("Expected validator to be invalid after failed check")
			}
			if v.Errors() == nil {
				t.Error("Expected validator to have errors")
			}
		})

		t.Run("multiple validation errors", func(t *testing.T) {
			v := NewValidator()
			v.Check(RequiredString("field1", ""))
			v.Check(RequiredString("field2", ""))

			if v.IsValid() {
				t.Error("Expected validator to be invalid")
			}
			err := v.Errors()
			if err == nil {
				t.Error("Expected validator to have errors")
			}
			if !strings.Contains(err.Error(), "field1") || !strings.Contains(err.Error(), "field2") {
				t.Error("Expected both field names in error message")
			}
		})

		t.Run("mixed valid and invalid checks", func(t *testing.T) {
			v := NewValidator()
			v.Check(RequiredString("validField", "valid"))
			v.Check(RequiredString("invalidField", ""))

			if v.IsValid() {
				t.Error("Expected validator to be invalid")
			}
			err := v.Errors()
			if err == nil {
				t.Error("Expected validator to have errors")
			}
			if !strings.Contains(err.Error(), "invalidField") {
				t.Error("Expected invalid field name in error message")
			}
		})

		t.Run("fluent interface", func(t *testing.T) {
			v := NewValidator()
			result := v.Check(RequiredString("field1", "valid")).Check(RequiredString("field2", "valid"))

			if result != v {
				t.Error("Expected Check to return the same validator instance")
			}
			if !v.IsValid() {
				t.Error("Expected validator to be valid after valid checks")
			}
		})

		t.Run("non-validation error handling", func(t *testing.T) {
			v := NewValidator()
			v.Check(errors.New("generic error"))

			if v.IsValid() {
				t.Error("Expected validator to be invalid")
			}
			err := v.Errors()
			if err == nil {
				t.Error("Expected validator to have errors")
			}

			if !strings.Contains(err.Error(), "unknown") {
				t.Error("Expected 'unknown' field in converted error")
			}
		})
	})
}
