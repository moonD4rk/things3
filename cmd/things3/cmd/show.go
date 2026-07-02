package cmd

import (
	"github.com/spf13/cobra"

	"github.com/moond4rk/things3"
	"github.com/moond4rk/things3/cmd/things3/internal/resolve"
)

func newShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "show <query>",
		Short:   "Show an item by UUID, prefix, or title (Quick Find)",
		GroupID: groupLookup,
		Example: "  things3 show 3x1QqJqf\n  things3 show \"Write report\"\n  things3 show meeting",
		Args:    cobra.ExactArgs(1),
		RunE:    withClient(runShow),
	}
}

func runShow(cmd *cobra.Command, args []string, client *things3.Client) error {
	ctx := cmd.Context()
	matches, err := resolve.Resolve(ctx, client, args[0])
	if err != nil {
		return err
	}

	switch len(matches) {
	case 0:
		return &NotFoundError{Query: args[0]}
	case 1:
		return showOne(cmd, client, matches[0])
	default:
		items := make([]mixedItem, len(matches))
		for i, m := range matches {
			if m.Kind == resolve.KindProject {
				items[i] = projectMixed(m.Project)
			} else {
				items[i] = todoMixed(m.Todo)
			}
		}
		return outputMixed(cmd, items)
	}
}

func showOne(cmd *cobra.Command, client *things3.Client, m resolve.Match) error {
	ctx := cmd.Context()
	if m.Kind == resolve.KindProject {
		todos, err := client.Todos().InProject(m.UUID()).Status().Incomplete().All(ctx)
		if err != nil {
			return err
		}
		return outputProjectDetail(cmd, m.Project, todos)
	}
	// Re-fetch to load the checklist, which Resolve does not populate.
	todo, err := client.Todos().WithUUID(m.UUID()).Status().Any().IncludeChecklist().First(ctx)
	if err != nil {
		return err
	}
	return outputTodoDetail(cmd, todo)
}
