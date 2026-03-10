package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/moond4rk/things3"
)

// NewTodoCmd creates the todo command for viewing a single todo.
func NewTodoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "todo <identifier>",
		Short: "Show a todo by UUID prefix, title keyword, or search query",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := things3.NewClient()
			if err != nil {
				return err
			}
			defer client.Close()

			identifier := args[0]
			byTitle, _ := cmd.Flags().GetBool("title")
			bySearch, _ := cmd.Flags().GetBool("search")

			if byTitle && bySearch {
				return fmt.Errorf("--title and --search are mutually exclusive")
			}

			var q things3.TodoQueryBuilder
			switch {
			case byTitle:
				q = client.Todos().WithTitle(identifier).Status().Any()
			case bySearch:
				q = client.Todos().Search(identifier).Status().Any()
			default:
				q = client.Todos().WithUUIDPrefix(identifier).Status().Any()
			}

			todos, err := q.IncludeChecklist().All(cmd.Context())
			if err != nil {
				return err
			}

			if len(todos) == 1 {
				return outputTodoDetail(cmd, &todos[0])
			}
			return outputTodos(cmd, todos)
		},
	}

	cmd.Flags().BoolP("title", "t", false, "match by title keyword")
	cmd.Flags().BoolP("search", "s", false, "search in title, notes, and area")

	return cmd
}
