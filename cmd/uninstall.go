package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wave-cli/wave-core/internal/config"
	"github.com/wave-cli/wave-core/internal/runner"
	"github.com/wave-cli/wave-core/internal/ui"
)

// NewUninstallCmd creates the 'wave uninstall' command.
func NewUninstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall <plugin>",
		Short: "Remove an installed plugin",
		Long:  "Removes a plugin and all its versions, assets, and configuration.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pluginName := args[0]

			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("getting home directory: %w", err)
			}

			configPath := filepath.Join(homeDir, ".wave", "config")
			pluginsDir := filepath.Join(homeDir, ".wave", "plugins")

			gc, err := config.ParseGlobalConfig(configPath)
			if err != nil {
				return fmt.Errorf("reading config: %w", err)
			}

			// Find the full name
			fullName, _, found := runner.LookupPlugin(pluginName, gc.Plugins)
			if !found {
				// Try as full name
				if _, ok := gc.Plugins[pluginName]; ok {
					fullName = pluginName
					found = true
				}
			}

			if !found {
				return fmt.Errorf("plugin %q is not installed", pluginName)
			}

			// Create spinner for uninstall progress
			spinner := ui.NewSpinner(os.Stderr, fmt.Sprintf("Uninstalling %s...", fullName))
			spinner.Start()

			// Remove plugin directory - use only plugin name (not org/name)
			// fullName is "org/name", we extract just "name"
			shortName := fullName
			if parts := strings.SplitN(fullName, "/", 2); len(parts) == 2 {
				shortName = parts[1]
			}
			pluginDir := filepath.Join(pluginsDir, shortName)
			if err := os.RemoveAll(pluginDir); err != nil {
				spinner.StopWithError(fmt.Sprintf("Failed to remove %s", fullName))
				return fmt.Errorf("removing plugin directory: %w", err)
			}

			// Remove from global config
			delete(gc.Plugins, fullName)
			if err := config.WriteGlobalConfig(configPath, gc); err != nil {
				spinner.StopWithError("Failed to update config")
				return fmt.Errorf("updating config: %w", err)
			}

			spinner.StopWithSuccess(fmt.Sprintf("Uninstalled %s", fullName))
			return nil
		},
	}
}
