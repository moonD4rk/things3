package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/moond4rk/things3"
	"github.com/moond4rk/things3/cmd/things3/internal/resolve"
)

const flagArea = "area"

// Values of the projects --status flag.
const (
	statusIncomplete = "incomplete"
	statusCompleted  = "completed"
	statusCanceled   = "canceled"
	statusAny        = "any"
)

func newProjectsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "projects",
		Short:   "List projects",
		GroupID: groupCollections,
		Example: "  things3 projects\n  things3 projects --area Work\n  things3 projects --status completed --sort date --desc",
		Args:    cobra.NoArgs,
		RunE:    withClient(runProjects),
	}
	cmd.Flags().String(flagArea, "", "filter by area (name, prefix, or UUID)")
	cmd.Flags().String(flagStatus, statusIncomplete, "which projects: incomplete, completed, canceled, or any")
	return cmd
}

func runProjects(cmd *cobra.Command, _ []string, client *things3.Client) error {
	var q things3.ProjectQueryBuilder
	switch status, _ := cmd.Flags().GetString(flagStatus); status {
	case statusIncomplete:
		q = client.Projects().Status().Incomplete()
	case statusCompleted:
		q = client.Projects().Status().Completed()
	case statusCanceled:
		q = client.Projects().Status().Canceled()
	case statusAny:
		q = client.Projects().Status().Any()
	default:
		return fmt.Errorf("invalid --%s %q: use incomplete, completed, canceled, or any", flagStatus, status)
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
	return outputItems(cmd, projectItems(projects), nil, defaultRow)
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
