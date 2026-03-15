package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// ParseGlobalConfig reads and parses ~/.wave/config from the given path.
func ParseGlobalConfig(path string) (*GlobalConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading global config: %w", err)
	}

	var gc GlobalConfig
	if _, err := toml.Decode(string(data), &gc); err != nil {
		return nil, fmt.Errorf("parsing global config: %w", err)
	}

	// Ensure maps are initialized
	if gc.Plugins == nil {
		gc.Plugins = make(map[string]string)
	}

	return &gc, nil
}

// WriteGlobalConfig writes a GlobalConfig to the given path as TOML.
func WriteGlobalConfig(path string, gc *GlobalConfig) error {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating config file: %w", err)
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(gc); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
}

// DefaultGlobalConfig returns a GlobalConfig with sensible defaults.
// homeDir should be the user's home directory.
func DefaultGlobalConfig(homeDir string) *GlobalConfig {
	return &GlobalConfig{
		Core: CoreConfig{
			LogsDir: filepath.Join(homeDir, ".wave", "logs"),
		},
		Projects: ProjectsConfig{
			Folders: []string{},
		},
		Plugins: make(map[string]string),
	}
}
