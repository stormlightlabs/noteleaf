package utils

import (
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Logger is the global application logger
var Logger *log.Logger

// NewLogger creates a new logger with the specified level and format
func NewLogger(level string, format string) *log.Logger {
	logger := log.New(os.Stderr)

	switch strings.ToLower(level) {
	case "debug":
		logger.SetLevel(log.DebugLevel)
	case "info":
		logger.SetLevel(log.InfoLevel)
	case "warn", "warning":
		logger.SetLevel(log.WarnLevel)
	case "error":
		logger.SetLevel(log.ErrorLevel)
	default:
		logger.SetLevel(log.InfoLevel)
	}

	// Set format
	if format == "json" {
		logger.SetFormatter(log.JSONFormatter)
	} else {
		logger.SetFormatter(log.TextFormatter)
		logger.SetReportTimestamp(true)
	}

	return logger
}

// GetLogger returns the global logger, creating a default one if it doesn't exist
func GetLogger() *log.Logger {
	if Logger == nil {
		Logger = NewLogger("info", "text")
	}
	return Logger
}

func Titlecase(s string) string {
	return cases.Title(language.Und, cases.NoLower).String(s)
}
