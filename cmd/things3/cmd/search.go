package cmd

import (
	"github.com/spf13/cobra"

	"github.com/moond4rk/things3"
)

func newSearchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "search <query>",
		Short:   "Full-text search across todos and projects",
		GroupID: groupLookup,
		Example: "  things3 search meeting\n  things3 search report --json",
		Args:    cobra.ExactArgs(1),
		RunE:    withClient(runSearch),
	}
	return cmd
}

func runSearch(cmd *cobra.Command, args []string, client *things3.Client) error {
	ctx := cmd.Context()
	todos, err := client.Todos().Search(args[0]).Status().Any().All(ctx)
	if err != nil {
		return err
	}
	projects, err := client.Projects().Search(args[0]).Status().Any().All(ctx)
	if err != nil {
		return err
	}
	items := make([]mixedItem, 0, len(todos)+len(projects))
	for i := range todos {
		items = append(items, todoMixed(&todos[i]))
	}
	for i := range projects {
		items = append(items, projectMixed(&projects[i]))
	}
	return outputMixedList(cmd, items)
}
