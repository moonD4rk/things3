package cmd

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/moond4rk/things3"
	"github.com/moond4rk/things3/cmd/things3/internal/resolve"
)

func newEditCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "edit <query>",
		Short:   "Edit a todo or project's attributes",
		GroupID: groupActions,
		Example: `  things3 edit "Buy milk" --title "Buy oat milk"
  things3 edit a1b2c3d4 --add-tags urgent
  things3 edit "Report" --deadline 2026-08-01`,
		Args: cobra.ExactArgs(1),
		RunE: withClient(runEdit),
	}
	cmd.Flags().String(flagTitle, "", "set the title")
	cmd.Flags().String(flagNotes, "", "replace the notes")
	cmd.Flags().String(flagAppendNotes, "", "append to the notes")
	cmd.Flags().String(flagDeadline, "", "set the deadline (YYYY-MM-DD)")
	cmd.Flags().Bool(flagClearDeadline, false, "clear the deadline")
	cmd.Flags().String(flagTags, "", "replace tags (comma-separated)")
	cmd.Flags().String(flagAddTags, "", "add tags (comma-separated)")
	cmd.MarkFlagsMutuallyExclusive(flagDeadline, flagClearDeadline)
	addWriteFlags(cmd)
	return cmd
}

func runEdit(cmd *cobra.Command, args []string, client *things3.Client) error {
	f := cmd.Flags()
	if !anyChanged(f, flagTitle, flagNotes, flagAppendNotes, flagDeadline, flagClearDeadline, flagTags, flagAddTags) {
		return errors.New("nothing to edit; pass at least one flag")
	}

	ctx := cmd.Context()
	match, err := resolve.ResolveOne(ctx, client, args[0])
	if err != nil {
		return fromResolveError(err)
	}
	baseline := matchModifiedAt(match)

	var builder urlBuilder
	if match.Kind == resolve.KindProject {
		u, flagErr := applyEditFlags(client.UpdateProject(match.UUID()), f)
		if flagErr != nil {
			return flagErr
		}
		builder = u
	} else {
		u, flagErr := applyEditFlags(client.UpdateTodo(match.UUID()), f)
		if flagErr != nil {
			return flagErr
		}
		builder = u
	}
	return runWrite(cmd, "edit", builder, modifiedVerifier("edit", match, baseline, client))
}
