// Package env provides access to wave environment variables for plugins.
// wave-core sets these variables before executing a plugin, providing
// context about the plugin and project.
package env

import (
	"os"
)

// PluginEnv holds wave environment variables available to plugins.
type PluginEnv struct {
	Name        string // WAVE_PLUGIN_NAME - plugin name (e.g., "flow")
	Version     string // WAVE_PLUGIN_VERSION - plugin version (e.g., "1.2.0")
	Dir         string // WAVE_PLUGIN_DIR - plugin installation directory
	Assets      string // WAVE_PLUGIN_ASSETS - plugin assets directory
	ProjectRoot string // WAVE_PROJECT_ROOT - root directory of the current project
}

// Get reads wave environment variables set by wave-core.
// Returns a populated PluginEnv with all available environment info.
func Get() PluginEnv {
	return PluginEnv{
		Name:        os.Getenv("WAVE_PLUGIN_NAME"),
		Version:     os.Getenv("WAVE_PLUGIN_VERSION"),
		Dir:         os.Getenv("WAVE_PLUGIN_DIR"),
		Assets:      os.Getenv("WAVE_PLUGIN_ASSETS"),
		ProjectRoot: os.Getenv("WAVE_PROJECT_ROOT"),
	}
}

// GetName returns the plugin name from WAVE_PLUGIN_NAME.
func GetName() string {
	return os.Getenv("WAVE_PLUGIN_NAME")
}

// GetVersion returns the plugin version from WAVE_PLUGIN_VERSION.
func GetVersion() string {
	return os.Getenv("WAVE_PLUGIN_VERSION")
}

// GetDir returns the plugin directory from WAVE_PLUGIN_DIR.
func GetDir() string {
	return os.Getenv("WAVE_PLUGIN_DIR")
}

// GetAssets returns the plugin assets directory from WAVE_PLUGIN_ASSETS.
func GetAssets() string {
	return os.Getenv("WAVE_PLUGIN_ASSETS")
}

// GetProjectRoot returns the project root from WAVE_PROJECT_ROOT.
func GetProjectRoot() string {
	return os.Getenv("WAVE_PROJECT_ROOT")
}
