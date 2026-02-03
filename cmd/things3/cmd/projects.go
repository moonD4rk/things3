package cmd

import (
	"github.com/spf13/cobra"

	"github.com/moond4rk/things3"
)

// NewProjectsCmd creates the projects command.
func NewProjectsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "projects",
		Short: "List all incomplete projects",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := things3.NewClient()
			if err != nil {
				return err
			}
			defer client.Close()

			tasks, err := client.Projects(cmd.Context())
			if err != nil {
				return err
			}

			return outputTasks(cmd, tasks)
		},
	}
}
