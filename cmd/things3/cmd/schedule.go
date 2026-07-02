package cmd

import (
	"github.com/spf13/cobra"

	"github.com/moond4rk/things3"
	"github.com/moond4rk/things3/cmd/things3/internal/resolve"
)

func newScheduleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "schedule <query> <when>",
		Short:   "Schedule a todo or project (the app's When)",
		GroupID: groupActions,
		Example: "  things3 schedule \"Buy milk\" today\n  things3 schedule a1b2c3d4 evening\n  things3 schedule \"Report\" 2026-08-01",
		Args:    cobra.ExactArgs(2),
		RunE:    withClient(runSchedule),
	}
	addWriteFlags(cmd)
	return cmd
}

func runSchedule(cmd *cobra.Command, args []string, client *things3.Client) error {
	match, err := resolve.ResolveOne(cmd.Context(), client, args[0])
	if err != nil {
		return fromResolveError(err)
	}
	baseline := matchModifiedAt(match)

	var builder urlBuilder
	if match.Kind == resolve.KindProject {
		u, whenErr := things3.ParseWhen(client.UpdateProject(match.UUID()), args[1])
		if whenErr != nil {
			return whenErr
		}
		builder = u
	} else {
		u, whenErr := things3.ParseWhen(client.UpdateTodo(match.UUID()), args[1])
		if whenErr != nil {
			return whenErr
		}
		builder = u
	}
	return runWrite(cmd, "schedule", builder, modifiedVerifier("schedule", match, baseline, client))
}
