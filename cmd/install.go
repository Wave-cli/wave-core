package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wave-cli/wave-core/internal/config"
	"github.com/wave-cli/wave-core/internal/downloader"
	"github.com/wave-cli/wave-core/internal/pluginmgmt"
	"github.com/wave-cli/wave-core/internal/version"
)

// NewInstallCmd creates the 'wave install' command.
func NewInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install <org/plugin>[@version]",
		Short: "Install a plugin from GitHub Releases",
		Long:  "Downloads and installs a wave plugin from its GitHub release assets.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, err := pluginmgmt.ParsePluginRef(args[0])
			if err != nil {
				return fmt.Errorf("invalid plugin reference: %w", err)
			}

			printer.Info("Installing %s...", ref.FullName())

			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("getting home directory: %w", err)
			}

			pluginsDir := filepath.Join(homeDir, ".wave", "plugins")
			configPath := filepath.Join(homeDir, ".wave", "config")

			// Get GitHub token from environment if available
			token := os.Getenv("GITHUB_TOKEN")

			client := downloader.NewClient("", token)
			err = client.InstallPlugin(ref.Org, ref.Name, ref.Version, pluginsDir)
			if err != nil {
				return fmt.Errorf("installing %s: %w", ref.FullName(), err)
			}

			// Read installed version from Waveplugin
			reg := pluginmgmt.NewRegistry(pluginsDir)
			installedVersion := ref.Version
			if wp, err := reg.ReadWaveplugin(ref.FullName()); err == nil {
				installedVersion = wp.Plugin.Version

				// Check compatibility
				info := version.Get()
				if !version.SatisfiesMin(info.Version, wp.Compatibility.MinWaveVersion) {
					printer.Warn("Plugin requires wave >= %s (you have %s)",
						wp.Compatibility.MinWaveVersion, info.Version)
				}
			}

			// Update global config
			gc, err := config.ParseGlobalConfig(configPath)
			if err != nil {
				gc = config.DefaultGlobalConfig(homeDir)
			}
			gc.Plugins[ref.FullName()] = installedVersion
			if err := config.WriteGlobalConfig(configPath, gc); err != nil {
				return fmt.Errorf("updating config: %w", err)
			}

			printer.Success("Installed %s v%s", ref.FullName(), trimV(installedVersion))
			return nil
		},
	}
}

func trimV(v string) string {
	return strings.TrimPrefix(v, "v")
}
