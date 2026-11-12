package cmd

import (
	"github.com/spf13/cobra"

	"github.com/moond4rk/things3"
	"github.com/moond4rk/things3/cmd/things3/internal/change"
	"github.com/moond4rk/things3/cmd/things3/internal/resolve"
)

func newScheduleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "schedule <query> <when>",
		Short:   "Schedule a todo or project (the app's When)",
		GroupID: groupActions,
		Example: "  things3 schedule \"Buy milk\" today\n  things3 schedule a1b2c3d4 evening\n" +
			"  things3 schedule \"Report\" 2026-08-01 --reminder 09:00",
		Args: cobra.ExactArgs(2),
		RunE: withClient(runSchedule),
	}
	cmd.Flags().String(flagReminder, "", "also set a reminder time (HH:MM); needs a day, not anytime or someday")
	addWriteFlags(cmd)
	return cmd
}

func runSchedule(cmd *cobra.Command, args []string, client *things3.Client) error {
	schedule := change.Schedule{When: args[1]}
	if cmd.Flags().Changed(flagReminder) {
		clock, err := change.ParseClock(mustString(cmd.Flags(), flagReminder))
		if err != nil {
			return err
		}
		schedule.Reminder = &clock
	}

	match, err := resolve.ResolveTarget(cmd.Context(), client, args[0], resolve.PreferOpen)
	if err != nil {
		return fromResolveError(err)
	}
	builder, err := schedule.Apply(client, match)
	if err != nil {
		return err
	}
	return runWrite(cmd, "schedule", builder, modifiedVerifier("schedule", match, matchModifiedAt(match), client))
}
