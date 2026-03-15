// echo is a test plugin for wave-core end-to-end testing.
//
// Behavior:
//   - Reads config JSON from stdin
//   - If first arg is "fail", emits a structured wave error and exits 1
//   - Otherwise prints "OK echo <args> <key=value pairs from config>"
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

type waveError struct {
	WaveError bool   `json:"wave_error"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	Details   string `json:"details,omitempty"`
}

func main() {
	args := os.Args[1:]

	// Read config from stdin
	config := readConfig()

	// Check for fail command
	if len(args) > 0 && args[0] == "fail" {
		emitError("ECHO_FAIL", "intentional failure for testing", "use a different command")
		os.Exit(1)
	}

	// Build output
	parts := []string{"OK", "echo"}
	parts = append(parts, args...)

	// Append config key=value pairs (sorted for deterministic output)
	if len(config) > 0 {
		keys := make([]string, 0, len(config))
		for k := range config {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			parts = append(parts, fmt.Sprintf("%s=%v", k, config[k]))
		}
	}

	fmt.Println(strings.Join(parts, " "))
}

func readConfig() map[string]any {
	data, err := io.ReadAll(os.Stdin)
	if err != nil || len(data) == 0 {
		return nil
	}

	var cfg map[string]any
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil
	}
	return cfg
}

func emitError(code, message, details string) {
	e := waveError{
		WaveError: true,
		Code:      code,
		Message:   message,
		Details:   details,
	}
	json.NewEncoder(os.Stderr).Encode(e)
}
