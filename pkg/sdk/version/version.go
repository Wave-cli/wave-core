// Package version provides version information reading for wave plugins.
// Plugins can read project version from the Wavefile and access their own version.
package version

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// GetProjectVersion reads the project version from the Wavefile.
// It searches for a Wavefile starting from the current directory,
// walking up the directory tree until found.
// Returns the version string or an error if not found.
func GetProjectVersion() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting current directory: %w", err)
	}
	return GetProjectVersionFrom(cwd)
}

// GetProjectVersionFrom reads the project version from a Wavefile,
// starting the search from the specified directory.
func GetProjectVersionFrom(startDir string) (string, error) {
	wavefilePath, err := discoverWavefile(startDir)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(wavefilePath)
	if err != nil {
		return "", fmt.Errorf("reading Wavefile: %w", err)
	}

	version, err := parseVersionFromWavefile(data)
	if err != nil {
		return "", err
	}

	return version, nil
}

// GetPluginVersion returns the current plugin version from WAVE_PLUGIN_VERSION.
func GetPluginVersion() string {
	return os.Getenv("WAVE_PLUGIN_VERSION")
}

// discoverWavefile walks up from startDir looking for a file named "Wavefile".
func discoverWavefile(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", fmt.Errorf("resolving start dir: %w", err)
	}

	for {
		candidate := filepath.Join(dir, "Wavefile")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no Wavefile found (searched up from %s)", startDir)
		}
		dir = parent
	}
}

// parseVersionFromWavefile extracts the version from Wavefile content.
// Looks for version = "x.y.z" in the [project] section.
func parseVersionFromWavefile(data []byte) (string, error) {
	// Simple regex to find version in [project] section
	// Matches: version = "1.0.0" or version = '1.0.0'
	re := regexp.MustCompile(`(?m)^\s*version\s*=\s*["']([^"']+)["']`)
	matches := re.FindSubmatch(data)
	if matches == nil {
		return "", fmt.Errorf("no version found in Wavefile")
	}
	return string(matches[1]), nil
}
