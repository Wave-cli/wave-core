package errors

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ANSI color codes
const (
	colorRed   = "\033[31m"
	colorReset = "\033[0m"
)

// FormatError renders a PluginError into a human-friendly string for terminal display.
// When debug is true, it prints the raw JSON to stderr.
// When debug is false, it shows a colored user-friendly message.
func FormatError(pluginName, pluginVersion string, pe *PluginError, logPath string, debug bool) string {
	if debug {
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

// ParseStderr scans stderr output for a structured wave error JSON.
// It looks through each line for a JSON object with wave_error: true.
// Returns nil if no structured error is found.
func ParseStderr(stderr []byte) *PluginError {
	if len(stderr) == 0 {
		return nil
	}

	lines := strings.Split(string(stderr), "\n")
	// Search lines in reverse (error is typically the last structured output)
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if pe := tryParseWaveError(line); pe != nil {
			return pe
		}
	}

	return nil
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
