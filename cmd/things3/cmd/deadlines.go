package cmd

import (
	"github.com/spf13/cobra"

	"github.com/moond4rk/things3"
)

// NewDeadlinesCmd creates the deadlines command.
func NewDeadlinesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "deadlines",
		Short: "List tasks with deadlines",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := things3.NewClient()
			if err != nil {
				return err
			}
			defer client.Close()

			tasks, err := client.Deadlines(cmd.Context())
			if err != nil {
				return err
			}

			return outputTasks(cmd, tasks)
		},
	}
}
