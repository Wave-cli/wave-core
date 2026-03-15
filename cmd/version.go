package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wave-cli/wave-core/internal/version"
)

// NewVersionCmd creates the 'wave version' command.
func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print wave version information",
		Run: func(cmd *cobra.Command, args []string) {
			info := version.Get()
			cmd.Println(info.Full())
		},
	}
}
