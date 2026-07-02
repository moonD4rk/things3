package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// Command group IDs, displayed in help in this order.
const (
	groupViews       = "views"
	groupCollections = "collections"
	groupLookup      = "lookup"
	groupActions     = "actions"
)

// NewRootCmd builds the things3 root command with all subcommands registered.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "things3",
		Short: "Query and control Things 3 from the terminal",
		Long: `things3 reads the Things 3 database directly and writes through the Things
URL scheme, mirroring the app's sidebar views and interaction verbs.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetOut(os.Stdout)
	root.SetErr(os.Stderr)

	addGlobalFlags(root)

	root.AddGroup(
		&cobra.Group{ID: groupViews, Title: "Views:"},
		&cobra.Group{ID: groupCollections, Title: "Collections:"},
		&cobra.Group{ID: groupLookup, Title: "Lookup:"},
		&cobra.Group{ID: groupActions, Title: "Actions:"},
	)

	registerCommands(root)

	return root
}

// addGlobalFlags registers the persistent flag surface shared by every command:
// the mutually exclusive output-format switches, the display limit, the database
// override, and the list-shaping flags. The list flags are honored by the list
// commands and accepted but ignored elsewhere, giving one uniform flag surface.
func addGlobalFlags(root *cobra.Command) {
	pf := root.PersistentFlags()
	pf.Bool(flagText, false, "output as plain text (default)")
	pf.BoolP(flagJSON, "j", false, "output as JSON")
	pf.BoolP(flagYAML, "y", false, "output as YAML")
	pf.IntP(flagLimit, "n", 0, "max items to display (0 = unlimited)")
	pf.String(flagDB, "", "Things database path (overrides THINGSDB)")
	pf.Var(newPageValue(), flagPage, "page number, 1-based (list commands)")
	pf.Bool(flagAll, false, "show all items without pagination (list commands)")
	pf.Var(newSortValue(), flagSort, "sort by: date, created, modified, title (list commands)")
	pf.Bool(flagDesc, false, "reverse the --sort order (list commands)")
	pf.String(flagTag, "", "keep only items carrying this tag, case-insensitive (list commands)")
	root.MarkFlagsMutuallyExclusive(flagText, flagJSON, flagYAML)
}

// registerCommands attaches every subcommand to the root in help order.
func registerCommands(root *cobra.Command) {
	root.AddCommand(
		newTodayCmd(),
		newInboxCmd(),
		newUpcomingCmd(),
		newAnytimeCmd(),
		newSomedayCmd(),
		newLogbookCmd(),
		newDeadlinesCmd(),
		newTrashCmd(),
		newProjectsCmd(),
		newAreasCmd(),
		newTagsCmd(),
		newShowCmd(),
		newSearchCmd(),
		newAddCmd(),
		newDoneCmd(),
		newCancelCmd(),
		newScheduleCmd(),
		newMoveCmd(),
		newEditCmd(),
		newOpenCmd(),
		NewVersionCmd(),
	)
}
