package errors

import (
	"encoding/json"
	"fmt"
	"strings"
)

// FormatError renders a PluginError into a human-friendly string for terminal display.
// When debug is true, it shows the full JSON error structure.
// When debug is false, it shows only the user-friendly message.
func FormatError(pluginName, pluginVersion string, pe *PluginError, logPath string, debug bool) string {
	if debug {
		return formatDebugError(pluginName, pluginVersion, pe, logPath)
	}
	return formatSimpleError(pe)
}

// formatSimpleError returns only the error message for normal users.
func formatSimpleError(pe *PluginError) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s", pe.Message)
	if pe.Details != "" {
		fmt.Fprintf(&b, "\n%s", pe.Details)
	}
	return b.String()
}

// formatDebugError returns the full error structure for debugging.
func formatDebugError(pluginName, pluginVersion string, pe *PluginError, logPath string) string {
	var b strings.Builder

	// Show full JSON structure
	jsonBytes, _ := json.MarshalIndent(pe, "  ", "  ")
	fmt.Fprintf(&b, "ERROR [%s]\n", pe.Code)
	fmt.Fprintf(&b, "  %s\n", string(jsonBytes))

	fmt.Fprintf(&b, "\n  Plugin: %s", pluginName)
	if pluginVersion != "" {
		fmt.Fprintf(&b, " v%s", pluginVersion)
	}
	fmt.Fprintln(&b)

	if logPath != "" {
		fmt.Fprintf(&b, "  Logged: %s\n", logPath)
	}

	return b.String()
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
