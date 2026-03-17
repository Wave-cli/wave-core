package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/wave-cli/wave-core/internal/downloader"
	"github.com/wave-cli/wave-core/internal/pluginmgmt"
	"github.com/wave-cli/wave-core/internal/ui"
)

// NewRestoreCmd creates the 'wave restore' command.
func NewRestoreCmd() *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Reinstall all plugins from config",
		Long:  "Downloads and reinstalls all plugins listed in the global config file.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if globalCfg == nil || len(globalCfg.Plugins) == 0 {
				printer.Info("No plugins to restore.")
				return nil
			}

			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("getting home directory: %w", err)
			}

			pluginsDir := filepath.Join(homeDir, ".wave", "plugins")

			// Get GitHub token from environment if available
			token := os.Getenv("GITHUB_TOKEN")
			client := downloader.NewClient("", token)

			if dryRun {
				printer.Info("Dry run - would restore these plugins:")
				for fullName, version := range globalCfg.Plugins {
					fmt.Fprintf(os.Stdout, "  %s@%s\n", fullName, version)
				}
				return nil
			}

			total := len(globalCfg.Plugins)
			restored := 0
			failed := 0

			for fullName, version := range globalCfg.Plugins {
				ref, err := pluginmgmt.ParsePluginRef(fullName + "@" + version)
				if err != nil {
					printer.Warn("Skipping %s: invalid reference", fullName)
					failed++
					continue
				}

				spinner := ui.NewSpinner(os.Stderr, fmt.Sprintf("Restoring %s@%s...", fullName, version))
				spinner.Start()

				err = client.InstallPlugin(ref.Org, ref.Name, ref.Version, pluginsDir)
				if err != nil {
					spinner.StopWithError(fmt.Sprintf("Failed to restore %s", fullName))
					printer.Warn("  %v", err)
					failed++
					continue
				}

				spinner.StopWithSuccess(fmt.Sprintf("Restored %s@%s", fullName, version))
				restored++
			}

			if failed > 0 {
				printer.Warn("Restored %d/%d plugins (%d failed)", restored, total, failed)
			} else {
				printer.Success("Restored all %d plugins", total)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be restored without downloading")

	return cmd
}
