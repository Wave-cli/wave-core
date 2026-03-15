// Package bootstrap handles first-run setup of the wave home directory.
// It ensures ~/.wave/, ~/.wave/config, ~/.wave/plugins/, and the logs
// directory all exist before any other wave operations run.
package bootstrap

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/wave-cli/wave-core/internal/config"
)

// Ensure guarantees that the wave home directory structure and config file
// exist. If ~/.wave/config already exists, it is read and returned without
// modification. If it doesn't exist, a default config is created.
//
// homeDir is the user's home directory (typically os.UserHomeDir()).
func Ensure(homeDir string) (*config.GlobalConfig, error) {
	waveDir := filepath.Join(homeDir, ".wave")
	configPath := filepath.Join(waveDir, "config")
	pluginsDir := filepath.Join(waveDir, "plugins")

	// Create base directories
	dirs := []string{waveDir, pluginsDir}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return nil, fmt.Errorf("creating directory %s: %w", d, err)
		}
	}

	// If config exists, read and return it
	if _, err := os.Stat(configPath); err == nil {
		gc, err := config.ParseGlobalConfig(configPath)
		if err != nil {
			return nil, fmt.Errorf("reading existing config: %w", err)
		}

		// Ensure logs dir exists (it may reference a custom path)
		if gc.Core.LogsDir != "" {
			logsDir := expandHome(gc.Core.LogsDir, homeDir)
			if err := os.MkdirAll(logsDir, 0755); err != nil {
				return nil, fmt.Errorf("creating logs directory %s: %w", logsDir, err)
			}
		}

		return gc, nil
	}

	// First run: create default config
	gc := config.DefaultGlobalConfig(homeDir)

	// Create logs directory
	if err := os.MkdirAll(gc.Core.LogsDir, 0755); err != nil {
		return nil, fmt.Errorf("creating logs directory: %w", err)
	}

	// Write config file
	if err := config.WriteGlobalConfig(configPath, gc); err != nil {
		return nil, fmt.Errorf("writing default config: %w", err)
	}

	return gc, nil
}

// expandHome replaces a leading ~ with the actual home directory.
func expandHome(path, homeDir string) string {
	if len(path) >= 2 && path[:2] == "~/" {
		return filepath.Join(homeDir, path[2:])
	}
	if path == "~" {
		return homeDir
	}
	return path
}
