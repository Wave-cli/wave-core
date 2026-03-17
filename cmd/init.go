package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/wave-cli/wave-core/internal/config"
)

// NewInitCmd creates the 'wave init' command.
func NewInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init [project-name]",
		Short: "Initialize a new Wavefile in the current directory",
		Long:  "Creates a Wavefile in the current directory with default project metadata.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting current directory: %w", err)
			}

			wavefilePath := filepath.Join(cwd, "Wavefile")

			// Check if Wavefile already exists
			if _, err := os.Stat(wavefilePath); err == nil {
				return fmt.Errorf("Wavefile already exists in %s", cwd)
			}

			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("getting home directory: %w", err)
			}
			configPath := filepath.Join(homeDir, ".wave", "config")

			// Load global config to get default org
			gc, err := config.ParseGlobalConfig(configPath)
			if err != nil {
				gc = config.DefaultGlobalConfig(homeDir)
			}

			// Use provided project name or fall back to directory name
			var projectName string
			if len(args) > 0 && args[0] != "" {
				projectName = args[0]
			} else {
				projectName = filepath.Base(cwd)
			}

			// Use default org from config as owner
			owner := gc.User.Org

			content := fmt.Sprintf(`[project]
name = %q
version = "0.1.0"
owner = %q
category = ""
tags = []
`, projectName, owner)

			if err := os.WriteFile(wavefilePath, []byte(content), 0644); err != nil {
				return fmt.Errorf("writing Wavefile: %w", err)
			}

			// Add project folder to global config if not already present
			if !containsFolder(gc.Projects.Folders, cwd) {
				gc.Projects.Folders = append(gc.Projects.Folders, cwd)
			}

			// Write updated global config
			if err := config.WriteGlobalConfig(configPath, gc); err != nil {
				return fmt.Errorf("updating global config: %w", err)
			}

			printer.Success("Created Wavefile in %s", cwd)
			return nil
		},
	}
}

// containsFolder checks if a folder path is already in the list.
func containsFolder(folders []string, folder string) bool {
	for _, f := range folders {
		if f == folder {
			return true
		}
	}
	return false
}
