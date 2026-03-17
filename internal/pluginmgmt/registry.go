package pluginmgmt

import (
	"fmt"
	"path/filepath"
)

// Registry manages installed plugins under a plugins directory.
// Layout: <pluginsDir>/<name>/<version>/bin/<name>
// A "current" symlink points to the active version directory.
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

// ResolveBinary returns the absolute path to the installed plugin binary
// via the "current" symlink: <pluginsDir>/<name>/current/bin/<name>.
func (r *Registry) ResolveBinary(fullName string) (string, error) {
	ref, err := ParsePluginRef(fullName)
	if err != nil {
		return "", err
	}

	// Use only the plugin name, not org/name
	binPath := filepath.Join(r.pluginsDir, ref.Name, "current", "bin", ref.Name)

	// Verify the symlink chain is valid by resolving it.
	resolved, err := filepath.EvalSymlinks(binPath)
	if err != nil {
		return "", fmt.Errorf("plugin %q not installed: %w", fullName, err)
	}
	_ = resolved // we return the symlink-based path for consistency

	return binPath, nil
}

// ResolveAssets returns the path to the plugin's assets directory
// via the "current" symlink: <pluginsDir>/<name>/current/assets.
func (r *Registry) ResolveAssets(fullName string) (string, error) {
	ref, err := ParsePluginRef(fullName)
	if err != nil {
		return "", err
	}

	// Use only the plugin name, not org/name
	assetsPath := filepath.Join(r.pluginsDir, ref.Name, "current", "assets")

	// Verify the symlink chain is valid.
	if _, err := filepath.EvalSymlinks(assetsPath); err != nil {
		return "", fmt.Errorf("plugin %q assets not found: %w", fullName, err)
	}

	return assetsPath, nil
}

// ReadWaveplugin reads and parses the Waveplugin metadata file for an
// installed plugin via the "current" symlink.
func (r *Registry) ReadWaveplugin(fullName string) (*Waveplugin, error) {
	ref, err := ParsePluginRef(fullName)
	if err != nil {
		return nil, err
	}

	// Use only the plugin name, not org/name
	wpPath := filepath.Join(r.pluginsDir, ref.Name, "current", "Waveplugin")
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
