package pluginmgmt

import (
	"fmt"
	"os"
	"path/filepath"
)

// Registry manages installed plugins under a plugins directory.
// Layout: <pluginsDir>/<name>/bin/<name>
type Registry struct {
	pluginsDir string
}

// InstalledPlugin describes a plugin entry from the global config.
type InstalledPlugin struct {
	FullName string
	Version  string
}

// NewRegistry creates a Registry rooted at pluginsDir
// (typically ~/.wave/plugins).
func NewRegistry(pluginsDir string) *Registry {
	return &Registry{pluginsDir: pluginsDir}
}

// ResolveBinary returns the absolute path to the installed plugin binary:
// <pluginsDir>/<name>/bin/<name>.
func (r *Registry) ResolveBinary(fullName string) (string, error) {
	ref, err := ParsePluginRef(fullName)
	if err != nil {
		return "", err
	}

	// Use only the plugin name, not org/name
	binPath := filepath.Join(r.pluginsDir, ref.Name, "bin", ref.Name)

	if _, err := os.Stat(binPath); err != nil {
		return "", fmt.Errorf("plugin %q not installed: %w", fullName, err)
	}

	return binPath, nil
}

// ResolveAssets returns the path to the plugin's assets directory:
// <pluginsDir>/<name>/assets.
func (r *Registry) ResolveAssets(fullName string) (string, error) {
	ref, err := ParsePluginRef(fullName)
	if err != nil {
		return "", err
	}

	// Use only the plugin name, not org/name
	assetsPath := filepath.Join(r.pluginsDir, ref.Name, "assets")

	if _, err := os.Stat(assetsPath); err != nil {
		return "", fmt.Errorf("plugin %q assets not found: %w", fullName, err)
	}

	return assetsPath, nil
}

// ReadWaveplugin reads and parses the Waveplugin metadata file for an
// installed plugin.
func (r *Registry) ReadWaveplugin(fullName string) (*Waveplugin, error) {
	ref, err := ParsePluginRef(fullName)
	if err != nil {
		return nil, err
	}

	// Use only the plugin name, not org/name
	wpPath := filepath.Join(r.pluginsDir, ref.Name, "Waveplugin")
	return ParseWaveplugin(wpPath)
}

// ListInstalled converts the plugins map (from global config) into a
// slice of InstalledPlugin entries.
func (r *Registry) ListInstalled(plugins map[string]string) []InstalledPlugin {
	list := make([]InstalledPlugin, 0, len(plugins))
	for name, version := range plugins {
		list = append(list, InstalledPlugin{
			FullName: name,
			Version:  version,
		})
	}
	return list
}
