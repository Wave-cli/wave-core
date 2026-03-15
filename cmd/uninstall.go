package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/wave-cli/wave-core/internal/config"
	"github.com/wave-cli/wave-core/internal/runner"
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

			printer.Info("Uninstalling %s...", fullName)

			// Remove plugin directory
			parts := filepath.SplitList(fullName)
			if len(parts) == 0 {
				// Use string split for org/name
				pluginDir := filepath.Join(pluginsDir, filepath.FromSlash(fullName))
				os.RemoveAll(pluginDir)
			} else {
				pluginDir := filepath.Join(pluginsDir, filepath.FromSlash(fullName))
				os.RemoveAll(pluginDir)
			}

			// Remove from global config
			delete(gc.Plugins, fullName)
			if err := config.WriteGlobalConfig(configPath, gc); err != nil {
				return fmt.Errorf("updating config: %w", err)
			}

			printer.Success("Uninstalled %s", fullName)
			return nil
		},
	}
}
