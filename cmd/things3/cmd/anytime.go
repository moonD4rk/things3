package cmd

import (
	"github.com/spf13/cobra"

	"github.com/moond4rk/things3"
)

// NewAnytimeCmd creates the anytime command.
func NewAnytimeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "anytime",
		Short: "List tasks in the Anytime list",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := things3.NewClient()
			if err != nil {
				return err
			}
			defer client.Close()

			tasks, err := client.Anytime(cmd.Context())
			if err != nil {
				return err
			}

			return outputTasks(cmd, tasks)
		},
	}
}
