package cmd

import (
	"github.com/spf13/cobra"

	"github.com/moond4rk/things3"
)

// NewLogbookCmd creates the logbook command.
func NewLogbookCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logbook",
		Short: "List completed and canceled tasks",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := things3.NewClient()
			if err != nil {
				return err
			}
			defer client.Close()

			tasks, err := client.Logbook(cmd.Context())
			if err != nil {
				return err
			}

			return outputTasks(cmd, tasks)
		},
	}
}
