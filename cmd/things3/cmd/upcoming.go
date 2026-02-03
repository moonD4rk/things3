package cmd

import (
	"github.com/spf13/cobra"

	"github.com/moond4rk/things3"
)

// NewUpcomingCmd creates the upcoming command.
func NewUpcomingCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "upcoming",
		Short: "List tasks scheduled for future dates",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := things3.NewClient()
			if err != nil {
				return err
			}
			defer client.Close()

			tasks, err := client.Upcoming(cmd.Context())
			if err != nil {
				return err
			}

			return outputTasks(cmd, tasks)
		},
	}
}
