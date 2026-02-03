package cmd

import (
	"github.com/spf13/cobra"

	"github.com/moond4rk/things3"
)

// NewSomedayCmd creates the someday command.
func NewSomedayCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "someday",
		Short: "List tasks in the Someday list",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := things3.NewClient()
			if err != nil {
				return err
			}
			defer client.Close()

			tasks, err := client.Someday(cmd.Context())
			if err != nil {
				return err
			}

			return outputTasks(cmd, tasks)
		},
	}
}
