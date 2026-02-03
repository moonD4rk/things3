package cmd

import (
	"github.com/spf13/cobra"

	"github.com/moond4rk/things3"
)

// NewInboxCmd creates the inbox command.
func NewInboxCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "inbox",
		Short: "List tasks in the Inbox",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := things3.NewClient()
			if err != nil {
				return err
			}
			defer client.Close()

			tasks, err := client.Inbox(cmd.Context())
			if err != nil {
				return err
			}

			return outputTasks(cmd, tasks)
		},
	}
}
