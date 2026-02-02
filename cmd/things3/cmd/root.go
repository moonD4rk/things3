package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "things3",
	Short: "CLI for Things 3 task management",
	Long: `things3 is a command-line interface for interacting with Things 3.

It provides read-only access to your Things 3 database and allows you to
query tasks, projects, areas, and tags from the command line.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)
}

func exitWithError(err error) {
	fmt.Fprintln(os.Stderr, "Error:", err)
	os.Exit(1)
}
