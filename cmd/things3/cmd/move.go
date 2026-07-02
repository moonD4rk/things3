package cmd

import (
	"context"
	"errors"
	"strings"

	"github.com/spf13/cobra"

	"github.com/moond4rk/things3"
	"github.com/moond4rk/things3/cmd/things3/internal/resolve"
)

func newMoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "move <query>",
		Short:   "Move a todo or project to a project or area (the app's Move)",
		GroupID: groupActions,
		Example: "  things3 move \"Buy milk\" --to Groceries\n  things3 move a1b2c3d4 --to Work",
		Args:    cobra.ExactArgs(1),
		RunE:    withClient(runMove),
	}
	cmd.Flags().String(flagTo, "", "destination project or area (required)")
	_ = cmd.MarkFlagRequired(flagTo)
	addWriteFlags(cmd)
	return cmd
}

func runMove(cmd *cobra.Command, args []string, client *things3.Client) error {
	dest, _ := cmd.Flags().GetString(flagTo)
	switch {
	case strings.EqualFold(dest, nameInbox):
		return errors.New("the Things URL scheme cannot move items to Inbox; use the Things app")
	case isWhenKeyword(dest):
		return errors.New(`"--to" takes a project or area; use "things3 schedule" for dates`)
	}

	ctx := cmd.Context()
	match, err := resolve.ResolveOne(ctx, client, args[0])
	if err != nil {
		return fromResolveError(err)
	}
	baseline := matchModifiedAt(match)

	builder, err := moveBuilder(ctx, client, match, dest)
	if err != nil {
		return err
	}
	return runWrite(cmd, "move", builder, modifiedVerifier("move", match, baseline, client))
}

func moveBuilder(ctx context.Context, client *things3.Client, match resolve.Match, dest string) (urlBuilder, error) {
	if match.Kind == resolve.KindProject {
		// Projects can only move to areas.
		area, err := resolve.Area(ctx, client, dest)
		if err != nil {
			if isNotFound(err) {
				if _, perr := resolve.Project(ctx, client, dest); perr == nil {
					return nil, errors.New("a project can only move to an area, not another project")
				}
			}
			return nil, fromResolveError(err)
		}
		return client.UpdateProject(match.UUID()).AreaID(area.UUID), nil
	}

	// Todo: try a project destination first, then an area.
	project, err := resolve.Project(ctx, client, dest)
	if err == nil {
		return client.UpdateTodo(match.UUID()).ListID(project.UUID), nil
	}
	if !isNotFound(err) {
		return nil, fromResolveError(err)
	}
	area, err := resolve.Area(ctx, client, dest)
	if err != nil {
		return nil, fromResolveError(err)
	}
	return client.UpdateTodo(match.UUID()).ListID(area.UUID), nil
}
