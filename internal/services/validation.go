package services

import (
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"
)

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	dateFormats = []string{
		"2006-01-02",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02 15:04:05",
	}
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return "no validation errors"
	}

	if len(e) == 1 {
		return e[0].Error()
	}

	var messages []string
	for _, err := range e {
		messages = append(messages, err.Error())
	}
	return fmt.Sprintf("multiple validation errors: %s", strings.Join(messages, "; "))
}

// RequiredString validates that a string field is not empty
func RequiredString(name, value string) error {
	if strings.TrimSpace(value) == "" {
		return ValidationError{Field: name, Message: "is required and cannot be empty"}
	}
	return nil
}

// ValidURL validates that a string is a valid URL
func ValidURL(name, value string) error {
	if value == "" {
		return nil
	}

	parsed, err := url.Parse(value)
	if err != nil {
		return ValidationError{Field: name, Message: "must be a valid URL"}
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ValidationError{Field: name, Message: "must use http or https scheme"}
	}

	return nil
}

// ValidEmail validates that a string is a valid email address
func ValidEmail(name, value string) error {
	if value == "" {
		return nil
	}

	if !emailRegex.MatchString(value) {
		return ValidationError{Field: name, Message: "must be a valid email address"}
	}

	return nil
}

// StringLength validates string length constraints
func StringLength(name, value string, min, max int) error {
	length := len(strings.TrimSpace(value))

	if min > 0 && length < min {
		return ValidationError{Field: name, Message: fmt.Sprintf("must be at least %d characters long", min)}
	}

	if max > 0 && length > max {
		return ValidationError{Field: name, Message: fmt.Sprintf("must not exceed %d characters", max)}
	}

	return nil
}

// ValidDate validates that a string can be parsed as a date in supported formats
func ValidDate(name, value string) error {
	if value == "" {
		return nil
	}

	for _, format := range dateFormats {
		if _, err := time.Parse(format, value); err == nil {
			return nil
		}
	}

	return ValidationError{
		Field:   name,
		Message: "must be a valid date (YYYY-MM-DD, YYYY-MM-DDTHH:MM:SSZ, etc.)",
	}
}

// PositiveID validates that an ID is positive
func PositiveID(name string, value int64) error {
	if value <= 0 {
		return ValidationError{Field: name, Message: "must be a positive integer"}
	}
	return nil
}

// ValidEnum validates that a value is one of the allowed enum values
func ValidEnum(name, value string, allowedValues []string) error {
	if value == "" {
		return nil
	}

	if slices.Contains(allowedValues, value) {
		return nil
	}

	message := fmt.Sprintf("must be one of: %s", strings.Join(allowedValues, ", "))
	return ValidationError{Field: name, Message: message}
}

// ValidFilePath validates that a string looks like a valid file path
func ValidFilePath(name, value string) error {
	if value == "" {
		return nil
	}

	if strings.Contains(value, "..") {
		return ValidationError{Field: name, Message: "cannot contain '..' path traversal"}
	}

	if strings.ContainsAny(value, "<>:\"|?*") {
		return ValidationError{Field: name, Message: "contains invalid characters"}
	}

	return nil
}

// Validator provides a fluent interface for validation
type Validator struct {
	errors ValidationErrors
}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	return &Validator{}
}

// Check adds a validation check
func (v *Validator) Check(err error) *Validator {
	if err != nil {
		if valErr, ok := err.(ValidationError); ok {
			v.errors = append(v.errors, valErr)
		} else {
			v.errors = append(v.errors, ValidationError{Field: "unknown", Message: err.Error()})
		}
	}
	return v
}

// IsValid returns true if no validation errors occurred
func (v *Validator) IsValid() bool {
	return len(v.errors) == 0
}

// Errors returns all validation errors
func (v *Validator) Errors() error {
	if len(v.errors) == 0 {
		return nil
	}
	return v.errors
}
