package cmd

import (
	"github.com/spf13/cobra"

	"github.com/moond4rk/things3"
)

// NewSearchCmd creates the search command.
func NewSearchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search for tasks by title or UUID prefix",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := things3.NewClient()
			if err != nil {
				return err
			}
			defer client.Close()

			query := args[0]
			byUUID, _ := cmd.Flags().GetBool("uuid")

			var tasks []things3.Task
			if byUUID {
				// Search by UUID prefix
				tasks, err = client.Tasks().
					WithUUIDPrefix(query).
					Status().Any().
					All(cmd.Context())
			} else {
				// Search by title/notes
				tasks, err = client.Search(cmd.Context(), query)
			}
			if err != nil {
				return err
			}

			return outputTasks(cmd, tasks)
		},
	}

	cmd.Flags().BoolP("uuid", "u", false, "search by UUID prefix instead of title")

	return cmd
}
