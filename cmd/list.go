package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"
	"github.com/wave-cli/wave-core/internal/pluginmgmt"
)

// NewListCmd creates the 'wave list' command.
func NewListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed plugins",
		Long:  "Shows all installed wave plugins with their versions and descriptions.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if globalCfg == nil || len(globalCfg.Plugins) == 0 {
				printer.Info("No plugins installed.")
				return nil
			}

			homeDir, _ := os.UserHomeDir()
			pluginsDir := filepath.Join(homeDir, ".wave", "plugins")
			reg := pluginmgmt.NewRegistry(pluginsDir)

			list := reg.ListInstalled(globalCfg.Plugins)

			// Sort by name
			sort.Slice(list, func(i, j int) bool {
				return list[i].FullName < list[j].FullName
			})

			printer.Info("Installed plugins:\n")
			for _, p := range list {
				description := ""
				if wp, err := reg.ReadWaveplugin(p.FullName); err == nil {
					description = wp.Plugin.Description
				}

				if description != "" {
					fmt.Fprintf(os.Stdout, "  %s  v%s  - %s\n", p.FullName, p.Version, description)
				} else {
					fmt.Fprintf(os.Stdout, "  %s  v%s\n", p.FullName, p.Version)
				}
			}

			return nil
		},
	}
}
