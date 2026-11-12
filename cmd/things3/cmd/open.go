package cmd

import (
	"context"
	"errors"
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

const flagSearch = "search"

func newOpenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "open [query|view]",
		Short:   "Reveal an item, built-in list, or search in Things.app",
		GroupID: groupActions,
		Example: "  things3 open\n  things3 open today\n  things3 open \"Buy milk\"\n  things3 open --search invoice",
		Args:    cobra.MaximumNArgs(1),
		RunE:    withClient(runOpen),
	}
	cmd.Flags().String(flagSearch, "", "open the app's search for this text instead of an item or list")
	cmd.Flags().Bool(flagDryRun, false, "print the things:/// URL without executing it")
	return cmd
}

func runOpen(cmd *cobra.Command, args []string, client *things3.Client) error {
	result := writeResult{Action: actionOpen, Verified: true, Message: nameToday}
	nav, err := openTarget(cmd, args, client, &result)
	if err != nil {
		return err
	}
	return runWrite(cmd, actionOpen, nav, func(context.Context) writeResult {
		return result
	})
}

// openTarget builds the navigation URL and fills in what result reports.
func openTarget(cmd *cobra.Command, args []string, client *things3.Client, result *writeResult) (urlBuilder, error) {
	search, _ := cmd.Flags().GetString(flagSearch)
	switch {
	case search != "" && len(args) > 0:
		return nil, errors.New("--search opens the app's search; it cannot be combined with an item or view")
	case search != "":
		result.Message = "search: " + search
		return client.SearchBuilder(search), nil
	case len(args) == 0:
		return client.ShowBuilder().List(things3.ListToday), nil
	}

	if list, ok := viewLists[strings.ToLower(args[0])]; ok {
		result.Message = strings.ToLower(args[0])
		return client.ShowBuilder().List(list), nil
	}
	match, err := resolve.ResolveOne(cmd.Context(), client, args[0], resolve.PreferOpen)
	if err != nil {
		return nil, fromResolveError(err)
	}
	result.Message = match.Title()
	result.UUID = match.UUID()
	result.Type = string(match.Kind)
	return client.ShowBuilder().ID(match.UUID()), nil
}
