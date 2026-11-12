package cmd

import (
	"fmt"
	"slices"
	"time"

	"github.com/spf13/cobra"

	"github.com/moond4rk/things3"
)

const flagDays = "days"

// daysWindow reads --days, rejecting a negative window rather than ignoring it,
// so the flag has the same contract as the MCP tools' days argument.
func daysWindow(cmd *cobra.Command) (int, error) {
	days, _ := cmd.Flags().GetInt(flagDays)
	if days < 0 {
		return 0, fmt.Errorf("invalid --%s %d: the window must be zero or positive", flagDays, days)
	}
	return days, nil
}

// view is one sidebar list: how to fetch it and how to lay it out in text mode.
type view struct {
	fetch func(cmd *cobra.Command, client *things3.Client) ([]things3.Item, error)
	group func([]things3.Item) []itemGroup
	row   rowOptions
}

// run adapts a view to a command body.
func (v view) run(cmd *cobra.Command, _ []string, client *things3.Client) error {
	items, err := v.fetch(cmd, client)
	if err != nil {
		return err
	}
	return outputItems(cmd, items, v.group, v.row)
}

func newTodayCmd() *cobra.Command {
	v := view{
		fetch: func(cmd *cobra.Command, client *things3.Client) ([]things3.Item, error) {
			return client.Today(cmd.Context())
		},
		group: groupToday,
		row:   defaultRow,
	}
	return &cobra.Command{
		Use:     nameToday,
		Short:   "List the Today list, This Evening last",
		GroupID: groupViews,
		Example: "  things3 today\n  things3 today --sort title\n  things3 today --json",
		Args:    cobra.NoArgs,
		RunE:    withClient(v.run),
	}
}

func newInboxCmd() *cobra.Command {
	v := view{
		fetch: func(cmd *cobra.Command, client *things3.Client) ([]things3.Item, error) {
			return client.Inbox(cmd.Context())
		},
		row: defaultRow,
	}
	return &cobra.Command{
		Use:     nameInbox,
		Short:   "List todos in the Inbox",
		GroupID: groupViews,
		Example: "  things3 inbox\n  things3 inbox --json",
		Args:    cobra.NoArgs,
		RunE:    withClient(v.run),
	}
}

func newUpcomingCmd() *cobra.Command {
	v := view{
		fetch: func(cmd *cobra.Command, client *things3.Client) ([]things3.Item, error) {
			days, err := daysWindow(cmd)
			if err != nil {
				return nil, err
			}
			items, err := client.Upcoming(cmd.Context())
			if err != nil || days == 0 {
				return items, err
			}
			return withinDays(items, things3.Item.UpcomingDate, client.Now(), days), nil
		},
		group: groupByUpcomingDate,
		row:   defaultRow,
	}
	cmd := &cobra.Command{
		Use:     nameUpcoming,
		Short:   "List scheduled items, repeating series, and deadlines by date",
		GroupID: groupViews,
		Example: "  things3 upcoming\n  things3 upcoming --days 7\n  things3 upcoming --tag work",
		Args:    cobra.NoArgs,
		RunE:    withClient(v.run),
	}
	cmd.Flags().Int(flagDays, 0, "limit to the next N days (0 = all)")
	return cmd
}

func newAnytimeCmd() *cobra.Command {
	v := view{
		fetch: func(cmd *cobra.Command, client *things3.Client) ([]things3.Item, error) {
			return client.Anytime(cmd.Context())
		},
		group: groupByContainer,
		// Rows drop the container segment because the view already groups by it.
		row: noContainerRow,
	}
	return &cobra.Command{
		Use:     nameAnytime,
		Short:   "List available todos grouped by project or area",
		GroupID: groupViews,
		Example: "  things3 anytime\n  things3 anytime --json",
		Args:    cobra.NoArgs,
		RunE:    withClient(v.run),
	}
}

func newSomedayCmd() *cobra.Command {
	v := view{
		fetch: func(cmd *cobra.Command, client *things3.Client) ([]things3.Item, error) {
			return client.Someday(cmd.Context())
		},
		row: defaultRow,
	}
	return &cobra.Command{
		Use:     nameSomeday,
		Short:   "List Someday todos and projects",
		GroupID: groupViews,
		Example: "  things3 someday",
		Args:    cobra.NoArgs,
		RunE:    withClient(v.run),
	}
}

