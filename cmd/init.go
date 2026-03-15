package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// NewInitCmd creates the 'wave init' command.
func NewInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a new Wavefile in the current directory",
		Long:  "Creates a Wavefile in the current directory with default project metadata.",
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

			// Use directory name as project name
			projectName := filepath.Base(cwd)

			content := fmt.Sprintf(`[project]
name = %q
version = "0.1.0"
owner = ""
category = ""
tags = []
`, projectName)

			if err := os.WriteFile(wavefilePath, []byte(content), 0644); err != nil {
				return fmt.Errorf("writing Wavefile: %w", err)
			}

			printer.Success("Created Wavefile in %s", cwd)
			return nil
		},
	}
}
