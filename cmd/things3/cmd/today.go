package cmd

import (
	"github.com/spf13/cobra"

	"github.com/moond4rk/things3"
)

// NewTodayCmd creates the today command.
func NewTodayCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "today",
		Short: "List tasks scheduled for today",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := things3.NewClient()
			if err != nil {
				return err
			}
			defer client.Close()

			tasks, err := client.Today(cmd.Context())
			if err != nil {
				return err
			}

			return outputTasks(cmd, tasks)
		},
	}
}
