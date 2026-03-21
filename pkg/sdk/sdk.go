// Package sdk provides helpers for wave plugin authors.
// Plugins import this package to read config from stdin,
// emit structured errors, and access wave environment variables.
//
// The SDK is organized into subpackages for modularity:
//   - sdk/config: Configuration reading and typed access
//   - sdk/error: Structured error emission
//   - sdk/env: Wave environment variables
//   - sdk/version: Project and plugin version reading
//
// This main package re-exports common types for convenience.
package sdk

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"

	sdkconfig "github.com/wave-cli/wave-core/pkg/sdk/config"
	sdkenv "github.com/wave-cli/wave-core/pkg/sdk/env"
	sdkerror "github.com/wave-cli/wave-core/pkg/sdk/error"
	sdkversion "github.com/wave-cli/wave-core/pkg/sdk/version"
)

// PluginEnv holds wave environment variables available to plugins.
// Re-exported from sdk/env for convenience.
type PluginEnv = sdkenv.PluginEnv

// WaveError is the JSON structure emitted to stderr for structured errors.
// Error codes MUST use lowercase-and-dashes format (e.g. "flow-no-command").
// Deprecated: Use sdk/error.WaveError instead. Errors are now emitted as plain text.
type WaveError struct {
	WaveError bool   `json:"wave_error"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	Details   string `json:"details,omitempty"`
}

// Config wraps the raw plugin configuration map and provides typed access.
// Re-exported from sdk/config for convenience.
type Config = sdkconfig.Config

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
	cfg, err := sdkconfig.ReadFrom(r)
	if err != nil {
		return nil, err
	}
	return &Plugin{
		Env:    sdkenv.Get(),
		Config: cfg,
	}, nil
}

// ReadConfig reads the plugin configuration from os.Stdin.
// wave-core passes the plugin's config section as a JSON object on stdin.
func ReadConfig() (*Config, error) {
	return sdkconfig.Read()
}

// ReadConfigFrom reads plugin configuration from an arbitrary reader.
// This is useful for testing.
func ReadConfigFrom(r io.Reader) (*Config, error) {
	return sdkconfig.ReadFrom(r)
}

// Err emits a structured wave error to stderr and exits with code 1.
// Code MUST be lowercase-and-dashes format (e.g. "flow-no-command").
// Message should be a human-readable description of the error.
func Err(code, message string) {
	sdkerror.Emit(code, message)
}

// ErrWithDetails emits a structured wave error with additional details.
// Code MUST be lowercase-and-dashes format (e.g. "flow-resolve-error").
// Details provides additional context (suggestions, available commands, etc.).
func ErrWithDetails(code, message, details string) {
	sdkerror.EmitWithDetails(code, message, details)
}

// FormatWaveError writes a structured wave error JSON to the given writer.
// This is the testable version of Err().
// Deprecated: Use sdk/error.Format instead. Errors are now emitted as plain text.
func FormatWaveError(w io.Writer, code, message, details string) {
	// For backwards compatibility, still emit JSON for existing plugins
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
	return sdkenv.Get()
}

// GetVersion returns the project version from the Wavefile.
// This reads the version from the [project] section of the nearest Wavefile.
func GetVersion() (string, error) {
	return sdkversion.GetProjectVersion()
}

// GetVersionFrom returns the project version from a Wavefile,
// starting the search from the specified directory.
func GetVersionFrom(startDir string) (string, error) {
	return sdkversion.GetProjectVersionFrom(startDir)
}

// GetPluginVersion returns the current plugin version from WAVE_PLUGIN_VERSION.
func GetPluginVersion() string {
	return sdkversion.GetPluginVersion()
}

// Errf creates a new error with a formatted message.
// This is a convenience wrapper around fmt.Errorf for plugin authors.
// It supports all fmt.Errorf formatting verbs including %w for wrapping errors.
func Errf(format string, args ...any) error {
	return sdkerror.Errf(format, args...)
}

// NewConfig creates a Config from an existing map.
// Useful for testing or constructing configs programmatically.
func NewConfig(data map[string]any) *Config {
	return sdkconfig.New(data)
}

// =============================================================================
// Backwards Compatibility - Inline Config methods
// =============================================================================

// ConfigData provides backwards-compatible inline Config implementation
// These are re-exported from the Config type alias

// Raw returns the underlying configuration map.
func Raw(c *Config) map[string]any {
	return c.Raw()
}

// Get returns the value for key and whether it exists.
func Get(c *Config, key string) (any, bool) {
	return c.Get(key)
}

// Has returns true if key exists in the config.
func Has(c *Config, key string) bool {
	return c.Has(key)
}

// String returns the string value for key, or ("", false) if missing/wrong type.
func String(c *Config, key string) (string, bool) {
	return c.String(key)
}

// Bool returns the bool value for key, or (false, false) if missing/wrong type.
func Bool(c *Config, key string) (bool, bool) {
	return c.Bool(key)
}

// Float returns the float64 value for key, or (0, false) if missing/wrong type.
func Float(c *Config, key string) (float64, bool) {
	return c.Float(key)
}

// Map returns the map value for key, or (nil, false) if missing/wrong type.
func Map(c *Config, key string) (map[string]any, bool) {
	return c.Map(key)
}

// Keys returns all top-level keys in sorted order.
func Keys(c *Config) []string {
	keys := make([]string, 0, len(c.Raw()))
	for k := range c.Raw() {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Errf creates a new error with a formatted message (standalone function).
func errf(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}