func newLogbookCmd() *cobra.Command {
	v := view{
		fetch: func(cmd *cobra.Command, client *things3.Client) ([]things3.Item, error) {
			days, err := daysWindow(cmd)
			if err != nil {
				return nil, err
			}
			var since time.Time
			if days > 0 {
				since = client.Now().AddDate(0, 0, -days)
			}
			return client.Logbook(cmd.Context(), since)
		},
		row: defaultRow,
	}
	cmd := &cobra.Command{
		Use:     nameLogbook,
		Short:   "List completed and canceled items, most recent first",
		GroupID: groupViews,
		Example: "  things3 logbook\n  things3 logbook --days 7\n  things3 logbook -n 5 --page 2",
		Args:    cobra.NoArgs,
		RunE:    withClient(v.run),
	}
	cmd.Flags().Int(flagDays, 30, "limit to the last N days (0 = all)")
	return cmd
}

func newDeadlinesCmd() *cobra.Command {
	v := view{
		fetch: func(cmd *cobra.Command, client *things3.Client) ([]things3.Item, error) {
			days, err := daysWindow(cmd)
			if err != nil {
				return nil, err
			}
			items, err := client.Deadlines(cmd.Context())
			if err != nil || days == 0 {
				return items, err
			}
			return withinDays(items, things3.Item.Deadline, client.Now(), days), nil
		},
		row: defaultRow,
	}
	cmd := &cobra.Command{
		Use:     nameDeadlines,
		Short:   "List open items with deadlines, soonest first",
		GroupID: groupViews,
		Example: "  things3 deadlines\n  things3 deadlines --days 7\n  things3 deadlines --json",
		Args:    cobra.NoArgs,
		RunE:    withClient(v.run),
	}
	cmd.Flags().Int(flagDays, 0, "limit to deadlines within the next N days, including overdue (0 = all)")
	return cmd
}

func newTrashCmd() *cobra.Command {
	v := view{
		fetch: func(cmd *cobra.Command, client *things3.Client) ([]things3.Item, error) {
			return client.Trash(cmd.Context())
		},
		row: defaultRow,
	}
	return &cobra.Command{
		Use:     nameTrash,
		Short:   "List trashed todos and projects",
		GroupID: groupViews,
		Example: "  things3 trash\n  things3 trash --json",
		Args:    cobra.NoArgs,
		RunE:    withClient(v.run),
	}
}

// withinDays keeps the items whose date falls on or before now+days. An item
// without that date is dropped: it has no place in a date window.
func withinDays(items []things3.Item, date func(things3.Item) *time.Time, now time.Time, days int) []things3.Item {
	cutoff := now.AddDate(0, 0, days)
	return slices.DeleteFunc(items, func(it things3.Item) bool {
		d := date(it)
		return d == nil || d.After(cutoff)
	})
}

// groupToday sections a page into Today / This Evening, but only when evening
// todos are present; otherwise a plain list reads cleaner.
func groupToday(items []things3.Item) []itemGroup {
	isEvening := func(it things3.Item) bool { return it.Todo != nil && it.Todo.Evening }
	if !slices.ContainsFunc(items, isEvening) {
		return nil
	}
	var today, evening []things3.Item
	for _, it := range items {
		if isEvening(it) {
			evening = append(evening, it)
			continue
		}
		today = append(today, it)
	}
	groups := make([]itemGroup, 0, 2)
	if len(today) > 0 {
		groups = append(groups, itemGroup{Header: "Today", Items: today})
	}
	return append(groups, itemGroup{Header: "This Evening", Items: evening})
}

// groupByUpcomingDate groups date-sorted items under "YYYY-MM-DD Weekday"
// headers, keyed by the day Upcoming files each one under.
func groupByUpcomingDate(items []things3.Item) []itemGroup {
	return groupBy(items, func(it things3.Item) string {
		if d := it.UpcomingDate(); d != nil {
			return d.Format("2006-01-02 Monday")
		}
		return "No date"
	})
}

// groupByContainer groups todos by their project, else area, else "No Project",
// preserving first-seen order.
func groupByContainer(items []things3.Item) []itemGroup {
	return groupBy(items, func(it things3.Item) string {
		switch {
		case it.Todo != nil && it.Todo.ProjectTitle != "":
			return it.Todo.ProjectTitle
		case it.Todo != nil && it.Todo.AreaTitle != "":
			return it.Todo.AreaTitle
		default:
			return "No Project"
		}
	})
}

// groupBy buckets items by header in first-seen order.
func groupBy(items []things3.Item, header func(things3.Item) string) []itemGroup {
	var groups []itemGroup
	index := map[string]int{}
	for _, it := range items {
		h := header(it)
		i, ok := index[h]
		if !ok {
			i = len(groups)
			index[h] = i
			groups = append(groups, itemGroup{Header: h})
		}
		groups[i].Items = append(groups[i].Items, it)
	}
	return groups
}
