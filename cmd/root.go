// Package cmd implements the wave CLI commands using Cobra.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wave-cli/wave-core/internal/bootstrap"
	"github.com/wave-cli/wave-core/internal/config"
	"github.com/wave-cli/wave-core/internal/errors"
	"github.com/wave-cli/wave-core/internal/pluginmgmt"
	"github.com/wave-cli/wave-core/internal/runner"
	"github.com/wave-cli/wave-core/internal/ui"
)

var (
	printer   *ui.Printer
	globalCfg *config.GlobalConfig
)

// NewRootCmd creates and returns the root cobra command.
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "wave",
		Short: "wave - modular CLI orchestrator",
		Long:  "wave is a plugin-based CLI orchestrator. Run plugins via: wave <plugin> [args...]",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Set up printer based on flags
			level := ui.LevelNormal
			if viper.GetBool("quiet") {
				level = ui.LevelQuiet
			} else if viper.GetBool("debug") {
				level = ui.LevelDebug
			} else if viper.GetBool("verbose") {
				level = ui.LevelVerbose
			}
			printer = ui.NewPrinter(os.Stderr, level)

			// Bootstrap wave home directory
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("getting home directory: %w", err)
			}

			gc, err := bootstrap.Ensure(homeDir)
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			globalCfg = gc

			return nil
		},
		// When an unknown command is given, try to run it as a plugin
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	// Persistent flags bound via Viper
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().Bool("debug", false, "Debug output")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Suppress non-error output")
	rootCmd.PersistentFlags().Bool("manual", false, "Run plugin without Wavefile config (manual mode)")

	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
	viper.BindPFlag("manual", rootCmd.PersistentFlags().Lookup("manual"))

	// Register built-in commands
	rootCmd.AddCommand(NewVersionCmd())
	rootCmd.AddCommand(NewInitCmd())
	rootCmd.AddCommand(NewConfigCmd())
	rootCmd.AddCommand(NewInstallCmd())
	rootCmd.AddCommand(NewUninstallCmd())
	rootCmd.AddCommand(NewListCmd())
	rootCmd.AddCommand(NewRestoreCmd())

	return rootCmd
}

// RegisterPlugins dynamically adds installed plugins as sub-commands.
func RegisterPlugins(rootCmd *cobra.Command, gc *config.GlobalConfig) {
	if gc == nil || gc.Plugins == nil {
		return
	}

	homeDir, _ := os.UserHomeDir()
	pluginsDir := filepath.Join(homeDir, ".wave", "plugins")
	reg := pluginmgmt.NewRegistry(pluginsDir)

	for fullName := range gc.Plugins {
		ref, err := pluginmgmt.ParsePluginRef(fullName)
		if err != nil {
			continue
		}

		shortName := ref.Name
		fn := fullName // capture for closure

		// Read Waveplugin for description
		description := "wave plugin"
		if wp, err := reg.ReadWaveplugin(fullName); err == nil {
			description = wp.Plugin.Description
		}

		cmd := &cobra.Command{
			Use:                shortName,
			Short:              description,
			DisableFlagParsing: true,
			RunE: func(cmd *cobra.Command, args []string) error {
				return runPlugin(fn, args, gc, pluginsDir, reg)
			},
		}
		rootCmd.AddCommand(cmd)
	}
}

// runPlugin executes an installed plugin.
func runPlugin(fullName string, args []string, gc *config.GlobalConfig, pluginsDir string, reg *pluginmgmt.Registry) error {
	ref, _ := pluginmgmt.ParsePluginRef(fullName)

	// Resolve binary
	binPath, err := reg.ResolveBinary(fullName)
	if err != nil {
		return fmt.Errorf("plugin %q: %w", ref.Name, err)
	}

	// Get plugin version
	version := gc.Plugins[fullName]

	// Try to find project root and Wavefile (unless in manual mode)
	var section map[string]any
	projectRoot := ""
	cwd, _ := os.Getwd()

	if !viper.GetBool("manual") {
		if wfPath, err := config.DiscoverWavefile(cwd); err == nil {
			projectRoot = filepath.Dir(wfPath)
			if wf, err := config.ParseWavefile(wfPath); err == nil {
				section = wf.Sections[ref.Name]
			}
		}
	}

	// Execute with streaming
	result, err := runner.StreamExecute(binPath, args, section, ref.Name, version, projectRoot, os.Stdout, os.Stderr)
	if err != nil {
		return err
	}

	// Handle structured errors
	if result.PluginError != nil {
		pe := result.PluginError

		// Log error
		logsDir := gc.Core.LogsDir
		errors.LogError(logsDir, ref.Name, pe, args)

		// Format and display
		logFile := filepath.Join(logsDir, "daily.log")
		printer.Error("%s", errors.FormatError(ref.Name, version, pe, logFile))
	}

	if result.ExitCode != 0 {
		os.Exit(result.ExitCode)
	}

	return nil
}

// Execute runs the root command.
func Execute() {
	rootCmd := NewRootCmd()

	// Bootstrap first to register plugins
	homeDir, _ := os.UserHomeDir()
	if gc, err := bootstrap.Ensure(homeDir); err == nil {
		RegisterPlugins(rootCmd, gc)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
