package cmd

import (
	"github.com/spf13/cobra"

	"github.com/moond4rk/things3"
)

// NewSearchCmd creates the search command.
func NewSearchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search for todos by title or UUID prefix",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := things3.NewClient()
			if err != nil {
				return err
			}
			defer client.Close()

			query := args[0]
			byUUID, _ := cmd.Flags().GetBool("uuid")

			var q things3.TodoQueryBuilder
			if byUUID {
				q = client.Todos().WithUUIDPrefix(query).Status().Any()
			} else {
				q = client.Todos().Search(query).Status().Any()
			}

			todos, err := q.All(cmd.Context())
			if err != nil {
				return err
			}
			return outputTodos(cmd, todos)
		},
	}

	cmd.Flags().BoolP("uuid", "u", false, "search by UUID prefix instead of title")

	return cmd
}
