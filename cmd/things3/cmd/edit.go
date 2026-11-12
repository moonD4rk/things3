package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/moond4rk/things3"
	"github.com/moond4rk/things3/cmd/things3/internal/change"
	"github.com/moond4rk/things3/cmd/things3/internal/resolve"
)

func newEditCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "edit <query>",
		Short:   "Edit a todo or project's attributes",
		GroupID: groupActions,
		Example: `  things3 edit "Buy milk" --title "Buy oat milk"
  things3 edit a1b2c3d4 --add-tags urgent --reminder 09:00
  things3 edit "Report" --deadline 2026-08-01
  things3 edit "Pack" --append-checklist Passport --append-checklist Charger`,
		Args: cobra.ExactArgs(1),
		RunE: withClient(runEdit),
	}
	f := cmd.Flags()
	f.String(flagTitle, "", "set the title")
	f.String(flagNotes, "", "replace the notes")
	f.Bool(flagClearNotes, false, "clear the notes")
	f.String(flagPrependNotes, "", "prepend to the notes")
	f.String(flagAppendNotes, "", "append to the notes")
	f.String(flagDeadline, "", "set the deadline (YYYY-MM-DD)")
	f.Bool(flagClearDeadline, false, "clear the deadline")
	f.String(flagReminder, "", "set the reminder time (HH:MM), keeping the scheduled day")
	f.String(flagTags, "", "replace tags (comma-separated)")
	f.Bool(flagClearTags, false, "remove every tag")
	f.String(flagAddTags, "", "add tags (comma-separated)")
	f.StringArray(flagChecklist, nil, "replace the checklist with this item (repeatable; todos only)")
	f.Bool(flagClearChecklist, false, "remove the checklist")
	f.StringArray(flagPrependChecklist, nil, "add a checklist item at the top (repeatable)")
	f.StringArray(flagAppendChecklist, nil, "add a checklist item at the bottom (repeatable)")
	addWriteFlags(cmd)
	return cmd
}

func runEdit(cmd *cobra.Command, args []string, client *things3.Client) error {
	edit, err := editFromFlags(cmd.Flags())
	if err != nil {
		return err
	}
	if edit.Empty() {
		return &change.InputError{Message: "nothing to edit; pass at least one flag"}
	}

	match, err := resolve.ResolveTarget(cmd.Context(), client, args[0], resolve.PreferOpen)
	if err != nil {
		return fromResolveError(err)
	}
	builder, err := edit.Apply(client, match)
	if err != nil {
		return err
	}
	return runWrite(cmd, "edit", builder, modifiedVerifier("edit", match, matchModifiedAt(match), client))
}

// editFromFlags reads the edit flags that were set into a change.
func editFromFlags(f *pflag.FlagSet) (*change.Edit, error) {
	edit := &change.Edit{}
	if f.Changed(flagTitle) {
		edit.Title = new(mustString(f, flagTitle))
	}
	if f.Changed(flagNotes) {
		edit.Notes = new(mustString(f, flagNotes))
	}
	edit.ClearNotes, _ = f.GetBool(flagClearNotes)
	edit.PrependNotes, _ = f.GetString(flagPrependNotes)
	edit.AppendNotes, _ = f.GetString(flagAppendNotes)

	if f.Changed(flagDeadline) {
		d, err := change.ParseDate(flagDeadline, mustString(f, flagDeadline))
		if err != nil {
			return nil, err
		}
		edit.Deadline = &d
	}
	edit.ClearDeadline, _ = f.GetBool(flagClearDeadline)

	if f.Changed(flagReminder) {
		clock, err := change.ParseClock(mustString(f, flagReminder))
		if err != nil {
			return nil, err
		}
		edit.Reminder = &clock
	}

	edit.Tags = splitTags(mustString(f, flagTags))
	edit.ClearTags, _ = f.GetBool(flagClearTags)
	edit.AddTags = splitTags(mustString(f, flagAddTags))

	edit.Checklist, _ = f.GetStringArray(flagChecklist)
	edit.ClearChecklist, _ = f.GetBool(flagClearChecklist)
	edit.PrependChecklist, _ = f.GetStringArray(flagPrependChecklist)
	edit.AppendChecklist, _ = f.GetStringArray(flagAppendChecklist)
	return edit, nil
}

func mustString(f *pflag.FlagSet, name string) string {
	v, _ := f.GetString(name)
	return v
}
