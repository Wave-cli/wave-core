// Package sdk provides helpers for wave plugin authors.
// Plugins import this package to read config from stdin,
// emit structured errors, and access wave environment variables.
package sdk

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// PluginEnv holds wave environment variables available to plugins.
type PluginEnv struct {
	Name        string // WAVE_PLUGIN_NAME
	Version     string // WAVE_PLUGIN_VERSION
	Dir         string // WAVE_PLUGIN_DIR
	Assets      string // WAVE_PLUGIN_ASSETS
	ProjectRoot string // WAVE_PROJECT_ROOT
}

// WaveError is the JSON structure emitted to stderr for structured errors.
type WaveError struct {
	WaveError bool   `json:"wave_error"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	Details   string `json:"details,omitempty"`
}

// ReadConfig reads the plugin configuration from os.Stdin.
// wave-core passes the plugin's config section as a JSON object on stdin.
func ReadConfig() (map[string]any, error) {
	return ReadConfigFrom(os.Stdin)
}

// ReadConfigFrom reads plugin configuration from an arbitrary reader.
// This is useful for testing.
func ReadConfigFrom(r io.Reader) (map[string]any, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading config from stdin: %w", err)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("empty config input")
	}

	var cfg map[string]any
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config JSON: %w", err)
	}

	return cfg, nil
}

// Err emits a structured wave error to stderr and exits with code 1.
// Use this as the primary way to report errors from plugins.
func Err(code, message string) {
	FormatWaveError(os.Stderr, code, message, "")
	os.Exit(1)
}

// ErrWithDetails emits a structured wave error with additional details.
func ErrWithDetails(code, message, details string) {
	FormatWaveError(os.Stderr, code, message, details)
	os.Exit(1)
}

// FormatWaveError writes a structured wave error JSON to the given writer.
// This is the testable version of Err().
func FormatWaveError(w io.Writer, code, message, details string) {
	e := WaveError{
		WaveError: true,
		Code:      code,
		Message:   message,
		Details:   details,
	}
	json.NewEncoder(w).Encode(e)
}

// GetPluginEnv reads wave environment variables set by wave-core.
func GetPluginEnv() PluginEnv {
	return PluginEnv{
		Name:        os.Getenv("WAVE_PLUGIN_NAME"),
		Version:     os.Getenv("WAVE_PLUGIN_VERSION"),
		Dir:         os.Getenv("WAVE_PLUGIN_DIR"),
		Assets:      os.Getenv("WAVE_PLUGIN_ASSETS"),
		ProjectRoot: os.Getenv("WAVE_PROJECT_ROOT"),
	}
}
