// package shared contains constants used across the codebase
package shared

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/log"
)

var (
	ErrConfig error = fmt.Errorf("configuration error")
)

func ConfigError(m string, err error) error {
	return errors.Join(ErrConfig, fmt.Errorf("%s: %w", m, err))
}

func IsConfigError(err error) bool {
	return errors.Is(err, ErrConfig)
}

// CompoundWriter writes every payload to two sinks:
//  1. a primary sink (typically [os.Stdout] or [os.Stderr])
//  2. a secondary sink (typically [*os.File])
//
// It satisfies io.Writer.
type CompoundWriter struct {
	primary   io.Writer
	secondary io.Writer
}

// New creates a new [CompoundWriter]
func New(primary io.Writer, secondary io.Writer) *CompoundWriter {
	return &CompoundWriter{
		primary:   primary,
		secondary: secondary,
	}
}

func LogWithStdErr(w io.WriteCloser) *CompoundWriter {
	return New(os.Stderr, w)
}

func LogWithStdOut(w io.WriteCloser) *CompoundWriter {
	return New(os.Stdout, w)
}

// Write writes p to both instances of [io.Writer]
func (cw *CompoundWriter) Write(p []byte) (int, error) {
	var err error
	var n1, n2 int

	if n1, err = cw.primary.Write(p); err != nil {
		return n1, err
	}
	if n2, err = cw.secondary.Write(p); err != nil {
		return n2, err
	}
	return len(p), nil
}

func FallbackLogger() *log.Logger {
	return log.NewWithOptions(os.Stderr, log.Options{
		Prefix:          "[DEBUG]",
		ReportTimestamp: true,
		ReportCaller:    true,
		TimeFormat:      time.Kitchen,
		Level:           log.DebugLevel,
	})
}

// NewDebugLoggerWithFile creates a new debug logger that writes to both stderr and a log file
func NewDebugLoggerWithFile(configDir string) *log.Logger {
	logger := FallbackLogger()
	logsDir := filepath.Join(configDir, "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return logger
	}

	logFile := filepath.Join(logsDir, fmt.Sprintf("publication_%s.log", time.Now().Format("2006-01-02")))
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return logger
	}

	w := LogWithStdErr(file)
	logger.SetOutput(w)
	return logger
}
