package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/wave-cli/wave-core/internal/config"
)

// NewConfigCmd creates the config command.
func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage global wave configuration",
	}

	cmd.AddCommand(NewConfigSetCmd())
	return cmd
}

// NewConfigSetCmd updates fields in the global config.
func NewConfigSetCmd() *cobra.Command {
	var name string
	var org string

	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Update global config values",
		RunE: func(cmd *cobra.Command, args []string) error {
			if globalCfg == nil {
				return fmt.Errorf("global config not loaded")
			}

			nameSet := cmd.Flags().Changed("name")
			orgSet := cmd.Flags().Changed("org")
			if !nameSet && !orgSet {
				return fmt.Errorf("no config fields provided")
			}

			if nameSet {
				globalCfg.User.Name = name
			}
			if orgSet {
				globalCfg.User.Org = org
			}

			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("getting home directory: %w", err)
			}

			configPath := filepath.Join(homeDir, ".wave", "config")
			if err := config.WriteGlobalConfig(configPath, globalCfg); err != nil {
				return err
			}

			if printer != nil {
				printer.Success("Updated global config")
			}
			return nil
		},
	}

	setCmd.Flags().StringVar(&name, "name", "", "User name")
	setCmd.Flags().StringVar(&org, "org", "", "Organization")

	return setCmd
}
