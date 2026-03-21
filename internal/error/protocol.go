// Package errors defines the wave error protocol and provides parsing,
// formatting, and logging for plugin errors.
package errors

// PluginError represents a structured error from a wave plugin.
// Plugins emit this as JSON on stderr with wave_error: true.
type PluginError struct {
	WaveError bool   `json:"wave_error"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	Details   string `json:"details,omitempty"`
}
