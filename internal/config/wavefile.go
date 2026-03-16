package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// ParseWavefile reads and parses a Wavefile from the given path.
func ParseWavefile(path string) (*Wavefile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading Wavefile: %w", err)
	}

	// First pass: decode everything into a raw map
	var raw map[string]any
	if _, err := toml.Decode(string(data), &raw); err != nil {
		return nil, fmt.Errorf("parsing Wavefile: %w", err)
	}

	// Extract [project] section
	wf := &Wavefile{
		Sections: make(map[string]map[string]any),
	}

	if projRaw, ok := raw["project"]; ok {
		// Re-encode and decode into ProjectMeta for clean type mapping
		if projMap, ok := projRaw.(map[string]any); ok {
			wf.Project = decodeProjectMeta(projMap)
		}
	}

	// All other top-level sections are plugin sections
	for key, val := range raw {
		if key == "project" {
			continue
		}
		if section, ok := val.(map[string]any); ok {
			wf.Sections[key] = section
		}
	}

	return wf, nil
}

// decodeProjectMeta extracts ProjectMeta from a raw map.
func decodeProjectMeta(m map[string]any) ProjectMeta {
	pm := ProjectMeta{}

	if v, ok := m["name"].(string); ok {
		pm.Name = v
	}
	if v, ok := m["version"].(string); ok {
		pm.Version = v
	}
	if v, ok := m["owner"].(string); ok {
		pm.Owner = v
	}
	if v, ok := m["category"].(string); ok {
		pm.Category = v
	}
	if tags, ok := m["tags"].([]any); ok {
		for _, t := range tags {
			if s, ok := t.(string); ok {
				pm.Tags = append(pm.Tags, s)
			}
		}
	}

	return pm
}

// DiscoverWavefile walks up from startDir looking for a file named "Wavefile".
// Returns the absolute path to the Wavefile, or an error if not found.
func DiscoverWavefile(startDir string) (string, error) {
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
			// Reached filesystem root
			return "", fmt.Errorf("no Wavefile found (searched up from %s)", startDir)
		}
		dir = parent
	}
}
