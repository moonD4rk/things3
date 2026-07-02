package cmd

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	"github.com/moond4rk/things3"
	"github.com/moond4rk/things3/cmd/things3/internal/resolve"
)

// viewLists maps view names to their built-in Things list IDs.
var viewLists = map[string]things3.ListID{
	nameInbox:     things3.ListInbox,
	nameToday:     things3.ListToday,
	nameUpcoming:  things3.ListUpcoming,
	nameAnytime:   things3.ListAnytime,
	nameSomeday:   things3.ListSomeday,
	nameLogbook:   things3.ListLogbook,
	nameDeadlines: things3.ListDeadlines,
}

func newOpenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "open [query|view]",
		Short:   "Reveal an item or built-in list in Things.app",
		GroupID: groupActions,
		Example: "  things3 open\n  things3 open today\n  things3 open \"Buy milk\"",
		Args:    cobra.MaximumNArgs(1),
		RunE:    withClient(runOpen),
	}
	cmd.Flags().Bool(flagDryRun, false, "print the things:/// URL without executing it")
	return cmd
}

func runOpen(cmd *cobra.Command, args []string, client *things3.Client) error {
	nav := client.ShowBuilder()
	result := writeResult{Action: actionOpen, Verified: true, Message: nameToday}

	switch {
	case len(args) == 0:
		nav = nav.List(things3.ListToday)
	default:
		if list, ok := viewLists[strings.ToLower(args[0])]; ok {
			nav = nav.List(list)
			result.Message = strings.ToLower(args[0])
		} else {
			match, err := resolve.ResolveOne(cmd.Context(), client, args[0])
			if err != nil {
				return fromResolveError(err)
			}
			nav = nav.ID(match.UUID())
			result.Message = match.Title()
			result.UUID = match.UUID()
			result.Type = string(match.Kind)
		}
	}

	return runWrite(cmd, actionOpen, nav, func(context.Context) writeResult {
		return result
	})
}
