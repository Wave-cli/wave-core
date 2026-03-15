// Package pluginmgmt manages plugin metadata, name resolution, and the
// local plugin registry (installed plugins under ~/.wave/plugins/).
package pluginmgmt

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// Waveplugin represents the parsed Waveplugin metadata file that ships
// with every plugin release.
type Waveplugin struct {
	Plugin        PluginMeta        `toml:"plugin"`
	Compatibility CompatibilityMeta `toml:"compatibility"`
	Assets        AssetsMeta        `toml:"assets"`
}

// PluginMeta holds the [plugin] section of a Waveplugin file.
type PluginMeta struct {
	Name        string `toml:"name"`
	Version     string `toml:"version"`
	Description string `toml:"description"`
	Creator     string `toml:"creator"`
	License     string `toml:"license"`
	Homepage    string `toml:"homepage"`
}

// CompatibilityMeta holds the [compatibility] section.
type CompatibilityMeta struct {
	MinWaveVersion string `toml:"min_wave_version"`
}

// AssetsMeta holds the [assets] section.
type AssetsMeta struct {
	Schema string   `toml:"schema"` // Optional schema filename (e.g. "Waveschema")
	Files  []string `toml:"files"`
}

// ParseWaveplugin reads and decodes a Waveplugin TOML file at path.
// It guarantees Assets.Files is never nil (empty slice for minimal plugins).
func ParseWaveplugin(path string) (*Waveplugin, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read Waveplugin: %w", err)
	}

	var wp Waveplugin
	if err := toml.Unmarshal(data, &wp); err != nil {
		return nil, fmt.Errorf("parse Waveplugin: %w", err)
	}

	// Ensure Files is never nil so callers don't need nil checks.
	if wp.Assets.Files == nil {
		wp.Assets.Files = []string{}
	}

	return &wp, nil
}
