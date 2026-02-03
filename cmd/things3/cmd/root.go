package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// Flag names for consistent access across commands.
const (
	flagLimit = "limit"
	flagJSON  = "json"
	flagYAML  = "yaml"
)

// NewRootCmd creates the root command for things3 CLI.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "things3",
		Short: "CLI for Things 3 task management",
		Long: `things3 is a command-line interface for interacting with Things 3.

It provides read-only access to your Things 3 database and allows you to
query tasks, projects, areas, and tags from the command line.`,
	}
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)

	// Global flags
	cmd.PersistentFlags().IntP(flagLimit, "n", 0, "max items to display (0 for unlimited)")
	cmd.PersistentFlags().BoolP(flagJSON, "j", false, "output as JSON")
	cmd.PersistentFlags().BoolP(flagYAML, "y", false, "output as YAML")

	// Register subcommands
	cmd.AddCommand(NewVersionCmd())
	cmd.AddCommand(NewInboxCmd())
	cmd.AddCommand(NewTodayCmd())
	cmd.AddCommand(NewUpcomingCmd())
	cmd.AddCommand(NewAnytimeCmd())
	cmd.AddCommand(NewSomedayCmd())
	cmd.AddCommand(NewLogbookCmd())
	cmd.AddCommand(NewDeadlinesCmd())
	cmd.AddCommand(NewProjectsCmd())
	cmd.AddCommand(NewAreasCmd())
	cmd.AddCommand(NewTagsCmd())
	cmd.AddCommand(NewSearchCmd())

	return cmd
}
