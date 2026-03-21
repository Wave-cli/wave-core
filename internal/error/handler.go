package errors

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// ANSI color codes
const (
	colorRed   = "\033[31m"
	colorReset = "\033[0m"
)

// isDebugMode checks if WAVE_DEBUG environment variable is set to "1".
func isDebugMode() bool {
	return os.Getenv("WAVE_DEBUG") == "1"
}

// FormatError renders a PluginError into a human-friendly string for terminal display.
// When WAVE_DEBUG=1, it prints the raw JSON to stderr.
// Otherwise, it shows a colored user-friendly message.
func FormatError(pluginName, pluginVersion string, pe *PluginError, logPath string) string {
	if isDebugMode() {
		return formatDebugError(pe)
	}
	return formatSimpleError(pe)
}

// formatSimpleError returns a colored, user-friendly error message.
// Format:
//   - With message: "error code: message\ndetails" (all in red)
//   - Without message: "error code\ndetails" (all in red)
//   - Details only shown if present
func formatSimpleError(pe *PluginError) string {
	var b strings.Builder

	// Build the error line
	if pe.Message != "" {
		fmt.Fprintf(&b, "%s%s: %s%s", colorRed, pe.Code, pe.Message, colorReset)
	} else {
		fmt.Fprintf(&b, "%s%s%s", colorRed, pe.Code, colorReset)
	}

	// Add details on a new line if present
	if pe.Details != "" {
		fmt.Fprintf(&b, "\n%s%s%s", colorRed, pe.Details, colorReset)
	}

	return b.String()
}

// formatDebugError returns the raw JSON error for debugging.
func formatDebugError(pe *PluginError) string {
	jsonBytes, _ := json.Marshal(pe)
	return string(jsonBytes)
}

// ParseStderr scans stderr output for a wave error.
// It supports two formats:
//  1. JSON with wave_error: true (legacy format)
//  2. Plain text: "code: message\ndetails" (preferred format)
//
// Returns nil if no structured error is found.
func ParseStderr(stderr []byte) *PluginError {
	if len(stderr) == 0 {
		return nil
	}

	lines := strings.Split(string(stderr), "\n")

	// Try JSON format first (search in reverse - error is typically last)
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if pe := tryParseWaveError(line); pe != nil {
			return pe
		}
	}

	// Try plain text format: "code: message\ndetails"
	return tryParsePlainError(string(stderr))
}

// tryParseWaveError attempts to parse a single line as a wave error JSON.
func tryParseWaveError(line string) *PluginError {
	// Quick check before attempting JSON parse
	if !strings.Contains(line, "wave_error") {
		return nil
	}

	var pe PluginError
	if err := jsonUnmarshal([]byte(line), &pe); err != nil {
		return nil
	}

	if !pe.WaveError {
		return nil
	}
	if pe.Code == "" || pe.Message == "" {
		return nil
	}

	return &pe
}

// tryParsePlainError attempts to parse stderr as plain text error format.
// Format: "code: message\ndetails" or "code: message"
// Returns nil if the format doesn't match.
func tryParsePlainError(stderr string) *PluginError {
	stderr = strings.TrimSpace(stderr)
	if stderr == "" {
		return nil
	}

	lines := strings.SplitN(stderr, "\n", 2)
	firstLine := lines[0]

	// First line should be "code: message"
	colonIdx := strings.Index(firstLine, ":")
	if colonIdx == -1 {
		return nil
	}

	code := strings.TrimSpace(firstLine[:colonIdx])
	message := strings.TrimSpace(firstLine[colonIdx+1:])

	// Validate code looks like an error code (lowercase-and-dashes)
	if !isValidErrorCode(code) {
		return nil
	}

	pe := &PluginError{
		WaveError: true,
		Code:      code,
		Message:   message,
	}

	// Second line (if present) is details
	if len(lines) > 1 {
		pe.Details = strings.TrimSpace(lines[1])
	}

	return pe
}

// isValidErrorCode checks if a string looks like a wave error code.
// Valid codes are lowercase-and-dashes with at least one dash (e.g., "flow-resolve-error").
func isValidErrorCode(s string) bool {
	if s == "" {
		return false
	}
	hasDash := false
	for _, c := range s {
		if c == '-' {
			hasDash = true
		} else if !(c >= 'a' && c <= 'z') {
			return false
		}
	}
	return hasDash
}
