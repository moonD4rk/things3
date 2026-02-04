package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

// NewVersionCmd creates the version command.
func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, _ []string) {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "things3 %s\n", version)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  commit: %s\n", commit)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  built:  %s\n", buildDate)
		},
	}
}
