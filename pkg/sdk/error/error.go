// Package error provides structured error handling for wave plugins.
// Plugins use this package to emit errors that wave-core can parse and display.
//
// Error Format:
// By default, plugins emit errors as human-readable plain text to stderr.
// When WAVE_DEBUG=1, errors are emitted as JSON for debugging.
// The error code should be lowercase-and-dashes format (e.g. "flow-no-command").
package error

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// WaveError represents a structured error from a wave plugin.
// This is used internally but errors are emitted as plain text by default.
// When WAVE_DEBUG=1, errors are emitted as JSON.
type WaveError struct {
	WaveError bool   `json:"wave_error"`        // Always true for wave errors
	Code      string `json:"code"`              // Error code in lowercase-and-dashes format
	Message   string `json:"message"`           // Human-readable error message
	Details   string `json:"details,omitempty"` // Additional context (optional)
}

// isDebugMode checks if WAVE_DEBUG environment variable is set to "1".
func isDebugMode() bool {
	return os.Getenv("WAVE_DEBUG") == "1"
}

// Error implements the error interface.
func (e *WaveError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s\n%s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Emit writes an error to stderr and exits with code 1.
// Code MUST be lowercase-and-dashes format (e.g. "flow-no-command").
// Message should be a human-readable description of the error.
func Emit(code, message string) {
	Format(os.Stderr, code, message, "")
	os.Exit(1)
}

// EmitWithDetails writes an error with additional details and exits.
// Code MUST be lowercase-and-dashes format (e.g. "flow-resolve-error").
// Details provides additional context (suggestions, available commands, etc.).
func EmitWithDetails(code, message, details string) {
	Format(os.Stderr, code, message, details)
	os.Exit(1)
}

// Format writes an error to the given writer without exiting.
// This is the testable version of Emit().
// When WAVE_DEBUG=1, outputs JSON format. Otherwise outputs plain text.
func Format(w io.Writer, code, message, details string) {
	if isDebugMode() {
		formatJSON(w, code, message, details)
	} else {
		formatPlain(w, code, message, details)
	}
}

// formatPlain writes error in plain text format: "code: message\ndetails"
func formatPlain(w io.Writer, code, message, details string) {
	if details != "" {
		fmt.Fprintf(w, "%s: %s\n%s\n", code, message, details)
	} else {
		fmt.Fprintf(w, "%s: %s\n", code, message)
	}
}

// formatJSON writes error in JSON format for debugging.
func formatJSON(w io.Writer, code, message, details string) {
	e := WaveError{
		WaveError: true,
		Code:      code,
		Message:   message,
		Details:   details,
	}
	json.NewEncoder(w).Encode(e)
}

// New creates a new WaveError without emitting it.
// Useful when you need to return an error rather than immediately exit.
func New(code, message string) *WaveError {
	return &WaveError{
		Code:    code,
		Message: message,
	}
}

// NewWithDetails creates a new WaveError with additional details.
func NewWithDetails(code, message, details string) *WaveError {
	return &WaveError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// Errf creates a new error with a formatted message.
// This is a convenience wrapper around fmt.Errorf for plugin authors.
// It supports all fmt.Errorf formatting verbs including %w for wrapping errors.
func Errf(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}
