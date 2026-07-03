package cmd

import (
	"slices"
	"time"

	"github.com/spf13/cobra"

	"github.com/moond4rk/things3"
)

const flagDays = "days"

func newTodayCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     nameToday,
		Short:   "List today's todos, including This Evening",
		GroupID: groupViews,
		Example: "  things3 today\n  things3 today --sort title\n  things3 today --json",
		Args:    cobra.NoArgs,
		RunE:    withClient(runToday),
	}
	return cmd
}

func runToday(cmd *cobra.Command, _ []string, client *things3.Client) error {
	todos, err := client.Today(cmd.Context())
	if err != nil {
		return err
	}
	return outputTodoList(cmd, todos, todayGrouper, defaultRow)
}

// todayGrouper sections a page into Today / This Evening, but only when evening
// todos are present; otherwise a plain list reads cleaner.
func todayGrouper(todos []things3.Todo) []todoGroup {
	if !hasEvening(todos) {
		return nil
	}
	return groupTodayTodos(todos)
}

func hasEvening(todos []things3.Todo) bool {
	for i := range todos {
		if todos[i].Evening {
			return true
		}
	}
	return false
}

func newInboxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "inbox",
		Short:   "List todos in the Inbox",
		GroupID: groupViews,
		Example: "  things3 inbox\n  things3 inbox --json",
		Args:    cobra.NoArgs,
		RunE:    withClient(runInbox),
	}
	return cmd
}

func runInbox(cmd *cobra.Command, _ []string, client *things3.Client) error {
	todos, err := client.Todos().Start().Inbox().Status().Incomplete().All(cmd.Context())
	if err != nil {
		return err
	}
	return outputTodoList(cmd, todos, nil, defaultRow)
}

func newUpcomingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "upcoming",
		Short:   "List scheduled todos grouped by date",
		GroupID: groupViews,
		Example: "  things3 upcoming\n  things3 upcoming --days 7\n  things3 upcoming --tag work",
		Args:    cobra.NoArgs,
		RunE:    withClient(runUpcoming),
	}
	cmd.Flags().Int(flagDays, 0, "limit to todos scheduled within the next N days (0 = all)")
	return cmd
}

func runUpcoming(cmd *cobra.Command, _ []string, client *things3.Client) error {
	days, _ := cmd.Flags().GetInt(flagDays)
	todos, err := client.Upcoming(cmd.Context())
	if err != nil {
		return err
	}
	if days > 0 {
		cutoff := time.Now().AddDate(0, 0, days)
		todos = slices.DeleteFunc(todos, func(t things3.Todo) bool {
			return t.StartDate == nil || t.StartDate.After(cutoff)
		})
	}
	return outputTodoList(cmd, todos, groupByStartDate, defaultRow)
}

func newAnytimeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     nameAnytime,
		Short:   "List Anytime todos grouped by project or area",
		GroupID: groupViews,
		Example: "  things3 anytime\n  things3 anytime --json",
		Args:    cobra.NoArgs,
		RunE:    withClient(runAnytime),
	}
	return cmd
}

func runAnytime(cmd *cobra.Command, _ []string, client *things3.Client) error {
	todos, err := client.Todos().Start().Anytime().Status().Incomplete().All(cmd.Context())
	if err != nil {
		return err
	}
	// Rows drop the container segment because the view already groups by it.
	return outputTodoList(cmd, todos, groupByContainer, noContainerRow)
}

func newSomedayCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     nameSomeday,
		Short:   "List Someday todos with no scheduled date",
		GroupID: groupViews,
		Example: "  things3 someday",
		Args:    cobra.NoArgs,
		RunE:    withClient(runSomeday),
	}
	return cmd
}

func runSomeday(cmd *cobra.Command, _ []string, client *things3.Client) error {
	todos, err := client.Todos().
		StartDate().Exists(false).
		Start().Someday().
		Status().Incomplete().
		All(cmd.Context())
	if err != nil {
		return err
	}
	return outputTodoList(cmd, todos, nil, defaultRow)
}

func newLogbookCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "logbook",
		Short:   "List completed and canceled todos, most recent first",
		GroupID: groupViews,
		Example: "  things3 logbook\n  things3 logbook --days 7\n  things3 logbook -n 5 --page 2",
		Args:    cobra.NoArgs,
		RunE:    withClient(runLogbook),
	}
	cmd.Flags().Int(flagDays, 30, "limit to the last N days (0 = all)")
	return cmd
}

func runLogbook(cmd *cobra.Command, _ []string, client *things3.Client) error {
	days, _ := cmd.Flags().GetInt(flagDays)
	q := client.Todos().Status().Any().StopDate().Exists(true)
	if days > 0 {
		q = q.StopDate().After(time.Now().AddDate(0, 0, -days))
	}
	todos, err := q.All(cmd.Context())
	if err != nil {
		return err
	}
	sortByStopTimeDesc(todos)
	return outputTodoList(cmd, todos, nil, defaultRow)
}

func newDeadlinesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     nameDeadlines,
		Short:   "List todos with deadlines, soonest first",
		GroupID: groupViews,
		Example: "  things3 deadlines\n  things3 deadlines --days 7\n  things3 deadlines --json",
		Args:    cobra.NoArgs,
		RunE:    withClient(runDeadlines),
	}
	cmd.Flags().Int(flagDays, 0, "limit to deadlines within the next N days, including overdue (0 = all)")
	return cmd
}

func runDeadlines(cmd *cobra.Command, _ []string, client *things3.Client) error {
	days, _ := cmd.Flags().GetInt(flagDays)
	q := client.Todos().Deadline().Exists(true).Status().Incomplete()
	if days > 0 {
		q = q.Deadline().OnOrBefore(time.Now().AddDate(0, 0, days))
	}
	todos, err := q.All(cmd.Context())
	if err != nil {
		return err
	}
	slices.SortStableFunc(todos, func(a, b things3.Todo) int {
		return compareTimePtrAsc(a.Deadline, b.Deadline)
	})
	return outputTodoList(cmd, todos, nil, defaultRow)
}

func newTrashCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "trash",
		Short:   "List trashed todos and projects",
		GroupID: groupViews,
		Example: "  things3 trash\n  things3 trash --json",
		Args:    cobra.NoArgs,
		RunE:    withClient(runTrash),
	}
	return cmd
}

func runTrash(cmd *cobra.Command, _ []string, client *things3.Client) error {
	ctx := cmd.Context()
	todos, err := client.Todos().Trashed(true).Status().Any().All(ctx)
	if err != nil {
		return err
	}
	projects, err := client.Projects().Trashed(true).Status().Any().All(ctx)
	if err != nil {
		return err
	}
	items := make([]mixedItem, 0, len(todos)+len(projects))
	for i := range todos {
		items = append(items, todoMixed(&todos[i]))
	}
	for i := range projects {
		items = append(items, projectMixed(&projects[i]))
	}
	return outputMixedList(cmd, items)
}

// groupTodayTodos splits Today into the Today section and, when present, the
// This Evening section (the app's bottom section).
func groupTodayTodos(todos []things3.Todo) []todoGroup {
	today := make([]things3.Todo, 0, len(todos))
	var evening []things3.Todo
	for i := range todos {
		if todos[i].Evening {
			evening = append(evening, todos[i])
			continue
		}
		today = append(today, todos[i])
	}
	var groups []todoGroup
	if len(today) > 0 {
		groups = append(groups, todoGroup{Header: "Today", Todos: today})
	}
	if len(evening) > 0 {
		groups = append(groups, todoGroup{Header: "This Evening", Todos: evening})
	}
	return groups
}

// groupByStartDate groups start-date-sorted todos under "YYYY-MM-DD Weekday"
// headers in ascending order.
func groupByStartDate(todos []things3.Todo) []todoGroup {
	var groups []todoGroup
	index := map[string]int{}
	for i := range todos {
		header := "No date"
		if todos[i].StartDate != nil {
			header = todos[i].StartDate.Format("2006-01-02 Monday")
		}
		appendToGroup(&groups, index, header, &todos[i])
	}
	return groups
}

// groupByContainer groups todos by their project, else area, else "No Project",
// preserving first-seen order.
func groupByContainer(todos []things3.Todo) []todoGroup {
	var groups []todoGroup
	index := map[string]int{}
	for i := range todos {
		appendToGroup(&groups, index, containerHeader(&todos[i]), &todos[i])
	}
	return groups
}

func containerHeader(t *things3.Todo) string {
	switch {
	case t.ProjectTitle != "":
		return t.ProjectTitle
	case t.AreaTitle != "":
		return t.AreaTitle
	default:
		return "No Project"
	}
}

func appendToGroup(groups *[]todoGroup, index map[string]int, header string, todo *things3.Todo) {
	if i, ok := index[header]; ok {
		(*groups)[i].Todos = append((*groups)[i].Todos, *todo)
		return
	}
	index[header] = len(*groups)
	*groups = append(*groups, todoGroup{Header: header, Todos: []things3.Todo{*todo}})
}

// compareTimePtrAsc orders non-nil times ascending, sorting nil last.
func compareTimePtrAsc(a, b *time.Time) int {
	switch {
	case a == nil && b == nil:
		return 0
	case a == nil:
		return 1
	case b == nil:
		return -1
	default:
		return a.Compare(*b)
	}
}

// stopTime returns when a todo left the active list: completion time, else
// cancellation time, else nil.
func stopTime(t *things3.Todo) *time.Time {
	if t.CompletedAt != nil {
		return t.CompletedAt
	}
	return t.CanceledAt
}

// sortByStopTimeDesc orders todos by stop time descending so -n N yields the N
// most recent entries. Todos without a stop time sort last.
func sortByStopTimeDesc(todos []things3.Todo) {
	slices.SortStableFunc(todos, func(a, b things3.Todo) int {
		ta, tb := stopTime(&a), stopTime(&b)
		switch {
		case ta == nil && tb == nil:
			return 0
		case ta == nil:
			return 1
		case tb == nil:
			return -1
		default:
			return tb.Compare(*ta)
		}
	})
}
