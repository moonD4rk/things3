package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/moond4rk/things3"
)

// NewProjectCmd creates the project command for viewing a single project.
func NewProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project <identifier>",
		Short: "Show a project by UUID prefix, title keyword, or search query",
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

			var q things3.ProjectQueryBuilder
			switch {
			case byTitle:
				q = client.Projects().WithTitle(identifier).Status().Any()
			case bySearch:
				q = client.Projects().Search(identifier).Status().Any()
			default:
				q = client.Projects().WithUUIDPrefix(identifier).Status().Any()
			}

			projects, err := q.All(cmd.Context())
			if err != nil {
				return err
			}

			if len(projects) == 0 {
				return fmt.Errorf("no project matches %q", identifier)
			}
			if len(projects) == 1 {
				todos, err := client.Todos().
					InProject(projects[0].UUID).
					Status().Incomplete().
					All(cmd.Context())
				if err != nil {
					return err
				}
				return outputProjectDetail(cmd, &projects[0], todos)
			}
			return outputProjects(cmd, projects)
		},
	}

	cmd.Flags().BoolP("title", "t", false, "match by title keyword")
	cmd.Flags().BoolP("search", "s", false, "search in title, notes, and area")

	return cmd
}
