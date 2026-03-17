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

	cmd.AddCommand(NewConfigSetUserCmd())
	cmd.AddCommand(NewConfigSetOrgCmd())
	return cmd
}

// NewConfigSetUserCmd sets the user name in global config.
func NewConfigSetUserCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set-user <name>",
		Short: "Set the user name in global config",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("getting home directory: %w", err)
			}

			configPath := filepath.Join(homeDir, ".wave", "config")
			gc, err := config.ParseGlobalConfig(configPath)
			if err != nil {
				gc = config.DefaultGlobalConfig(homeDir)
			}

			gc.User.Name = name

			if err := config.WriteGlobalConfig(configPath, gc); err != nil {
				return fmt.Errorf("writing config: %w", err)
			}

			if printer != nil {
				printer.Success("Set user name to %q", name)
			}
			return nil
		},
	}
}

// NewConfigSetOrgCmd sets the organization in global config.
func NewConfigSetOrgCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set-org <org>",
		Short: "Set the default organization in global config",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			org := args[0]

			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("getting home directory: %w", err)
			}

			configPath := filepath.Join(homeDir, ".wave", "config")
			gc, err := config.ParseGlobalConfig(configPath)
			if err != nil {
				gc = config.DefaultGlobalConfig(homeDir)
			}

			gc.User.Org = org

			if err := config.WriteGlobalConfig(configPath, gc); err != nil {
				return fmt.Errorf("writing config: %w", err)
			}

			if printer != nil {
				printer.Success("Set organization to %q", org)
			}
			return nil
		},
	}
}
