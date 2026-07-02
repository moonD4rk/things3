package cmd

import (
	"github.com/spf13/cobra"

	"github.com/moond4rk/things3"
	"github.com/moond4rk/things3/cmd/things3/internal/resolve"
)

const flagArea = "area"

func newProjectsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "projects",
		Short:   "List projects",
		GroupID: groupCollections,
		Example: "  things3 projects\n  things3 projects --area Work\n  things3 projects --all --sort title",
		Args:    cobra.NoArgs,
		RunE:    withClient(runProjects),
	}
	cmd.Flags().String(flagArea, "", "filter by area (name, prefix, or UUID)")
	return cmd
}

func runProjects(cmd *cobra.Command, _ []string, client *things3.Client) error {
	// --all doubles as the shared "no pagination" flag and the "include closed
	// projects" status filter.
	q := client.Projects().Status().Incomplete()
	if all, _ := cmd.Flags().GetBool(flagAll); all {
		q = client.Projects().Status().Any()
	}
	if area, _ := cmd.Flags().GetString(flagArea); area != "" {
		a, err := resolve.Area(cmd.Context(), client, area)
		if err != nil {
			return fromResolveError(err)
		}
		q = q.InArea(a.UUID)
	}
	projects, err := q.All(cmd.Context())
	if err != nil {
		return err
	}
	return outputProjectList(cmd, projects)
}

func newAreasCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "areas",
		Short:   "List all areas",
		GroupID: groupCollections,
		Example: "  things3 areas\n  things3 areas --json",
		Args:    cobra.NoArgs,
		RunE:    withClient(runAreas),
	}
}

func runAreas(cmd *cobra.Command, _ []string, client *things3.Client) error {
	areas, err := client.Areas().All(cmd.Context())
	if err != nil {
		return err
	}
	return outputAreas(cmd, areas)
}

func newTagsCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "tags",
		Short:   "List all tags",
		GroupID: groupCollections,
		Example: "  things3 tags\n  things3 tags --json",
		Args:    cobra.NoArgs,
		RunE:    withClient(runTags),
	}
}

func runTags(cmd *cobra.Command, _ []string, client *things3.Client) error {
	tags, err := client.Tags().All(cmd.Context())
	if err != nil {
		return err
	}
	return outputTags(cmd, tags)
}
