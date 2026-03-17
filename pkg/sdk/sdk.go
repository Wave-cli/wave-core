// Package sdk provides helpers for wave plugin authors.
// Plugins import this package to read config from stdin,
// emit structured errors, and access wave environment variables.
package sdk

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
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
// Error codes MUST use lowercase-and-dashes format (e.g. "flow-no-command").
type WaveError struct {
	WaveError bool   `json:"wave_error"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	Details   string `json:"details,omitempty"`
}

// Config wraps the raw plugin configuration map and provides typed access.
// wave-core passes the plugin's entire TOML section as a JSON object on stdin.
// The plugin receives it as-is — no schema validation, no transformation.
type Config struct {
	data map[string]any
}

// Raw returns the underlying configuration map.
func (c *Config) Raw() map[string]any {
	return c.data
}

// Get returns the value for key and whether it exists.
func (c *Config) Get(key string) (any, bool) {
	v, ok := c.data[key]
	return v, ok
}

// Has returns true if key exists in the config.
func (c *Config) Has(key string) bool {
	_, ok := c.data[key]
	return ok
}

// String returns the string value for key, or ("", false) if missing/wrong type.
func (c *Config) String(key string) (string, bool) {
	v, ok := c.data[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

// Bool returns the bool value for key, or (false, false) if missing/wrong type.
func (c *Config) Bool(key string) (bool, bool) {
	v, ok := c.data[key]
	if !ok {
		return false, false
	}
	b, ok := v.(bool)
	return b, ok
}

// Float returns the float64 value for key, or (0, false) if missing/wrong type.
// JSON numbers are always decoded as float64.
func (c *Config) Float(key string) (float64, bool) {
	v, ok := c.data[key]
	if !ok {
		return 0, false
	}
	f, ok := v.(float64)
	return f, ok
}

// Map returns the map value for key, or (nil, false) if missing/wrong type.
func (c *Config) Map(key string) (map[string]any, bool) {
	v, ok := c.data[key]
	if !ok {
		return nil, false
	}
	m, ok := v.(map[string]any)
	return m, ok
}

// Keys returns all top-level keys in sorted order.
func (c *Config) Keys() []string {
	keys := make([]string, 0, len(c.data))
	for k := range c.data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Plugin bundles the environment and config for a wave plugin.
// This is the primary entry point for plugin authors.
type Plugin struct {
	Env    PluginEnv
	Config *Config
}

// Init reads config from os.Stdin and env vars, returning a ready Plugin.
// This is the main entry point for plugin initialization.
func Init() (*Plugin, error) {
	return InitFrom(os.Stdin)
}

// InitFrom reads config from the given reader and env vars.
// Useful for testing.
func InitFrom(r io.Reader) (*Plugin, error) {
	cfg, err := ReadConfigFrom(r)
	if err != nil {
		return nil, err
	}
	return &Plugin{
		Env:    GetPluginEnv(),
		Config: cfg,
	}, nil
}

// ReadConfig reads the plugin configuration from os.Stdin.
// wave-core passes the plugin's config section as a JSON object on stdin.
func ReadConfig() (*Config, error) {
	return ReadConfigFrom(os.Stdin)
}

// ReadConfigFrom reads plugin configuration from an arbitrary reader.
// This is useful for testing.
func ReadConfigFrom(r io.Reader) (*Config, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading config from stdin: %w", err)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("empty config input")
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing config JSON: %w", err)
	}

	return &Config{data: raw}, nil
}

// Err emits a structured wave error to stderr and exits with code 1.
// Code MUST be lowercase-and-dashes format (e.g. "flow-no-command").
// Message should be a human-readable description of the error.
func Err(code, message string) {
	FormatWaveError(os.Stderr, code, message, "")
	os.Exit(1)
}

// ErrWithDetails emits a structured wave error with additional details.
// Code MUST be lowercase-and-dashes format (e.g. "flow-resolve-error").
// Details provides additional context (suggestions, available commands, etc.).
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

// Errf creates a new error with a formatted message.
// This is a convenience wrapper around fmt.Errorf for plugin authors.
// It supports all fmt.Errorf formatting verbs including %w for wrapping errors.
func Errf(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}
